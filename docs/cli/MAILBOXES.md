# Mailboxes

Manage user mailboxes and email accounts. Mailboxes are user accounts that can receive and send email, with authentication and storage.

## Available Actions
- [`list`](#list) - List all mailboxes in a table or output as JSON
- [`create`](#create) - Create a new mailbox
- [`describe`](#describe) - Show detailed information about a mailbox
- [`patch`](#patch) - Update an existing mailbox
- [`rename`](#rename) - Change the email address of a mailbox
- [`disable`](#disable)/[`enable`](#enable) - Disable or enable a mailbox
- [`delete`](#delete)/[`restore`](#restore) - Delete or restore a mailbox

## List
Shows a table of all mailboxes or outputs them as JSON. Can be filtered by domain.

### Usage
```sh
mailctl list mailboxes [domain...] [flags]
```

### Flags
- `-v`, `--verbose` - Show detailed information with timestamps
- `-j`, `--json` - Output in JSON format
- `-a`, `--all` - Include soft-deleted objects in the output
- `-d`, `--deleted` - Show only soft-deleted objects

### Examples
```sh
mailctl list mailboxes                    # All mailboxes
mailctl list mailboxes example.com        # For specific domain
mailctl list mailboxes example.com test.com  # Multiple domains
```

## Create
Creates a new mailbox with the specified email address.

### Usage
```sh
mailctl create mailboxes <email> [email...] [flags]
```

### Flags
- `--password` - Set password interactively (prompts)
- `--password-stdin` - Read password from stdin
- `--password-method string` - Password hashing method (default: "bcrypt")
- `--quota int32` - Mailbox quota in bytes
- `--transport string` - Transport name for this mailbox
- `--login-disabled` - Disable login (authentication)
- `--receiving-disabled` - Disable receiving email
- `--sending-disabled` - Disable sending email

### Examples
```sh
# Create mailbox (interactive password prompt)
mailctl create mailboxes user@example.com --password

# Create with password from stdin
echo "secretpassword" | mailctl create mailboxes user@example.com --password-stdin

# Create with login disabled
mailctl create mailboxes user@example.com --login-disabled

# Create with quota
mailctl create mailboxes user@example.com --quota 5368709120  # 5GB in bytes
```

## Patch
Updates properties of an existing mailbox.

### Usage
```sh
mailctl patch mailbox <email> [flags]
```

### Flags
- `--password` - Update password interactively (prompts)
- `--password-stdin` - Read new password from stdin
- `--password-method string` - Password hashing method (default: "bcrypt")
- `--quota int32` - New quota in bytes
- `--transport string` - New transport name
- `--login-enabled bool` - Enable or disable login
- `--receiving-enabled bool` - Enable or disable receiving email
- `--sending-enabled bool` - Enable or disable sending email

### Examples
```sh
# Update password (interactive)
mailctl patch mailbox user@example.com --password

# Update password from stdin
echo "newpassword" | mailctl patch mailbox user@example.com --password-stdin

# Update quota
mailctl patch mailbox user@example.com --quota 10737418240  # 10GB

# Enable/disable login
mailctl patch mailbox user@example.com --login-enabled=false
```

## Rename
Changes the email address of an existing mailbox (including domain).

### Usage
```sh
mailctl rename mailbox <old-email> <new-email>
```

## Enable
Enables features of a disabled mailbox. It does not throw an error if the mailbox is already enabled.

### Usage
```sh
mailctl enable mailboxes <email> [email...] [flags]
```

### Flags
- `-l`, `--login` - Enable login (authentication)
- `-r`, `--receiving` - Enable receiving email
- `-s`, `--sending` - Enable sending email

## Disable
Disables features of an active mailbox. It does not throw an error if the mailbox is already disabled.

### Usage
```sh
mailctl disable mailboxes <email> [email...] [flags]
```

### Flags
- `-l`, `--login` - Disable login (authentication)
- `-r`, `--receiving` - Disable receiving email
- `-s`, `--sending` - Disable sending email

## Delete
Soft-deletes a mailbox. The mailbox can be restored later. Use `--permanent` to permanently delete it.

### Usage
```sh
mailctl delete mailboxes <email> [email...] [flags]
```

### Flags
- `-f`, `--force` - Soft-delete the mailbox, even if it is already (updates the deletion timestamp)
- `-p`, `--permanent` - Permanently delete the mailbox

## Restore
Restores a soft-deleted mailbox.

### Usage
```sh
mailctl restore mailboxes <email> [email...]
```
