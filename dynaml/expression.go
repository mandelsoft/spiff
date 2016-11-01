package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type Status interface {
	error
	Issue(fmt string, args ...interface{}) (issue yaml.Issue, localError bool, failed bool)
	HasError() bool
}

type SourceProvider interface {
	SourceName() string
}

type Binding interface {
	SourceProvider
	GetLocalBinding() map[string]yaml.Node
	FindFromRoot([]string) (yaml.Node, bool)
	FindReference([]string) (yaml.Node, bool)
	FindInStubs([]string) (yaml.Node, bool)

	WithScope(step map[string]yaml.Node) Binding
	WithLocalScope(step map[string]yaml.Node) Binding
	WithPath(step string) Binding
	WithSource(source string) Binding
	RedirectOverwrite(path []string) Binding

	Path() []string
	StubPath() []string

	Flow(source yaml.Node, shouldOverride bool) (yaml.Node, Status)
}

type EvaluationInfo struct {
	RedirectPath []string
	Replace      bool
	Merged       bool
	Preferred    bool
	KeyName      string
	Source       string
	LocalError   bool
	Failed       bool
	Undefined    bool
	Issue        yaml.Issue
	yaml.NodeFlags
}

func (e EvaluationInfo) SourceName() string {
	return e.Source
}

func DefaultInfo() EvaluationInfo {
	return EvaluationInfo{nil, false, false, false, "", "", false, false, false, yaml.Issue{}, 0}
}

type Expression interface {
	Evaluate(Binding, bool) (interface{}, EvaluationInfo, bool)
}

func (i *EvaluationInfo) Error(msgfmt interface{}, args ...interface{}) (interface{}, EvaluationInfo, bool) {
	i.LocalError = true
	i.Issue = yaml.NewIssue(msgfmt.(string), args...)
	return nil, *i, false
}

func (i *EvaluationInfo) SetError(msgfmt interface{}, args ...interface{}) {
	i.LocalError = true
	i.Issue = yaml.NewIssue(msgfmt.(string), args...)
}

func (i *EvaluationInfo) PropagateError(value interface{}, state Status, msgfmt string, args ...interface{}) (interface{}, EvaluationInfo, bool) {
	i.Issue, i.LocalError, i.Failed = state.Issue(msgfmt, args...)
	if i.LocalError {
		value = nil
	}
	return value, *i, !i.LocalError
}

func (i EvaluationInfo) Join(o EvaluationInfo) EvaluationInfo {
	if o.RedirectPath != nil {
		i.RedirectPath = o.RedirectPath
	}
	i.Replace = o.Replace // replace only by directly using the merge node
	i.Preferred = i.Preferred || o.Preferred
	i.Merged = i.Merged || o.Merged
	if o.KeyName != "" {
		i.KeyName = o.KeyName
	}
	if o.Issue.Issue != "" {
		i.Issue = o.Issue
	}
	if o.LocalError {
		i.LocalError = true
	}
	if o.Failed {
		i.Failed = true
	}
	if o.Undefined {
		i.Undefined = true
	}
	i.NodeFlags |= o.NodeFlags
	return i
}

func ResolveExpressionOrPushEvaluation(e *Expression, resolved *bool, info *EvaluationInfo, binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	val, infoe, ok := (*e).Evaluate(binding, locally)
	if info != nil {
		infoe = (*info).Join(infoe)
	}
	if !ok {
		return nil, infoe, false
	}

	if v, ok := val.(Expression); ok {
		*e = v
		*resolved = false
		return nil, infoe, true
	} else {
		return val, infoe, true
	}
}

func ResolveIntegerExpressionOrPushEvaluation(e *Expression, resolved *bool, info *EvaluationInfo, binding Binding, locally bool) (int64, EvaluationInfo, bool) {
	value, infoe, ok := ResolveExpressionOrPushEvaluation(e, resolved, info, binding, locally)

	if value == nil {
		return 0, infoe, ok
	}

	i, ok := value.(int64)
	if ok {
		return i, infoe, true
	} else {
		infoe.Issue = yaml.NewIssue("integer operand required")
		return 0, infoe, false
	}
}

func ResolveExpressionListOrPushEvaluation(list *[]Expression, resolved *bool, info *EvaluationInfo, binding Binding, locally bool) ([]interface{}, EvaluationInfo, bool) {
	values := make([]interface{}, len(*list))
	pushed := make([]Expression, len(*list))
	infoe := EvaluationInfo{}
	ok := true

	copy(pushed, *list)

	for i, _ := range pushed {
		values[i], infoe, ok = ResolveExpressionOrPushEvaluation(&pushed[i], resolved, info, binding, locally)
		info = &infoe
		if !ok {
			return nil, infoe, false
		}
	}
	*list = pushed
	return values, infoe, true

}
