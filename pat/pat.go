package pat

import (
	"fmt"
	"strings"
)

func Matches(pat, y interface{}) (bool, error) {
	switch v1 := pat.(type) {
	case Constraint:
		return v1.Matches(y)
	case map[string]interface{}:
		switch v2 := y.(type) {
		case map[string]interface{}:
			for k, x := range v1 {
				if z, have := v2[k]; have {
					if ok, err := Matches(x, z); !ok || err != nil {
						return ok, err
					}
				} else {
					if c, is := x.(*Exists); is {
						return !c.Value, nil
					}
				}
			}
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

type Constraints []Constraint

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

// Cfg can store limits and options for parsing patterns.
type Cfg struct {
}

var DefaultCfg = &Cfg{}

// ParsePattern just calls DefaultCfg.ParsePattern().
func ParsePattern(x interface{}) (Constraint, error) {
	return DefaultCfg.ParsePattern(x)
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

type Constraint interface {
	Matches(msg interface{}) (bool, error)
}

type Map map[string]interface{}

func (c Map) Matches(x interface{}) (bool, error) {
	m, is := x.(map[string]interface{})
	if !is {
		return false, nil
	}
	for p, v := range m {
		if v2, have := c[p]; have {
			if c1, is := v2.(Constraint); is {
				if ok, err := c1.Matches(v); !ok || err != nil {
					return ok, err
				}
			}
		}
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

func ToNumber(x interface{}) (float64, error) {
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
	x, err := ToNumber(msg)
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
	return c.Value, nil
}
