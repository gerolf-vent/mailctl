package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	localPartAllowed = regexp.MustCompile("(?i)^[a-z0-9!#$%&'*+\\-/=?^_`{|}~.]+$")
)

type EmailAddress struct {
	LocalPart  string
	DomainFQDN string
}

func (ea *EmailAddress) String() string {
	return fmt.Sprintf("%s@%s", ea.LocalPart, ea.DomainFQDN)
}

func ParseEmailAddress(address string) (EmailAddress, error) {
	address = strings.TrimSpace(address)
	if strings.Count(address, "@") != 1 {
		return EmailAddress{}, fmt.Errorf("invalid email format: expected name@domain")
	}

	parts := strings.SplitN(address, "@", 2)
	localPart := parts[0]
	domainPart := parts[1]

	if err := validateLocalPart(localPart); err != nil {
		return EmailAddress{}, err
	}

	parsedDomain, err := ParseDomainFQDN(domainPart)
	if err != nil {
		return EmailAddress{}, err
	}

	return EmailAddress{
		LocalPart:  localPart,
		DomainFQDN: parsedDomain,
	}, nil
}

type EmailAddressOrWildcard struct {
	LocalPart  *string
	DomainFQDN string
}

func (ea *EmailAddressOrWildcard) IsWildcard() bool {
	return ea.LocalPart == nil
}

func (ea *EmailAddressOrWildcard) String() string {
	localPart := ""
	if ea.LocalPart != nil {
		localPart = *ea.LocalPart
	}
	return fmt.Sprintf("%s@%s", localPart, ea.DomainFQDN)
}

func ParseEmailAddressOrWildcard(address string) (EmailAddressOrWildcard, error) {
	address = strings.TrimSpace(address)
	if strings.Count(address, "@") != 1 {
		return EmailAddressOrWildcard{}, fmt.Errorf("invalid email format: expected name@domain")
	}

	parts := strings.SplitN(address, "@", 2)
	localPartRaw := parts[0]
	domainPart := parts[1]

	var localPart *string
	if localPartRaw != "" {
		if err := validateLocalPart(localPartRaw); err != nil {
			return EmailAddressOrWildcard{}, err
		}
		localPart = &localPartRaw
	}

	parsedDomain, err := ParseDomainFQDN(domainPart)
	if err != nil {
		return EmailAddressOrWildcard{}, err
	}

	return EmailAddressOrWildcard{
		LocalPart:  localPart,
		DomainFQDN: parsedDomain,
	}, nil
}

func validateLocalPart(local string) error {
	if len(local) < 1 || len(local) > 64 {
		return fmt.Errorf("local part length must be between 1 and 64 characters")
	}
	if strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") {
		return fmt.Errorf("local part cannot start or end with a dot")
	}
	if strings.Contains(local, "..") {
		return fmt.Errorf("local part cannot contain consecutive dots")
	}
	if !localPartAllowed.MatchString(local) {
		return fmt.Errorf("local part contains invalid characters")
	}
	return nil
}
