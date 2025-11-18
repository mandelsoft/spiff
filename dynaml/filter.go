package dynaml

import (
	"github.com/mandelsoft/spiff/yaml"
)

///////////////////////////////////////////////////////////////////////////////
//  filter list context
///////////////////////////////////////////////////////////////////////////////

type filterContext struct {
	defaultContext
}

var FilterListContext = &filterContext{defaultContext{brackets: "[]", keyword: "filter", supported: LIST_SUPPORT}}
var FilterMapContext = &filterContext{defaultContext{brackets: "{}", keyword: "filter", supported: MAP_SUPPORT}}

func (c *filterContext) CreateMappingAggregation(source interface{}) MappingAggregation {
	switch source.(type) {
	case []yaml.Node:
		m := &filterList{}
		m.mapToList = *newMapToList(source, m)
		return m
	case map[string]yaml.Node:
		m := &filterMap{}
		m.mapMapToMap = *newMapMapToMap(source, m)
		return m
	default:
		return nil
	}
}

type filterList struct {
	mapToList
}

func (m *filterList) Add(key interface{}, value interface{}, n yaml.Node, info EvaluationInfo) error {
	if info.Undefined {
		return nil
	}
	if value == nil {
		return nil
	}
	if toBool(value) {
		m.result = append(m.result, n)
	}
	return nil
}

type filterMap struct {
	mapMapToMap
}

func (m *filterMap) Add(key interface{}, value interface{}, n yaml.Node, info EvaluationInfo) error {
	if info.Undefined {
		return nil
	}
	if value == nil {
		return nil
	}
	if toBool(value) {
		m.result[key.(string)] = n
	}
	return nil
}
