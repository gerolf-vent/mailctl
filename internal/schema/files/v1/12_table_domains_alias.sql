/***************************************************************
 * Table for alias domains
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

CREATE TABLE domains_alias (
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
    enabled BOOLEAN NOT NULL DEFAULT(true),  -- Whether this alias domain is active
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_domains_alias_fqdn ON domains_alias(fqdn) WHERE deleted_at IS NULL;

CREATE TRIGGER trigger_updated_at
    BEFORE UPDATE ON domains_alias
    FOR EACH ROW
    EXECUTE FUNCTION hook_update_updated_at();

CREATE TRIGGER trigger_audit
    AFTER INSERT OR UPDATE OR DELETE ON domains_alias
    FOR EACH ROW
    EXECUTE FUNCTION hook_audit();

CREATE TRIGGER trigger_register_shared_id
    BEFORE INSERT OR UPDATE OR DELETE ON domains_alias
    FOR EACH ROW
    EXECUTE FUNCTION hook_register_shared_id('domains_id');

CREATE TRIGGER trigger_register_shared_id_recipientable
    BEFORE INSERT OR UPDATE OR DELETE ON domains_alias
    FOR EACH ROW
    EXECUTE FUNCTION hook_register_shared_id('domains_id_recipientable');

CREATE TRIGGER trigger_check_fqdn
    BEFORE INSERT OR UPDATE OR DELETE ON domains_alias
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_domains_uniq('fqdn');

CREATE TRIGGER trigger_cascade_soft_delete_domains_canonical
    AFTER UPDATE ON domains_alias
    FOR EACH ROW
    EXECUTE FUNCTION hook_cascade_soft_delete('domains_canonical', 'target_domain_id');
