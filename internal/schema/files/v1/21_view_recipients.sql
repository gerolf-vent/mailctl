/***************************************************************
 * View for all recipients
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

CREATE VIEW recipients AS 
    SELECT
        ID,
        domain_id,
        name,
        receiving_enabled,
        'mailbox' AS type,
        created_at,
        updated_at,
        deleted_at
    FROM mailboxes
UNION ALL
    SELECT
        ID,
        domain_id,
        name,
        enabled AS receiving_enabled,
        'alias' AS type,
        created_at,
        updated_at,
        deleted_at
    FROM aliases
UNION ALL
    SELECT
        ID,
        domain_id,
        name,
        enabled AS receiving_enabled,
        'relayed' AS type,
        created_at,
        updated_at,
        deleted_at
    FROM recipients_relayed;
