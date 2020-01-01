// Package category implements functions to request and manipulate category data from tesco
package category

import (
	"database/sql"
	"fmt"
	"github.com/mattburman/tesco/pkg/category"
)

var Get = category.Get

// ScrapeToSqlite scrapes the products in a category URL to sqlite3
func ScrapeToSqlite(url string, concurrency int) error {
	// set up db
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
		return fmt.Errorf("failed to create prepared statement for productResults: %v", err)
	}

	// the channel we will receive products on
	productResults := make(chan category.ProductResult)

	// start some insertion workers that take insertion jobs from the channel
	for i := 1; i <= 8; i++ {
		go func() {
			for reqResult := range productResults {
				fmt.Println(reqResult.Id)
				res, err := insertProduct.Exec(reqResult.Id, reqResult.Json)
				if err != nil {
					fmt.Printf("failed to execute query for %v: %v\n", reqResult.Id, err)
					continue
				}
				rowCnt, err := res.RowsAffected()
				if err != nil {
					fmt.Printf("failed to get RowsAffected for attempted insertion of %v: %v\n", reqResult.Id, err)
					continue
				}
				if rowCnt == 0 {
					fmt.Printf("no insertion for %v", reqResult.Id)
					continue
				}
				lastID, err := res.LastInsertId()
				if err != nil {
					fmt.Printf("failed to get LastInsertId for attempted insertion of %v: %v\n", reqResult.Id, err)
					continue
				}
				fmt.Printf("inserted product: %v\n", lastID)
			}
		}()
	}

	// scrape the category to place products on the productResults channel
	err = category.Scrape(url, concurrency, productResults, db)
	if err != nil {
		return fmt.Errorf("failed to scrape productResults: %v", err)
	}

	return nil
}
