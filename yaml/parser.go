package yaml

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/cloudfoundry-incubator/candiedyaml"
	"reflect"
	"time"
)

type NonStringKeyError struct {
	Key interface{}
}

func (e NonStringKeyError) Error() string {
	return fmt.Sprintf("map key must be a string: %#v", e.Key)
}

func Parse(sourceName string, source []byte) (Node, error) {
	docs, err := ParseMulti(sourceName, source)
	if err != nil {
		return nil, err
	}
	if len(docs) > 1 {
		return nil, fmt.Errorf("multi document not possible")
	}
	return docs[0], err
}

func ParseMulti(sourceName string, source []byte) ([]Node, error) {
	docs := []Node{}
	r := bytes.NewBuffer(source)
	d := candiedyaml.NewDecoder(r)

	for d.HasNext() {
		var parsed interface{}
		err := d.Decode(&parsed)
		if err != nil {
			return nil, err
		}
		n, err := sanitize(sourceName, parsed)
		if err != nil {
			return nil, err
		}
		docs = append(docs, n)
	}
	return docs, nil
}

func sanitize(sourceName string, root interface{}) (Node, error) {
	switch rootVal := root.(type) {
	case time.Time:
		return NewNode(rootVal.Format("2019-01-08T10:06:26Z"), sourceName), nil
	case map[interface{}]interface{}:
		sanitized := map[string]Node{}

		for key, val := range rootVal {
			str, ok := key.(string)
			if !ok {
				return nil, NonStringKeyError{key}
			}

			sub, err := sanitize(sourceName, val)
			if err != nil {
				return nil, err
			}

			sanitized[str] = sub
		}

		return NewNode(sanitized, sourceName), nil

	case []interface{}:
		sanitized := []Node{}

		for _, val := range rootVal {
			sub, err := sanitize(sourceName, val)
			if err != nil {
				return nil, err
			}

			sanitized = append(sanitized, sub)
		}

		return NewNode(sanitized, sourceName), nil

	case string, []byte, int64, float64, bool, nil:
		return NewNode(rootVal, sourceName), nil
	}

	return nil, errors.New(fmt.Sprintf("unknown type (%s) during sanitization: %#v\n", reflect.TypeOf(root).String(), root))
}
