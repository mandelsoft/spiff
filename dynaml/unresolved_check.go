package dynaml

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type UnresolvedNodes struct {
	Nodes []UnresolvedNode
}

type UnresolvedNode struct {
	yaml.Node

	Context []string
	Path    []string
}

func (e UnresolvedNodes) Issue(msgfmt string, args ...interface{}) (result yaml.Issue, localError bool, failed bool) {
	format := ""
	result = yaml.NewIssue(msgfmt, args...)
	localError = false
	failed = true

	for _, node := range e.Nodes {
		issue := node.Issue()
		msg := issue.Issue
		if msg != "" {
			msg = "\t" + tag(node) + msg
		}
		if node.HasError() {
			localError = true
		}
		if !node.Failed() {
			failed = false
		}
		switch node.Value().(type) {
		case Expression:
			format = "\t(( %s ))\tin %s\t%s\t(%s)%s"
		default:
			format = "\t%s\tin %s\t%s\t(%s)%s"
		}
		message := fmt.Sprintf(
			format,
			node.Value(),
			node.SourceName(),
			strings.Join(node.Context, "."),
			strings.Join(node.Path, "."),
			msg,
		)
		issue.Issue = message
		result.Nested = append(result.Nested, issue)
	}
	return
}

func (e UnresolvedNodes) HasError() bool {
	for _, node := range e.Nodes {
		issue := node.Issue()
		msg := issue.Issue
		if msg != "" {
			return true
		}
	}
	return false
}

func (e UnresolvedNodes) Error() string {
	message := "unresolved nodes:"
	format := ""

	for _, node := range e.Nodes {
		issue := node.Issue()
		msg := issue.Issue
		if msg != "" {
			msg = "\t" + tag(node) + msg
		}
		switch node.Value().(type) {
		case Expression:
			format = "%s\n\t(( %s ))\tin %s\t%s\t(%s)%s"
		default:
			format = "%s\n\t%s\tin %s\t%s\t(%s)%s"
		}
		message = fmt.Sprintf(
			format,
			message,
			node.Value(),
			node.SourceName(),
			strings.Join(node.Context, "."),
			strings.Join(node.Path, "."),
			msg,
		)
		message += nestedIssues("\t", issue)
	}

	return message
}

func tag(node yaml.Node) string {
	tag := " "
	if !node.Failed() {
		tag = "@"
	} else {
		tag = "-"
	}
	if node.HasError() {
		tag = "*"
	}
	return tag
}

func nestedIssues(gap string, issue yaml.Issue) string {
	message := ""
	if issue.Nested != nil {
		for _, sub := range issue.Nested {
			message = message + "\n" + gap + sub.Issue
			message += nestedIssues(gap+"\t", sub)
		}
	}
	return message
}

func FindUnresolvedNodes(root yaml.Node, context ...string) (result []UnresolvedNode) {
	if root == nil {
		return result
	}

	var nodes []UnresolvedNode
	dummy := []string{"dummy"}

	switch val := root.Value().(type) {
	case map[string]yaml.Node:
		for key, val := range val {
			nodes = append(
				nodes,
				FindUnresolvedNodes(val, addContext(context, key)...)...,
			)
		}

	case []yaml.Node:
		for i, val := range val {
			context := addContext(context, fmt.Sprintf("[%d]", i))

			nodes = append(
				nodes,
				FindUnresolvedNodes(val, context...)...,
			)
		}

	case Expression:
		var path []string
		switch val := root.Value().(type) {
		case AutoExpr:
			path = val.Path
		case MergeExpr:
			path = val.Path
		}

		nodes = append(nodes, UnresolvedNode{
			Node:    root,
			Context: context,
			Path:    path,
		})

	case TemplateValue:
		context := addContext(context, fmt.Sprintf("&"))

		nodes = append(
			nodes,
			FindUnresolvedNodes(val.Orig, context...)...,
		)

	case string:
		if s := yaml.EmbeddedDynaml(root); s != nil {
			_, err := Parse(*s, dummy, dummy)
			if err != nil {
				nodes = append(nodes, UnresolvedNode{
					Node:    yaml.IssueNode(root, true, false, yaml.Issue{Issue: fmt.Sprintf("unparseable expression")}),
					Context: context,
					Path:    []string{},
				})
			}
		}
	}

	for _, n := range nodes {
		if n.GetAnnotation().HasError() {
			result = append(result, n)
		}
	}
	for _, n := range nodes {
		if !n.GetAnnotation().HasError() && !n.GetAnnotation().Failed() {
			result = append(result, n)
		}
	}
	for _, n := range nodes {
		if !n.GetAnnotation().HasError() && n.GetAnnotation().Failed() {
			result = append(result, n)
		}
	}
	return result
}

func ResetUnresolvedNodes(root yaml.Node) yaml.Node {
	if root == nil {
		return root
	}

	switch elem := root.Value().(type) {
	case map[string]yaml.Node:
		for key, val := range elem {
			elem[key] = ResetUnresolvedNodes(val)
		}

	case []yaml.Node:
		for i, val := range elem {
			elem[i] = ResetUnresolvedNodes(val)
		}

	case Expression:
		root = node(fmt.Sprintf("(( %s ))", elem), nil)
	}

	return root
}

func addContext(context []string, step string) []string {
	dup := make([]string, len(context))
	copy(dup, context)
	return append(dup, step)
}

func isExpression(val interface{}) bool {
	if val == nil {
		return false
	}
	_, ok := val.(Expression)
	return ok
}

func isLocallyResolved(node yaml.Node) bool {
	return isLocallyResolvedValue(node.Value())
}

func isLocallyResolvedValue(value interface{}) bool {
	switch v := value.(type) {
	case Expression:
		return false
	case map[string]yaml.Node:
		if !yaml.IsMapResolved(v) {
			return false
		}
	case []yaml.Node:
		if !yaml.IsListResolved(v) {
			return false
		}
	default:
	}

	return true
}

func isResolved(node yaml.Node) bool {
	return node == nil || isResolvedValue(node.Value())
}

func isResolvedValue(val interface{}) bool {
	if val == nil {
		return true
	}
	switch v := val.(type) {
	case Expression:
		return false
	case []yaml.Node:
		for _, n := range v {
			if !isResolved(n) {
				return false
			}
		}
		return true
	case map[string]yaml.Node:
		for _, n := range v {
			if !isResolved(n) {
				return false
			}
		}
		return true

	case string:
		if yaml.EmbeddedDynaml(node(val, nil)) != nil {
			return false
		}
		return true
	default:
		return true
	}
}
