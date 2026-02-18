package engine

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"fmt"
)

func toNumber(v any) (float64, bool) {
	switch x := v.(type) {
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case json.Number:
		f, err := x.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(x, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func compareNumbers(a, b any, op string) bool {
	af, aok := toNumber(a)
	bf, bok := toNumber(b)
	if !aok || !bok {
		// try dates if RFC3339 strings
		at, aok2 := toTime(a)
		bt, bok2 := toTime(b)
		if aok2 && bok2 {
			switch op {
			case ">":
				return at.After(bt)
			case ">=":
				return at.After(bt) || at.Equal(bt)
			case "<":
				return at.Before(bt)
			case "<=":
				return at.Before(bt) || at.Equal(bt)
			}
		}
		return false
	}
	switch op {
	case ">":
		return af > bf
	case ">=":
		return af >= bf
	case "<":
		return af < bf
	case "<=":
		return af <= bf
	}
	return false
}

func toTime(v any) (time.Time, bool) {
	s, ok := v.(string)
	if !ok {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// compareAny: -1 if a<b, 0 if equal, 1 if a>b
func compareAny(a, b any) int {
	// number
	if af, aok := toNumber(a); aok {
		if bf, bok := toNumber(b); bok {
			if af < bf {
				return -1
			}
			if af > bf {
				return 1
			}
			return 0
		}
	}

	// date (RFC3339 strings)
	if at, aok := toTime(a); aok {
		if bt, bok := toTime(b); bok {
			if at.Before(bt) {
				return -1
			}
			if at.After(bt) {
				return 1
			}
			return 0
		}
	}

	// string
	as, aok := a.(string)
	bs, bok := b.(string)
	if aok && bok {
		if as < bs {
			return -1
		}
		if as > bs {
			return 1
		}
		return 0
	}

	aa := toString(a)
	bb := toString(b)
	if aa < bb {
		return -1
	}
	if aa > bb {
		return 1
	}
	return 0
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	default:
		return strings.TrimSpace(fmtAny(v))
	}
}

func fmtAny(v any) string {
    // avoid importing fmt in hot path unless needed
    switch x := v.(type) {
    case nil:
        return ""
    case string:
        return x
    case []byte:
        return string(x)
    default:
        // fallback — use fmt only when really needed
        return fmt.Sprintf("%v", v)
    }
}
