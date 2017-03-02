package flow

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mandelsoft/spiff/debug"
	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/yaml"
)

func Flow(source yaml.Node, stubs ...yaml.Node) (yaml.Node, error) {
	return NewEnvironment(stubs, source.SourceName()).Flow(source, true)
}

func get_inherited_flags(env dynaml.Binding) yaml.NodeFlags {
	overridden, found := env.FindInStubs(env.StubPath())
	if found {
		return overridden.Flags() & yaml.FLAG_TEMPORARY
	}
	return 0
}

func flow(root yaml.Node, env dynaml.Binding, shouldOverride bool) yaml.Node {
	if root == nil {
		return root
	}

	flags := root.Flags()
	replace := root.ReplaceFlag()
	redirect := root.RedirectPath()
	preferred := root.Preferred()
	merged := root.Merged()
	keyName := root.KeyName()
	source := root.SourceName()

	if redirect != nil {
		env = env.RedirectOverwrite(redirect)
	}

	debug.Debug("/// FLOW %v: %+v\n", env.Path(), root)
	if !replace {
		if _, ok := root.Value().(dynaml.Expression); !ok && merged {
			debug.Debug("  skip handling of merged node")
			return root
		}
		switch val := root.Value().(type) {
		case map[string]yaml.Node:
			return flowMap(root, env)

		case []yaml.Node:
			return flowList(root, env)

		case dynaml.Expression:
			debug.Debug("??? eval %T: %+v\n", val, val)
			env := env
			if root.SourceName() != env.SourceName() {
				env = env.WithSource(root.SourceName())
			}
			info := dynaml.DefaultInfo()
			var eval interface{} = nil
			m, ok := val.(dynaml.MarkerExpr)
			if ok && m.Has(dynaml.TEMPLATE) {
				debug.Debug("found template declaration\n")
				val := m.TemplateExpression(root)
				if val == nil {
					root = yaml.IssueNode(root, true, false, yaml.NewIssue("empty template value"))
					debug.Debug("??? failed ---> KEEP\n")
					if !shouldOverride {
						return root
					}
				}
				debug.Debug("  value template %s", val)
				eval = dynaml.TemplateValue{env.Path(), val, root}
			} else {
				eval, info, ok = val.Evaluate(env, false)
			}
			replace = replace || info.Replace
			flags |= info.NodeFlags
			debug.Debug("??? ---> %+v\n", eval)
			if !ok {
				root = yaml.IssueNode(root, true, false, info.Issue)
				debug.Debug("??? failed ---> KEEP\n")
				if !shouldOverride {
					return root
				}
			} else {
				if info.SourceName() != "" {
					source = info.SourceName()
				}
				result := yaml.NewNode(eval, source)
				_, ok = eval.(string)
				if ok {
					// map result to potential expression
					result = flowString(result, env)
				}
				_, expr := result.Value().(dynaml.Expression)

				if len(info.Issue.Issue) != 0 {
					result = yaml.IssueNode(result, false, info.Failed, info.Issue)
				}
				if info.Undefined {
					debug.Debug("   UNDEFINED")
					result = yaml.UndefinedNode(result)
				}
				// preserve accumulated node attributes
				if preferred || info.Preferred {
					debug.Debug("   PREFERRED")
					result = yaml.PreferredNode(result)
				}

				if info.KeyName != "" {
					keyName = info.KeyName
					result = yaml.KeyNameNode(result, keyName)
				}
				if len(info.RedirectPath) > 0 {
					redirect = info.RedirectPath
				}
				if len(redirect) > 0 {
					debug.Debug("   REDIRECT -> %v\n", redirect)
					result = yaml.RedirectNode(result.Value(), result, redirect)
				}

				if replace {
					debug.Debug("   REPLACE\n")
					result = yaml.ReplaceNode(result.Value(), result, redirect)
				} else {
					if merged || info.Merged {
						debug.Debug("   MERGED\n")
						result = yaml.MergedNode(result)
					}
				}
				if (flags | result.Flags()) != result.Flags() {
					result = yaml.AddFlags(result, flags)
				}
				if expr || result.Merged() || !shouldOverride || result.Preferred() {
					debug.Debug("   prefer expression over override")
					debug.Debug("??? ---> %+v\n", result)
					return result
				}
				debug.Debug("???   try override\n")
				replace = result.ReplaceFlag()
				root = result
			}

		case string:
			result := flowString(root, env)
			if result != nil {
				_, ok := result.Value().(dynaml.Expression)
				if ok {
					// analyse expression before overriding
					return result
				}
			}
		}
	}

	if !merged && root.StandardOverride() && shouldOverride {
		debug.Debug("/// lookup stub %v -> %v\n", env.Path(), env.StubPath())
		overridden, found := env.FindInStubs(env.StubPath())
		if found {
			root = overridden
			if keyName != "" {
				root = yaml.KeyNameNode(root, keyName)
			}
			if replace {
				root = yaml.ReplaceNode(root.Value(), root, redirect)
			} else {
				if redirect != nil {
					root = yaml.RedirectNode(root.Value(), root, redirect)
				} else {
					if merged {
						root = yaml.MergedNode(root)
					}
				}
			}
			root = yaml.AddFlags(root, flags)
		}
	}

	return root
}

