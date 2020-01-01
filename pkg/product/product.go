package product

import (
	"crypto/sha1"
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
	perGrams          *regexp.Regexp = regexp.MustCompile(`^Per (?P<grams>\d+)g$`)
	perServingGrams   *regexp.Regexp = regexp.MustCompile(`.*\((?P<grams>\d+)g\).*`)
	floatGrams        *regexp.Regexp = regexp.MustCompile(`(?P<grams>\d+(\.\d+)?)g`)
	floatKcal         *regexp.Regexp = regexp.MustCompile(`\d+kJ / (?P<kcal>\d+)kcal`)
)

type Macros struct {
	per     string
	size    float64
	carbs   float64
	protein float64
	fat     float64
	kcal    float64
}

type Source struct {
	url  string
	id   string
	name string
}

type Product struct {
	name                            string
	source                          Source
	description                     []string
	raw                             string
	hashOfRawValueLastUsedToCompute string
	perComp                         Macros
	perServing                      Macros
}

// NewProduct constructs a Product from a raw tesco json response string
func NewProduct(raw string, url string) (*Product, error) {
	results := gjson.GetMany(
		raw,
		"pageTitle",
		"product.description",
		"product.details.nutritionInfo.#.name",
		"product.details.nutritionInfo.#.perComp",
		"product.details.nutritionInfo.#.perServing",
	)
	name := results[0].String()

	id, err := URLToID(url)
	if err != nil {
		return nil, fmt.Errorf("unable to extract ID from %v: %v", url, err)
	}

	source := Source{url: url, id: id, name: "tesco"}

	descriptionResults := results[1].Array()
	description := make([]string, len(descriptionResults))
	for i, result := range descriptionResults {
		description[i] = result.String()
	}

	h := sha1.New()
	h.Write([]byte(raw))
	hashOfRawValueLastUsedToCompute := fmt.Sprintf("%x", h.Sum(nil))

	nutrientNames := results[2].Array()
	nutrientPerComps := results[3].Array()
	nutrientPerServings := results[4].Array()

	nutrientIndices := make(map[string]int)
	for i, nutrient := range nutrientNames {
		nutrientIndices[nutrient.String()] = i
	}

	perComp := Macros{}
	perServing := Macros{}

	i := nutrientIndices["Typical Values"]
	perComp.per = nutrientPerComps[i].String()
	perServing.per = nutrientPerServings[i].String()
	match := perGrams.FindSubmatch([]byte(perComp.per))
	if len(match) == 2 {
		perComp.size, _ = strconv.ParseFloat(string(match[1]), 64)
	}
	match = perServingGrams.FindSubmatch([]byte(perServing.per))
	if len(match) == 2 {
		perServing.size, _ = strconv.ParseFloat(string(match[1]), 64)
	}

	i = nutrientIndices["Fat"]
	match = floatGrams.FindSubmatch([]byte(nutrientPerComps[i].String()))
	if len(match) > 1 {
		perComp.fat, _ = strconv.ParseFloat(string(match[1]), 64)
	}
	match = floatGrams.FindSubmatch([]byte(nutrientPerServings[i].String()))
	if len(match) > 1 {
		perServing.fat, _ = strconv.ParseFloat(string(match[1]), 64)
	}

	i = nutrientIndices["Protein"]
	match = floatGrams.FindSubmatch([]byte(nutrientPerComps[i].String()))
	if len(match) > 1 {
		perComp.protein, _ = strconv.ParseFloat(string(match[1]), 64)
	}
	match = floatGrams.FindSubmatch([]byte(nutrientPerServings[i].String()))
	if len(match) > 1 {
		perServing.protein, _ = strconv.ParseFloat(string(match[1]), 64)
	}

	i = nutrientIndices["Carbohydrate"]
	match = floatGrams.FindSubmatch([]byte(nutrientPerComps[i].String()))
	if len(match) > 1 {
		perComp.carbs, _ = strconv.ParseFloat(string(match[1]), 64)
	}
	match = floatGrams.FindSubmatch([]byte(nutrientPerServings[i].String()))
	if len(match) > 1 {
		perServing.carbs, _ = strconv.ParseFloat(string(match[1]), 64)
	}

	i = nutrientIndices["Energy"]
	match = floatKcal.FindSubmatch([]byte(nutrientPerComps[i].String()))
	fmt.Println(string(match[1]))
	if len(match) > 1 {
		perComp.kcal, _ = strconv.ParseFloat(string(match[1]), 64)
	}
	match = floatKcal.FindSubmatch([]byte(nutrientPerServings[i].String()))
	if len(match) > 1 {
		perServing.kcal, _ = strconv.ParseFloat(string(match[1]), 64)
	}

	product := Product{
		name:                            name,
		source:                          source,
		description:                     description,
		raw:                             raw,
		hashOfRawValueLastUsedToCompute: hashOfRawValueLastUsedToCompute,
		perComp:                         perComp,
		perServing:                      perServing,
	}

	return &product, nil
}

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

var urlRegex *regexp.Regexp = regexp.MustCompile(`https://www.tesco.com/groceries/en-GB/products/(?P<ID>\d+)`)

// URLToID extracts the ID from a product URL
func URLToID(url string) (string, error) {
	match := urlRegex.FindSubmatch([]byte(url))
	if len(match) == 0 {
		return "", fmt.Errorf("could not extract id from url: %v", url)
	}
	return string(match[1]), nil
}
