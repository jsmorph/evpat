package pat

import (
	"fmt"
	"strings"
)

// Constraint is probably really a pattern.  Should have been called
// "Pattern", I guess.
type Constraint interface {
	Matches(msg interface{}) (bool, error)
}

// Cfg can store limits and options for parsing patterns.
type Cfg struct {
}

var DefaultCfg = &Cfg{}

// ParsePattern just calls DefaultCfg.ParsePattern().
//
// ToDo: Fix name of function or name of return type.
func ParsePattern(x interface{}) (Constraint, error) {
	return DefaultCfg.ParsePattern(x)
}

// Matches is a predicate that reports whether its second argument
// matches the pattern given in the first argument.
//
// This function is called recursively on it's first argument.  When
// that first argument is a Constraint, the the Matches method of that
// interface is called.  Otherwise, the match check is mostly literal.
func Matches(pat, y interface{}) (bool, error) {
	switch v1 := pat.(type) {
	case Constraint:
		return v1.Matches(y)
	case map[string]interface{}:
		switch v2 := y.(type) {
		case map[string]interface{}:
			if len(v1) == 0 {
				return true, nil
			}
			for k, x := range v1 {
				if z, have := v2[k]; have {
					if c, is := x.(*Exists); is {
						return c.Value, nil
					}
					ok, err := Matches(x, z)
					if err != nil {
						return false, err
					}
					return ok, nil
				}
				if c, is := x.(*Exists); is {
					return !c.Value, nil
				}
				return false, nil
			}
		default:
			return false, nil
		}

	case string:
		switch v2 := y.(type) {
		case string:
			return v1 == v2, nil
		}
	case int:
		switch v2 := y.(type) {
		case int:
			return v1 == v2, nil
		case int64:
			return int64(v1) == v2, nil
		}
	case int64:
		switch v2 := y.(type) {
		case int:
			return v1 == int64(v2), nil
		case int64:
			return v1 == v2, nil
		}
	case float64:
		switch v2 := y.(type) {
		case float64:
			return v1 == v2, nil
		}
	}

	return false, nil
}

// Constraints is just a list of Constraints, and a Constraints is
// itself a Constraint.
type Constraints []Constraint

var Missing = struct{}{}

// Matches performs matching with the special Constraint array
// behavior.
func (cs Constraints) Matches(x interface{}) (bool, error) {
	if len(cs) == 0 {
		return true, nil
	}

	xs, is := x.([]interface{})
	if !is {
		xs = []interface{}{x}
	}

	for _, x := range xs {
		for _, c := range cs {
			if ok, err := c.Matches(x); ok && err == nil {
				// Drop errors?
				return true, nil
			}
		}
	}

	return false, nil
}

func (cfg *Cfg) parseConstraint(x interface{}) (Constraint, error) {
	switch vv := x.(type) {
	default:
		return &Literal{
			Value: x,
		}, nil
	case map[string]interface{}:

		if y, have := vv["prefix"]; have {
			s, is := y.(string)
			if !is {
				return nil, fmt.Errorf("bad prefix '%#v'", y)
			}
			return &Prefix{
				Value: s,
			}, nil
		}

		if y, have := vv["exists"]; have {
			b, is := y.(bool)
			if !is {
				return nil, fmt.Errorf("bad exists value '%#v'", y)
			}
			return &Exists{
				Value: b,
			}, nil
		}

		if y, have := vv["numeric"]; have {
			ys, is := y.([]interface{})
			if !is {
				return nil, fmt.Errorf("bad numeric '%#v'", y)
			}
			if len(ys)%2 != 0 {
				return nil, fmt.Errorf("bad numeric array size '%#v'", ys)
			}
			ns := &Numeric{
				Predicates: make([]NumericPredicate, 0, len(ys)/2),
			}
			for i := 0; i < len(ys); i += 2 {
				s, is := ys[i].(string)
				if !is {
					return nil, fmt.Errorf("bad numeric relation '%#v'", ys[i])
				}
				p := NumericPredicate{
					Relation: s,
				}
				switch vv := ys[i+1].(type) {
				case int:
					p.Value = float64(vv)
				case float64:
					p.Value = vv
				case int64:
					p.Value = float64(vv)
				default:
					return nil, fmt.Errorf("bad numeric relation value '%#v'", ys[i+1])
				}
				ns.Predicates = append(ns.Predicates, p)
			}
			return ns, nil
		}

		return Map(vv), nil
	}
}

