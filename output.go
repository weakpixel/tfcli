package tfcli

import (
	"encoding/json"
	"fmt"
)

type tfOutVarible struct {
	Sensitive bool        `mapstructure:"sensitive"`
	Type      string      `mapstructure:"type"`
	Value     interface{} `mapstructure:"value"`
}

func readOutVars(bytes []byte) (map[string]string, error) {
	resMap := make(map[string]string)

	if bytes == nil {
		return resMap, nil
	}

	var outData map[string]tfOutVarible
	err := json.Unmarshal(bytes, &outData)
	if err != nil {
		return resMap, fmt.Errorf("unable to decode terraform output. Original error: %s", err)
	}

	for varName, variable := range outData {
		if variable.Type == "string" {
			resMap[varName] = variable.Value.(string)
		} else {
			varValueBytes, err := json.Marshal(variable.Value)
			if err != nil {
				return resMap, fmt.Errorf("unable to marshal variable value for variable %s. Original error: %s", varName, err)
			}
			resMap[varName] = string(varValueBytes)
		}
	}

	return resMap, nil
}
