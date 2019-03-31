package dynaml

import (
	"fmt"
	"github.com/mandelsoft/spiff/yaml"
	"net"
	"strings"
)

type Validator func(value interface{}, binding Binding, args ...interface{}) (bool, string, string, error, bool)

var validators = map[string]Validator{}

func RegisterValidator(name string, f Validator) {
	validators[name] = f
}

func func_validate(arguments []interface{}, binding Binding) (bool, interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()
	if len(arguments) < 2 {
		info.Error("at least two arguments required for validate")
		return true, nil, info, false
	}

	value := arguments[0]

	for i, a := range arguments[1:] {
		r, _, f, err, valid := validate(value, NewNode(a, binding), binding)
		if err != nil {
			info.SetError("condition %d has problem: %s", i+1, err)
			return true, nil, info, false
		}
		if !valid {
			return false, nil, info, true
		}
		if !r {
			info.SetError("condition %d failed: %s", i+1, f)
			return true, nil, info, false
		}
	}
	return true, value, info, true
}

func validate(value interface{}, cond yaml.Node, binding Binding) (bool, string, string, error, bool) {
	if cond == nil || cond.Value() == nil {
		return true, "empty condition", "no empty condition", nil, true
	}
	switch v := cond.Value().(type) {
	case string:
		return _validate(value, v, binding)
	case LambdaValue:
		return _validate(value, v, binding)
	case []yaml.Node:
		if len(v) == 0 {
			return false, "", "", fmt.Errorf("validation type missing"), true
		}
		return _validate(value, v[0].Value(), binding, v[1:]...)
	default:
		return false, "", "", fmt.Errorf("invalid validation check type: %s", ExpressionType(v)), true
	}
}

func _validate(value interface{}, cond interface{}, binding Binding, args ...yaml.Node) (bool, string, string, error, bool) {
	var err error
	switch v := cond.(type) {
	case LambdaValue:
		if len(v.lambda.Names) != len(args)+1 {
			return false, "", "", fmt.Errorf("argument count mismatch for lambda %s: expected %d, found %d", v, len(v.lambda.Names), len(args)+1), true
		}
		vargs := []interface{}{value}
		for _, a := range args {
			vargs = append(vargs, a.Value())
		}
		valid, r, info, ok := v.Evaluate(false, false, vargs, binding, false)

		if !valid {
			if !ok {
				err = fmt.Errorf("%s", info.Issue.Issue)
			}
			return false, "", "", err, false
		}
		return toBool(r), fmt.Sprintf("%s succeeded", v), fmt.Sprintf("%s failed", v), nil, true
	case string:
		not := strings.HasPrefix(v, "!")
		if not {
			v = v[1:]
		}
		r, t, f, err, resolved := handleStringType(value, v, binding, args...)
		if !resolved || err != nil {
			return false, "", "", err, resolved
		}
		if not {
			return !r, f, t, err, resolved
		} else {
			return r, t, f, err, resolved
		}
	default:
		return false, "", "", fmt.Errorf("unexpected validation type %q", ExpressionType(v)), true
	}
}

func handleStringType(value interface{}, op string, binding Binding, args ...yaml.Node) (bool, string, string, error, bool) {
	reason := "("
	switch op {
	case "and":
		for _, c := range args {
			r, t, f, err, resolved := validate(value, c, binding)
			if err != nil || !resolved {
				return false, "", "", err, resolved
			}
			if reason != "(" {
				reason += " and "
			}
			reason += t
			if !r {
				return false, reason + ")", f, nil, true
			}
		}
		reason = reason + ")"
		return true, reason, reason, nil, true
	case "or":
		for _, c := range args {
			r, t, f, err, resolved := validate(value, c, binding)
			if err != nil || !resolved {
				return false, "", "", err, resolved
			}
			if reason != "(" {
				reason += " and "
			}
			reason += f
			if r {
				return true, t, reason + ")", nil, true
			}
		}
		reason = reason + ")"
		return false, reason, reason, nil, true
	case "empty":
		switch v := value.(type) {
		case string:
			return v == "", "is empty", "is not empty", nil, true
		case []yaml.Node:
			return len(v) == 0, "is empty", "is not empty", nil, true
		case map[string]yaml.Node:
			return len(v) == 0, "is empty", "is not empty", nil, true
		default:
			return false, "", "", fmt.Errorf("invalid type for empty: %s", ExpressionType(v)), true
		}
	case "type":
		e := ExpressionType(value)
		for _, t := range args {
			s, err := StringValue("type arg", t.Value())
			if err != nil {
				return false, "", "", err, true
			}
			if s == e {
				return true, fmt.Sprintf("is of type %s", s), fmt.Sprintf("is of type %s", s), nil, true
			}
			if reason != "(" {
				reason += " and "
			}
			reason += fmt.Sprintf("is not of type %s", s)
		}
		return false, reason + ")", reason + ")", nil, true
	case "dnsname":
		s, err := StringValue(op, value)
		if err != nil {
			return false, "", "", err, true
		}
		if err := IsWildcardDNS1123Subdomain(s); err != nil {
			if err := IsDNS1123Subdomain(s); err != nil {
				return false, "is dns name", fmt.Sprintf("is no dns name: %s", err), nil, true
			}
		}
		return true, "is dns name", "is no dns name", nil, true
	case "dnslabel":
		s, err := StringValue(op, value)
		if err != nil {
			return false, "", "", err, true
		}
		if err := IsDNS1123Label(s); err != nil {
			return false, "is dns label", fmt.Sprintf("is no dns label: %s", err), nil, true
		}
		return true, "is dns label", "is no dns label", nil, true
	case "dnsdomain":
		s, err := StringValue(op, value)
		if err != nil {
			return false, "", "", err, true
		}
		if err := IsDNS1123Subdomain(s); err != nil {
			return false, "is dns domain", fmt.Sprintf("is no dns domain: %s", err), nil, true
		}
		return true, "is dns domain", "is no dns domain", nil, true
	case "wildcarddnsdomain":
		s, err := StringValue(op, value)
		if err != nil {
			return false, "", "", err, true
		}
		if err := IsWildcardDNS1123Subdomain(s); err != nil {
			return false, "is wildcard dns domain", fmt.Sprintf("is no wildcard dns domain: %s", err), nil, true
		}
		return true, "is wildcard dns domain", "is no wildcard dns domain", nil, true
	case "ip":
		s, err := StringValue(op, value)
		if err != nil {
			return false, "", "", err, true
		}
		if ip := net.ParseIP(s); ip == nil {
			return false, "is ip address", fmt.Sprintf("is no ip address: %s", s), nil, true
		}
		return true, "is ip address", "is no ip address", nil, true
	case "cidr":
		s, err := StringValue(op, value)
		if err != nil {
			return false, "", "", err, true
		}
		if _, _, err := net.ParseCIDR(s); err != nil {
			return false, "is CIDR", fmt.Sprintf("is no CIDR: %s", err), nil, true
		}
		return true, "is CIDR", "is no CIDR", nil, true
	default:
		v := validators[op]
		if v != nil {
			vargs := []interface{}{}
			for _, a := range args {
				vargs = append(vargs, a.Value())
			}
			return v(value, binding, vargs...)
		}
		return false, "", "", fmt.Errorf("unknown validation operator %q", op), true
	}
}

func StringValue(msg string, v interface{}) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("%s requires string, but got %s", msg, ExpressionType(v))
	}
	return s, nil
}
