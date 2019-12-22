// Package category implements functions to request and manipulate category data from tesco
package category

import (
	"fmt"
	"net/url"
)

// AddCountToURL returns the passed url with a count=48 query parameter
func AddCountToURL(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", fmt.Errorf("%v was not a valid URL: %v", u, err)
	}
	q := parsed.Query()
	q.Set("count", "48")
	parsed.RawQuery = q.Encode()
	u = fmt.Sprint(parsed)
	return u, nil
}


