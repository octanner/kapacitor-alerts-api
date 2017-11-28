package main

import (
	"github.com/gin-gonic/gin"
	_5xx "kapacitor-alerts-api/5xx"
	memory "kapacitor-alerts-api/memory"
)

func main() {

	router := gin.Default()
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


	router.Run()

}
