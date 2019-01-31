package dynaml

import (
	"github.com/mandelsoft/spiff/yaml"
)

func func_merge(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) < 1 {
		return info.Error("at least one argument required for merge function")
	}

	args := make([]yaml.Node, len(arguments))

	for i, arg := range arguments {
		temp, ok := arg.(TemplateValue)
		if ok {
			arg = node_copy(temp.Prepared).Value()
		}
		m, ok := arg.(map[string]yaml.Node)
		if !ok {
			return info.Error("argument %d for merge function is no map or map template", i+1)
		}
		args[i] = yaml.NewNode(m, "dynaml")
	}
	result, err := binding.Cascade(binding, args[0], false, args[1:]...)
	if err != nil {
		info.SetError("merging failed: %s", err)
		return nil, info, false
	}

	return result.Value(), info, true
}
