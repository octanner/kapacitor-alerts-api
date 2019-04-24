package main

import (
	"io/ioutil"
	_5xx "kapacitor-alerts-api/5xx"
	crashed "kapacitor-alerts-api/crashed"
	memory "kapacitor-alerts-api/memory"
	released "kapacitor-alerts-api/released"
	"kapacitor-alerts-api/utils"
	"log"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// InitDB - Run any available creation and migration scripts
func InitDB(db *sqlx.DB) {
	buf, err := ioutil.ReadFile("./create.sql")
	if err != nil {
		log.Println("Error: Unable to run migration scripts, could not load create.sql.")
		log.Fatalln(err)
	}
	_, err = db.Query(string(buf))
	if err != nil {
		log.Println("Error: Unable to run migration scripts, execution failed.")
		log.Fatalln(err)
	}
}

// DbMiddleware - Add a SQL database connection to the Gin context
func DbMiddleware(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

func main() {
	router := gin.Default()

	url := os.Getenv("DATABASE_URL")
	pool := utils.GetDB(url)
	InitDB(pool)

	router.Use(DbMiddleware(pool))

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
