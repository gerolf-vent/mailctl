package test

import (
	"database/sql"
	"regexp"
	"strings"
)

func lookupDomain(domainID int) (fqdn string, dType string, enabled bool, deletedAt sql.NullTime, ok bool) {
	if d, ok := fixtures.DomainsManaged[domainID]; ok {
		return d.FQDN, "managed", d.Enabled, d.DeletedAt, true
	}
	if d, ok := fixtures.DomainsRelayed[domainID]; ok {
		return d.FQDN, "relayed", d.Enabled, d.DeletedAt, true
	}
	if d, ok := fixtures.DomainsAlias[domainID]; ok {
		return d.FQDN, "alias", d.Enabled, d.DeletedAt, true
	}
	if d, ok := fixtures.DomainsCanonical[domainID]; ok {
		return d.FQDN, "canonical", d.Enabled, d.DeletedAt, true
	}
	return "", "", false, sql.NullTime{}, false
}

func SQLPatternToRegex(pattern string) *regexp.Regexp {
	var b strings.Builder
	b.WriteString("^")

	escaped := false
	for _, r := range pattern {
		switch {
		case escaped:
			b.WriteString(regexp.QuoteMeta(string(r)))
			escaped = false
		case r == '\\':
			escaped = true
		case r == '%':
			b.WriteString(".*")
		case r == '_':
			b.WriteString(".")
		default:
			b.WriteString(regexp.QuoteMeta(string(r)))
		}
	}

	if escaped {
		b.WriteString(regexp.QuoteMeta("\\"))
	}

	b.WriteString("$")
	return regexp.MustCompile(b.String())
}
