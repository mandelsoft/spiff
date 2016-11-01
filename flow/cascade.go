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

		templates[i] = Cleanup(flowed, testLocal)
	}

	result, err := Flow(template, templates...)
	if err == nil {
		result = Cleanup(result, testTemporary)
	}
	return result, err
}

func testTemporary(node yaml.Node) bool {
	return node.Temporary() || node.Local()
}
func testLocal(node yaml.Node) bool {
	return node.Local()
}

func Cleanup(node yaml.Node, test func(yaml.Node) bool) yaml.Node {
	if node == nil {
		return nil
	}
	value := node.Value()
	switch v := value.(type) {
	case []yaml.Node:
		r := []yaml.Node{}
		for _, e := range v {
			if !test(e) {
				r = append(r, Cleanup(e, test))
			}
		}
		value = r

	case map[string]yaml.Node:
		r := map[string]yaml.Node{}
		for k, e := range v {
			if !test(e) {
				r[k] = Cleanup(e, test)
			}
		}
		value = r
	}
	return yaml.ReplaceValue(value, node)
}
