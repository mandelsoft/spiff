package dynaml

import (
	"github.com/mandelsoft/spiff/yaml"
	"io/ioutil"
)

func func_listFiles(directory bool, arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 1 {
		return info.Error("list requires exactly one arguments")
	}

	name, ok := arguments[0].(string)
	if !ok {
		return info.Error("list: argument must be a string")
	}

	if name == "" {
		return info.Error("list: argument is empty string")
	}

	if !checkExistence(name, true) {
		return info.Error("list: %q is no directory or does not exist", name)
	}

	files, err := ioutil.ReadDir(name)
	if err != nil {
		return info.Error("list: %q:  error reading directory", name, err)
	}
	result := []yaml.Node{}
	for _, f := range files {
		if f.IsDir() == directory {
			result = append(result, NewNode(f.Name(), binding))
		}
	}
	return result, info, true
}
