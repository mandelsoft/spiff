package spiffing

import (
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/flow"
	"github.com/mandelsoft/spiff/yaml"
)

// MODE_OS_ACCESS allows os command execution (pipe, exec)
const MODE_OS_ACCESS = flow.MODE_OS_ACCESS

//  MODE_FILE_ACCESS allows file access to virtual filesystem
const MODE_FILE_ACCESS = flow.MODE_FILE_ACCESS

// MODE_DEFAULT (default) enables all os related spiff functions
const MODE_DEFAULT = MODE_OS_ACCESS | MODE_FILE_ACCESS

// Node is a document node of the processing representation of a document
type Node = yaml.Node

// Options described the processing options
type Options = flow.Options

// Functions provides access to a set of spiff functions used to extend
// the standrd function set
type Functions = dynaml.Registry

// Spiff is a configuration end execution context for
// executing spiff operations
type Spiff interface {
	WithEncryptionKey(key string) Spiff
	WithMode(mode int) Spiff
	WithFileSystem(fs vfs.FileSystem) Spiff
	WithFunctions(functions Functions) Spiff
	WithValues(values map[string]interface{}) (Spiff, error)

	FileSystem() vfs.FileSystem
	FileSource(path string) Source

	Unmarshal(name string, source []byte) (Node, error)
	UnmarshalSource(source Source) (Node, error)
	Marshal(node Node) ([]byte, error)
	DetermineState(node Node) Node
	Normalize(node Node) (interface{}, error)

	Cascade(template Node, stubs []Node, states ...Node) (Node, error)
	PrepareStubs(stubs ...Node) ([]Node, error)
	ApplyStubs(template Node, preparedstubs []Node) (Node, error)
}

// Source is used to get access to a template or stub source data and name
type Source interface {
	Name() string
	Data() ([]byte, error)
}
