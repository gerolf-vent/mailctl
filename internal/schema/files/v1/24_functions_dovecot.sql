/***************************************************************
 * Dovecot shorthand functions
 *
 * @author Gerolf Vent <dev@gerolfvent.de>
 ***************************************************************/

/**
 * Dovecot: UserDB lookup function for mailboxes.
 *
 * @param $1 domain name
 * @param $2 user name
 */
CREATE FUNCTION dovecot.userdb_mailboxes(VARCHAR(256), VARCHAR(256)) RETURNS TABLE(quota_storage_size VARCHAR(256)) AS $$
    SELECT
        CASE
            WHEN m.storage_quota IS NULL THEN
                '0'
            ELSE
                CONCAT(m.storage_quota, 'M')
        END AS quota_storage_size
    FROM mailboxes m
    WHERE
        m.domain_id = (
            SELECT ID FROM domains_managed
            WHERE
                fqdn = $1 AND
                -- Don't check for enabled here for consistency with passdb
                deleted_at IS NULL
        ) AND
        m.name = $2 AND
        -- Don't check for login_enabled, because thats handled when authenticating (passdb)
        m.deleted_at IS NULL
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Dovecot: PassDB lookup function for mailboxes.
 *
 * @param $1 domain name
 * @param $2 user name
 */
CREATE FUNCTION dovecot.passdb_mailboxes(VARCHAR(256), VARCHAR(256)) RETURNS TABLE(password VARCHAR(1024), nologin BOOLEAN, reason VARCHAR(256)) AS $$
    SELECT
        CASE
            WHEN m.login_enabled IS false OR dm.enabled IS false THEN
                NULL
            ELSE
                m.password_hash
        END AS password,
        CASE
            WHEN m.login_enabled IS false OR dm.enabled IS false THEN
                true
            WHEN m.password_hash IS NULL THEN
                true
            ELSE
                NULL
        END AS nologin,
        CASE
            WHEN m.login_enabled IS false OR dm.enabled IS false THEN
                'Login is disabled.'
            WHEN m.password_hash IS NULL THEN
                'No password set.'
            ELSE
                NULL
        END AS reason
    FROM mailboxes m
    JOIN domains_managed dm ON dm.ID = m.domain_id
    WHERE
        dm.fqdn = $1 AND
        dm.deleted_at IS NULL AND
        m.name = $2 AND
        m.deleted_at IS NULL
$$ LANGUAGE SQL SECURITY DEFINER;

/**
 * Dovecot: PassDB lookup function for remotes.
 *
 * @param $1 login name
 */
CREATE FUNCTION dovecot.passdb_remotes(VARCHAR(256)) RETURNS TABLE(password VARCHAR(1024), nologin BOOLEAN, reason VARCHAR(256)) AS $$
    SELECT
        CASE
            WHEN r.enabled IS false THEN
                NULL
            ELSE
                r.password_hash
        END AS password,
        CASE
            WHEN r.enabled IS false THEN
                true
            WHEN r.password_hash IS NULL THEN
                true
            ELSE
                NULL
        END AS nologin,
        CASE
            WHEN r.enabled IS false THEN
                'Remote is disabled.'
            WHEN r.password_hash IS NULL THEN
                'No password set.'
            ELSE
                NULL
        END AS reason
    FROM remotes r
    WHERE
        r.name = $1 AND
        r.deleted_at IS NULL
$$ LANGUAGE SQL SECURITY DEFINER;
