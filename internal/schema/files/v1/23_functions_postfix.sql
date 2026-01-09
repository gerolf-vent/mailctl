/***************************************************************
 * Postfix shorthand functions
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

/**
 * Postfix: Looks up an active domain by its FQDN.
 *
 * This also resolves alias domains to the real (managed or relayed) domain.
 * Use `SELECT ID FROM postfix.domain_by_fqdn($1)` to get the domain id for a given
 * FQDN.
 *
 * @param $1 domain fqdn
 */
CREATE FUNCTION postfix.domain_by_fqdn(VARCHAR(256)) RETURNS TABLE(ID INT, type VARCHAR(16), enabled BOOLEAN, deleted_at TIMESTAMPTZ) AS $$
    SELECT ID, type, enabled, deleted_at FROM
        (
            -- Check managed domains
            SELECT
                dm.ID,
                'managed' AS type,
                dm.enabled,
                dm.deleted_at
            FROM domains_managed dm
            WHERE
                dm.fqdn = $1
            LIMIT 1
        )
        UNION ALL
        (
            -- Check relayed domains
            SELECT
                dr.ID,
                'relayed' AS type,
                dr.enabled,
                dr.deleted_at
            FROM domains_relayed dr
            WHERE
                dr.fqdn = $1
            LIMIT 1
        )
        UNION ALL
        (
            -- Check alias domains
            SELECT
                da.ID,
                'alias' AS type,
                da.enabled,
                da.deleted_at
            FROM domains_alias da
            WHERE
                da.fqdn = $1
            LIMIT 1
        )
        UNION ALL
        (
            -- Check canoncical domains
            SELECT
                dc.ID,
                'canonical' AS type,
                dc.enabled,
                dc.deleted_at
            FROM domains_canonical dc
            WHERE
                dc.fqdn = $1
            LIMIT 1
        )
    LIMIT 1
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Postfix: Creates a string representation of a transport.
 *
 * @param $1 method
 * @param $2 host
 * @param $3 port (nullable)
 * @param $4 mx_lookup (boolean)
 */
CREATE FUNCTION postfix.transport_string(VARCHAR(32), VARCHAR(256), SMALLINT, BOOLEAN) RETURNS VARCHAR(512) AS $$
    SELECT
        CASE
            WHEN $4 = false AND $3 IS NULL THEN
                $1 || ':[' || $2 || ']'
            WHEN $4 = false AND $3 IS NOT NULL THEN
                $1 || ':[' || $2 || ']:' || $3::VARCHAR
            WHEN $3 IS NOT NULL THEN
                $1 || ':' || $2 || ':' || $3::VARCHAR
            ELSE
                $1 || ':' || $2
        END AS result
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Postfix: Looks up the transport for a recipient thats handled by this system.
 *
 * @param $1 recipient domain fqdn
 * @param $2 recipient name
 */
CREATE FUNCTION postfix.transport_maps(VARCHAR(256), VARCHAR(256)) RETURNS TABLE(result VARCHAR(512)) AS $$
    SELECT
        postfix.transport_string(t.method, t.host, t.port, t.mx_lookup) AS result
    FROM (
        SELECT COALESCE(
            (
                -- First check for mailbox transport
                SELECT COALESCE(m.transport_id, dm.transport_id)
                FROM mailboxes m
                JOIN domains_managed dm ON m.domain_id = dm.ID
                WHERE
                    dm.fqdn = $1 AND
                    dm.enabled = true AND
                    dm.deleted_at IS NULL AND
                    m.name = $2 AND
                    m.receiving_enabled = true AND
                    m.deleted_at IS NULL
            ),
            (
                -- Then check for relayed recipient transport
                SELECT dr.transport_id
                FROM recipients_relayed rr
                JOIN domains_relayed dr ON rr.domain_id = dr.ID
                WHERE
                    dr.fqdn = $1 AND
                    dr.enabled = true AND
                    dr.deleted_at IS NULL AND
                    rr.name = $2 AND
                    rr.enabled = true AND
                    rr.deleted_at IS NULL
            )
        ) AS transport_id
    ) AS r
    JOIN transports t ON r.transport_id = t.ID
    WHERE t.deleted_at IS NULL
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Postfix: Looks up whether a domain contains aliases only.
 *
 * @param $1 domain fqdn
 */
CREATE FUNCTION postfix.virtual_alias_domains(VARCHAR(256)) RETURNS TABLE(result VARCHAR(4)) AS $$
    SELECT 'OK' AS result
    FROM domains_alias da
    WHERE
        da.enabled = true AND
        da.fqdn = $1 AND
        da.deleted_at IS NULL
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Postfix: Looks up all target mail addresses for a given alias mail address.
 * The query will be empty if the provided mail address is not an alias.
 *
 * @param $1 domain part of the alias address
 * @param $2 name part of the alias address
 * @param $3 Maximum recursion depth
 */
