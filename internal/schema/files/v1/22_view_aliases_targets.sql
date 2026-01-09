/***************************************************************
 * View for all alias targets
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

CREATE VIEW aliases_targets AS
    SELECT
        art.ID,
        art.alias_id,
        d.fqdn,
        r.name,
        art.forwarding_to_target_enabled,
        art.sending_from_target_enabled,
        'recursive' AS type,
        art.created_at,
        art.updated_at,
        art.deleted_at
    FROM aliases_targets_recursive art
    JOIN recipients r ON art.recipient_id = r.ID
    JOIN domains d ON r.domain_id = d.ID
UNION ALL
    SELECT
        ID,
        alias_id,
        fqdn,
        name,
        forwarding_to_target_enabled,
        NULL AS sending_from_target_enabled,
        'foreign' AS type,
        created_at,
        updated_at,
        deleted_at
    FROM aliases_targets_foreign;
