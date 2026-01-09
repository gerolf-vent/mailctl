/***************************************************************
 * Stalwart shorthand functions
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

/**
 * Stalwart: name lookup function.
 * The full email address is used as the name to prevent collisions between
 * different domains.
 *
 * @param $1 name of mailbox (full email address)
 */
CREATE FUNCTION stalwart.name(TEXT) RETURNS TABLE(name TEXT, type TEXT, email TEXT, secret TEXT, description TEXT, quota TEXT) AS $$
    SELECT
        CONCAT(m.name, '@', dm.fqdn) AS name, -- Use full email as name
        'individual' AS type,  -- Required for mailboxes
        CONCAT(m.name, '@', dm.fqdn) AS email,  -- There is only one email per mailbox
        m.password_hash AS secret,
        '' AS description,
        (m.storage_quota * 1024 * 1024) AS quota  -- Quota is expected in bytes
    FROM mailboxes m
    JOIN domains_managed dm ON m.domain_id = dm.ID
    WHERE
        CONCAT(m.name, '@', dm.fqdn) = $1 AND
        dm.enabled = true AND
        dm.deleted_at IS NULL AND
        m.receiving_enabled = true AND
        m.deleted_at IS NULL
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Stalwart: members lookup function.
 * The groups feature is not supported. This function is defined to
 * may implement it in the future.
 *
 * @param $1 name of mailbox (full email address)
 */
CREATE FUNCTION stalwart.members(TEXT) RETURNS TABLE(member_of TEXT) AS $$
    SELECT '' AS member_of WHERE false
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Stalwart: recipients lookup function.
 * The lookup checks if the mailbox exists and if so, returns the email
 * address as name.
 *
 * @param $1 full email address (for mail delivery)
 */
CREATE FUNCTION stalwart.recipients(TEXT) RETURNS TABLE(email TEXT) AS $$
    SELECT
        CONCAT(m.name, '@', dm.fqdn) AS email
    FROM mailboxes m
    JOIN domains_managed dm ON m.domain_id = dm.ID
    WHERE
        CONCAT(m.name, '@', dm.fqdn) = $1 AND
        dm.enabled = true AND
        dm.deleted_at IS NULL AND
        m.receiving_enabled = true AND
        m.deleted_at IS NULL
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Stalwart: emails lookup function.
 * As the name of a mailbox is the full email address, this function
 * behaves the same as the recipients function.
 *
 * @param $1 full email address (for mail delivery)
 */
CREATE FUNCTION stalwart.emails(TEXT) RETURNS TABLE(address TEXT) AS $$
    SELECT
        CONCAT(m.name, '@', dm.fqdn) AS address
    FROM mailboxes m
    JOIN domains_managed dm ON m.domain_id = dm.ID
    WHERE
        CONCAT(m.name, '@', dm.fqdn) = $1 AND
        dm.enabled = true AND
        dm.deleted_at IS NULL AND
        m.receiving_enabled = true AND
        m.deleted_at IS NULL
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Stalwart: secrets lookup function.
 *
 * @param $1 name of mailbox (full email address)
 */
CREATE FUNCTION stalwart.secrets(TEXT) RETURNS TABLE(secret TEXT) AS $$
    SELECT
        m.password_hash AS secret
    FROM mailboxes m
    JOIN domains_managed dm ON m.domain_id = dm.ID
    WHERE
        CONCAT(m.name, '@', dm.fqdn) = $1 AND
        dm.enabled = true AND
        dm.deleted_at IS NULL AND
        m.receiving_enabled = true AND
        m.deleted_at IS NULL
$$ LANGUAGE SQL SECURITY DEFINER;
