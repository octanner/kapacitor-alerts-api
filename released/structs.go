package released

import (
	structs "kapacitor-alerts-api/structs"

	"gopkg.in/guregu/null.v3/zero"
)

type ReleaseVars struct {
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
}

type ReleaseTaskList struct {
	Tasks []struct {
		ID   string      `json:"id"`
		Vars ReleaseVars `json:"vars"`
	} `json:"tasks"`
}

type ReleaseTaskSpec struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Dbrps      []structs.DbrpSpec `json:"dbrps"`
	Status     string             `json:"status"`
	Script     string             `json:"script"`
	App        string             `json:"app"`
	Slack      string             `json:"slack"`
	Post       string             `json:"post"`
	Email      string             `json:"email"`
	EmailArray []string           `json:"emailarray"`
	//    Opsgenie string         `json:"opsgenie"`
	Vars map[string]structs.Var `json:"vars"`
}

// ReleasedDBTask - Used for retrieval of task information from the database
type ReleasedDBTask struct {
	App   string      `json:"app"`
	Slack zero.String `json:"slack"`
	Post  zero.String `json:"post"`
	Email zero.String `json:"email"`
}
