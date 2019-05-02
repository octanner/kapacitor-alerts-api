package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" //driver
)

//GetDB centralized access point
func GetDB(uri string) *sqlx.DB {
	db, dberr := sqlx.Open("postgres", uri)
	if dberr != nil {
		fmt.Println(dberr)
		return nil
	}
	// not available in 1.5 golang, youll want to turn it on for v1.6 or higher once upgraded.
	//pool.SetConnMaxLifetime(time.ParseDuration("1h"));
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(20)
	return db
}

// GetDBFromContext - Get the SQL DB connection from the Gin context
func GetDBFromContext(c *gin.Context) (*sqlx.DB, error) {
	db, exists := c.Get("db")
	if !exists {
		return nil, errors.New("DB not available in context")
	}
	return db.(*sqlx.DB), nil
}

// DBMiddleware - Add a SQL database connection to the Gin context
func DBMiddleware(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

// InitDB - Run any available creation and migration scripts
func InitDB(db *sqlx.DB) {
	buf, err := ioutil.ReadFile("./create.sql")
	if err != nil {
		log.Println("Error: Unable to run migration scripts, could not load create.sql.")
		log.Fatalln(err)
	}
	_, err = db.Exec(string(buf))
	if err != nil {
		log.Println("Error: Unable to run migration scripts, execution failed.")
		log.Fatalln(err)
	}
}
