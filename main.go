package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type (
	loginReq struct {
		Name     string `json:"name" binding:"alphanum,min=3,max=256"`
		Password string `json:"password" binding:"min=3,max=256"`
	}

	createUserReq struct {
		Name     string `json:"name" binding:"alphanum,min=3,max=256"`
		Password string `json:"password" binding:"min=3,max=256"`
		Admin    bool   `json:"admin"`
	}

	updateUserReq struct {
		Name     string `json:"name" binding:"omitempty,alphanum,min=3,max=256"`
		Password string `json:"password" binding:"omitempty,min=3,max=256"`
		Admin    bool   `json:"admin"`
	}

	orderReq []struct {
		ID              uint `json:"id" binding:"min=1"`
		OrderedQuantity int  `json:"orderedQuantity"`
	}

	createEntryReq struct {
		OrderedQuantity int `json:"orderedQuantity"`
	}
)

const period = 2 * time.Minute

func main() {
	dataSourceName := os.Getenv("DATABASE_FILE")
	if dataSourceName == "" {
		dataSourceName = "db.sqlite3"
	}

	db, err := NewSqliteDatabase(dataSourceName)
	if err != nil {
		panic(err)
	}

	count, err := db.Users.Count()
	if err != nil {
		panic(err)
	}

	if count == 0 {
		if _, err := db.Users.Create("admin", "boursière", true); err != nil {
			panic(err)
		}
	}

	broker := NewBroker()
	entriesBroker := NewBroker()
	go func() {
		p := period.Milliseconds()
		n := time.Now().UnixMilli()
		t := time.UnixMilli((n/p + 1) * p)
		time.Sleep(time.Until(t))

		tick := time.NewTicker(period)
		for {
			err := db.Beers.UpdatePrices()
			if err != nil {
				panic(err)
			}

			beers, err := db.Beers.All()
			if err != nil {
				panic(err)
			}

			entries, err := db.Entries.All()
			if err != nil {
				panic(err)
			}

			broker.Broadcast(gin.H{
				"type": "update",
				"data": beers,
			})

			entriesBroker.Broadcast(gin.H{
				"type": "update",
				"data": entries,
			})

			<-tick.C
		}
	}()

	router := gin.Default()
	router.Use(noCache)
	if gin.IsDebugging() {
		router.Use(debugCORS())
	}

	// Take a look at ./doc/routes.md for detailed route descriptions.

	// Get the list of all beers.
	router.GET("/api/beers", func(c *gin.Context) {
		beers, err := db.Beers.All()
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, beers)
	})

	// Delete all existing beers and upload new ones.
	router.POST("/api/beers", auth(db.Users, true), func(c *gin.Context) {
		if c.ContentType() != "text/csv" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}

		beers, err := LoadBeersFromCSV(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}

		if err := db.Beers.DeleteAll(); err != nil {
			panic(err)
		}

		if err := db.Entries.DeleteAll(); err != nil {
			panic(err)
		}

		for i := range beers {
			if err := db.Beers.Create(&beers[i]); err != nil {
				panic(err)
			}
		}

		broker.Broadcast(gin.H{
			"type": "update",
			"data": beers,
		})

		c.JSON(http.StatusCreated, beers)
	})

	// Get real-time updates of beers' status.
	router.GET("/api/beers/events", broker.ServeHTTP)

	// Order beers.
	router.POST("/api/beers/order", auth(db.Users, false), func(c *gin.Context) {
		var req orderReq
		if err := c.BindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}

		for _, order := range req {
			if err := db.Beers.MakeOrder(order.ID, order.OrderedQuantity); err != nil {
				panic(err)
			}
		}

		broker.Broadcast(gin.H{
			"type": "order",
			"data": req,
		})

		c.Status(http.StatusNoContent)
	})

	// Get administration statistics about the event.
	router.GET("/api/beers/stats", auth(db.Users, true), func(c *gin.Context) {
		profit, err := db.Beers.EstimatedProfit()
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"estimatedProfit": profit,
		})
	})

	// Get the list of all users.
	router.GET("/api/users", auth(db.Users, true), func(c *gin.Context) {
		users, err := db.Users.All()
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, users)
	})

	// Create a new user.
	router.POST("/api/users", auth(db.Users, true), func(c *gin.Context) {
		var req createUserReq
		if err := c.BindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}

		user, err := db.Users.Create(req.Name, req.Password, req.Admin)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "non_unique_name"})
			return
		}

		c.JSON(http.StatusCreated, user)
	})

	// Edit a user.
	router.PATCH("/api/users/:id", auth(db.Users, false), func(c *gin.Context) {
		id64, err := strconv.ParseUint(c.Param("id"), 10, 0)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}

		var req updateUserReq
		if err := c.BindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}

		id := uint(id64)
		client := c.MustGet("user").(User)
		if !client.Admin && (client.ID != id || req.Admin) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
			return
		}

		user, err := db.Users.ByID(id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "invalid_id"})
			return
		}

		user.Admin = req.Admin
		if req.Name != "" {
			user.Name = req.Name
		}
		if req.Password != "" {
			user.SetPassword(req.Password)
		}

		if err := db.Users.Update(&user); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "non_unique_name"})
			return
		}

		c.JSON(http.StatusOK, user)
	})

	// Delete a user.
	router.DELETE("/api/users/:id", auth(db.Users, true), func(c *gin.Context) {
		id64, err := strconv.ParseUint(c.Param("id"), 10, 0)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}

		id := uint(id64)
		if err := db.Users.Delete(id); err != nil {
			panic(err)
		}

		c.Status(http.StatusNoContent)
	})

	// Generate a new access token.
	router.POST("/api/users/token", func(c *gin.Context) {
		var req loginReq
		if err := c.BindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}

		user, err := db.Users.ByName(req.Name)
		if err != nil || !user.CheckPassword(req.Password) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "wrong_credentials"})
			return
		}

		token, err := db.Users.CreateToken(user.ID)
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"admin": user.Admin,
			"token": token,
		})
	})

	// Delete an access token.
	router.DELETE("/api/users/token", auth(db.Users, false), func(c *gin.Context) {
		token := c.MustGet("token").(string)
		if err := db.Users.DeleteToken(token); err != nil {
			panic(err)
		}

		c.Status(http.StatusNoContent)
	})

	// Get the list of all entries.
	router.GET("/api/entries", auth(db.Users, false), func(c *gin.Context) {
		entries, err := db.Entries.All()
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, entries)
	})

	// Get real-time updates of entries' status.
	router.GET("/api/entries/events", entriesBroker.ServeHTTP)

	// Add a new entry sale.
	router.POST("/api/entries", auth(db.Users, false), func(c *gin.Context) {
		fmt.Println("new entry")
		var req createEntryReq
		if err := c.BindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
			return
		}

		entry, err := db.Entries.Create(req.OrderedQuantity)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "non_unique_name"})
			return
		}

		entries, err := db.Entries.All()
		if err != nil {
			panic(err)
		}

		entriesBroker.Broadcast(gin.H{
			"type": "order",
			"data": entries,
		})

		c.JSON(http.StatusCreated, entry)
	})

	// Get statistic about entries.
	router.GET("/api/entries/stat", auth(db.Users, false), func(c *gin.Context) {
		count, err := db.Entries.Count()
		if err != nil {
			panic(err)
		}

		c.JSON(http.StatusOK, count)
	})

	router.Run()
}

// noCache is a middleware that forbids clients to cache any response.
func noCache(c *gin.Context) {
	c.Writer.Header().Set("Cache-Control", "no-store")
}

// debugCORS is a middleware that allows clients to make cross-origin requests
// with credentials. It should only be used while debugging!
func debugCORS() gin.HandlerFunc {
	cfg := cors.DefaultConfig()
	cfg.AllowOrigins = []string{
		"http://localhost:3000",
		"http://localhost:5000",
	}
	cfg.AllowCredentials = true
	cfg.AddAllowHeaders("Authorization")
	return cors.New(cfg)
}

// auth is a middleware that authenticates requests using access tokens.
//
// Unauthenticated requests are denied and, if admin is set to true, regular
// users are denied too.
//
// If a request is successfully authenticated, the user and its token are
// stored in the context for later use.
func auth(users UserManager, admin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.Request.Header.Get("Authorization")
		token := strings.TrimPrefix(auth, "Bearer ")
		if auth == token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}

		user, err := users.ByToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
			return
		}

		if admin && !user.Admin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "unauthorized"})
			return
		}

		c.Set("user", user)
		c.Set("token", token)
	}
}
