package structs

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
