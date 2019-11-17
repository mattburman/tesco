package cmd

import (
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

var (
	productf          string         = "https://www.tesco.com/groceries/en-GB/products/%v"
	dataRegexp        *regexp.Regexp = regexp.MustCompile(`data-props="({.*})"`)
	invalidProductIDf string         = "%v is an invalid productID"
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
		data, err := getProduct(productID)
		if err != nil {
			return fmt.Errorf("failed to get product: %v", err)
		}
		fmt.Println(*data)
		return nil
	},
}

// getProduct returns the product data
// or an error for parameter, network or request failures
func getProduct(id string) (*string, error) {
	idint, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("product ID was not an integer: %v", err)
	}
	if idint < 100000000 {
		return nil, fmt.Errorf(invalidProductIDf, id)
	}

	productURL := fmt.Sprintf(productf, id)

	resp, err := http.Get(productURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("request error: %v", err)
	}

	resources, err := extractResources(string(body))
	if err != nil {
		return nil, fmt.Errorf("unable to extract resources: %v", err)
	}

	data := gjson.Get(*resources, "productDetails.data")
	if !data.Exists() {
		return nil, errors.New("unable to extract productDetails.data")
	}

	s := data.String()

	return &s, nil
}

// extractResources takes a Tesco HTML response body and returns the resources json from data-props
func extractResources(body string) (*string, error) {
	matches := dataRegexp.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return nil, fmt.Errorf("failed to extract data from data-props")
	}
	json := html.UnescapeString(matches[1])

	if !gjson.Valid(json) {
		return nil, errors.New("invalid json")
	}

	if err := gjson.Get(json, "error"); err.Exists() {
		return nil, fmt.Errorf("error returned from Tesco: %v", err.String())
	}

	resources := gjson.Get(json, "resources").String()

	return &resources, nil
}

func init() {
	RootCmd.AddCommand(productCmd)
}
