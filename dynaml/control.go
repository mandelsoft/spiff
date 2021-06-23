package dynaml

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/spiff/yaml"
)

type ControlFunction func(val yaml.Node, node yaml.Node, m, opts map[string]yaml.Node, env Binding) (yaml.Node, bool)

type Control struct {
	Name     string
	Options  map[string]bool
	Function ControlFunction
}

var controls = map[string]*Control{}
var templateoptions = map[string]struct{}{}

func RegisterControl(name string, f ControlFunction, opts ...string) {
	m := map[string]bool{}
	for _, o := range opts {
		t := false
		if strings.HasPrefix(o, "*") {
			t = true
			o = o[1:]
		}
		m[o] = t

		if _, ok := templateoptions[o]; ok && !t {
			panic(fmt.Sprintf("ambigious control option template setting for %q", o))
		}
		if t {
			templateoptions[o] = struct{}{}
		}
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
	for _, o := range opts {
		if strings.HasPrefix(o, "*") {
			o = o[1:]
		}
		if old, ok := controls[o]; ok {
			if old != nil {
				panic(fmt.Sprintf("option %q for control %q already defined as control", o, c.Name))
			}
		}
		controls[o] = nil
	}
}

func ControlIssue(control string, node yaml.Node, msg string, args ...interface{}) (yaml.Node, bool) {
	var issue yaml.Issue
	if len(args) == 0 {
		issue = yaml.NewIssue("%s", msg)
	} else {
		issue = yaml.NewIssue(msg, args...)
	}
	return ControlIssueByIssue(control, node, issue, true)
}

func ControlIssueByIssue(control string, node yaml.Node, issue yaml.Issue, final bool) (yaml.Node, bool) {
	if control == "" {
		control = "<control>"
	} else {
		control = fmt.Sprintf("<%s control>", control)
	}
	if !final {
		return yaml.IssueNode(node, true, true, issue), false
	}
	return yaml.IssueNode(yaml.NewNode(control, node.SourceName()), true, true, issue), false
}

func IsControl(val interface{}, env Binding) (bool, error) {
	if n, ok := val.(map[string]yaml.Node); ok {
		c, _, _, _, err := GetControl(n, env)
		return c != nil, err
	}
	return false, nil
}

func RequireTemplate(opt string) bool {
	if strings.HasPrefix(opt, "<<") {
		_, ok := templateoptions[opt[2:]]
		return ok
	}
	return false
}

func GetControl(m map[string]yaml.Node, env Binding) (*Control, yaml.Node, map[string]yaml.Node, map[string]yaml.Node, error) {
	if env.GetFeatures().ControlEnabled() {
		var name string
		var val yaml.Node
		var control *Control
		opts := map[string]yaml.Node{}
		fields := map[string]yaml.Node{}
		for k, v := range m {
			if strings.HasPrefix(k, "<<") {
				n := k[2:]
				if n != "" && n != "<" && n[0] != '!' {
					c, ok := controls[n]
					if !ok {
						return nil, nil, nil, nil, fmt.Errorf("unknown control or control option %q", k)
					}
					if c != nil {
						if control != nil {
							return nil, nil, nil, nil, fmt.Errorf("multiple controls %q and %q", name, k)
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
			return control, val, fields, opts, control.CheckOpts(opts)
		}
	}
	return nil, nil, nil, nil, nil
}

func (c *Control) CheckOpts(opts map[string]yaml.Node) error {
	for o := range opts {
		if _, ok := c.Options[o]; !ok {
			return fmt.Errorf("invalid option %q for control %q", o, c.Name)
		}
	}
	return nil
}
