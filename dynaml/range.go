package dynaml

import (
	"fmt"

	"github.com/mandelsoft/spiff/yaml"
)

type RangeExpr struct {
	Start Expression
	End   Expression
}

func (e RangeExpr) getRange(binding Binding) (int64, int64, EvaluationInfo, bool, bool) {
	resolved := true

	start, info, ok := ResolveIntegerExpressionOrPushEvaluation(&e.Start, &resolved, nil, binding, false)
	if !ok {
		return 0, 0, info, false, resolved
	}
	end, info, ok := ResolveIntegerExpressionOrPushEvaluation(&e.End, &resolved, &info, binding, false)
	if !ok {
		return 0, 0, info, false, resolved
	}
	return start, end, info, true, resolved
}

func (e RangeExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	start, end, info, ok, resolved := e.getRange(binding)

	if !ok {
		return nil, info, false
	}
	if !resolved {
		return e, info, true
	}

	nodes := []yaml.Node{}
	delta := int64(1)
	if start > end {
		delta = -1
	}
	for i := start; i*delta <= end*delta; i += delta {
		nodes = append(nodes, node(i, binding))
	}

	return nodes, info, true
}

func (e RangeExpr) String() string {
	return fmt.Sprintf("[%s..%s]", e.Start, e.End)
}
