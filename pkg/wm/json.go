package wm

import "encoding/json"

// IsJSON returns true if provided str is valid json
func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}
