package cmd

import (
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/schema"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var SchemaDropUserCmd = &cobra.Command{
	Use:   "drop-user [username]",
	Short: "Drop an application-specific database user",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flagNameFile, _ := cmd.Flags().GetString("name-file")
		flagEnvPrefix, _ := cmd.Flags().GetString("env-prefix")

		argUsername := ""
		if len(args) > 0 {
			argUsername = args[0]
		}

		userName, err := ResolveDynamicSourceArg(argUsername, flagNameFile, fmt.Sprintf("%s_NAME", flagEnvPrefix))
		if err != nil {
			utils.PrintError(fmt.Errorf("failed to resolve username: %w", err))
			return nil
		}

		dbConn, err := db.Connect()
		if err != nil {
			utils.PrintErrorWithMessage("Failed to connect to database", err)
			return nil
		}
		defer dbConn.Close()

		dbConfig := db.GetConfig()

		err = schema.DropUser(dbConn, dbConfig.DBName, dbConfig.User, userName)
		if err != nil {
			utils.PrintErrorWithMessage("Failed to drop database user", err)
			return nil
		}

		utils.PrintSuccess("Database user dropped successfully")
		return nil
	},
}

func init() {
	SchemaDropUserCmd.Flags().String("name-file", "", "Read username from file")
	SchemaDropUserCmd.Flags().String("env-prefix", "MAILCTL_USER", "Prefix for environment variable")
}
