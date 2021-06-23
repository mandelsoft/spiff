package control

import (
	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/yaml"
)

func init() {
	dynaml.RegisterControl("switch", flowSwitch, "default")
}

func flowSwitch(val yaml.Node, node yaml.Node, fields, opts map[string]yaml.Node, env dynaml.Binding) (yaml.Node, bool) {
	switch v := val.Value().(type) {
	case dynaml.Expression:
		return node, false
	case string:
		sub := yaml.EmbeddedDynaml(val, env.GetState().InterpolationEnabled())
		if sub != nil {
			return node, false
		}
		if s, ok := fields[v]; ok {
			return s, true
		}
		if s, ok := opts["default"]; ok {
			return s, true
		}
		return dynaml.ControlIssue("switch", node, "invalid switch value: %q", v)
	default:
		return dynaml.ControlIssue("switch", node, "invalid switch value type: %s", dynaml.ExpressionType(v))
	}
}
