package utils

import "encoding/json"

func JSONBytesToMap(b []byte) (map[string]interface{}, error) {
	var m map[string]interface{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
