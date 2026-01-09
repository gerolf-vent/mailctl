# CLI Reference

## Configuration
`mailctl` is configured completely by environment variables. The following variables are supported:
| Variable | Description                    | Default Value          |
| -------- | ------------------------------ | ---------------------- |
| `DB_HOST` | Database host                 | `localhost`            |
| `DB_PORT` | Database port                 | `5432`                 |
| `DB_NAME` | Database name                 | `mail`               |
| `DB_USER` | Database user                 | `mail`            |
| `DB_PASSWORD` | Database password         | (empty)                |
| `DB_SSLMODE` | SSL mode for database connection | `disable`          |
| `DB_TLSCACERT` | Path to CA certificate file for SSL verification | (empty) |

## Commands

### Database Objects
| Type                                 | Description                                           |
| ------------------------------------------- | ----------------------------------------------------- |
| [Domains](DOMAINS.md)                       | Mail domains (managed, relayed, alias, canonical)                |
| [Mailboxes](MAILBOXES.md)                   | User mailboxes and authentication                     |
| [Aliases](ALIASES.md)                       | Virtual aliases                                       |
| [Alias Targets](ALIAS-TARGETS.md)           | Recursive and external target addresses of aliases (including send-as)   |
| [Catchall Targets](CATCHALL-TARGETS.md)     | Catch-all target addresses for domains                |
| [Relayed Recipients](RELAYED-RECIPIENTS.md) | Individual recipients on relayed domains              |
| [Transports](TRANSPORTS.md)                 | Mail transport configurations                         |
| [Remotes](REMOTES.md)                       | Remote SMTP relay credentials                         |
| [Send Grants](SEND-GRANTS.md)               | Permissions for remotes to send as specific addresses |

Each object type supports a subset of the following actions:
- `list` - List all objects in a table or output as JSON
- `create` - Create a new object
- `patch` - Update an existing object
- `rename` - Change the identifier of an object (e.g., email address, domain name)
- `disable`/`enable` - Disable or enable an object
- `delete`/`restore` - Delete or restore an object

These precede the object type in the command. For example, to list all domains, use:
```sh
mailctl list domains
```

### Schema Management
Following actions are available:
- `status` - Show current schema version and applied migrations
- `upgrade` - Apply pending migrations to upgrade the schema
- `purge` - Remove all data and drop all tables (use with caution!)
- `ensure-user` - Create/sync application-specific database user with limited permissions
- `drop-user` - Drop application-specific database user

Unlike other commands, these actions are specified after `schema`. For example, to upgrade the schema, use:
```sh
mailctl schema upgrade
```

See [Schema](SCHEMA.md) for the full command reference.

## Tips & Tricks

### Shell Completion
```sh
# Generate completion script for your shell
mailctl completion bash > /etc/bash_completion.d/mailctl
mailctl completion zsh > ~/.zsh/completions/_mailctl
mailctl completion fish > ~/.config/fish/completions/mailctl.fish
mailctl completion powershell > mailctl.ps1
```
