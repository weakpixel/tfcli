package tfcli

import (
	"testing"
)

func TestReadOutVars(t *testing.T) {
	tables := []struct {
		output     []byte
		outVars    map[string]string
		shouldFail bool
	}{
		{
			output: []byte(`{                                         
				"resource_group": {
					"sensitive": false,
					"type": "string",
					"value": "my-rg"
				}, 
				"kubeconfig": {
					"sensitive": false,
					"type": "string",
					"value": "apiVersion: v1\nclusters:\n- cluster:\n    certificate-authority-data: LS0keykey==\n    server: https://id27182-b8s-default-76f11749.hcp.uksouth.azmk8s.io:443\n  name: id27182-b8s-default\ncontexts:\n- context:\n    cluster: id27182-b8s-default\n    user: clusterUser_id27182-b8s-default_id27182-b8s-default\n  name: id27182-b8s-default\ncurrent-context: id27182-b8s-default\nkind: Config\npreferences: {}\nusers:\n- name: clusterUser_id27182-b8s-default_id27182-b8s-default\n  user:\n    client-certificate-data: keykey=\n    client-key-data: keykey\n    token: mytoken"
				}
			}
			`),
			outVars: map[string]string{
				"resource_group": "my-rg",
				"kubeconfig":     "apiVersion: v1\nclusters:\n- cluster:\n    certificate-authority-data: LS0keykey==\n    server: https://id27182-b8s-default-76f11749.hcp.uksouth.azmk8s.io:443\n  name: id27182-b8s-default\ncontexts:\n- context:\n    cluster: id27182-b8s-default\n    user: clusterUser_id27182-b8s-default_id27182-b8s-default\n  name: id27182-b8s-default\ncurrent-context: id27182-b8s-default\nkind: Config\npreferences: {}\nusers:\n- name: clusterUser_id27182-b8s-default_id27182-b8s-default\n  user:\n    client-certificate-data: keykey=\n    client-key-data: keykey\n    token: mytoken",
			},
			shouldFail: false,
		},
	}

	for _, table := range tables {
		outVars, err := readOutVars(table.output)
		if err != nil {
			if !table.shouldFail {
				t.Errorf("readOutVars() failed to read variables: %s. Original error: %s", string(table.output), err)
			}
		} else {
			if table.shouldFail {
				t.Errorf("readOutVars() didn't failed to read variables: %s", string(table.output))
			} else {
				for key, value := range table.outVars {
					if outVars[key] != value {
						t.Errorf("readOutVars() didn't read variable properly. Get %s, must be %s", outVars[key], value)
					}
				}
			}
		}
	}
}
