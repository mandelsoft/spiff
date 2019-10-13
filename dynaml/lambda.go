package dynaml

import (
	"fmt"
	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/yaml"
	"reflect"
)

func staticScope(binding Binding) Binding {
	self, _ := binding.FindReference([]string{yaml.SELF})
	if self != nil {
		return self.Resolver().(Binding)
	}
	return binding
}

type Parameter struct {
	Name    string
	Default Expression
}

func (p Parameter) String() string {
	if p.Default != nil {
		return fmt.Sprintf("%s=%s", p.Name, p.Default)
	}
	return p.Name
}

type LambdaExpr struct {
	Parameters []Parameter
	VarArgs    bool
	E          Expression
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
	var params []Parameter
	for i, p := range e.Parameters {
		if p.Default != nil {
			if params == nil {
				params = e.Parameters[:i]
			}
			v, info, ok := p.Default.Evaluate(binding, locally)
			if !ok {
				debug.Debug("failed LAMBDA Default Arg Evaluation %d: %s", i, info.Issue.Issue)
				nested := info.Issue
				info.SetError("evaluation of lambda call default argument %d failed: %s: %s", i, e, p.Default)
				info.Issue.Nested = append(info.Issue.Nested, nested)
				info.Issue.Sequence = true
				return nil, info, ok
			}
			if isExpression(value) {
				debug.Debug("delay LAMBDA expression for default argument %d", i)
				return nil, info, ok
			}
			params = append(params, Parameter{p.Name, ValueExpr{v}})
		}
	}
	if params == nil {
		params = e.Parameters
	}

	debug.Debug("LAMBDA VALUE with resolver %+v\n", binding)
	return LambdaValue{params, e, binding.GetStaticBinding(), staticScope(binding)}, info, true
}

func (e LambdaExpr) String() string {
	str := ""
	sep := ""
	for _, p := range e.Parameters {
		str = fmt.Sprintf("%s%s%s", str, sep, p)
		sep = ","
	}
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
		value, info, ok := lexpr.Evaluate(binding, locally)
		if !ok || isExpression(value) {
			return value, info, ok
		}
		lambda = value.(LambdaValue)

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
	Parameters []Parameter
	lambda     LambdaExpr
	static     map[string]yaml.Node
	resolver   Binding
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
		return fmt.Sprintf("%#v", v)
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

func (e LambdaValue) Evaluate(inline bool, curry, autocurry bool, args []interface{}, binding Binding, locally bool) (bool, interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()
	required := len(e.lambda.Parameters)

	if e.lambda.VarArgs {
		if required > 0 {
			required--
		}
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
	if len(args) > len(e.lambda.Parameters) {
		info.Issue = yaml.NewIssue("found %d argument(s), but expects %d", len(args), len(e.lambda.Parameters))
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
		inp[e.lambda.Parameters[i].Name] = NewNode(v, binding)
	}

	if curry || (autocurry && len(args) < len(e.lambda.Parameters) && e.lambda.Parameters[len(e.lambda.Parameters)-1].Default == nil) {
		debug.Debug("LAMBDA CALL: currying %+v\n", inp)
		rest := []Parameter{}
		if len(args) < len(e.lambda.Parameters) {
			rest = e.lambda.Parameters[len(args):]
		}
		return true, LambdaValue{
			rest,
			LambdaExpr{rest, e.lambda.VarArgs && len(rest) > 0, e.lambda.E},
			inp,
			e.resolver,
		}, DefaultInfo(), true
	}
	if len(args) < len(e.lambda.Parameters) {
		if e.lambda.Parameters[len(args)].Default != nil {
			for i := len(args); i < len(e.lambda.Parameters); i++ {
				inp[e.lambda.Parameters[i].Name] = NewNode(e.lambda.Parameters[i].Default.(ValueExpr).Value, binding)
			}
		} else {
			info.SetError("expected %d arguments, but found %d", len(e.lambda.Parameters), len(args))
			return false, nil, info, false
		}
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
