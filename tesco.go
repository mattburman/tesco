// tesco is a cli for accessing tesco macronutrient data
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"

	"github.com/mattburman/mappp"
	"github.com/urfave/cli"
)

type ProductData struct {
	AisleID               string        `json:"aisleId"`
	AisleName             string        `json:"aisleName"`
	AlternativeCategories []interface{} `json:"alternativeCategories"`
	AsyncPage             bool          `json:"asyncPage"`
	BackToURL             string        `json:"backToUrl"`
	Breadcrumbs           []struct {
		Label  string `json:"label"`
		LinkTo string `json:"linkTo,omitempty"`
		CatID  string `json:"catId,omitempty"`
	} `json:"breadcrumbs"`
	CanonicalURL   string `json:"canonicalUrl"`
	DepartmentID   string `json:"departmentId"`
	DepartmentName string `json:"departmentName"`
	Experiments    struct {
		Oop507 string `json:"oop507"`
	} `json:"experiments"`
	PageDescription string `json:"pageDescription"`
	PageTitle       string `json:"pageTitle"`
	Product         struct {
		AisleName           string      `json:"aisleName"`
		AverageWeight       interface{} `json:"averageWeight"`
		BaseProductID       string      `json:"baseProductId"`
		BrandName           string      `json:"brandName"`
		BulkBuyLimit        int         `json:"bulkBuyLimit"`
		BulkBuyLimitGroupID interface{} `json:"bulkBuyLimitGroupId"`
		BulkBuyLimitMessage string      `json:"bulkBuyLimitMessage"`
		CatchWeightList     interface{} `json:"catchWeightList"`
		DefaultImageURL     string      `json:"defaultImageUrl"`
		DepartmentName      string      `json:"departmentName"`
		DepositAmount       interface{} `json:"depositAmount"`
		Description         []string    `json:"description"`
		Details             struct {
			Additives           interface{} `json:"additives"`
			AlcoholInfo         interface{} `json:"alcoholInfo"`
			AllergenInfo        interface{} `json:"allergenInfo"`
			BoxContents         interface{} `json:"boxContents"`
			BrandMarketing      interface{} `json:"brandMarketing"`
			CookingInstructions struct {
				CookingGuidelines []interface{} `json:"cookingGuidelines"`
				CookingMethods    []struct {
					Instructions []string `json:"instructions"`
					Name         string   `json:"name"`
					Time         string   `json:"time"`
				} `json:"cookingMethods"`
				CookingPrecautions []string `json:"cookingPrecautions"`
				Microwave          struct {
					Chilled struct {
						Detail       interface{} `json:"detail"`
						Instructions []string    `json:"instructions"`
					} `json:"chilled"`
					Frozen struct {
						Detail       interface{}   `json:"detail"`
						Instructions []interface{} `json:"instructions"`
					} `json:"frozen"`
				} `json:"microwave"`
				OtherInstructions []interface{} `json:"otherInstructions"`
				Oven              struct {
					Chilled struct {
						Instructions []interface{} `json:"instructions"`
						Temperature  interface{}   `json:"temperature"`
						Time         interface{}   `json:"time"`
					} `json:"chilled"`
					Frozen struct {
						Instructions []interface{} `json:"instructions"`
						Temperature  interface{}   `json:"temperature"`
						Time         interface{}   `json:"time"`
					} `json:"frozen"`
				} `json:"oven"`
			} `json:"cookingInstructions"`
			Directions           interface{} `json:"directions"`
			Dosage               interface{} `json:"dosage"`
			DrainedWeight        interface{} `json:"drainedWeight"`
			Features             interface{} `json:"features"`
			FreezingInstructions interface{} `json:"freezingInstructions"`
			GuidelineDailyAmount struct {
				DailyAmounts []struct {
					Name    string `json:"name"`
					Percent string `json:"percent"`
					Rating  string `json:"rating"`
					Value   string `json:"value"`
				} `json:"dailyAmounts"`
				Title interface{} `json:"title"`
			} `json:"guidelineDailyAmount"`
			HazardInfo struct {
				ChemicalName string        `json:"chemicalName"`
				ProductName  string        `json:"productName"`
				SignalWord   string        `json:"signalWord"`
				Statements   []string      `json:"statements"`
				SymbolCodes  []interface{} `json:"symbolCodes"`
			} `json:"hazardInfo"`
			HealthClaims          interface{}   `json:"healthClaims"`
			Healthmark            interface{}   `json:"healthmark"`
			Ingredients           []interface{} `json:"ingredients"`
			LowerAgeLimit         interface{}   `json:"lowerAgeLimit"`
			ManufacturerMarketing interface{}   `json:"manufacturerMarketing"`
			MarketingTextInfo     []string      `json:"marketingTextInfo"`
			NappyInfo             interface{}   `json:"nappyInfo"`
			NetContents           interface{}   `json:"netContents"`
			NumberOfUses          string        `json:"numberOfUses"`
			NutritionInfo         []struct {
				Name                string      `json:"name"`
				PerComp             string      `json:"perComp"`
				PerServing          string      `json:"perServing"`
				ReferenceIntake     interface{} `json:"referenceIntake"`
				ReferencePercentage interface{} `json:"referencePercentage"`
			} `json:"nutritionInfo"`
			NutritionalClaims interface{} `json:"nutritionalClaims"`
			OriginInformation []struct {
				Title string `json:"title"`
				Value string `json:"value"`
			} `json:"originInformation"`
			OtherInformation          interface{} `json:"otherInformation"`
			OtherNutritionInformation interface{} `json:"otherNutritionInformation"`
			PackSize                  interface{} `json:"packSize"`
			PreparationAndUsage       []string    `json:"preparationAndUsage"`
			PreparationGuidelines     interface{} `json:"preparationGuidelines"`
			ProductMarketing          []string    `json:"productMarketing"`
			RecyclingInfo             interface{} `json:"recyclingInfo"`
			SafetyWarning             interface{} `json:"safetyWarning"`
			Storage                   []string    `json:"storage"`
			UpperAgeLimit             interface{} `json:"upperAgeLimit"`
			Warnings                  []string    `json:"warnings"`
		} `json:"details"`
		DisplayType                string        `json:"displayType"`
		DistributorAddress         interface{}   `json:"distributorAddress"`
		FoodIcons                  []interface{} `json:"foodIcons"`
		GroupBulkBuyLimit          int           `json:"groupBulkBuyLimit"`
		Gtin                       string        `json:"gtin"`
		ID                         string        `json:"id"`
		Images                     []string      `json:"images"`
		ImporterAddress            interface{}   `json:"importerAddress"`
		Increment                  int           `json:"increment"`
		IsForSale                  bool          `json:"isForSale"`
		IsInFavourites             interface{}   `json:"isInFavourites"`
		IsNew                      bool          `json:"isNew"`
		IsRestrictedOrderAmendment interface{}   `json:"isRestrictedOrderAmendment"`
		ManufacturerAddress        interface{}   `json:"manufacturerAddress"`
		MaxQuantityAllowed         int           `json:"maxQuantityAllowed"`
		MaxWeight                  interface{}   `json:"maxWeight"`
		MinWeight                  interface{}   `json:"minWeight"`
		MultiPackDetails           interface{}   `json:"multiPackDetails"`
		ProductType                string        `json:"productType"`
		RestrictedDelivery         interface{}   `json:"restrictedDelivery"`
		Restrictions               []interface{} `json:"restrictions"`
		ReturnTo                   struct {
			AddressLine1  string      `json:"addressLine1"`
			AddressLine10 interface{} `json:"addressLine10"`
			AddressLine11 interface{} `json:"addressLine11"`
			AddressLine12 interface{} `json:"addressLine12"`
			AddressLine2  string      `json:"addressLine2"`
			AddressLine3  string      `json:"addressLine3"`
			AddressLine4  interface{} `json:"addressLine4"`
			AddressLine5  interface{} `json:"addressLine5"`
			AddressLine6  interface{} `json:"addressLine6"`
			AddressLine7  interface{} `json:"addressLine7"`
			AddressLine8  interface{} `json:"addressLine8"`
			AddressLine9  interface{} `json:"addressLine9"`
		} `json:"returnTo"`
		Reviews struct {
			Entries []interface{} `json:"entries"`
			Errors  []interface{} `json:"errors"`
			Info    struct {
				Count  int `json:"count"`
				Offset int `json:"offset"`
				Page   int `json:"page"`
				Total  int `json:"total"`
			} `json:"info"`
			Product struct {
				Tpnb string      `json:"tpnb"`
				Tpnc interface{} `json:"tpnc"`
			} `json:"product"`
			Stats struct {
				CountsPerRatingLevel interface{} `json:"countsPerRatingLevel"`
				CreatedOn            interface{} `json:"createdOn"`
				NoOfReviews          interface{} `json:"noOfReviews"`
				OverallRating        interface{} `json:"overallRating"`
				OverallRatingRange   interface{} `json:"overallRatingRange"`
			} `json:"stats"`
		} `json:"reviews"`
		ShelfLife              interface{} `json:"shelfLife"`
		Status                 string      `json:"status"`
		SuperDepartmentName    string      `json:"superDepartmentName"`
		TimeRestrictedDelivery interface{} `json:"timeRestrictedDelivery"`
		Title                  string      `json:"title"`
	} `json:"product"`
	Promotions      []interface{} `json:"promotions"`
	Recommendations struct {
		Data       interface{} `json:"data"`
		Experiment struct {
			Experiment string `json:"experiment"`
			Variation  string `json:"variation"`
		} `json:"experiment"`
	} `json:"recommendations"`
	Referer             string `json:"referer"`
	RestOfShelfURL      string `json:"restOfShelfUrl"`
	Robots              string `json:"robots"`
	ShelfID             string `json:"shelfId"`
	ShelfName           string `json:"shelfName"`
	StructuredData      string `json:"structuredData"`
	SuperDepartmentID   string `json:"superDepartmentId"`
	SuperDepartmentName string `json:"superDepartmentName"`
	Template            string `json:"template"`
}

