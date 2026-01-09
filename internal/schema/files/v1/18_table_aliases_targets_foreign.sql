/***************************************************************
 * Table for foreign alias targets
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

CREATE TABLE aliases_targets_foreign (
    ID INT PRIMARY KEY
        REFERENCES shared.aliases_targets_id(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE
        DEFAULT nextval('shared.aliases_targets_id_seq'),
    alias_id INT NOT NULL  -- Alias from which to forward emails
        REFERENCES aliases(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE,
    fqdn VARCHAR(256) NOT NULL  -- Domain part of target mail address
        CHECK (check_domain_fqdn(fqdn)),
    name VARCHAR(256) NOT NULL  -- Name part of target mail address
        CHECK (check_mail_address_name(name)),
    forwarding_to_target_enabled BOOLEAN NOT NULL DEFAULT(true),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    UNIQUE (alias_id, fqdn, name)
);

CREATE INDEX idx_aliases_targets_foreign_fqdn_and_name ON aliases_targets_foreign(fqdn, name) WHERE deleted_at IS NULL;
CREATE INDEX idx_aliases_targets_foreign_alias_id ON aliases_targets_foreign(alias_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_aliases_targets_foreign_forwarding_enabled ON aliases_targets_foreign(forwarding_to_target_enabled) WHERE deleted_at IS NULL;

CREATE TRIGGER trigger_updated_at
    BEFORE UPDATE ON aliases_targets_foreign
    FOR EACH ROW
    EXECUTE FUNCTION hook_update_updated_at();

CREATE TRIGGER trigger_audit
    AFTER INSERT OR UPDATE OR DELETE ON aliases_targets_foreign
    FOR EACH ROW
    EXECUTE FUNCTION hook_audit();

-- Make sure the alias target id is globally unique
CREATE TRIGGER trigger_register_shared_id
    BEFORE INSERT OR UPDATE OR DELETE ON aliases_targets_foreign
    FOR EACH ROW
    EXECUTE FUNCTION hook_register_shared_id('aliases_targets_id');

CREATE TRIGGER trigger_check_foreign_key_soft_delete_alias
    BEFORE INSERT OR UPDATE ON aliases_targets_foreign
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_foreign_key_soft_delete('alias_id', 'aliases');

-- Ensures that foreign alias targets are actually foreign (and not exist in the database)
CREATE FUNCTION hook_check_foreign_alias_targets_domain()
RETURNS TRIGGER AS $$
BEGIN
    -- Ignore updates that do not change the fqdn value
    IF TG_OP = 'UPDATE' AND NEW.fqdn = OLD.fqdn THEN
        RETURN NEW;
    END IF;

    IF EXISTS (
        SELECT 1 FROM domains_managed WHERE fqdn = NEW.fqdn LIMIT 1
    ) OR EXISTS (
        SELECT 1 FROM domains_relayed WHERE fqdn = NEW.fqdn LIMIT 1
    ) OR EXISTS (
        SELECT 1 FROM domains_alias WHERE fqdn = NEW.fqdn LIMIT 1
    ) OR EXISTS (
        SELECT 1 FROM domains_canonical WHERE fqdn = NEW.fqdn LIMIT 1
    ) THEN
        RAISE EXCEPTION 'Domain "%" is not foreign (exists).', NEW.fqdn;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_check_foreign_alias_targets
    BEFORE INSERT OR UPDATE ON aliases_targets_foreign
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_foreign_alias_targets_domain();
