# Aliases

Manage email aliases for forwarding, distribution and send-as. Aliases allow you to forward emails to one or more target addresses and also send-as with mailboxes.

## Available Actions
- [`list`](#list) - List all aliases in a table or output as JSON
- [`create`](#create) - Create a new alias
- [`patch`](#patch) - Update an existing alias
- [`rename`](#rename) - Change the email address of an alias
- [`disable`](#disable)/[`enable`](#enable) - Disable or enable an alias
- [`delete`](#delete)/[`restore`](#restore) - Delete or restore an alias

## List
Shows a table of all aliases or outputs them as JSON. Can be filtered by domain.

### Usage
```sh
mailctl list aliases [domain...] [flags]
```

### Flags
- `-v`, `--verbose` - Show detailed information with timestamps
- `-j`, `--json` - Output in JSON format
- `-a`, `--all` - Include soft-deleted objects in the output
- `-d`, `--deleted` - Show only soft-deleted objects

## Create
Creates a new alias. Note: After creating an alias, you need to add targets using the `create alias-target` command.

### Usage
```sh
mailctl create aliases <email> [email...] [flags]
```

### Examples
```sh
# Create alias
mailctl create aliases alias@example.com

# Create multiple aliases
mailctl create aliases alias1@example.com alias2@example.com
```

## Patch
Updates properties of an existing alias.

### Usage
```sh
mailctl patch aliases <email> [email...] [flags]
```

### Flags
- `--enabled bool` - Enable or disable the alias

## Rename
Changes the email address of an existing alias (including domain).

### Usage
```sh
mailctl rename alias <old-email> <new-email>
```

## Enable
Enables a disabled alias. It does not throw an error if the alias is already enabled.

### Usage
```sh
mailctl enable aliases <email> [email...]
```

## Disable
Disables an active alias. It does not throw an error if the alias is already disabled.

### Usage
```sh
mailctl disable aliases <email> [email...]
```

## Delete
Soft-deletes an alias. The alias can be restored later. Use `--permanent` to permanently delete it.

### Usage
```sh
mailctl delete aliases <email> [email...] [flags]
```

### Flags
- `-f`, `--force` - Soft-delete the alias, even if it is already (updates the deletion timestamp)
- `-p`, `--permanent` - Permanently delete the alias

## Restore
Restores a soft-deleted alias.

### Usage
```sh
mailctl restore aliases <email> [email...]
```