/*
 * compatibility issue. A single merge node was always optional
 * means: <<: (( merge )) == <<: (( merge || nil ))
 * the first pass, just parses the dynaml
 * only the second pass, evaluates a dynaml node!
 */
func simpleMergeCompatibilityCheck(initial bool, node yaml.Node) bool {
	if !initial {
		merge, ok := node.Value().(dynaml.MergeExpr)
		return ok && !merge.Required
	}
	return false
}

func flowMap(root yaml.Node, env dynaml.Binding) yaml.Node {
	var flags yaml.NodeFlags
	flags = get_inherited_flags(env)
	processed := true
	template := false
	rootMap := root.Value().(map[string]yaml.Node)

	env = env.WithScope(rootMap)

	redirect := root.RedirectPath()
	replace := root.ReplaceFlag()
	newMap := make(map[string]yaml.Node)

	sortedKeys := getSortedKeys(rootMap)

	debug.Debug("HANDLE MAP %v\n", env.Path())

	// iteration order matters for the "<<" operator, it must be the first key in the map that is handled
	for i := range sortedKeys {
		key := sortedKeys[i]
		val := rootMap[key]

		if key == "<<" {
			_, initial := val.Value().(string)
			base := flow(val, env, false)
			debug.Debug("flow to %#v\n", base.Value())
			_, ok := base.Value().(dynaml.Expression)
			if ok {
				m, ok := base.Value().(dynaml.MarkerExpr)
				if ok {
					debug.Debug("found marker\n")
					flags |= m.GetFlags()
					if flags.Temporary() {
						debug.Debug("found temporary declaration\n")
					}
					if flags.Local() {
						debug.Debug("found local declaration\n")
					}
				}
				if ok && m.Has(dynaml.TEMPLATE) {
					debug.Debug("found template declaration\n")
					processed = false
					template = true
					val = m.TemplateExpression(root)
					if val == nil {
						continue
					}
					debug.Debug("  insert expression: %v\n", val)
				} else {
					if simpleMergeCompatibilityCheck(initial, base) {
						continue
					}
					val = base
				}
				processed = false
			} else {
				baseMap, ok := base.Value().(map[string]yaml.Node)
				if base != nil && base.RedirectPath() != nil {
					redirect = base.RedirectPath()
					env = env.RedirectOverwrite(redirect)
				}
				if ok {
					for k, v := range baseMap {
						newMap[k] = v
					}
				}
				replace = base.ReplaceFlag()
				if ok || base.Value() == nil || yaml.EmbeddedDynaml(base) == nil {
					// still ignore non dynaml value (might be strange but compatible)
					if replace {
						break
					}
					continue
				} else {
					val = base
				}
			}
		} else {
			if processed {
				val = flow(val, env.WithPath(key), true)
			}
		}

		debug.Debug("MAP %v (%s)%s\n", env.Path(), val.KeyName(), key)
		if !val.Undefined() {
			newMap[key] = val
		}
	}

	debug.Debug("MAP DONE %v\n", env.Path())
	var result interface{}
	if template {
		debug.Debug(" as template\n")
		result = dynaml.TemplateValue{env.Path(), yaml.NewNode(newMap, root.SourceName()), root}
	} else {
		result = newMap
	}
	var node yaml.Node
	if replace {
		node = yaml.ReplaceNode(result, root, redirect)
	} else {
		node = yaml.RedirectNode(result, root, redirect)
	}
	if (flags | node.Flags()) != node.Flags() {
		node = yaml.AddFlags(node, flags)
	}

	return node
}

