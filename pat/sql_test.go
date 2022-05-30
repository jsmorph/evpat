package pat

import (
	"encoding/json"
	"fmt"
	"testing"
)

func ParseJSON(s string) interface{} {
	var x interface{}
	if err := json.Unmarshal([]byte(s), &x); err != nil {
		panic(err)
	}
	return x
}

func TestPat(t *testing.T) {
	f := func(p string) func(t *testing.T) {
		return func(t *testing.T) {
			p, err := ParsePattern(ParseJSON(p))
			if err != nil {
				t.Fatal(err)
			}
			fmt.Printf("%#v\n", p)
		}
	}

	t.Run("", f(`{"desires":{"likes":["tacos","chips"]}}`))
	t.Run("", f(`["tacos","chips"]`))
	t.Run("", f(`{"desires":{"likes":["tacos","chips"],"needs":[42]}}`))
}
