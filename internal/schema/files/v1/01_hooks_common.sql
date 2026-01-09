/***************************************************************
 * Trigger hooks
 *
 * Various common trigger hooks useful for many/all tables.
 * Every table should use the following hooks:
 * - hook_update_updated_at
 * - hook_audit
 * Tables with shared global ids:
 * - hook_register_shared_id
 * Tables with foreign keys:
 * - hook_cascade_soft_delete
 * - hook_check_foreign_key_soft_delete
 * - hook_prohibit_delete_in_use
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

/**
 * Auto-update update timestamp on every update
 *
 * This just sets the "updated_at" column to the current timestamp on every update.
 * You can create a trigger like this:
 * ```sql
 * CREATE TRIGGER trigger_your_table_updated_at
 *     BEFORE UPDATE ON your_table
 *     FOR EACH ROW
 *     EXECUTE FUNCTION hook_update_updated_at();
 * ```
 */
CREATE FUNCTION hook_update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

/**
 * Generic audit trigger function
 *
 * This captures changes (INSERT, UPDATE, DELETE) to an audit.log table.
 * You can create a trigger like this:
 * ```sql
 * CREATE TRIGGER trigger_your_table_audit
 *     AFTER INSERT OR UPDATE OR DELETE ON your_table
 *     FOR EACH ROW
 *     EXECUTE FUNCTION hook_audit();
 * ```
 */
CREATE FUNCTION hook_audit()
RETURNS TRIGGER AS $$
DECLARE
    changed_by TEXT;
BEGIN
    BEGIN
        changed_by := current_setting('mailctl.current_user');
    EXCEPTION WHEN undefined_object THEN
        changed_by := current_user;
    END;

    IF TG_OP = 'DELETE' THEN
        INSERT INTO audit.log (table_name, record_id, operation, data_old, changed_by)
        VALUES (TG_TABLE_NAME, OLD.ID, TG_OP, row_to_json(OLD), changed_by);
        RETURN OLD;
    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO audit.log (table_name, record_id, operation, data_old, data_new, changed_by)
        VALUES (TG_TABLE_NAME, NEW.ID, TG_OP, row_to_json(OLD), row_to_json(NEW), changed_by);
        RETURN NEW;
    ELSIF TG_OP = 'INSERT' THEN
        INSERT INTO audit.log (table_name, record_id, operation, data_new, changed_by)
        VALUES (TG_TABLE_NAME, NEW.ID, TG_OP, row_to_json(NEW), changed_by);
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

/**
 * Register a new id in a shared.* table
 *
 * This enforces global id space across tables that share a single id sequence.
 * You can create a trigger like this:
 * ```sql
 * CREATE TRIGGER trigger_your_table_register_shared_id
 *     BEFORE INSERT OR UPDATE OR DELETE ON your_table
 *     FOR EACH ROW
 *     EXECUTE FUNCTION hook_register_shared_id('domains_id');
 * ```
 *
 * @param TG_ARGV[0] Target shared table name (without schema), e.g. 'domains_id'
 */
CREATE FUNCTION hook_register_shared_id()
RETURNS TRIGGER AS $$
BEGIN
    -- Passthrough updates that do not change the id or deleted_at
    IF TG_OP = 'UPDATE' AND NEW.ID = OLD.ID AND NEW.deleted_at = OLD.deleted_at THEN
        RETURN NEW;
    END IF;

    IF TG_OP = 'DELETE' THEN
        EXECUTE format('DELETE FROM shared.%I WHERE ID = $1;', TG_ARGV[0]) USING OLD.ID;
        RETURN OLD;
    END IF;

    IF TG_OP = 'INSERT' THEN
        EXECUTE format('INSERT INTO shared.%I (ID, deleted_at) VALUES ($1, $2);', TG_ARGV[0]) USING NEW.ID, NEW.deleted_at;
    END IF;

    IF TG_OP = 'UPDATE' THEN
        EXECUTE format('UPDATE shared.%I SET ID = $2, deleted_at = $3 WHERE ID = $1;', TG_ARGV[0]) USING OLD.ID, NEW.ID, NEW.deleted_at;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

