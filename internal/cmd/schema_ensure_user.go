package cmd

import (
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/schema"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var SchemaEnsureUserCmd = &cobra.Command{
	Use:   "ensure-user [username]",
	Short: "Create/sync application-specific database user with limited permissions",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagType, _ := cmd.Flags().GetString("type")
		flagPasswordPrompt, _ := cmd.Flags().GetBool("password")
		flagPasswordStdin, _ := cmd.Flags().GetBool("password-stdin")
		flagTypeFile, _ := cmd.Flags().GetString("type-file")
		flagNameFile, _ := cmd.Flags().GetString("name-file")
		flagPasswordFile, _ := cmd.Flags().GetString("password-file")
		flagEnvPrefix, _ := cmd.Flags().GetString("env-prefix")

		argUsername := ""
		if len(args) > 0 {
			argUsername = args[0]
		}

		userType, err := ResolveDynamicSourceArg(flagType, flagTypeFile, fmt.Sprintf("%s_TYPE", flagEnvPrefix))
		if err != nil {
			utils.PrintError(fmt.Errorf("failed to retreive user type: %w", err))
			return nil
		}

		userName, err := ResolveDynamicSourceArg(argUsername, flagNameFile, fmt.Sprintf("%s_NAME", flagEnvPrefix))
		if err != nil {
			utils.PrintError(fmt.Errorf("failed to retreive username: %w", err))
			return nil
		}

		var password string
		if flagPasswordPrompt || flagPasswordStdin {
			password, err = ReadPassword(flagPasswordStdin)
			if err != nil {
				utils.PrintErrorWithMessage("failed to read password", err)
				return nil
			}
		} else {
			password, err = ResolveDynamicSourceArg("", flagPasswordFile, fmt.Sprintf("%s_PASSWORD", flagEnvPrefix))
			if err != nil {
				utils.PrintError(fmt.Errorf("failed to retreive password: %w", err))
				return nil
			}
		}

		dbConn, err := db.Connect()
		if err != nil {
			utils.PrintErrorWithMessage("Failed to connect to database", err)
			return nil
		}
		defer dbConn.Close()

		dbConfig := db.GetConfig()

		err = schema.EnsureUser(dbConn, dbConfig.DBName, userType, schema.User{
			Name:     userName,
			Password: password,
		})
		if err != nil {
			utils.PrintErrorWithMessage("Failed to sync database user", err)
			return nil
		}

		utils.PrintSuccess("Database user synced successfully")
		return nil
	},
}

func init() {
	SchemaEnsureUserCmd.Flags().StringP("type", "t", "", "User type: 'manager', 'postfix', 'dovecot' or 'stalwart'")
	SchemaEnsureUserCmd.Flags().BoolP("password", "p", false, "Set password interactively (prompts)")
	SchemaEnsureUserCmd.Flags().Bool("password-stdin", false, "Read password from stdin")
	SchemaEnsureUserCmd.Flags().String("type-file", "", "Read user type from file")
	SchemaEnsureUserCmd.Flags().String("name-file", "", "Read username from file")
	SchemaEnsureUserCmd.Flags().String("password-file", "", "Read password from file")
	SchemaEnsureUserCmd.Flags().String("env-prefix", "MAILCTL_USER", "Prefix for environment variables")
}
