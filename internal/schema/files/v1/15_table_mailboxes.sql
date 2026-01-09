/***************************************************************
 * Table for mailboxes
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

CREATE TABLE mailboxes (
    ID INT PRIMARY KEY
        REFERENCES shared.recipients_id(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE
        DEFAULT nextval('shared.recipients_id_seq'),
    domain_id INT NOT NULL
        REFERENCES domains_managed(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE,
    name VARCHAR(256) NOT NULL
        CHECK (check_mail_address_name(name)),
    transport_id INT
        REFERENCES transports(ID)
            ON DELETE RESTRICT
            ON UPDATE CASCADE,
    password_hash VARCHAR(1024),  -- Password hash for direct auth (e.g., bcrypt, argon2)
    storage_quota INT,  -- in MB
    login_enabled BOOLEAN NOT NULL DEFAULT(true),  -- Whether the user can login to the server with a mail client
    receiving_enabled BOOLEAN NOT NULL DEFAULT(true),  -- Whether the user can receive mails
    sending_enabled BOOLEAN NOT NULL DEFAULT(true),  -- Whether the user can send mails
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    FOREIGN KEY (domain_id, name)
        REFERENCES shared.recipients_uniq(domain_id, name)
            ON DELETE CASCADE
            ON UPDATE CASCADE
);

CREATE INDEX idx_mailboxes_domain_and_name ON mailboxes(domain_id, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_mailboxes_domain_id ON mailboxes(domain_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_mailboxes_transport_id ON mailboxes(transport_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_mailboxes_domain_transport ON mailboxes(domain_id, transport_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_mailboxes_enabled_flags ON mailboxes(login_enabled, receiving_enabled, sending_enabled) WHERE deleted_at IS NULL;
CREATE INDEX idx_mailboxes_domain_name_sending ON mailboxes(domain_id, name, sending_enabled) WHERE deleted_at IS NULL;
CREATE INDEX idx_mailboxes_domain_name_receiving ON mailboxes(domain_id, name, receiving_enabled) WHERE deleted_at IS NULL;

CREATE TRIGGER trigger_updated_at
    BEFORE UPDATE ON mailboxes
    FOR EACH ROW
    EXECUTE FUNCTION hook_update_updated_at();

CREATE TRIGGER trigger_audit
    AFTER INSERT OR UPDATE OR DELETE ON mailboxes
    FOR EACH ROW
    EXECUTE FUNCTION hook_audit();

-- Make sure the recipient id is globally unique
CREATE TRIGGER trigger_register_shared_id
    BEFORE INSERT OR UPDATE OR DELETE ON mailboxes
    FOR EACH ROW
    EXECUTE FUNCTION hook_register_shared_id('recipients_id');

-- Make sure the recipient (domain_id, name) is globally unique
CREATE TRIGGER trigger_check_recipients_uniq
    BEFORE INSERT OR UPDATE OR DELETE ON mailboxes
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_recipients_uniq('domain_id', 'name');

CREATE TRIGGER trigger_cascade_soft_delete_aliases_targets_recursive
    AFTER UPDATE ON mailboxes
    FOR EACH ROW
    EXECUTE FUNCTION hook_cascade_soft_delete('aliases_targets_recursive', 'recipient_id');

CREATE TRIGGER trigger_cascade_soft_delete_domains_catchall_targets
    AFTER UPDATE ON mailboxes
    FOR EACH ROW
    EXECUTE FUNCTION hook_cascade_soft_delete('domains_catchall_targets', 'recipient_id');

CREATE TRIGGER trigger_check_foreign_key_soft_delete_domain
    BEFORE INSERT OR UPDATE ON mailboxes
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_foreign_key_soft_delete('domain_id', 'domains_managed');

CREATE TRIGGER trigger_check_foreign_key_soft_delete_transport
    BEFORE INSERT OR UPDATE ON mailboxes
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_foreign_key_soft_delete('transport_id', 'transports');
