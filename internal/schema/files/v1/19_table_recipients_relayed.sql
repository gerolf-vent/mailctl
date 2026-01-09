/***************************************************************
 * Table for relayed recipients
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

CREATE TABLE recipients_relayed (
    ID INT PRIMARY KEY
        REFERENCES shared.recipients_id(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE
        DEFAULT nextval('shared.recipients_id_seq'),
    domain_id INT NOT NULL
        REFERENCES domains_relayed(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE,
    name VARCHAR(256) NOT NULL
        CHECK (check_mail_address_name(name)),
    enabled BOOLEAN NOT NULL DEFAULT(true),  -- Whether relaying for this recipient is enabled
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    FOREIGN KEY (domain_id, name)
        REFERENCES shared.recipients_uniq(domain_id, name)
            ON DELETE CASCADE
            ON UPDATE CASCADE
);

CREATE INDEX idx_recipients_relayed_domain_and_name ON recipients_relayed(domain_id, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_recipients_relayed_domain_id ON recipients_relayed(domain_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_recipients_relayed_domain_name ON recipients_relayed(domain_id, name) WHERE deleted_at IS NULL;

CREATE TRIGGER trigger_updated_at
    BEFORE UPDATE ON recipients_relayed
    FOR EACH ROW
    EXECUTE FUNCTION hook_update_updated_at();

CREATE TRIGGER trigger_audit
    AFTER INSERT OR UPDATE OR DELETE ON recipients_relayed
    FOR EACH ROW
    EXECUTE FUNCTION hook_audit();

-- Make sure the recipient id is globally unique
CREATE TRIGGER trigger_register_shared_id
    BEFORE INSERT OR UPDATE OR DELETE ON recipients_relayed
    FOR EACH ROW
    EXECUTE FUNCTION hook_register_shared_id('recipients_id');

-- Make sure the recipient (domain_id, name) is globally unique
CREATE TRIGGER trigger_check_recipients_uniq
    BEFORE INSERT OR UPDATE OR DELETE ON recipients_relayed
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_recipients_uniq('domain_id', 'name');

CREATE TRIGGER trigger_check_foreign_key_soft_delete_domain
    BEFORE INSERT OR UPDATE ON recipients_relayed
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_foreign_key_soft_delete('domain_id', 'domains_relayed');

-- Cascade to aliases_targets_recursive where this recipient is a target
CREATE TRIGGER trigger_cascade_soft_delete_alias_targets
    AFTER UPDATE ON recipients_relayed
    FOR EACH ROW
    EXECUTE FUNCTION hook_cascade_soft_delete('aliases_targets_recursive', 'recipient_id');

-- Cascade to domains_catchall_targets where this recipient is a target
CREATE TRIGGER trigger_cascade_soft_delete_catchall_targets
    AFTER UPDATE ON recipients_relayed
    FOR EACH ROW
    EXECUTE FUNCTION hook_cascade_soft_delete('domains_catchall_targets', 'recipient_id');
