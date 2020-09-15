// package spiffing is a wrapper for internal spiff functionality
// distributed over multiple packaes to offer a coherent interface
// for using spiff as go library

package spiffing

import (
	"os"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/flow"
	"github.com/mandelsoft/spiff/yaml"
)

type sourceBase struct {
	name string
}

func (s *sourceBase) Name() string {
	return s.name
}

type sourceFile struct {
	sourceBase
	fs vfs.FileSystem
}

// NewSourceFile returns a source based on a file in a virtual filesystem
// If no filesystem is given the os filesystem is used by default
func NewSourceFile(path string, optfs ...vfs.FileSystem) Source {
	var fs vfs.FileSystem
	if len(optfs) > 0 {
		fs = optfs[0]
	}
	if fs == nil {
		fs = osfs.New()
	}
	return &sourceFile{sourceBase{path}, fs}
}

func (s *sourceFile) Data() ([]byte, error) {
	return vfs.ReadFile(s.fs, s.name)
}

type sourceData struct {
	sourceBase
	data []byte
}

// NewSourceData creates a source based on yaml data
func NewSourceData(name string, data []byte) Source {
	return &sourceData{sourceBase{name}, data}
}

func (s *sourceData) Data() ([]byte, error) {
	return s.data, nil
}

////////////////////////////////////////////////////////////////////////////////

type spiff struct {
	key       string
	mode      int
	fs        vfs.FileSystem
	opts      flow.Options
	values    map[string]yaml.Node
	functions Functions

	binding dynaml.Binding
}

// NewFunctions provides a new registry for additional spiff functions
func NewFunctions() Functions {
	return dynaml.NewRegistry()
}

// New create a new default spiff context.
func New() Spiff {
	return &spiff{
		key:  os.Getenv("SPIFF_ENCRYPTION_KEY"),
		mode: MODE_DEFAULT,
	}
}

func (s *spiff) reset() Spiff {
	s.binding = nil
	return s
}

// WithEncryptionKey creates a new context with
// dedicated encryption key used for the spiff encryption feature
func (s spiff) WithEncryptionKey(key string) Spiff {
	s.key = key
	return s.reset()
}

// WithMode creates a new context with the given processing mode.
// (see MODE constants)
func (s spiff) WithMode(mode int) Spiff {
	if s.fs != nil {
		mode = mode & ^MODE_OS_ACCESS
	}
	s.mode = mode
	return s.reset()
}

// WithFileSystem creates a new context with the given
// virtual filesystem used for filesystem functions during
// prcessing. Setting a filesystem disables the command
// execution functions.
func (s spiff) WithFileSystem(fs vfs.FileSystem) Spiff {
	s.fs = fs
	if fs != nil {
		s.mode = s.mode & ^MODE_OS_ACCESS
	}
	return s.reset()
}

// WithFunctions creates a new context with the given
// additional function definitions
func (s spiff) WithFunctions(functions Functions) Spiff {
	s.functions = functions
	return s.reset()
}

// WithValues creates a new context with the given
// additional structured values usable by path expressions
// during processing.
// It is highly recommended to decide for a common root
// value (like `values`) to minimize the blocked root
// elements in the processed documents.
func (s spiff) WithValues(values map[string]interface{}) (Spiff, error) {
	if values != nil {
		nodes, err := yaml.Sanitize("values", values)
		if err != nil {
			return nil, err
		}
		s.values = nodes.Value().(map[string]yaml.Node)
	} else {
		s.values = nil
	}
	return s.reset(), nil
}

// FileSystem return the virtual filesystem set for the execution context.
func (s *spiff) FileSystem() vfs.FileSystem {
	return s.fs
}

// FileSource create a new file source based on the configured file system.
func (s *spiff) FileSource(path string) Source {
	return NewSourceFile(path, s.fs)
}

// Cascade processes a template with a list of given subs and state
// documents
func (s *spiff) Cascade(template Node, stubs []Node, states ...Node) (Node, error) {
	if s.binding == nil {
		s.binding = flow.NewEnvironment(
			nil, "context", flow.NewState(s.key, s.mode, s.fs).SetFunctions(s.functions))
		if s.values != nil {
			s.binding = s.binding.WithLocalScope(s.values)
		}
	}
	return flow.Cascade(s.binding, template, s.opts, append(stubs, states...)...)
}

// PrepareStubs processes a list a stubs and returns a prepared
// represenation usable to process a template
func (s *spiff) PrepareStubs(stubs ...Node) ([]Node, error) {
	return flow.PrepareStubs(s.binding, s.opts.Partial, stubs...)
}

// ApplyStubs uses already prepared subs to process a template.
func (s *spiff) ApplyStubs(template Node, preparedstubs []Node) (Node, error) {
	return flow.Apply(s.binding, template, preparedstubs, s.opts)
}

// Unmarshal parses a single document yaml representation and
// returns the internal representation
func (s *spiff) Unmarshal(name string, source []byte) (Node, error) {
	return yaml.Unmarshal(name, source)
}

// Unmarshal parses a single source and
// returns the internal representation
func (s *spiff) UnmarshalSource(source Source) (Node, error) {
	data, err := source.Data()
	if err != nil {
		return nil, err
	}
	return yaml.Unmarshal(source.Name(), data)
}

// UnmarshalMulti parses a multi document yaml representation and
// returns the list of documents in the internal representation
func (s *spiff) UnmarshalMulti(name string, source []byte) ([]Node, error) {
	return yaml.UnmarshalMulti(name, source)
}

// UnmarshalMulti parses a multi document source and
// returns the list of documents in the internal representation
func (s *spiff) UnmarshalMultiSource(source Source) ([]Node, error) {
	data, err := source.Data()
	if err != nil {
		return nil, err
	}
	return yaml.UnmarshalMulti(source.Name(), data)
}

// DetermineState extracts the intended new state representation from
// a processing result.
func (s *spiff) DetermineState(node Node) Node {
	return flow.DetermineState(node)
}

// Marshal transform the internal node representation into a
// yaml representation
func (s *spiff) Marshal(node Node) ([]byte, error) {
	return yaml.Marshal(node)
}

// Normalize transform the node representation to a regular go value representation
// consisting of map[string]interface{}`, `[]interface{}`, `string `boolean`,
// `int64`, `float64` and []byte objects
func (s *spiff) Normalize(node Node) (interface{}, error) {
	return yaml.Normalize(node)
}
