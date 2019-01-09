package dynaml

import (
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/yaml"
)

var fileCache = map[string][]byte{}

func func_read(cached bool, arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) > 2 {
		return info.Error("read takes a maximum of two arguments")
	}

	file, ok := arguments[0].(string)
	if !ok {
		return info.Error("string value required for file path")
	}

	t := "text"
	if strings.HasSuffix(file, ".yml") || strings.HasSuffix(file, ".yaml") {
		t = "yaml"
	}
	if len(arguments) > 1 {
		t, ok = arguments[1].(string)
		if !ok {
			return info.Error("string value required for type")
		}

	}

	var err error

	data := fileCache[file]
	if !cached || data == nil {
		debug.Debug("reading %s file %s\n", t, file)
		if strings.HasPrefix(file, "http:") || strings.HasPrefix(file, "https:") {
			response, err := http.Get(file)
			if err != nil {
				return info.Error("error getting [%s]: %s", file, err)
			} else {
				defer response.Body.Close()
				contents, err := ioutil.ReadAll(response.Body)
				if err != nil {
					return info.Error("error getting body [%s]: %s", file, err)
				}
				data = contents
			}
		} else {
			data, err = ioutil.ReadFile(file)
			if err != nil {
				return info.Error("error reading [%s]: %s", path.Clean(file), err)
			}
		}
		fileCache[file] = data
	}

	switch t {
	case "yaml":
		node, err := yaml.Parse(file, data)
		if err != nil {
			return info.Error("error parsing stub [%s]: %s", path.Clean(file), err)
		}
		debug.Debug("resolving yaml file\n")
		result, state := binding.Flow(node, false)
		if state != nil {
			debug.Debug("resolving yaml file failed: " + state.Error())
			return info.PropagateError(nil, state, "resolution of yaml file '%s' failed", file)
		}
		debug.Debug("resolving yaml file succeeded")
		info.Source = file
		return result.Value(), info, true
	case "import":
		node, err := yaml.Parse(file, data)
		if err != nil {
			return info.Error("error parsing stub [%s]: %s", path.Clean(file), err)
		}
		info.Source = file
		return node.Value(), info, true

	case "text":
		info.Source = file
		return string(data), info, true

	default:
		return info.Error("invalid file type [%s] %s", path.Clean(file), t)
	}
}
