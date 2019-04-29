package crashed

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
*    POST     /task/crashed         TestCreateCrashedTask
*    PATCH    /task/crashed         TestUpdateCrashedTask
*    DELETE 	/task/crashed/:app    TestDeleteCrashedTask
*    GET      /tasks/crashed        TestCreateCrashedTask
*    GET      /task/crashed/:app    TestCreateCrashedTask
 */

// setupRouter - Setup Gin routes for current test type
func setupRouter() *gin.Engine {
	pool := utils.GetDB(os.Getenv("DATABASE_URL"))
	if pool == nil {
		log.Panicln("Unable to connect to database")
	}

	router := gin.Default()
	gin.SetMode(gin.DebugMode)
	router.Use(utils.DBMiddleware(pool))

	router.POST("/task/crashed", ProcessCrashedRequest)
	router.GET("/task/crashed/:app", GetCrashedTask)
	router.PATCH("/task/crashed", ProcessCrashedRequest)
	router.DELETE("/task/crashed/:app", DeleteCrashedTask)
	router.GET("/tasks/crashed", ListCrashedTasks)

	return router
}

// TestCreateCrashedTask - Make sure that creating a crashed task works and we can successfully get info about the created task
func TestCreateCrashedTask(t *testing.T) {
	router := setupRouter()

	// Create a new task
	var task CrashedDBTask
	task.App = "gotest-voltron"
	task.Slack = zero.StringFrom("#cobra")
	taskBytes, err := json.Marshal(task)
	assert.Nil(t, err, "Converting from CrashedDBTask to JSON should not throw an error")

	req, _ := http.NewRequest("POST", "/task/crashed", bytes.NewBuffer(taskBytes))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "HTTP response code for POST /task/crashed should be 201")

	// Check that the new task exists and contains expected data
	req, _ = http.NewRequest("GET", "/task/crashed/gotest-voltron", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "HTTP response code for GET /task/crashed/:app should be 200")

	var returnedTask CrashedDBTask
	err = json.Unmarshal([]byte(w.Body.String()), &returnedTask)

	assert.Nil(t, err, "Converting from JSON to CrashedDBTask should not throw an error")
	assert.Equal(t, returnedTask.App, task.App, "Task app name should match")
	assert.Equal(t, returnedTask.Slack, task.Slack, "Task slack should match")
	assert.Equal(t, returnedTask.Email, task.Email, "Task email should match")
	assert.Equal(t, returnedTask.Post, task.Post, "Task post should match")

	// Check that the new task exists in the list of all tasks
	req, _ = http.NewRequest("GET", "/tasks/crashed", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "HTTP response code for GET /tasks/crashed should be 200")

	var returnedTasks []CrashedDBTask
	err = json.Unmarshal([]byte(w.Body.String()), &returnedTasks)

	assert.Nil(t, err, "Converting from JSON to []CrashedDBTask should not throw an error")

	var foundTask CrashedDBTask

	for _, n := range returnedTasks {
		if n.App == task.App {
			foundTask = n
		}
	}

	assert.NotNil(t, foundTask, "Task should exist in returned list of tasks")

	assert.Equal(t, foundTask.App, task.App, "Task app name should match")
	assert.Equal(t, foundTask.Slack, task.Slack, "Task slack should match")
	assert.Equal(t, foundTask.Email, task.Email, "Task email should match")
	assert.Equal(t, foundTask.Post, task.Post, "Task post should match")
}

// TestUpdateCrashedTask - Make sure that updating a crashed task's config works
func TestUpdateCrashedTask(t *testing.T) {
	router := setupRouter()

	// Create a task with updated information
	// Slack - #cobra => nil
	// Post - nil => http://example.com/
	var task CrashedDBTask
	task.App = "gotest-voltron"
	task.Post = zero.StringFrom("http://example.com/")
	taskBytes, err := json.Marshal(task)
	assert.Nil(t, err, "Converting from CrashedDBTask to JSON should not throw an error")

	req, _ := http.NewRequest("PATCH", "/task/crashed", bytes.NewBuffer(taskBytes))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "HTTP response code for PATCH /task/crashed should be 201")

	// Check that the new task exists and contains expected data
	req, _ = http.NewRequest("GET", "/task/crashed/gotest-voltron", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "HTTP response code for GET /task/crashed/:app should be 200")

	var returnedTask CrashedDBTask
	err = json.Unmarshal([]byte(w.Body.String()), &returnedTask)

	// API sets nil slack channels to #
	task.Slack = zero.StringFrom("#")

	assert.Nil(t, err, "Converting from JSON to CrashedDBTask should not throw an error")
	assert.Equal(t, returnedTask.App, task.App, "Task app name should match")
	assert.Equal(t, returnedTask.Slack, task.Slack, "Task slack should match")
	assert.Equal(t, returnedTask.Email, task.Email, "Task email should match")
	assert.Equal(t, returnedTask.Post, task.Post, "Task post should match")
}

// TestDeleteCrashedTask - Make sure that deleting a crashed task works and that we cannot access it anymore
func TestDeleteCrashedTask(t *testing.T) {
	router := setupRouter()

	// Delete the task that was created in TestCreateCrashedTask
	req, _ := http.NewRequest("DELETE", "/task/crashed/gotest-voltron", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "HTTP response code for DELETE /task/crashed/:app should be 200")

	// Check that the deleted task doesn't exist anymore
	req, _ = http.NewRequest("GET", "/task/crashed/gotest-voltron", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "HTTP response code for GET /task/crashed/:app on invalid app should be 404")
}
