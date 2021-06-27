package control

import (
	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/yaml"
)

func init() {
	dynaml.RegisterControl("type", flowType, "default")
}

func flowType(ctx *dynaml.ControlContext) (yaml.Node, bool) {
	t := "undef"
	if ctx.Value.Value() != nil {
		switch v := ctx.Value.Value().(type) {
		case dynaml.Expression:
			_, info, _ := v.Evaluate(ctx, false)
			if !info.Undefined {
				return ctx.Node, false
			}
		default:
			sub := yaml.EmbeddedDynaml(ctx.Value, ctx.GetState().InterpolationEnabled())
			if sub != nil || !dynaml.IsResolvedNode(ctx.Value, ctx) {
				return ctx.Node, false
			}

			t = dynaml.ExpressionType(v)
		}
	} else {
		t = "nil"
	}

	return selected(ctx, t)
}
