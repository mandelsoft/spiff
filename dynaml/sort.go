package dynaml

import (
	"github.com/mandelsoft/spiff/yaml"
	"sort"
)

func func_sort(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 1 {
		return info.Error("sort takes exactly one argument")
	}

	list, ok := arguments[0].([]yaml.Node)
	if !ok {
		return info.Error("argument for sort must be a list")
	}

	for i := range list {
		if _, ok := list[i].Value().(string); !ok {
			return info.Error("list elements must be strings")
		}
	}

	less := func(i, j int) bool {
		a, _ := list[i].Value().(string)
		b, _ := list[j].Value().(string)
		return a < b
	}

	sort.SliceStable(list, less)
	return list, info, true
}
