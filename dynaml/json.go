package dynaml

import (
	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/mandelsoft/spiff/yaml"
	"strings"
)

func func_as_json(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 1 {
		return info.Error("asjson takes exactly one argument")
	}

	result, err := yaml.ValueToJSON(arguments[0])
	if err != nil {
		return info.Error("cannot jsonencode: %s", err)
	}
	return string(result), info, true
}

func func_as_yaml(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 1 {
		return info.Error("asyaml takes exactly one argument")
	}

	result, err := candiedyaml.Marshal(arguments[0])
	if err != nil {
		return info.Error("cannot yamlencode: %s", err)
	}
	return string(result), info, true
}

func func_parse_yaml(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 1 {
		return info.Error("parseyaml takes exactly one argument")
	}

	str, ok := arguments[0].(string)
	if !ok {
		return info.Error("first argument for parseyaml must be a string")
	}
	name := strings.Join(binding.Path(), ".")
	node, err := yaml.Parse(name, []byte(str))
	if err != nil {
		return info.Error("error parsing stub [%s]: %s", name, err)
	}

	return node.Value(), info, true
}