func flowList(root yaml.Node, env dynaml.Binding) yaml.Node {
	rootList := root.Value().([]yaml.Node)

	debug.Debug("HANDLE LIST %v\n", env.Path())
	merged, process, replaced, redirectPath, keyName, flags := processMerges(root, rootList, env)

	if process {
		debug.Debug("process list (key: %s) %v\n", keyName, env.Path())
		newList := []yaml.Node{}
		if len(redirectPath) > 0 {
			env = env.RedirectOverwrite(redirectPath)
		}
		for idx, val := range merged.([]yaml.Node) {
			step, resolved := stepName(idx, val, keyName, env)
			debug.Debug("  step %s\n", step)
			if resolved {
				val = flow(val, env.WithPath(step), false)
			}
			if !val.Undefined() {
				newList = append(newList, val)
			}
		}

		merged = newList
	}

	if keyName != "" {
		root = yaml.KeyNameNode(root, keyName)
	}
	debug.Debug("LIST DONE (%s)%v\n", root.KeyName(), env.Path())

	if replaced {
		root = yaml.ReplaceNode(merged, root, redirectPath)
	} else {
		if len(redirectPath) > 0 {
			root = yaml.RedirectNode(merged, root, redirectPath)
		} else {
			root = yaml.SubstituteNode(merged, root)
		}
	}
	if (flags | root.Flags()) != root.Flags() {
		return yaml.AddFlags(root, flags)
	}
	return root
}

func flowString(root yaml.Node, env dynaml.Binding) yaml.Node {

	sub := yaml.EmbeddedDynaml(root)
	if sub == nil {
		return root
	}
	debug.Debug("dynaml: %v: %s\n", env.Path(), *sub)
	expr, err := dynaml.Parse(*sub, env.Path(), env.StubPath())
	if err != nil {
		return root
	}

	return yaml.SubstituteNode(expr, root)
}

func stepName(index int, value yaml.Node, keyName string, env dynaml.Binding) (string, bool) {
	if keyName == "" {
		keyName = "name"
	}
	name, ok := yaml.FindString(value, keyName)
	if ok {
		return keyName + ":" + name, true
	}

	step := fmt.Sprintf("[%d]", index)
	v, ok := yaml.FindR(true, value, keyName)
	if ok && v.Value() != nil {
		debug.Debug("found raw %s", keyName)
		_, ok := v.Value().(dynaml.Expression)
		if ok {
			v = flow(v, env.WithPath(step), false)
			_, ok := v.Value().(dynaml.Expression)
			if ok {
				return step, false
			}
		}
		name, ok = v.Value().(string)
		if ok {
			return keyName + ":" + name, true
		}
	} else {
		debug.Debug("raw %s not found", keyName)
	}
	return step, true
}

