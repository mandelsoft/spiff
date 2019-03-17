package dynaml

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"github.com/mandelsoft/spiff/yaml"
	"io"
	"sort"
	"strings"
)

func func_archive(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) < 1 || len(arguments) > 2 {
		return info.Error("archive takes one or two arguments")
	}

	mode := "tar"

	if len(arguments) == 2 {
		str, ok := arguments[1].(string)
		if !ok {
			return info.Error("second argument for hash must be a string")
		}
		mode = str
	}

	data, ok := arguments[0].(map[string]yaml.Node)
	if !ok {
		return info.Error("first argument for hash must be a file map ")
	}

	files := map[string][]byte{}
	for file, val := range data {
		if val == nil {
			continue
		}
		var content []byte
		switch v := val.Value().(type) {
		case string:
			content = []byte(v)
			if hasTag(file, "#") {
				b, err := base64.StdEncoding.DecodeString(v)
				if err == nil {
					content = b
				}
			}
		case int64:
			content = []byte(fmt.Sprintf("%d", v))
		case bool:
			content = []byte(fmt.Sprintf("%t", v))
		case map[string]yaml.Node, []yaml.Node:
			c, infoa, ok := func_as_yaml([]interface{}{v}, binding)
			if !ok {
				return info.Error("cannot convert %s: %s", file, infoa.Issue.Issue)
			}
			content = []byte(c.(string))
		default:
			return info.Error("invalid file content type %s", ExpressionType(v))
		}
		files[file] = content
	}

	var buf bytes.Buffer
	var err error
	switch mode {
	case "targz":
		zipper := gzip.NewWriter(&buf)
		err = tar_archive(zipper, files)
		zipper.Close()
		if err != nil {
			return info.Error("archiving %s failed: %s", mode, err)
		}
	case "tar":
		err = tar_archive(&buf, files)
		if err != nil {
			return info.Error("archiving %s failed: %s", mode, err)
		}
	default:
		return info.Error("invalid archive type '%s'", mode)
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), info, true
}

func getSortedMapKeys(unsortedMap map[string][]byte) []string {
	keys := make([]string, len(unsortedMap))
	i := 0
	for k, _ := range unsortedMap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

func hasTag(file, tag string) bool {
	for strings.HasPrefix(file, "*") || strings.HasPrefix(file, "#") || strings.HasPrefix(file, "-") {
		if strings.HasPrefix(file, tag) {
			return true
		}
		stop := strings.HasPrefix(file, "-")
		file = file[1:]
		if stop {
			break
		}
	}
	return false
}

func tar_archive(w io.Writer, files map[string][]byte) error {
	tw := tar.NewWriter(w)
	defer tw.Close()
	keys := getSortedMapKeys(files)
	for _, file := range keys {
		mode := int64(0600)
		if hasTag(file, "*") {
			mode = 0744
		}
		for strings.HasPrefix(file, "*") || strings.HasPrefix(file, "#") || strings.HasPrefix(file, "-") {
			stop := strings.HasPrefix(file, "-")
			file = file[1:]
			if stop {
				break
			}
		}
		header := &tar.Header{
			Name: file,
			Mode: mode,
			Size: int64(len(files[file])),
		}
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if _, err := tw.Write([]byte(files[file])); err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return nil
}
