package compile

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/flow"
	"github.com/mandelsoft/spiff/yaml"
)

var mapType = reflect.TypeOf(map[string]interface{}{})
var arrayType = reflect.TypeOf([]interface{}{})

type CompileError struct {
	Path    []string
	Message error
}

func (c CompileError) Error() string {
	return fmt.Sprintf("%s: %s", strings.Join(c.Path, "."), c.Message)
}

type CompileErrors []CompileError

func (c CompileErrors) Error() string {
	if len(c) == 0 {
		return ""
	}
	s := ""
	for _, e := range c {
		s = s + "\n" + e.Error()
	}
	return s[1:]
}

// Len is the number of elements in the collection.
func (c CompileErrors) Len() int {
	return len(c)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (c CompileErrors) Less(i, j int) bool {
	return strings.Compare(c[i].Error(), c[j].Error()) < 0
}

// Swap swaps the elements with indexes i and j.
func (c CompileErrors) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c *CompileErrors) Add(path []string, err error) {
	if err != nil {
		*c = append(*c, CompileError{path, err})
	}
}

func (c *CompileErrors) Append(errs CompileErrors) {
	*c = append(*c, errs...)
}

func Compile(sourceName string, root interface{}) (yaml.Node, CompileErrors) {
	env := flow.NewEnvironment(nil, sourceName)
	return compile(env, root)
}

func compile(env dynaml.Binding, root interface{}) (yaml.Node, CompileErrors) {
	var errors CompileErrors

	switch rootVal := root.(type) {
	case time.Time:
		return yaml.NewNode(rootVal.Format("2019-01-08T10:06:26Z"), env.SourceName()), errors
	case string:
		_, err := flow.FlowString(yaml.NewNode(rootVal, env.SourceName()), env)
		errors.Add(env.Path(), err)
		return yaml.NewNode(rootVal, env.SourceName()), errors

	case map[interface{}]interface{}:
		sanitized := map[string]yaml.Node{}

		for key, val := range rootVal {
			str, ok := key.(string)
			if !ok {
				errors.Add(env.Path(), yaml.NonStringKeyError{key})
			} else {
				sub, errs := compile(env.WithPath(str), val)
				if errs == nil {
					sanitized[str] = sub
				}
				errors.Append(errs)
			}
		}

		return yaml.NewNode(sanitized, env.SourceName()), errors

	case []interface{}:
		sanitized := []yaml.Node{}

		for index, val := range rootVal {
			sub, errs := compile(env.WithPath(fmt.Sprintf("[%d]", index)), val)
			if errs == nil {
				sanitized = append(sanitized, sub)
			}
			errors.Append(errs)
		}

		return yaml.NewNode(sanitized, env.SourceName()), errors

	case map[string]interface{}:
		sanitized := map[string]yaml.Node{}

		for key, val := range rootVal {
			sub, errs := compile(env.WithPath(key), val)
			if errs == nil {
				sanitized[key] = sub
			}
			errors.Append(errs)
		}

		return yaml.NewNode(sanitized, env.SourceName()), errors
	case int:
		return yaml.NewNode(int64(rootVal), env.SourceName()), errors
	case int32:
		return yaml.NewNode(int64(rootVal), env.SourceName()), errors
	case float32:
		return yaml.NewNode(float64(rootVal), env.SourceName()), errors
	case []byte, int64, float64, bool, nil:
		return yaml.NewNode(rootVal, env.SourceName()), errors
	default:
		value := reflect.ValueOf(root)
		if value.Type().ConvertibleTo(mapType) {
			return compile(env, value.Convert(mapType).Interface())
		}
		if value.Type().ConvertibleTo(arrayType) {
			return compile(env, value.Convert(arrayType).Interface())
		}
		errors.Add(env.Path(), fmt.Errorf("unknown type (%s)", reflect.TypeOf(root).String()))
		return nil, errors
	}

}
