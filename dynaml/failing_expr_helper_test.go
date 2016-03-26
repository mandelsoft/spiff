package dynaml

type FailingExpr struct{}

func (FailingExpr) Evaluate(Binding, bool) (interface{}, EvaluationInfo, bool) {
	return nil, DefaultInfo(), false
}
