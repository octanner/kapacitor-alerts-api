package _5xx

import (
	structs "kapacitor-alerts-api/structs"
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
	ID        string                 `json:id"`
	Type      string                 `json:"type"`
	Dbrps     []structs.DbrpSpec     `json:"dbrps"`
	Status    string                 `json:"status"`
	Script    string                 `json:"script"`
	App       string                 `json:"app"`
	Fqdn      string                 `json:"fqdn"`
	Tolerance string                 `json:"tolerance"`
        Sigma     string                 `json:"sigma"`
	Slack     string                 `json:"slack"`
	Post      string                 `json:"post"`
	Email     string                 `json:"email"`
	Vars      map[string]structs.Var `json:"vars"`
}

type _5xxTaskList struct {
	Tasks []struct {
		ID   string   `json:"id"`
		Vars _5xxVars `json:"vars"`
	} `json:"tasks"`
}
