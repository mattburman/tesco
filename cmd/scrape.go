package cmd

import (
	"github.com/spf13/cobra"
)

var concurrency int

var ScrapeCmd = &cobra.Command{
	Use:   "scrape <type>",
	Short: "scrape tesco urls and persist the data",
}

func init() {
	ScrapeCmd.PersistentFlags().IntVar(&concurrency, "concurrency", 3, "number of simultaneous requests")
	RootCmd.AddCommand(ScrapeCmd)
}
