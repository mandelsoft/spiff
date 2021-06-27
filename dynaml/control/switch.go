package control

import (
	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/yaml"
)

func init() {
	dynaml.RegisterControl("switch", flowSwitch, "default")
}

func flowSwitch(ctx *dynaml.ControlContext) (yaml.Node, bool) {
	if e, ok := ctx.Value.Value().(dynaml.Expression); ok {
		_, info, _ := e.Evaluate(ctx, false)
		if info.Undefined {
			return yaml.UndefinedNode(dynaml.NewNode(nil, ctx)), true
		}
	}
	if node, ok := dynaml.ControlReady(ctx, true); !ok {
		return node, false
	}
	return selected(ctx, ctx.Value.Value())
}

func selected(ctx *dynaml.ControlContext, key interface{}) (yaml.Node, bool) {
	var result yaml.Node
	if key != nil {
		switch v := key.(type) {
		case string:
			result = ctx.Field(v)
		default:
			return dynaml.ControlIssue(ctx, "invalid switch value type: %s", dynaml.ExpressionType(v))
		}
	}
	if result == nil {
		result = ctx.Option("default")
	}
	if result != nil {
		return dynaml.ControlValue(ctx, result)
	}
	return dynaml.ControlIssue(ctx, "invalid switch value: %q", key)
}
