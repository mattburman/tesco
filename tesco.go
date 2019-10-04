// tesco is a cli for accessing tesco macronutrient data
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

var productf string = "https://www.tesco.com/groceries/en-GB/products/%v"
var dataRegexp *regexp.Regexp = regexp.MustCompile(`data-redux-state="(\[.*\])"`)

func main() {
	var productID int64
	flag.Int64Var(&productID, "pid", 0, productf)
	flag.Parse()
	res, err := getProduct(productID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)
}

// getProduct returns the product data
// or an error for parameter, network or request failures
func getProduct(id int64) (string, error) {
	if id < 100000000 {
		return "", fmt.Errorf("getProduct: %v is an invalid productID", id)
	}

	productURL := fmt.Sprintf(productf, id)

	resp, err := http.Get(productURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	matches := dataRegexp.FindStringSubmatch(string(body))

	if len(matches) < 2 {
		return "", fmt.Errorf("getProduct: failed to extract data for %v", id)
	}

	data := matches[1]

	return data, nil
}
