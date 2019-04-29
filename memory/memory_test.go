package memory

import (
	"bytes"
	"encoding/json"
	"errors"
	"kapacitor-alerts-api/utils"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"gopkg.in/guregu/null.v3/zero"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

/********************************************************
*    Endpoints tested:
*
*    Method   Endpoint                 Function
*    ---------------------------------------------------
*    POST     /task/memory             TestCreateMemoryTask
*    PATCH    /task/memory             TestUpdateMemoryTask
*    DELETE 	/task/memory/:app        TestDeleteMemoryTask
*    GET      /tasks/memory            TestCreateMemoryTask
*    GET      /tasks/memory/:app       TestCreateMemoryTask
*    GET      /tasks/memory/:app/id    TestCreateMemoryTask
 */

// setupRouter - Setup Gin routes for current test type
func setupRouter() *gin.Engine {
	pool := utils.GetDB(os.Getenv("DATABASE_URL"))
	if pool == nil {
		log.Panicln("Unable to connect to database")
	}

	router := gin.Default()
	router.Use(utils.DBMiddleware(pool))

	router.POST("/task/memory", ProcessInstanceMemoryRequest)
	router.PATCH("/task/memory", ProcessInstanceMemoryRequest)
	router.DELETE("/task/memory/:app/:id", DeleteMemoryTask)
	router.GET("/tasks/memory/:app", GetMemoryTasksForApp)
	router.GET("/tasks/memory/:app/:id", GetMemoryTask)
	router.GET("/tasks/memory", ListMemoryTasks)

	return router
}

// TestCreateMemoryTask - Make sure that creating a memory task works and we can successfully get info about the created task
func TestCreateMemoryTask(t *testing.T) {
	router := setupRouter()

	// Create two new tasks (testing multiple dyno types)
	var task1 MemoryDBTaskTest
	task1.ID = ""
	task1.App = "gotest-voltron"
	task1.Dynotype = "web"
	task1.Crit = "1000"
	task1.Warn = "750"
	task1.Wind = "12h"
	task1.Every = "1m"
	task1.Slack = zero.StringFrom("#cobra")

	var task2 MemoryDBTaskTest
	task2.ID = ""
	task2.App = "gotest-voltron"
	task2.Dynotype = "worker"
	task2.Crit = "750"
	task2.Warn = "500"
	task2.Wind = "6h"
	task2.Every = "30s"
	task2.Slack = zero.StringFrom("#cobra")

	task1Bytes, err := json.Marshal(task1)
	assert.Nil(t, err, "Converting from MemoryDBTaskTest to JSON should not throw an error")

	task2Bytes, err := json.Marshal(task2)
	assert.Nil(t, err, "Converting from MemoryDBTaskTest to JSON should not throw an error")

	req, _ := http.NewRequest("POST", "/task/memory", bytes.NewBuffer(task1Bytes))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "HTTP response code for POST /task/memory should be 201")

	req, _ = http.NewRequest("POST", "/task/memory", bytes.NewBuffer(task2Bytes))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "HTTP response code for POST /task/memory should be 201")

	// Check that both tasks exist, contain expected data
	req, _ = http.NewRequest("GET", "/tasks/memory/gotest-voltron", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "HTTP response code for GET /tasks/memory/:app should be 200")

	var returnedTasks []MemoryDBTask
	err = json.Unmarshal([]byte(w.Body.String()), &returnedTasks)
	assert.Nil(t, err, "Converting from JSON to MemoryDBTaskTest should not throw an error")

	// Check each task for valid data
	for _, returnedTask := range returnedTasks {
		var compareTask MemoryDBTaskTest
		if returnedTask.Dynotype == "web" {
			compareTask = task1
			task1.ID = returnedTask.ID
		} else if returnedTask.Dynotype == "worker" {
			compareTask = task2
			task2.ID = returnedTask.ID
		} else {
			assert.Error(t, errors.New("Invalid dynotype"), "Task list should not contain a dynotype other than web or worker")
		}

		crit, _ := strconv.Atoi(compareTask.Crit)
		warn, _ := strconv.Atoi(compareTask.Warn)

		assert.Equal(t, returnedTask.App, compareTask.App, "Task app name should match")
		assert.Equal(t, returnedTask.Dynotype, compareTask.Dynotype, "Task dynotype should match")
		assert.Equal(t, returnedTask.Crit, crit, "Task crit should match")
		assert.Equal(t, returnedTask.Warn, warn, "Task warn should match")
		assert.Equal(t, returnedTask.Wind, compareTask.Wind, "Task window should match")
		assert.Equal(t, returnedTask.Every, compareTask.Every, "Task every should match")
		assert.Equal(t, returnedTask.Slack, compareTask.Slack, "Task slack should match")
		assert.Equal(t, returnedTask.Email, compareTask.Email, "Task email should match")
		assert.Equal(t, returnedTask.Post, compareTask.Post, "Task post should match")
	}

	// Check individual task endpoints
	for _, returnedTask := range returnedTasks {
		req, _ = http.NewRequest("GET", "/tasks/memory/gotest-voltron/"+returnedTask.ID, nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "HTTP response code for GET /tasks/memory/:app/:id should be 200")

		var verifyTask MemoryDBTask
		err = json.Unmarshal([]byte(w.Body.String()), &verifyTask)
		assert.Nil(t, err, "Converting from JSON to MemoryDBTaskTest should not throw an error")

		assert.Equal(t, returnedTask.App, verifyTask.App, "Task app name should match")
		assert.Equal(t, returnedTask.Dynotype, verifyTask.Dynotype, "Task dynotype should match")
		assert.Equal(t, returnedTask.Crit, verifyTask.Crit, "Task crit should match")
		assert.Equal(t, returnedTask.Warn, verifyTask.Warn, "Task warn should match")
		assert.Equal(t, returnedTask.Wind, verifyTask.Wind, "Task window should match")
		assert.Equal(t, returnedTask.Every, verifyTask.Every, "Task every should match")
		assert.Equal(t, returnedTask.Slack, verifyTask.Slack, "Task slack should match")
		assert.Equal(t, returnedTask.Email, verifyTask.Email, "Task email should match")
		assert.Equal(t, returnedTask.Post, verifyTask.Post, "Task post should match")
	}

	// Check that the new tasks exist in the list of all tasks
	req, _ = http.NewRequest("GET", "/tasks/memory", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "HTTP response code for GET /tasks/memory should be 200")

	var returnedTasks2 []MemoryDBTask
	err = json.Unmarshal([]byte(w.Body.String()), &returnedTasks2)

	assert.Nil(t, err, "Converting from JSON to []MemoryDBTask should not throw an error")

	found := []bool{false, false}

	for _, returnedTask := range returnedTasks2 {
		var verifyTask MemoryDBTaskTest
		if returnedTask.ID == task1.ID {
			verifyTask = task1
			found[0] = true
		} else if returnedTask.ID == task2.ID {
			verifyTask = task2
			found[1] = true
		} else {
			continue
		}

		crit, _ := strconv.Atoi(verifyTask.Crit)
		warn, _ := strconv.Atoi(verifyTask.Warn)

		assert.Equal(t, returnedTask.App, verifyTask.App, "Task app name should match")
		assert.Equal(t, returnedTask.Dynotype, verifyTask.Dynotype, "Task dynotype should match")
		assert.Equal(t, returnedTask.Crit, crit, "Task crit should match")
		assert.Equal(t, returnedTask.Warn, warn, "Task warn should match")
		assert.Equal(t, returnedTask.Wind, verifyTask.Wind, "Task window should match")
		assert.Equal(t, returnedTask.Every, verifyTask.Every, "Task every should match")
		assert.Equal(t, returnedTask.Slack, verifyTask.Slack, "Task slack should match")
		assert.Equal(t, returnedTask.Email, verifyTask.Email, "Task email should match")
		assert.Equal(t, returnedTask.Post, verifyTask.Post, "Task post should match")
	}

	assert.True(t, found[0], "Task 1 should exist in list of all tasks")
	assert.True(t, found[1], "Task 2 should exist in list of all tasks")
}

