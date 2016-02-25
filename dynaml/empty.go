package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type EmptyHashExpr struct{}

func (e EmptyHashExpr) Evaluate(binding Binding) (interface{}, EvaluationInfo, bool) {
	return make(map[string]yaml.Node), DefaultInfo(), true
}

func (e EmptyHashExpr) String() string {
	return "{}"
}
