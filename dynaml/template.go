package dynaml

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/spiff/debug"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type TemplateExpr struct {
}

func (e TemplateExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()
	info.Issue = yaml.NewIssue("&template only usable to declare templates")
	return nil, info, false
}

func (e TemplateExpr) String() string {
	return fmt.Sprintf("&template")
}

type SubstitutionExpr struct {
	Template Expression
	Val      TemplateValue
	Node     yaml.Node
}

func (e SubstitutionExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	if e.Node == nil {
		debug.Debug("evaluating expression to determine template\n")
		n, info, ok := e.Template.Evaluate(binding, false)
		if !ok || isExpression(n) {
			return e, info, ok
		}
		e.Val, ok = n.(TemplateValue)
		if !ok {
			return info.Error("template value required")
		} else {
			e.Node = node_copy(e.Val.Prepared)
		}
	}
	debug.Debug("resolving template '%s'\n", strings.Join(e.Val.Path, "."))
	result, state := binding.Flow(e.Node, false)
	info := DefaultInfo()
	if state != nil {
		if state.HasError() {
			debug.Debug("resolving template failed: " + state.Error())
			return info.PropagateError(e, state, "resolution of template '%s' failed", strings.Join(e.Val.Path, "."))
		} else {
			debug.Debug("resolving template delayed: " + state.Error())
			return e, info, true
		}
	}
	debug.Debug("resolving template succeeded")
	info.Source = result.SourceName()
	return result.Value(), info, true
}

func (e SubstitutionExpr) String() string {
	return fmt.Sprintf("*(%s)", e.Template)
}

type TemplateValue struct {
	Path     []string
	Prepared yaml.Node
	Orig     yaml.Node
}

func (e TemplateValue) MarshalYAML() (tag string, value interface{}, err error) {
	return e.Orig.MarshalYAML()
}

func node_copy(node yaml.Node) yaml.Node {
	if node == nil {
		return nil
	}
	switch val := node.Value().(type) {
	case []yaml.Node:
		list := make([]yaml.Node, len(val))
		for i, v := range val {
			list[i] = node_copy(v)
		}
		return yaml.NewNode(list, node.SourceName())
	case map[string]yaml.Node:
		m := make(map[string]yaml.Node)
		for k, v := range val {
			m[k] = node_copy(v)
		}
		return yaml.NewNode(m, node.SourceName())
	}
	return yaml.NewNode(node.Value(), node.SourceName())
}