// TestUpdateMemoryTask - Make sure that updating a memory task's config works
func TestUpdateMemoryTask(t *testing.T) {
	router := setupRouter()

	// Create a task with updated information
	// Crit - 1000 => 750
	// Warn - 750 => 500
	// Wind - 12h => 1d
	// Every - 1m => 2m
	// Slack - #cobra => nil
	// Post - nil => http://example.com/
	var task MemoryDBTaskTest
	task.ID = ""
	task.App = "gotest-voltron"
	task.Dynotype = "web"
	task.Crit = "1000"
	task.Warn = "750"
	task.Wind = "1d"
	task.Every = "2m"
	task.Post = zero.StringFrom("http://example.com")

	taskBytes, err := json.Marshal(task)
	assert.Nil(t, err, "Converting from MemoryDBTaskTest to JSON should not throw an error")

	req, _ := http.NewRequest("PATCH", "/task/memory", bytes.NewBuffer(taskBytes))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "HTTP response code for PATCH /task/memory should be 201")

	// Check that the new task exists and contains expected data
	req, _ = http.NewRequest("GET", "/tasks/memory/gotest-voltron/gotest-voltron-sample.memory_total-web", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "HTTP response code for GET /tasks/memory/:app/:id should be 200")

	var returnedTask MemoryDBTask
	err = json.Unmarshal([]byte(w.Body.String()), &returnedTask)
	assert.Nil(t, err, "JSON error")

	// API sets nil slack channels to #
	task.Slack = zero.StringFrom("#")

	crit, _ := strconv.Atoi(task.Crit)
	warn, _ := strconv.Atoi(task.Warn)

	assert.Equal(t, returnedTask.App, task.App, "Task app name should match")
	assert.Equal(t, returnedTask.Dynotype, task.Dynotype, "Task dynotype should match")
	assert.Equal(t, returnedTask.Crit, crit, "Task crit should match")
	assert.Equal(t, returnedTask.Warn, warn, "Task warn should match")
	assert.Equal(t, returnedTask.Wind, task.Wind, "Task window should match")
	assert.Equal(t, returnedTask.Every, task.Every, "Task every should match")
	assert.Equal(t, returnedTask.Slack, task.Slack, "Task slack should match")
	assert.Equal(t, returnedTask.Email, task.Email, "Task email should match")
	assert.Equal(t, returnedTask.Post, task.Post, "Task post should match")
}

