package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
)

type Vars struct {
	App struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"app"`
	Crit struct {
		Type        string `json:"type"`
		Value       int    `json:"value"`
		Description string `json:"description"`
	} `json:"crit"`
	Dynotype struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"dynotype"`
	Dynotyperequest struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"dynotyperequest"`
	Every struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"every"`
	ID struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"id"`
	Metric struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"metric"`
	Slack struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"slack"`
	Type struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"type"`
	Warn struct {
		Type        string `json:"type"`
		Value       int    `json:"value"`
		Description string `json:"description"`
	} `json:"warn"`
	Window struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"window"`
	Email struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"email"`
	Post struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"post"`
}

type TaskList struct {
	Tasks []struct {
		ID   string `json:"id"`
		Vars Vars   `json:"vars"`
	} `json:"tasks"`
}

type Var struct {
	Value       interface{} `json:"value"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type DbrpSpec struct {
	Db string `json:"db"`
	Rp string `json:"rp"`
}

type TaskSpec struct {
	ID     string     `json:"id"`
	Type   string     `json:"type"`
	Dbrps  []DbrpSpec `json:"dbrps"`
	Status string     `json:"status"`
	Script string     `json:"script"`
	App    string     `json:"app"`
	Crit   string     `json:"crit"`
	Warn   string     `json:"warn"`
	Slack  string     `json:"slack"`
	Window string     `json:"window"`
	Every  string     `json:"every"`
	Post   string     `json:"post"`
	Email  string     `json:"email"`
	//	Opsgenie string         `json:"opsgenie"`
	Dynotype string         `json:"dynotype"`
	Metric   string         `json:"metric"`
	Vars     map[string]Var `json:"vars"`
}

const memoryalerttemplate = `
batch
    |query('''
select mean(value)/1024/1024 as value from "opentsdb"."autogen"."[[ .Metric ]]" where "app"='[[ .App ]]' and "dyno" [[ .Dynotype ]]
    ''')
        .period([[ .Window ]])
        .every([[ .Every ]])
        .groupBy('app','dyno')
    |eval(lambda: ceil("value")).as('rvalue').keep('value','rvalue')
    |alert()
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
        .email('[[ .Email ]]')
        [[end]]
        [[if .Post]]
        .post('[[ .Post ]]')
        [[end]]

`

//       .opsGenie().recipients(['[[ .Opsgenie ]]'])
func main() {

	router := gin.Default()
	router.POST("/task/memory", processInstanceMemoryRequest)
	router.PATCH("/task/memory", processInstanceMemoryRequest)
	router.DELETE("/task/memory/:app/:id", deleteTask)
	router.GET("/tasks/memory/:app", getTasksForApp)
	router.GET("/tasks/memory/:app/:id", getTask)
	router.Run()

}

func processInstanceMemoryRequest(c *gin.Context) {
	var vars map[string]Var
	vars = make(map[string]Var)

	bodybytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	var dbrps []DbrpSpec
	var dbrp DbrpSpec
	var task TaskSpec
	err = json.Unmarshal(bodybytes, &task)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}

	task.ID = task.App + "-memory"

	task.Metric = "sample.memory_total"
	vars = addvar("metric", task.Metric, "string", vars)
	vars = addvar("dynotyperequest", task.Dynotype, "string", vars)
	if task.Dynotype != "all" {
		task.ID = task.App + "-" + task.Metric + "-" + task.Dynotype
		vars = addvar("id", task.ID, "string", vars)

	}
	if task.Dynotype == "all" {
		task.ID = task.App + "-" + task.Metric + "-all"
		vars = addvar("id", task.ID, "string", vars)
	}

	task.Type = "batch"
	vars = addvar("type", task.Type, "string", vars)

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
	vars = addvar("dynotype", task.Dynotype, "string", vars)

	task.Dbrps = dbrps
	task.Script = ""
	task.Status = "enabled"
        if !strings.HasPrefix(task.Slack, "#") && !strings.HasPrefix(task.Slack,"@"){
          task.Slack = "#"+task.Slack
        }

	t := template.Must(template.New("memoryalerttemplate").Delims("[[", "]]").Parse(memoryalerttemplate))
	var sb bytes.Buffer
	swr := bufio.NewWriter(&sb)
	err = t.Execute(swr, task)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	swr.Flush()
	task.Script = string(sb.Bytes())
	vars = addvar("app", task.App, "string", vars)
	vars = addvar("crit", task.Crit, "int", vars)
	vars = addvar("warn", task.Warn, "int", vars)
	vars = addvar("slack", task.Slack, "string", vars)
	vars = addvar("window", task.Window, "string", vars)
	vars = addvar("every", task.Every, "string", vars)
	vars = addvar("post", task.Post, "string", vars)
	vars = addvar("email", task.Email, "string", vars)

	task.Vars = vars

	bodybytes, err = json.Marshal(task)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	if c.Request.Method == "DELETE" {
		deleteInstanceMemoryTask(task, c)
		c.String(200, "")
	}
	if c.Request.Method == "POST" {
		createInstanceMemoryTask(task, c)
	}
	if c.Request.Method == "PATCH" {
		deleteInstanceMemoryTask(task, c)
		createInstanceMemoryTask(task, c)
	}

}

