package flow

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/yaml"
)

func init() {
	RegisterControl("switch", flowSwitch, "default")
	RegisterControl("if", flowIf, "then", "else")
}

type ControlFunction func(val yaml.Node, node yaml.Node, m, opts map[string]yaml.Node, env dynaml.Binding) yaml.Node

type Control struct {
	Name     string
	Options  map[string]struct{}
	Function ControlFunction
}

var controls = map[string]*Control{}

func RegisterControl(name string, f ControlFunction, opts ...string) {
	m := map[string]struct{}{}
	for _, o := range opts {
		m[o] = struct{}{}
	}
	c := &Control{
		Name:     name,
		Options:  m,
		Function: f,
	}
	if _, ok := controls[c.Name]; ok {
		panic(fmt.Sprintf("control or option %q already defined", c.Name))
	}
	controls[c.Name] = c
	for o := range c.Options {
		if old, ok := controls[o]; ok {
			if old != nil {
				panic(fmt.Sprintf("option %q for control %q already defined as control", o, c.Name))
			}
		}
		controls[o] = nil
	}
}

func ControlIssue(control string, node yaml.Node, msg string, args ...interface{}) yaml.Node {
	var issue yaml.Issue
	if len(args) == 0 {
		issue = yaml.NewIssue("%s", msg)
	} else {
		issue = yaml.NewIssue(msg, args...)
	}
	if control == "" {
		control = "<control>"
	} else {
		control = fmt.Sprintf("<%s control>", control)
	}
	return yaml.IssueNode(yaml.NewNode(control, node.SourceName()), true, true, issue)
}

func flowControl(node yaml.Node, env dynaml.Binding) yaml.Node {
	if m, ok := node.Value().(map[string]yaml.Node); ok && env.GetState().ControlEnabled() {
		var name string
		var val yaml.Node
		var control *Control
		opts := map[string]yaml.Node{}
		fields := map[string]yaml.Node{}
		for k, v := range m {
			if strings.HasPrefix(k, "<<") {
				n := k[2:]
				if n != "" && n != "<" && strings.Trim(n, "!") != "" {
					c, ok := controls[n]
					if !ok {
						return ControlIssue("", node, "unknown control or control option %q", k)
					}
					if c != nil {
						if control != nil {
							return ControlIssue("", node, "multiple controls %q and %q", name, k)
						}
						name = k
						control = c
						val = v
					} else {
						opts[n] = v
					}
				}
				continue
			}
			fields[k] = v
		}

		if control != nil {
			if err := control.CheckOpts(opts); err != nil {
				return ControlIssue(control.Name, node, err.Error())
			}
			return control.Function(val, node, fields, opts, env)
		}
	}
	return node
}

func (c *Control) CheckOpts(opts map[string]yaml.Node) error {
	for o := range opts {
		if _, ok := c.Options[o]; !ok {
			return fmt.Errorf("invalid option %q for control %q", o, c.Name)
		}
	}
	return nil
}

func flowSwitch(val yaml.Node, node yaml.Node, fields, opts map[string]yaml.Node, env dynaml.Binding) yaml.Node {
	switch v := val.Value().(type) {
	case dynaml.Expression:
		return node
	case string:
		sub := yaml.EmbeddedDynaml(val, env.GetState().InterpolationEnabled())
		if sub != nil {
			return node
		}
		if s, ok := fields[v]; ok {
			return s
		}
		if s, ok := opts["default"]; ok {
			return s
		}
		return ControlIssue("switch", node, "invalid switch value: %q", v)
	default:
		return ControlIssue("switch", node, "invalid switch value type: %s", dynaml.ExpressionType(v))
	}
}

func flowIf(val yaml.Node, node yaml.Node, fields, opts map[string]yaml.Node, env dynaml.Binding) yaml.Node {
	for range fields {
		return ControlIssue("if", node, "no regular fields allowed in if control")
	}
	switch v := val.Value().(type) {
	case dynaml.Expression:
		return node
	case bool:
		if v {
			if e, ok := opts["then"]; ok {
				return e
			}
			return yaml.UndefinedNode(yaml.NewNode(nil, node.SourceName()))
		} else {
			if e, ok := opts["else"]; ok {
				return e
			}
			return yaml.UndefinedNode(yaml.NewNode(nil, node.SourceName()))
		}
	default:
		return ControlIssue("if", node, "invalid condition value type: %s", dynaml.ExpressionType(v))
	}
}
