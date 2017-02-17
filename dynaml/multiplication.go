package dynaml

import (
	"fmt"
	"net"
)

type MultiplicationExpr struct {
	A Expression
	B Expression
}

func (e MultiplicationExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
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
		return aint * bint, info, true
	}

	str, ok := a.(string)
	if ok {
		ip, cidr, err := net.ParseCIDR(str)
		if err != nil {
			return info.Error("CIDR or int argument required for multiplication: %s", err)
		}
		ones, _ := cidr.Mask.Size()
		size := int64(1 << (32 - uint32(ones)))
		ip = IPAdd(ip.Mask(cidr.Mask), size*bint)
		return (&net.IPNet{ip, cidr.Mask}).String(), info, true
	}
	return info.Error("CIDR or int argument required as first argument for multiplication")
}

func (e MultiplicationExpr) String() string {
	return fmt.Sprintf("%s * %s", e.A, e.B)
}
