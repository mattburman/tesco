// tesco is a cli for accessing tesco macronutrient data
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli"
)

var (
	productf          string         = "https://www.tesco.com/groceries/en-GB/products/%v"
	dataRegexp        *regexp.Regexp = regexp.MustCompile(`data-props="({.*})"`)
	invalidProductIDf string         = "%v is an invalid productID"
)

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

				db, err := sql.Open("sqlite3", "./data.db")
				if err != nil {
					return fmt.Errorf("failed to open db: %v", err)
				}
				defer db.Close()
				err = db.Ping()
				if err != nil {
					return fmt.Errorf("unable to connect to db: %v", err)
				}
				insertProduct, err := db.Prepare("INSERT INTO products(id, source) VALUES(?, 'tesco')")
				if err != nil {
					return fmt.Errorf("failed to create prepared statement for products: %v", err)
				}

				data, err := getCategory(url)
				if err != nil {
					return fmt.Errorf("failed to get category: %v", err)
				}
				productIDs, err := categoryToProductIDs(data)
				if err != nil {
					return fmt.Errorf("failed to extract products from category")
				}

				type task struct {
					id int64
				}
				type result struct {
					raw string
					id  int64
				}
				tasks := make(chan task)
				go func() {
					for _, id := range *productIDs {
						fmt.Println(id)
						tasks <- task{id: id}
					}
					close(tasks)
				}()

				results := make(chan result)
				var wg sync.WaitGroup
				wg.Add(concurrency)
				go func() {
					wg.Wait()
					close(results)
				}()

				for i := 0; i < concurrency; i++ {
					go func() {
						defer wg.Done()
						for t := range tasks {
							product, err := getProduct(t.id)
							if err != nil {
								fmt.Printf("could not get product data for %v: %v\n", t.id, err)
								continue
							}
							results <- result{raw: *product, id: t.id}
						}
					}()
				}

				for r := range results {
					go func(reqResult result) {
						res, err := insertProduct.Exec(reqResult.id)
						if err != nil {
							fmt.Printf("failed to execute query for %v: %v\n", reqResult.id, err)
							return
						}
						rowCnt, err := res.RowsAffected()
						if err != nil {
							fmt.Printf("failed to get RowsAffected for attempted insertion of %v: %v\n", reqResult.id, err)
							return
						}
						if rowCnt == 0 {
							fmt.Printf("no insertion for %v", reqResult.id)
							return
						}
						lastID, err := res.LastInsertId()
						if err != nil {
							fmt.Printf("failed to get LastInsertId for attempted insertion of %v: %v\n", reqResult.id, err)
							return
						}
						fmt.Printf("inserted product %v\n", lastID)
					}(r)
				}
				fmt.Printf("results: %v\n", results)

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
func categoryToProductIDs(category *string) (*[]int64, error) {
	ids := gjson.Get(*category, "results.productItems.#.product.id")
	if !ids.Exists() {
		return nil, fmt.Errorf("unable to extract product ids from category")
	}

	idCount := gjson.Get(ids.String(), "#")
	idSlice := make([]int64, idCount.Int())

	i := 0
	ids.ForEach(func(key, value gjson.Result) bool {
		idSlice[i] = value.Int()
		i++
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
