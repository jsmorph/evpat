package pat

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

type TestCase struct {
	Pat     interface{} `json:"pat"`
	Msg     interface{} `json:"msg"`
	Matches bool        `json:"matches"`
	Error   bool        `json:"error,omitempty"`
}

func TestBasic(t *testing.T) {

	bs, err := ioutil.ReadFile("tests.json")
	if err != nil {
		t.Fatal(err)
	}

	var cases []TestCase
	if err := json.Unmarshal(bs, &cases); err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		t.Run(JSON(tc), func(t *testing.T) {
			c, err := ParsePattern(tc.Pat)
			if err != nil {
				t.Fatal(err)
			}
			got, err := c.Matches(tc.Msg)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.Matches {
				t.Fatal(got)
			}
		})
	}
}
