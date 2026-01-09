# Database Schema

```mermaid
erDiagram
    %% Transport relationships
    transports ||--o{ domains_managed : "default transport"
    transports ||--o{ domains_relayed : "transport"
    transports ||--o{ mailboxes : "optional transport override"
    
    %% Domain relationships
    domains_managed ||--o{ mailboxes : "contains"
    domains_managed ||--o{ aliases : "contains"
    domains_managed ||--o{ domains_catchall_targets : "has catchall"
    
    domains_relayed ||--o{ recipients_relayed : "contains"
    domains_relayed ||--o{ aliases : "contains"
    domains_relayed ||--o{ domains_catchall_targets : "has catchall"
    
    domains_alias ||--o{ aliases : "contains"
    domains_alias ||--o{ domains_catchall_targets : "has catchall"
    
    domains_canonical }o--|| domains_managed : "canonical target"
    domains_canonical }o--|| domains_relayed : "canonical target"
    domains_canonical }o--|| domains_alias : "canonical target"
    
    %% Alias relationships
    aliases ||--o{ aliases_targets_recursive : "forwards to"
    aliases ||--o{ aliases_targets_foreign : "forwards to external"
    
    %% Catchall relationships
    domains_catchall_targets }o--|| mailboxes : "targets"
    domains_catchall_targets }o--|| aliases : "targets"
    domains_catchall_targets }o--|| recipients_relayed : "targets"
    
    %% Recursive alias relationships
    aliases_targets_recursive }o--|| mailboxes : "targets"
    aliases_targets_recursive }o--|| aliases : "targets recursively"
    aliases_targets_recursive }o--|| recipients_relayed : "targets"
    
    %% Remote relationships
    remotes ||--o{ remotes_send_grants : "has grants"
    remotes_send_grants }o--|| domains_managed : "for domain"
    remotes_send_grants }o--|| domains_relayed : "for domain"
    remotes_send_grants }o--|| domains_alias : "for domain"
    
    %% Table definitions
    transports {
        int ID PK
        varchar name UK
        varchar method
        varchar host
        smallint port
        boolean mx_lookup
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    
    domains_managed {
        int ID PK "shared.domains_id"
        varchar fqdn UK
        int transport_id FK "transports"
        boolean enabled
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    
    domains_relayed {
        int ID PK "shared.domains_id"
        varchar fqdn UK
        int transport_id FK "transports"
        boolean enabled
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    
    domains_alias {
        int ID PK "shared.domains_id"
        varchar fqdn UK
        boolean enabled
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    
    domains_canonical {
        int ID PK "shared.domains_id"
        varchar fqdn UK
        int target_domain_id FK "shared.domains_id_recipientable"
        boolean enabled
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    
    domains_catchall_targets {
        int ID PK
        int domain_id FK "shared.domains_id_recipientable"
        int recipient_id FK "shared.recipients_id"
        boolean forwarding_to_target_enabled
        boolean fallback_only
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    
    mailboxes {
        int ID PK "shared.recipients_id"
        int domain_id FK "domains_managed"
        varchar name
        int transport_id FK "transports"
        varchar password_hash
        int storage_quota
        boolean login_enabled
        boolean receiving_enabled
        boolean sending_enabled
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    
    aliases {
        int ID PK "shared.recipients_id"
        int domain_id FK "shared.domains_id_recipientable"
        varchar name
        boolean enabled
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    
    aliases_targets_recursive {
        int ID PK "shared.aliases_targets_id"
        int alias_id FK "aliases"
        int recipient_id FK "shared.recipients_id"
        boolean forwarding_to_target_enabled
        boolean sending_from_target_enabled
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    
    aliases_targets_foreign {
        int ID PK "shared.aliases_targets_id"
        int alias_id FK "aliases"
        varchar fqdn
        varchar name
        boolean forwarding_to_target_enabled
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    
    recipients_relayed {
        int ID PK "shared.recipients_id"
        int domain_id FK "domains_relayed"
        varchar name
        boolean enabled
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    
    remotes {
        int ID PK
        varchar name UK
        varchar password_hash
        boolean enabled
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
    
    remotes_send_grants {
        int ID PK
        int remote_id FK "remotes"
        int domain_id FK "shared.domains_id_recipientable"
        varchar name
        timestamptz created_at
        timestamptz updated_at
        timestamptz deleted_at
    }
```

## Schema Overview

### Core Concepts

#### Transports
Define how mail is delivered (SMTP, LMTP, etc.). Used by domains and can be overridden per mailbox.

#### Domain Types
- **Managed Domains**: Fully managed domains with mailboxes and aliases
- **Relayed Domains**: Domains that relay mail to external systems
- **Alias Domains**: Domains that only contain aliases
- **Canonical Domains**: Domain name mappings for address rewriting

#### Recipients
- **Mailboxes**: Actual email accounts with storage and authentication
- **Aliases**: Email addresses that forward to other recipients (internal or external)
- **Relayed Recipients**: Specific recipients on relayed domains

#### Remotes
External systems or users that can authenticate and send mail through specific grants.

### Soft Deletes

All tables support soft deletes via the `deleted_at` timestamp field. This allows for:
- Data recovery
- Audit trails
- Preventing cascading deletes from breaking audit logs

### Shared ID Sequences

The schema uses shared sequences for:
- **domains_id**: Unique IDs across all domain types
- **domains_id_recipientable**: Subset of domains that can have recipients (managed, relayed, alias)
- **recipients_id**: Unique IDs across mailboxes, aliases, and relayed recipients
- **aliases_targets_id**: Unique IDs across all alias target types

This ensures global uniqueness while allowing polymorphic relationships.

## Integration

See documentation for integration with:
- [Postfix](../integrations/POSTFIX.md)
- [Dovecot](../integrations/DOVECOT.md)
