package utils

import (
	"errors"
	"fmt"

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
