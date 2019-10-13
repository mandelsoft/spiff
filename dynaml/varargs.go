package dynaml

import "fmt"

type VarArgsExpression interface {
	Expression
	IsVarArg() bool
}

type VarArgsExpr struct {
	Expression
}

func (e VarArgsExpr) String() string {
	return fmt.Sprintf("%s...", e.Expression)
}

func (e VarArgsExpr) Evaluate(binding Binding, locally bool) (interface{}, EvaluationInfo, bool) {
	return e.Expression.Evaluate(binding, locally)
}

func (e VarArgsExpr) IsVarArg() bool {
	return true
}

func IsVarArg(e Expression) bool {
	va, ok := e.(VarArgsExpression)
	return ok && va.IsVarArg()
}

func KeepVarArg(e Expression, orig Expression) Expression {
	if va, ok := orig.(VarArgsExpression); ok && va.IsVarArg() {
		if _, ok := e.(VarArgsExpression); !ok {
			return VarArgsExpr{e}
		}
	}
	return e
}
