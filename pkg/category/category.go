// Package category implements functions to request and manipulate category data from tesco
package category

import (
	"errors"
	"fmt"
	"github.com/mattburman/tesco/internal/category"
	"github.com/mattburman/tesco/pkg/product"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
)

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

// ToProductIDs takes a product category JSON string and returns extracted product IDs
func ToProductIDs(category *string) (*[]string, error) {
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

