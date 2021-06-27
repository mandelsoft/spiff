package flow

import (
	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/yaml"

	_ "github.com/mandelsoft/spiff/dynaml/control"
)

func flowControl(node yaml.Node, undef map[string]yaml.Node, env dynaml.Binding) (yaml.Node, bool, bool) {
	flags := node.GetAnnotation().Flags()
	resolved := false
	is := false
	if m, ok := node.Value().(map[string]yaml.Node); ok {
		control, val, fields, opts, err := dynaml.GetControl(m, undef, env)
		ctx := &dynaml.ControlContext{
			Control: control,
			Value:   val,
			Node:    node,
			Fields:  fields,
			Options: opts,
			Binding: env,
		}
		if control != nil {
			if err == nil {
				is = true
				node, resolved = control.Function(ctx)
			}
		}
		if err != nil {
			node, resolved = dynaml.ControlIssue(ctx, err.Error())
		}
	}
	if resolved {
		if flags != 0 {
			node = yaml.AddFlags(node, flags)
		}
	}
	return node, is, resolved
}
