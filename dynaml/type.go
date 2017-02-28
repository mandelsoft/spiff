package dynaml

import (
	"github.com/mandelsoft/spiff/yaml"
)

func func_type(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 1 {
		info.Error("exactly one argument required for function 'type'")
	}

	switch arguments[0].(type) {
	case string:
		return "string", info, true
	case int64:
		return "int", info, true
	case bool:
		return "bool", info, true
	case []yaml.Node:
		return "list", info, true
	case map[string]yaml.Node:
		return "map", info, true
	case TemplateValue:
		return "template", info, true
	case LambdaValue:
		return "lambda", info, true
	case nil:
		return "nil", info, true
	default:
		return info.Error("unknown type for %v", arguments[0])
	}
}
