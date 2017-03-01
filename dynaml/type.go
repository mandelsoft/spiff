package dynaml

import (
	"github.com/mandelsoft/spiff/yaml"
)

func func_type(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 1 {
		info.Error("exactly one argument required for function 'type'")
	}

	tn := expression_type(arguments[0])
	if tn == "" {
		return info.Error("unknown type for %v", arguments[0])
	} else {
		return tn, info, true
	}
}

func expression_type(elem interface{}) string {
	switch elem.(type) {
	case string:
		return "string"
	case int64:
		return "int"
	case bool:
		return "bool"
	case []yaml.Node:
		return "list"
	case map[string]yaml.Node:
		return "map"
	case TemplateValue:
		return "template"
	case LambdaValue:
		return "lambda"
	case nil:
		return "nil"
	default:
		return ""
	}
}
