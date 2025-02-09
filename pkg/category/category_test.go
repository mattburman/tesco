package category

import "testing"

func TestAddCountToURL(t *testing.T) {
	tables := []struct {
		input string
		want  string
	}{
		{
			"https://www.product.com/groceries/en-GB/shop/fresh-food/fresh-meat-and-poultry/fresh-beef/all",
			"https://www.product.com/groceries/en-GB/shop/fresh-food/fresh-meat-and-poultry/fresh-beef/all?count=48",
		},
		{
			"https://www.product.com/groceries/en-GB/shop/fresh-food/fresh-meat-and-poultry/fresh-beef/all?count=24",
			"https://www.product.com/groceries/en-GB/shop/fresh-food/fresh-meat-and-poultry/fresh-beef/all?count=48",
		},
		{
			"https://www.product.com/groceries/en-GB/shop/fresh-food/fresh-meat-and-poultry/fresh-beef/all?random=24",
			"https://www.product.com/groceries/en-GB/shop/fresh-food/fresh-meat-and-poultry/fresh-beef/all?count=48&random=24",
		},
	}

	for _, tc := range tables {
		out, _ := AddCountToURL(tc.input)
		if out != tc.want {
			t.Errorf("count was not added correctly. got: %v, want: %v for url: %v", out, tc.want, tc.input)
		}
	}
}
