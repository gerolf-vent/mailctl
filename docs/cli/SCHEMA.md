# Schema Management

Database schema management commands for creating, upgrading, and managing the mailctl database.

## Available Actions
- [`status`](#status) - Show current database schema version
- [`upgrade`](#upgrade) - Upgrade database schema to latest version
- [`purge`](#purge) - Permanently delete all data and schema
- [`ensure-user`](#ensure-user) - Create and sync application-specific database users
- [`drop-user`](#drop-user) - Drop application-specific database user

## Status
Shows the current database schema version and whether migrations are pending.

### Usage
```sh
mailctl schema status
```

## Upgrade
Upgrades the database schema to the latest version by applying pending migrations.

### Usage
```sh
mailctl schema upgrade
```

## Purge
Permanently deletes all content from the database and removes the schema.

>[!WARNING]
> This action is irreversible and will delete all data. You have to use the `--confirm` flag to actually perform the purge.

### Usage
```sh
mailctl schema purge --confirm
```

### Flags
- `--confirm` - Required flag to confirm the purge operation

## Ensure User
Creates and syncs an application-specific user with password and limited permissions in the database.

There are four types of users:
| Type | Description |
| ---- | ----------- |
| `manager` | User that can alter the database state (create/patch/delete domains, mailboxes, aliases, etc. but not the schema) |
| `postfix` | User for read-only access to the postfix functions |
| `dovecot` | User for read-only access to the dovecot functions |
| `stalwart` | User for read-only access to the stalwart functions |

You can provide the username, password and type in various ways:

1. As command arguments:
   ```sh
   mailctl schema ensure-user <username> -p --type <type>
   ```
   
   One of `-p`/`--password` or `--password-stdin` must be provided, otherwise the command will fail.
2. From files by specifying any of the following flags:
   - `--type-file` - Read user type from file
   - `--name-file` - Read username from file
   - `--password-file` - Read password from file
3. Via the environment variables:
   - `MAILCTL_USER_TYPE`
   - `MAILCTL_USER_NAME`
   - `MAILCTL_USER_PASSWORD`
   
   You can change the prefix `MAILCTL_USER` using `--env-prefix`.

If any of the required information (type, username, password) is missing, the command will fail. Multiple sources can be combined, e.g., type from file, username from argument and password from stdin. The sources have the precedence as listed above.

### Usage
```sh
mailctl schema ensure-user [username] [flags]
```

### Flags
- `-t`, `--type string` - User type: 'manager', 'postfix', 'dovecot' or 'stalwart'
- `-p`, `--password` - Set password interactively (prompts)
- `--password-stdin` - Read password from stdin
- `--type-file string` - Read user type from file
- `--name-file string` - Read username from file
- `--password-file string` - Read password from file
- `--env-prefix string` - Prefix for environment variables (default: "MAILCTL_USER")

### Examples
```sh
# Create a user that can alter the database state
mailctl schema ensure-user mailctl_manager --type manager

# Create a user for read-only access to the postfix functions
echo "verylongpassword" | mailctl schema ensure-user --type postfix --env-prefix MAILCTL_POSTFIX_USER --password-stdin

# Create a user with type/username from file and password from stdin
echo "dovecotpassword" | mailctl schema ensure-user --type-file /path/to/typefile --name-file /path/to/usernamefile --password-stdin
```

## Drop user
Drops an application-specific database user.

The flags and arguments behave the same way as in `ensure-user`, but without the type or password options.

### Usage
```sh
mailctl schema drop-user [username] [flags]
```

### Flags
- `--name-file string` - Read username from file
- `--env-prefix string` - Prefix for environment variable (default: "MAILCTL_USER")
