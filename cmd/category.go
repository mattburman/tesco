package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/tidwall/gjson"
)

var concurrency int

var categoryCmd = &cobra.Command{
	Use:   "category <url>",
	Short: "scrape category by URL and persist to data.db",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("No URL supplied")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(concurrency)
		url := args[0]
		fmt.Println(url)

		url, err := addCountToURL(url)
		if err != nil {
			return fmt.Errorf("unable to parse url: %v", err)
		}

		db, err := sql.Open("sqlite3", "./data.db")
		if err != nil {
			return fmt.Errorf("failed to open db: %v", err)
		}
		defer db.Close()
		err = db.Ping()
		if err != nil {
			return fmt.Errorf("unable to connect to db: %v", err)
		}
		insertProduct, err := db.Prepare("INSERT INTO products(id, source, raw) VALUES(?, 'tesco', ?)")
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

		unfetchedProductIDs, err := getUnfetchedProductIDs(db, productIDs)
		if err != nil {
			return fmt.Errorf("failed to get unfetched products from DB: %v", err)
		}
		fmt.Printf("unfetchedProductIDs: %v\n", unfetchedProductIDs)

		type task struct {
			id string
		}
		type result struct {
			raw string
			id  string
		}
		tasks := make(chan task)
		go func() {
			for _, id := range *unfetchedProductIDs {
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
					wg.Add(1)
				}
			}()
		}

		for r := range results {
			go func(reqResult result) {
				defer wg.Done()
				res, err := insertProduct.Exec(reqResult.id, reqResult.raw)
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
}

// getUnfetchedProductIDs returns the productIDs supplied that do not exist in the DB
func getUnfetchedProductIDs(db *sql.DB, toFetchProductIDs *[]string) (*[]string, error) {
	numIDs := len(*toFetchProductIDs)
	placeholders := strings.TrimSuffix(strings.Repeat("?,", numIDs), ",")
	sql := fmt.Sprintf("SELECT id FROM products WHERE id IN(%v) AND source='tesco'", placeholders)

	ids := make([]interface{}, numIDs)
	for i := range ids {
		ids[i] = (*toFetchProductIDs)[i]
	}
	rows, err := db.Query(sql, ids...)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing products from DB: %v", err)
	}

	var id string
	existingProducts := make([]string, 0, numIDs)
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("failed to scan line of product: %v", err))
		}
		existingProducts = append(existingProducts, id)
	}

	sort.Strings(existingProducts)
	unfetchedIDs := make([]string, 0, numIDs)
	for _, pid := range *toFetchProductIDs {
		i := sort.SearchStrings(existingProducts, pid)
		fmt.Println(i, len(existingProducts))
		if i < len(existingProducts) && existingProducts[i] == pid { // product exists in db
			continue
		}
		unfetchedIDs = append(unfetchedIDs, pid)
	}

	return &unfetchedIDs, nil
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

// categoryToProductIDs takes a tesco category JSON string and returns extracted product IDs
func categoryToProductIDs(category *string) (*[]string, error) {
	ids := gjson.Get(*category, "results.productItems.#.product.id")
	if !ids.Exists() {
		return nil, fmt.Errorf("unable to extract product ids from category")
	}

	idCount := gjson.Get(ids.String(), "#")
	idSlice := make([]string, idCount.Int())

	i := 0
	ids.ForEach(func(key, value gjson.Result) bool {
		idSlice[i] = value.String()
		i++
		return true
	})

	return &idSlice, nil
}

// addCountToURL returns the passed url with a count=48 query parameter
func addCountToURL(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("%v was not a valid URL: %v", u, err)
	}
	q := parsed.Query()
	q.Set("count", "48")
	parsed.RawQuery = q.Encode()
	u = fmt.Sprint(parsed)
	return u, nil
}

func init() {
	categoryCmd.Flags().IntVar(&concurrency, "concurrency", 3, "number of simultaneous requests")
	RootCmd.AddCommand(categoryCmd)
}
