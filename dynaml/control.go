package dynaml

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/spiff/yaml"
)

type ControlFunction func(val yaml.Node, node yaml.Node, m, opts map[string]yaml.Node, env Binding) (yaml.Node, bool)

type Control struct {
	name     string
	options  map[string]bool
	function ControlFunction
}

func (c *Control) Name() string {
	return c.name
}

func (c *Control) Options() []string {
	l := []string{}
	for n := range c.options {
		l = append(l, n)
	}
	return l
}

func (c *Control) HasOption(name string) bool {
	_, ok := c.options[name]
	return ok
}

func (c *Control) IsTemplateOption(name string) bool {
	return c.options[name]
}

func (c *Control) Function(val yaml.Node, node yaml.Node, m, opts map[string]yaml.Node, env Binding) (yaml.Node, bool) {
	return c.function(val, node, m, opts, env)
}

type Controls interface {
	RegisterControl(name string, f ControlFunction, opts ...string) error
	LookupControl(name string) (*Control, bool)
	IsTemplateControlOption(name string) bool
}

type controlRegistry struct {
	controls        map[string]*Control
	templateoptions map[string]struct{}
}

func newControls() *controlRegistry {
	return &controlRegistry{map[string]*Control{}, map[string]struct{}{}}
}

func NewControls() Controls {
	r := newControls()

	for n, c := range control_registry.controls {
		r.controls[n] = c
	}
	for n := range control_registry.templateoptions {
		r.templateoptions[n] = struct{}{}
	}
	return r
}

func (r *controlRegistry) RegisterControl(name string, f ControlFunction, opts ...string) error {
	m := map[string]bool{}
	for _, o := range opts {
		t := false
		if strings.HasPrefix(o, "*") {
			t = true
			o = o[1:]
		}
		m[o] = t

		if _, ok := r.templateoptions[o]; ok && !t {
			return fmt.Errorf("ambigious control option template setting for %q", o)
		}
		if t {
			r.templateoptions[o] = struct{}{}
		}
	}
	c := &Control{
		name:     name,
		options:  m,
		function: f,
	}
	if _, ok := r.controls[c.name]; ok {
		return fmt.Errorf("control or option %q already defined", c.name)
	}
	r.controls[c.name] = c
	for _, o := range opts {
		if strings.HasPrefix(o, "*") {
			o = o[1:]
		}
		if old, ok := r.controls[o]; ok {
			if old != nil {
				return fmt.Errorf("option %q for control %q already defined as control", o, c.name)
			}
		}
		r.controls[o] = nil
	}
	return nil
}

func (r *controlRegistry) LookupControl(name string) (*Control, bool) {
	c, ok := r.controls[name]
	return c, ok
}

func (r *controlRegistry) IsTemplateControlOption(name string) bool {
	_, ok := r.templateoptions[name]
	return ok
}

var control_registry = newControls()

func RegisterControl(name string, f ControlFunction, opts ...string) {
	err := control_registry.RegisterControl(name, f, opts...)
	if err != nil {
		panic(err.Error())
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

func RequireTemplate(opt string, env Binding) bool {
	registry := env.GetState().GetRegistry()
	if strings.HasPrefix(opt, "<<") {
		return registry.IsTemplateControlOption(opt[2:])
	}
	return false
}

func GetControl(m map[string]yaml.Node, env Binding) (*Control, yaml.Node, map[string]yaml.Node, map[string]yaml.Node, error) {
	if env.GetFeatures().ControlEnabled() {
		registry := env.GetState().GetRegistry()
		var name string
		var val yaml.Node
		var control *Control
		opts := map[string]yaml.Node{}
		fields := map[string]yaml.Node{}
		for k, v := range m {
			if strings.HasPrefix(k, "<<") {
				n := k[2:]
				if n != "" && n != "<" && n[0] != '!' {
					c, ok := registry.LookupControl(n)
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
		if _, ok := c.options[o]; !ok {
			return fmt.Errorf("invalid option %q for control %q", o, c.name)
		}
	}
	return nil
}