CREATE FUNCTION postfix.virtual_alias_maps(VARCHAR(256), VARCHAR(256), INT) RETURNS TABLE(result VARCHAR(513)) AS $$
    WITH RECURSIVE
        -- Build recursive alias target chain
        aliases_targets_chain AS (
            -- Base case: Initial match for the provided alias address + catch-all aliases
            SELECT alias_id, recipient_id, fallback_only, is_catchall, 0 as depth
            FROM (
                -- Include the alias itself (for foreign targets lookup)
                SELECT a.ID AS alias_id, NULL::INT AS recipient_id, false AS fallback_only, false AS is_catchall
                FROM aliases a
                WHERE
                    a.domain_id IN (
                        SELECT ID
                        FROM postfix.domain_by_fqdn($1)
                        WHERE
                            enabled = true AND
                            deleted_at IS NULL
                    ) AND
                    a.name = $2 AND
                    a.enabled = true AND
                    a.deleted_at IS NULL

                UNION ALL

                -- Include matches for explicit aliases (recursive targets)
                SELECT atr.alias_id, atr.recipient_id, false AS fallback_only, false AS is_catchall
                FROM aliases_targets_recursive atr
                WHERE
                    atr.alias_id IN (
                        SELECT a.ID
                        FROM aliases a
                        WHERE
                            a.domain_id IN (
                                SELECT ID
                                FROM postfix.domain_by_fqdn($1)
                                WHERE
                                    enabled = true AND
                                    deleted_at IS NULL
                            ) AND
                            a.name = $2 AND
                            a.enabled = true AND
                            a.deleted_at IS NULL
                    ) AND
                    atr.forwarding_to_target_enabled = true AND
                    atr.deleted_at IS NULL

                UNION ALL

                -- Include matches for catch-all aliases in the same domain
                SELECT NULL AS alias_id, dct.recipient_id, dct.fallback_only AS fallback_only, true AS is_catchall
                FROM domains_catchall_targets dct
                WHERE
                    dct.domain_id IN (
                        SELECT ID
                        FROM postfix.domain_by_fqdn($1)
                        WHERE
                            enabled = true AND
                            deleted_at IS NULL
                    ) AND
                    dct.forwarding_to_target_enabled = true AND
                    dct.deleted_at IS NULL
            )

            UNION ALL

            -- Recursive over the alias targets (check for recipients that are aliases themselves)
            SELECT atr.alias_id, atr.recipient_id, atc.fallback_only, atc.is_catchall, atc.depth + 1 AS depth
            FROM aliases_targets_recursive atr
            JOIN aliases_targets_chain atc ON atc.recipient_id = atr.alias_id
            JOIN aliases a ON a.ID = atr.alias_id
            LEFT JOIN domains_managed dm ON dm.ID = a.domain_id
            LEFT JOIN domains_relayed dr ON dr.ID = a.domain_id
            LEFT JOIN domains_alias da ON da.ID = a.domain_id
            WHERE
                atr.forwarding_to_target_enabled = true AND
                atr.deleted_at IS NULL AND
                a.enabled = true AND
                a.deleted_at IS NULL AND
                COALESCE(dm.enabled, dr.enabled, da.enabled) = true AND
                COALESCE(dm.deleted_at, dr.deleted_at, da.deleted_at) IS NULL AND
                atc.depth < $3  -- Prevent infinite recursion
        ) CYCLE alias_id SET is_cycle USING path,

        -- Collect all recipients from the built chain
        recipients AS (
            (
                -- Collect mailbox addresses
                SELECT m.name, dm.fqdn, atc.fallback_only, atc.is_catchall
                FROM aliases_targets_chain atc
                JOIN mailboxes m ON m.ID = atc.recipient_id
                JOIN domains_managed dm ON m.domain_id = dm.ID
                WHERE
                    dm.enabled = true AND
                    dm.deleted_at IS NULL AND
                    m.receiving_enabled = true AND
                    m.deleted_at IS NULL
            )
            UNION ALL
            (
                -- Collect relayed recipient addresses
                SELECT rr.name, dr.fqdn, atc.fallback_only, atc.is_catchall
                FROM aliases_targets_chain atc
                JOIN recipients_relayed rr ON rr.ID = atc.recipient_id
                JOIN domains_relayed dr ON rr.domain_id = dr.ID
                WHERE
                    dr.enabled = true AND
                    dr.deleted_at IS NULL AND
                    rr.enabled = true AND
                    rr.deleted_at IS NULL
            )
            UNION ALL
            (
                -- Collect foreign target addresses
                SELECT ft.name, ft.fqdn, atc.fallback_only, atc.is_catchall
                FROM aliases_targets_chain atc
                JOIN aliases_targets_foreign ft ON ft.alias_id = atc.alias_id
                WHERE
                    ft.forwarding_to_target_enabled = true AND
                    ft.deleted_at IS NULL
            )
        ),

        stats AS (
            SELECT EXISTS (SELECT 1 FROM recipients r WHERE r.is_catchall = false) AS has_non_catchall
        )

    -- Apply fallback-only selection
    SELECT DISTINCT CONCAT(r.name, '@', r.fqdn)::VARCHAR(513) AS result
    FROM recipients r, stats s
    WHERE
        r.is_catchall = false OR  -- Always include non-catchall targets
        r.fallback_only = false OR  -- Include catch-all targets that are not fallback-only
        (r.fallback_only = true AND s.has_non_catchall = false)  -- Include fallback-only catch-all targets only if no non-catchall targets exist
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Postfix: Looks up whether a domain contains mailboxes.
 *
 * @param $1 domain fqdn
 */
