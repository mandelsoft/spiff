package control

import (
	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/yaml"
)

func init() {
	dynaml.RegisterControl("if", flowIf, "then", "else")
}

func flowIf(val yaml.Node, node yaml.Node, fields, opts map[string]yaml.Node, env dynaml.Binding) (yaml.Node, bool) {
	for range fields {
		return dynaml.ControlIssue("if", node, "no regular fields %v allowed in if control", yaml.GetSortedKeys(fields))
	}
	switch v := val.Value().(type) {
	case dynaml.Expression:
		return node, false
	case bool:
		if v {
			if e, ok := opts["then"]; ok {
				return e, true
			}
			return yaml.UndefinedNode(yaml.NewNode(nil, node.SourceName())), true
		} else {
			if e, ok := opts["else"]; ok {
				return e, true
			}
			return yaml.UndefinedNode(yaml.NewNode(nil, node.SourceName())), true
		}
	default:
		sub := yaml.EmbeddedDynaml(val, env.GetState().InterpolationEnabled())
		if sub != nil || !dynaml.IsResolvedNode(val, env) {
			return node, false
		}
		return dynaml.ControlIssue("if", node, "invalid condition value type: %s", dynaml.ExpressionType(v))
	}
}
