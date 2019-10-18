// tesco is a cli for accessing tesco macronutrient data
package main

import (
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"

	"github.com/tidwall/gjson"
	"github.com/urfave/cli"
)

var productf string = "https://www.tesco.com/groceries/en-GB/products/%v"
var dataRegexp *regexp.Regexp = regexp.MustCompile(`data-props="({.*})"`)

var invalidProductIDf string = "%v is an invalid productID"

func main() {
	app := cli.NewApp()

	app.Name = "tesco"
	app.Usage = "query the tesco site"

	app.Commands = []cli.Command{
		{
			Name:    "product",
			Aliases: []string{"p"},
			Usage:   "get product by product ID",
			Flags: []cli.Flag{
				cli.Int64Flag{
					Name:     "pid",
					Usage:    "Product ID of the product",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				pid := c.Int64("pid")
				if pid < 100000000 {
					return fmt.Errorf(invalidProductIDf, pid)
				}

				data, err := getProduct(pid)
				if err != nil {
					return err
				}
				fmt.Println(*data)

				return nil
			},
		},
		{
			Name:    "category",
			Aliases: []string{"c"},
			Usage:   "get category data by url",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "url",
					Usage:    "A category URL",
					Required: true,
				},
				cli.IntFlag{
					Name:  "concurrency",
					Usage: "Concurrent number of GET requests",
					Value: 1,
				},
			},
			Action: func(c *cli.Context) error {
				url := c.String("url")
				concurrency := c.Int("concurrency")

				data, err := getCategory(url)
				if err != nil {
					return fmt.Errorf("failed to get category: %v", err)
				}
				productIDs, err := categoryToProductIDs(data)
				if err != nil {
					return fmt.Errorf("failed to extract products from category")
				}
				fmt.Println(productIDs)
				fmt.Println(concurrency)

				return nil
			},
		},
	}

	app.OnUsageError = func(c *cli.Context, err error, isSubcommand bool) error {
		cli.ShowCommandHelp(c, c.Command.Name)
		return nil
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
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

// categoryToProductIDs takes a tesco category JSON string and returns extracted product IDs
func categoryToProductIDs(category *string) (*[]string, error) {
	ids := gjson.Get(*category, "results.productItems.#.product.id")
	if !ids.Exists() {
		return nil, fmt.Errorf("unable to extract product ids from category")
	}

	idCount := gjson.Get(ids.String(), "#")
	idSlice := make([]string, idCount.Int())

	ids.ForEach(func(key, value gjson.Result) bool {
		idSlice = append(idSlice, value.String())
		return true
	})

	return &idSlice, nil
}

// getCategory takes a tesco category page and returns the data
// or an error for parameter, network or request failures
func getCategory(url string) (*string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request error: %v", err)
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

	data := gjson.Get(*resources, "productsByCategory.data|@pretty")
	if !data.Exists() {
		return nil, errors.New("unable to access products data in resources")
	}
	s := data.String()

	return &s, nil
}

// getProduct returns the product data
// or an error for parameter, network or request failures
func getProduct(id int64) (*string, error) {
	if id < 100000000 {
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
