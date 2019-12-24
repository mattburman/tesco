package collecting

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"html"
)

// ExtractResources takes data-props and returns the resources as json
func ExtractResources(dataProps string) (*string, error) {
	json := html.UnescapeString(dataProps)

	if !gjson.Valid(json) {
		return nil, errors.New("invalid json")
	}

	if err := gjson.Get(json, "error"); err.Exists() {
		return nil, fmt.Errorf("error returned from Tesco: %v", err.String())
	}

	resources := gjson.Get(json, "resources").String()

	return &resources, nil
}

