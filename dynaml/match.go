package dynaml

import (
	"regexp"
	"strconv"

	"github.com/mandelsoft/spiff/yaml"
)

func func_match(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 2 {
		return info.Error("match takes exactly two arguments")
	}

	pattern, ok := arguments[0].(string)
	if !ok {
		return info.Error("pattern string for argument one of function match required")
	}

	if arguments[1] == nil {
		return false, info, true
	}

	elem := ""
	switch v := arguments[1].(type) {
	case string:
		elem = v
	case int64:
		elem = strconv.FormatInt(v, 10)
	case bool:
		elem = strconv.FormatBool(v)
	default:
		return info.Error("simple value for argument two of function match required")
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return info.Error("match: %s", err)
	}

	list := re.FindStringSubmatch(elem)
	newList := make([]yaml.Node, len(list))
	for i, v := range list {
		newList[i] = node(v, info)
	}
	return newList, info, true
}
