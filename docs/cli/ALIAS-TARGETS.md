# Alias Targets

Manage targets for aliases. Alias targets define where emails sent to an alias should be forwarded to and which addresses can send as the alias.

## Available Actions
- [`list`](#list) - List all alias targets in a table or output as JSON
- [`create`](#create) - Create a new alias target
- [`patch`](#patch) - Update an existing alias target
- [`disable`](#disable)/[`enable`](#enable) - Disable or enable an alias target
- [`delete`](#delete)/[`restore`](#restore) - Delete or restore an alias target

## List
Shows a table of all targets for one or more aliases or outputs them as JSON.

### Usage
```sh
mailctl list alias-targets [alias-email...] [flags]
```

### Flags
- `-v`, `--verbose` - Show detailed information with timestamps
- `-j`, `--json` - Output in JSON format
- `-a`, `--all` - Include soft-deleted objects in the output
- `-d`, `--deleted` - Show only soft-deleted objects

## Create
Creates a new alias target, adding one or more target addresses to an alias.

### Usage
```sh
mailctl create alias-target <alias-email> <target-email> [target-email...] [flags]
```

### Flags
- `-f`, `--forward bool` - Enable forwarding to target (default: false)
- `-s`, `--send bool` - Enable sending from target (default: false)

### Examples
```sh
# Add target with forwarding enabled
mailctl create alias-target alias@example.com target@example.com --forward

# Add target with send permission
mailctl create alias-target alias@example.com mailbox@example.com --send

# Add multiple targets
mailctl create alias-target alias@example.com target1@example.com target2@example.com --forward
```

## Patch
Updates properties of an existing alias target.

### Usage
```sh
mailctl patch alias-target <alias-email> <target-email> [flags]
```

### Flags
- `-f`, `--forward bool` - Enable/disable forwarding to target
- `-s`, `--send bool` - Enable/disable sending from target

### Examples
```sh
# Enable forwarding
mailctl patch alias-target alias@example.com target@example.com --forward=true

# Disable sending
mailctl patch alias-target alias@example.com target@example.com --send=false
```

## Enable
Enables forwarding and/or sending on an alias target. Use flags to select which property to enable.

### Usage
```sh
mailctl enable alias-target <alias-email> <target-email> [<alias-email> <target-email>...] [flags]
```

### Flags
- `-f`, `--forward` - Enable forwarding to target
- `-s`, `--send` - Enable sending from target

## Disable
Disables forwarding and/or sending on an alias target. Use flags to select which property to disable.

### Usage
```sh
mailctl disable alias-target <alias-email> <target-email> [<alias-email> <target-email>...] [flags]
```

### Flags
- `-f`, `--forward` - Disable forwarding to target
- `-s`, `--send` - Disable sending from target

## Delete
Soft-deletes an alias target. The alias target can be restored later. Use `--permanent` to permanently delete it.

### Usage
```sh
mailctl delete alias-target <alias-email> <target-email> [<alias-email> <target-email>...] [flags]
```

### Flags
- `-f`, `--force` - Soft-delete the alias target, even if it is already (updates the deletion timestamp)
- `-p`, `--permanent` - Permanently delete the alias target

## Restore
Restores a soft-deleted alias target.

### Usage
```sh
mailctl restore alias-target <alias-email> <target-email> [<alias-email> <target-email>...]
```
