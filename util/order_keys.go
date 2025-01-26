package util

import (
	"fmt"
	"reflect"
	"sort"
)

func OrderedMapKeysOf(m interface{}) []string {
	v := reflect.ValueOf(m)

	if v.Kind() == reflect.Map {
		keys := make([]string, len(v.MapKeys()))
		i := 0
		for _, k := range v.MapKeys() {
			keys[i] = fmt.Sprint(k.Interface())
			i++
		}

		sort.Strings(keys)
		return keys
	}

	return []string{}
}
