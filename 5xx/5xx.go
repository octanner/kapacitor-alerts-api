package _5xx

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io/ioutil"
	structs "kapacitor-alerts-api/structs"
	"kapacitor-alerts-api/utils"
	"net/http"
	"os"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
)

const _5xxalerttemplate = `batch
    |query('''
select count("value") from "opentsdb"."autogen"./router.status.(5.*)/ where "fqdn"='[[ .Fqdn ]]'
    ''')
        .period(10m)
        .every(1m)
    |eval(lambda: sigma("count"))
        .as('sigma')
        .keep('count', 'sigma')
    |alert()
        .crit(lambda: "sigma" > [[ .Sigma ]])
        .warn(lambda: ("sigma" <= [[ .Sigma ]] AND "sigma" >= 0.1) )
        .stateChangesOnly()
        [[if .Slack ]]
        .slack()
        .channel('[[ .Slack ]]')
        [[end]]
        .message('[[ .Fqdn ]]: {{ if eq .Level "CRITICAL" }}Excessive 5xxs {{ end }}{{ if eq .Level "OK" }}5xxs back to normal {{ end }}{{ if eq .Level "INFO" }}5xxs Returning to Normal {{ end }}{{ if eq .Level "WARNING" }}Elevated 5xxs {{ end }} Metric: {{ .Name }}  Sigma: {{ index .Fields "sigma" | printf "%0.2f" }} Count: {{ index .Fields "count" }}')
        .details('''
<h3>{{ .Message }}</h3>
<a href="https://membanks.octanner.io/dashboard/db/alamo-router-scanner?var-url=[[ .Fqdn ]]&from=now-1h&to=now&panelId=4&fullscreen">Link To Memory Banks</a>
''')
        [[if .Email]][[ range $email := .EmailArray ]]
        .email('[[ $email ]]')[[end]][[end]]
        [[if .Post]]
        .post('[[ .Post ]]')
        [[end]]    
 
`

// getTaskByName - Get a task from the database
func getTaskByName(app string, c *gin.Context) (*_5xxDBTask, error) {
	db, err := utils.GetDBFromContext(c)
	if err != nil {
		return nil, errors.New("Unable to access database")
	}

	task := _5xxDBTask{}

	err = db.Get(&task, "SELECT * FROM _5xx_tasks WHERE app=$1", app)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.New("Unable to access database")
	}
	return &task, nil
}

// Process5xxRequest - Create structs for creating and updating 5xx tasks
func Process5xxRequest(c *gin.Context) {
	var vars map[string]structs.Var
	vars = make(map[string]structs.Var)

	bodybytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	var dbrps []structs.DbrpSpec
	var dbrp structs.DbrpSpec
	var task _5xxTaskSpec
	err = json.Unmarshal(bodybytes, &task)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	task.ID = task.App + "-5xx"

	task.Type = "batch"
	vars = utils.AddVar("type", task.Type, "string", vars)
	var tolerancelist map[string]string
	tolerancelist = make(map[string]string)
	tolerancelist["low"] = "0.5"
	tolerancelist["medium"] = "1.0"
	tolerancelist["high"] = "1.5"
	task.Sigma = tolerancelist[task.Tolerance]
	dbrp.Db = "opentsdb"
	dbrp.Rp = "autogen"
	dbrps = append(dbrps, dbrp)
	task.Dbrps = dbrps
	task.Script = ""
	task.Status = "enabled"

	if !strings.HasPrefix(task.Slack, "#") && !strings.HasPrefix(task.Slack, "@") {
		task.Slack = "#" + task.Slack
	}

	task.EmailArray = strings.Split(task.Email, ",")

	t := template.Must(template.New("_5xxalerttemplate").Delims("[[", "]]").Parse(_5xxalerttemplate))
	var sb bytes.Buffer
	swr := bufio.NewWriter(&sb)
	err = t.Execute(swr, task)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	swr.Flush()
	task.Script = string(sb.Bytes())
	vars = utils.AddVar("id", task.ID, "string", vars)
	vars = utils.AddVar("app", task.App, "string", vars)
	vars = utils.AddVar("fqdn", task.Fqdn, "string", vars)
	vars = utils.AddVar("tolerance", task.Tolerance, "string", vars)
	vars = utils.AddVar("sigma", task.Sigma, "string", vars)
	vars = utils.AddVar("slack", task.Slack, "string", vars)
	vars = utils.AddVar("post", task.Post, "string", vars)
	vars = utils.AddVar("email", task.Email, "string", vars)

	task.Vars = vars

	bodybytes, err = json.Marshal(task)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	if c.Request.Method == "POST" {
		create5xxTask(task, c)
	} else if c.Request.Method == "PATCH" {
		// Check if task exists before trying to patch it
		_, err := getTaskByName(task.App, c)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(404, nil)
			} else {
				utils.ReportError(err, c, "")
			}
			return
		}
		delete5xxTask(task.App, c)
		create5xxTask(task, c)
	}
}

// create5xxTask - Create a new task in Kapacitor and save it to the database
func create5xxTask(task _5xxTaskSpec, c *gin.Context) {
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
		"INSERT INTO _5xx_tasks VALUES ($1, $2, $3, $4, $5)",
		task.App, task.Tolerance, task.Slack, task.Post, task.Email,
	)

	if err != nil {
		utils.ReportError(err, c, "Unable to save to database")
		return
	}

	c.String(201, "")
}

// Delete5xxTask - Verify that a task exists in the database before talking to Kapacitor
func Delete5xxTask(c *gin.Context) {
	app := c.Param("app")

	// Check if task exists before trying to delete it
	_, err := getTaskByName(app, c)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, nil)
		} else {
			utils.ReportError(err, c, "")
		}
		return
	}

	delete5xxTask(app, c)
	c.String(200, "")
}

func delete5xxTask(app string, c *gin.Context) {
	db, err := utils.GetDBFromContext(c)
	if err != nil {
		utils.ReportError(err, c, "Unable to access database")
		return
	}

	client := http.Client{}
	req, err := http.NewRequest("DELETE", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks/"+app+"-5xx", nil)
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

	_, err = db.Exec("DELETE FROM _5xx_tasks WHERE app=$1", app)
	if err != nil {
		utils.ReportError(err, c, "Unable to access database")
		return
	}
}

// Get5xxTask - Get the config of a specific 5xx task from the database
func Get5xxTask(c *gin.Context) {
	app := c.Param("app")

	task, err := getTaskByName(app, c)
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

// List5xxTasks - Get the config of all 5xx tasks from the database
func List5xxTasks(c *gin.Context) {
	db, err := utils.GetDBFromContext(c)
	if err != nil {
		utils.ReportError(err, c, "Unable to access database")
		return
	}

	tasks := []_5xxDBTask{}

	err = db.Select(&tasks, "SELECT * FROM _5xx_tasks ORDER BY app ASC")
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

// Get5xxTaskState - Get the current state of a 5xx task for an app from Kapacitor
func Get5xxTaskState(c *gin.Context) {
	var stateresp _5xxSimpleTaskState
	app := c.Param("app")

	_, err := getTaskByName(app, c)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, nil)
		} else {
			utils.ReportError(err, c, "")
		}
		return
	}

	client := http.Client{}
	req, err := http.NewRequest("GET", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1preview/alerts/topics?pattern=*"+app+"-5xx*", nil)
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
	var taskstate _5xxTaskState
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	err = json.Unmarshal(bodybytes, &taskstate)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	stateresp.App = app
	stateresp.State = taskstate.Topics[0].Level

	c.JSON(200, stateresp)
}
