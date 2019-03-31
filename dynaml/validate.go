package dynaml

import (
	"fmt"
	"github.com/mandelsoft/spiff/yaml"
	"net"
	"regexp"
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

	for i, c := range arguments[1:] {
		r, _, f, err, valid := validate(value, NewNode(c, binding), binding)
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
			return ValidatorErrorf("validation type missing")
		}
		return _validate(value, v[0].Value(), binding, v[1:]...)
	default:
		return ValidatorErrorf("invalid validation check type: %s", ExpressionType(v))
	}
}

func _validate(value interface{}, cond interface{}, binding Binding, args ...yaml.Node) (bool, string, string, error, bool) {
	var err error
	switch v := cond.(type) {
	case LambdaValue:
		if len(v.lambda.Names) != len(args)+1 {
			return ValidatorErrorf("argument count mismatch for lambda %s: expected %d, found %d", v, len(v.lambda.Names), len(args)+1)
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
		l, ok := r.([]yaml.Node)
		if ok {
			switch len(l) {
			case 1:
				r = l[0].Value()
				break
			case 2:
				t, err := StringValue("lambda validator", l[1].Value())
				if err != nil {
					return ValidatorErrorf("lambda validator result index %d: %s", 1, err)
				}
				return toBool(l[0].Value()), t, t, nil, true
			case 3:
				t, err := StringValue("lambda validator", l[1].Value())
				if err != nil {
					return ValidatorErrorf("lambda validator result index %d: %s", 1, err)
				}
				f, err := StringValue("lambda validator", l[2].Value())
				if err != nil {
					return ValidatorErrorf("lambda validator result index %d: %s", 2, err)
				}
				return toBool(l[0].Value()), t, f, nil, true
			default:
				return ValidatorErrorf("invalid result length of validator %s", v)
			}
		}
		return toBool(r), fmt.Sprintf("%s succeeded", v), fmt.Sprintf("%s failed", v), nil, true
	case string:
		not := strings.HasPrefix(v, "!")
		if not {
			v = v[1:]
		} else {
			if v == "" {
				return ValidatorErrorf("empty validator type")
			}
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
		return ValidatorErrorf("unexpected validation type %q", ExpressionType(v))
	}
}

func handleStringType(value interface{}, op string, binding Binding, args ...yaml.Node) (bool, string, string, error, bool) {
	reason := "("
	optional := false
	switch op {
	case "list":
		l, ok := value.([]yaml.Node)
		if !ok {
			return false, "is a list", "is no list", nil, true
		}
		if len(args) == 0 {
			return true, "is a list", "is no list", nil, true
		}
		for i, e := range l {
			for j, c := range args {
				r, t, f, err, valid := validate(e.Value(), c, binding)
				if err != nil {
					return ValidatorErrorf("list entry %d condition %d: %s", i, j, err)
				}
				if !valid {
					return false, "", "", nil, false
				}
				if !r {
					return false, fmt.Sprintf("entry %d condition %d %s", i, j, t), fmt.Sprintf("entry %d condition %d %s", i, j, f), nil, true
				}
			}
		}
		return true, "all entries match all conditions", "all entries match all conditions", nil, true

	case "map":
		l, ok := value.(map[string]yaml.Node)
		if !ok {
			return false, "is a map", "is no map", nil, true
		}
		if len(args) == 0 {
			return true, "is a map", "is no map", nil, true
		}
		var ck yaml.Node
		if len(args) > 2 {
			return ValidatorErrorf("map validator takes a maximum of two arguments, got %d", len(args))
		}
		if len(args) == 2 {
			ck = args[0]
		}
		ce := args[len(args)-1]

		for k, e := range l {
			if ck != nil {
				r, t, f, err, valid := validate(k, ck, binding)
				if err != nil {
					return ValidatorErrorf("map key %q %s", k, err)
				}
				if !valid {
					return false, "", "", nil, false
				}
				if !r {
					return false, fmt.Sprintf("map key %q %s", k, t), fmt.Sprintf("map key %q %s", k, f), nil, true
				}
			}

			r, t, f, err, valid := validate(e.Value(), ce, binding)
			if err != nil {
				return ValidatorErrorf("map entry %q: %s", k, err)
			}
			if !valid {
				return false, "", "", nil, false
			}
			if !r {
				return false, fmt.Sprintf("map entry %q %s", k, t), fmt.Sprintf("map entry %q %s", k, f), nil, true
			}
		}
		return true, "all map entries and keys match", "all map entries and keys match", nil, true

	case "optionalfield":
		optional = true
		fallthrough
	case "mapfield":
		l, ok := value.(map[string]yaml.Node)
		if !ok {
			return false, "is a map", "is no map", nil, true
		}
		if len(args) == 0 || len(args) > 2 {
			return ValidatorErrorf("%s reqires one or two arguments", op)
		}
		field, err := StringValue(op, args[0].Value())
		if err != nil {
			return ValidatorErrorf("field name must be string")
		}
		val, ok := l[field]
		if !ok {
			if optional {
				return true, fmt.Sprintf("has no optional field %q", field), "oops", nil, true
			}
			return false, fmt.Sprintf("has field %q", field), fmt.Sprintf("has no field %q", field), nil, true
		}
		if len(args) == 2 {
			r, t, f, err, valid := validate(val.Value(), args[1], binding)
			if err != nil {
				return ValidatorErrorf("map entry %q %s", field, err)
			}
			if !valid {
				return false, "", "", nil, false
			}
			if !r {
				return false, fmt.Sprintf("map entry %q %s", field, t), fmt.Sprintf("map key %q %s", field, f), nil, true
			}
			return true, fmt.Sprintf("map entry %q %s", field, t), fmt.Sprintf("map entry %q %s", field, t), nil, true
		}
		return true, fmt.Sprintf("map entry %q exists", field), fmt.Sprintf("map entry %q exists", field), nil, true

	case "and", "not", "":
		if len(args) == 0 {
			return ValidatorErrorf("validator argument required")
		}
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
		if len(args) == 0 {
			return ValidatorErrorf("validator argument required")
		}
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
			return ValidatorErrorf("invalid type for empty: %s", ExpressionType(v))
		}
	case "valueset":
		if len(args) != 1 {
			return ValidatorErrorf("valueset requires a list argument with possible values")
		}
		l, ok := args[0].Value().([]yaml.Node)
		if !ok {
			return ValidatorErrorf("valueset requires a list argument with possible values")
		}
		for _, v := range l {
			if ok, _, _ := compareEquals(value, v.Value()); ok {
				return true, "matches valueset", "oops", nil, true
			}
		}
		s, ok := value.(string)
		if ok {
			return false, fmt.Sprint("valid value %q", s), fmt.Sprintf("invalid value %q", s), nil, true
		}
		i, ok := value.(int64)
		if ok {
			return false, fmt.Sprint("valid value %d", i), fmt.Sprintf("invalid value %d", i), nil, true
		}
		return false, "valid value", "invalid value", nil, true

	case "match":
		if len(args) != 1 {
			return ValidatorErrorf("match requires a regexp argument")
		}
		s, ok := args[0].Value().(string)
		if !ok {
			return ValidatorErrorf("match requires a regexp argument")
		}

		re, err := regexp.Compile(s)
		if err != nil {
			return ValidatorErrorf("regexp %s: %s", s, err)
		}
		s, ok = value.(string)
		if !ok {
			return ValidatorErrorf("no string to match regexp")
		}
		if !re.MatchString(s) {
			return false, fmt.Sprintf("invalid value %q", s), fmt.Sprintf("invalid value %q", s), nil, true
		}
		return true, fmt.Sprintf("valid value %q", s), fmt.Sprintf("valid value %q", s), nil, true

	case "type":
		e := ExpressionType(value)
		for _, t := range args {
			s, err := StringValue("type arg", t.Value())
			if err != nil {
				return ValidatorErrorf("%s: %s", op, err)
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
			return ValidatorErrorf("%s: %s", op, err)
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
			return ValidatorErrorf("%s: %s", op, err)
		}
		if err := IsDNS1123Label(s); err != nil {
			return false, "is dns label", fmt.Sprintf("is no dns label: %s", err), nil, true
		}
		return true, "is dns label", "is no dns label", nil, true
	case "dnsdomain":
		s, err := StringValue(op, value)
		if err != nil {
			return ValidatorErrorf("%s: %s", op, err)
		}
		if err := IsDNS1123Subdomain(s); err != nil {
			return false, "is dns domain", fmt.Sprintf("is no dns domain: %s", err), nil, true
		}
		return true, "is dns domain", "is no dns domain", nil, true
	case "wildcarddnsdomain":
		s, err := StringValue(op, value)
		if err != nil {
			return ValidatorErrorf("%s: %s", op, err)
		}
		if err := IsWildcardDNS1123Subdomain(s); err != nil {
			return false, "is wildcard dns domain", fmt.Sprintf("is no wildcard dns domain: %s", err), nil, true
		}
		return true, "is wildcard dns domain", "is no wildcard dns domain", nil, true
	case "ip":
		s, err := StringValue(op, value)
		if err != nil {
			return ValidatorErrorf("%s: %s", op, err)
		}
		if ip := net.ParseIP(s); ip == nil {
			return false, "is ip address", fmt.Sprintf("is no ip address: %s", s), nil, true
		}
		return true, "is ip address", "is no ip address", nil, true
	case "cidr":
		s, err := StringValue(op, value)
		if err != nil {
			return ValidatorErrorf("%s: %s", op, err)
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
		return ValidatorErrorf("unknown validation operator %q", op)
	}
}

func StringValue(msg string, v interface{}) (string, error) {
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("%s requires string, but got %s", msg, ExpressionType(v))
	}
	return s, nil
}

func ValidatorErrorf(msgfmt string, args ...interface{}) (bool, string, string, error, bool) {
	return false, "", "", fmt.Errorf(msgfmt, args...), true
}
