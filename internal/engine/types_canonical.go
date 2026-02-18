package engine

import (
	"bytes"
	"encoding/json"
	"strings"
	"time"

	"testDB/internal/types"
)

// canonicalizeDocument converts a types.Document → map[string]any with canonical types
// (time.Time for dates, int64/float64 for numbers, etc.)
func canonicalizeDocument(doc types.Document) types.Document {
	result := canonicalizeAny(doc)
	// safe type assertion - in practice this should always be map[string]any
	m, ok := result.(map[string]any)
	if !ok {
		// this case should practically never happen
		return types.Document{}
	}
	return types.Document(m)
}

// canonicalizeAny recursively converts values to "canonical" Go types
func canonicalizeAny(v any) any {
	switch x := v.(type) {
	case types.Document:
		m := make(map[string]any, len(x))
		for k, v2 := range x {
			m[k] = canonicalizeAny(v2)
		}
		return m

	case map[string]any:
		// Extended JSON / MongoDB extended JSON style → ISODate
		if len(x) == 1 {
			if ds, ok := x["$date"]; ok {
				if s, ok2 := ds.(string); ok2 {
					// RFC3339Nano supports nanoseconds and various timezone formats
					if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
						return t
					}
					// You may also want to support legacy MongoDB $date formats here
					// (milliseconds since epoch, etc.) — but RFC3339 is most common today
				}
			}
		}

		// normal map → recurse
		m := make(map[string]any, len(x))
		for k, v2 := range x {
			m[k] = canonicalizeAny(v2)
		}
		return m

	case []any:
		out := make([]any, 0, len(x))
		for _, it := range x {
			out = append(out, canonicalizeAny(it))
		}
		return out

	case json.Number:
		s := x.String()

		// Fast path: no decimal point or exponent → likely integer
		if !strings.ContainsAny(s, ".eE") {
			if i, err := x.Int64(); err == nil {
				return i
			}
		}

		// Otherwise → try float
		if f, err := x.Float64(); err == nil {
			return f
		}

		// Very rare fallback (invalid number format)
		return s

	default:
		// bool, string, time.Time, nil, etc. → pass through
		return x
	}
}

// canonicalizeAnyMap is a convenience wrapper when you already have map[string]any
func canonicalizeAnyMap(m map[string]any) map[string]any {
	if m == nil {
		return make(map[string]any)
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = canonicalizeAny(v)
	}
	return out
}

// decodeJSONUseNumber is useful when you want json.Number instead of float64
// (typically used in API/HTTP handlers before canonicalization)
func decodeJSONUseNumber(b []byte, out any) error {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	return dec.Decode(out)
}