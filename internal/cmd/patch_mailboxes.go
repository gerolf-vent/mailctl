package cmd

import (
	"database/sql"
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var PatchMailboxesCmd = &cobra.Command{
	Use:   "mailbox <email>",
	Short: "Update an existing mailbox",
	Long:  "Updates properties of an existing mailbox.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagPassword, _ := cmd.Flags().GetBool("password")
		flagPasswordStdin, _ := cmd.Flags().GetBool("password-stdin")
		flagPasswordMethod, _ := cmd.Flags().GetString("password-method")
		flagPasswordNo, _ := cmd.Flags().GetBool("no-password")

		if (flagPassword || flagPasswordStdin) && flagPasswordNo {
			return fmt.Errorf("cannot use --no-password with --password or --password-stdin")
		}

		if flagPassword && flagPasswordStdin {
			return fmt.Errorf("cannot use both --password and --password-stdin")
		}

		argEmails := ParseEmailArgs(args)
		if len(argEmails) != len(args) {
			return fmt.Errorf("invalid email arguments")
		}

		argEmail := argEmails[0]

		options := db.MailboxesPatchOptions{}
		if flagPassword || flagPasswordStdin {
			passwordHash, err := ReadPasswordHashed(flagPasswordMethod, flagPasswordStdin)
			if err != nil {
				utils.PrintErrorWithMessage("failed to read password", err)
				return nil
			}
			options.PasswordHash = &sql.NullString{Valid: true, String: passwordHash}
		}
		if flagPasswordNo {
			options.PasswordHash = &sql.NullString{Valid: false}
		}
		if cmd.Flags().Changed("quota") {
			q, _ := cmd.Flags().GetInt32("quota")
			if q <= 0 {
				options.Quota = &sql.NullInt32{Valid: false}
			} else {
				options.Quota = &sql.NullInt32{Valid: true, Int32: q}
			}
		}
		if cmd.Flags().Changed("transport") {
			transportName, _ := cmd.Flags().GetString("transport")
			if transportName == "-" {
				options.TransportName = &sql.NullString{Valid: false}
			} else {
				options.TransportName = &sql.NullString{Valid: true, String: transportName}
			}
		}
		if cmd.Flags().Changed("login") {
			v, _ := cmd.Flags().GetBool("login")
			options.Login = &v
		}
		if cmd.Flags().Changed("receiving") {
			v, _ := cmd.Flags().GetBool("receiving")
			options.Receiving = &v
		}
		if cmd.Flags().Changed("sending") {
			v, _ := cmd.Flags().GetBool("sending")
			options.Sending = &v
		}

		runner := db.TxRunner{
			Exec: func(tx *sql.Tx) error {
				return db.Mailboxes(tx).Patch(argEmail, options)
			},
			ItemString:     argEmail.String(),
			FailureMessage: "failed to patch mailbox",
			SuccessMessage: "Successfully patched mailbox",
		}

		runner.Run()
		return nil
	},
}

func init() {
	PatchMailboxesCmd.Flags().Bool("password", false, "Update password interactively (prompts)")
	PatchMailboxesCmd.Flags().String("password-method", "bcrypt", "Password hashing method")
	PatchMailboxesCmd.Flags().Bool("password-stdin", false, "Read new password from stdin")
	PatchMailboxesCmd.Flags().Bool("no-password", false, "Remove password")
	PatchMailboxesCmd.Flags().Int32("quota", 0, "New quota in bytes")
	PatchMailboxesCmd.Flags().String("transport", "", "New transport name")
	PatchMailboxesCmd.Flags().Bool("login-enabled", true, "Enable or disable login")
	PatchMailboxesCmd.Flags().Bool("receiving-enabled", true, "Enable or disable receiving email")
	PatchMailboxesCmd.Flags().Bool("sending-enabled", true, "Enable or disable sending email")
}
