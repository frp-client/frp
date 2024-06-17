package format

import "encoding/json"

func ToJsonString(v any) string {
	buf, _ := json.Marshal(v)
	return string(buf)
}
