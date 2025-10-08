package dynaml

import (
	"fmt"

	"github.com/mandelsoft/spiff/yaml"
)

func func_merge(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) < 1 {
		return info.Error("at least one argument required for merge function")
	}

	maps, msg := getMapList(arguments)
	if maps == nil {
		return info.Error("merge: %s", msg)
	}

	args := make([]yaml.Node, len(maps))
	for i, m := range maps {
		args[i] = yaml.NewNode(m, "dynaml")
	}
	result, err := binding.Cascade(binding, args[0], false, args[1:]...)
	if err != nil {
		info.SetError("merging failed: %s", err)
		return nil, info, false
	}

	return result.Value(), info, true
}

func getMap(n int, arg interface{}) (map[string]yaml.Node, error) {
	temp, ok := arg.(TemplateValue)
	if ok {
		arg, ok := node_copy(temp.Prepared).Value().(map[string]yaml.Node)
		if !ok {
			return nil, fmt.Errorf("%d: template is not a map template", n+1)
		}
		return arg, nil
	}
	m, ok := arg.(map[string]yaml.Node)
	if ok {
		return m, nil
	}
	return nil, fmt.Errorf("%d: no map or map template, but %s", n+1, ExpressionType(arg))
}

func getMapList(arguments []interface{}) ([]map[string]yaml.Node, string) {
	args := []map[string]yaml.Node{}

	if len(arguments) == 1 {
		l, ok := arguments[0].([]yaml.Node)
		if ok {
			for i, e := range l {
				m, err := getMap(i, e.Value())
				if err != nil {
					return nil, fmt.Sprintf("entry of list argument: %s", err)
				}
				args = append(args, m)
			}
			if len(args) == 0 {
				return nil, "no map found for merge"
			}
		}
	}
	if len(args) == 0 {
		for i, arg := range arguments {
			m, err := getMap(i, arg)
			if err != nil {
				return nil, fmt.Sprintf("argument %s", err)
			}
			args = append(args, m)
		}
	}
	return args, ""
}
