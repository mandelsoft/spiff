package flow

import (
	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/yaml"

	_ "github.com/mandelsoft/spiff/dynaml/control"
)

func flowControl(node yaml.Node, env dynaml.Binding) (yaml.Node, bool, bool) {
	flags := node.GetAnnotation().Flags()
	resolved := false
	is := false
	if m, ok := node.Value().(map[string]yaml.Node); ok {
		control, val, fields, opts, err := dynaml.GetControl(m, env)
		if control != nil {
			if err == nil {
				is = true
				node, resolved = control.Function(val, node, fields, opts, env)
			}
		}
		if err != nil {
			node, resolved = dynaml.ControlIssue("", node, err.Error())
		}
	}
	if resolved {
		if flags != 0 {
			node = yaml.AddFlags(node, flags)
		}
	}
	return node, is, resolved
}
