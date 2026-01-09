# Transports

Manage mail transport configurations. Transports define how mail should be delivered (method, host, port, and MX lookup settings).

## Available Actions
- [`list`](#list) - List all transports in a table or output as JSON
- [`create`](#create) - Create a new transport
- [`describe`](#describe) - Show detailed information about a transport
- [`patch`](#patch) - Update an existing transport
- [`rename`](#rename) - Change the name of a transport
- [`delete`](#delete)/[`restore`](#restore) - Delete or restore a transport

## List
Shows a table of all transports or outputs them as JSON.

### Usage
```sh
mailctl list transports [flags]
```

### Flags
- `-v`, `--verbose` - Show detailed information with timestamps
- `-j`, `--json` - Output in JSON format
- `-a`, `--all` - Include soft-deleted objects in the output
- `-d`, `--deleted` - Show only soft-deleted objects

## Create
Creates a new mail transport configuration.

### Usage
```sh
mailctl create transport <name> [flags]
```

### Flags
- `-m`, `--method string` - Transport method (required)
- `-H`, `--host string` - Transport host (required)
- `-p`, `--port uint16` - Transport port
- `-x`, `--mx-lookup bool` - Enable MX lookup for this transport

### Examples
```sh
# Create SMTP transport
mailctl create transport smtp-out --method smtp --host smtp.relay.com

# Create transport with custom port
mailctl create transport smtp-relay --method smtp --host smtp.relay.com --port 587

# Create transport with MX lookup
mailctl create transport mx-relay --method relay --host relay.com --mx-lookup
```

## Patch
Updates properties of an existing transport.

### Usage
```sh
mailctl patch transports <name> [name...] [flags]
```

### Flags
- `-m`, `--method string` - Transport method
- `-H`, `--host string` - Transport host
- `-p`, `--port uint16` - Transport port
- `-x`, `--mx-lookup bool` - Enable MX lookup for this transport

### Examples
```sh
# Update host
mailctl patch transports smtp-out --host new-smtp.relay.com

# Update port
mailctl patch transports smtp-out --port 465

# Enable MX lookup
mailctl patch transports smtp-out --mx-lookup=true
```

## Rename
Changes the name of an existing transport.

### Usage
```sh
mailctl rename transport <old-name> <new-name>
```

## Delete
Soft-deletes a transport. The transport can be restored later. Use `--permanent` to permanently delete it.

### Usage
```sh
mailctl delete transports <name> [name...] [flags]
```

### Flags
- `-f`, `--force` - Soft-delete the transport, even if it is already (updates the deletion timestamp)
- `-p`, `--permanent` - Permanently delete the transport

## Restore
Restores a soft-deleted transport.

### Usage
```sh
mailctl restore transports <name> [name...]
```
