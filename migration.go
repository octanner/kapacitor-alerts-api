package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"kapacitor-alerts-api/structs"
	utils "kapacitor-alerts-api/utils"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
)

type kapTask struct {
	ID   string `json:"id"`
	Vars struct {
		App       structs.Var `json:"app"`
		Crit      structs.Var `json:"crit,omitempty"`
		Dyno      structs.Var `json:"dynotyperequest,omitempty"`
		Email     structs.Var `json:"email,omitempty"`
		Every     structs.Var `json:"every,omitempty"`
		Post      structs.Var `json:"post,omitempty"`
		Slack     structs.Var `json:"slack,omitempty"`
		Tolerance structs.Var `json:"tolerance,omitempty"`
		Warn      structs.Var `json:"warn,omitempty"`
		Window    structs.Var `json:"window,omitempty"`
	} `json:"vars"`
}

type kapResponse struct {
	Tasks []kapTask `json:"tasks"`
}

// checkTarget - Ensure that slack, post, and email have values and are not nil
func checkTarget(task kapTask) (string, string, string) {
	var slack, post, email string
	if task.Vars.Slack.Value == nil {
		slack = ""
	} else {
		slack = task.Vars.Slack.Value.(string)
	}

	if task.Vars.Post.Value == nil {
		post = ""
	} else {
		post = task.Vars.Post.Value.(string)
	}

	if task.Vars.Email.Value == nil {
		email = ""
	} else {
		email = task.Vars.Email.Value.(string)
	}

	return slack, post, email
}

// saveMemoryTask - Save a memory task to the database
func saveMemoryTask(task kapTask, db *sqlx.DB) error {
	slack, post, email := checkTarget(task)

	_, err := db.Exec(
		"INSERT INTO memory_tasks VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		task.ID, task.Vars.App.Value.(string), task.Vars.Dyno.Value.(string),
		task.Vars.Crit.Value.(float64), task.Vars.Warn.Value.(float64),
		task.Vars.Window.Value.(string), task.Vars.Every.Value.(string),
		slack, post, email,
	)
	return err
}

// save5xxTask - Save a 5xx task to the database
func save5xxTask(task kapTask, db *sqlx.DB) error {
	slack, post, email := checkTarget(task)

	_, err := db.Exec(
		"INSERT INTO _5xx_tasks VALUES ($1, $2, $3, $4, $5)",
		task.Vars.App.Value.(string), task.Vars.Tolerance.Value.(string),
		slack, post, email,
	)
	return err
}

// saveCrashedTask - Save a crashed task to the database
func saveCrashedTask(task kapTask, db *sqlx.DB) error {
	slack, post, email := checkTarget(task)

	_, err := db.Exec(
		"INSERT INTO crashed_tasks VALUES ($1, $2, $3, $4)",
		task.Vars.App.Value.(string), slack, post, email,
	)
	return err
}

// saveReleasedTask - Save a released task to the database
func saveReleasedTask(task kapTask, db *sqlx.DB) error {
	slack, post, email := checkTarget(task)

	_, err := db.Exec(
		"INSERT INTO released_tasks VALUES ($1, $2, $3, $4)",
		task.Vars.App.Value.(string), slack, post, email,
	)
	return err
}

// runMigration - Clear the database and import all tasks from Kapacitor
func runMigration(db *sqlx.DB) {
	fmt.Println()
	fmt.Println("==============================")
	fmt.Println("      DATABASE MIGRATION      ")
	fmt.Println("==============================")
	fmt.Println()

	start := time.Now()
	reg, _ := regexp.Compile(`^.*((-sample\.memory_total-(\w+))|-(release|5xx|crash))$`)

	fmt.Println("Re-creating database...")
	// Drop all tables from the database (if exists)
	_, err := db.Exec(`
		do $$
		begin
			DROP TABLE IF EXISTS memory_tasks;
			DROP TABLE IF EXISTS _5xx_tasks;
			DROP TABLE IF EXISTS crashed_tasks;
			DROP TABLE IF EXISTS released_tasks;
		end
		$$;
	`)
	if err != nil {
		fmt.Println("✖ Error: Unable to migrate database from Kapacitor - error clearing database")
		log.Fatalln(err)
	}

	// Recreate tables
	utils.InitDB(db)

	fmt.Println("✓ Database successfully recreated.")
	fmt.Println()

	fmt.Println("Fetching tasks from Kapacitor...")
	// Get all tasks from Kapacitor
	resp, err := http.Get(os.Getenv("KAPACITOR_URL") + "/kapacitor/v1/tasks?pattern=*")
	if err != nil {
		fmt.Println("✖ Error: Unable to migrate database from Kapacitor - error fetching tasks from Kapacitor")
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("✖ Error: Unable to migrate database from Kapacitor - error fetching tasks from Kapacitor")
		log.Fatalln(err)
	}

	var k kapResponse
	err = json.Unmarshal(bodybytes, &k)
	if err != nil {
		fmt.Println("✖ Error: Unable to migrate database from Kapacitor - error fetching tasks from Kapacitor")
		log.Fatalln(err)
	}

	fmt.Println("✓ " + strconv.Itoa(len(k.Tasks)) + " tasks fetched from the Kapacitor API.")
	fmt.Println()

	fmt.Println("Importing tasks into the database...")
	fmt.Println()
	var success, fail int

	// For each task, determine type and save config to the appropriate database
	for _, task := range k.Tasks {
		res := reg.FindStringSubmatch(task.ID)
		if res == nil {
			fmt.Println("Skipping " + task.ID + "...")
			continue
		}

		if res[3] != "" {
			err = saveMemoryTask(task, db)
		} else if res[4] != "" {
			if res[4] == "release" {
				err = saveReleasedTask(task, db)
			} else if res[4] == "5xx" {
				err = save5xxTask(task, db)
			} else if res[4] == "crash" {
				err = saveCrashedTask(task, db)
			} else {
				continue
			}
		}

		if err != nil {
			fmt.Println("✖ Error: Could not migrate " + task.ID + " to the database.")
			fail++
		} else {
			fmt.Println("✓ Successfully imported " + task.ID + " to the database.")
			success++
		}
	}
	fmt.Println("✓ Imported " + strconv.Itoa(success) + " tasks in " + time.Since(start).String())
	if fail > 0 {
		fmt.Println("✖ There were " + strconv.Itoa(success) + "errors, see fmt for details.")
	} else {
		fmt.Println("✓ All memory, 5xx, crashed, and released tasks successfully imported.")
	}

	fmt.Println()
	fmt.Println("Database migration complete!")
	fmt.Println()
}
