package dynaml

import (
	"fmt"
	"strings"

	"github.com/mandelsoft/spiff/yaml"
)

type FakeBinding struct {
	FoundFromRoot   map[string]yaml.Node
	FoundReferences map[string]yaml.Node
	FoundInStubs    map[string]yaml.Node

	path     []string
	stubPath []string
}

func (c FakeBinding) GetTempName([]byte) (string, error) {
	return "", fmt.Errorf("no temp names")
}

func (c FakeBinding) Outer() Binding {
	return nil
}

func (c FakeBinding) Path() []string {
	return c.path
}

func (c FakeBinding) StubPath() []string {
	return c.stubPath
}

func (c FakeBinding) SourceName() string {
	return "test"
}

func (c FakeBinding) RedirectOverwrite([]string) Binding {
	return c
}

func (c FakeBinding) WithScope(map[string]yaml.Node) Binding {
	return c
}

func (c FakeBinding) WithLocalScope(map[string]yaml.Node) Binding {
	return c
}

func (c FakeBinding) WithPath(step string) Binding {
	newPath := make([]string, len(c.path))
	copy(newPath, c.path)
	c.path = append(newPath, step)

	newPath = make([]string, len(c.stubPath))
	copy(newPath, c.stubPath)
	c.stubPath = append(newPath, step)
	return c
}

func (c FakeBinding) WithSource(source string) Binding {
	return c
}

func (c FakeBinding) GetStaticBinding() map[string]yaml.Node {
	return map[string]yaml.Node{}
}

func (c FakeBinding) GetRootBinding() map[string]yaml.Node {
	return c.FoundFromRoot
}

func (c FakeBinding) FindFromRoot(path []string) (yaml.Node, bool) {
	p := strings.Join(path, ".")
	if len(path) == 0 {
		p = ""
	}
	val, found := c.FoundFromRoot[p]
	return val, found
}

func (c FakeBinding) FindReference(path []string) (yaml.Node, bool) {
	val, found := c.FoundReferences[strings.Join(path, ".")]
	return val, found
}

func (c FakeBinding) FindInStubs(path []string) (yaml.Node, bool) {
	val, found := c.FoundInStubs[strings.Join(path, ".")]
	return val, found
}

func (c FakeBinding) Flow(source yaml.Node, shouldOverride bool) (yaml.Node, Status) {
	return nil, nil
}

func (c FakeBinding) Cascade(outer Binding, template yaml.Node, partial bool, templates ...yaml.Node) (yaml.Node, error) {
	return nil, nil
}
