/***************************************************************
 * Table for remotes
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

CREATE TABLE remotes (
    ID SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL UNIQUE,
    password_hash VARCHAR(1024),  -- Hashed authentication data (e.g., bcrypt, argon2)
    enabled BOOLEAN NOT NULL DEFAULT(true),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_remotes_enabled ON remotes(enabled) WHERE deleted_at IS NULL;
CREATE INDEX idx_remotes_name_enabled ON remotes(name, enabled) WHERE deleted_at IS NULL;

CREATE TRIGGER trigger_updated_at
    BEFORE UPDATE ON remotes
    FOR EACH ROW
    EXECUTE FUNCTION hook_update_updated_at();

CREATE TRIGGER trigger_audit
    AFTER INSERT OR UPDATE OR DELETE ON remotes
    FOR EACH ROW
    EXECUTE FUNCTION hook_audit();

CREATE TRIGGER trigger_cascade_soft_delete
    AFTER UPDATE ON remotes
    FOR EACH ROW
    EXECUTE FUNCTION hook_cascade_soft_delete('remotes_send_grants', 'remote_id');
