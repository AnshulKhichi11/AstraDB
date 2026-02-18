package engine

import (
	"errors"
	"regexp"
	"strings"

	"testDB/internal/types"
)

func matchesFilter(doc types.Document, filter map[string]any) bool {
	if len(filter) == 0 {
		return true
	}

	// top-level logical
	if orVal, ok := filter["$or"]; ok {
		arr, ok := orVal.([]any)
		if !ok || len(arr) == 0 {
			return false
		}
		for _, item := range arr {
			m, ok := item.(map[string]any)
			if ok && matchesFilter(doc, m) {
				return true
			}
		}
		return false
	}

	if andVal, ok := filter["$and"]; ok {
		arr, ok := andVal.([]any)
		if !ok || len(arr) == 0 {
			return false
		}
		for _, item := range arr {
			m, ok := item.(map[string]any)
			if !ok {
				return false
			}
			if !matchesFilter(doc, m) {
				return false
			}
		}
		return true
	}

	// field conditions
	for key, want := range filter {
		if strings.HasPrefix(key, "$") {
			continue
		}

		got, exists := getNestedField(doc, key) // A2: nested address.city
		if opMap, ok := want.(map[string]any); ok {
			// operator object
			for op, opVal := range opMap {
				switch op {
				case "$exists":
					b, ok := opVal.(bool)
					if !ok {
						return false
					}
					if b && !exists {
						return false
					}
					if !b && exists {
						return false
					}

				case "$gt":
					if !compareNumbers(got, opVal, ">") {
						return false
					}
				case "$gte":
					if !compareNumbers(got, opVal, ">=") {
						return false
					}
				case "$lt":
					if !compareNumbers(got, opVal, "<") {
						return false
					}
				case "$lte":
					if !compareNumbers(got, opVal, "<=") {
						return false
					}
				case "$ne":
					if compareAny(got, opVal) == 0 {
						return false
					}
				case "$in":
					if !matchIn(got, opVal) {
						return false
					}
				case "$nin":
					if matchIn(got, opVal) {
						return false
					}
				case "$regex":
					if !matchRegex(got, opVal) {
						return false
					}
				case "$not":
					// $not: { $regex: "..."}  OR { $gt: 5 } etc.
					sub, ok := opVal.(map[string]any)
					if !ok {
						return false
					}
					tmp := map[string]any{key: sub}
					if matchesFilter(doc, tmp) {
						return false
					}
				case "$elemMatch":
					// field must be array; elemMatch is filter object for each element
					sub, ok := opVal.(map[string]any)
					if !ok {
						return false
					}
					if !matchElemMatch(got, sub) {
						return false
					}
				default:
					return false
				}
			}
			continue
		}

		// direct equality needs field exist
		if !exists {
			return false
		}
		if compareAny(got, want) != 0 {
			return false
		}
	}

	return true
}

func getNestedField(doc types.Document, path string) (any, bool) {
	parts := strings.Split(path, ".")
	var cur any = doc

	for _, p := range parts {
		var m map[string]any

		switch x := cur.(type) {
		case types.Document:
			m = map[string]any(x) // ✅ named type -> plain map
		case map[string]any:
			m = x
		default:
			return nil, false
		}

		v, ok := m[p]
		if !ok {
			return nil, false
		}
		cur = v
	}
	return cur, true
}


func matchIn(docValue any, arrayValue any) bool {
	arr, ok := arrayValue.([]any)
	if !ok {
		return false
	}
	for _, item := range arr {
		if compareAny(docValue, item) == 0 {
			return true
		}
	}
	return false
}

func matchRegex(got any, pattern any) bool {
	s, ok := got.(string)
	if !ok {
		return false
	}
	pat, ok := pattern.(string)
	if !ok {
		return false
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

func matchElemMatch(got any, subFilter map[string]any) bool {
	arr, ok := got.([]any)
	if !ok {
		return false
	}
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
	if matchesFilter(types.Document(m), subFilter) { return true }
} else if m2, ok := item.(types.Document); ok {
	if matchesFilter(m2, subFilter) { return true }
}

	}
	return false
}

func applyUpdateOperators(doc types.Document, update map[string]any) error {
	// Support: $set, $inc, $unset, $push, $pull, $rename
	for op, payload := range update {
		switch op {
		case "$set":
			m, ok := payload.(map[string]any)
			if !ok {
				return errors.New("$set must be object")
			}
			for k, v := range m {
				setNestedField(doc, k, v)
			}

		case "$inc":
			m, ok := payload.(map[string]any)
			if !ok {
				return errors.New("$inc must be object")
			}
			for k, v := range m {
				cur, _ := getNestedField(doc, k)
				cv, _ := toNumber(cur)
				iv, ok := toNumber(v)
				if !ok {
					return errors.New("$inc value must be number")
				}
				setNestedField(doc, k, cv+iv)
			}

		case "$unset":
			m, ok := payload.(map[string]any)
			if !ok {
				return errors.New("$unset must be object")
			}
			for k := range m {
				unsetNestedField(doc, k)
			}

		case "$push":
			m, ok := payload.(map[string]any)
			if !ok {
				return errors.New("$push must be object")
			}
			for k, v := range m {
				cur, exists := getNestedField(doc, k)
				if !exists || cur == nil {
					setNestedField(doc, k, []any{v})
					continue
				}
				arr, ok := cur.([]any)
				if !ok {
					return errors.New("$push target must be array")
				}
				arr = append(arr, v)
				setNestedField(doc, k, arr)
			}

		case "$pull":
			m, ok := payload.(map[string]any)
			if !ok {
				return errors.New("$pull must be object")
			}
			for k, v := range m {
				cur, exists := getNestedField(doc, k)
				if !exists || cur == nil {
					continue
				}
				arr, ok := cur.([]any)
				if !ok {
					return errors.New("$pull target must be array")
				}
				out := make([]any, 0, len(arr))
				for _, it := range arr {
					if compareAny(it, v) != 0 {
						out = append(out, it)
					}
				}
				setNestedField(doc, k, out)
			}

		case "$rename":
			m, ok := payload.(map[string]any)
			if !ok {
				return errors.New("$rename must be object")
			}
			for from, toAny := range m {
				to, ok := toAny.(string)
				if !ok {
					return errors.New("$rename value must be string")
				}
				val, exists := getNestedField(doc, from)
				if exists {
					unsetNestedField(doc, from)
					setNestedField(doc, to, val)
				}
			}

		default:
			return errors.New("unsupported update operator: " + op)
		}
	}
	return nil
}

func setNestedField(doc types.Document, path string, value any) {
	parts := strings.Split(path, ".")
	last := len(parts) - 1
	cur := map[string]any(doc)
	for i := 0; i < last; i++ {
		p := parts[i]
		next, ok := cur[p].(map[string]any)
		if !ok {
			next = map[string]any{}
			cur[p] = next
		}
		cur = next
	}
	cur[parts[last]] = value
}

func unsetNestedField(doc types.Document, path string) {
	parts := strings.Split(path, ".")
	last := len(parts) - 1
	cur := map[string]any(doc)
	for i := 0; i < last; i++ {
		p := parts[i]
		next, ok := cur[p].(map[string]any)
		if !ok {
			return
		}
		cur = next
	}
	delete(cur, parts[last])
}
