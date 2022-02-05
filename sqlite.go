package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"os"
	"path"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/qustavo/dotsql"
)

func NewSqliteDatabase(dataSourceName string) (Database, error) {
	var database Database

	db, err := sql.Open("sqlite3", dataSourceName+"?_foreign_keys=on")
	if err != nil {
		return database, err
	}

	dot, err := loadDotSqlFromDir("sql")
	if err != nil {
		return database, err
	}

	if _, err := dot.Exec(db, "init"); err != nil {
		return database, err
	}

	database.Beers = &sqliteBeerManager{db, dot}
	database.Users = &sqliteUserManager{db, dot}
	return database, err
}

func loadDotSqlFromDir(name string) (*dotsql.DotSql, error) {
	entries, err := os.ReadDir(name)
	if err != nil {
		return nil, err
	}

	var dots []*dotsql.DotSql
	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ".sql") {
			dot, err := dotsql.LoadFromFile(path.Join(name, entry.Name()))
			if err != nil {
				return nil, err
			}

			dots = append(dots, dot)
		}
	}

	return dotsql.Merge(dots...), nil
}

type sqliteBeerManager struct {
	db  *sql.DB
	dot *dotsql.DotSql
}

func (m sqliteBeerManager) All() ([]Beer, error) {
	rows, err := m.dot.Query(m.db, "beers/get-all")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	beers := []Beer{}
	for rows.Next() {
		var b Beer
		if err := rows.Scan(&b.ID, &b.BarID, &b.Name, &b.StockQuantity, &b.SoldQuantity, &b.PreviousSoldQuantity, &b.TotalSoldQuantity, &b.SellingPrice, &b.PreviousSellingPrice, &b.PurchasePrice, &b.AlcoholContent, &b.IncrCoef, &b.DecrCoef, &b.MinCoef, &b.MaxCoef); err != nil {
			return nil, err
		}

		beers = append(beers, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return beers, nil
}

func (m sqliteBeerManager) Create(b *Beer) error {
	result, err := m.dot.Exec(m.db, "beers/create", b.BarID, b.Name, b.StockQuantity, b.PurchasePrice, b.AlcoholContent, b.IncrCoef, b.DecrCoef, b.MinCoef, b.MaxCoef)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	if _, err := m.dot.Exec(m.db, "beers/update-price", id, b.PurchasePrice); err != nil {
		return err
	}

	b.ID = uint(id)
	b.SellingPrice = b.PurchasePrice
	b.PreviousSellingPrice = b.PurchasePrice
	return nil
}

func (m sqliteBeerManager) DeleteAll() error {
	_, err := m.dot.Exec(m.db, "beers/delete-all")
	if err != nil {
		return err
	}

	return nil
}

func (m sqliteBeerManager) EstimatedProfit() (float64, error) {
	row, err := m.dot.QueryRow(m.db, "beers/get-estimated-profit")
	if err != nil {
		return 0, err
	}

	var profit float64
	if err := row.Scan(&profit); err != nil {
		return 0, err
	}

	return profit, nil
}

func (m sqliteBeerManager) MakeOrder(id uint, amount int) error {
	if _, err := m.dot.Exec(m.db, "beers/make-order", id, amount); err != nil {
		return err
	}

	return nil
}

func (m sqliteBeerManager) UpdatePrice(id uint, price float64) error {
	if _, err := m.dot.Exec(m.db, "beers/update-price", id, price); err != nil {
		return err
	}

	return nil
}

func (m sqliteBeerManager) UpdatePrices() error {
	beers, err := m.All()
	if err != nil {
		return err
	}

	for _, beer := range beers {
		if err := m.UpdatePrice(beer.ID, beer.NewPrice()); err != nil {
			return err
		}
	}

	return nil
}

type sqliteUserManager struct {
	db  *sql.DB
	dot *dotsql.DotSql
}

func (m sqliteUserManager) All() ([]User, error) {
	rows, err := m.dot.Query(m.db, "users/get-all")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Password, &user.Admin); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (m sqliteUserManager) ByID(id uint) (User, error) {
	var user User

	row, err := m.dot.QueryRow(m.db, "users/get-by-id", id)
	if err != nil {
		return user, err
	}

	if err := row.Scan(&user.ID, &user.Name, &user.Password, &user.Admin); err != nil {
		return user, err
	}

	return user, nil
}

func (m sqliteUserManager) ByName(name string) (User, error) {
	var user User

	row, err := m.dot.QueryRow(m.db, "users/get-by-name", name)
	if err != nil {
		return user, err
	}

	if err := row.Scan(&user.ID, &user.Name, &user.Password, &user.Admin); err != nil {
		return user, err
	}

	return user, nil
}

func (m sqliteUserManager) ByToken(token string) (User, error) {
	var user User

	row, err := m.dot.QueryRow(m.db, "users/get-by-token", token)
	if err != nil {
		return user, err
	}

	if err := row.Scan(&user.ID, &user.Name, &user.Password, &user.Admin); err != nil {
		return user, err
	}

	return user, nil
}

func (m sqliteUserManager) Create(name, password string, admin bool) (User, error) {
	user := User{Name: name, Admin: admin}
	user.SetPassword(password)
	result, err := m.dot.Exec(m.db, "users/create", user.Name, user.Password, user.Admin)
	if err != nil {
		return user, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return user, err
	}

	user.ID = uint(id)
	return user, nil
}

func (m sqliteUserManager) Update(u *User) error {
	if _, err := m.dot.Exec(m.db, "users/update", u.ID, u.Name, u.Password, u.Admin); err != nil {
		return err
	}

	return nil
}

func (m sqliteUserManager) Delete(id uint) error {
	if _, err := m.dot.Exec(m.db, "users/delete", id); err != nil {
		return err
	}

	return nil
}

func (m sqliteUserManager) CreateToken(userID uint) (string, error) {
	token := generateToken()
	if _, err := m.dot.Exec(m.db, "users/create-token", userID, token); err != nil {
		return token, err
	}

	return token, nil
}

func generateToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Read could fail but if it does we should probably panic anyway
		// (https://stackoverflow.com/a/42318347).
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes)
}

func (m sqliteUserManager) DeleteToken(token string) error {
	if _, err := m.dot.Exec(m.db, "users/delete-token", token); err != nil {
		return err
	}

	return nil
}
