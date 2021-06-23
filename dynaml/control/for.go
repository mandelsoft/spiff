package control

import (
	"fmt"

	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/yaml"
)

func init() {
	dynaml.RegisterControl("for", flowFor, "*do", "*mapkey")
}

type iteration struct {
	name    string
	values  []yaml.Node
	current int
}

func flowFor(val yaml.Node, node yaml.Node, fields, opts map[string]yaml.Node, env dynaml.Binding) (yaml.Node, bool) {
	for range fields {
		return dynaml.ControlIssue("for", node, "no regular fields allowed in for control")
	}
	if !dynaml.IsResolvedNode(val, env) {
		return node, false
	}
	body, ok := opts["do"]
	if !ok {
		return dynaml.ControlIssue("for", node, "do fields required in for control")
	}
	if !dynaml.IsResolvedNode(body, env) {
		return node, false
	}

	var mapkey *dynaml.SubstitutionExpr
	k, ok := opts["mapkey"]
	if ok {
		if !dynaml.IsResolvedNode(k, env) {
			return node, false
		}
		if t, ok := k.Value().(dynaml.TemplateValue); ok {
			mapkey = &dynaml.SubstitutionExpr{dynaml.ValueExpr{t}}
		} else {
			return dynaml.ControlIssue("for", node, "mapkey must be an expression")
		}
	}

	var subst *dynaml.SubstitutionExpr
	if t, ok := body.Value().(dynaml.TemplateValue); ok {
		subst = &dynaml.SubstitutionExpr{dynaml.ValueExpr{t}}
	}
	iterations := []iteration{}
	switch def := val.Value().(type) {
	case map[string]yaml.Node:
		vars := yaml.GetSortedKeys(def)
		iterations = make([]iteration, len(vars))
		for i, v := range vars {
			values, ok := def[v].Value().([]yaml.Node)
			if !ok {
				return dynaml.ControlIssue("for", node, "control variable %q required list value but got %s", v, dynaml.ExpressionType(def[v].Value()))
			}
			if len(values) == 0 {
				return nil, true
			}
			iterations[len(iterations)-i-1] = iteration{v, values, 0}
		}
	case []yaml.Node:
		iterations = make([]iteration, len(def))
		for i, v := range def {
			spec, ok := v.Value().(map[string]yaml.Node)
			if !ok {
				return dynaml.ControlIssue("for", node, "control variable list entry requires may but got %s", dynaml.ExpressionType(v.Value()))
			}
			if len(spec) != 2 {
				return dynaml.ControlIssue("for", node, "control variable list entry requires two fields: name and values")
			}
			n := spec["name"]
			if n == nil {
				return dynaml.ControlIssue("for", node, "control variable list entry requires name field")
			}
			name, ok := n.Value().(string)
			if !ok {
				return dynaml.ControlIssue("for", node, "control variable name must be of type string but got %s", dynaml.ExpressionType(n.Value()))
			}
			l := spec["values"]
			if l == nil {
				return dynaml.ControlIssue("for", node, "control variable list entry requires values field")
			}
			values, ok := l.Value().([]yaml.Node)
			if !ok {
				return dynaml.ControlIssue("for", node, "control variable values must be of type list but got %s", dynaml.ExpressionType(l.Value()))
			}
			iterations[len(iterations)-i-1] = iteration{name, values, 0}
		}
	default:
		return dynaml.ControlIssue("for", node, "value field must be map but got %s", dynaml.ExpressionType(def))
	}

	var resultlist []yaml.Node
	var resultmap map[string]yaml.Node

	if mapkey != nil {
		resultmap = map[string]yaml.Node{}
	} else {
		resultlist = []yaml.Node{}
	}

	done := true
	issue := yaml.Issue{}
outer:
	for {
		// do
		inp := map[string]yaml.Node{}
		for i := 0; i < len(iterations); i++ {
			inp[iterations[i].name] = iterations[i].values[iterations[i].current]
			inp["index-"+iterations[i].name] = yaml.NewNode(int64(iterations[i].current), "for")
		}
		scope := env.WithLocalScope(inp)
		key := ""
		if mapkey != nil {
			k, info, ok := mapkey.Evaluate(scope, false)
			if !ok {
				done = false
				issue.Nested = append(issue.Nested, controlVariablesIssue(iterations, info.Issue))
			}
			if key, ok = k.(string); !ok {
				done = false
				issue.Nested = append(issue.Nested, controlVariablesIssue(iterations, yaml.NewIssue("map key must be string, but found %s", dynaml.ExpressionType(k))))
			}
		}
		if subst != nil {
			v, info, ok := subst.Evaluate(scope, false)
			if !ok {
				done = false
				issue.Nested = append(issue.Nested, controlVariablesIssue(iterations, info.Issue))
			} else {
				if dynaml.IsExpression(v) {
					done = false
				} else {
					if mapkey != nil {
						resultmap[key] = yaml.NewNode(v, node.SourceName())
					} else {
						resultlist = append(resultlist, yaml.NewNode(v, node.SourceName()))
					}
				}
			}
		} else {
			if mapkey != nil {
				resultmap[key] = opts["do"]
			} else {
				resultlist = append(resultlist, opts["do"])
			}
		}

		for i := 0; i <= len(iterations); i++ {
			if i == len(iterations) {
				break outer
			}
			iterations[i].current++
			if iterations[i].current < len(iterations[i].values) {
				break
			}
			iterations[i].current = 0
		}
	}
	if !done {
		if len(issue.Nested) > 0 {
			issue.Issue = "error evaluationg for body"
			return dynaml.ControlIssueByIssue("for", node, issue, false)
		}
		return node, false
	}
	if resultlist != nil {
		return yaml.NewNode(resultlist, node.SourceName()), true
	}
	return yaml.NewNode(resultmap, node.SourceName()), true
}

func controlVariablesIssue(iterations []iteration, issue yaml.Issue) yaml.Issue {
	desc := fmt.Sprintf("control variables: ")
	sep := ""
	for _, i := range iterations {
		desc = fmt.Sprintf("%s%s %s=%s", desc, sep, i.name, dynaml.Shorten(dynaml.Short(i.values[i.current].Value(), false)))
		sep = ";"
	}
	issue.Issue = fmt.Sprintf("%s: %s", desc, issue.Issue)
	return issue
}
