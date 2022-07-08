package dynaml

import (
	"fmt"

	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/yaml"
)

type DynamicExpr struct {
	Root  Expression
	Index Expression
}

func (e DynamicExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {

	// if root is a reference expression and the type is known allow for element selection if element is resolved
	// regardless of the resolution state of the root
	// enables .["dyn"]
	_, isRef := e.Root.(ReferenceExpr)

	root, info, ok := e.Root.Evaluate(binding, locally || isRef)
	if !ok {
		return nil, info, false
	}

	if !isLocallyResolvedValue(root, binding) {
		return e, info, true
	}

	locally = locally || info.Raw
	/*
		if !locally && !isResolvedValue(root, binding) {
			info.Issue = yaml.NewIssue("'%s' unresolved", e.Expression)
			return e, info, true
		}
	*/

	dyn, infoe, ok := e.Index.Evaluate(binding, locally)

	info.Join(infoe)
	if !ok {
		return nil, info, false
	}
	if !isResolvedValue(dyn, binding) {
		return e, info, true
	}

	debug.Debug("dynamic reference: %+v\n", dyn)

	var qual []string
	switch v := dyn.(type) {
	case int64:
		_, ok := root.([]yaml.Node)
		if !ok {
			return info.Error("index requires array expression")
		}
		qual = []string{fmt.Sprintf("[%d]", v)}
	case string:
		qual = []string{v}
	case []yaml.Node:
		qual = make([]string, len(v))
		for i, e := range v {
			switch v := e.Value().(type) {
			case int64:
				qual[i] = fmt.Sprintf("[%d]", v)
			case string:
				qual[i] = v
			default:
				return info.Error("index or field name required for reference qualifier")
			}
		}
	default:
		return info.Error("index or field name required for reference qualifier")
	}
	return NewReferenceExpr(qual...).find(func(end int, path []string) (yaml.Node, bool) {
		return yaml.Find(NewNode(root, nil), binding.GetFeatures(), path[:end+1]...)
	}, binding, locally)
}

func (e DynamicExpr) String() string {
	return fmt.Sprintf("%s.[%s]", e.Root, e.Index)
}
