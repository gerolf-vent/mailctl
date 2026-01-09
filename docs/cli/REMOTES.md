# Remotes

Manage remote SMTP relay server credentials. Remotes define the connection details and authentication for external SMTP servers.

## Available Actions
- [`list`](#list) - List all remotes in a table or output as JSON
- [`create`](#create) - Create a new remote
- [`patch`](#patch) - Update an existing remote
- [`rename`](#rename) - Change the hostname of a remote
- [`disable`](#disable)/[`enable`](#enable) - Disable or enable a remote
- [`delete`](#delete)/[`restore`](#restore) - Delete or restore a remote

## List
Shows a table of all remotes or outputs them as JSON.

### Usage
```sh
mailctl list remotes [flags]
```

### Flags
- `-v`, `--verbose` - Show detailed information with timestamps
- `-j`, `--json` - Output in JSON format
- `-a`, `--all` - Include soft-deleted objects in the output
- `-d`, `--deleted` - Show only soft-deleted objects

## Create
Creates a new remote SMTP relay server configuration.

### Usage
```sh
mailctl create remotes <name> [<name>...] [flags]
```

### Flags
- `-u`, `--username string` - SMTP username (required)
- `-p`, `--password` - Set password interactively
- `--password-method string` - Password hashing method (default: "bcrypt", options: "bcrypt" or "argon2id")
- `--password-hash-options string` - Password hash options (bcrypt: <cost>; argon2id: m=<number>,t=<number>,p=<number>)
- `--password-stdin` - Set password from stdin
- `--port int` - SMTP port (default: 25)
- `-d`, `--disabled` - Create remote in disabled state

### Examples
```sh
# Create with interactive password prompt
mailctl create remotes smtp.relay.com --username myuser --password

# Create with password from stdin
echo "password" | mailctl create remotes smtp.relay.com --username myuser --password-stdin

# Create with custom bcrypt cost
mailctl create remotes smtp.relay.com --username myuser --password --password-hash-options "14"

# Create with argon2id instead of default bcrypt
mailctl create remotes smtp.relay.com --username myuser --password --password-method argon2id

# Create disabled remote
mailctl create remotes smtp.relay.com --username myuser --disabled --password
```

## Patch
Updates properties of an existing remote.

### Usage
```sh
mailctl patch remote <name> [flags]
```

### Flags
- `-e`, `--enabled bool` - Enable or disable the remote
- `-p`, `--password` - Update password interactively
- `--password-method string` - Password hashing method (default: "bcrypt", options: "bcrypt" or "argon2id")
- `--password-hash-options string` - Password hash options (bcrypt: <cost>; argon2id: m=<number>,t=<number>,p=<number>)
- `--password-stdin` - Set password from stdin
- `--no-password` - Remove password

### Examples
```sh
# Update password (interactive)
mailctl patch remote smtp.relay.com --password

# Update password from stdin
echo "newpassword" | mailctl patch remote smtp.relay.com --password-stdin

# Update password with higher bcrypt cost
mailctl patch remote smtp.relay.com --password --password-hash-options "14"

# Remove password
mailctl patch remote smtp.relay.com --no-password

# Enable/disable remote
mailctl patch remote smtp.relay.com --enabled=false
```

## Rename
Changes the name of an existing remote.

### Usage
```sh
mailctl rename remote <old-name> <new-name>
```

## Enable
Enables a disabled remote. It does not throw an error if the remote is already enabled.

### Usage
```sh
mailctl enable remotes <hostname> [hostname...]
```

## Disable
Disables an active remote. It does not throw an error if the remote is already disabled.

### Usage
```sh
mailctl disable remotes <hostname> [hostname...]
```

## Delete
Soft-deletes a remote. The remote can be restored later. Use `--permanent` to permanently delete it.

### Usage
```sh
mailctl delete remotes <hostname> [hostname...] [flags]
```

### Flags
- `-f`, `--force` - Soft-delete the remote, even if it is already (updates the deletion timestamp)
- `-p`, `--permanent` - Permanently delete the remote

## Restore
Restores a soft-deleted remote.

### Usage
```sh
mailctl restore remotes <hostname> [hostname...]
```
