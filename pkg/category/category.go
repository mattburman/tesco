// Package category implements functions to request and manipulate category data from tesco
package category

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/mattburman/tesco/internal/category"
	"github.com/mattburman/tesco/pkg/collecting"
	"github.com/mattburman/tesco/pkg/product"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
)

type ProductResult struct {
	Id   string
	Url  string
	Json string
}

// Get takes a product category page and returns the data
// or an error for parameter, network or request failures
func Get(url string) (*string, error) {
	url, err := category.AddCountToURL(url)
	if err != nil {
		return nil, fmt.Errorf("unable to parse url: %v", err)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request error: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("request error: %v", err)
	}

	resources, err := product.ExtractResources(string(body))
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

func Scrape(url string, concurrency int, productResults chan ProductResult, db *sql.DB) error {
	url, err := category.AddCountToURL(url)
	if err != nil {
		return fmt.Errorf("unable to parse url: %v", err)
	}

	productCollector := colly.NewCollector(
		colly.Async(true),
	)
	productCollector.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: concurrency})
	productCollector.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})
	productCollector.OnHTML("[data-props]", func(e *colly.HTMLElement) {
		productJson, err := collecting.ExtractResources(e.Attr("data-props"))
		if err != nil {
			fmt.Printf("error extracting resources from data-props: %v\n", err)
			return
		}
		url := e.Request.URL.String()
		id, err := product.URLToID(url)
		if err != nil {
			fmt.Printf("could not get id from url: %v", err)
		}
		productResults <- ProductResult{Id: id, Url: url, Json: *productJson}
	})

	categoryCollector := colly.NewCollector(
		colly.Async(true),
	)
	categoryCollector.OnHTML("[data-props]", func(e *colly.HTMLElement) {
		categoryJson, err := collecting.ExtractResources(e.Attr("data-props"))
		if err != nil {
			fmt.Printf("error extracting resources from data-props: %v\n", err)
		}

		productIDs, err := ToProductIDs(categoryJson)
		if err != nil {
			fmt.Printf("error extracting productIDs: %v\n", err)
			return
		}

		unfetchedProductIDs, err := product.GetUnfetchedProductIDs(db, productIDs)
		if err != nil {
			fmt.Printf("failed to get unfetched productResults from DB: %v", err)
		}

		for _, productID := range *unfetchedProductIDs {
			productCollector.Visit(fmt.Sprintf("https://www.tesco.com/groceries/en-GB/products/%v", productID))
		}
		productCollector.Wait()
	})
	categoryCollector.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	categoryCollector.Visit(url)
	categoryCollector.Wait()
	close(productResults)
	return nil
}

// ToProductIDs takes a product category result JSON string and returns extracted product IDs
func ToProductIDs(category *string) (*[]string, error) {
	ids := gjson.Get(*category, "productsByCategory.data.results.productItems.#.product.id")
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

