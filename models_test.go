package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestLoadBeersFromCSV(t *testing.T) {
	tests := []struct {
		csv  string
		want []Beer
	}{
		{
			csv:  "name\nmyname\n",
			want: []Beer{{Name: "myname"}},
		},
		{
			csv:  "name\nmyname\nyourname\n",
			want: []Beer{{Name: "myname"}, {Name: "yourname"}},
		},
		{
			csv:  "barId\n4\n",
			want: []Beer{{BarID: 4}},
		},
		{
			csv:  "a,b,c\n1,2,3\n",
			want: []Beer{{}},
		},
		{
			csv:  "purchasePrice\n45.2\n",
			want: []Beer{{PurchasePrice: 45.2}},
		},
		{
			csv:  "purchasePrice\n\"45,2€\"\n",
			want: []Beer{{PurchasePrice: 45.2}},
		},
		{
			csv:  "purchasePrice\n\"45,2 €\"\n",
			want: []Beer{{PurchasePrice: 45.2}},
		},
		{
			csv:  "purchasePrice,barId,name\n4 €,1,\"ho ho\"\n",
			want: []Beer{{BarID: 1, Name: "ho ho", PurchasePrice: 4}},
		},
	}

	for _, test := range tests {
		r := strings.NewReader(test.csv)
		got, err := LoadBeersFromCSV(r)
		if err != nil {
			t.Errorf("LoadBeersFromCSV() failed: %v", err)
		}

		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("LoadBeersFromCSV() = %v; got %v", test.want, got)
		}
	}
}

func TestNewPrice(t *testing.T) {
	tests := []struct {
		beer Beer
		want float64
	}{
		{
			beer: Beer{
				StockQuantity:        48,
				SoldQuantity:         5,
				PreviousSoldQuantity: 2,
				TotalSoldQuantity:    7,
				SellingPrice:         1.2,
				PurchasePrice:        1.2,
				IncrCoef:             0.02,
				DecrCoef:             0.05,
				MinCoef:              0.8,
				MaxCoef:              1.2,
			},
			want: 1.26, // 3 sold units more
		},
		{
			beer: Beer{
				StockQuantity:        48,
				SoldQuantity:         8,
				PreviousSoldQuantity: 10,
				TotalSoldQuantity:    13,
				SellingPrice:         1.2,
				PurchasePrice:        1.2,
				IncrCoef:             0.02,
				DecrCoef:             0.05,
				MinCoef:              0.8,
				MaxCoef:              1.2,
			},
			want: 1.1, // 2 sold units less
		},
		{
			beer: Beer{
				StockQuantity:        24,
				SoldQuantity:         0,
				PreviousSoldQuantity: 10,
				TotalSoldQuantity:    10,
				SellingPrice:         1.2,
				PurchasePrice:        1.2,
				IncrCoef:             0.02,
				DecrCoef:             0.05,
				MinCoef:              0.8,
				MaxCoef:              1.2,
			},
			want: 0.96, // 10 sold units less but MinCoef of 0.8
		},
		{
			beer: Beer{
				StockQuantity:        48,
				SoldQuantity:         28,
				PreviousSoldQuantity: 1,
				TotalSoldQuantity:    29,
				SellingPrice:         1,
				PurchasePrice:        1,
				IncrCoef:             0.01,
				DecrCoef:             0.01,
				MinCoef:              1,
				MaxCoef:              1.1,
			},
			want: 1.1, // 27 sold units more but MaxCoef of 1.1
		},
	}

	for _, test := range tests {
		got := test.beer.NewPrice()
		want := test.want
		if got < want-1e-3 || got > want+1e-3 {
			t.Errorf("beer.NewPrice() = %v; want %v", got, want)
		}
	}
}

func TestPassword(t *testing.T) {
	var user User
	user.SetPassword("helloworld")

	if !user.CheckPassword("helloworld") {
		t.Error("user.CheckPassword() failed but shouldn't")
	}
	if user.CheckPassword("helloword") {
		t.Error("user.CheckPassword() succeeded but shouldn't")
	}
}
