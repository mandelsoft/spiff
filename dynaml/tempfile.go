package dynaml

import (
	"io/ioutil"
	"os"
	"strconv"
)

func func_tempfile(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) < 1 || len(arguments) > 2 {
		return info.Error("temp_file requires exactly one or two arguments")
	}

	data, _ := getArg(0, arguments[0], true)

	name, err := binding.GetTempName([]byte(data))
	if err != nil {
		return info.Error("cannot create temporary file: %s", err)
	}

	permissions := int64(0644)
	if len(arguments) == 2 {
		switch v := arguments[1].(type) {
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

	err = ioutil.WriteFile(name, []byte(data), os.FileMode(permissions))
	if err != nil {
		return info.Error("cannot write file: %s", err)
	}

	return name, info, true
}
