/**
 * Dovecot: PassDB lookup function for mailboxes.
 *
 * @param $1 domain name
 * @param $2 user name
 */
CREATE FUNCTION dovecot.ensure_password_scheme(password_hash TEXT)
RETURNS TEXT AS $$
BEGIN
    -- If already has a scheme prefix (starts with {), return as-is
    IF password_hash LIKE '{%' THEN
        RETURN password_hash;
    END IF;
    
    -- Bcrypt hashes start with $2a$, $2b$, or $2y$
    IF password_hash ~ '^\$2[aby]\$' THEN
        RETURN '{BLF-CRYPT}' || password_hash;

    -- ARGON2ID hashes start with $argon2id$
    ELSIF password_hash ~ '^\$argon2id\$' THEN
        RETURN '{ARGON2ID}' || password_hash;

    END IF;
    
    RETURN password_hash;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

/**
 * Dovecot: PassDB lookup function for mailboxes.
 *
 * @version 2
 * @param $1 domain name
 * @param $2 user name
 */
CREATE OR REPLACE FUNCTION dovecot.passdb_mailboxes(VARCHAR(256), VARCHAR(256)) RETURNS TABLE(password VARCHAR(1024), nologin BOOLEAN, reason VARCHAR(256)) AS $$
    SELECT
        CASE
            WHEN m.login_enabled IS false OR dm.enabled IS false THEN
                NULL
            ELSE
                dovecot.ensure_password_scheme(m.password_hash)
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
 * @version 2
 * @param $1 login name
 */
CREATE OR REPLACE FUNCTION dovecot.passdb_remotes(VARCHAR(256)) RETURNS TABLE(password VARCHAR(1024), nologin BOOLEAN, reason VARCHAR(256)) AS $$
    SELECT
        CASE
            WHEN r.enabled IS false THEN
                NULL
            ELSE
                dovecot.ensure_password_scheme(r.password_hash)
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
