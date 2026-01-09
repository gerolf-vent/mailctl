# Catchall Targets

Manage catch-all addresses that receive all or only unmatched emails for a domain.

## Available Actions
- [`list`](#list) - List all catchall targets in a table or output as JSON
- [`create`](#create) - Create a new catchall target
- [`describe`](#describe) - Show detailed information about a catchall target
- [`patch`](#patch) - Update an existing catchall target
- [`disable`](#disable)/[`enable`](#enable) - Disable or enable a catchall target
- [`delete`](#delete)/[`restore`](#restore) - Delete or restore a catchall target

## List
Shows a table of all catchall targets or outputs them as JSON. Can be filtered by domain.

### Usage
```sh
mailctl list catchall-targets [domain...] [flags]
```

### Flags
- `-v`, `--verbose` - Show detailed information with timestamps
- `-j`, `--json` - Output in JSON format
- `-a`, `--all` - Include soft-deleted objects in the output
- `-d`, `--deleted` - Show only soft-deleted objects

## Create
Creates a new catchall target for a domain.

### Usage
```sh
mailctl create catchall-targets <domain> <target-email> [target-email...] [flags]
```

### Flags
- `-f`, `--forward bool` - Enable forwarding to target (default: true)
- `-b`, `--only-fallback bool` - Only forward if no other recipient matched (default: true)

### Examples
```sh
# Create catchall for domain
mailctl create catchall-targets example.com catchall@example.com

# Create without forwarding
mailctl create catchall-targets example.com catchall@example.com --forward=false

# Create without fallback-only behavior
mailctl create catchall-targets example.com catchall@example.com --only-fallback=false
```

## Patch
Updates properties of an existing catchall target.

### Usage
```sh
mailctl patch catchall-targets <domain> <target-email> [target-email...] [flags]
```

### Flags
- `-f`, `--forward bool` - Enable/disable forwarding to target
- `-b`, `--only-fallback bool` - Enable/disable only-fallback behavior

### Examples
```sh
mailctl patch catchall-targets example.com catchall@example.com --only-fallback=true
mailctl patch catchall-targets example.com catchall@example.com --forward=false
```

## Enable
Enables a disabled catchall target. This will affect the forwarding capability only. It does not throw an error if the catchall target is already enabled.

### Usage
```sh
mailctl enable catchall-targets <domain> <target> [<domain> <target>...]
```

## Disable
Disables an active catchall target. This will affect the forwarding capability only. It does not throw an error if the catchall target is already disabled.

### Usage
```sh
mailctl disable catchall-target <domain> <target> [<domain> <target>...]
```

## Delete
Soft-deletes a catchall target. The catchall target can be restored later. Use `--permanent` to permanently delete it.

### Usage
```sh
mailctl delete catchall-target <domain> <target> [<domain> <target>...] [flags]
```

### Flags
- `-f`, `--force` - Soft-delete the catchall target, even if it is already (updates the deletion timestamp)
- `-p`, `--permanent` - Permanently delete the catchall target

## Restore
Restores a soft-deleted catchall target.

### Usage
```sh
mailctl restore catchall-targets <domain> <target> [<domain> <target>...]
```