CREATE FUNCTION postfix.virtual_mailbox_domains(VARCHAR(256)) RETURNS TABLE(result VARCHAR(4)) AS $$
    SELECT 'OK' AS result
    FROM domains_managed
    WHERE
        enabled = true AND
        fqdn = $1 AND
        deleted_at IS NULL
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Postfix: Looks up wether the given address is a mailbox.
 *
 * @param $1 domain fqdn
 * @param $2 mailbox name
 */
CREATE FUNCTION postfix.virtual_mailbox_maps(VARCHAR(256), VARCHAR(256)) RETURNS TABLE(result VARCHAR(4)) AS $$
    SELECT 'OK' AS result
    FROM mailboxes m
    WHERE
        m.domain_id IN (
            SELECT ID
            FROM domains_managed
            WHERE
                fqdn = $1 AND
                enabled = true AND
                deleted_at IS NULL
        ) AND
        m.name = $2 AND
        m.receiving_enabled = true AND
        m.deleted_at IS NULL
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Postfix: Looks up all remote SASL login names of mailboxes that are allowed to send from a given mail address.
 *
 * @param $1 domain part of the sender address
 * @param $2 name part of the sender address
 * @param $3 maximum recursion depth
 */
CREATE FUNCTION postfix.smtpd_sender_login_maps_mailboxes(VARCHAR(256), VARCHAR(256), INT) RETURNS TABLE(result VARCHAR(513)) AS $$
    WITH RECURSIVE
        -- Build recursive alias target chain
        aliases_targets_chain AS (
            SELECT alias_id, recipient_id, 0 as depth
            FROM aliases_targets_recursive
            WHERE
                alias_id IN (
                    SELECT ID
                    FROM aliases
                    WHERE
                        -- Match the provided domain
                        domain_id IN (
                            SELECT ID
                            FROM postfix.domain_by_fqdn($1)
                            WHERE
                                enabled = true AND
                                deleted_at IS NULL
                        ) AND
                        -- Match the provided alias address
                        name = $2 AND
                        enabled = true AND
                        deleted_at IS NULL
                ) AND
                sending_from_target_enabled = true AND
                deleted_at IS NULL

            UNION ALL

            -- Recursive over the alias targets (check for recipients that are aliases themselves)
            SELECT atr.alias_id, atr.recipient_id, atc.depth + 1
            FROM aliases_targets_recursive atr
            JOIN aliases_targets_chain atc ON atc.recipient_id = atr.alias_id
            JOIN aliases a ON a.id = atr.alias_id
            LEFT JOIN domains_managed dm ON dm.ID = a.domain_id
            LEFT JOIN domains_relayed dr ON dr.ID = a.domain_id
            LEFT JOIN domains_alias da ON da.ID = a.domain_id
            WHERE
                atr.sending_from_target_enabled = true AND
                atr.deleted_at IS NULL AND
                a.enabled = true AND
                a.deleted_at IS NULL AND
                COALESCE(dm.enabled, dr.enabled, da.enabled) = true AND
                COALESCE(dm.deleted_at, dr.deleted_at, da.deleted_at) IS NULL AND
                atc.depth < $3  -- Prevent infinite recursion
        ) CYCLE alias_id SET is_cycle USING path,
    
        recipients AS (
            (
                -- Collect all mailboxes from the built chain
                SELECT CONCAT(m.name, '@', d.fqdn)::VARCHAR(513) AS result
                FROM aliases_targets_chain c
                JOIN mailboxes m ON m.ID = c.recipient_id
                JOIN domains_managed d ON m.domain_id = d.ID
                WHERE
                    d.enabled = true AND
                    d.deleted_at IS NULL AND
                    m.sending_enabled = true AND
                    /* Don't check for login_enabled, because thats handled when authenticating */
                    m.deleted_at IS NULL
            )
            UNION ALL
            (
            -- Direct match for mailbox (in case no alias is used)
            SELECT CONCAT(m.name, '@', dm.fqdn)::VARCHAR(513) AS result
            FROM mailboxes m
            JOIN domains_managed dm ON dm.ID = m.domain_id
            WHERE
                dm.fqdn = $1 AND
                dm.enabled = true AND
                dm.deleted_at IS NULL AND
                m.name = $2 AND
                m.sending_enabled = true AND
                /* Don't check for login_enabled, because thats handled when authenticating */
                m.deleted_at IS NULL
            )
        )

    SELECT DISTINCT r.result FROM recipients r
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Postfix: Looks up whether a domain is relayed by this system.
 *
 * @param $1 domain fqdn
 */
