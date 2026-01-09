/***************************************************************
 * Table for catch-all targets
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

CREATE TABLE domains_catchall_targets (
    ID SERIAL PRIMARY KEY,
    domain_id INT NOT NULL  -- Domain from which to forward emails
        REFERENCES shared.domains_id_recipientable(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE,
    recipient_id INT NOT NULL  -- Target recipient to which to forward emails
        REFERENCES shared.recipients_id(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE,
    forwarding_to_target_enabled BOOLEAN NOT NULL DEFAULT(true),
    fallback_only BOOLEAN NOT NULL DEFAULT(true),  -- true: only if no non-catchall match; false: always
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    UNIQUE(domain_id, recipient_id)
);

CREATE INDEX idx_domains_catchall_targets_domain_id ON domains_catchall_targets(domain_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_domains_catchall_targets_recipient_id ON domains_catchall_targets(recipient_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_domains_catchall_targets_both_ids ON domains_catchall_targets(domain_id, recipient_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_domains_catchall_targets_forwarding_enabled ON domains_catchall_targets(forwarding_to_target_enabled) WHERE deleted_at IS NULL;

CREATE TRIGGER trigger_updated_at
    BEFORE UPDATE ON domains_catchall_targets
    FOR EACH ROW
    EXECUTE FUNCTION hook_update_updated_at();

CREATE TRIGGER trigger_audit
    AFTER INSERT OR UPDATE OR DELETE ON domains_catchall_targets
    FOR EACH ROW
    EXECUTE FUNCTION hook_audit();

CREATE TRIGGER trigger_check_foreign_key_soft_delete_domain
    BEFORE INSERT OR UPDATE ON domains_catchall_targets
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_foreign_key_soft_delete('domain_id', 'shared.domains_id');

CREATE TRIGGER trigger_check_foreign_key_soft_delete_recipient
    BEFORE INSERT OR UPDATE ON domains_catchall_targets
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_foreign_key_soft_delete('recipient_id', 'shared.recipients_id');
