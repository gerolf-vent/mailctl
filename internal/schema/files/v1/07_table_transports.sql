/***************************************************************
 * Table for transports
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

CREATE TABLE transports (
    ID SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL UNIQUE,  -- Simplifies referencing a specific target
    method VARCHAR(32) NOT NULL,
    host VARCHAR(256) NOT NULL,
    port SMALLINT CHECK (port > 0 AND port <= 65535) DEFAULT 25,
    mx_lookup BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE TRIGGER trigger_updated_at
    BEFORE UPDATE ON transports
    FOR EACH ROW
    EXECUTE FUNCTION hook_update_updated_at();

CREATE TRIGGER trigger_audit
    AFTER INSERT OR UPDATE OR DELETE ON transports
    FOR EACH ROW
    EXECUTE FUNCTION hook_audit();

-- Ensures that transports can't be deleted if they are in use by domains (which require a transport)
CREATE TRIGGER trigger_prohibit_delete_in_use_domains_managed
    BEFORE UPDATE OR DELETE ON transports
    FOR EACH ROW
    EXECUTE FUNCTION hook_prohibit_delete_in_use('domains_managed', 'transport_id');

CREATE TRIGGER trigger_prohibit_delete_in_use_domains_relayed
    BEFORE UPDATE OR DELETE ON transports
    FOR EACH ROW
    EXECUTE FUNCTION hook_prohibit_delete_in_use('domains_relayed', 'transport_id');