var productf string = "https://www.tesco.com/groceries/en-GB/products/%v"
var dataRegexp *regexp.Regexp = regexp.MustCompile(`data-props="({.*})"`)

var invalidProductIDf string = "%v is an invalid productID"

func main() {
	app := cli.NewApp()

	app.Name = "tesco"
	app.Usage = "query the tesco site"

	app.Commands = []cli.Command{
		{
			Name:    "product",
			Aliases: []string{"p"},
			Usage:   "get product by product ID",
			Flags: []cli.Flag{
				cli.Int64Flag{
					Name:     "pid",
					Usage:    "Product ID of the product",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				pid := c.Int64("pid")
				if pid < 100000000 {
					return fmt.Errorf(invalidProductIDf, pid)
				}

				data, err := getProduct(pid)
				if err != nil {
					return err
				}
				fmt.Println(data)

				return nil
			},
		},
		{
			Name:    "category",
			Aliases: []string{"c"},
			Usage:   "get category data by url",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "url",
					Usage:    "A category URL",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				url := c.String("url")

				data, err := getCategory(url)
				if err != nil {
					return err
				}
				fmt.Println(data)

				return nil
			},
		},
	}

	app.OnUsageError = func(c *cli.Context, err error, isSubcommand bool) error {
		cli.ShowCommandHelp(c, c.Command.Name)
		return nil
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func extractResources(resp *http.Response) (map[string]map[string]interface{}, error) {
	var resourceMap map[string]map[string]interface{}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resourceMap, fmt.Errorf("Network error")
	}

	matches := dataRegexp.FindStringSubmatch(string(body))

	if len(matches) < 2 {
		return resourceMap, fmt.Errorf("Failed to extract data")
	}

	jsonString := html.UnescapeString(matches[1])
	var jsonMap map[string]map[string]map[string]interface{}

	json.Unmarshal([]byte(jsonString), &jsonMap)

	_, errPresent := jsonMap["error"]
	if errPresent {
		return resourceMap, errors.New("Error returned from Tesco")
	}

	resourceMap, ok := jsonMap["resources"]
	if !ok {
		return resourceMap, errors.New("Unable to access resources")
	}

	return resourceMap, nil
}

// getCategory takes a tesco category page and returns the data
// or an error for parameter, network or request failures
func getCategory(url string) (string, error) {
	var data string

	resp, err := http.Get(url)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	resources, err := extractResources(resp)
	if err != nil {
		return data, fmt.Errorf("getCategory: Unable to extract resources for url: %v", url)
	}

	productsByCategory, ok := resources["productsByCategory"]
	if !ok {
		return data, fmt.Errorf("getCategory: Unable to access productsByCategory for url: %v", url)
	}

	mappp.Pp(productsByCategory)

	return "", nil
}

// getProduct returns the product data
// or an error for parameter, network or request failures
func getProduct(id int64) (ProductData, error) {
	var data ProductData

	if id < 100000000 {
		return data, fmt.Errorf(invalidProductIDf, id)
	}

	productURL := fmt.Sprintf(productf, id)

	resp, err := http.Get(productURL)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	resources, err := extractResources(resp)
	if err != nil {
		return data, fmt.Errorf("getProduct: Unable to extract resources for pid: %v", id)
	}

	productDetails, ok := resources["productDetails"]
	if !ok {
		return data, fmt.Errorf("getProduct: Unable to access productDetails for pid: %v", id)
	}

	dataMap, ok := productDetails["data"]
	if !ok {
		return data, fmt.Errorf("getProduct: Unable to access data for pid: %v", id)
	}

	dataStr, err := json.Marshal(dataMap)
	if err != nil {
		return data, fmt.Errorf("getProduct: Unable to marshal data to map for pid: %v", id)
	}

	err = json.Unmarshal(dataStr, &data)
	if err != nil {
		return data, fmt.Errorf("getProduct: Unable to unmarshal to ProductData for pid: %v", id)
	}

	return data, nil
}
