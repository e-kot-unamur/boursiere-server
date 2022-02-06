package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"

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
	Create(b *Beer) error
	DeleteAll() error
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
	Update(u *User) error
	Delete(id uint) error
	CreateToken(userID uint) (string, error)
	DeleteToken(token string) error
}

// Beer represents a type of beer from the database.
type Beer struct {
	ID                   uint    `json:"id" csv:"-"`
	BarID                uint    `json:"barId" csv:"barId"`
	Name                 string  `json:"name" csv:"name"`
	StockQuantity        int     `json:"stockQuantity" csv:"stockQuantity"`
	SoldQuantity         int     `json:"-" csv:"-"`
	PreviousSoldQuantity int     `json:"-" csv:"-"`
	TotalSoldQuantity    int     `json:"totalSoldQuantity" csv:"-"`
	SellingPrice         float64 `json:"sellingPrice" csv:"-"`
	PreviousSellingPrice float64 `json:"previousSellingPrice" csv:"-"`
	PurchasePrice        float64 `json:"-" csv:"purchasePrice"`
	BottleSize           float64 `json:"bottleSize" csv:"bottleSize"`
	AlcoholContent       float64 `json:"alcoholContent" csv:"alcoholContent"`
	IncrCoef             float64 `json:"-" csv:"incrCoef"`
	DecrCoef             float64 `json:"-" csv:"decrCoef"`
	MinCoef              float64 `json:"-" csv:"minCoef"`
	MaxCoef              float64 `json:"-" csv:"maxCoef"`
}

var beerType = reflect.TypeOf(Beer{})
var beerFields map[string]reflect.StructField

func init() {
	beerFields = make(map[string]reflect.StructField)
	for i := 0; i < beerType.NumField(); i++ {
		field := beerType.Field(i)
		title := field.Tag.Get("csv")
		if title == "-" {
			continue
		}
		if title == "" {
			title = field.Name
		}
		beerFields[title] = field
	}
}

// LoadBeersFromCSV parses CSV data and generates beers. It mimics Unmarshal
// from the standard library.
//
// The first row is interpreted as column names and takes the `csv` struct tag
// into account. Missing columns are ignored and `csv:"-"` tags are omitted.
//
// For columns of type float64, "," are replaced with "." to handle French
// decimal commas. Furthermore, trailling spaces (" ") and euro symbols ("€")
// are removed.
func LoadBeersFromCSV(source io.Reader) ([]Beer, error) {
	r := csv.NewReader(source)
	titles, err := r.Read()
	if err != nil {
		return nil, err
	}

	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	mapping := map[string]int{}
	for i, title := range titles {
		if _, ok := beerFields[title]; ok {
			mapping[title] = i
		}
	}

	beers := []Beer{}
	for _, record := range records {
		beer := reflect.New(beerType)
		ptr := reflect.Indirect(beer)

		for title, field := range beerFields {
			i, ok := mapping[title]
			if !ok {
				continue
			}

			s := record[i]
			f := ptr.Field(field.Index[0])

			switch field.Type.Kind() {
			case reflect.Int:
				v, err := strconv.ParseInt(s, 10, 0)
				if err != nil {
					return nil, err
				}
				f.SetInt(v)

			case reflect.Uint:
				v, err := strconv.ParseUint(s, 10, 0)
				if err != nil {
					return nil, err
				}
				f.SetUint(v)

			case reflect.Float64:
				s = strings.TrimRight(s, " €")
				s = strings.ReplaceAll(s, ",", ".")
				v, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return nil, err
				}
				f.SetFloat(v)

			case reflect.String:
				f.SetString(s)

			default:
				panic(fmt.Sprintf("unmanaged type %v for field %v", field.Type, field.Name))
			}
		}

		beers = append(beers, *beer.Interface().(*Beer))
	}

	return beers, nil
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

// User represents a user from the database.
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
