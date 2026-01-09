/***************************************************************
 * Table for managed domains
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

CREATE TABLE domains_managed (
    ID INT PRIMARY KEY
        REFERENCES shared.domains_id(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE
        DEFAULT nextval('shared.domains_id_seq'),
    fqdn VARCHAR(256) NOT NULL
        REFERENCES shared.domains_uniq(fqdn)
            ON DELETE CASCADE
            ON UPDATE CASCADE
        CHECK (check_domain_fqdn(fqdn)),
    transport_id INT NOT NULL  -- Default transport for this domain which can be overridden for each mailbox
        REFERENCES transports(ID)
            ON DELETE RESTRICT
            ON UPDATE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT(true),  -- Whether receiving, sending and login on this domain is enabled for all recipients
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_domains_managed_fqdn ON domains_managed(fqdn) WHERE deleted_at IS NULL;
CREATE INDEX idx_domains_managed_transport_id ON domains_managed(transport_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_domains_managed_enabled ON domains_managed(enabled) WHERE deleted_at IS NULL;

CREATE TRIGGER trigger_updated_at
    BEFORE UPDATE ON domains_managed
    FOR EACH ROW
    EXECUTE FUNCTION hook_update_updated_at();

CREATE TRIGGER trigger_audit
    AFTER INSERT OR UPDATE OR DELETE ON domains_managed
    FOR EACH ROW
    EXECUTE FUNCTION hook_audit();

CREATE TRIGGER trigger_register_shared_id
    BEFORE INSERT OR UPDATE OR DELETE ON domains_managed
    FOR EACH ROW
    EXECUTE FUNCTION hook_register_shared_id('domains_id');

CREATE TRIGGER trigger_register_shared_id_recipientable
    BEFORE INSERT OR UPDATE OR DELETE ON domains_managed
    FOR EACH ROW
    EXECUTE FUNCTION hook_register_shared_id('domains_id_recipientable');

CREATE TRIGGER trigger_check_domains_uniq
    BEFORE INSERT OR UPDATE OR DELETE ON domains_managed
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_domains_uniq('fqdn');

CREATE TRIGGER trigger_cascade_soft_delete_domains_canonical
    AFTER UPDATE ON domains_managed
    FOR EACH ROW
    EXECUTE FUNCTION hook_cascade_soft_delete('domains_canonical', 'target_domain_id');

CREATE TRIGGER trigger_cascade_soft_delete_mailboxes
    AFTER UPDATE ON domains_managed
    FOR EACH ROW
    EXECUTE FUNCTION hook_cascade_soft_delete('mailboxes', 'domain_id');

CREATE TRIGGER trigger_cascade_soft_delete_aliases
    AFTER UPDATE ON domains_managed
    FOR EACH ROW
    EXECUTE FUNCTION hook_cascade_soft_delete('aliases', 'domain_id');

CREATE TRIGGER trigger_cascade_soft_delete_catchall_targets
    AFTER UPDATE ON domains_managed
    FOR EACH ROW
    EXECUTE FUNCTION hook_cascade_soft_delete('domains_catchall_targets', 'domain_id');

CREATE TRIGGER trigger_cascade_soft_delete_remotes_send_grants
    AFTER UPDATE ON domains_managed
    FOR EACH ROW
    EXECUTE FUNCTION hook_cascade_soft_delete('remotes_send_grants', 'domain_id');

CREATE TRIGGER trigger_check_foreign_key_soft_delete
    BEFORE INSERT OR UPDATE ON domains_managed
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_foreign_key_soft_delete('transport_id', 'transports');
