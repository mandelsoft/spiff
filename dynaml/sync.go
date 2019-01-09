package dynaml

import (
	"fmt"
	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/yaml"
)

type SyncExpr struct {
	Sub Expression
	Cond Expression
	Value Expression
}

func (e SyncExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	resolved := true
	var value interface{}
	var info EvaluationInfo

	result:= map[string]yaml.Node{}

	value, infoe, ok := ResolveExpressionOrPushEvaluation(&e.Sub, &resolved, nil, binding, false)

	if !ok {
		debug.Debug("sync arg failed\n")
		result[CATCH_VALID]=node(false, binding)
		result[CATCH_ERROR]=node(infoe.Issue.Issue, binding)
	} else {
		if !resolved {
			return e, info, true
		}
	}

	debug.Debug("sync arg succeeded\n")
	result[CATCH_VALID]=node(true, binding)
	result[CATCH_ERROR]=node("", binding)
	result[CATCH_VALUE]=node(value, binding)

	cond, infoc, ok := e.Cond.Evaluate(binding.WithLocalScope(result),false)
	if !ok {
		return info.AnnotateError(infoc,"condition evaluation failed)")
	}

	switch v:=cond.(type) {
	case bool:
		if !v {
			debug.Debug("sync condition is false\n")
			return e, infoe, ok
		}
	case Expression:
		return e, info, true
	default:
		return info.Error("condition must evaluate to bool")
	}

	if e.Value!=nil  {
		debug.Debug("evaluating sync value\n")
		value, infov, ok := e.Value.Evaluate(binding.WithLocalScope(result),false)
		if !ok {
			return info.AnnotateError(infoc,"value expression failed)")
		}
		if _, ok := value.(Expression); ok {
			return e, infov, true
		}
		return value, infov, ok
	}
	debug.Debug("returning sync value\n")
	return value, infoe, ok
}

func (e SyncExpr) String() string {
	return fmt.Sprintf("sync(%s)", e.Sub)
}

func (e CallExpr) sync(binding Binding) (interface{}, EvaluationInfo, bool) {
	var info EvaluationInfo
	switch len(e.Arguments) {
	case 2:
		return &SyncExpr{e.Arguments[0], e.Arguments[1], nil}, info, true
	case 3:
		return &SyncExpr{e.Arguments[0], e.Arguments[1], e.Arguments[2]}, info, true
	default:
		return info.Error("2 or 3 arguments required for sync")
	}
}


