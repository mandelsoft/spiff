// package spiffing is a wrapper for internal spiff functionality
// distributed over multiple packaes to offer a coherent interface
// for using spiff as go library

package spiffing

import (
	"os"

	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/flow"
	"github.com/mandelsoft/spiff/yaml"
)

const MODE_OS_ACCESS = flow.MODE_OS_ACCESS
const MODE_FILE_ACCESS = flow.MODE_FILE_ACCESS

const MODE_DEFAULT = MODE_OS_ACCESS | MODE_FILE_ACCESS

type Node = yaml.Node
type Options = flow.Options

type Spiff interface {
	WithEncryptionKey(key string) Spiff
	WithMode(mode int) Spiff
	WithFileSystem(fs vfs.FileSystem) Spiff

	Unmarshal(name string, source []byte) (Node, error)
	Marshal(node Node) ([]byte, error)
	DetermineState(node Node) Node

	Cascade(template Node, stubs []Node, states ...Node) (Node, error)
	PrepareStubs(stubs ...Node) ([]Node, error)
	ApplyStubs(template Node, preparedstubs []Node) (Node, error)
}

type spiff struct {
	key  string
	mode int
	fs   vfs.FileSystem

	opts    flow.Options
	binding dynaml.Binding
}

func New() *spiff {
	return &spiff{
		key:  os.Getenv("SPIFF_ENCRYPTION_KEY"),
		mode: MODE_DEFAULT,
	}
}

func (s spiff) WithEncryptionKey(key string) Spiff {
	s.key = key
	return &s
}

func (s spiff) WithMode(mode int) Spiff {
	s.mode = mode
	return &s
}

func (s spiff) WithFileSystem(fs vfs.FileSystem) Spiff {
	s.fs = fs
	return &s
}

func (s *spiff) Cascade(template Node, stubs []Node, states ...Node) (Node, error) {
	if s.binding == nil {
		s.binding = flow.NewEnvironment(nil, "context", flow.NewState(s.key, s.mode, s.fs))
	}
	return flow.Cascade(s.binding, template, s.opts, append(stubs, states...)...)
}

func (s *spiff) PrepareStubs(stubs ...Node) ([]Node, error) {
	return flow.PrepareStubs(s.binding, s.opts.Partial, stubs...)
}

func (s *spiff) ApplyStubs(template Node, preparedstubs []Node) (Node, error) {
	return flow.Apply(s.binding, template, preparedstubs, s.opts)
}

func (s *spiff) Unmarshal(name string, source []byte) (Node, error) {
	return yaml.Unmarshal(name, source)
}

func (s *spiff) UnmarshalMulti(name string, source []byte) ([]Node, error) {
	return yaml.UnmarshalMulti(name, source)
}

func (s *spiff) DetermineState(node Node) Node {
	return flow.DetermineState(node)
}

func (s *spiff) Marshal(node Node) ([]byte, error) {
	return yaml.Marshal(node)
}
