package spiffing

import (
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
