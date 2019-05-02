package _5xx

import (
	structs "kapacitor-alerts-api/structs"

	"gopkg.in/guregu/null.v3/zero"
)

type _5xxVars struct {
	App struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"app"`
	ID struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"id"`
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
	Fqdn struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"fqdn"`
	Tolerance struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"tolerance"`
	Sigma struct {
		Type        string `json:"type"`
		Value       string `json:"value"`
		Description string `json:"description"`
	} `json:"sigma"`
}

type _5xxTaskSpec struct {
	ID         string                 `json:id"`
	Type       string                 `json:"type"`
	Dbrps      []structs.DbrpSpec     `json:"dbrps"`
	Status     string                 `json:"status"`
	Script     string                 `json:"script"`
	App        string                 `json:"app"`
	Fqdn       string                 `json:"fqdn"`
	Tolerance  string                 `json:"tolerance"`
	Sigma      string                 `json:"sigma"`
	Slack      string                 `json:"slack"`
	Post       string                 `json:"post"`
	Email      string                 `json:"email"`
	EmailArray []string               `json:"emailarray"`
	Vars       map[string]structs.Var `json:"vars"`
}

type _5xxTaskList struct {
	Tasks []struct {
		ID   string   `json:"id"`
		Vars _5xxVars `json:"vars"`
	} `json:"tasks"`
}

type _5xxTaskState struct {
	Link struct {
		Rel  string `json:"rel"`
		Href string `json:"href"`
	} `json:"link"`
	Topics []struct {
		Link struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"link"`
		ID         string `json:"id"`
		Level      string `json:"level"`
		Collected  int    `json:"collected"`
		EventsLink struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"events-link"`
		HandlersLink struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"handlers-link"`
	} `json:"topics"`
}

type _5xxSimpleTaskState struct {
	App   string `json:"app"`
	State string `json:"state"`
}

type _5xxDBTask struct {
	App       string      `json:"app"`
	Tolerance string      `json:"tolerance"`
	Slack     zero.String `json:"slack"`
	Post      zero.String `json:"post"`
	Email     zero.String `json:"email"`
}
