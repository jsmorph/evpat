package pat

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Statement struct {
	S    string
	Args []interface{}
}

type WhereConjunct Statement

func (c *Literal) Where(bName, bVal, vName string) *WhereConjunct {
	return &WhereConjunct{
		S:    fmt.Sprintf("(%s = ? AND %s = ?)", bName, vName),
		Args: []interface{}{bVal, c.Value},
	}
}

func (c *AnythingBut) Where(bName, bVal, vName string) *WhereConjunct {
	return &WhereConjunct{
		S:    fmt.Sprintf("(%s = ? AND %s IS NOT ?)", bName, vName),
		Args: []interface{}{bVal, c.Value},
	}
}

func (c *Prefix) Where(bName, bVal, vName string) *WhereConjunct {
	return &WhereConjunct{
		S:    fmt.Sprintf("(%s = ? AND %s LIKE ? + '%%')", bName, vName),
		Args: []interface{}{bVal, c.Value},
	}
}

func (c *Exists) Where(bName, bVal, vName string) *WhereConjunct {
	return &WhereConjunct{
		S:    fmt.Sprintf("(%s = ? AND %s IS NOT NULL)", bName, vName),
		Args: []interface{}{bVal},
	}
}

func (cs Constraints) Where(bName, bVal, vName string) *WhereConjunct {
	vals := make([]interface{}, 0, len(cs)+1)
	vals = append(vals, bVal)
	nargs := 0
	for _, c := range cs {
		switch vv := c.(type) {
		case *Literal:
			vals = append(vals, vv.Value)
			nargs++
		default:
			panic(fmt.Errorf("can't handle %T", c))
		}
	}
	return &WhereConjunct{
		S: fmt.Sprintf("(%s = ? AND %s IN (%s))",
			bName, vName, qmarks(nargs)),
		Args: vals,
	}
}

func (c Map) Where(bName, bVal, vName string) *WhereConjunct {
	return &WhereConjunct{
		S:    fmt.Sprintf("(%s = ? AND %s IS NOT IMPLEMENTED)", bName, vName),
		Args: []interface{}{bVal},
	}
}

func qmarks(n int) string {
	acc := make([]string, n)
	for i := range acc {
		acc[i] = "?"
	}
	return strings.Join(acc, ",")
}

func (c *Numeric) Where(bName, bVal, vName string) *WhereConjunct {
	conjuncts := make([]*WhereConjunct, 0, len(c.Predicates))
	for _, p := range c.Predicates {
		conjunct := &WhereConjunct{
			S:    fmt.Sprintf("(%s = ? AND %s %s ?)", bName, vName, p.Relation),
			Args: []interface{}{bVal, p.Value},
		}
		conjuncts = append(conjuncts, conjunct)
	}

	return And(conjuncts)
}

func And(conjuncts []*WhereConjunct) *WhereConjunct {
	var (
		ss   = make([]string, 0, len(conjuncts))
		args = make([]interface{}, 0, len(conjuncts))
	)
	for _, c := range conjuncts {
		ss = append(ss, c.S)
		args = append(args, c.Args...)
	}
	conjunct := &WhereConjunct{
		S:    strings.Join(ss, " AND "),
		Args: args,
	}

	return conjunct
}

func GenerateQuery(eventsTable, indexTable string, t0 time.Time, pattern interface{}) (*Statement, error) {
	p, err := ParsePattern(pattern)
	if err != nil {
		return nil, err
	}
	var (
		ss            = make([]string, 0, 4)
		args          = make([]interface{}, 0, 4)
		joins         = make([]string, 0, 4)
		previousTable = "t0"
		i             = 0
	)

	FlattenPattern(p, func(b *Branch) {
		fmt.Printf("branch %#v\n", *b)
		var (
			table = fmt.Sprintf("t%d", i)
			bName = table + ".branch"
			bVal  = strings.Join(b.Limbs, "/")
			vName = table + ".value"
			w     = b.Constraint.Where(bName, bVal, vName)
			s     = fmt.Sprintf("(%s.ts > ? AND %s)", table, w.S)
		)
		ss = append(ss, s)
		args = append(args, TS(t0))
		args = append(args, w.Args...)
		if 0 < i {
			joins = append(joins,
				fmt.Sprintf("LEFT JOIN %s AS %s ON %s.eid = %s.eid",
					indexTable, table, previousTable, table))
		}
		previousTable = table
		i++
	}, nil)

	s := &Statement{
		S: fmt.Sprintf(`
SELECT t0.eid AS eid, event
FROM %s AS t0
%s
LEFT JOIN %s ON %s.eid = t0.eid
WHERE 
 %s
`,
			indexTable,
			strings.Join(joins, "\n"),
			eventsTable, eventsTable,
			strings.Join(ss, " AND \n ")),
		Args: args,
	}

	return s, nil
}

func TS(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

func GenerateInsert(eventsTable, indexTable string, t0 time.Time, eid interface{}, event interface{}) ([]*Statement, error) {
	js, err := json.Marshal(&event)
	if err != nil {
		return nil, err
	}

	var (
		args = make([]interface{}, 0, 4)
		vals = make([]string, 0, 4)
	)

	f := func(prefix []string, val interface{}) error {
		args = append(args, eid)
		args = append(args, TS(t0))
		args = append(args, strings.Join(prefix, "/"))
		args = append(args, val)
		vals = append(vals, "(?,?,?,?)")
		return nil
	}

	if err := FlattenEvent(event, f, nil); err != nil {
		return nil, err
	}

	return []*Statement{
		&Statement{
			S: fmt.Sprintf("INSERT INTO %s (eid,ts,branch,value) VALUES %s",
				indexTable, strings.Join(vals, ", ")),
			Args: args,
		},
		&Statement{
			S: fmt.Sprintf("INSERT INTO %s (eid,event) VALUES (?,?)",
				eventsTable),
			Args: []interface{}{eid, string(js)},
		},
	}, nil
}

func CreateEventsTables(eventsTable, indexTable string) []*Statement {
	// ToDo: Foreign key, ...

	return []*Statement{
		&Statement{
			S: fmt.Sprintf("CREATE TABLE %s(eid TEXT PRIMARY KEY, event TEXT)",
				eventsTable),
		},
		&Statement{
			S: fmt.Sprintf("CREATE TABLE %s(eid TEXT, ts TEXT, branch TEXT, value TEXT)",
				indexTable),
		},
	}
}
