package dynaml

import (
	"fmt"
	"net"
)

type DivisionExpr struct {
	A Expression
	B Expression
}

func (e DivisionExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
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

	if bint == 0 {
		return info.Error("division by zero")
	}

	aint, ok := a.(int64)
	if ok {
		return aint / bint, info, true
	}

	str, ok := a.(string)
	if ok {
		ip, cidr, err := net.ParseCIDR(str)
		if err != nil {
			return info.Error("CIDR or int argument required as first argument for division: %s", err)
		}
		ones, bits := cidr.Mask.Size()
		ip = ip.Mask(cidr.Mask)
		round := false
		for bint > 1 {
			if bint%2 == 1 {
				round = true
			}
			bint = bint / 2
			ones++
		}
		if round {
			ones++
		}
		if ones > 32 {
			return info.Error("divisor too large for CIDR network size")
		}
		return (&net.IPNet{ip, net.CIDRMask(ones, bits)}).String(), info, true
	}
	return info.Error("CIDR or int argument required as first argument for division")
}

func (e DivisionExpr) String() string {
	return fmt.Sprintf("%s / %s", e.A, e.B)
}
