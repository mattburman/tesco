// tesco is a cli for accessing tesco macronutrient data
package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
)

func main() {
	var productID int64
	flag.Int64Var(&productID, "pid", 0, "https://www.tesco.com/groceries/en-GB/products/<productID>")
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
	return strconv.FormatInt(id, 10), nil
}
