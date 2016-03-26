package dynaml

import (
	"strings"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func func_split(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 2 {
		return info.Error("split takes exactly 2 arguments")
	}

	sep, ok := arguments[0].(string)
	if !ok {
		return info.Error("first argument for split must be a string")
	}
	str, ok := arguments[1].(string)
	if !ok {
		return info.Error("second argument for split must be a string")
	}

	array := strings.Split(str, sep)
	result := make([]yaml.Node, len(array))
	for i, e := range array {
		result[i] = node(e, binding)
	}
	return result, info, true
}
