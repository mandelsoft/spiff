package dynaml

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/yaml"
)

type ComparisonExpr struct {
	A  Expression
	Op string
	B  Expression
}

func (e ComparisonExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
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

	var result bool
	var infor EvaluationInfo

	switch e.Op {
	case "==":
		result, infor, ok = compareEquals(a, b)
	case "!=":
		result, infor, ok = compareEquals(a, b)
		result = !result
	case "<=", "<", ">", ">=":
		switch va := a.(type) {
		case int64:
			vb, ok := b.(int64)
			if !ok {
				return infor.Error("comparision %s only for integers or strings", e.Op)
			}
			switch e.Op {
			case "<=":
				result = va <= vb
			case "<":
				result = va < vb
			case ">":
				result = va > vb
			case ">=":
				result = va >= vb
			}

		case string:
			vb, ok := b.(string)
			if !ok {
				return infor.Error("comparision %s only for strings or integers", e.Op)
			}
			switch e.Op {
			case "<=":
				result = va <= vb
			case "<":
				result = va < vb
			case ">":
				result = va > vb
			case ">=":
				result = va >= vb
			}
		}
	}
	infor = info.Join(infor)

	if !ok {
		return nil, infor, false
	}
	return result, infor, true
}

func (e ComparisonExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.A, e.Op, e.B)
}

func compareEquals(a, b interface{}) (bool, EvaluationInfo, bool) {
	info := DefaultInfo()

	debug.Debug("compare a '%#v'\n", a)
	debug.Debug("compare b '%#v' \n", b)
	if a == nil && b == nil {
		return true, info, true
	}
	switch va := a.(type) {
	case string:
		var vb string
		switch v := b.(type) {
		case string:
			vb = v
		case int64:
			vb = strconv.FormatInt(v, 10)
		case LambdaValue:
			vb = v.String()
		case bool:
			vb = strconv.FormatBool(v)
		default:
			info.Issue = yaml.NewIssue("types uncomparable")
			return false, info, false
		}
		return va == vb, info, true

	case int64:
		var vb int64
		var err error
		switch v := b.(type) {
		case string:
			vb, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return false, info, true
			}
		case int64:
			vb = v
		case bool:
			if v {
				vb = 1
			} else {
				vb = 0
			}
		default:
			info.Issue = yaml.NewIssue("types uncomparable")
			return false, info, false
		}
		return va == vb, info, true

	case LambdaValue:
		switch v := b.(type) {
		case string:
			return va.String() == v, info, true
		case LambdaValue:
			return reflect.DeepEqual(va, v), info, true
		default:
			info.Issue = yaml.NewIssue("types uncomparable")
			return false, info, false
		}

	case []yaml.Node:
		vb, ok := b.([]yaml.Node)
		if !ok || len(va) != len(vb) {
			debug.Debug("compare list len mismatch")
			break
		}
		for i, v := range vb {
			result, info, _ := compareEquals(va[i].Value(), v.Value())
			if !result {
				debug.Debug(fmt.Sprintf("compare list entry %d mismatch", i))
				return false, info, true
			}
		}
		return true, info, true

	case map[string]yaml.Node:
		vb, ok := b.(map[string]yaml.Node)
		if !ok || len(va) != len(vb) {
			break
		}

		for k, v := range vb {
			result, info, _ := compareEquals(va[k].Value(), v.Value())
			if !result {
				return false, info, true
			}
		}
		return true, info, true

	}

	return false, info, true
}
