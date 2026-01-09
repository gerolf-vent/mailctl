# Stalwart Integration
There are native SQL functions in the `stalwart` schema that satisfy Stalwart's SQL directory queries. Point Stalwart to these functions to resolve accounts, secrets, and recipient checks directly from your `mailctl` database. For least privilege, create a dedicated database role via `mailctl schema ensure-user --type stalwart`.

> [!IMPORTANT]
> The SQL functions only support Stalwart as an MDA (mail delivery) not an MTA (mail routing and sending), so you still need Postfix for mail routing and allowing mailboxes to send mail.

## Prerequisites
- Stalwart Mail Server 0.13+ with PostgreSQL (SQL directory) enabled
- Dedicated PostgreSQL user for Postfix (optional but highly recommended)

## Available functions
| Stalwart lookup | SQL function | Description |
|-----------------|--------------|-------------|
| `name` | `stalwart.name($1)` | Returns account metadata (name, type, email, secret, description, quota in bytes) for a mailbox |
| `members` | `stalwart.members($1)` | Placeholder for future group membership (currently returns no rows) |
| `recipients` | `stalwart.recipients($1)` | Confirms a mailbox exists and is allowed to receive mail |
| `emails` | `stalwart.emails($1)` | Alias of `recipients`, resolves a mailbox email address |
| `secrets` | `stalwart.secrets($1)` | Returns the password hash for login (Argon2id with `{CRYPT}` prefix) |

All lookups expect the full email address as the single parameter.

## Configuration
Stalwart's SQL directory can be pointed at the `mailctl` database with prepared statements. Example (TOML-style) snippet:

```toml
# Select the SQL directory provider
directory.default = "sql"

[directory.sql]
driver = "postgres"
dsn = "postgres://mailctl_stalwart:secure-and-long-password@db.example.com:5432/mailctl?sslmode=require"
prepare = true

[directory.sql.query]
name = "SELECT name, type, email, secret, description, quota FROM stalwart.name($1)"
members = "SELECT member_of FROM stalwart.members($1)"
recipients = "SELECT email FROM stalwart.recipients($1)"
emails = "SELECT address FROM stalwart.emails($1)"
secrets = "SELECT secret FROM stalwart.secrets($1)"
```

Notes:
- Keep `prepare = true` so parameters are bound safely; `$1` is the placeholder for PostgreSQL.
- Ensure the `mailctl_stalwart` role (or whichever user you choose) has `USAGE` on the `stalwart` schema and `EXECUTE` on its functions. `mailctl schema ensure-user --type stalwart` grants these automatically.
- `quota` is returned in bytes; Stalwart expects this unit.

After updating the config, reload or restart Stalwart to apply the changes.