/**
 * Cascade soft deletes and restores from parent to child tables
 * When a parent is soft-deleted, children are soft-deleted with the same timestamp
 * When a parent is restored, children with matching deleted_at timestamp are also restored
 *
 * @param TG_ARGV[0] Child table name (which holds the foreign key)
 * @param TG_ARGV[1] Foreign key column name in the child table
 */
CREATE FUNCTION hook_cascade_soft_delete()
RETURNS TRIGGER AS $$
BEGIN
    -- Handle soft delete operations
    IF TG_OP = 'UPDATE' AND OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        -- Soft delete child records with the same timestamp
        EXECUTE format(
            'UPDATE %I SET deleted_at = $1 WHERE %I = $2 AND deleted_at IS NULL',
            TG_ARGV[0], TG_ARGV[1]
        ) USING NEW.deleted_at, NEW.ID;
    END IF;
    
    -- Handle restore operations (deleted_at was set, now being cleared)
    IF TG_OP = 'UPDATE' AND OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NULL THEN
        -- Restore child records that were deleted at the same time as the parent
        EXECUTE format(
            'UPDATE %I SET deleted_at = NULL WHERE %I = $1 AND deleted_at = $2',
            TG_ARGV[0], TG_ARGV[1]
        ) USING NEW.ID, OLD.deleted_at;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

/**
 * Checks that a foreign key does not point to a soft-deleted row.
 * This prevents:
 * 1. Inserting a new row referencing a soft-deleted parent.
 * 2. Updating a row to reference a soft-deleted parent.
 * 3. Restoring a row that references a soft-deleted parent.
 *
 * @param TG_ARGV[0] Foreign key column name in the child table
 * @param TG_ARGV[1] Parent table name (which holds the primary key)
 */
CREATE FUNCTION hook_check_foreign_key_soft_delete()
RETURNS TRIGGER AS $$
DECLARE
    parent_id INT;
    parent_deleted BOOLEAN;
BEGIN
    -- If the child row is deleted, we don't care about the parent's state
    IF NEW.deleted_at IS NOT NULL THEN
        RETURN NEW;
    END IF;

    -- Get the value of the foreign key column
    EXECUTE format('SELECT ($1).%I', TG_ARGV[0]) INTO parent_id USING NEW;

    -- Skip check if the foreign key is NULL
    IF parent_id IS NULL THEN
        RETURN NEW;
    END IF;

    EXECUTE format(
        'SELECT EXISTS(SELECT 1 FROM %s WHERE ID = $1 AND deleted_at IS NOT NULL)',
        TG_ARGV[1]
    ) INTO parent_deleted USING parent_id;
    
    IF parent_deleted THEN
        RAISE EXCEPTION 'referenced parent in % is soft-deleted', TG_ARGV[1];
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

/**
 * Prohibits deletion of an entry if it is referenced by another table with a non-nullable
 * foreign key. This is useful when the foreign key has ON DELETE RESTRICT but the column
 * does not allow NULL values.
 *
 * You can create a trigger like this:
 * ```sql
 * CREATE TRIGGER trigger_prohibit_delete_in_use
 *     BEFORE UPDATE OR DELETE ON your_table
 *     FOR EACH ROW
 *     EXECUTE FUNCTION hook_prohibit_delete_in_use('child_table', 'fk_column');
 * ```
 *
 * @param TG_ARGV[0] Child table name (which holds the foreign key)
 * @param TG_ARGV[1] Foreign key column name in the child table
 */
CREATE FUNCTION hook_prohibit_delete_in_use()
RETURNS TRIGGER AS $$
DECLARE
    in_use BOOLEAN;
BEGIN
    -- Passthrough updates that are not a soft-delete
    IF TG_OP = 'UPDATE' AND (NEW.deleted_at IS NULL OR NEW.deleted_at = OLD.deleted_at) THEN
        RETURN NEW;
    END IF;

    EXECUTE format(
        'SELECT EXISTS(SELECT 1 FROM %I WHERE %I = $1 AND deleted_at IS NULL LIMIT 1)',
        TG_ARGV[0], TG_ARGV[1]
    ) INTO in_use USING OLD.ID;

    IF in_use THEN
        RAISE EXCEPTION 'cannot delete %: in use by % (%.%)',
            TG_TABLE_NAME, TG_ARGV[0], TG_ARGV[0], TG_ARGV[1];
    END IF;

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