// ParsePattern parses a Constraint from a plain value.
//
// ToDo: Fix name of function or name of return type.
func (cfg *Cfg) ParsePattern(x interface{}) (Constraint, error) {
	switch vv := x.(type) {
	case []interface{}:
		cs := make(Constraints, len(vv))
		for i, v := range vv {
			c, err := cfg.parseConstraint(v)
			if err != nil {
				return nil, err
			}
			cs[i] = c
		}
		return cs, nil
	case map[string]interface{}:
		m := make(Map, len(vv))
		for k, v := range vv {
			c, err := cfg.ParsePattern(v)
			if err != nil {
				return nil, err
			}
			m[k] = c
		}
		return m, nil
	default:
		return &Literal{
			Value: x,
		}, nil
	}
}

// pass is a Constraint with a Matches method that always returns
// true.
type pass struct {
}

// Matches always returns true.
func (c *pass) Matches(msg interface{}) (bool, error) {
	return true, nil
}

// Pass is a handy (?) Constraint with a Matches method that always
// returns true.
var Pass = &pass{}

type Map map[string]interface{}

func (c Map) Matches(x interface{}) (bool, error) {
	m, is := x.(map[string]interface{})
	if !is {
		return false, nil
	}
	if 0 == len(c) {
		return true, nil
	}
	for p, v1 := range c {
		if v2, have := m[p]; have {
			if pc, is := v1.(*Exists); is {
				return pc.Value, nil
			}
			return Matches(v1, v2)
		}
		if pc, is := v1.(*Exists); is {
			return !pc.Value, nil
		}
		if pc, is := v1.(Constraints); is {
			return pc.Matches(Missing)
		}
		return false, nil
	}

	return true, nil
}

type Literal struct {
	Value interface{}
}

func (c *Literal) Matches(x interface{}) (bool, error) {
	return Matches(c.Value, x)
}

type Prefix struct {
	Value string
}

func (c *Prefix) Matches(x interface{}) (bool, error) {
	s, is := x.(string)
	if !is {
		return false, nil
	}
	return strings.HasPrefix(s, c.Value), nil
}

type Numeric struct {
	Predicates []NumericPredicate
}

type NumericPredicate struct {
	Relation string
	Value    float64
}

func toNumber(x interface{}) (float64, error) {
	switch vv := x.(type) {
	case float64:
		return vv, nil
	case int:
		return float64(vv), nil
	case int64:
		return float64(vv), nil
	case uint64:
		return float64(vv), nil
	}
	return 0, fmt.Errorf("bad number %#v", x)

}

func (c *NumericPredicate) Matches(msg interface{}) (bool, error) {
	x, err := toNumber(msg)
	if err != nil {
		return false, nil
	}
	switch c.Relation {
	default:
		return false, fmt.Errorf("unknown numeric relation '%s'", c.Relation)
	case "<":
		return x < c.Value, nil
	case "<=":
		return x <= c.Value, nil
	case ">":
		return x > c.Value, nil
	case ">=":
		return x >= c.Value, nil
	case "=":
		return x == c.Value, nil
	}
	return false, nil
}

func (c *Numeric) Matches(msg interface{}) (bool, error) {
	for _, p := range c.Predicates {
		if matches, err := p.Matches(msg); !matches || err != nil {
			return matches, err
		}
	}
	return true, nil
}

type Exists struct {
	Value bool
}

func (c *Exists) Matches(msg interface{}) (bool, error) {
	if msg == Missing {
		return !c.Value, nil
	}
	return c.Value, nil
}
