package _5xx

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	structs "kapacitor-alerts-api/structs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
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
        [[if .Slack ]]
        .slack()
        .channel('[[ .Slack ]]')
        [[end]]
        .message('[[ .Fqdn ]]: {{ if eq .Level "CRITICAL" }}Excessive 5xxs {{ end }}{{ if eq .Level "OK" }}5xxs back to normal {{ end }}{{ if eq .Level "INFO" }}5xxs Returning to Normal {{ end }}{{ if eq .Level "WARNING" }}Elevated 5xxs {{ end }} Metric: {{ .Name }}  Sigma: {{ index .Fields "sigma" | printf "%0.2f" }} Count: {{ index .Fields "count" }}')
        .details('''
<h3>{{ .Message }}</h3>
<a href="https://membanks.octanner.io/dashboard/db/alamo-router-scanner?var-url=[[ .Fqdn ]]&from=now-1h&to=now&panelId=4&fullscreen">Link To Memory Banks</a>
''')
        [[if .Email]]
        .email('[[ .Email ]]')
        [[end]]
        [[if .Post]]
        .post('[[ .Post ]]')
        [[end]]    
 
`

//https://membanks.octanner.io/dashboard/db/alamo-router-scanner?var-url=obertbin-voltron.alamoapp.octanner.io&from=now-1h&to=now&panelId=4&fullscreen

func Process5xxRequest(c *gin.Context) {
	var vars map[string]structs.Var
	vars = make(map[string]structs.Var)

	bodybytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	var dbrps []structs.DbrpSpec
	var dbrp structs.DbrpSpec
	var task _5xxTaskSpec
	err = json.Unmarshal(bodybytes, &task)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}

	task.ID = task.App + "-5xx"

	task.Type = "batch"
	vars = addvar("type", task.Type, "string", vars)
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

	t := template.Must(template.New("_5xxalerttemplate").Delims("[[", "]]").Parse(_5xxalerttemplate))
	var sb bytes.Buffer
	swr := bufio.NewWriter(&sb)
	err = t.Execute(swr, task)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	swr.Flush()
	task.Script = string(sb.Bytes())
	vars = addvar("id", task.ID, "string", vars)
	vars = addvar("app", task.App, "string", vars)
	vars = addvar("fqdn", task.Fqdn, "string", vars)
	vars = addvar("tolerance", task.Tolerance, "string", vars)
	vars = addvar("sigma", task.Sigma, "string", vars)
	vars = addvar("slack", task.Slack, "string", vars)
	vars = addvar("post", task.Post, "string", vars)
	vars = addvar("email", task.Email, "string", vars)

	task.Vars = vars

	bodybytes, err = json.Marshal(task)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	if c.Request.Method == "DELETE" {
		delete5xxTask(task, c)
		c.String(200, "")
	}
	if c.Request.Method == "POST" {
		create5xxTask(task, c)
	}
	if c.Request.Method == "PATCH" {
		delete5xxTask(task, c)
		create5xxTask(task, c)
	}

}
func create5xxTask(task _5xxTaskSpec, c *gin.Context) {

	client := http.Client{}

	p, err := json.Marshal(task)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	req, err := http.NewRequest("POST", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks", bytes.NewBuffer(p))
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	defer resp.Body.Close()
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	if resp.StatusCode != 200 {
		fmt.Println(string(bodybytes))
		var er structs.ErrorResponse
		err = json.Unmarshal(bodybytes, &er)
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(500, er)
		return

	}
	c.String(201, "")

}

func Delete5xxTask(c *gin.Context) {

	var task _5xxTaskSpec
	client := http.Client{}
	req, err := http.NewRequest("GET", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks", nil)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	defer resp.Body.Close()
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	var tasklist _5xxTaskList
	err = json.Unmarshal(bodybytes, &tasklist)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	for _, element := range tasklist.Tasks {
		if strings.HasPrefix(element.ID, c.Param("app")) && strings.HasSuffix(element.ID, "-5xx") {
			simpletask, err := convertToSimple5xxTask(element.ID, element.Vars)
			if err != nil {
				fmt.Println("error while converting")
				fmt.Println(err)
			}
			task = simpletask

		}
	}
	delete5xxTask(task, c)

}

func delete5xxTask(task _5xxTaskSpec, c *gin.Context) {
	client := http.Client{}
	req, err := http.NewRequest("DELETE", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks/"+task.ID, nil)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	defer resp.Body.Close()
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	if resp.StatusCode != 204 {
		fmt.Println(string(bodybytes))
		var er structs.ErrorResponse
		err = json.Unmarshal(bodybytes, &er)
		if err != nil {
			fmt.Println(err)
		}
		c.JSON(500, er)
		return

	}
}

func Get5xxTask(c *gin.Context) {

	var tasktoreturn _5xxTaskSpec
	client := http.Client{}
	req, err := http.NewRequest("GET", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks", nil)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	defer resp.Body.Close()
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	var tasklist _5xxTaskList
	err = json.Unmarshal(bodybytes, &tasklist)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	for _, element := range tasklist.Tasks {
		if strings.HasPrefix(element.ID, c.Param("app")) && strings.HasSuffix(element.ID, "-5xx") {
			simpletask, err := convertToSimple5xxTask(element.ID, element.Vars)
			if err != nil {
				fmt.Println(err)
			}
			tasktoreturn = simpletask

		}
	}
	c.JSON(200, tasktoreturn)

}
func List5xxTasks(c *gin.Context) {
	var tasks []_5xxTaskSpec
	client := http.Client{}
	req, err := http.NewRequest("GET", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks", nil)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	defer resp.Body.Close()
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	var tasklist _5xxTaskList
	err = json.Unmarshal(bodybytes, &tasklist)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}

	for _, element := range tasklist.Tasks {
		simpletask, err := convertToSimple5xxTask(element.ID, element.Vars)
		if err != nil {
			fmt.Println(err)
		}
		if strings.HasSuffix(simpletask.ID, "-5xx") {
			tasks = append(tasks, simpletask)
		}
	}
	c.JSON(200, tasks)

}

func addvar(name string, value string, vtype string, flistin map[string]structs.Var) (flistout map[string]structs.Var) {
	if value != "" {
		var var1 structs.Var
		if vtype == "string" {
			var1.Value = value
		}
		if vtype == "int" {
			intvalue, _ := strconv.Atoi(value)
			var1.Value = intvalue
		}
		if vtype == "float" {
			floatvalue, _ := strconv.ParseFloat(value, 64)
			var1.Value = floatvalue
		}

		var1.Type = vtype
		var1.Description = name
		flistin[name] = var1

	}
	return flistin

}

func convertToSimple5xxTask(id string, vars _5xxVars) (t _5xxTaskSpec, e error) {
	var tasktoreturn _5xxTaskSpec
	tasktoreturn.ID = id
	tasktoreturn.App = vars.App.Value
	tasktoreturn.Fqdn = vars.Fqdn.Value
	tasktoreturn.Tolerance = vars.Tolerance.Value
	tasktoreturn.Slack = vars.Slack.Value
	tasktoreturn.Email = vars.Email.Value
	tasktoreturn.Post = vars.Post.Value
	tasktoreturn.Sigma = vars.Sigma.Value

	return tasktoreturn, nil
}

func Get5xxTaskState(c *gin.Context) {
	var stateresp _5xxSimpleTaskState
	state, err := get5xxTaskState(c.Param("app"))
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	stateresp.App = c.Param("app")
	stateresp.State = state
	c.JSON(200, stateresp)
	return
}

func get5xxTaskState(app string) (s string, e error) {

	var state string

	client := http.Client{}
	req, err := http.NewRequest("GET", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1preview/alerts/topics?pattern=*"+app+"-5xx*", nil)
	if err != nil {
		return state, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return state, err
	}
	defer resp.Body.Close()
	var taskstate _5xxTaskState
	bodybytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return state, err
	}
	err = json.Unmarshal(bodybytes, &taskstate)
	if err != nil {
		return state, err
	}

	state = taskstate.Topics[0].Level
	return state, nil

}
