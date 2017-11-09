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
	"text/template"
)


type ErrorResponse struct {
    Error string `json:"error"`
}

type DbrpSpec struct {
	Db string `json:"db"`
	Rp string `json:"rp"`
}

type TaskSpec struct {
	ID       string     `json:"id"`
	Type     string     `json:"type"`
	Dbrps    []DbrpSpec `json:"dbrps"`
	Status   string     `json:"status"`
	Script   string     `json:"script"`
	App      string     `json:"app"`
	Crit     string     `json:"crit"`
	Warn     string     `json:"warn"`
	Slack    string     `json:"slack"`
	Window   string     `json:"window"`
	Every    string     `json:"every"`
	Post     string     `json:"post"`
	Email    string     `json:"email"`
	Opsgenie string     `json:"opsgenie"`
	Dynotype string     `json:"dynotype"`
	Metric   string     `json:"metric"`
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
	router.POST("/task", createTask)
	router.Run()

}

func createTask(c *gin.Context) {

	bodybytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Println(err)
	}

	var dbrps []DbrpSpec
	var dbrp DbrpSpec
	var task TaskSpec
	err = json.Unmarshal(bodybytes, &task)
	if err != nil {
		fmt.Println(err)
                var er ErrorResponse
                er.Error="Server Error while reading response"
                c.JSON(500, er)
                return
	}

	task.ID = task.App + "-memory"
	if task.Dynotype != "*" {
		task.ID = task.App + "-" + task.Metric + "-" + task.Dynotype
	}
	if task.Dynotype == "*" {
		task.ID = task.App + "-" + task.Metric + "-all"
	}
	task.Type = "batch"
	dbrp.Db = "opentsdb"
	dbrp.Rp = "autogen"
	dbrps = append(dbrps, dbrp)
	if task.Dynotype == "*" {
		task.Dynotype = " =~ /.*/ "
	} else if task.Dynotype == "web" {
		task.Dynotype = " !~ /--/ "
	} else {
		task.Dynotype = " =~ /" + task.Dynotype + "/ "
	}

	task.Dbrps = dbrps
	task.Script = ""
	task.Status = "enabled"
	t := template.Must(template.New("memoryalerttemplate").Delims("[[", "]]").Parse(memoryalerttemplate))
	var sb bytes.Buffer
	swr := bufio.NewWriter(&sb)
	err = t.Execute(swr, task)
	if err != nil {
		fmt.Println(err)
                var er ErrorResponse
                er.Error="Server Error while reading response"
                c.JSON(500, er)
                return
	}
	swr.Flush()
	fmt.Println(string(sb.Bytes()))
	task.Script = string(sb.Bytes())

	client := http.Client{}

	p, err := json.Marshal(task)
	if err != nil {
		fmt.Println(err)
                var er ErrorResponse
                er.Error="Server Error while reading response"
                c.JSON(500, er)
                return
	}

	//	req, err := http.NewRequest("POST", "http://10.84.24.142:9092/kapacitor/v1/tasks", bytes.NewBuffer(p))
	fmt.Println(string(p))
	req, err := http.NewRequest("POST", os.Getenv("KAPACITOR_URL")+"/kapacitor/v1/tasks", bytes.NewBuffer(p))
	if err != nil {
		fmt.Println(err)
                var er ErrorResponse
                er.Error="Server Error while reading response"
                c.JSON(500, er)
                return
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
                var er ErrorResponse
                er.Error="Server Error while reading response"
                c.JSON(500, er)
                return
	}
	defer resp.Body.Close()
	bodybytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
                var er ErrorResponse
                er.Error="Server Error while reading response"
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
