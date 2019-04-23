package utils

import (
	"kapacitor-alerts-api/structs"
	"strconv"
)

// AddVar - Add a variable to a map[string]structs.Var -
//				 	$name: { Type: $type, Description: $name, Value: $value}
func AddVar(name string, value string, vtype string, flistin map[string]structs.Var) (flistout map[string]structs.Var) {
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
