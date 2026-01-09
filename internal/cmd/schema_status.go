package cmd

import (
	"fmt"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/schema"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var SchemaStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current schema version",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbConn, err := db.Connect()
		if err != nil {
			utils.PrintErrorWithMessage("failed to connect to database", err)
			return nil
		}
		defer dbConn.Close()

		currentVersion, err := schema.GetCurrentVersion(dbConn)
		if err != nil {
			utils.PrintErrorWithMessage("Failed to get current schema version", err)
			return nil
		}

		latestVersion, err := schema.GetLatestAvailableVersion()
		if err != nil {
			utils.PrintErrorWithMessage("Failed to get latest available schema version", err)
			return nil
		}

		if currentVersion > 0 && currentVersion < latestVersion {
			fmt.Println(utils.YellowStyle.Bold(true).Render("Schema upgrade available"))
			fmt.Printf("Installed schema: v%d\n", currentVersion)
			fmt.Printf("Available schema: v%d\n", latestVersion)
		} else if currentVersion > 0 && currentVersion == latestVersion {
			utils.PrintSuccess("Schema is up to date")
			fmt.Printf("Installed schema: v%d\n", currentVersion)
		} else {
			fmt.Println(utils.YellowStyle.Bold(true).Render("Schema is not installed"))
			fmt.Printf("Available schema: v%d\n", latestVersion)
		}
		return nil
	},
}
