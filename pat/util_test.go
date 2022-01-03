package pat

import (
	"encoding/json"
)

func P(s string) interface{} {
	var x interface{}
	if err := json.Unmarshal([]byte(s), &x); err != nil {
		panic(err)
	}
	return x
}
