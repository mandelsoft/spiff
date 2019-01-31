package dynaml

import (
	"io/ioutil"
	"os"
	"strconv"
)

func func_write(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	var err error
	info := DefaultInfo()

	if len(arguments) < 2 || len(arguments) > 3 {
		return info.Error("write requires two arguments")
	}
	file, ok := getArg(0, arguments[0])
	if !ok || file == "" {
		return info.Error("file argument must be a non-empty string")
	}
	data, _ := getArg(1, arguments[1])
	permissions := int64(0644)
	if len(arguments) == 3 {
		switch v := arguments[2].(type) {
		case string:
			permissions, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return info.Error("permissions must be given as int or int string: %s", err)
			}
		case int64:
			permissions = v
		default:
			return info.Error("permissions must be given as int or int string")
		}
	}

	err = ioutil.WriteFile(file, []byte(data), os.FileMode(permissions))
	if err != nil {
		return info.Error("cannot write file: %s", err)
	}

	return convertOutput([]byte(data))
}
