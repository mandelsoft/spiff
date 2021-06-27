package flow

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/onsi/gomega"

	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/yaml"
)

type MatcherSupport struct {
	gomega.OmegaMatcher
	features []string
	Expected yaml.Node
	Stubs    []yaml.Node
	actual   yaml.Node
}

func (matcher *MatcherSupport) WithFeatures(features ...string) *MatcherSupport {
	matcher.features = features
	return matcher
}

func (s *MatcherSupport) createEnv() dynaml.Binding {
	features := features.FeatureFlags{}
	for _, name := range s.features {
		features.Set(name, true)
	}
	return NewEnvironment(nil, "", NewDefaultState().SetFeatures(features))
}

func FlowAs(expected yaml.Node, stubs ...yaml.Node) *FlowAsMatcher {
	matcher := &FlowAsMatcher{}
	support := &MatcherSupport{OmegaMatcher: matcher, Expected: expected, Stubs: stubs}
	matcher.MatcherSupport = support
	return matcher
}

type FlowAsMatcher struct {
	*MatcherSupport
}

func (matcher *FlowAsMatcher) Match(source interface{}) (success bool, err error) {
	if source == nil && matcher.Expected == nil {
		return false, fmt.Errorf("Refusing to compare <nil> to <nil>.")
	}
	env := matcher.createEnv()
	matcher.actual, err = NestedFlow(env, source.(yaml.Node), matcher.Stubs...)
	if err != nil {
		return false, err
	}

	matcher.actual = Cleanup(matcher.actual, discardTemporary)

	if matcher.actual.EquivalentToNode(matcher.Expected) {
		return true, nil
	} else {
		return false, nil
	}

	return
}

func formatMessage(actual yaml.Node, message string, expected yaml.Node) string {
	return fmt.Sprintf("Expected%s\n%s%s", formatYAML(actual), message, formatYAML(expected))
}

func formatYAML(yaml yaml.Node) string {
	formatted, err := candiedyaml.Marshal(yaml)
	if err != nil {
		return fmt.Sprintf("\n\t<%T> %#v", yaml, yaml)
	}

	return fmt.Sprintf("\n\t%s", strings.Replace(string(formatted), "\n", "\n\t", -1))
}

func (matcher *FlowAsMatcher) FailureMessage(actual interface{}) (message string) {
	return formatMessage(matcher.actual, "to flow as", matcher.Expected)
}

func (matcher *FlowAsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return formatMessage(matcher.actual, "not to flow as", matcher.Expected)
}

func FlowToErr(expected string, stubs ...yaml.Node) *FlowErrAsMatcher {
	expected = `unresolved nodes:
` + expected
	matcher := &FlowErrAsMatcher{Expected: expected}
	support := &MatcherSupport{OmegaMatcher: matcher, Stubs: stubs}
	matcher.MatcherSupport = support
	return matcher
}

type FlowErrAsMatcher struct {
	*MatcherSupport
	Expected string
	actual   string
}

func (matcher *FlowErrAsMatcher) Match(source interface{}) (success bool, err error) {
	env := matcher.createEnv()
	_, err = NestedFlow(env, source.(yaml.Node), matcher.Stubs...)
	if err == nil {
		return false, fmt.Errorf("no error reported")
	}
	matcher.actual = err.Error()
	return matcher.actual == matcher.Expected, nil
}

func formatErrorMessage(actual string, message string, expected string) string {
	return fmt.Sprintf("Expected\n%s\n%s\n%s", actual, message, expected)
}

func (matcher *FlowErrAsMatcher) FailureMessage(actual interface{}) (message string) {
	return formatErrorMessage(matcher.actual, "to be equal to", matcher.Expected)
}

func (matcher *FlowErrAsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return formatErrorMessage(matcher.actual, "not to be equla to", matcher.Expected)
}
