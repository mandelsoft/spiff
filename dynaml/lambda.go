package dynaml

import (
	"fmt"
	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/yaml"
	"reflect"
	"strings"
)

func staticScope(binding Binding) Binding {
	self, _ := binding.FindReference([]string{yaml.SELF})
	if self != nil {
		return self.Resolver().(Binding)
	}
	return binding
}

type LambdaExpr struct {
	Names   []string
	VarArgs bool
	E       Expression
}

func isInline(e Expression) bool {
	_, ok := e.(LambdaExpr)
	return ok
}

func keys(m map[string]yaml.Node) []string {
	s := []string{}
	for k := range m {
		s = append(s, k)
	}
	return s
}

func (e LambdaExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()
	debug.Debug("LAMBDA VALUE with resolver %+v\n", binding)
	return LambdaValue{e, binding.GetStaticBinding(), staticScope(binding)}, info, true
}

func (e LambdaExpr) String() string {
	str := strings.Join(e.Names, ",")
	if e.VarArgs {
		str += "..."
	}
	return fmt.Sprintf("lambda|%s|->%s", str, e.E)
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
			return info.Error("cannot parse lamba expression '%s': %s", err)
		}
		lexpr, ok := expr.(LambdaExpr)
		if !ok {
			debug.Debug("no lambda expression: %T\n", expr)
			return info.Error("'%s' is no lambda expression", v)
		}
		lambda = LambdaValue{lexpr, binding.GetStaticBinding(), staticScope(binding)}

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
	static   map[string]yaml.Node
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

func short(val interface{}, all bool) string {
	switch v := val.(type) {
	case []yaml.Node:
		s := "["
		sep := ""
		for _, e := range v {
			s = fmt.Sprintf("%s%s%s", s, sep, short(e.Value(), all))
			sep = ", "
		}
		return s + "]"
	case map[string]yaml.Node:
		s := "{"
		sep := ""
		for _, k := range getSortedKeys(v) {
			if all || k != "_" {
				s = fmt.Sprintf("%s%s%s: %s", s, sep, k, short(v[k].Value(), all))
				sep = ", "
			}
		}
		return s + "}"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (e LambdaValue) String() string {
	binding := ""
	if len(e.static) > 0 {
		binding = short(e.static, false)
	}
	return fmt.Sprintf("%s%s", shorten(binding), e.lambda)
}

func shorten(s string) string {
	if len(s) > 40 {
		s = s[:17] + " ... " + s[len(s)-17:]
	}
	return s
}

func (e LambdaValue) MarshalYAML() (tag string, value interface{}, err error) {
	return "", "(( " + e.lambda.String() + " ))", nil
}

func (e LambdaValue) Evaluate(inline bool, curry bool, args []interface{}, binding Binding, locally bool) (bool, interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()
	required := len(e.lambda.Names)

	if e.lambda.VarArgs {
		required--
		if len(args) > required {
			varargs := []yaml.Node{}
			for _, a := range args[required:] {
				varargs = append(varargs, yaml.NewNode(a, binding.SourceName()))
			}
			args = append(args[:required], varargs)
		} else {
			if len(args) == required {
				args = append(args, []yaml.Node{})
			}
		}
	}
	if len(args) > len(e.lambda.Names) {
		info.Issue = yaml.NewIssue("found %d argument(s), but expects %d", len(args), len(e.lambda.Names))
		return false, nil, info, false
	}
	inp := map[string]yaml.Node{}
	for n, v := range e.static {
		//fmt.Printf("  static %s: %s\n", n, ExpressionType(v.Value()))
		inp[n] = v
	}
	debug.Debug("LAMBDA CALL: inherit local %+v\n", inp)
	for i, v := range args {
		//fmt.Printf("  dyn %s: %s\n", e.lambda.Names[i], ExpressionType(v))
		inp[e.lambda.Names[i]] = NewNode(v, binding)
	}

	if len(args) < len(e.lambda.Names) {
		if curry {
			debug.Debug("LAMBDA CALL: currying %+v\n", inp)
			rest := e.lambda.Names[len(args):]
			return true, LambdaValue{LambdaExpr{rest, e.lambda.VarArgs, e.lambda.E}, inp, e.resolver}, DefaultInfo(), true
		}
		info.SetError("expected %d arguments, but found %d", len(e.lambda.Names), len(args))
		return false, nil, info, false
	}
	if !inline {
		debug.Debug("LAMBDA CALL: staticScope %+v\n", e.resolver)
		inp[yaml.SELF] = yaml.ResolverNode(NewNode(e, binding), e.resolver)
		debug.Debug("LAMBDA CALL: effective local %+v\n", inp)
	}
	value, info, ok := e.lambda.E.Evaluate(binding.WithLocalScope(inp), locally)
	if !ok {
		debug.Debug("failed LAMBDA CALL: %s", info.Issue.Issue)
		nested := info.Issue
		info.SetError("evaluation of lambda expression failed: %s: %s", e, shorten(short(inp, false)))
		info.Issue.Nested = append(info.Issue.Nested, nested)
		info.Issue.Sequence = true
		return false, nil, info, ok
	}
	if isExpression(value) {
		debug.Debug("delay LAMBDA CALL")
		return false, nil, info, ok
	}
	return true, value, info, ok
}
