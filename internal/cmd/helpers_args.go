package cmd

import (
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/utils"
)

func ParseEmailOrWildcardArgs(args []string) []utils.EmailAddressOrWildcard {
	emails := make([]utils.EmailAddressOrWildcard, 0, len(args))
	for _, argEmail := range args {
		parsedEmail, err := utils.ParseEmailAddressOrWildcard(argEmail)
		if err != nil {
			utils.PrintError(fmt.Errorf("invalid email format: %s: %w", argEmail, err))
		}
		emails = append(emails, parsedEmail)
	}

	return emails
}

func ParseEmailArgs(args []string) []utils.EmailAddress {
	emails := make([]utils.EmailAddress, 0, len(args))
	for _, argEmail := range args {
		parsedEmail, err := utils.ParseEmailAddress(argEmail)
		if err != nil {
			utils.PrintError(fmt.Errorf("invalid email format: %s: %w", argEmail, err))
		}
		emails = append(emails, parsedEmail)
	}

	return emails
}

func ParseDomainFQDNArgs(args []string) []string {
	domains := make([]string, 0, len(args))
	for _, arg := range args {
		domainFQDN, err := utils.ParseDomainFQDN(arg)
		if err != nil {
			utils.PrintError(fmt.Errorf("invalid domain format: %s: %w", arg, err))
		}
		domains = append(domains, domainFQDN)
	}
	return domains
}
