package dynaml

import (
	"fmt"

	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/yaml"
)

type LambdaExpr struct {
	Names []string
	E     Expression
}

func (e LambdaExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()
	return LambdaValue{e, binding.GetLocalBinding(), binding}, info, true
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
		lambda = LambdaValue{lexpr, binding.GetLocalBinding(), binding}

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
	lambda  LambdaExpr
	local   map[string]yaml.Node
	binding Binding
}

func (e LambdaValue) String() string {
	binding := ""
	if len(e.local) > 0 {
		binding = "{"
		for n, v := range e.local {
			if n != "_" {
				binding += fmt.Sprintf("%s: %v,", n, v.Value())
			}
		}
		binding += "}"
	}
	return fmt.Sprintf("%s%s", binding, e.lambda)
}

func (e LambdaValue) MarshalYAML() (tag string, value interface{}, err error) {
	return "", "(( " + e.lambda.String() + " ))", nil
}

func (e LambdaValue) Evaluate(args []interface{}, binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(args) > len(e.lambda.Names) {
		info.Issue = yaml.NewIssue("found %d argument(s), but expects %d", len(args), len(e.lambda.Names))
		return nil, info, false
	}
	inp := map[string]yaml.Node{}
	for n, v := range e.local {
		inp[n] = v
	}
	debug.Debug("LAMBDA CALL: inherit local %+v\n", inp)
	inp[yaml.SELF] = yaml.ResolverNode(node(e, binding), e.binding)
	for i, v := range args {
		inp[e.lambda.Names[i]] = node(v, binding)
	}
	debug.Debug("LAMBDA CALL: effective local %+v\n", inp)

	if len(args) < len(e.lambda.Names) {
		rest := e.lambda.Names[len(args):]
		return LambdaValue{LambdaExpr{rest, e.lambda.E}, inp, e.binding}, DefaultInfo(), true
	}

	return e.lambda.E.Evaluate(binding.WithLocalScope(inp), locally)
}