CREATE FUNCTION postfix.relay_domains(VARCHAR(256)) RETURNS TABLE(result VARCHAR(4)) AS $$
    SELECT 'OK' AS result
    FROM domains_relayed
    WHERE
        enabled = true AND
        fqdn = $1 AND
        deleted_at IS NULL
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Postfix: Looks up whether the given address is a recipient on a relayed server.
 *
 * @param $1 domain fqdn
 * @param $2 recipient name
 */
CREATE FUNCTION postfix.relay_recipient_maps(VARCHAR(256), VARCHAR(256)) RETURNS TABLE(result VARCHAR(4)) AS $$
    SELECT 'OK' AS result
    FROM recipients_relayed rr
    WHERE
        rr.domain_id IN (
            SELECT ID
            FROM domains_relayed
            WHERE
                fqdn = $1 AND
                enabled = true AND
                deleted_at IS NULL
        ) AND
        rr.name = $2 AND
        rr.enabled = true AND
        rr.deleted_at IS NULL
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Postfix: Looks up all SASL login names of remotes that are allowed to send from a given mail address.
 *
 * @param $1 domain part of sender email address
 * @param $2 local part of sender email address
 */
CREATE FUNCTION postfix.smtpd_sender_login_maps_remotes(VARCHAR(256), VARCHAR(256)) RETURNS TABLE(result VARCHAR(513)) AS $$
    SELECT r.name AS result
    FROM remotes_send_grants rsg
    JOIN remotes r ON r.ID = rsg.remote_id
    WHERE
        -- Check for matching domain and enabled status
        rsg.domain_id IN (
            SELECT ID
            FROM postfix.domain_by_fqdn($1)
            WHERE
                enabled = true AND
                deleted_at IS NULL
        ) AND
        r.enabled = true AND
        r.deleted_at IS NULL AND
        $2 LIKE rsg.name ESCAPE '\' AND
        rsg.deleted_at IS NULL
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Postfix: Looks up all (mailboxes and remotes) SASL login names that are allowed to send from a given mail address.
 *
 * @param $1 domain part of sender email address
 * @param $2 local part of sender email address
 * @param $3 maximum recursion depth
 */
CREATE FUNCTION postfix.smtpd_sender_login_maps(VARCHAR(256), VARCHAR(256), INT) RETURNS TABLE(result VARCHAR(513)) AS $$
    SELECT result FROM postfix.smtpd_sender_login_maps_mailboxes($1, $2, $3)
    UNION ALL
    SELECT result FROM postfix.smtpd_sender_login_maps_remotes($1, $2)
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Postfix: Canonicalize alias domains
 *
 * @param $1 recipient domain fqdn
 * @param $2 recipient name
 */
CREATE FUNCTION postfix.canonical_maps(VARCHAR(256), VARCHAR(256)) RETURNS TABLE(result VARCHAR(513)) AS $$
    SELECT CONCAT($2, '@', COALESCE(dm.fqdn, dr.fqdn, da.fqdn)) AS result
    FROM domains_canonical dc
    LEFT JOIN domains_managed dm ON dm.ID = dc.target_domain_id
    LEFT JOIN domains_relayed dr ON dr.ID = dc.target_domain_id
    LEFT JOIN domains_alias da ON da.ID = dc.target_domain_id
    WHERE
        dc.fqdn = $1 AND
        dc.enabled = true AND
        dc.deleted_at IS NULL AND
        COALESCE(dm.enabled, dr.enabled, da.enabled) = true AND
        COALESCE(dm.deleted_at, dr.deleted_at, da.deleted_at) IS NULL
    LIMIT 1
$$ LANGUAGE SQL SECURITY DEFINER;
