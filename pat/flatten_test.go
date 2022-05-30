package pat

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestDB(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var (
		eventsTable = "events"
		indexTable  = "eventsind"
	)

	for _, s := range CreateEventsTables(eventsTable, indexTable) {
		_, err := db.Exec(s.S)
		if err != nil {
			t.Fatal(err)
		}
	}

	t0 := time.Now().UTC()

	{

		eid := 100

		e := ParseJSON(`{"likes":"tacos","needs":42}`)

		ss, err := GenerateInsert(eventsTable, indexTable, t0, eid, e)
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range ss {
			if _, err := db.Exec(s.S, s.Args...); err != nil {
				t.Fatal(err)
			}
		}
	}

	t0 = t0.Add(-10 * time.Second)

	{
		x := ParseJSON(`{"likes":["tacos","queso"],"needs":[42]}`)
		q, err := GenerateQuery(eventsTable, indexTable, t0, x)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("generate:\n%s\n", q.S)
		fmt.Printf("%v\n", q.Args)

		rs, err := db.Query(q.S, q.Args...)
		if err != nil {
			t.Fatal(err)
		}
		defer rs.Close()
		for rs.Next() {
			var (
				eid   string
				event string
			)
			if err = rs.Scan(&eid, &event); err != nil {
				t.Fatal(err)
			}
			fmt.Printf("result: %v: %s\n", eid, event)
		}

	}
}