func deleteInstanceMemoryTask(task TaskSpec, c *gin.Context) {
	client := http.Client{}
	req, err := http.NewRequest("DELETE", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks/"+task.ID, nil)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	defer resp.Body.Close()
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	if resp.StatusCode != 204 {
		fmt.Println(string(bodybytes))
		var er ErrorResponse
		err = json.Unmarshal(bodybytes, &er)
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(500, er)
		return

	}
}

func createInstanceMemoryTask(task TaskSpec, c *gin.Context) {

	client := http.Client{}

	p, err := json.Marshal(task)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	req, err := http.NewRequest("POST", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks", bytes.NewBuffer(p))
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	defer resp.Body.Close()
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	if resp.StatusCode != 200 {
		fmt.Println(string(bodybytes))
		var er ErrorResponse
		err = json.Unmarshal(bodybytes, &er)
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(500, er)
		return

	}
	c.String(201, "")

}

func addvar(name string, value string, vtype string, flistin map[string]Var) (flistout map[string]Var) {
	if value != "" {
		var var1 Var
		if vtype == "string" {
			var1.Value = value
		}
		if vtype == "int" {
			intvalue, _ := strconv.Atoi(value)
			var1.Value = intvalue
		}

		var1.Type = vtype
		var1.Description = name
		flistin[name] = var1

	}
	return flistin

}

func deleteTask(c *gin.Context) {

	var task TaskSpec
	client := http.Client{}
	req, err := http.NewRequest("GET", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks", nil)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	defer resp.Body.Close()
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	var tasklist TaskList
	err = json.Unmarshal(bodybytes, &tasklist)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	for _, element := range tasklist.Tasks {
		if strings.HasPrefix(element.ID, c.Param("app")) && strings.HasSuffix(element.ID, "-"+c.Param("id")) {
			simpletask, err := convertToSimpleTask(element.ID, element.Vars)
			if err != nil {
				fmt.Println(err)
			}
			task = simpletask

		}
	}
	deleteInstanceMemoryTask(task, c)

}
func getTask(c *gin.Context) {

	var tasktoreturn TaskSpec
	client := http.Client{}
	req, err := http.NewRequest("GET", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks", nil)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	defer resp.Body.Close()
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	var tasklist TaskList
	err = json.Unmarshal(bodybytes, &tasklist)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	for _, element := range tasklist.Tasks {
		if strings.HasPrefix(element.ID, c.Param("app")) && strings.HasSuffix(element.ID, "-"+c.Param("id")) {
			simpletask, err := convertToSimpleTask(element.ID, element.Vars)
			if err != nil {
				fmt.Println(err)
			}
			tasktoreturn = simpletask

		}
	}
	c.JSON(200, tasktoreturn)

}

func getTasksForApp(c *gin.Context) {
	var tasks []TaskSpec
	client := http.Client{}
	req, err := http.NewRequest("GET", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks", nil)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	defer resp.Body.Close()
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	var tasklist TaskList
	err = json.Unmarshal(bodybytes, &tasklist)
	if err != nil {
		fmt.Println(err)
		var er ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}

	for _, element := range tasklist.Tasks {
		if strings.HasPrefix(element.ID, c.Param("app")) {
			simpletask, err := convertToSimpleTask(element.ID, element.Vars)
			if err != nil {
				fmt.Println(err)
			}
			tasks = append(tasks, simpletask)

		}
	}
	c.JSON(200, tasks)
}

func convertToSimpleTask(id string, vars Vars) (t TaskSpec, e error) {
	var tasktoreturn TaskSpec
	tasktoreturn.ID = id
	tasktoreturn.App = vars.App.Value

	tasktoreturn.Dynotype = vars.Dynotyperequest.Value

	tasktoreturn.Window = vars.Window.Value
	tasktoreturn.Every = vars.Every.Value
	tasktoreturn.Crit = strconv.Itoa(vars.Crit.Value)
	tasktoreturn.Warn = strconv.Itoa(vars.Warn.Value)
	tasktoreturn.Slack = vars.Slack.Value
	tasktoreturn.Email = vars.Email.Value
	tasktoreturn.Post = vars.Post.Value

	return tasktoreturn, nil
}
