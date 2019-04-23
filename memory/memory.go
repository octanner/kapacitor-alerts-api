package memory

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
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

// ProcessInstanceMemoryRequest - Create structs for creating and updating memory tasks
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
		createInstanceMemoryTask(task, c)
	} else if c.Request.Method == "PATCH" {
		deleteInstanceMemoryTask(task.ID, c)
		createInstanceMemoryTask(task, c)
	}
}

// createInstanceMemoryTask - Create memory task in Kapacitor and the database
func createInstanceMemoryTask(task MemoryTaskSpec, c *gin.Context) {
	db, err := utils.GetDBFromContext(c)
	if err != nil {
		utils.ReportError(err, c, "Unable to access database")
		return
	}

	client := http.Client{}

	p, err := json.Marshal(task)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	req, err := http.NewRequest("POST", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks", bytes.NewBuffer(p))
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	defer resp.Body.Close()
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	if resp.StatusCode != 200 {
		var er structs.ErrorResponse
		err = json.Unmarshal(bodybytes, &er)
		utils.ReportError(err, c, "")
		return
	}

	_, err = db.Exec(
		"insert into memory_tasks values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		task.ID, task.App, task.Vars["dynotyperequest"].Value, task.Crit, task.Warn,
		task.Window, task.Every, task.Slack, task.Post, task.Email,
	)

	if err != nil {
		utils.ReportError(err, c, "Unable to save to database")
		return
	}

	c.String(201, "")
}

// DeleteMemoryTask - Verify that a task exists in the database before talking to Kapacitor
func DeleteMemoryTask(c *gin.Context) {
	app := c.Param("app")
	id := c.Param("id")

	db, err := utils.GetDBFromContext(c)
	if err != nil {
		utils.ReportError(err, c, "Unable to access database")
		return
	}

	task := MemoryDBTask{}

	err = db.Get(&task, "select * from memory_tasks where id=$1 and app=$2", id, app)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.ReportInvalidRequest(c, "Task not found")
		} else {
			utils.ReportError(err, c, "Unable to access database")
		}
		return
	}

	deleteInstanceMemoryTask(task.ID, c)
	c.String(200, "")
}

// deleteInstanceMemoryTask - Delete memory task from Kapacitor and the database
func deleteInstanceMemoryTask(id string, c *gin.Context) {
	db, err := utils.GetDBFromContext(c)
	if err != nil {
		utils.ReportError(err, c, "Unable to access database")
		return
	}

	client := http.Client{}
	req, err := http.NewRequest("DELETE", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks/"+id, nil)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	defer resp.Body.Close()
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	if resp.StatusCode != 204 {
		var er structs.ErrorResponse
		err = json.Unmarshal(bodybytes, &er)
		utils.ReportError(err, c, "")
		return
	}

	_, err = db.Exec("delete from memory_tasks where id=$1", id)
	if err != nil {
		utils.ReportError(err, c, "Unable to access database")
		return
	}
}

// GetMemoryTask - Get the config of a specific memory task from the database
func GetMemoryTask(c *gin.Context) {
	app := c.Param("app")
	id := c.Param("id")

	db, err := utils.GetDBFromContext(c)
	if err != nil {
		utils.ReportError(err, c, "Unable to access database")
		return
	}

	task := MemoryDBTask{}

	err = db.Get(&task, "select * from memory_tasks where id=$1 and app=$2", id, app)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(200, nil)
		} else {
			utils.ReportError(err, c, "Unable to access database")
		}
		return
	}

	c.JSON(200, task)
}

// GetMemoryTasksForApp - Get the config of all memory tasks for an app from the database
func GetMemoryTasksForApp(c *gin.Context) {
	app := c.Param("app")

	db, err := utils.GetDBFromContext(c)
	if err != nil {
		utils.ReportError(err, c, "Unable to access database")
		return
	}

	tasks := []MemoryDBTask{}

	err = db.Select(&tasks, "select * from memory_tasks where app=$1", app)
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

// ListMemoryTasks - Get the config of all memory tasks from the database
func ListMemoryTasks(c *gin.Context) {
	db, err := utils.GetDBFromContext(c)
	if err != nil {
		utils.ReportError(err, c, "Unable to access database")
		return
	}

	tasks := []MemoryDBTask{}

	err = db.Select(&tasks, "select * from memory_tasks order by app asc")
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
