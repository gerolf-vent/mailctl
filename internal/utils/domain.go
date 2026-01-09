package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	domainAllowedChars = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)
	domainPattern      = regexp.MustCompile(`(?i)^([a-z0-9]([a-z0-9\-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$`)
)

// ParseDomainFQDN validates a domain using the same rules as the SQL schema.
func ParseDomainFQDN(fqdn string) (string, error) {
	fqdn = strings.TrimSpace(fqdn)

	if len(fqdn) < 3 || len(fqdn) > 253 {
		return "", fmt.Errorf("domain length must be between 3 and 253 characters")
	}

	if !domainAllowedChars.MatchString(fqdn) {
		return "", fmt.Errorf("domain contains invalid characters")
	}

	if !domainPattern.MatchString(fqdn) {
		return "", fmt.Errorf("domain is not a valid FQDN")
	}

	return fqdn, nil
}
