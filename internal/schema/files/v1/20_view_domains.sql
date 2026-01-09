/***************************************************************
 * View for all domains
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

CREATE VIEW domains AS
    SELECT
        'managed' AS type,
        ID,
        fqdn,
        transport_id,
        NULL::integer AS target_domain_id,
        enabled,
        created_at,
        updated_at,
        deleted_at
    FROM domains_managed
UNION ALL
    SELECT
        'relayed' AS type,
        ID,
        fqdn,
        transport_id,
        NULL::integer AS target_domain_id,
        enabled,
        created_at,
        updated_at,
        deleted_at
    FROM domains_relayed
UNION ALL
    SELECT
        'alias' AS type,
        ID,
        fqdn,
        NULL::integer AS transport_id,
        NULL::integer AS target_domain_id,
        enabled,
        created_at,
        updated_at,
        deleted_at
    FROM domains_alias
UNION ALL
    SELECT
        'canonical' AS type,
        ID,
        fqdn,
        NULL::integer AS transport_id,
        target_domain_id,
        enabled,
        created_at,
        updated_at,
        deleted_at
    FROM domains_canonical;
