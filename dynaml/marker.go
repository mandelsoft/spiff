package dynaml

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/yaml"
)

const (
	TEMPORARY = "&temporary"
	TEMPLATE  = "&template"
	LOCAL     = "&local"
	INJECT    = "&inject"
	DEFAULT   = "&default"
	STATE     = "&state"
	DYNAMIC   = "&dynamic" // POC
)

type MarkerExpr struct {
	list []string
	expr Expression
}

func (e MarkerExpr) IncludesDirectMerge() bool {
	return IncludesDirectMerge(e.expr)
}

func NewTemplateMarker(expr Expression) MarkerExpr {
	return MarkerExpr{list: []string{TEMPLATE}, expr: expr}
}

func (e MarkerExpr) Add(marker string) MarkerExpr {
	for _, v := range e.list {
		if v == marker {
			return e
		}
	}
	e.list = append(e.list, marker)
	return e
}

func (e MarkerExpr) String() string {
	if e.expr != nil {
		return fmt.Sprintf("%s (%s)", strings.Join(e.list, " "), e.expr)
	}
	return fmt.Sprintf("%s", strings.Join(e.list, " "))
}

func (e MarkerExpr) GetTag() string {
	for _, m := range e.list {
		if strings.HasPrefix(m, "&tag:") {
			return m[5:]
		}
	}
	return ""
}

func (e MarkerExpr) GetFlags() yaml.NodeFlags {
	var flags yaml.NodeFlags
	for _, m := range e.list {
		switch m {
		case TEMPORARY:
			flags.SetTemporary()
		case LOCAL:
			flags.SetLocal()
		case INJECT:
			flags.SetInject()
		case DEFAULT:
			flags.SetDefault()
		case STATE:
			flags.SetState()
		case DYNAMIC:
			flags.SetDynamic()
		}
	}
	return flags
}

func (e MarkerExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if e.Has(TEMPLATE) {
		info.Issue = yaml.NewIssue("&template only usable as marker for templates")
		return nil, info, false
	}
	info.AddFlags(e.GetFlags())
	if e.expr != nil {
		result, infoe, ok := e.expr.Evaluate(binding, locally)
		infoe = infoe.Join(info)
		return result, infoe, ok
	}
	return nil, info, true
}

func (e MarkerExpr) setExpression(expr Expression) MarkerExpr {
	e.expr = expr
	return e
}

func (e MarkerExpr) Has(t string) bool {
	for _, v := range e.list {
		if v == t {
			return true
		}
	}
	return false
}

func (e MarkerExpr) add(m string) MarkerExpr {
	e.list = append(e.list, m)
	return e
}

func (e MarkerExpr) TemplateExpression(orig yaml.Node) yaml.Node {
	nlist := []string{}
	for _, m := range e.list {
		if m != TEMPLATE {
			debug.Debug(" preserving marker %s", m)
			nlist = append(nlist, m)
		} else {
			debug.Debug(" omitting marker %s", m)
		}
	}
	m := MarkerExpr{nlist, e.expr}

	// TODO: remove case
	if false && len(nlist) > 0 {
		return yaml.SubstituteNode(fmt.Sprintf("(( %s ))", m), orig)
	}
	if e.expr != nil {
		return yaml.AddFlags(yaml.SubstituteNode(fmt.Sprintf("(( %s ))", e.expr), orig), m.GetFlags())
	}
	return nil
}

func (e MarkerExpr) MarshalYAML() (tag string, value interface{}, err error) {
	return "", fmt.Sprintf("(( %s ))", e.String()), nil
}

func newMarkerExpr(m string) MarkerExpr {
	return MarkerExpr{list: []string{m}}
}

type MarkerExpressionExpr struct {
	contents string
	expr     Expression
}

func (e MarkerExpressionExpr) IncludesDirectMerge() bool {
	return IncludesDirectMerge(e.expr)
}

func (e MarkerExpressionExpr) String() string {
	return e.contents
}

func (e MarkerExpressionExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	return e.expr.Evaluate(binding, locally)
}
