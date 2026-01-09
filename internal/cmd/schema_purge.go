package cmd

import (
	"fmt"
	"os"

	"github.com/gerolf-vent/mailctl/internal/db"
	"github.com/gerolf-vent/mailctl/internal/schema"
	"github.com/gerolf-vent/mailctl/internal/utils"
	"github.com/spf13/cobra"
)

var SchemaPurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge all schema and data from the database",
	Long:  "Purge the entire database by dropping all tables, sequences, functions, and types.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Require confirmation flag
		confirm, err := cmd.Flags().GetBool("confirm")
		if err != nil {
			utils.PrintErrorWithMessage("Failed to read confirmation flag", err)
			return nil
		}
		if !confirm {
			fmt.Printf("This will %s:\n", utils.RedStyle.Render("PERMANENTLY DELETE"))
			fmt.Println("  • All tables and their data")
			fmt.Println("  • All sequences")
			fmt.Println("  • All functions")
			fmt.Println("  • All custom types")
			fmt.Println("To proceed, use '--confirm'")
			os.Exit(1)
		}

		dbConn, err := db.Connect()
		if err != nil {
			utils.PrintErrorWithMessage("Failed to connect to database", err)
			return nil
		}
		defer dbConn.Close()

		fmt.Println("Purging database...")

		err = schema.Purge(dbConn)
		if err != nil {
			utils.PrintErrorWithMessage("Failed to purge database", err)
			return nil
		}

		utils.PrintSuccess("Schema purge completed successfully")
		return nil
	},
}

func init() {
	SchemaPurgeCmd.Flags().BoolP("confirm", "c", false, "Confirm database purge (required)")
}