func processMerges(orig yaml.Node, root []yaml.Node, env dynaml.Binding) (interface{}, bool, bool, []string, string, yaml.NodeFlags) {
	var flags yaml.NodeFlags
	flags = get_inherited_flags(env)
	spliced := []yaml.Node{}
	process := true
	template := false
	keyName := orig.KeyName()
	replaced := orig.ReplaceFlag()
	redirectPath := orig.RedirectPath()

	for _, val := range root {
		if val == nil {
			continue
		}

		inlineNode, ok := yaml.UnresolvedListEntryMerge(val)
		if ok {
			debug.Debug("*** %+v\n", inlineNode.Value())
			_, initial := inlineNode.Value().(string)
			result := flow(inlineNode, env, false)
			if result.KeyName() != "" {
				keyName = result.KeyName()
			}
			debug.Debug("=== (%s)%+v\n", keyName, result)
			_, ok := result.Value().(dynaml.Expression)
			if ok {
				if simpleMergeCompatibilityCheck(initial, inlineNode) {
					continue
				}
				m, ok := result.Value().(dynaml.MarkerExpr)
				if ok {
					flags |= m.GetFlags()
					if ok && m.Has(dynaml.TEMPLATE) {
						debug.Debug("found template declaration\n")
						template = true
						process = false
						result = m.TemplateExpression(orig)
						if result == nil {
							continue
						}
						debug.Debug("  insert expression: %v\n", result)
					}
				}
				newMap := make(map[string]yaml.Node)
				newMap["<<"] = result
				val = yaml.SubstituteNode(newMap, orig)
				process = false
			} else {
				inline, ok := result.Value().([]yaml.Node)

				if ok {
					inlineNew := newEntries(inline, root, keyName)
					replaced = result.ReplaceFlag()
					redirectPath = result.RedirectPath()
					if replaced {
						spliced = inlineNew
						process = false
						break
					} else {
						spliced = append(spliced, inlineNew...)
					}
				}
				if ok || result.Value() == nil || yaml.EmbeddedDynaml(result) == nil {
					// still ignore non dynaml value (might be strange but compatible)
					continue
				}
			}
		}

		val, newKey := ProcessKeyTag(val)
		if newKey != "" {
			keyName = newKey
		}
		spliced = append(spliced, val)
	}

	var result interface{}
	if template {
		debug.Debug(" as template\n")
		result = dynaml.TemplateValue{env.Path(), yaml.NewNode(spliced, orig.SourceName()), orig}
	} else {
		result = spliced
	}

	debug.Debug("--> %+v  proc=%v replaced=%v redirect=%v key=%s\n", result, process, replaced, redirectPath, keyName)
	return result, process, replaced, redirectPath, keyName, flags
}

func ProcessKeyTag(val yaml.Node) (yaml.Node, string) {
	keyName := ""

	m, ok := val.Value().(map[string]yaml.Node)
	if ok {
		found := false
		for key, _ := range m {
			split := strings.Index(key, ":")
			if split > 0 {
				if key[:split] == "key" {
					keyName = key[split+1:]
					found = true
				}
			}
		}
		if found {
			newMap := make(map[string]yaml.Node)
			for key, v := range m {
				split := strings.Index(key, ":")
				if split > 0 {
					if key[:split] == "key" {
						key = key[split+1:]
					}
				}
				newMap[key] = v
			}
			return yaml.SubstituteNode(newMap, val), keyName
		}
	}
	return val, keyName
}

func newEntries(a []yaml.Node, b []yaml.Node, keyName string) []yaml.Node {
	if keyName == "" {
		keyName = "name"
	}
	old := yaml.KeyNameNode(yaml.NewNode(b, "some map"), keyName)
	added := []yaml.Node{}

	for _, val := range a {
		name, ok := yaml.FindStringR(true, val, keyName)
		if ok {
			_, found := yaml.FindR(true, old, name) // TODO
			if found {
				continue
			}
		}

		added = append(added, val)
	}

	return added
}

func getSortedKeys(unsortedMap map[string]yaml.Node) []string {
	keys := make([]string, len(unsortedMap))
	i := 0
	for k, _ := range unsortedMap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
