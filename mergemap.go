package mergemap

import (
	"errors"
	"reflect"
)

var ErrTooDeep = errors.New("too deep")

type Config struct {
	depthCheck  bool
	maxDepth    int
	appendSlice bool
	keyStringer func(mapKey reflect.Value) string
}

func WithMaxDepth(maxDepth int) func(*Config) {
	return func(config *Config) {
		config.depthCheck = true
		config.maxDepth = maxDepth
	}
}

func WithAppendSlice(config *Config) {
	config.appendSlice = true
}

// key stringer function is used when setting a map key (which must be a string) from a reflect.Value.
// if not provided, reflect.Value's String method is used.
func WithKeyStringer(f func(mapKey reflect.Value) string) func(*Config) {
	return func(config *Config) {
		config.keyStringer = f
	}
}

// Merge recursively merges the src and dst maps. Key conflicts are resolved by
// preferring src, or recursively descending, if both src and dst are maps.
func Merge(dst, src map[string]interface{}, opts ...func(*Config)) (map[string]interface{}, error) {
	config := &Config{}
	for _, opt := range opts {
		opt(config)
	}
	return merge(dst, src, 0, config)
}

func merge(dst, src map[string]interface{}, depth int, config *Config) (map[string]interface{}, error) {
	if config.depthCheck && depth > config.maxDepth {
		return nil, ErrTooDeep
	}
	for key, srcVal := range src {
		if dstVal, ok := dst[key]; ok {
			srcMap, srcMapOk := mapify(srcVal, config.keyStringer)
			dstMap, dstMapOk := mapify(dstVal, config.keyStringer)
			if srcMapOk && dstMapOk {
				var err error
				srcVal, err = merge(dstMap, srcMap, depth+1, config)
				if err != nil {
					return nil, err
				}
			} else if config.appendSlice {
				// if both are slices, then append
				srcSlice, srcSliceOk := srcVal.([]interface{})
				dstSlice, dstSliceOk := dstVal.([]interface{})
				if srcSliceOk && dstSliceOk {
					for _, elem := range srcSlice {
						dstSlice = append(dstSlice, elem)
					}
					srcVal = dstSlice
				}
			}
		}
		dst[key] = srcVal
	}
	return dst, nil
}

func mapify(i interface{}, keyStringer func(mapKey reflect.Value) string) (map[string]interface{}, bool) {
	value := reflect.ValueOf(i)
	if value.Kind() == reflect.Map {
		m := map[string]interface{}{}
		for _, k := range value.MapKeys() {
			kString := k.String()
			if keyStringer != nil {
				kString = keyStringer(k)
			}
			m[kString] = value.MapIndex(k).Interface()
		}
		return m, true
	}
	return nil, false
}
