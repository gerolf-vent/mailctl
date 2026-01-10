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
mailctl list mailboxes [flags] [<domain>...]
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
mailctl create mailboxes [flags] <email> [<email>...]
```

### Flags
- `-p`, `--password` - Set password interactively (prompts)
- `--password-stdin` - Read password from stdin
- `--password-method string` - Password hashing method (default: "argon2id", options: "bcrypt" or "argon2id")
- `--password-hash-options string` - Password hash options (bcrypt: `<cost>`; argon2id: m=`<number>`,t=`<number>`,p=`<number>`)
- `-q`, `--quota int32` - Mailbox quota in bytes
- `--transport string` - Transport name for this mailbox
- `-l`, `--login-disabled` - Disable login (authentication)
- `-r`, `--receiving-disabled` - Disable receiving email
- `-s`, `--sending-disabled` - Disable sending email

### Examples
```sh
# Create mailbox (interactive password prompt)
mailctl create mailboxes user@example.com --password

# Create with password from stdin
echo "secretpassword" | mailctl create mailboxes user@example.com --password-stdin

# Create with custom password hash options (argon2id with higher security)
mailctl create mailboxes user@example.com --password --password-hash-options "m=131072,t=3,p=4"

# Create with bcrypt and custom cost
mailctl create mailboxes user@example.com --password --password-method bcrypt --password-hash-options "12"

# Create with login disabled
mailctl create mailboxes user@example.com --login-disabled

# Create with quota
mailctl create mailboxes user@example.com --quota 5368709120  # 5GB in bytes
```

## Patch
Updates properties of an existing mailbox.

### Usage
```sh
mailctl patch mailboxes [flags] <email> [<email>...]
```

### Flags
- `-p`, `--password` - Update password interactively (prompts)
- `--password-stdin` - Read new password from stdin
- `--password-method string` - Password hashing method (default: "argon2id", options: "bcrypt" or "argon2id")
- `--password-hash-options string` - Password hash options (bcrypt: `<cost>`; argon2id: m=`<number>`,t=`<number>`,p=`<number>`)
- `--no-password` - Remove password
- `-q`, `--quota int32` - New quota in bytes
- `--transport string` - New transport name
- `-l`, `--login bool` - Enable or disable login
- `-r`, `--receiving bool` - Enable or disable receiving email
- `-s`, `--sending bool` - Enable or disable sending email

### Examples
```sh
# Update password (interactive)
mailctl patch mailbox user@example.com --password

# Update password from stdin
echo "newpassword" | mailctl patch mailbox user@example.com --password-stdin

# Update password with custom argon2id settings
mailctl patch mailbox user@example.com --password --password-hash-options "m=262144,t=4,p=8"

# Remove password
mailctl patch mailbox user@example.com --no-password

# Update quota
mailctl patch mailbox user@example.com --quota 10737418240  # 10GB

# Enable/disable login
mailctl patch mailbox user@example.com --login=false
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
mailctl enable mailboxes [flags] <email> [<email>...]
```

### Flags
- `-l`, `--login` - Enable login (authentication)
- `-r`, `--receiving` - Enable receiving email
- `-s`, `--sending` - Enable sending email

## Disable
Disables features of an active mailbox. It does not throw an error if the mailbox is already disabled.

### Usage
```sh
mailctl disable mailboxes [flags] <email> [<email>...]
```

### Flags
- `-l`, `--login` - Disable login (authentication)
- `-r`, `--receiving` - Disable receiving email
- `-s`, `--sending` - Disable sending email

## Delete
Soft-deletes a mailbox. The mailbox can be restored later. Use `--permanent` to permanently delete it.

### Usage
```sh
mailctl delete mailboxes [flags] <email> [<email>...]
```

### Flags
- `-f`, `--force` - Soft-delete the mailbox, even if it is already (updates the deletion timestamp)
- `-p`, `--permanent` - Permanently delete the mailbox

## Restore
Restores a soft-deleted mailbox.

### Usage
```sh
mailctl restore mailboxes <email> [<email>...]
```
