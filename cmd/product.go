package cmd

import (
	"fmt"
	"github.com/mattburman/tesco/pkg/product"
	"strconv"

	"github.com/spf13/cobra"
)

var productCmd = &cobra.Command{
	Use:   "product <id>",
	Short: "get product by product ID",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("No product ID supplied")
		}
		idint, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("product ID was not an integer: %v", err)
		}
		if idint < 100000000 {
			return fmt.Errorf("Product ID %v is less than 100000000", idint)
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		productID := args[0]
		data, err := product.GetProduct(productID)
		if err != nil {
			return fmt.Errorf("failed to get product: %v", err)
		}
		fmt.Println(*data)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(productCmd)
}
