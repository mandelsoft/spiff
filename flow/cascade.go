package flow

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func Cascade(template yaml.Node, partial bool, templates ...yaml.Node) (yaml.Node, error) {
	for i := len(templates) - 1; i >= 0; i-- {
		flowed, err := Flow(templates[i], templates[i+1:]...)
		if !partial && err != nil {
			return nil, err
		}

		templates[i] = flowed
	}

	result, err := Flow(template, templates...)
	if err == nil {
		result = Cleanup(result)
	}
	return result, err
}

func Cleanup(node yaml.Node) yaml.Node {
	if node == nil {
		return nil
	}
	value := node.Value()
	switch v := value.(type) {
	case []yaml.Node:
		r := []yaml.Node{}
		for _, e := range v {
			if !e.Temporary() {
				r = append(r, Cleanup(e))
			}
		}
		value = r

	case map[string]yaml.Node:
		r := map[string]yaml.Node{}
		for k, e := range v {
			if !e.Temporary() {
				r[k] = Cleanup(e)
			}
		}
		value = r
	}
	return yaml.ReplaceValue(value, node)
}
