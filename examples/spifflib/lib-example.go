package main

import (
	"fmt"
	"os"

	"github.com/mandelsoft/spiff/flow"
	"github.com/mandelsoft/spiff/spiffing"
)

var state = `
state: {}
`
var stub = `
ages:
  alice: 25
  bob: (( alice + 1 ))
`

var template = `
state:
  <<<: (( &state ))
  random: (( rand("[:alnum:]", 10) )) 
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
	spiff := spiffing.New()
	pstate, err := spiff.Unmarshal("state", []byte(state))
	Error(err)
	pstub, err := spiff.Unmarshal("stub", []byte(stub))
	Error(err)
	ptempl, err := spiff.Unmarshal("template", []byte(template))
	Error(err)
	result, err := spiff.Cascade(ptempl, []spiffing.Node{pstub}, pstate)
	Error(err)
	b, err := spiff.Marshal(result)
	Error(err)
	newstate, err := spiff.Marshal(flow.DetermineState(result))
	Error(err)
	fmt.Printf("==== new state ===\n")
	fmt.Printf("%s\n", string(newstate))
	fmt.Printf("==== result ===\n")
	fmt.Printf("%s\n", string(b))
}
