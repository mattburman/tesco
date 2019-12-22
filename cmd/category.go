package cmd

import (
	"database/sql"
	"fmt"
	"github.com/mattburman/tesco/pkg/category"
	"github.com/mattburman/tesco/pkg/product"
	"sync"

	"github.com/spf13/cobra"
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
		url := args[0]

		db, err := sql.Open("sqlite3", "./data.db")
		if err != nil {
			return fmt.Errorf("failed to open db: %v", err)
		}
		defer db.Close()
		err = db.Ping()
		if err != nil {
			return fmt.Errorf("unable to connect to db: %v", err)
		}
		insertProduct, err := db.Prepare("INSERT INTO products(id, source, raw) VALUES(?, 'product', ?)")
		if err != nil {
			return fmt.Errorf("failed to create prepared statement for products: %v", err)
		}

		data, err := category.Get(url)
		if err != nil {
			return fmt.Errorf("failed to get category: %v", err)
		}
		productIDs, err := category.ToProductIDs(data)
		if err != nil {
			return fmt.Errorf("failed to extract products from category")
		}

		unfetchedProductIDs, err := product.GetUnfetchedProductIDs(db, productIDs)
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
					product, err := product.GetProduct(t.id)
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
				fmt.Printf("inserted product: %v\n", lastID)
			}(r)
		}

		fmt.Printf("results: %v\n", results)

		return nil
	},
}

func init() {
	categoryCmd.Flags().IntVar(&concurrency, "concurrency", 3, "number of simultaneous requests")
	RootCmd.AddCommand(categoryCmd)
}
