package main

import (
	"database/sql"
	"reflect"
	"testing"
)

func newSqliteBeerManager() sqliteBeerManager {
	db, err := NewSqliteDatabase(":memory:")
	if err != nil {
		panic(err)
	}

	return db.Beers.(sqliteBeerManager)
}

func (m sqliteBeerManager) mustExec(name string, args ...interface{}) sql.Result {
	result, err := m.dot.Exec(m.db, name, args...)
	if err != nil {
		panic(err)
	}

	return result
}

/*
func newSqliteUserManager() sqliteUserManager {
	db, err := NewSqliteDatabase(":memory:")
	if err != nil {
		panic(err)
	}

	return db.Users.(sqliteUserManager)
}
*/

func TestAllBeersWithoutHistory(t *testing.T) {
	beers := newSqliteBeerManager()
	beers.mustExec("testing/insert-beers")

	got, err := beers.All()
	if err != nil {
		t.Errorf("beers.All() failed: %v", err)
	}

	want := []Beer{
		{
			ID:                   1,
			BarID:                1,
			Name:                 "Bush",
			StockQuantity:        24,
			SoldQuantity:         0,
			PreviousSoldQuantity: 0,
			TotalSoldQuantity:    0,
			SellingPrice:         1.3,
			PreviousSellingPrice: 1.3,
			PurchasePrice:        1.3,
			AlcoholContent:       12,
			IncrCoef:             0.01,
			DecrCoef:             0.02,
			MinCoef:              0.8,
			MaxCoef:              1.2,
		},
		{
			ID:                   2,
			BarID:                3,
			Name:                 "TK",
			StockQuantity:        48,
			SoldQuantity:         0,
			PreviousSoldQuantity: 0,
			TotalSoldQuantity:    0,
			SellingPrice:         1.2,
			PreviousSellingPrice: 1.2,
			PurchasePrice:        1.2,
			AlcoholContent:       8.4,
			IncrCoef:             0.02,
			DecrCoef:             0.02,
			MinCoef:              0.8,
			MaxCoef:              1.2,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("beers.All() = %v; got %v", want, got)
	}
}

func TestAllBeersWithHistory(t *testing.T) {
	beers := newSqliteBeerManager()
	beers.mustExec("testing/insert-beers")
	beers.mustExec("testing/insert-history")

	got, err := beers.All()
	if err != nil {
		t.Errorf("beers.All() failed: %v", err)
	}

	want := []Beer{
		{
			ID:                   1,
			BarID:                1,
			Name:                 "Bush",
			StockQuantity:        24,
			SoldQuantity:         5,
			PreviousSoldQuantity: 23,
			TotalSoldQuantity:    38,
			SellingPrice:         1.2,
			PreviousSellingPrice: 1.4,
			PurchasePrice:        1.3,
			AlcoholContent:       12,
			IncrCoef:             0.01,
			DecrCoef:             0.02,
			MinCoef:              0.8,
			MaxCoef:              1.2,
		},
		{
			ID:                   2,
			BarID:                3,
			Name:                 "TK",
			StockQuantity:        48,
			SoldQuantity:         10,
			PreviousSoldQuantity: 9,
			TotalSoldQuantity:    22,
			SellingPrice:         1.2,
			PreviousSellingPrice: 1,
			PurchasePrice:        1.2,
			AlcoholContent:       8.4,
			IncrCoef:             0.02,
			DecrCoef:             0.02,
			MinCoef:              0.8,
			MaxCoef:              1.2,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("beers.All() = %v; got %v", want, got)
	}
}

func TestEstimatedProfit(t *testing.T) {
	beers := newSqliteBeerManager()
	beers.mustExec("testing/insert-beers")
	beers.mustExec("testing/insert-history")

	got, err := beers.EstimatedProfit()
	if err != nil {
		t.Errorf("beers.EstimatedProfit() failed: %v", err)
	}

	want := 10.4
	if got < want-1e-3 || got > want+1e-3 {
		t.Errorf("beers.EstimatedProfit() = %v; want %v", got, want)
	}
}
