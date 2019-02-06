package dynaml

import (
	"fmt"
	"sort"

	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/yaml"
)

type MapExpr struct {
	A      Expression
	Lambda Expression
}

func (e MapExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
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

	debug.Debug("map: using lambda %+v\n", lambda)
	var result []yaml.Node
	switch value.(type) {
	case []yaml.Node:
		result, info, ok = mapList(value.([]yaml.Node), lambda, binding, mapResult)

	case map[string]yaml.Node:
		result, info, ok = mapMap(value.(map[string]yaml.Node), lambda, binding, mapResult)

	default:
		return info.Error("map or list required for mapping")
	}
	if !ok {
		return nil, info, false
	}
	if result == nil {
		return e, info, true
	}
	debug.Debug("map: --> %+v\n", result)
	return result, info, true
}

func (e MapExpr) String() string {
	lambda, ok := e.Lambda.(LambdaExpr)
	if ok {
		return fmt.Sprintf("map[%s%s]", e.A, fmt.Sprintf("%s", lambda)[len("lambda"):])
	} else {
		return fmt.Sprintf("map[%s|%s]", e.A, e.Lambda)
	}
}

type MappingResult func(mapped interface{}, n yaml.Node, info EvaluationInfo) yaml.Node

func mapResult(mapped interface{}, n yaml.Node, info EvaluationInfo) yaml.Node {
	if mapped != nil {
		return node(mapped, info)
	}
	return nil
}

func mapList(source []yaml.Node, e LambdaValue, binding Binding, mapper MappingResult) ([]yaml.Node, EvaluationInfo, bool) {
	inp := make([]interface{}, len(e.lambda.Names))
	result := []yaml.Node{}
	info := DefaultInfo()

	if len(e.lambda.Names) > 2 {
		info.Error("mapping expression take a maximum of 2 arguments")
		return nil, info, false
	}
	for i, n := range source {
		debug.Debug("map:  mapping for %d: %+v\n", i, n)
		inp[0] = i
		inp[len(inp)-1] = n.Value()
		mapped, info, ok := e.Evaluate(inp, binding, false)
		if !ok {
			debug.Debug("map:  %d %+v: failed\n", i, n)
			return nil, info, false
		}

		_, ok = mapped.(Expression)
		if ok {
			debug.Debug("map:  %d unresolved  -> KEEP\n")
			return nil, info, true
		}
		debug.Debug("map:  %d --> %+v\n", i, mapped)
		node := mapper(mapped, n, info)
		if node != nil {
			result = append(result, node)
		}
	}
	return result, info, true
}

func mapMap(source map[string]yaml.Node, e LambdaValue, binding Binding, mapper MappingResult) ([]yaml.Node, EvaluationInfo, bool) {
	inp := make([]interface{}, len(e.lambda.Names))
	result := []yaml.Node{}
	info := DefaultInfo()

	keys := getSortedKeys(source)
	for _, k := range keys {
		n := source[k]
		debug.Debug("map:  mapping for %s: %+v\n", k, n)
		inp[0] = k
		inp[len(inp)-1] = n.Value()
		mapped, info, ok := e.Evaluate(inp, binding, false)
		if !ok {
			debug.Debug("map:  %s %+v: failed\n", k, n)
			return nil, info, false
		}

		_, ok = mapped.(Expression)
		if ok {
			debug.Debug("map:  %d unresolved  -> KEEP\n")
			return nil, info, true
		}
		debug.Debug("map:  %s --> %+v\n", k, mapped)
		node := mapper(mapped, n, info)
		if node != nil {
			result = append(result, node)
		}
	}
	return result, info, true
}

func getSortedKeys(unsortedMap map[string]yaml.Node) []string {
	keys := make([]string, len(unsortedMap))
	i := 0
	for k, _ := range unsortedMap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
