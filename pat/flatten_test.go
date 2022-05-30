package pat

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestFlatten(t *testing.T) {
	x := ParseJSON(`{"likes":["tacos","queso"],"needs":[42],"when":["now","later"]}`)
	p, err := ParsePattern(x)
	if err != nil {
		t.Fatal(err)
	}
	var (
		ss            = make([]string, 0, 4)
		args          = make([]interface{}, 0, 4)
		i             = 0
		joins         = make([]string, 0, 4)
		realTable     = "pats"
		previousTable = "t0"
		t0            = time.Now().UTC()
	)

	FlattenPattern(p, func(b *Branch) {
		fmt.Printf("branch %#v\n", *b)
		var (
			table = fmt.Sprintf("t%d", i)
			bName = table + ".branch"
			bVal  = strings.Join(b.Limbs, "/")
			vName = table + ".value"
			w     = b.Constraint.Where(bName, bVal, vName)
			s     = fmt.Sprintf("(%s AND %s.ts > ?)", w.S, table)
		)
		ss = append(ss, s)
		args = append(args, t0)
		args = append(args, w.Args...)
		if 0 < i {
			joins = append(joins,
				fmt.Sprintf("LEFT JOIN %s AS %s ON %s.eid = %s.eid",
					realTable, table, previousTable, table))
		}
		previousTable = table
		i++
	}, nil)

	fmt.Printf("SELECT * FROM %s AS t0\n", realTable)
	clause := fmt.Sprintf("WHERE %s", strings.Join(ss, " AND "))
	fmt.Printf("%s\n", clause)
	fmt.Printf("%v\n", args)
	fmt.Printf("%s\n", strings.Join(joins, ", "))

	{
		q, err := GenerateQuery("pats", t0, x)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("generate:\n%s\n", q.S)
		fmt.Printf("%v\n", q.Args)
	}

	e := ParseJSON(`{"likes":"tacos","when":{"every":"day"}}`)

	g := func(prefix []string, val interface{}) {
		fmt.Printf("%v %#v\n", prefix, val)
		fmt.Printf("INSERT INTO %s (ts,branch,value) VALUES (?,?,?)\n", realTable)
		fmt.Printf("%v\n", []interface{}{time.Now().UTC(), strings.Join(prefix, "/"), val})
	}

	FlattenEvent(e, g, nil)
}

func TestDB(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	realTable := "events"

	{

		sqlStmt := `CREATE TABLE events (eid TEXT, ts TEXT, branch TEXT, value TEXT)`
		_, err = db.Exec(sqlStmt)
		if err != nil {
			t.Fatal(err)
		}
	}

	t0 := time.Now().UTC()

	{

		eid := 100

		e := ParseJSON(`{"likes":"tacos","needs":42}`)

		g := func(prefix []string, val interface{}) {

			s := fmt.Sprintf("INSERT INTO %s (eid,ts,branch,value) VALUES (?,?,?,?)\n", realTable)
			args := []interface{}{eid, TS(t0), strings.Join(prefix, "/"), val}
			fmt.Printf("%s%v\n", s, args)

			if _, err = db.Exec(s, args...); err != nil {
				t.Fatal(err)
			}
		}

		FlattenEvent(e, g, nil)
	}

	t0 = t0.Add(-10 * time.Second)

	{
		x := ParseJSON(`{"likes":["tacos","queso"],"needs":[42]}`)
		q, err := GenerateQuery(realTable, t0, x)
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
			var eid string
			if err = rs.Scan(&eid); err != nil {
				t.Fatal(err)
			}
			fmt.Printf("result: %v\n", eid)
		}

	}
}
