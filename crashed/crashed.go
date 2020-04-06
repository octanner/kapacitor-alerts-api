package crashed

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

const crashalerttemplate = `
	batch
		|	query('''
				select text,title,app,tags from "opentsdb"."retention_policy"."events" where "app"='[[ .App ]]' and "title"= 'crashed' and text ='App crashed' and tags =~ /[[ .Space ]],[[ .Shortapp ]],/
			''')
				.period(60s)
				.every(61s)
		|	alert()
				.warn(lambda: 1 > 0)
				[[if .Slack ]]
					.slack()
					.channel('[[ .Slack ]]')
				[[end]]
				.message('{{ index .Fields "app" }} crashed. Info: {{ index .Fields "tags" }}')
				.details('''
					<h3>{{ .Message }}</h3>
					{{ index .Fields "app" }} crashed. Info: {{ index .Fields "tags" }} 
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

// getTaskByName - Get a task from the database
func getTaskByName(app string, c *gin.Context) (*CrashedDBTask, error) {
	db, err := utils.GetDBFromContext(c)
	if err != nil {
		return nil, errors.New("Unable to access database")
	}

	task := CrashedDBTask{}

	err = db.Get(&task, "SELECT * FROM crashed_tasks WHERE app=$1", app)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.New("Unable to access database")
	}
	return &task, nil
}

// ProcessCrashedRequest - POST | PATCH /task/crashed
func ProcessCrashedRequest(c *gin.Context) {
	var vars map[string]structs.Var
	vars = make(map[string]structs.Var)

	bodybytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	var dbrps []structs.DbrpSpec
	var dbrp structs.DbrpSpec
	var task CrashedTaskSpec
	err = json.Unmarshal(bodybytes, &task)
	if err != nil {
		utils.ReportError(err, c, "Server Error while reading response")
		return
	}

	task.ID = task.App + "-crash"
	task.Type = "batch"
	vars = utils.AddVar("type", task.Type, "string", vars)

	dbrp.Db = "opentsdb"
	dbrp.Rp = "retention_policy"
	dbrps = append(dbrps, dbrp)
	task.Dbrps = dbrps
	task.Script = ""
	task.Status = "enabled"
	task.Shortapp, task.Dynotype, task.Space = parseforparts(task.App)
	if !strings.HasPrefix(task.Slack, "#") && !strings.HasPrefix(task.Slack, "@") {
		task.Slack = "#" + task.Slack
	}

	task.EmailArray = strings.Split(task.Email, ",")

	t := template.Must(template.New("crashalerttemplate").Delims("[[", "]]").Parse(crashalerttemplate))
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
		err = createCrashedTask(task, c)
		if err != nil {
			utils.ReportError(err, c, "")
			return
		}
	}

	if c.Request.Method == "PATCH" {
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

		err = deleteCrashedTask(task.App, c)
		if err != nil {
			utils.ReportError(err, c, "")
			return
		}

		err = createCrashedTask(task, c)
		if err != nil {
			utils.ReportError(err, c, "")
			return
		}
	}

	c.String(201, "")
}

func deleteCrashedTask(app string, c *gin.Context) error {
	db, err := utils.GetDBFromContext(c)
	if err != nil {
		return errors.New("Unable to access database")
	}

	client := http.Client{}

	req, err := http.NewRequest("DELETE", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks/"+app+"-crash", nil)
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

	_, err = db.Exec("DELETE FROM crashed_tasks WHERE app=$1", app)
	if err != nil {
		return errors.New("Unable to access database")
	}

	return nil
}

// createCrashedTask - Create a task in Kapacitor and save the config to the database
func createCrashedTask(task CrashedTaskSpec, c *gin.Context) error {
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
		"INSERT INTO crashed_tasks VALUES ($1, $2, $3, $4)",
		task.App, task.Slack, task.Post, task.Email,
	)

	if err != nil {
		return errors.New("Unable to save to database")
	}

	return nil
}

// DeleteCrashedTask - DELETE /task/crashed/:app
func DeleteCrashedTask(c *gin.Context) {
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

	// Delete the task from Kapacitor and remove the config from the database
	err = deleteCrashedTask(app, c)
	if err != nil {
		utils.ReportError(err, c, "")
	} else {
		c.String(200, "")
	}
}

// GetCrashedTask - GET /task/crashed/:app
func GetCrashedTask(c *gin.Context) {
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

// ListCrashedTasks - GET /tasks/crashed
func ListCrashedTasks(c *gin.Context) {
	db, err := utils.GetDBFromContext(c)
	if err != nil {
		utils.ReportError(err, c, "Unable to access database")
		return
	}

	tasks := []CrashedDBTask{}

	err = db.Select(&tasks, "SELECT * FROM crashed_tasks ORDER BY app ASC")
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

func parseforparts(full string) (a string, d string, s string) {
	if strings.Contains(full, "--") {
		a = strings.Split(full, "--")[0]
		d = strings.Split(strings.Split(full, "--")[1], "-")[0]
		s = strings.Join(strings.Split(strings.Split(full, "--")[1], "-")[1:], "-")
	} else {
		a = strings.Split(full, "-")[0]
		d = "web"
		s = strings.Join(strings.Split(full, "-")[1:], "-")
	}
	return a, d, s
}
