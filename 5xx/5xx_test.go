package _5xx

import (
	"bytes"
	"encoding/json"
	"kapacitor-alerts-api/utils"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gopkg.in/guregu/null.v3/zero"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

/********************************************************
*    Endpoints tested:
*
*    Method   Endpoint              Function
*    ---------------------------------------------------
*    POST     /task/5xx             TestCreate5xxTask
*    PATCH    /task/5xx             TestUpdate5xxTask
*    DELETE 	/task/5xx/:app        TestDelete5xxTask
*    GET      /tasks/5xx            TestCreate5xxTask
*    GET      /task/5xx/:app        TestCreate5xxTask
*    GET      /task/5xx/:app/state  TestGet5xxTaskState
 */

// setupRouter - Setup Gin routes for current test type
func setupRouter() *gin.Engine {
	pool := utils.GetDB(os.Getenv("DATABASE_URL"))
	if pool == nil {
		log.Panicln("Unable to connect to database")
	}

	router := gin.Default()
	router.Use(utils.DBMiddleware(pool))

	router.POST("/task/5xx", Process5xxRequest)
	router.PATCH("/task/5xx", Process5xxRequest)
	router.DELETE("/task/5xx/:app", Delete5xxTask)
	router.GET("/task/5xx/:app", Get5xxTask)
	router.GET("/task/5xx/:app/state", Get5xxTaskState)
	router.GET("/tasks/5xx", List5xxTasks)

	return router
}

// TestCreate5xxTask - Make sure that creating a 5xx task works and we can successfully get info about the created task
func TestCreate5xxTask(t *testing.T) {
	router := setupRouter()

	// Create a new task
	var task _5xxDBTask
	task.App = "gotest-voltron"
	task.Tolerance = "low"
	task.Slack = zero.StringFrom("#cobra")
	taskBytes, err := json.Marshal(task)
	assert.Nil(t, err, "Converting from _5xxDBTask to JSON should not throw an error")

	req, _ := http.NewRequest("POST", "/task/5xx", bytes.NewBuffer(taskBytes))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "HTTP response code for POST /task/5xx should be 201")

	// Check that the new task exists and contains expected data
	req, _ = http.NewRequest("GET", "/task/5xx/gotest-voltron", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "HTTP response code for GET /task/5xx/:app should be 200")

	var returnedTask _5xxDBTask
	err = json.Unmarshal([]byte(w.Body.String()), &returnedTask)

	assert.Nil(t, err, "Converting from JSON to _5xxDBTask should not throw an error")
	assert.Equal(t, returnedTask.App, task.App, "Task app name should match")
	assert.Equal(t, returnedTask.Tolerance, task.Tolerance, "Task tolerance should match")
	assert.Equal(t, returnedTask.Slack, task.Slack, "Task slack should match")
	assert.Equal(t, returnedTask.Email, task.Email, "Task email should match")
	assert.Equal(t, returnedTask.Post, task.Post, "Task post should match")

	// Check that the new task exists in the list of all tasks
	req, _ = http.NewRequest("GET", "/tasks/5xx", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "HTTP response code for GET /tasks/5xx should be 200")

	var returnedTasks []_5xxDBTask
	err = json.Unmarshal([]byte(w.Body.String()), &returnedTasks)

	assert.Nil(t, err, "Converting from JSON to []_5xxDBTask should not throw an error")

	var foundTask _5xxDBTask

	for _, n := range returnedTasks {
		if n.App == task.App {
			foundTask = n
		}
	}

	assert.NotNil(t, foundTask, "Task should exist in returned list of tasks")

	assert.Equal(t, foundTask.App, task.App, "Task app name should match")
	assert.Equal(t, foundTask.Tolerance, task.Tolerance, "Task tolerance should match")
	assert.Equal(t, foundTask.Slack, task.Slack, "Task slack should match")
	assert.Equal(t, foundTask.Email, task.Email, "Task email should match")
	assert.Equal(t, foundTask.Post, task.Post, "Task post should match")
}

// TestGet5xxTaskStatus - Make sure that we can get status information about a 5xx task
func TestGet5xxTaskState(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/task/5xx/gotest-voltron/state", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "HTTP response code for GET /task/5xx/:app/state should be 200")

	var taskState _5xxSimpleTaskState
	err := json.Unmarshal([]byte(w.Body.String()), &taskState)

	assert.Nil(t, err, "Converting from JSON to _5xxSimpleTaskState should not throw an error")
	assert.Equal(t, taskState.App, "gotest-voltron", "Task app name should match")
	assert.Equal(t, taskState.State, "OK", "Task state should be OK")
}

// TestUpdate5xxTask - Make sure that updating a 5xx task's config works
func TestUpdate5xxTask(t *testing.T) {
	router := setupRouter()

	// Create a task with updated information
	// Slack - #cobra => nil
	// Post - nil => http://example.com/
	var task _5xxDBTask
	task.App = "gotest-voltron"
	task.Tolerance = "low"
	task.Post = zero.StringFrom("http://example.com/")
	taskBytes, err := json.Marshal(task)
	assert.Nil(t, err, "Converting from _5xxDBTask to JSON should not throw an error")

	req, _ := http.NewRequest("PATCH", "/task/5xx", bytes.NewBuffer(taskBytes))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "HTTP response code for PATCH /task/5xx should be 201")

	// Check that the new task exists and contains expected data
	req, _ = http.NewRequest("GET", "/task/5xx/gotest-voltron", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "HTTP response code for GET /task/5xx/:app should be 200")

	var returnedTask _5xxDBTask
	err = json.Unmarshal([]byte(w.Body.String()), &returnedTask)

	// API sets nil slack channels to #
	task.Slack = zero.StringFrom("#")

	assert.Nil(t, err, "Converting from JSON to _5xxDBTask should not throw an error")
	assert.Equal(t, returnedTask.App, task.App, "Task app name should match")
	assert.Equal(t, returnedTask.Tolerance, task.Tolerance, "Task tolerance should match")
	assert.Equal(t, returnedTask.Slack, task.Slack, "Task slack should match")
	assert.Equal(t, returnedTask.Email, task.Email, "Task email should match")
	assert.Equal(t, returnedTask.Post, task.Post, "Task post should match")
}

// TestDelete5xxTask - Make sure that deleting a 5xx task works and that we cannot access it anymore
func TestDelete5xxTask(t *testing.T) {
	router := setupRouter()

	// Delete the task that was created in TestCreate5xxTask
	req, _ := http.NewRequest("DELETE", "/task/5xx/gotest-voltron", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "HTTP response code for DELETE /task/5xx/:app should be 200")

	// Check that the deleted task doesn't exist anymore
	req, _ = http.NewRequest("GET", "/task/5xx/gotest-voltron", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "HTTP response code for GET /task/5xx/:app on invalid app should be 404")
}
