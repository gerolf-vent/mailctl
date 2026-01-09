package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var CreateMailboxesCmd = &cobra.Command{
	Use:     "mailboxes <email> [email...]",
	Aliases: []string{"mailbox"},
	Short:   "Create a new mailbox",
	Long:    "Creates a new mailbox with the specified email address.\nEmails must be in the format \"name@example.com\".",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagPassword, _ := cmd.Flags().GetBool("password")
		flagPasswordStdin, _ := cmd.Flags().GetBool("password-stdin")
		flagPasswordMethod, _ := cmd.Flags().GetString("password-method")
		flagQuota, _ := cmd.Flags().GetInt32("quota")
		flagTransportName, _ := cmd.Flags().GetString("transport")
		flagLoginDisabled, _ := cmd.Flags().GetBool("login-disabled")
		flagReceivingDisabled, _ := cmd.Flags().GetBool("receiving-disabled")
		flagSendingDisabled, _ := cmd.Flags().GetBool("sending-disabled")

		if flagPassword && flagPasswordStdin {
			return fmt.Errorf("cannot use both --password and --password-stdin")
		}

		if len(args) > 0 && (flagPassword || flagPasswordStdin) {
			return fmt.Errorf("cannot set password while creating multiple mailboxes")
		}

		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}

		options := db.MailboxesCreateOptions{
			LoginEnabled:     !flagLoginDisabled,
			ReceivingEnabled: !flagReceivingDisabled,
			SendingEnabled:   !flagSendingDisabled,
		}

		if flagPassword || flagPasswordStdin {
			passwordHash, err := ReadPasswordHashed(flagPasswordMethod, flagPasswordStdin)
			if err != nil {
				utils.PrintErrorWithMessage("failed to read password", err)
				return nil
			}
			options.PasswordHash.Valid = true
			options.PasswordHash.String = passwordHash
		}

		if flagQuota > 0 {
			options.Quota.Valid = true
			options.Quota.Int32 = flagQuota
		}

		if flagTransportName != "" {
			options.TransportName.Valid = true
			options.TransportName.String = flagTransportName
		}

		runner := db.TxForEachRunner[utils.EmailAddress]{
			Items: argEmails,
			Exec: func(tx *sql.Tx, item utils.EmailAddress) error {
				return db.Mailboxes(tx).Create(item, options)
			},
			ItemString:     func(item utils.EmailAddress) string { return item.String() },
			FailureMessage: "failed to create mailbox",
			SuccessMessage: "Successfully created mailbox",
		}

		runner.Run()
		return nil
	},
}

func init() {
	CreateMailboxesCmd.Flags().Bool("password", false, "Set password interactively (prompts)")
	CreateMailboxesCmd.Flags().String("password-method", "bcrypt", "Password hashing method (default: \"bcrypt\")")
	CreateMailboxesCmd.Flags().Bool("password-stdin", false, "Read password from stdin")
	CreateMailboxesCmd.Flags().Int32("quota", 0, "Mailbox quota in bytes")
	CreateMailboxesCmd.Flags().String("transport", "", "Transport name for this mailbox")
	CreateMailboxesCmd.Flags().Bool("login-disabled", false, "Disable login (authentication)")
	CreateMailboxesCmd.Flags().Bool("receiving-disabled", false, "Disable receiving email")
	CreateMailboxesCmd.Flags().Bool("sending-disabled", false, "Disable sending email")
}
