# Postfix Integration
There are native SQL functions implemented in the database schema which cover various maps for Postfix. You can use the Postfix PostgreSQL table lookup to query these functions as described below.
To improve security, it's recommended to create a dedicated PostgreSQL user with limited permissions for Postfix. This can be done via `mailctl schema ensure-user --type postfix`.

## Prerequisites
- Postfix 3.6+ with PostgreSQL support enabled
- Dedicated PostgreSQL user for Postfix (optional but highly recommended)

## Available Functions
| Postfix Option | SQL Function | Description | Used by Feature | Notes |
|----------------|--------------|-------------|--------------|-------|
| `transport_maps` | `postfix.transport_maps(%d, %u)` | Nexthop for emails | mailboxes, relayed recipients | |
| `canonical_maps` | `postfix.canonical_maps(%d, %u)` | Domain canonicalization/rewrites | canonical domains | |
| `virtual_alias_domains` | `postfix.virtual_alias_domains(%s)` | Identifies aliases only domains | |
| `virtual_alias_maps` | `postfix.virtual_alias_maps(%d, %u, 10000)` | Resolve alias addresses | aliases | last parameter is maximum recursion depth |
| `virtual_mailbox_domains` | `postfix.virtual_mailbox_domains(%s)` | Identifies mailbox domains | managed domains, mailboxes | |
| `virtual_mailbox_maps` | `postfix.virtual_mailbox_maps(%d, %u)` | Resolve mailbox addresses | mailboxes | |
| `relay_domains` | `postfix.relay_domains(%s)` | Identifies relayed domains | relayed domains, relayed recipients | |
| `relay_recipient_maps` | `postfix.relay_recipient_maps(%d, %u)` | Resolve relayed recipient addresses | relayed recipients | |
| `smtpd_sender_login_maps` | `postfix.smtpd_sender_login_maps(%d, %u, 10000)` | Resolves **all** logins (mailboxes and remotes) which are allowed to send from an address | aliases | last parameter is maximum recursion depth |
| `smtpd_sender_login_maps` | `postfix.smtpd_sender_login_maps_mailboxes(%d, %u, 10000)` | Resolves **only mailbox** logins which are allowed to send from an address | aliases | last parameter is maximum recursion depth |
| `smtpd_sender_login_maps` | `postfix.smtpd_sender_login_maps_remotes(%d, %u)` | Resolves **only remote** logins which are allowed to send from an address | aliases | |

## Configuration
You have to create a Postfix map configuration file for each map you want to use. The configuration file should look like this:
```
hosts = db.example.com
port = 5432
user = mailctl_postfix
password = secureandlongpassword123456789
dbname = mailctl
query = SELECT result FROM postfix.transport_maps(%d, %u);
```
All Postfix functions are located in the `postfix` schema, so don't forget to prefix the function names with `postfix.` in the query. The function name is usually the same as the Postfix option name, but you can refer to the table above for exact names and parameters.

After creating the map configuration file, you can reference it in `main.cf` like this:
```
transport_maps = pgsql:/path/to/your/mapfile
```

Thats it! Postfix should now be able to use the database functions for mail routing and address resolution.
