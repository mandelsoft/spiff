package dynaml

import (
	"github.com/mandelsoft/spiff/yaml"
	"os"
	"path/filepath"
)

func func_lookup(directory bool, arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	paths := []string{}

	switch len(arguments) {
	case 0, 1:
		return info.Error("lookup_file requires at least two arguments")
	default:
		for index, arg := range arguments[1:] {
			switch v := arg.(type) {
			case []yaml.Node:
				for _, p := range v {
					if p.Value() == nil {
						continue
					}
					switch v := p.Value().(type) {
					case string:
						paths = append(paths, v)
					default:
						return info.Error("lookup_file: argument %d must be a list of strings", index)
					}
				}
			case string:
				paths = append(paths, v)
			default:
				return info.Error("lookup_file: argument %d must be a string or a list of strings", index)
			}
		}
	}

	name, ok := arguments[0].(string)
	if !ok {
		return info.Error("lookup_file: first argument must be a string")
	}

	if name == "" {
		return info.Error("lookup_file: first argument is empty string")
	}

	result := []yaml.Node{}
	if filepath.IsAbs(name) {
		if checkExistence(name, directory) {
			result = append(result, NewNode(name, binding))
		}
		return result, info, true
	}

	for _, d := range paths {
		if d != "" {
			p := d + "/" + name
			if checkExistence(p, directory) {
				result = append(result, NewNode(p, binding))
			}
		}
	}
	return result, info, true
}

func checkExistence(path string, directory bool) bool {
	s, err := os.Stat(path)
	if os.IsNotExist(err) || err != nil {
		return false
	}
	return s.IsDir() == directory
}
