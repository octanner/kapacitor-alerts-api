package memory

import (
	structs "kapacitor-alerts-api/structs"

	"gopkg.in/guregu/null.v3/zero"
)

type MemoryVars struct {
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

type MemoryTaskList struct {
	Tasks []struct {
		ID   string     `json:"id"`
		Vars MemoryVars `json:"vars"`
	} `json:"tasks"`
}

type MemoryTaskSpec struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Dbrps      []structs.DbrpSpec `json:"dbrps"`
	Status     string             `json:"status"`
	Script     string             `json:"script"`
	App        string             `json:"app"`
	Crit       string             `json:"crit"`
	Warn       string             `json:"warn"`
	Slack      string             `json:"slack"`
	Window     string             `json:"window"`
	Every      string             `json:"every"`
	Post       string             `json:"post"`
	Email      string             `json:"email"`
	EmailArray []string           `json:"emailarray"`
	//    Opsgenie string         `json:"opsgenie"`
	Dynotype string                 `json:"dynotype"`
	Metric   string                 `json:"metric"`
	Vars     map[string]structs.Var `json:"vars"`
}

// MemoryDBTask - Used for retrieval of task information from the database
type MemoryDBTask struct {
	ID       string      `json:"id"`
	App      string      `json:"app"`
	Dynotype string      `json:"dynotype"`
	Crit     string      `json:"crit"`
	Warn     string      `json:"warn"`
	Wind     string      `json:"window"`
	Every    string      `json:"every"`
	Slack    zero.String `json:"slack"`
	Post     zero.String `json:"post"`
	Email    zero.String `json:"email"`
}
