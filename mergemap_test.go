package mergemap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestMerge(t *testing.T) {
	for _, tuple := range []struct {
		src      string
		dst      string
		expected string
	}{
		{
			src:      `{}`,
			dst:      `{}`,
			expected: `{}`,
		},
		{
			src:      `{"b":2}`,
			dst:      `{"a":1}`,
			expected: `{"a":1,"b":2}`,
		},
		{
			src:      `{"a":0}`,
			dst:      `{"a":1}`,
			expected: `{"a":0}`,
		},
		{
			src:      `{"a":{       "y":2}}`,
			dst:      `{"a":{"x":1       }}`,
			expected: `{"a":{"x":1, "y":2}}`,
		},
		{
			src:      `{"a":{"x":2}}`,
			dst:      `{"a":{"x":1}}`,
			expected: `{"a":{"x":2}}`,
		},
		{
			src:      `{"a":{       "y":7, "z":8}}`,
			dst:      `{"a":{"x":1, "y":2       }}`,
			expected: `{"a":{"x":1, "y":7, "z":8}}`,
		},
		{
			src:      `{"1": { "b":1, "2": { "3": {         "b":3, "n":[1,2]} }        }}`,
			dst:      `{"1": {        "2": { "3": {"a":"A",        "n":"xxx"} }, "a":3 }}`,
			expected: `{"1": { "b":1, "2": { "3": {"a":"A", "b":3, "n":[1,2]} }, "a":3 }}`,
		},
	} {
		var dst map[string]interface{}
		if err := json.Unmarshal([]byte(tuple.dst), &dst); err != nil {
			t.Error(err)
			continue
		}

		var src map[string]interface{}
		if err := json.Unmarshal([]byte(tuple.src), &src); err != nil {
			t.Error(err)
			continue
		}

		var expected map[string]interface{}
		if err := json.Unmarshal([]byte(tuple.expected), &expected); err != nil {
			t.Error(err)
			continue
		}

		got, err := Merge(dst, src)
		if err != nil {
			t.Errorf("expected nil error, got %s", err)
		}
		assert(t, expected, got)
	}
}

func TestMergeWithMaxDepth(t *testing.T) {
	dst := `{"a": {"b": {"c": "d"}}}`
	src := `{"a": {"b": {"c": "e"}}}`
	_, err := Merge(unmarshal(dst), unmarshal(src), WithMaxDepth(1))
	if err != ErrTooDeep {
		t.Errorf("expected %s, got %s", ErrTooDeep, err)
	}
}

func TestMergeWithAppendSlice(t *testing.T) {
	dst := `{"a": ["b", true, null]}`
	src := `{"a": [1, 1.5, {"c": "d"}]}`
	expected := `{"a": ["b", true, null, 1, 1.5, {"c": "d"}]}`
	got, err := Merge(unmarshal(dst), unmarshal(src), WithAppendSlice)
	if err != nil {
		t.Errorf("expected nil error, got %s", err)
	}
	assert(t, unmarshal(expected), got)
}

func TestMergeWithKeyStringer(t *testing.T) {
	dst := map[string]interface{}{"a": map[interface{}]interface{}{false: nil, nil: nil}}
	src := map[string]interface{}{"a": map[interface{}]interface{}{true: nil}}
	expected := map[string]interface{}{"a": map[string]interface{}{"true": nil, "false": nil, "null": nil}}

	keyStringer := func(mapKey reflect.Value) string {
		if mapKey.Kind() == reflect.Interface && mapKey.IsNil() {
			return "null"
		}
		return fmt.Sprint(mapKey)
	}
	got, err := Merge(dst, src, WithKeyStringer(keyStringer))
	if err != nil {
		t.Errorf("expected nil error, got %s", err)
	}
	assert(t, expected, got)
}

func assert(t *testing.T, expected, got map[string]interface{}) {
	expectedBuf, err := json.Marshal(expected)
	if err != nil {
		t.Error(err)
		return
	}
	gotBuf, err := json.Marshal(got)
	if err != nil {
		t.Error(err)
		return
	}
	if bytes.Compare(expectedBuf, gotBuf) != 0 {
		t.Errorf("expected %s, got %s", string(expectedBuf), string(gotBuf))
		return
	}
}

func unmarshal(s string) map[string]interface{} {
	var target map[string]interface{}
	if err := json.Unmarshal([]byte(s), &target); err != nil {
		panic(err)
	}
	return target
}
