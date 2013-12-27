package mergemap

import (
	"errors"
	"reflect"
)

var (
	MaxDepth = 32
)

// Merge recursively merges the src and dst maps. Key conflicts are resolved by
// preferring src, or recursively descending, if both src and dst are maps.
func Merge(dst, src map[string]interface{}) (map[string]interface{}, error) {
	return merge(dst, src, 0)
}

func merge(dst, src map[string]interface{}, depth int) (map[string]interface{}, error) {
	if depth > MaxDepth {
		return nil, errors.New("mergemap: maps too deeply nested")
	}
	for key, srcVal := range src {
		if dstVal, ok := dst[key]; ok {
			srcMap, srcMapOk := mapify(srcVal)
			dstMap, dstMapOk := mapify(dstVal)
			if srcMapOk && dstMapOk {
				var err error
				srcVal, err = merge(dstMap, srcMap, depth+1)
				if err != nil {
					return nil, err
				}
			}
		}
		dst[key] = srcVal
	}
	return dst, nil
}

func mapify(i interface{}) (map[string]interface{}, bool) {
	value := reflect.ValueOf(i)
	if value.Kind() == reflect.Map {
		m := map[string]interface{}{}
		for _, k := range value.MapKeys() {
			m[k.String()] = value.MapIndex(k).Interface()
		}
		return m, true
	}
	return map[string]interface{}{}, false
}
