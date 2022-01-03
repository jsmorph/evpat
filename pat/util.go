package pat

import (
	"bytes"
	"encoding/json"
)

func JSON(x interface{}) string {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(&x)
	if err != nil {
		panic(err)
	}
	return string(buffer.Bytes())
}
