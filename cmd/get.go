package cmd

import (
	"github.com/spf13/cobra"
)

var GetCmd = &cobra.Command{
	Use:   "get <type>",
	Short: "get tesco urls and output the data",
}

func init() {
	RootCmd.AddCommand(GetCmd)
}
