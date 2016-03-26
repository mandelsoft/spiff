package dynaml

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/cloudfoundry-incubator/spiff/debug"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

var fileCache = map[string][]byte{}

func func_read(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) > 2 {
		return info.Error("read takes a maximum of two arguments")
	}

	file, ok := arguments[0].(string)
	if !ok {
		return info.Error("string value requiredfor file path")
	}

	t := "text"
	if strings.HasSuffix(file, ".yml") {
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
	if data == nil {
		debug.Debug("reading %s file %s\n", t, file)
		data, err = ioutil.ReadFile(file)
		if err != nil {
			return info.Error("error reading [%s]: %s", path.Clean(file), err)
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
			return info.Error("yaml file resolution failed")
		}
		debug.Debug("resolving yaml file succeeded")
		info.Source = file
		return result.Value(), info, true

	case "text":
		info.Source = file
		return string(data), info, true

	default:
		return info.Error("invalid file type [%s] %s", path.Clean(file), t)
	}
}
