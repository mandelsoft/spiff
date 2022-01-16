package flow

import (
	"fmt"

	"github.com/mandelsoft/spiff/yaml"
)

func CascadeAs(expected yaml.Node, stubs ...yaml.Node) *CascadeAsMatcher {
	matcher := &CascadeAsMatcher{}
	support := &MatcherSupport{OmegaMatcher: matcher, Expected: expected, Stubs: stubs}
	matcher.MatcherSupport = support
	return matcher
}

type CascadeAsMatcher struct {
	*MatcherSupport
}

func (matcher *CascadeAsMatcher) WithFeatures(features ...string) *CascadeAsMatcher {
	matcher.features = features
	return matcher
}

func (matcher *CascadeAsMatcher) Match(source interface{}) (success bool, err error) {
	if source == nil && matcher.Expected == nil {
		return false, fmt.Errorf("Refusing to compare <nil> to <nil>.")
	}
	env := matcher.createEnv()
	matcher.actual, err = Cascade(env, source.(yaml.Node), Options{}, matcher.Stubs...)
	if err != nil {
		return false, err
	}

	if matcher.actual.EquivalentToNode(matcher.Expected) {
		return true, nil
	} else {
		return false, nil
	}

	return
}

func (matcher *CascadeAsMatcher) FailureMessage(actual interface{}) (message string) {
	return formatMessage(matcher.actual, "to flow as", matcher.Expected)
}

func (matcher *CascadeAsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return formatMessage(matcher.actual, "not to flow as", matcher.Expected)
}
