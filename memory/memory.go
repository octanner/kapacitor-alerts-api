package memory

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	structs "kapacitor-alerts-api/structs"
	utils "kapacitor-alerts-api/utils"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
)

const memoryalerttemplate = `
	batch
    |	query('''
				select mean(value)/1024/1024 as value from "opentsdb"."autogen"."[[ .Metric ]]" where "app"='[[ .App ]]' and "dyno" [[ .Dynotype ]]
    	''')
        .period([[ .Window ]])
        .every([[ .Every ]])
        .groupBy('app','dyno')
    |	eval(lambda: ceil("value")).as('rvalue').keep('value','rvalue')
    |	alert()
        .crit(lambda: "value" > [[ .Crit ]])
        .warn(lambda: "value" > [[ .Warn ]])
        .stateChangesOnly()
        [[if .Slack ]]
        	.slack()
        	.channel('[[ .Slack ]]')
        [[end]]
        .message('Memory is {{ .Level }} for {{ .Group }} : {{ index .Fields "rvalue" }} MB - limits [[ .Warn ]]/[[ .Crit ]]')
        .details('''
					<h3>{{ .Message }}</h3>
					<h3>Value: {{ index .Fields "rvalue" }}</h3>
				''')
				[[if .Email]]
					[[ range $email := .EmailArray ]] 
						.email('[[ $email ]]')
					[[end]]
				[[end]]
        [[if .Post]]
        	.post('[[ .Post ]]')
        [[end]]
`

// getTaskByID - Get a task from the database by its ID
func getTaskByID(id string, c *gin.Context) (*MemoryDBTask, error) {
	db, err := utils.GetDBFromContext(c)
	if err != nil {
		return nil, errors.New("Unable to access database")
	}

	task := MemoryDBTask{}

	err = db.Get(&task, "SELECT * FROM memory_tasks WHERE id=$1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.New("Unable to access database")
	}
	return &task, nil
}

func getTaskByNameAndDyno(app string, dyno string, c *gin.Context) (*MemoryDBTask, error) {
	db, err := utils.GetDBFromContext(c)
	if err != nil {
		return nil, errors.New("Unable to access database")
	}

	task := MemoryDBTask{}

	err = db.Get(&task, "SELECT * FROM memory_tasks WHERE app=$1 AND dynotype=$2", app, dyno)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.New("Unable to access database")
	}
	return &task, nil
}

// ProcessInstanceMemoryRequest - POST | PATCH /task/memory
func ProcessInstanceMemoryRequest(c *gin.Context) {
	var vars map[string]structs.Var
	vars = make(map[string]structs.Var)
	var dbrps []structs.DbrpSpec
	var dbrp structs.DbrpSpec
	var task MemoryTaskSpec

	bodybytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	err = json.Unmarshal(bodybytes, &task)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	task.ID = task.App + "-memory"
	task.Metric = "sample.memory_total"
	vars = utils.AddVar("metric", task.Metric, "string", vars)
	vars = utils.AddVar("dynotyperequest", task.Dynotype, "string", vars)

	if task.Dynotype != "all" {
		task.ID = task.App + "-" + task.Metric + "-" + task.Dynotype
	} else {
		task.ID = task.App + "-" + task.Metric + "-all"
	}
	vars = utils.AddVar("id", task.ID, "string", vars)

	task.Type = "batch"
	vars = utils.AddVar("type", task.Type, "string", vars)

	dbrp.Db = "opentsdb"
	dbrp.Rp = "autogen"
	dbrps = append(dbrps, dbrp)

	if task.Dynotype == "all" {
		task.Dynotype = " =~ /.*/ "
	} else if task.Dynotype == "web" {
		task.Dynotype = " !~ /--/ "
	} else {
		task.Dynotype = " =~ /" + task.Dynotype + "/ "
	}
	vars = utils.AddVar("dynotype", task.Dynotype, "string", vars)

	task.Dbrps = dbrps
	task.Script = ""
	task.Status = "enabled"

	if !strings.HasPrefix(task.Slack, "#") && !strings.HasPrefix(task.Slack, "@") {
		task.Slack = "#" + task.Slack
	}

	task.EmailArray = strings.Split(task.Email, ",")

	t := template.Must(template.New("memoryalerttemplate").Delims("[[", "]]").Parse(memoryalerttemplate))

	var sb bytes.Buffer
	swr := bufio.NewWriter(&sb)
	err = t.Execute(swr, task)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	swr.Flush()
	task.Script = string(sb.Bytes())
	vars = utils.AddVar("app", task.App, "string", vars)
	vars = utils.AddVar("crit", task.Crit, "int", vars)
	vars = utils.AddVar("warn", task.Warn, "int", vars)
	vars = utils.AddVar("slack", task.Slack, "string", vars)
	vars = utils.AddVar("window", task.Window, "string", vars)
	vars = utils.AddVar("every", task.Every, "string", vars)
	vars = utils.AddVar("post", task.Post, "string", vars)
	vars = utils.AddVar("email", task.Email, "string", vars)

	task.Vars = vars

	bodybytes, err = json.Marshal(task)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	if c.Request.Method == "POST" {
		err = createInstanceMemoryTask(task, c)
		if err != nil {
			utils.ReportError(err, c, "")
			return
		}
	}

	if c.Request.Method == "PATCH" {
		// Check if task exists before trying to patch it
		_, err := getTaskByID(task.ID, c)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(404, nil)
			} else {
				utils.ReportError(err, c, "")
			}
			return
		}

		err = deleteInstanceMemoryTask(task.ID, c)
		if err != nil {
			utils.ReportError(err, c, "")
			return
		}

		err = createInstanceMemoryTask(task, c)
		if err != nil {
			utils.ReportError(err, c, "")
			return
		}
	}

	c.String(201, "")
}

