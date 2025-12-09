package dynaml

import (
	"github.com/mandelsoft/spiff/yaml"
)

func func_deepmerge(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) < 1 {
		return nil, info, false
	}

	maps, msg := getMapList(arguments)
	if maps == nil {
		return info.Error("deepmerge: %s", msg)
	}
	result := make(map[string]yaml.Node)
	for _, v := range maps {
		deepmerge(result, v)
	}
	return result, info, true
}

func deepmerge(result map[string]yaml.Node, m map[string]yaml.Node) {
	for k, v := range m {
		if isMap(result[k]) && isMap(v) {
			r := make(map[string]yaml.Node)
			concatenateMap(r, result[k].Value().(map[string]yaml.Node))
			deepmerge(r, v.Value().(map[string]yaml.Node))
			v = yaml.NewNode(r, "<deepmerge>")
		}
		result[k] = v
	}
}
