/***************************************************************
 * Tables for globally unique domains
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

/**
 * Sequence for domain ids
 * Used to generate unique IDs for domains across multiple tables.
 */
CREATE SEQUENCE shared.domains_id_seq AS INT;

/**
 * Table of all domain ids
 * Ensures that each domain has a unique global ID.
 *
 * To use this with a table:
 * 1. Create a foreign key reference:
 *    ```sql
 *    FOREIGN KEY (ID)
 *      REFERENCES shared.domains_id(ID)
 *          ON DELETE CASCADE
 *          ON UPDATE CASCADE
 *    ```
 * 2. Use `DEFAULT nextval('shared.domains_id_seq')` to generate new IDs.
 * 3. Create triggers to call `hook_register_foreign_id` on INSERT, UPDATE, DELETE:
 *    ```sql
 *    CREATE TRIGGER trigger_register_foreign_id
 *        BEFORE INSERT OR UPDATE OR DELETE ON your_domain_table
 *        FOR EACH ROW
 *        EXECUTE FUNCTION hook_register_foreign_id('shared.domains_id');
 *    ```
 */
CREATE TABLE shared.domains_id (
    ID INT PRIMARY KEY DEFAULT nextval('shared.domains_id_seq'),
    deleted_at TIMESTAMPTZ 
);

CREATE INDEX idx_domains_id_not_deleted ON shared.domains_id(ID) WHERE deleted_at IS NULL;

/**
 * Table of domain ids, which recipients can be defined on
 * This excludes canonical domains themselves to ensure a flat mapping
 * of canonical domains to non-canonical domains.
 *
 * @see `shared.domains_id`
 */
CREATE TABLE shared.domains_id_recipientable (
    ID INT PRIMARY KEY
        REFERENCES shared.domains_id(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE,
    deleted_at TIMESTAMPTZ 
);

CREATE INDEX idx_domains_id_recipientable_not_deleted ON shared.domains_id_recipientable(ID) WHERE deleted_at IS NULL;

/**
 * Table for globally unique domains
 * Ensures that each domain (fqdn) is unique across all domain tables.
 *
 * To enforce this with a table:
 * 1. Create a foreign key reference:
 *    ```sql
 *    FOREIGN KEY (fqdn)
 *      REFERENCES shared.domains_uniq(fqdn)
 *          ON DELETE CASCADE
 *          ON UPDATE CASCADE
 *    ```
 * 2. Create triggers to call `hook_check_domains_uniq` on INSERT, UPDATE, DELETE:
 *    ```sql
 *    CREATE TRIGGER trigger_check_domains_uniq
 *        BEFORE INSERT OR UPDATE OR DELETE ON your_domain_table
 *        FOR EACH ROW
 *        EXECUTE FUNCTION hook_check_domains_uniq('fqdn');
 *    ```
 */
CREATE TABLE shared.domains_uniq (
    fqdn VARCHAR(256) PRIMARY KEY
        CHECK (check_domain_fqdn(fqdn))
);

/**
 * Trigger to ensure uniqueness of domains across multiple tables
 *
 * @param TG_ARGV[0] Name of the domain fqdn column
 */
CREATE FUNCTION hook_check_domains_uniq()
RETURNS TRIGGER AS $$
DECLARE
    old_val TEXT;
    new_val TEXT;
BEGIN
    -- Get column values dynamically using the column name from TG_ARGV[1]
    IF TG_OP = 'UPDATE' OR TG_OP = 'DELETE' THEN
        EXECUTE format('SELECT ($1).%I::TEXT', TG_ARGV[0]) INTO old_val USING OLD;
    END IF;
    IF TG_OP = 'UPDATE' OR TG_OP = 'INSERT' THEN
        EXECUTE format('SELECT ($1).%I::TEXT', TG_ARGV[0]) INTO new_val USING NEW;
    END IF;

    -- Passthrough updates that do not change the column value
    IF TG_OP = 'UPDATE' AND old_val = new_val THEN
        RETURN NEW;
    END IF;

    -- Cascade deallocations to the shared table
    IF TG_OP = 'DELETE' OR TG_OP = 'UPDATE' THEN
        DELETE FROM shared.domains_uniq WHERE fqdn = old_val;
    END IF;

    -- Allocate new entry in the shared table
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        INSERT INTO shared.domains_uniq (fqdn) VALUES (new_val);
    END IF;

    -- Return the appropriate row
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
