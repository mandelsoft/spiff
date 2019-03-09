package dynaml

import (
	"io/ioutil"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/yaml"
)

var fileCache = map[string][]byte{}

var templ_pattern = regexp.MustCompile(".*\\s+&template(\\(?|\\s+).*")

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
	if strings.HasSuffix(file, ".yml") || strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".json") {
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
	return parse(file, data, t, binding)
}

func parse(file string, data []byte, mode string, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()
	switch mode {
	case "template":
		n, err := yaml.Parse(file, data)
		orig := node_copy(n)

		switch v := orig.Value().(type) {
		case map[string]yaml.Node:
			if _, ok := v["<<"]; !ok {
				v["<<"] = node("(( &template ))", n)
			}
		case []yaml.Node:
			found := false
			for _, e := range v {
				if m, ok := e.Value().(map[string]yaml.Node); ok {
					if e, ok := m["<<"]; ok {
						s := yaml.EmbeddedDynaml(e)
						if s != nil && templ_pattern.MatchString(*s) {
							found = true
							break
						}
					}
				}
			}
			if !found {
				new := []yaml.Node{node(map[string]yaml.Node{"<<": node("(( &template ))", n)}, n)}
				new = append(new, v...)
				orig = node(new, n)
			}
		}
		if err != nil {
			return info.Error("error parsing file [%s]: %s", path.Clean(file), err)
		}
		result := NewTemplateValue(binding.Path(), n, orig, binding)
		return result, info, true

	case "yaml":
		node, err := yaml.Parse(file, data)
		if err != nil {
			return info.Error("error parsing file [%s]: %s", path.Clean(file), err)
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
	case "multiyaml":
		nodes, err := yaml.ParseMulti(file, data)
		if err != nil {
			return info.Error("error parsing file [%s]: %s", path.Clean(file), err)
		}
		for len(nodes) > 1 && nodes[len(nodes)-1].Value() == nil {
			nodes = nodes[:len(nodes)-1]
		}
		debug.Debug("resolving yaml list from file\n")
		info.Source = file
		result, state := binding.Flow(node(nodes, info), false)
		if state != nil {
			debug.Debug("resolving yaml file failed: " + state.Error())
			return info.PropagateError(nil, state, "resolution of yaml file '%s' failed", file)
		}
		debug.Debug("resolving yaml file succeeded")
		return result.Value(), info, true
	case "import":
		node, err := yaml.Parse(file, data)
		if err != nil {
			return info.Error("error parsing file [%s]: %s", path.Clean(file), err)
		}
		info.Source = file
		info.Raw = true
		debug.Debug("import yaml file succeeded")
		return node.Value(), info, true
	case "importmulti":
		nodes, err := yaml.ParseMulti(file, data)
		if err != nil {
			return info.Error("error parsing file [%s]: %s", path.Clean(file), err)
		}
		info.Source = file
		info.Raw = true
		for len(nodes) > 1 && nodes[len(nodes)-1].Value() == nil {
			nodes = nodes[:len(nodes)-1]
		}
		return nodes, info, true

	case "text":
		info.Source = file
		return string(data), info, true

	default:
		return info.Error("invalid file type [%s] %s", path.Clean(file), mode)
	}
}
