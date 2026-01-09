package cmd

import (
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/schema"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var SchemaUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade database schema to latest version",
	Long:  `Upgrade the database schema to the latest available version by applying pending migrations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		latestVersion, err := schema.GetLatestAvailableVersion()
		if err != nil {
			utils.PrintErrorWithMessage("Failed to get latest available schema version", err)
			return nil
		}

		dbConn, err := db.Connect()
		if err != nil {
			utils.PrintErrorWithMessage("Failed to connect to database", err)
			return nil
		}
		defer dbConn.Close()

		currentVersion, err := schema.GetCurrentVersion(dbConn)
		if err != nil {
			utils.PrintErrorWithMessage("Failed to get current schema version", err)
			return nil
		}

		if currentVersion >= latestVersion {
			utils.PrintSuccess("Schema is already up to date")
			return nil
		}

		if currentVersion == 0 {
			fmt.Printf("Installing schema v%d...\n", latestVersion)
		} else {
			fmt.Printf("Upgrading schema: v%d -> v%d...\n", currentVersion, latestVersion)
		}

		err = schema.Upgrade(dbConn, latestVersion)
		if err != nil {
			utils.PrintErrorWithMessage("Failed to upgrade schema", err)
			return nil
		}

		utils.PrintSuccess("Schema upgrade completed successfully")
		return nil
	},
}
