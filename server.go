package main

import (
	"github.com/gin-gonic/gin"
	_5xx "kapacitor-alerts-api/5xx"
	memory "kapacitor-alerts-api/memory"
        release "kapacitor-alerts-api/release"
        crashed "kapacitor-alerts-api/crashed"
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

        router.POST("/task/release", release.ProcessReleaseRequest)
        router.GET("/task/release/:app", release.GetReleaseTaskForApp)
        router.PATCH("/task/release", release.ProcessReleaseRequest)
        router.DELETE("/task/release/:app", release.DeleteReleaseTask)
        router.GET("/tasks/release", release.ListReleaseTasks)

        router.POST("/task/crashed", crashed.ProcessCrashedRequest)
        router.GET("/task/crashed/:app", crashed.GetCrashedTaskForApp)
        router.PATCH("/task/crashed", crashed.ProcessCrashedRequest)
        router.DELETE("/task/crashed/:app", crashed.DeleteCrashedTask)
        router.GET("/tasks/crashed", crashed.ListCrashedTasks)
	router.Run()

}
