// tesco is a cli for accessing tesco macronutrient data
package main

import (
	"flag"
	"fmt"
)

func main() {
	var productID int64
	flag.Int64Var(&productID, "pid", 0, "https://www.tesco.com/groceries/en-GB/products/<productID>")
	flag.Parse()
	fmt.Println(productID)
}
