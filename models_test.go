package main

import "testing"

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
				StockQuantity:        6,
				SoldQuantity:         4,
				PreviousSoldQuantity: 1,
				TotalSoldQuantity:    5,
				SellingPrice:         1,
				PurchasePrice:        1,
				IncrCoef:             0.01,
				DecrCoef:             0.01,
				MinCoef:              0,
				MaxCoef:              5,
			},
			want: 2.5, // only 1 unit left
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
