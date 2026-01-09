/***************************************************************
 * Table for remote send grants
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

CREATE TABLE remotes_send_grants (
    ID SERIAL PRIMARY KEY,
    remote_id INT NOT NULL
        REFERENCES remotes(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE,
    domain_id INT NOT NULL
        REFERENCES shared.domains_id_recipientable(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE,
    name VARCHAR(256) NOT NULL, -- Local part of the mail address (SQL pattern matching strings allowed)
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    UNIQUE(remote_id, domain_id, name)
);

CREATE INDEX idx_remotes_send_grants_remote_id ON remotes_send_grants(remote_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_remotes_send_grants_domain_id ON remotes_send_grants(domain_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_remotes_send_grants_name ON remotes_send_grants(name) WHERE deleted_at IS NULL;

CREATE TRIGGER trigger_updated_at
    BEFORE UPDATE ON remotes_send_grants
    FOR EACH ROW
    EXECUTE FUNCTION hook_update_updated_at();

CREATE TRIGGER trigger_audit
    AFTER INSERT OR UPDATE OR DELETE ON remotes_send_grants
    FOR EACH ROW
    EXECUTE FUNCTION hook_audit();

CREATE TRIGGER trigger_check_foreign_key_soft_delete
    BEFORE INSERT OR UPDATE ON remotes_send_grants
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_foreign_key_soft_delete('domain_id', 'shared.domains_id_recipientable');
