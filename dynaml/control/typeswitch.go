package control

import (
	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/yaml"
)

func init() {
	dynaml.RegisterControl("type", flowType, "default")
}

func flowType(val yaml.Node, node yaml.Node, fields, opts map[string]yaml.Node, env dynaml.Binding) (yaml.Node, bool) {
	t := "undef"
	switch v := val.Value().(type) {
	case dynaml.Expression:
		_, info, _ := v.Evaluate(env, false)
		if !info.Undefined {
			return node, false
		}
	default:
		sub := yaml.EmbeddedDynaml(val, env.GetState().InterpolationEnabled())
		if sub != nil || !dynaml.IsResolvedNode(val, env) {
			return node, false
		}

		t = dynaml.ExpressionType(v)
	}

	if s, ok := fields[t]; ok {
		return s, true
	}
	if s, ok := opts["default"]; ok {
		return s, true
	}
	return dynaml.ControlIssue("type", node, "invalid type switch type: %q", t)
}
