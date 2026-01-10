# Domains

There are four types of domains: managed, relayed, alias and canonical. Once a domain is created with a specific type you can't change it, unless you delete and recreate the domain (other properties as fqdn and transport can still be altered).

| Type        | Description                                                                                                                                          |
| ----------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `managed`   | Domains for which mailboxes can be created. Also supports aliases.                                                                                   |
| `relayed`   | Domains for which mail is relayed to an external SMTP server. For this you have to configure relayed recipients.                                     |
| `alias`     | Domains that only exist to define aliases on them. No mailboxes or relayed recipients can be created on alias domains.                               |
| `canonical` | Domains that rewrite all incoming/outgoing mail to a target domain. No mailboxes, relayed recipients or aliases can be created on canonical domains. |

## Available Actions
- [`list`](#list) - List all domains in a table or output as JSON
- [`create`](#create) - Create a new domain
- [`patch`](#patch) - Update an existing domain
- [`rename`](#rename) - Change the FQDN of a domain
- [`disable`](#disable)/[`enable`](#enable) - Disable or enable a domain
- [`delete`](#delete)/[`restore`](#restore) - Delete or restore a domain

## List
Shows a table of all domains or outputs them as JSON.

### Usage
```sh
mailctl list domains [flags]
```

### Flags
- `-v`, `--verbose` - Show detailed information with timestamps
- `-j`, `--json` - Output in JSON format
- `-a`, `--all` - Include soft-deleted objects in the output
- `-d`, `--deleted` - Show only soft-deleted objects

## Create
Creates a new domain of the specified type.

### Usage
```sh
mailctl create domains [flags] <fqdn> [<fqdn>...]
```

### Flags
- `-t`, `--type string` - Domain type: 'managed', 'relayed', 'alias' or 'canonical' (default: "managed")
- `--transport string` - Transport name (required for managed/relayed domains)
- `--target-domain string` - Target domain FQDN (required for canonical domains)
- `-d`, `--disabled` - Create in disabled state

### Examples
```sh
# Create managed domain
mailctl create domains example.com --type managed --transport mailboxes1

# Create relayed domain
mailctl create domains relay.example.com --type relayed --transport smtp-relay

# Create alias domain (only for defining aliases on it)
mailctl create domains aliasdomain.com --type alias

# Create canonical domain (rewrite domain to target domain)
mailctl create domains alias.example.com --type canonical --target-domain example.com

# Create disabled domain
mailctl create domains test.com --disabled
```

## Patch
Updates properties of an existing domain.

### Usage
```sh
mailctl patch domains [flags] <fqdn> [<fqdn>...]
```

### Flags
- `-e`, `--enabled bool` - Enable or disable the domain
- `--transport string` - New transport name (only for managed/relayed domains)
- `--target-domain string` - New target domain FQDN (only for canonical domains)

### Examples
```sh
# For managed/relayed domain
mailctl patch domains example.com --transport new-transport

# For canonical domain
mailctl patch domains alias.example.com --target-domain newtarget.com
```

## Rename
Changes the FQDN of an existing domain.

### Usage
```sh
mailctl rename domain <old-fqdn> <new-fqdn>
```

## Enable
Enables a disabled domain. It does not throw an error if the domain is already enabled.

### Usage
```sh
mailctl enable domains <fqdn> [<fqdn>...]
```

## Disable
Disables an active domain. It does not throw an error if the domain is already disabled.

### Usage
```sh
mailctl disable domains <fqdn> [<fqdn>...]
```

## Delete
Soft-deletes a domain. The domain can be restored later. Use `--permanent` to permanently delete it.

### Usage
```sh
mailctl delete domains [flags] <fqdn> [<fqdn>...]
```

### Flags
- `-f`, `--force` - Soft-delete the domain, even if it is already (updates the deletion timestamp)
- `-p`, `--permanent` - Permanently delete the domain

## Restore
Restores a soft-deleted domain.

### Usage
```sh
mailctl restore domains <fqdn> [<fqdn>...]
```
