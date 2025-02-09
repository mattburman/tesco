package cmd

import (
	"fmt"
	"github.com/mattburman/tesco/internal/category"
	"github.com/spf13/cobra"
)

var scrapeCategoryCmd = &cobra.Command{
	Use:   "category <url>",
	Short: "scrape category by URL and persist to data.db",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("No URL supplied")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]
		err := category.ScrapeToSqlite(url, concurrency)
		if err != nil {
			return err
		}
		return nil
	},
}

var getCategoryCmd = &cobra.Command{
	Use:   "category <url>",
	Short: "get category by URL",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("No URL supplied")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]
		category, err := category.Get(url)
		if err != nil {
			return err
		}
		fmt.Println(*category)
		return nil
	},
}

func init() {
	ScrapeCmd.AddCommand(scrapeCategoryCmd)
	GetCmd.AddCommand(getCategoryCmd)
}
