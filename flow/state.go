package flow

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"

	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/features"
	"github.com/mandelsoft/spiff/yaml"
)

const MODE_FILE_ACCESS = 1 // support file system access
const MODE_OS_ACCESS = 2   // support os commands like pipe and exec

type State struct {
	files         map[string]string // content hash to temp file name
	fileCache     map[string][]byte // file content cache
	key           string            // default encryption key
	mode          int
	fileSystem    vfs.VFS // virtual filesystem to use for filesystem based operations
	functions     dynaml.Registry
	interpolation bool
	tags          map[string]*dynaml.Tag
	docno         int // document number
}

var _ dynaml.State = &State{}

func NewState(key string, mode int, optfs ...vfs.FileSystem) *State {
	var fs vfs.FileSystem
	if len(optfs) > 0 {
		fs = optfs[0]
	}
	if fs == nil {
		fs = osfs.New()
	} else {
		mode = mode & ^MODE_OS_ACCESS
	}
	return &State{
		tags:       map[string]*dynaml.Tag{},
		files:      map[string]string{},
		fileCache:  map[string][]byte{},
		key:        key,
		mode:       mode,
		fileSystem: vfs.New(fs),
		docno:      1,
	}
}

func NewDefaultState() *State {
	return NewState(features.EncryptionKey(), MODE_OS_ACCESS|MODE_FILE_ACCESS)
}

func (s *State) SetFunctions(f dynaml.Registry) *State {
	s.functions = f
	return s
}

func (s *State) SetTags(tags ...*dynaml.Tag) *State {
	s.tags = map[string]*dynaml.Tag{}
	for _, v := range tags {
		s.tags[v.Name()] = v
	}
	return s
}

func (s *State) EnableInterpolation() {
	s.interpolation = true
}

func (s *State) SetInterpolation(b bool) *State {
	s.interpolation = b
	return s
}

func (s *State) InterpolationEnabled() bool {
	return s.interpolation
}

func (s *State) OSAccessAllowed() bool {
	return s.mode&MODE_OS_ACCESS != 0
}

func (s *State) FileAccessAllowed() bool {
	return s.mode&MODE_FILE_ACCESS != 0
}

func (s *State) FileSystem() vfs.VFS {
	return s.fileSystem
}

func (s *State) GetFunctions() dynaml.Registry {
	return s.functions
}

func (s *State) GetEncryptionKey() string {
	return s.key
}

func (s *State) GetTempName(data []byte) (string, error) {
	if !s.FileAccessAllowed() {
		return "", fmt.Errorf("tempname: no OS operations supported in this execution environment")
	}
	sum := sha512.Sum512(data)
	hash := base64.StdEncoding.EncodeToString(sum[:])

	name, ok := s.files[hash]
	if !ok {
		file, err := s.fileSystem.TempFile("", "spiff-")
		if err != nil {
			return "", err
		}
		name = file.Name()
		s.files[hash] = name
	}
	return name, nil
}

func (s *State) SetTag(name string, node yaml.Node, path []string) error {
	debug.Debug("setting tag: %v\n", path)
	old := s.tags[name]
	if old != nil {
		if old.Scope() != dynaml.TAG_LOCAL {
			return fmt.Errorf("duplicate tag %q: %s in foreign document", name, strings.Join(path, "."))
		}
		if !reflect.DeepEqual(path, old.Path()) {
			return fmt.Errorf("duplicate tag %q: %s <-> %s", name, strings.Join(path, "."), strings.Join(old.Path(), "."))
		}
	}
	s.tags[name] = dynaml.NewTag(name, Cleanup(node, discardTags), path, false)
	return nil
}

func (s *State) GetTag(name string) *dynaml.Tag {
	i, err := strconv.Atoi(name)
	if err == nil {
		if i <= 0 {
			i += s.docno
			if i <= 0 {
				return nil
			}
			name = fmt.Sprintf("%d", i)
		}
	}
	return s.tags[name]
}

func (s *State) ResetTags() {
	s.tags = map[string]*dynaml.Tag{}
	s.docno = 1
}

func (s *State) ResetStream() {
	n := map[string]*dynaml.Tag{}
	for _, v := range s.tags {
		if v.Scope() == dynaml.TAG_GLOBAL {
			n[v.Name()] = v
		}
	}
	s.docno = 1
	s.tags = n
}

func (s *State) PushDocument(node yaml.Node) {
	for _, t := range s.tags {
		t.ResetLocal()
	}
	if node != nil {
		s.SetTag(fmt.Sprintf("%d", s.docno), node, nil)
	}
	s.docno++
}

func (s *State) Cleanup() {
	for _, n := range s.files {
		s.fileSystem.Remove(n)
	}
	s.files = map[string]string{}
}

func (s *State) GetFileContent(file string, cached bool) ([]byte, error) {
	var err error

	data := s.fileCache[file]
	if !cached || data == nil {
		debug.Debug("reading file %s\n", file)
		if strings.HasPrefix(file, "http:") || strings.HasPrefix(file, "https:") {
			response, err := http.Get(file)
			if err != nil {
				return nil, fmt.Errorf("error getting [%s]: %s", file, err)
			} else {
				defer response.Body.Close()
				contents, err := ioutil.ReadAll(response.Body)
				if err != nil {
					return nil, fmt.Errorf("error getting body [%s]: %s", file, err)
				}
				data = contents
			}
		} else {
			data, err = s.fileSystem.ReadFile(file)
			if err != nil {
				return nil, fmt.Errorf("error reading [%s]: %s", path.Clean(file), err)
			}
		}
		s.fileCache[file] = data
	}
	return data, nil
}
