# Relayed Recipients

Manage individual recipients on relayed domains. Relayed recipients are specific email addresses on relayed domains that are allowed to receive mail.

## Available Actions
- [`list`](#list) - List all relayed recipients in a table or output as JSON
- [`create`](#create) - Create a new relayed recipient
- [`describe`](#describe) - Show detailed information about a relayed recipient
- [`patch`](#patch) - Update an existing relayed recipient
- [`rename`](#rename) - Change the email address of a relayed recipient
- [`disable`](#disable)/[`enable`](#enable) - Disable or enable a relayed recipient
- [`delete`](#delete)/[`restore`](#restore) - Delete or restore a relayed recipient

## List
Shows a table of all relayed recipients or outputs them as JSON. Can be filtered by domain.

### Usage
```sh
mailctl list relayed-recipients [flags] [<domain>...]
```

### Flags
- `-v`, `--verbose` - Show detailed information with timestamps
- `-j`, `--json` - Output in JSON format
- `-a`, `--all` - Include soft-deleted objects in the output
- `-d`, `--deleted` - Show only soft-deleted objects

## Create
Creates a new relayed recipient.

### Usage
```sh
mailctl create recipients-relayed [flags] <email> [<email>...]
```

### Flags
- `-d`, `--disabled` - Create relayed recipient in disabled state

## Patch
Updates properties of an existing relayed recipient.

### Usage
```sh
mailctl patch recipients-relayed [flags] <email> [<email>...]
```

### Flags
- `-e`, `--enabled bool` - Enable/disable relayed recipient

## Rename
Changes the email address of an existing relayed recipient (including domain).

### Usage
```sh
mailctl rename recipient-relayed <old-email> <new-email>
```

## Enable
Enables a disabled relayed recipient. It does not throw an error if the relayed recipient is already enabled.

### Usage
```sh
mailctl enable recipients-relayed <email> [<email>...]
```

## Disable
Disables an active relayed recipient. It does not throw an error if the relayed recipient is already disabled.

### Usage
```sh
mailctl disable recipients-relayed <email> [<email>...]
```

## Delete
Soft-deletes a relayed recipient. The relayed recipient can be restored later. Use `--permanent` to permanently delete it.

### Usage
```sh
mailctl delete recipients-relayed [flags] <email> [<email>...]
```

### Flags
- `-f`, `--force` - Soft-delete the relayed recipient, even if it is already (updates the deletion timestamp)
- `-p`, `--permanent` - Permanently delete the relayed recipient

## Restore
Restores a soft-deleted relayed recipient.

### Usage
```sh
mailctl restore recipients-relayed <email> [<email>...]
```
