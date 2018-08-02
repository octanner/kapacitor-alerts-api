package release

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

const releasealerttemplate = `
batch
    |query('''
select text,title,app from "opentsdb"."autogen"."events" where "app"='[[ .App ]]' and "title"= 'released'
    ''')
        .period(60s)
        .every(61s)
    |alert()
        .warn(lambda: 1 > 0)
        [[if .Slack ]]
        .slack()
        .channel('[[ .Slack ]]')
        [[end]]
        .message('{{ index .Fields "app" }} released.  New image is {{ index .Fields "text" }}')
        .details('''
<h3>{{ .Message }}</h3>
{{ index .Fields "app" }} released.  New image is {{ index .Fields "text" }}
''')
        [[if .Email]][[ range $email := .EmailArray ]] 
        .email('[[ $email ]]')[[end]][[end]]
        [[if .Post]]
        .post('[[ .Post ]]')
        [[end]]

`


func ProcessReleaseRequest(c *gin.Context) {
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
	var task ReleaseTaskSpec
	err = json.Unmarshal(bodybytes, &task)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}

	task.ID = task.App + "-release"
	task.Type = "batch"
	vars = addvar("type", task.Type, "string", vars)

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

	t := template.Must(template.New("releasealerttemplate").Delims("[[", "]]").Parse(releasealerttemplate))
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
fmt.Println(task.ID)
	vars = addvar("app", task.App, "string", vars)
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
		deleteReleaseTask(task, c)
		c.String(200, "")
	}
	if c.Request.Method == "POST" {
		createReleaseTask(task, c)
	}
	if c.Request.Method == "PATCH" {
		deleteReleaseTask(task, c)
		createReleaseTask(task, c)
	}

}

func deleteReleaseTask(task ReleaseTaskSpec, c *gin.Context) {
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

func createReleaseTask(task ReleaseTaskSpec, c *gin.Context) {

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

func DeleteReleaseTask(c *gin.Context) {

	var task ReleaseTaskSpec
	client := http.Client{}
	req, err := http.NewRequest("GET", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks?pattern="+c.Param("app")+"-release", nil)
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
    
	var tasklist ReleaseTaskList
	err = json.Unmarshal(bodybytes, &tasklist)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	for _, element := range tasklist.Tasks {
		if strings.HasPrefix(element.ID, c.Param("app")) && strings.HasSuffix(element.ID, "-release") {
			simpletask, err := convertToSimpleReleaseTask(element.ID, element.Vars)
			if err != nil {
				fmt.Println(err)
			}
			task = simpletask

		}
	}
	deleteReleaseTask(task, c)

}

func GetReleaseTask(c *gin.Context) {

	var tasktoreturn ReleaseTaskSpec
	client := http.Client{}
	req, err := http.NewRequest("GET", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks?pattern=*-release", nil)
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
	var tasklist ReleaseTaskList
	err = json.Unmarshal(bodybytes, &tasklist)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
	for _, element := range tasklist.Tasks {
		if strings.HasPrefix(element.ID, c.Param("app")) && strings.HasSuffix(element.ID, "-"+c.Param("id")) {
			simpletask, err := convertToSimpleReleaseTask(element.ID, element.Vars)
			if err != nil {
				fmt.Println(err)
			}
			tasktoreturn = simpletask

		}
	}
	c.JSON(200, tasktoreturn)

}

func GetReleaseTaskForApp(c *gin.Context) {
	//var tasks []ReleaseTaskSpec
	client := http.Client{}
	req, err := http.NewRequest("GET", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks?pattern="+c.Param("app")+"-release", nil)
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
	var tasklist ReleaseTaskList
	err = json.Unmarshal(bodybytes, &tasklist)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}
 var simpletask ReleaseTaskSpec
	for _, element := range tasklist.Tasks {
	//	if strings.HasPrefix(element.ID, c.Param("app")) {
			simpletask, err = convertToSimpleReleaseTask(element.ID, element.Vars)
			if err != nil {
				fmt.Println(err)
			}
	///		tasks = append(tasks, simpletask)
	//	}
	}
	c.JSON(200, simpletask)
}

func convertToSimpleReleaseTask(id string, vars ReleaseVars) (t ReleaseTaskSpec, e error) {
	var tasktoreturn ReleaseTaskSpec
	tasktoreturn.ID = id
	tasktoreturn.App = vars.App.Value
        tasktoreturn.Type = vars.Type.Value
	tasktoreturn.Slack = vars.Slack.Value
	tasktoreturn.Email = vars.Email.Value
	tasktoreturn.Post = vars.Post.Value

	return tasktoreturn, nil
}

func ListReleaseTasks(c *gin.Context) {
	var tasks []ReleaseTaskSpec
	client := http.Client{}
	req, err := http.NewRequest("GET", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks?pattern=*-sample.release_total-*", nil)
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
	var tasklist ReleaseTaskList
	err = json.Unmarshal(bodybytes, &tasklist)
	if err != nil {
		fmt.Println(err)
		var er structs.ErrorResponse
		er.Error = "Server Error while reading response"
		c.JSON(500, er)
		return
	}

	for _, element := range tasklist.Tasks {
		if strings.Contains(element.ID, "sample.release_total") {
			simpletask, err := convertToSimpleReleaseTask(element.ID, element.Vars)
			if err != nil {
				fmt.Println(err)
			}
			if simpletask.App != "" {
				tasks = append(tasks, simpletask)
			}
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

		var1.Type = vtype
		var1.Description = name
		flistin[name] = var1

	}
	return flistin

}
