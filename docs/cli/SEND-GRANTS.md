# Send Grants

Manage permissions for remotes to send email as specific addresses. Send grants control which email addresses or domains a remote SMTP server is allowed to send as.

The grant supports the wildcards `%` (string) and `_` (single character) for grants. For example, granting `%.example.com` allows sending as any subdomain of `example.com`. Use `\` to escape wildcard characters. In a shell you have to escape the backslash itself, so usually you have to write `\\%`.

## Available Actions
- [`list`](#list) - List all send grants in a table or output as JSON
- [`create`](#create) - Create a new send grant
- [`delete`](#delete)/[`restore`](#restore) - Delete or restore a send grant

## List
Shows a table of all send grants or outputs them as JSON. Can be filtered by remote name.

### Usage
```sh
mailctl list send-grants [name...] [flags]
```

### Flags
- `-v`, `--verbose` - Show detailed information with timestamps
- `-j`, `--json` - Output in JSON format
- `-a`, `--all` - Include soft-deleted objects in the output
- `-d`, `--deleted` - Show only soft-deleted objects

## Create
Creates a new send grant, allowing a remote to send as a specific email address or domain.

### Usage
```sh
mailctl create send-grant <remote> <email|domain> [flags]
```

### Examples
```sh
# Allow remote to send as specific address
mailctl create send-grant smtp.relay.com user@example.com

# Allow remote to send as entire domain
mailctl create send-grant smtp.relay.com example.com
```

## Delete
Soft-deletes a send grant. The send grant can be restored later. Use `--permanent` to permanently delete it.

### Usage
```sh
mailctl delete send-grant <remote> <email|domain> [flags]
```

### Flags
- `-f`, `--force` - Soft-delete the send grant, even if it is already (updates the deletion timestamp)
- `-p`, `--permanent` - Permanently delete the send grant

## Restore
Restores a soft-deleted send grant.

### Usage
```sh
mailctl restore send-grant <remote> <email|domain>
```
