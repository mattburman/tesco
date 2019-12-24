package product

import "testing"

func TestURLToID(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"https://www.tesco.com/groceries/en-GB/products/1",
			args{"https://www.tesco.com/groceries/en-GB/products/1"},
			"1",
			false,
		},
		{
			"https://www.tesco.com/groceries/en-GB/products/123456789",
			args{"https://www.tesco.com/groceries/en-GB/products/123456789"},
			"123456789",
			false,
		},
		{
			"https://www.tesco.com/groceries/en-GB/products/987654321/",
			args{"https://www.tesco.com/groceries/en-GB/products/987654321/"},
			"987654321",
			false,
		},
		{
			"https://www.tesco.com/groceries/en-GB/products/",
			args{"https://www.tesco.com/groceries/en-GB/products/"},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := URLToID(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("URLToID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("URLToID() got = %v, want %v", got, tt.want)
			}
		})
	}
}
