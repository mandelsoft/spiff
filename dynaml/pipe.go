package dynaml

import (
	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/yaml"
)

func func_pipe(cached bool, arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) < 2 {
		return nil, info, false
	}
	args := []string{}
	debug.Debug("pipe: found %d arguments for call\n", len(arguments))
	for i, arg := range arguments {
		if i == 1 {
			list, ok := arg.([]yaml.Node)
			if ok {
				debug.Debug("exec: found array as second argument\n")
				if len(arguments) == 2 && len(list) > 0 {
					// handle single list argument to gain command and argument
					for j, arg := range list {
						v, ok := getArg(j+1, arg.Value())
						if !ok {
							return info.Error("command argument must be string")
						}
						args = append(args, v)
					}
				} else {
					return info.Error("list not allowed for command argument")
				}
			} else {
				str, ok := arg.(string)
				if !ok {
					return info.Error("command argument must be string")
				}
				args = append(args, str)
			}
		} else {
			v, ok := getArg(i-1, arg)
			if !ok {
				return info.Error("command argument/content must be string")
			}
			args = append(args, v)
		}
	}
	result, err := cachedExecute(cached, &args[0], args[1:])
	if err != nil {
		return info.Error("execution '%s' failed", args[1])
	}

	return convertOutput(result)
}
