package dynaml

import (
	"fmt"
	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/yaml"
	"time"
)

type SyncExpr struct {
	Sub Expression
	Cond Expression
	Value Expression
	Timeout Expression
	first  time.Time
	last time.Time
}

func (e SyncExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	resolved := true
	var value interface{}
	var info EvaluationInfo

	errmsg:=""
	timeout:=5 * time.Minute
	if e.Timeout!=nil {
		t, infot, ok := ResolveIntegerExpressionOrPushEvaluation(&e.Timeout,&resolved, nil, binding, false)
		if !ok {
			return nil, infot, ok
		}
		if !resolved {
			return e, info, true
		}
		timeout=time.Second *time.Duration(t)
	}

	result:= map[string]yaml.Node{}

	expr:=e.Sub
	value, infoe, ok := ResolveExpressionOrPushEvaluation(&expr, &resolved, nil, binding, false)

	if !ok {
		debug.Debug("sync arg failed\n")
		result[CATCH_VALID]=node(false, binding)
		result[CATCH_ERROR]=node(infoe.Issue.Issue, binding)
		errmsg=infoe.Issue.Issue
	} else {
		if !resolved {
			return e, info, true
		}
		result[CATCH_VALID]=node(true, binding)
		result[CATCH_ERROR]=node("", binding)
		result[CATCH_VALUE]=node(value, binding)
	}

	debug.Debug("sync arg succeeded\n")

	cond, infoc, ok := e.Cond.Evaluate(binding.WithLocalScope(result),false)
	if !ok {
		return info.AnnotateError(infoc,"condition evaluation failed)")
	}

	switch v:=cond.(type) {
	case bool:
		if !v {
			debug.Debug("sync condition is false\n")
			e.last=time.Now()

			if e.last.Before(e.first.Add(timeout)) {
				return e, infoe, ok
			}
			if errmsg!="" {
				return nil, infoe, false
			} else {
				return info.Error("sync timeout reached")
			}
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
		if isExpression(value) {
			return e, infov, true
		}
		return value, infov, ok
	} else {
		if errmsg!="" {
			return errmsg, info, true
		}
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
		return &SyncExpr{e.Arguments[0], e.Arguments[1], nil, nil, time.Now(), time.Time{}}, info, true
	case 3:
		return &SyncExpr{e.Arguments[0], e.Arguments[1], e.Arguments[2], nil, time.Now(), time.Time{}}, info, true
	case 4:
		return &SyncExpr{e.Arguments[0], e.Arguments[1], e.Arguments[2], e.Arguments[3], time.Now(), time.Time{}}, info, true
	default:
		return info.Error("2 or 3 arguments required for sync")
	}
}


