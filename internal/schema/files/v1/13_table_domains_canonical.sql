/***************************************************************
 * Table for canonical domains
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

CREATE TABLE domains_canonical (
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
    target_domain_id INT NOT NULL
        REFERENCES shared.domains_id_recipientable(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT(true),  -- Whether this canonical domain is active
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_domains_canonical_fqdn ON domains_canonical(fqdn) WHERE deleted_at IS NULL;
CREATE INDEX idx_domains_canonical_target_domain_id ON domains_canonical(target_domain_id) WHERE deleted_at IS NULL;

CREATE TRIGGER trigger_updated_at
    BEFORE UPDATE ON domains_canonical
    FOR EACH ROW
    EXECUTE FUNCTION hook_update_updated_at();

CREATE TRIGGER trigger_audit
    AFTER INSERT OR UPDATE OR DELETE ON domains_canonical
    FOR EACH ROW
    EXECUTE FUNCTION hook_audit();

CREATE TRIGGER trigger_register_shared_id
    BEFORE INSERT OR UPDATE OR DELETE ON domains_canonical
    FOR EACH ROW
    EXECUTE FUNCTION hook_register_shared_id('domains_id');

CREATE TRIGGER trigger_check_domains_uniq
    BEFORE INSERT OR UPDATE OR DELETE ON domains_canonical
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_domains_uniq('fqdn');

CREATE TRIGGER trigger_check_foreign_key_soft_delete
    BEFORE INSERT OR UPDATE ON domains_canonical
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_foreign_key_soft_delete('target_domain_id', 'shared.domains_id_recipientable');
