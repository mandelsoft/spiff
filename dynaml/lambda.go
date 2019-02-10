package dynaml

import (
	"fmt"
	"reflect"

	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/yaml"
)

func staticScope(binding Binding) Binding {
	self, _ := binding.FindReference([]string{yaml.SELF})
	if self != nil {
		return self.Resolver().(Binding)
	}
	return binding
}

type LambdaExpr struct {
	Names []string
	E     Expression
}

func (e LambdaExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()
	debug.Debug("LAMBDA VALUE with resolver %+v\n", binding)

	return LambdaValue{e, binding.GetLocalBinding(), staticScope(binding)}, info, true
}

func (e LambdaExpr) String() string {
	str := ""
	for _, n := range e.Names {
		str += "," + n
	}
	return fmt.Sprintf("lambda|%s|->%s", str[1:], e.E)
}

type LambdaRefExpr struct {
	Source   Expression
	Path     []string
	StubPath []string
}

func (e LambdaRefExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	var lambda LambdaValue
	resolved := true
	value, info, ok := ResolveExpressionOrPushEvaluation(&e.Source, &resolved, nil, binding, false)
	if !ok {
		return nil, info, false
	}
	if !resolved {
		return e, info, false
	}

	switch v := value.(type) {
	case LambdaValue:
		lambda = v

	case string:
		debug.Debug("LRef: parsing '%s'\n", v)
		expr, err := Parse(v, e.Path, e.StubPath)
		if err != nil {
			debug.Debug("cannot parse: %s\n", err.Error())
			return info.Error("cannot parse lamba expression '%s'")
		}
		lexpr, ok := expr.(LambdaExpr)
		if !ok {
			debug.Debug("no lambda expression: %T\n", expr)
			return info.Error("'%s' is no lambda expression", v)
		}
		lambda = LambdaValue{lexpr, binding.GetLocalBinding(), staticScope(binding)}

	default:
		return info.Error("lambda reference must resolve to lambda value or string")
	}
	debug.Debug("found lambda: %s\n", lambda)
	return lambda, info, true
}

func (e LambdaRefExpr) String() string {
	return fmt.Sprintf("lambda %s", e.Source)
}

type LambdaValue struct {
	lambda   LambdaExpr
	local    map[string]yaml.Node
	resolver Binding
}

var _ StaticallyScopedValue = LambdaValue{}
var _ yaml.ComparableValue = TemplateValue{}

func (e LambdaValue) StaticResolver() Binding {
	return e.resolver
}

func (e LambdaValue) SetStaticResolver(binding Binding) StaticallyScopedValue {
	e.resolver = binding
	return e
}

func (e LambdaValue) EquivalentTo(val interface{}) bool {
	o, ok := val.(LambdaValue)
	return ok && reflect.DeepEqual(e.lambda, o.lambda)
}

func (e LambdaValue) String() string {
	binding := ""
	if len(e.local) > 0 {
		binding = "{"
		sep := ""
		for n, v := range e.local {
			if n != "_" {
				binding += fmt.Sprintf("%s%s: %v", sep, n, v.Value())
				sep = ", "
			}
		}
		binding += "}"
	}
	return fmt.Sprintf("%s%s", binding, e.lambda)
}

func (e LambdaValue) MarshalYAML() (tag string, value interface{}, err error) {
	return "", "(( " + e.lambda.String() + " ))", nil
}

func (e LambdaValue) Evaluate(args []interface{}, binding Binding, locally bool) (bool, interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(args) > len(e.lambda.Names) {
		info.Issue = yaml.NewIssue("found %d argument(s), but expects %d", len(args), len(e.lambda.Names))
		return false, nil, info, false
	}
	inp := map[string]yaml.Node{}
	for n, v := range e.local {
		inp[n] = v
	}
	debug.Debug("LAMBDA CALL: inherit local %+v\n", inp)
	for i, v := range args {
		inp[e.lambda.Names[i]] = node(v, binding)
	}

	if len(args) < len(e.lambda.Names) {
		debug.Debug("LAMBDA CALL: currying %+v\n", inp)
		rest := e.lambda.Names[len(args):]
		return true, LambdaValue{LambdaExpr{rest, e.lambda.E}, inp, e.resolver}, DefaultInfo(), true
	}
	debug.Debug("LAMBDA CALL: staticScope %+v\n", e.resolver)
	inp[yaml.SELF] = yaml.ResolverNode(node(e, binding), e.resolver)
	debug.Debug("LAMBDA CALL: effective local %+v\n", inp)
	value, info, ok := e.lambda.E.Evaluate(binding.WithLocalScope(inp), locally)
	if !ok {
		debug.Debug("failed LAMBDA CALL: %s", info.Issue)
		nested := info.Issue
		info.SetError("evaluation of lambda expression failed: %s", e)
		info.Issue.Nested = append(info.Issue.Nested, nested)
		return false, nil, info, ok
	}
	if isExpression(value) {
		debug.Debug("delay LAMBDA CALL")
		return false, nil, info, ok
	}
	return true, value, info, ok
}