// TestDeleteMemoryTask - Make sure that deleting a memory task works and that we cannot access it anymore
func TestDeleteMemoryTask(t *testing.T) {
	router := setupRouter()

	// Delete the tasks that were created in TestCreateMemoryTask
	req, _ := http.NewRequest("DELETE", "/task/memory/gotest-voltron/gotest-voltron-sample.memory_total-web", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "HTTP response code for DELETE /task/memory/:app/:id should be 200")

	req, _ = http.NewRequest("DELETE", "/task/memory/gotest-voltron/gotest-voltron-sample.memory_total-worker", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "HTTP response code for DELETE /task/memory/:app/:id should be 200")

	// Check that the deleted tasks don't exist anymore
	req, _ = http.NewRequest("GET", "/tasks/memory/gotest-voltron/gotest-voltron-sample.memory_total-web", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "HTTP response code for GET /task/memory/:app/:id on invalid app should be 404")

	req, _ = http.NewRequest("GET", "/tasks/memory/gotest-voltron/gotest-voltron-sample.memory_total-worker", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "HTTP response code for GET /task/memory/:app/:id on invalid app should be 404")

	req, _ = http.NewRequest("GET", "/tasks/memory/gotest-voltron", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]string
	err := json.Unmarshal([]byte(w.Body.String()), &response)
	assert.Nil(t, err, "Converting from JSON to map[string]string should not throw an error")

	assert.Equal(t, http.StatusOK, w.Code, "HTTP response for GET /tasks/memory/:app should be 200")
	assert.Equal(t, len(response), 0, "Result for GET /tasks/memory/:app should be empty when no tasks are present")
}
