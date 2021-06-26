package control

import (
	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/yaml"
)

func init() {
	dynaml.RegisterControl("merge", flowMerge)
}

func flowMerge(val yaml.Node, node yaml.Node, fields, opts map[string]yaml.Node, env dynaml.Binding) (yaml.Node, bool) {
	switch v := val.Value().(type) {
	case dynaml.Expression:
		return node, false
	case string:
		sub := yaml.EmbeddedDynaml(val, env.GetState().InterpolationEnabled())
		if sub != nil {
			return node, false
		}
		return dynaml.ControlIssue("merge", node, "invalid value type: %s", dynaml.ExpressionType(v))
	case map[string]yaml.Node:
		if !dynaml.IsResolvedNode(val, env) {
			return node, false
		}
		for k, e := range v {
			fields[k] = e
		}
	case []yaml.Node:
		if !dynaml.IsResolvedNode(val, env) {
			return node, false
		}
		for i, l := range v {
			if l.Value() != nil {
				if m, ok := l.Value().(map[string]yaml.Node); ok {
					for k, e := range m {
						fields[k] = e
					}
				} else {
					return dynaml.ControlIssue("merge", node, "entry %d: invalid entry type: %s", i, dynaml.ExpressionType(v))
				}
			}
		}

	default:
		if v != nil {
			return dynaml.ControlIssue("merge", node, "invalid value type: %s", dynaml.ExpressionType(v))
		}
	}
	return yaml.NewNode(fields, env.SourceName()), true
}
