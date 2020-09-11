package spiffing

import (
	"strings"
)

// Process just processes a template with the values set in the execution
// context. It directly takes and delivers byte array containing yaml data.
func Process(s Spiff, template Source) ([]byte, error) {
	templ, err := s.UnmarshalSource(template)
	if err != nil {
		return nil, err
	}
	result, err := s.Cascade(templ, nil)
	if err != nil {
		return nil, err
	}
	return s.Marshal(result)
}

// ProcessFile just processes a template give by a file with the values set in
// the execution context.
// The path name of the file is interpreted in the context of the filesystem
// found in the execution context, which is defaulted by the OS filesystem.
func ProcessFile(s Spiff, path string) ([]byte, error) {
	return Process(s, s.FileSource(path))
}

// ExecuteDynamlExpression just processes a plain dynaml expression with the values set in
// the execution context.
func EvaluateDynamlExpression(s Spiff, expr string) ([]byte, error) {
	r, err := Process(s, NewSourceData("dynaml", []byte("(( "+expr+" ))")))
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(r), "\n")
	if len(lines) == 2 && lines[1] == "" {
		return []byte(lines[0]), nil
	}
	return r, nil
}
