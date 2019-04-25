package main

import (
	"fmt"
	"io/ioutil"
	_5xx "kapacitor-alerts-api/5xx"
	crashed "kapacitor-alerts-api/crashed"
	memory "kapacitor-alerts-api/memory"
	released "kapacitor-alerts-api/released"
	utils "kapacitor-alerts-api/utils"
	"log"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// initDB - Run any available creation and migration scripts
func initDB(db *sqlx.DB) {
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

// dbMiddleware - Add a SQL database connection to the Gin context
func dbMiddleware(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

// checkEnv - Verify required environment variables exist
func checkEnv() {
	_, dbURL := os.LookupEnv("DATABASE_URL")
	_, kapURL := os.LookupEnv("KAPACITOR_URL")

	if !dbURL {
		panic("✖ Environment variable DATABASE_URL not found.")
	} else if !kapURL {
		panic("✖ Environment variable KAPACITOR_URL not found.")
	}
}

func main() {
	checkEnv()

	pool := utils.GetDB(os.Getenv("DATABASE_URL"))

	_, migrate := os.LookupEnv("RUN_MIGRATION")
	if migrate {
		fmt.Println("Detected $RUN_MIGRATION environment variable")
		runMigration(pool)
	} else {
		initDB(pool)
	}

	router := gin.Default()
	router.Use(dbMiddleware(pool))

	router.POST("/task/memory", memory.ProcessInstanceMemoryRequest)
	router.PATCH("/task/memory", memory.ProcessInstanceMemoryRequest)
	router.DELETE("/task/memory/:app/:id", memory.DeleteMemoryTask)
	router.GET("/tasks/memory/:app", memory.GetMemoryTasksForApp)
	router.GET("/tasks/memory/:app/:id", memory.GetMemoryTask)
	router.GET("/tasks/memory", memory.ListMemoryTasks)

	router.POST("/task/5xx", _5xx.Process5xxRequest)
	router.PATCH("/task/5xx", _5xx.Process5xxRequest)
	router.DELETE("/task/5xx/:app", _5xx.Delete5xxTask)
	router.GET("/task/5xx/:app", _5xx.Get5xxTask)
	router.GET("/task/5xx/:app/state", _5xx.Get5xxTaskState)
	router.GET("/tasks/5xx", _5xx.List5xxTasks)

	router.POST("/task/release", released.ProcessReleaseRequest)
	router.GET("/task/release/:app", released.GetReleaseTask)
	router.PATCH("/task/release", released.ProcessReleaseRequest)
	router.DELETE("/task/release/:app", released.DeleteReleaseTask)
	router.GET("/tasks/release", released.ListReleaseTasks)

	router.POST("/task/crashed", crashed.ProcessCrashedRequest)
	router.GET("/task/crashed/:app", crashed.GetCrashedTask)
	router.PATCH("/task/crashed", crashed.ProcessCrashedRequest)
	router.DELETE("/task/crashed/:app", crashed.DeleteCrashedTask)
	router.GET("/tasks/crashed", crashed.ListCrashedTasks)

	router.Run()
}
