package db

import (
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/gerolf-vent/mailctl/internal/utils"
)

const (
	MailboxesLoginAttemptsTable string = "audit.mailboxes_login_attempts"
)

type MailboxLoginAttempt struct {
	DomainFQDN    string    `json:"domainFQDN"`
	Name          string    `json:"name"`
	Succeeded     bool      `json:"succeeded"`
	FailureReason string    `json:"failureReason"`
	AttemptedAt   time.Time `json:"attemptedAt"`
}

type MailboxesLoginAttemptsListOptions struct {
	FilterDomains []string
	FilterEmails  []*utils.EmailAddress
}

type MailboxesLoginAttemptsRepository interface {
	List(options MailboxesLoginAttemptsListOptions) ([]MailboxLoginAttempt, error)
	CheckRateLimit(email utils.EmailAddress, count uint32, interval time.Duration) (ok bool, err error)
	Record(email utils.EmailAddress, succeeded bool, failureReason string) (err error)
}

type mailboxesLoginAttemptsRepository struct {
	r sq.BaseRunner
}

func MailboxesLoginAttempts(r sq.BaseRunner) MailboxesLoginAttemptsRepository {
	return &mailboxesLoginAttemptsRepository{
		r: r,
	}
}

func (r *mailboxesLoginAttemptsRepository) List(options MailboxesLoginAttemptsListOptions) ([]MailboxLoginAttempt, error) {
	rows, err := sq.
		Select(
			"domain_fqdn",
			"name",
			"succeeded",
			"failure_reason",
			"attempted_at",
		).
		From(MailboxesLoginAttemptsTable).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.r).
		Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []MailboxLoginAttempt
	for rows.Next() {
		var mla MailboxLoginAttempt
		if err := rows.Scan(
			&mla.DomainFQDN,
			&mla.Name,
			&mla.Succeeded,
			&mla.FailureReason,
			&mla.AttemptedAt,
		); err != nil {
			return nil, err
		}

		out = append(out, mla)
	}

	return out, nil
}

func (r *mailboxesLoginAttemptsRepository) CheckRateLimit(email utils.EmailAddress, count uint32, interval time.Duration) (ok bool, err error) {
	var attempts uint32

	err = sq.
		Select("COUNT(*)").
		From(MailboxesLoginAttemptsTable).
		Where(sq.Eq{
			"domain_fqdn": email.DomainFQDN,
			"name":        email.LocalPart,
		}).
		Where("attempted_at > ?", time.Now().Add(-1*interval)).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.r).
		QueryRow().
		Scan(&attempts)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil
		}
		return false, err
	}

	return attempts < count, nil
}

func (r *mailboxesLoginAttemptsRepository) Record(email utils.EmailAddress, succeeded bool, failureReason string) (err error) {
	q := sq.
		Insert(MailboxesLoginAttemptsTable).
		Columns("domain_fqdn", "name", "succeeded", "failure_reason").
		Values(email.DomainFQDN, email.LocalPart, succeeded, failureReason)

	return Exec(r.r, q, 1)
}
