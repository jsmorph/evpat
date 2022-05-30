package pat

type Branch struct {
	Limbs      []string
	Constraint Constraint
}

func FlattenPattern(x interface{}, f func(*Branch), prefix []string) {
	if prefix == nil {
		prefix = make([]string, 0, 4)
	}
	switch vv := x.(type) {
	case Map:
		for p, v := range vv {
			FlattenPattern(v, f, append(prefix, p))
		}
	case Constraint:
		b := &Branch{
			Limbs:      prefix,
			Constraint: vv,
		}
		f(b)
	}
}

func FlattenEvent(x interface{}, f func(k []string, v interface{}), prefix []string) {
	if prefix == nil {
		prefix = make([]string, 0, 4)
	}
	switch vv := x.(type) {
	case map[string]interface{}:
		for k, v := range vv {
			FlattenEvent(v, f, append(prefix, k))
		}
	case []interface{}:
		for _, v := range vv {
			FlattenEvent(v, f, prefix)
		}
	default:
		f(prefix, vv)
	}
}
