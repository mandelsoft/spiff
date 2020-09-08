package main

import (
	"fmt"
	"os"

	"github.com/mandelsoft/spiff/flow"
	"github.com/mandelsoft/spiff/yaml"
)

var stub = `
ages:
  alice: 25
  bob: (( alice + 1 ))
`

var template = `
ages: (( &temporary ))

example:
  sum: (( sum[ages|0|s,k,v|->s + v] ))
`

func Error(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	pstub, err := yaml.Unmarshal("stub", []byte(stub))
	Error(err)
	ptempl, err := yaml.Unmarshal("template", []byte(template))
	Error(err)
	env := flow.NewEnvironment(nil, "code")
	result, err := flow.Cascade(env, ptempl, flow.Options{}, pstub)
	Error(err)
	b, err := yaml.Marshal(result)
	Error(err)
	fmt.Printf("%s\n", string(b))
}
