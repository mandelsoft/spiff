package dynaml

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/spiff/debug"
)

type CallExpr struct {
	Function  Expression
	Arguments []Expression
}

func (e CallExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	resolved := true
	funcName := ""
	var value interface{}
	var info EvaluationInfo

	ref, okf := e.Function.(ReferenceExpr)
	if okf && len(ref.Path) == 1 && ref.Path[0] != "" && ref.Path[0] != "_" {
		funcName = ref.Path[0]
	} else {
		value, info, okf = ResolveExpressionOrPushEvaluation(&e.Function, &resolved, &info, binding, false)
		if okf && resolved {
			_, okf = value.(LambdaValue)
			if !okf {
				debug.Debug("function: no string or lambda value: %T\n", value)
				return info.Error("function call '%s' requires function name or lambda value", e.Function)
			}
		}
	}

	switch funcName {
	case "defined":
		return e.defined(binding)
	case "require":
		return e.require(binding)
	case "valid":
		return e.valid(binding)
	case "stub":
		return e.stub(binding)
	}

	values, info, ok := ResolveExpressionListOrPushEvaluation(&e.Arguments, &resolved, nil, binding, false)

	if !okf {
		debug.Debug("failed to resolve function: %s\n", info.Issue)
		return nil, info, false
	}

	if !ok {
		debug.Debug("call args failed\n")
		return nil, info, false
	}

	if !resolved {
		return e, info, true
	}

	var result interface{}
	var sub EvaluationInfo

	switch funcName {
	case "":
		debug.Debug("calling lambda function %#v\n", value)
		result, sub, ok = value.(LambdaValue).Evaluate(values, binding, false)

	case "static_ips":
		result, sub, ok = func_static_ips(e.Arguments, binding)

	case "join":
		result, sub, ok = func_join(values, binding)

	case "split":
		result, sub, ok = func_split(values, binding)

	case "trim":
		result, sub, ok = func_trim(values, binding)

	case "length":
		result, sub, ok = func_length(values, binding)

	case "uniq":
		result, sub, ok = func_uniq(values, binding)

	case "element":
		result, sub, ok = func_element(values, binding)

	case "compact":
		result, sub, ok = func_compact(values, binding)

	case "contains":
		result, sub, ok = func_contains(values, binding)

	case "index":
		result, sub, ok = func_index(values, binding)

	case "lastindex":
		result, sub, ok = func_lastindex(values, binding)

	case "replace":
		result, sub, ok = func_replace(values, binding)

	case "match":
		result, sub, ok = func_match(values, binding)

	case "exec":
		result, sub, ok = func_exec(values, binding)

	case "eval":
		result, sub, ok = func_eval(values, binding, locally)

	case "env":
		result, sub, ok = func_env(values, binding)

	case "read":
		result, sub, ok = func_read(values, binding)

	case "format":
		result, sub, ok = func_format(values, binding)

	case "error":
		result, sub, ok = func_error(values, binding)

	case "min_ip":
		result, sub, ok = func_minIP(values, binding)

	case "max_ip":
		result, sub, ok = func_maxIP(values, binding)

	case "num_ip":
		result, sub, ok = func_numIP(values, binding)

	case "makemap":
		result, sub, ok = func_makemap(values, binding)

	case "list_to_map":
		result, sub, ok = func_list_to_map(e.Arguments[0], values, binding)

	case "ipset":
		result, sub, ok = func_ipset(values, binding)

	case "merge":
		result, sub, ok = func_merge(values, binding)

	case "base64":
		result, sub, ok = func_base64(values, binding)
	case "base64_decode":
		result, sub, ok = func_base64_decode(values, binding)

	case "md5":
		result, sub, ok = func_md5(values, binding)

	case "substr":
		result, sub, ok = func_substr(values, binding)

	case "type":
		if info.Undefined {
			info.Undefined = false
			return "undef", info, ok
		} else {
			result, sub, ok = func_type(values, binding)
		}

	default:
		return info.Error("unknown function '%s'", funcName)
	}

	if ok && (result == nil || isExpression(result)) {
		return e, sub.Join(info), true
	}
	return result, sub.Join(info), ok
}

func (e CallExpr) String() string {
	args := make([]string, len(e.Arguments))
	for i, e := range e.Arguments {
		args[i] = fmt.Sprintf("%s", e)
	}

	return fmt.Sprintf("%s(%s)", e.Function, strings.Join(args, ", "))
}
