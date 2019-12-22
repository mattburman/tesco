package product

import (
	"database/sql"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

var (
	productf          string         = "https://www.tesco.com/groceries/en-GB/products/%v"
	dataRegexp        *regexp.Regexp = regexp.MustCompile(`data-props="({.*})"`)
	invalidProductIDf string         = "%v is an invalid productID"
)

// GetProduct returns the product data
// or an error for parameter, network or request failures
func GetProduct(id string) (*string, error) {
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

	resources, err := ExtractResources(string(body))
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

// ExtractResources takes a Tesco HTML response body and returns the resources json from data-props
func ExtractResources(body string) (*string, error) {
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

// GetUnfetchedProductIDs returns the productIDs supplied that do not exist in the DB
func GetUnfetchedProductIDs(db *sql.DB, toFetchProductIDs *[]string) (*[]string, error) {
	numIDs := len(*toFetchProductIDs)
	placeholders := strings.TrimSuffix(strings.Repeat("?,", numIDs), ",")
	sql := fmt.Sprintf("SELECT id FROM products WHERE id IN(%v) AND source='product'", placeholders)

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
		if i < len(existingProducts) && existingProducts[i] == pid { // product exists in db
			continue
		}
		unfetchedIDs = append(unfetchedIDs, pid)
	}

	return &unfetchedIDs, nil
}

