package dynaml

import (
	"fmt"
	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/yaml"
)

type SelectExpr struct {
	A      Expression
	Lambda Expression
}

func (e SelectExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	resolved := true

	debug.Debug("evaluate mapping\n")
	value, info, ok := ResolveExpressionOrPushEvaluation(&e.A, &resolved, nil, binding, true)
	if !ok {
		return nil, info, false
	}
	lvalue, infoe, ok := ResolveExpressionOrPushEvaluation(&e.Lambda, &resolved, nil, binding, false)
	if !ok {
		return nil, info, false
	}

	if !resolved {
		return e, info.Join(infoe), ok
	}

	lambda, ok := lvalue.(LambdaValue)
	if !ok {
		return infoe.Error("mapping requires a lambda value")
	}

	debug.Debug("select: using lambda %+v\n", lambda)
	var result []yaml.Node
	switch value.(type) {
	case []yaml.Node:
		result, info, ok = mapList(value.([]yaml.Node), lambda, binding, selectResult)

	case map[string]yaml.Node:
		result, info, ok = mapMap(value.(map[string]yaml.Node), lambda, binding, selectResult)

	default:
		return info.Error("map or list required for mapping")
	}
	if !ok {
		return nil, info, false
	}
	if result == nil {
		return e, info, true
	}
	debug.Debug("select: --> %+v\n", result)
	return result, info, true
}

func (e SelectExpr) String() string {
	lambda, ok := e.Lambda.(LambdaExpr)
	if ok {
		return fmt.Sprintf("select[%s%s]", e.A, fmt.Sprintf("%s", lambda)[len("lambda"):])
	} else {
		return fmt.Sprintf("select[%s|%s]", e.A, e.Lambda)
	}
}

func selectResult(mapped interface{}, n yaml.Node, info EvaluationInfo) yaml.Node {
	if mapped != nil && toBool(mapped) {
		return n
	}
	return nil
}
