package dynaml

import (
	"strings"
)

func func_replace(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) < 3 {
		return info.Error("replace requires at least 3 arguments")
	}
	if len(arguments) > 4 {
		return info.Error("replace does not take more than 4 arguments")
	}

	str, ok := arguments[0].(string)
	if !ok {
		return info.Error("first argument for replace must be a string")
	}
	src, ok := arguments[1].(string)
	if !ok {
		return info.Error("second argument for replace must be a string")
	}
	dst, ok := arguments[2].(string)
	if !ok {
		return info.Error("third argument for replace must be a string")
	}
	n := int64(-1)
	if len(arguments) > 3 {
		n, ok = arguments[3].(int64)
		if !ok {
			return info.Error("fourth argument for replace must be an integer")
		}
	}

	e := strings.Replace(str, src, dst, int(n))
	return e, info, true
}
