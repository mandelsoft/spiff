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

	b, info, ok := ResolveExpressionOrPushEvaluation(&e.B, &resolved, &info, binding, false)
	if !ok {
		return nil, info, false
	}

	if !resolved {
		return e, info, true
	}

	aint, ok := a.(int64)
	bint, bok := b.(int64)
	if ok {
		if !bok {
			return info.Error("integer operand required")
		}
		return aint - bint, info, true
	}

	str, ok := a.(string)
	if ok {
		ip := net.ParseIP(str)
		if ip != nil {
			if bok {
				return IPAdd(ip, -bint).String(), info, true
			}
			bstr, ok := b.(string)
			if ok {
				ipb := net.ParseIP(bstr)
				if ip != nil {
					if len(ip) != len(ipb) {
						return info.Error("IP type mismatch")
					}
					return DiffIP(ip, ipb), info, true
				}
				return info.Error("string argument for MINUS must be an IP address")
			}
			return info.Error("second argument of MINUS must be IP address or integer")
		}
		return info.Error("string argument for MINUS must be an IP address")
	}
	return info.Error("first argument of MINUS must be IP address or integer")
}

func (e SubtractionExpr) String() string {
	return fmt.Sprintf("%s - %s", e.A, e.B)
}
