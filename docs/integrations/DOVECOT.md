# Dovecot Integration (v2.4+)
There are native SQL functions implemented in the database schema which cover UserDB and PassDB lookups for Dovecot. You can use the Dovecot PostgreSQL configuration to query these functions as described below.
To improve security, it's recommended to create a dedicated PostgreSQL user with limited permissions for Dovecot. This can be done via `mailctl schema ensure-user --type dovecot`.

## Prerequisites
- Dovecot v2.4+ with PostgreSQL support enabled
- Dedicated PostgreSQL user for Dovecot (optional but highly recommended)

## Available functions

| Dovecot Lookup | SQL Function | Description |
|----------------|--------------|-------------|
| UserDB (Mailboxes) | `dovecot.userdb_mailboxes('%{user\|domain}', '%{user\|username}')` | Returns quota information for mailboxes |
| PassDB (Mailboxes) | `dovecot.passdb_mailboxes('%{user\|domain}', '%{user\|username}')` | Returns password hash and login status |
| PassDB (Remotes) | `dovecot.passdb_remotes('%{user}')` | Returns password hash and login status |

## Configuration
> [!NOTE]
> In Dovecot v2.4, the SQL configuration is defined directly in `dovecot.conf` (or included files). The separate `dovecot-sql.conf.ext` file is no longer used.

Here is an example configuration snippet for `dovecot.conf`:

```dovecot
# Define the SQL driver and connection details
sql_driver = pgsql

# Configure the connection (replace values with your actual DB details)
pgsql mailctl_db {
  parameters {
    host = db.example.com
    port = 5432
    dbname = mailctl
    user = mailctl_dovecot
    password = secureandlongpassword123456789
  }
}

# UserDB configuration for mailboxes
userdb sql {
  query = SELECT * FROM dovecot.userdb_mailboxes('%{user|domain}', '%{user|username}')
}

# PassDB configuration for mailboxes
passdb sql {
  query = \
    SELECT \
      password, \
      nologin, \
      reason \
    FROM dovecot.passdb_mailboxes('%{user|domain}', '%{user|username}')
}
```

### Remotes Authentication
If you also want to authenticate "Remotes" (e.g. for relaying via SMTP), you can add a separate PassDB block or use a combined query. Since Remotes don't necessarily have a domain part in the same way mailboxes do, the query differs.

For Remotes, the query would be:
```dovecot
passdb sql_remotes {
  query = \
    SELECT \
      password, \
      nologin, \
      reason \
    FROM dovecot.passdb_remotes('%{user}')
}
```

## Notes
- The `default_pass_scheme` is now handled differently or defaults to detecting the scheme from the hash (e.g. `{CRYPT}`). `mailctl` uses Argon2id with the `{CRYPT}` prefix, which Dovecot supports.
- Ensure that the `mailctl_dovecot` user has `USAGE` on the `dovecot` schema and `EXECUTE` permissions on the functions. The `mailctl schema ensure-user` command handles this for you.
