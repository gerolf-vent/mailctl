/***************************************************************
 * Table for recursive alias targets
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

CREATE TABLE aliases_targets_recursive (
    ID INT PRIMARY KEY
        REFERENCES shared.aliases_targets_id(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE
        DEFAULT nextval('shared.aliases_targets_id_seq'),
    alias_id INT NOT NULL  -- Alias from which to forward emails
        REFERENCES aliases(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE,
    recipient_id INT NOT NULL  -- Target recipient to which to forward emails
        REFERENCES shared.recipients_id(ID)
            ON DELETE CASCADE
            ON UPDATE CASCADE,
    forwarding_to_target_enabled BOOLEAN NOT NULL DEFAULT(true),
    sending_from_target_enabled BOOLEAN NOT NULL DEFAULT(false),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    UNIQUE(alias_id, recipient_id)
);

CREATE INDEX idx_aliases_targets_recursive_alias_id ON aliases_targets_recursive(alias_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_aliases_targets_recursive_recipient_id ON aliases_targets_recursive(recipient_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_aliases_targets_recursive_both_ids ON aliases_targets_recursive(alias_id, recipient_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_aliases_targets_recursive_forwarding_enabled ON aliases_targets_recursive(forwarding_to_target_enabled) WHERE deleted_at IS NULL;
CREATE INDEX idx_aliases_targets_recursive_sending_enabled ON aliases_targets_recursive(sending_from_target_enabled) WHERE deleted_at IS NULL;

CREATE TRIGGER trigger_updated_at
    BEFORE UPDATE ON aliases_targets_recursive
    FOR EACH ROW
    EXECUTE FUNCTION hook_update_updated_at();

CREATE TRIGGER trigger_audit
    AFTER INSERT OR UPDATE OR DELETE ON aliases_targets_recursive
    FOR EACH ROW
    EXECUTE FUNCTION hook_audit();

-- Make sure the alias target id is globally unique
CREATE TRIGGER trigger_register_shared_id
    BEFORE INSERT OR UPDATE OR DELETE ON aliases_targets_recursive
    FOR EACH ROW
    EXECUTE FUNCTION hook_register_shared_id('aliases_targets_id');

CREATE TRIGGER trigger_check_foreign_key_soft_delete_alias
    BEFORE INSERT OR UPDATE ON aliases_targets_recursive
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_foreign_key_soft_delete('alias_id', 'aliases');

CREATE TRIGGER trigger_check_foreign_key_soft_delete_recipient
    BEFORE INSERT OR UPDATE ON aliases_targets_recursive
    FOR EACH ROW
    EXECUTE FUNCTION hook_check_foreign_key_soft_delete('recipient_id', 'shared.recipients_id');