// createInstanceMemoryTask - Create memory task in Kapacitor and save config to the database
func createInstanceMemoryTask(task MemoryTaskSpec, c *gin.Context) error {
	db, err := utils.GetDBFromContext(c)
	if err != nil {
		return errors.New("Unable to access database")
	}

	client := http.Client{}

	p, err := json.Marshal(task)
	if err != nil {
		return errors.New("Server Error while reading response")
	}

	req, err := http.NewRequest("POST", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks", bytes.NewBuffer(p))
	if err != nil {
		return errors.New("Server Error while reading response")
	}

	resp, err := client.Do(req)
	if err != nil {
		return errors.New("Server Error while reading response")
	}

	defer resp.Body.Close()
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("Server Error while reading response")
	}

	if resp.StatusCode != 200 {
		var er structs.ErrorResponse
		err = json.Unmarshal(bodybytes, &er)
		if err != nil {
			return errors.New("Server Error while reading response")
		}
		return errors.New(er.Error)
	}

	_, err = db.Exec(
		"INSERT INTO memory_tasks VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		task.ID, task.App, task.Vars["dynotyperequest"].Value, task.Crit, task.Warn,
		task.Window, task.Every, task.Slack, task.Post, task.Email,
	)

	if err != nil {
		return errors.New("Unable to save to database")
	}

	return nil
}

// DeleteMemoryTask - DELETE /task/memory/:app/:dyno
func DeleteMemoryTask(c *gin.Context) {
	app := c.Param("app")
	dyno := c.Param("dyno")

	// Check if task exists before trying to delete it
	task, err := getTaskByNameAndDyno(app, dyno, c)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, nil)
		} else {
			utils.ReportError(err, c, "")
		}
		return
	}

	err = deleteInstanceMemoryTask(task.ID, c)
	if err != nil {
		utils.ReportError(err, c, "")
		return
	}

	c.String(200, "")
}

// deleteInstanceMemoryTask - Delete memory task from Kapacitor and the database
func deleteInstanceMemoryTask(id string, c *gin.Context) error {
	db, err := utils.GetDBFromContext(c)
	if err != nil {
		return errors.New("Unable to access database")
	}

	client := http.Client{}
	req, err := http.NewRequest("DELETE", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks/"+id, nil)
	if err != nil {
		return errors.New("Server Error while reading response")
	}

	resp, err := client.Do(req)
	if err != nil {
		return errors.New("Server Error while reading response")
	}

	defer resp.Body.Close()
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("Server Error while reading response")
	}

	if resp.StatusCode != 204 {
		var er structs.ErrorResponse
		err = json.Unmarshal(bodybytes, &er)
		if err != nil {
			return errors.New("Server Error while reading response")
		}
		return errors.New(er.Error)
	}

	_, err = db.Exec("DELETE FROM memory_tasks WHERE id=$1", id)
	if err != nil {
		return errors.New("Unable to access database")
	}

	return nil
}

// GetMemoryTask - GET /task/memory/:app/:dyno
func GetMemoryTask(c *gin.Context) {
	app := c.Param("app")
	dyno := c.Param("dyno")

	task, err := getTaskByNameAndDyno(app, dyno, c)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, nil)
		} else {
			utils.ReportError(err, c, "")
		}
		return
	}

	c.JSON(200, task)
}

// GetMemoryTasksForApp - GET /tasks/memory/:app
func GetMemoryTasksForApp(c *gin.Context) {
	app := c.Param("app")

	db, err := utils.GetDBFromContext(c)
	if err != nil {
		utils.ReportError(err, c, "Unable to access database")
		return
	}

	tasks := []MemoryDBTask{}

	err = db.Select(&tasks, "SELECT * FROM memory_tasks WHERE app=$1 ORDER BY id ASC", app)
	if err != nil {
		utils.ReportError(err, c, "Unable to access database")
		return
	}

	if len(tasks) == 0 {
		c.JSON(200, nil)
		return
	}

	c.JSON(200, tasks)
}

// ListMemoryTasks - GET /tasks/memory
func ListMemoryTasks(c *gin.Context) {
	db, err := utils.GetDBFromContext(c)
	if err != nil {
		utils.ReportError(err, c, "Unable to access database")
		return
	}

	tasks := []MemoryDBTask{}

	err = db.Select(&tasks, "SELECT * FROM memory_tasks ORDER BY app ASC")
	if err != nil {
		utils.ReportError(err, c, "Unable to access database")
		return
	}

	if len(tasks) == 0 {
		c.JSON(200, nil)
		return
	}

	c.JSON(200, tasks)
}
