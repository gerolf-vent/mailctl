package db

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

func PostfixTransportMaps(r sq.BaseRunner, email utils.EmailAddress) (string, error) {
	var res string
	err := sq.
		Select("result").
		Suffix("FROM postfix.transport_maps(?, ?)", email.DomainFQDN, email.LocalPart).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRow().
		Scan(&res)
	if err != nil {
		return "", err
	}
	return res, nil
}

func PostfixVirtualAliasDomains(r sq.BaseRunner, fqdn string) (string, error) {
	var res string
	err := sq.
		Select("result").
		Suffix("FROM postfix.virtual_alias_domains(?)", fqdn).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRow().
		Scan(&res)
	if err != nil {
		return "", err
	}
	return res, nil
}

func PostfixVirtualAliasMaps(r sq.BaseRunner, email utils.EmailAddress, limit uint32) ([]string, error) {
	var results []string
	rows, err := sq.
		Select("result").
		Suffix("FROM postfix.virtual_alias_maps(?, ?, ?)", email.DomainFQDN, email.LocalPart, limit).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var res string
		if err := rows.Scan(&res); err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	return results, nil
}

func PostfixVirtualMailboxDomains(r sq.BaseRunner, fqdn string) (string, error) {
	var res string
	err := sq.
		Select("result").
		Suffix("FROM postfix.virtual_mailbox_domains(?)", fqdn).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRow().
		Scan(&res)
	if err != nil {
		return "", err
	}
	return res, nil
}

func PostfixVirtualMailboxMaps(r sq.BaseRunner, email utils.EmailAddress) (string, error) {
	var res string
	err := sq.
		Select("result").
		Suffix("FROM postfix.virtual_mailbox_maps(?, ?)", email.DomainFQDN, email.LocalPart).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRow().
		Scan(&res)
	if err != nil {
		return "", err
	}
	return res, nil
}

func PostfixSMTPDSenderLoginMapsMailboxes(r sq.BaseRunner, email utils.EmailAddress, limit uint32) ([]string, error) {
	var results []string
	rows, err := sq.
		Select("result").
		Suffix("FROM postfix.smtpd_sender_login_maps_mailboxes(?, ?, ?)", email.DomainFQDN, email.LocalPart, limit).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var res string
		if err := rows.Scan(&res); err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	return results, nil
}

func PostfixRelayDomains(r sq.BaseRunner, fqdn string) (string, error) {
	var res string
	err := sq.
		Select("result").
		Suffix("FROM postfix.relay_domains(?)", fqdn).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRow().
		Scan(&res)
	if err != nil {
		return "", err
	}
	return res, nil
}

func PostfixRelayRecipientMaps(r sq.BaseRunner, email utils.EmailAddress) (string, error) {
	var res string
	err := sq.
		Select("result").
		Suffix("FROM postfix.relay_recipient_maps(?, ?)", email.DomainFQDN, email.LocalPart).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRow().
		Scan(&res)
	if err != nil {
		return "", err
	}
	return res, nil
}

func PostfixSMTPDSenderLoginMapsRemotes(r sq.BaseRunner, email utils.EmailAddress) ([]string, error) {
	var results []string
	rows, err := sq.
		Select("result").
		Suffix("FROM postfix.smtpd_sender_login_maps_remotes(?, ?)", email.DomainFQDN, email.LocalPart).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var res string
		if err := rows.Scan(&res); err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	return results, nil
}

func PostfixCanonicalMaps(r sq.BaseRunner, email utils.EmailAddress) (string, error) {
	var res string
	err := sq.
		Select("result").
		Suffix("FROM postfix.canonical_maps(?, ?)", email.DomainFQDN, email.LocalPart).
		PlaceholderFormat(sq.Dollar).
		RunWith(r).
		QueryRow().
		Scan(&res)
	if err != nil {
		return "", err
	}
	return res, nil
}
