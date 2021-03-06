package pat

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
)

type TestCase struct {
	// AWS indicates that this test should work with AWS
	// EventBridge TestEventPattern.
	//
	// See TestAWS() below.
	AWS bool `json:"aws,omitempty"`

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

	cfg := &Cfg{}

	for _, tc := range cases {
		t.Run(JSON(tc), func(t *testing.T) {
			c, err := cfg.ParsePattern(tc.Pat)
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

func BenchmarkBasic(b *testing.B) {

	bs, err := ioutil.ReadFile("tests.json")
	if err != nil {
		b.Fatal(err)
	}

	var cases []TestCase
	if err := json.Unmarshal(bs, &cases); err != nil {
		b.Fatal(err)
	}

	cfg := &Cfg{}

	var (
		pats = make([]Constraint, len(cases))
		msgs = make([]interface{}, len(cases))
	)

	for i, tc := range cases {
		c, err := cfg.ParsePattern(tc.Pat)
		if err != nil {
			b.Fatal(err)
		}
		pats[i] = c
		msgs[i] = tc.Msg
	}

	b.ResetTimer()

	n := 0
LOOP:
	for {
		for i, p := range pats {
			if _, err := p.Matches(msgs[i]); err != nil {
				b.Fatal(err)
			}
			n++
			if b.N <= n {
				break LOOP
			}
		}
	}
}

// TestWithAWS calls the AWS TestEventPattern API via the AWS Go SDK v2.
func TestWithAWS(t *testing.T) {
	if os.Getenv("AWS_PROFILE") == "" &&
		os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		t.Skip("AWS access not configured")
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("unable to load AWS SDK config, %v", err)
	}

	svc := eventbridge.NewFromConfig(cfg)

	bs, err := ioutil.ReadFile("tests.json")
	if err != nil {
		t.Fatal(err)
	}

	var cases []TestCase
	if err := json.Unmarshal(bs, &cases); err != nil {
		t.Fatal(err)
	}

	// The AWS TestEventPattern API insists on having certain
	// top-level values.

	ensure := func(msg interface{}) interface{} {
		m, is := msg.(map[string]interface{})
		if !is {
			t.Fatalf("%s (%T) isn't a map", JSON(msg), msg)
		}
		for _, p := range []string{
			"id", "account", "source", "region", "detail-type"} {
			if _, have := m[p]; !have {
				m[p] = "something"
			}
		}
		m["time"] = time.Now().UTC().Format(time.RFC3339)
		m["resources"] = []string{"a", "b"}

		return m
	}

	for _, tc := range cases {
		if !tc.AWS {
			continue
		}

		t.Run(JSON(tc), func(t *testing.T) {
			in := &eventbridge.TestEventPatternInput{
				Event:        aws.String(JSON(ensure(tc.Msg))),
				EventPattern: aws.String(JSON(tc.Pat)),
			}
			out, err := svc.TestEventPattern(ctx, in)
			if err != nil {
				if tc.Error {
					return
				} else {
					t.Fatal(err)
				}
			}
			if out.Result != tc.Matches {
				t.Fatal(out.Result)
			}
		})
	}
}
