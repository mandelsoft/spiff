package flow

import (
	"crypto/sha512"
	"encoding/base64"
	"io/ioutil"
	"os"
)

type State struct {
	files map[string]string
}

func NewState() *State {
	return &State{map[string]string{}}
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
