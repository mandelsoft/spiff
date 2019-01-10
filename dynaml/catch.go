package dynaml

import (
	"fmt"
	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/yaml"
)

const CATCH_ERROR = "error"
const CATCH_VALUE = "value"
const CATCH_VALID = "valid"

type CatchExpr struct {
	Sub Expression
}

func (e CatchExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	resolved := true
	var value interface{}
	var info EvaluationInfo

	result := map[string]yaml.Node{}

	value, infoe, ok := ResolveExpressionOrPushEvaluation(&e.Sub, &resolved, nil, binding, false)

	if !ok {
		debug.Debug("catch arg failed\n")
		result[CATCH_VALID] = node(false, binding)
		result[CATCH_ERROR] = node(infoe.Issue.Issue, binding)
		return result, info, true
	}

	if !resolved {
		return e, info, true
	}

	debug.Debug("catch arg succeeded\n")
	result[CATCH_VALID] = node(true, binding)
	result[CATCH_ERROR] = node("", binding)
	result[CATCH_VALUE] = node(value, binding)
	return result, info, ok
}

func (e CatchExpr) String() string {
	return fmt.Sprintf("catch(%s)", e.Sub)
}

func (e CallExpr) catch(binding Binding) (interface{}, EvaluationInfo, bool) {
	var info EvaluationInfo
	if len(e.Arguments) != 1 {
		return info.Error("catch requires a single argument")
	}
	return &CatchExpr{e.Arguments[0]}, info, true
}
