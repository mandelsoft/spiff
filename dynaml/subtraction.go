package dynaml

import (
	"fmt"
	"net"
)

type SubtractionExpr struct {
	A Expression
	B Expression
}

func (e SubtractionExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	resolved := true

	a, info, ok := ResolveExpressionOrPushEvaluation(&e.A, &resolved, nil, binding, false)
	if !ok {
		return nil, info, false
	}

	bint, info, ok := ResolveIntegerExpressionOrPushEvaluation(&e.B, &resolved, &info, binding, false)
	if !ok {
		return nil, info, false
	}

	if !resolved {
		return e, info, true
	}

	aint, ok := a.(int64)
	if ok {
		return aint - bint, info, true
	}

	str, ok := a.(string)
	if ok {
		ip := net.ParseIP(str)
		if ip != nil {
			return IPAdd(ip, -bint).String(), info, true
		}
		return info.Error("string argument for MINUS must be an IP address")
	}
	return info.Error("first argument of MINUS must be IP address or integer")
}

func (e SubtractionExpr) String() string {
	return fmt.Sprintf("%s - %s", e.A, e.B)
}
