/***************************************************************
 * Tables for globally unique recipients
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

/**
 * Sequence for recipient ids
 * Used to generate unique IDs for recipients across multiple tables.
 */
CREATE SEQUENCE shared.recipients_id_seq AS INT;

/**
 * Table of recipient ids
 * Ensures that each recipient has a unique global ID.
 *
 * To use this with a table:
 * 1. Create a foreign key reference:
 *    ```sql
 *    FOREIGN KEY (ID)
 *      REFERENCES shared.recipients_id(ID)
 *          ON DELETE CASCADE
 *          ON UPDATE CASCADE
 *    ```
 * 2. Use `DEFAULT nextval('shared.recipients_id_seq')` to generate new IDs.
 * 3. Create triggers to call `hook_register_foreign_id` on INSERT, UPDATE, DELETE:
 *    ```sql
 *    CREATE TRIGGER trigger_register_foreign_id
 *        BEFORE INSERT OR UPDATE OR DELETE ON your_recipient_table
 *        FOR EACH ROW
 *        EXECUTE FUNCTION hook_register_foreign_id('shared.recipients_id');
 *    ```
 */
CREATE TABLE shared.recipients_id (
    ID INT PRIMARY KEY DEFAULT nextval('shared.recipients_id_seq'),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_recipients_id_not_deleted ON shared.recipients_id(ID) WHERE deleted_at IS NULL;

/**
 * Table for globally unique recipients
 * Ensures that each (domain_id, name) pair is unique across all recipient tables.
 *
 * To enforce this with a table:
 * 1. Create a foreign key reference:
 *    ```sql
 *    FOREIGN KEY (domain_id, name)
 *      REFERENCES shared.recipients_uniq(domain_id, name)
 *          ON DELETE CASCADE
 *          ON UPDATE CASCADE
 *    ```
 * 2. Create triggers to call `hook_check_recipients_uniq` on INSERT, UPDATE, DELETE:
 *    ```sql
 *    CREATE TRIGGER trigger_check_recipients_uniq
 *        BEFORE INSERT OR UPDATE OR DELETE ON your_recipient_table
 *        FOR EACH ROW
 *        EXECUTE FUNCTION hook_check_recipients_uniq('domain_id', 'name');
 *    ```
 */
CREATE TABLE shared.recipients_uniq (
    domain_id INT NOT NULL  -- Domain of the recipient
        REFERENCES shared.domains_id_recipientable(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE,
    name VARCHAR(256) NOT NULL  -- Name part of the recipient email address
        CHECK (check_mail_address_name(name)),
    PRIMARY KEY (domain_id, name)
);

/**
 * Trigger to ensure uniqueness of recipients across multiple tables
 *
 * @param TG_ARGV[0] Name of the recipient domain_id column
 * @param TG_ARGV[1] Name of the recipient name column
 */
CREATE FUNCTION hook_check_recipients_uniq()
RETURNS TRIGGER AS $$
DECLARE
    old_domain_id INT;
    old_name TEXT;
    new_domain_id INT;
    new_name TEXT;
BEGIN
    -- Get column values dynamically using the column names from TG_ARGV
    IF TG_OP = 'UPDATE' OR TG_OP = 'DELETE' THEN
        EXECUTE format('SELECT ($1).%I::INT', TG_ARGV[0]) INTO old_domain_id USING OLD;
        EXECUTE format('SELECT ($1).%I::TEXT', TG_ARGV[1]) INTO old_name USING OLD;
    END IF;
    IF TG_OP = 'UPDATE' OR TG_OP = 'INSERT' THEN
        EXECUTE format('SELECT ($1).%I::INT', TG_ARGV[0]) INTO new_domain_id USING NEW;
        EXECUTE format('SELECT ($1).%I::TEXT', TG_ARGV[1]) INTO new_name USING NEW;
    END IF; 

    -- Passthrough updates that do not change the column values
    IF TG_OP = 'UPDATE' AND old_domain_id = new_domain_id AND old_name = new_name THEN
        RETURN NEW;
    END IF;

    -- Cascade deallocations to the shared table
    IF TG_OP = 'DELETE' OR TG_OP = 'UPDATE' THEN
        DELETE FROM shared.recipients_uniq WHERE domain_id = old_domain_id AND name = old_name;
    END IF;

    -- Allocate new entry in the shared table
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        INSERT INTO shared.recipients_uniq (domain_id, name) VALUES (new_domain_id, new_name);
    END IF;

    -- Return the appropriate row
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
