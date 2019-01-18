package dynaml

import (
	"crypto/md5"
	"fmt"
)

func func_md5(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 1 {
		return info.Error("md5 takes exactly one arguments")
	}

	str, ok := arguments[0].(string)
	if !ok {
		return info.Error("first argument for md5 must be a string")
	}

	result := md5.Sum([]byte(str))
	return fmt.Sprintf("%x", result), info, true
}
