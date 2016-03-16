package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func func_contains(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 2 {
		return info.Error("contains takes exactly two arguments")
	}

	list, ok := arguments[0].([]yaml.Node)
	if !ok {
		return info.Error("list expected for argument one of function contains required")
	}

	if arguments[1] == nil {
		return false, info, true
	}

	elem := arguments[1]

	for _, v := range list {
		r, _, _ := compareEquals(v.Value(), elem)
		if r {
			return true, info, true
		}
	}
	return false, info, true
}
