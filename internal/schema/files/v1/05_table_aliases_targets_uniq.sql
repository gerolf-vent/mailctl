/***************************************************************
 * Tables for globally unique alias targets
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

/**
 * Sequence for alias target ids
 * Used to generate unique IDs for alias targets across multiple tables.
 */
CREATE SEQUENCE shared.aliases_targets_id_seq AS INT;

/**
 * Table of alias target ids
 * Ensures that each alias target has a unique global ID.
 *
 * To use this with a table:
 * 1. Create a foreign key reference:
 *    ```sql
 *    FOREIGN KEY (ID)
 *      REFERENCES shared.aliases_targets_id(ID)
 *          ON DELETE CASCADE
 *          ON UPDATE CASCADE
 *    ```
 * 2. Use `DEFAULT nextval('shared.aliases_targets_id_seq')` to generate new IDs.
 * 3. Create triggers to call `hook_register_foreign_id` on INSERT, UPDATE, DELETE:
 *    ```sql
 *    CREATE TRIGGER trigger_register_foreign_id
 *        BEFORE INSERT OR UPDATE OR DELETE ON your_alias_target_table
 *        FOR EACH ROW
 *        EXECUTE FUNCTION hook_register_foreign_id('shared.aliases_targets_id');
 *    ```
 */
CREATE TABLE shared.aliases_targets_id (
    ID INT PRIMARY KEY DEFAULT nextval('shared.aliases_targets_id_seq'),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_aliases_targets_id_not_deleted ON shared.aliases_targets_id(ID) WHERE deleted_at IS NULL;
