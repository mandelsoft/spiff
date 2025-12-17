package dynaml

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/spiff/legacy/candiedyaml"
	"github.com/mandelsoft/spiff/yaml"

	"github.com/gowebpki/jcs"
)

func func_as_json(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	var err error

	info := DefaultInfo()

	canon := false
	switch len(arguments) {
	case 1:
	case 2:
		canon = toBool(arguments[1])
	default:
		if len(arguments) != 1 {
			return info.Error("asjson takes at least one argument, and optionally a second one")
		}
	}

	result, err := yaml.ValueToJSON(arguments[0])
	if err != nil {
		return info.Error("cannot jsonencode: %s", err)
	}
	if canon {
		result, err = CanonicalizedJson(arguments[0])
		if err != nil {
			return info.Error("%s", err)
		}
	}
	return string(result), info, true
}

func func_as_jcs(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 1 {
		return info.Error("asjson takes exactly one argument")
	}

	result, err := CanonicalizedJson(arguments[0])
	if err != nil {
		return info.Error("%s", err)
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

	if len(arguments) < 1 || len(arguments) > 2 {
		return info.Error("parse takes one or two arguments")
	}

	str, ok := arguments[0].(string)
	if !ok {
		return info.Error("first argument for parse must be a string")
	}

	mode := "import"
	if len(arguments) > 1 {
		mode, ok = arguments[1].(string)
		if !ok {
			return info.Error("second argument for parse must be a string")
		}
	}

	name := strings.Join(binding.Path(), ".")
	return ParseData(name, []byte(str), mode, binding)
}

func CanonicalizedJson(node interface{}) ([]byte, error) {
	result, err := yaml.ValueToJSON(node)
	if err != nil {
		return nil, fmt.Errorf("cannot jsonencode: %w", err)
	}
	result, err = jcs.Transform(result)
	if err != nil {
		return nil, fmt.Errorf("cannot canonicalize: %w", err)
	}

	return result, nil
}
