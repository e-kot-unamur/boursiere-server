package main

import (
	"database/sql"
	"reflect"
	"testing"
)

func newSqliteBeerManager() *sqliteBeerManager {
	db, err := NewSqliteDatabase(":memory:")
	if err != nil {
		panic(err)
	}
	return db.Beers.(*sqliteBeerManager)
}

func (m sqliteBeerManager) mustExec(name string, args ...interface{}) sql.Result {
	result, err := m.dot.Exec(m.db, name, args...)
	if err != nil {
		panic(err)
	}
	return result
}

func (m sqliteBeerManager) mustCount(table string) int {
	count := -1
	row := m.db.QueryRow("SELECT COUNT(*) FROM " + table)
	if err := row.Scan(&count); err != nil {
		panic(err)
	}
	return count
}

func newSqliteUserManager() *sqliteUserManager {
	db, err := NewSqliteDatabase(":memory:")
	if err != nil {
		panic(err)
	}
	return db.Users.(*sqliteUserManager)
}

func (m sqliteUserManager) mustExec(name string, args ...interface{}) sql.Result {
	result, err := m.dot.Exec(m.db, name, args...)
	if err != nil {
		panic(err)
	}
	return result
}

func (m sqliteUserManager) mustCount(table string) int {
	count := -1
	row := m.db.QueryRow("SELECT COUNT(*) FROM " + table)
	if err := row.Scan(&count); err != nil {
		panic(err)
	}
	return count
}

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
			BottleSize:           33,
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
			BottleSize:           33,
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
			BottleSize:           33,
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
			BottleSize:           33,
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

func TestCreateBeer(t *testing.T) {
	beers := newSqliteBeerManager()

	if err := beers.Create(&Beer{}); err != nil {
		t.Errorf("beers.Create() failed: %v", err)
	}

	beersCount := beers.mustCount("beers")
	if beersCount != 1 {
		t.Errorf("beersCount = 1; got %v", beersCount)
	}

	historyCount := beers.mustCount("history")
	if historyCount != 1 {
		t.Errorf("historyCount = 1; got %v", historyCount)
	}
}

func TestCreateBeersUpdating(t *testing.T) {
	beers := newSqliteBeerManager()

	got := Beer{
		BarID:          2,
		Name:           "test",
		StockQuantity:  6,
		PurchasePrice:  2.22,
		BottleSize:     25,
		AlcoholContent: 6,
		IncrCoef:       0.02,
		DecrCoef:       0.03,
		MinCoef:        0.9,
		MaxCoef:        2.5,
	}
	if err := beers.Create(&got); err != nil {
		t.Errorf("beers.Create() failed: %v", err)
	}

	want := Beer{
		ID:                   1, // updated
		BarID:                2,
		Name:                 "test",
		StockQuantity:        6,
		SellingPrice:         2.22, // updated
		PreviousSellingPrice: 2.22, // updated
		PurchasePrice:        2.22,
		BottleSize:           25,
		AlcoholContent:       6,
		IncrCoef:             0.02,
		DecrCoef:             0.03,
		MinCoef:              0.9,
		MaxCoef:              2.5,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("beers.Create(&b), b = %v; got %v", want, got)
	}
}

func TestDeleteAllBeers(t *testing.T) {
	beers := newSqliteBeerManager()
	beers.mustExec("testing/insert-beers")
	beers.mustExec("testing/insert-history")

	if err := beers.DeleteAll(); err != nil {
		t.Errorf("beers.DeleteAll() failed: %v", err)
	}

	beersCount := beers.mustCount("beers")
	if beersCount != 0 {
		t.Errorf("beersCount = 0; got %v", beersCount)
	}

	historyCount := beers.mustCount("history")
	if historyCount != 0 {
		t.Errorf("historyCount = 0; got %v", historyCount)
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

func TestAllUsers(t *testing.T) {
	users := newSqliteUserManager()
	users.mustExec("testing/insert-users")

	got, err := users.All()
	if err != nil {
		t.Errorf("users.All() failed: %v", err)
	}

	want := []User{
		{ID: 1, Name: "admin", Password: []byte("hashedpwd"), Admin: true},
		{ID: 2, Name: "bob", Password: []byte("hashedhash"), Admin: false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("users.All() = %v; got %v", want, got)
	}
}

func TestUserByID(t *testing.T) {
	users := newSqliteUserManager()
	users.mustExec("testing/insert-users")

	got, err := users.ByID(2)
	if err != nil {
		t.Errorf("users.ByID(2) failed: %v", err)
	}

	want := User{ID: 2, Name: "bob", Password: []byte("hashedhash"), Admin: false}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("users.ByID(2) = %v; got %v", want, got)
	}
}

func TestUserByName(t *testing.T) {
	users := newSqliteUserManager()
	users.mustExec("testing/insert-users")

	got, err := users.ByName("admin")
	if err != nil {
		t.Errorf("users.ByName(\"admin\") failed: %v", err)
	}

	want := User{ID: 1, Name: "admin", Password: []byte("hashedpwd"), Admin: true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("users.ByName(\"admin\") = %v: got %v", want, got)
	}
}

func TestUserByToken(t *testing.T) {
	users := newSqliteUserManager()
	users.mustExec("testing/insert-users")
	users.mustExec("testing/insert-tokens")

	got, err := users.ByToken("amazingtoken")
	if err != nil {
		t.Errorf("users.ByToken() failed: %v", err)
	}

	want := User{ID: 1, Name: "admin", Password: []byte("hashedpwd"), Admin: true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("users.ByToken() = %v: got %v", want, got)
	}
}

func TestUserByTokenWithBadToken(t *testing.T) {
	users := newSqliteUserManager()
	users.mustExec("testing/insert-users")
	users.mustExec("testing/insert-tokens")

	if _, err := users.ByToken("amazedtoken"); err == nil {
		t.Errorf("users.ByToken() succeeded but shouldn't")
	}
}

func TestUserByTokenWithoutTokens(t *testing.T) {
	users := newSqliteUserManager()
	users.mustExec("testing/insert-users")

	if _, err := users.ByToken("amazingtoken"); err == nil {
		t.Errorf("users.ByToken() succeeded but shouldn't")
	}
}

func TestCreateUser(t *testing.T) {
	users := newSqliteUserManager()

	got, err := users.Create("alice", "secret", true)
	if err != nil {
		t.Errorf("users.Create() failed: %v", err)
	}

	want := User{ID: 1, Name: "alice", Password: got.Password, Admin: true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("users.Create() = %v; got %v", want, got)
	}
	if !got.CheckPassword("secret") {
		t.Errorf("user.CheckPassword() failed but shouldn't")
	}
}

func TestDeleteUser(t *testing.T) {
	users := newSqliteUserManager()
	users.mustExec("testing/insert-users")
	users.mustExec("testing/insert-tokens")

	if err := users.Delete(1); err != nil {
		t.Errorf("users.Delete() failed: %v", err)
	}

	usersCount := users.mustCount("users")
	if usersCount != 1 {
		t.Errorf("usersCount = 1; got %v", usersCount)
	}

	tokensCount := users.mustCount("tokens")
	if tokensCount != 1 {
		t.Errorf("tokensCount = 1; got %v", tokensCount)
	}
}
