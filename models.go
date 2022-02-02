package main

import (
	"math"

	"golang.org/x/crypto/bcrypt"
)

// Database gives access to all models that can be stored.
type Database struct {
	Beers BeerManager
	Users UserManager
}

// BeerManager includes all possible operations on the Beer model.
type BeerManager interface {
	All() ([]Beer, error)
	EstimatedProfit() (float64, error)
	MakeOrder(id uint, amount int) error
	UpdatePrice(id uint, price float64) error
	UpdatePrices() error
}

// UserManager includes all possible operations on the User model.
type UserManager interface {
	All() ([]User, error)
	ByID(id uint) (User, error)
	ByName(name string) (User, error)
	ByToken(token string) (User, error)
	Create(name, password string, admin bool) (User, error)
	Update(user *User) error
	Delete(id uint) error
	CreateToken(userID uint) (string, error)
	DeleteToken(token string) error
}

// Beer represents a type of beer from the database.
type Beer struct {
	ID                   uint    `json:"id"`
	BarID                uint    `json:"barId"`
	Name                 string  `json:"name"`
	StockQuantity        int     `json:"stockQuantity"`
	SoldQuantity         int     `json:"-"`
	PreviousSoldQuantity int     `json:"-"`
	TotalSoldQuantity    int     `json:"totalSoldQuantity"`
	SellingPrice         float64 `json:"sellingPrice"`
	PreviousSellingPrice float64 `json:"previousSellingPrice"`
	PurchasePrice        float64 `json:"-"`
	AlcoholContent       float64 `json:"alcoholContent"`
	IncrCoef             float64 `json:"-"`
	DecrCoef             float64 `json:"-"`
	MinCoef              float64 `json:"-"`
	MaxCoef              float64 `json:"-"`
}

// NewPrice computes and returns the beer's new price based on its current
// quantity, price and sold quantity of the last period.
func (b *Beer) NewPrice() float64 {
	price := b.SellingPrice
	stock := b.StockQuantity - b.TotalSoldQuantity
	delta := float64(b.SoldQuantity - b.PreviousSoldQuantity)

	if stock > 10 {
		if delta > 0 {
			price += b.IncrCoef * delta
		} else {
			price += b.DecrCoef * delta
		}
	} else if stock > 5 {
		price = 2.1 * b.PurchasePrice
	} else {
		price = 2.5 * b.PurchasePrice
	}

	minPrice := b.MinCoef * b.PurchasePrice
	maxPrice := b.MaxCoef * b.PurchasePrice
	return math.Min(math.Max(minPrice, price), maxPrice)
}

// User represents an user from the database.
//
// Its Password is actually a hash and should not be accessed directly but
// through SetPassword and CheckPassword instead.
type User struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Password []byte `json:"-"`
	Admin    bool   `json:"admin"`
}

// SetPassword hashes, salts and updates a password.
func (u *User) SetPassword(password string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// GenerateFromPassword could fail if bad parameters are given, or if it
		// cannot access the OS's secure RNG (https://stackoverflow.com/q/57032884).
		// In either case, we should panic.
		panic(err)
	}

	u.Password = hash
}

// CheckPassword tests a given password against the stored hash.
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(u.Password, []byte(password))
	return err == nil
}
