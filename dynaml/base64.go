package dynaml

import (
	"encoding/base64"
)

func func_base64(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 1 {
		return info.Error("base64 takes exactly one argument")
	}

	str, ok := arguments[0].(string)
	if !ok {
		return info.Error("first argument for base64 must be a string")
	}

	result := base64.StdEncoding.EncodeToString([]byte(str))
	return result, info, true
}

func func_base64_decode(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 1 {
		return info.Error("base64_decode takes exactly one argument")
	}

	str, ok := arguments[0].(string)
	if !ok {
		return info.Error("first argument for base64_decode must be a string")
	}

	result, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return info.Error("cannot decode string")
	}
	return string(result), info, true
}
