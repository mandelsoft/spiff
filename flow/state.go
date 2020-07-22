package flow

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"github.com/mandelsoft/spiff/debug"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

type State struct {
	files     map[string]string // content hash to temp file name
	fileCache map[string][]byte // file content cache
	key       string            // default encryption key
	osaccess  bool              // allow OS access
}

func NewState(key string, osaccess bool) *State {
	return &State{map[string]string{}, map[string][]byte{}, key, osaccess}
}

func (s *State) OSAccessAllowed() bool {
	return s.osaccess
}

func (s *State) GetEncryptionKey() string {
	return s.key
}

func (s *State) GetTempName(data []byte) (string, error) {
	sum := sha512.Sum512(data)
	hash := base64.StdEncoding.EncodeToString(sum[:])

	name, ok := s.files[hash]
	if !ok {
		file, err := ioutil.TempFile("", "spiff-")
		if err != nil {
			return "", err
		}
		name = file.Name()
		s.files[hash] = name
	}
	return name, nil
}

func (s *State) Cleanup() {
	for _, n := range s.files {
		os.Remove(n)
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
			data, err = ioutil.ReadFile(file)
			if err != nil {
				return nil, fmt.Errorf("error reading [%s]: %s", path.Clean(file), err)
			}
		}
		s.fileCache[file] = data
	}
	return data, nil
}
