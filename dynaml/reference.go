package dynaml

import (
	"strings"

	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/yaml"
)

type ReferenceExpr struct {
	Tag  string
	Path []string
}

func NewReferenceExpr(path ...string) ReferenceExpr {
	return ReferenceExpr{"", path}
}

func NewTaggedReferenceExpr(tag string, path ...string) ReferenceExpr {
	return ReferenceExpr{tag, path}
}

func (e ReferenceExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	var tag *Tag
	fromRoot := e.Path[0] == ""

	debug.Debug("reference: (%s)%v\n", e.Tag, e.Path)
	if e.Tag != "" {
		info := DefaultInfo()
		if e.Tag != "doc:0" {
			tag = binding.GetState().GetTag(e.Tag)
			if tag == nil {
				return info.Error("tag '%s' not found", e.Tag)
			}
			if len(e.Path) == 1 && e.Path[0] == "" {
				return tag.Node().Value(), info, true
			}
		} else {
			if len(e.Path) == 1 && e.Path[0] == "" {
				return info.Error("no reference to actual document possible")
			}
			fromRoot = true
		}
	}
	return e.find(func(end int, path []string) (yaml.Node, bool) {
		if fromRoot {
			start := 0
			if e.Path[0] == "" {
				start = 1
			}
			return binding.FindFromRoot(path[start : end+1])
		} else {
			if tag != nil {
				return yaml.Find(tag.Node(), path...)
			}
			return binding.FindReference(path[:end+1])
		}
	}, binding, locally)
}

func (e ReferenceExpr) String() string {
	tag := ""
	if e.Tag != "" {
		tag = e.Tag + "::"
	}
	return tag + strings.Join(e.Path, ".")
}

func (e ReferenceExpr) find(f func(int, []string) (node yaml.Node, x bool), binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	var step yaml.Node
	var ok bool

	info := DefaultInfo()
	debug.Debug("resolving ref [%v]", e.Path)
	for i := 0; i < len(e.Path); i++ {
		step, ok = f(i, e.Path)

		debug.Debug("  %d: %v %+v\n", i, ok, step)
		if !ok {
			return info.Error("'%s' not found", strings.Join(e.Path[0:i+1], "."))
		}

		if !isLocallyResolved(step) {
			debug.Debug("  locally unresolved %T\n", step.Value())
			if _, ok := step.Value().(Expression); ok {
				info.Issue = yaml.NewIssue("'%s' unresolved", strings.Join(e.Path[0:i+1], "."))
			} else {
				info.Issue = yaml.NewIssue("'%s' not complete", strings.Join(e.Path[0:i+1], "."))
			}
			info.Failed = step.Failed() || step.HasError()
			return e, info, true
		}
	}

	if !locally && !isResolvedValue(step.Value()) {
		debug.Debug("  unresolved\n")
		info.Issue = yaml.NewIssue("'%s' unresolved", strings.Join(e.Path, "."))
		info.Failed = step.Failed() || step.HasError()
		return e, info, true
	}

	debug.Debug("reference %v -> %+v\n", e.Path, step)
	info.KeyName = step.KeyName()
	return value(yaml.ReferencedNode(step)), info, true
}
