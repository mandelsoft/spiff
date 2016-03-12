package dynaml

import (
	"fmt"

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
	Node     yaml.Node
}

func (e SubstitutionExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	if e.Node == nil {
		debug.Debug("evaluating expression to determine template\n")
		n, info, ok := e.Template.Evaluate(binding, false)
		if !ok || isExpression(n) {
			return e, info, ok
		}
		val, ok := n.(TemplateValue)
		if !ok {
			info.Issue = yaml.NewIssue("template value required")
			return nil, info, false
		} else {
			e.Node = node_copy(val.Prepared)
		}
	}
	debug.Debug("resolving template\n")
	result, state := binding.Flow(e.Node, false)
	info := DefaultInfo()
	if state != nil {
		if state.HasError() {
			debug.Debug("resolving template failed: " + state.Error())
			info.Issue = state.Issue("template resolution failed")
			return e, info, false
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
