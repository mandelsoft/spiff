package main

import (
	"fmt"
	"math"
	"os"

	"github.com/mandelsoft/spiff/dynaml"
	"github.com/mandelsoft/spiff/spiffing"
)

func func_pow(arguments []interface{}, binding dynaml.Binding) (interface{}, dynaml.EvaluationInfo, bool) {
	info := dynaml.DefaultInfo()

	if len(arguments) != 2 {
		return info.Error("pow takes 2 arguments")
	}

	a, b, err := dynaml.NumberOperands(arguments[0], arguments[1])

	if err != nil {
		return info.Error("%s", err)
	}
	_, i := a.(int64)
	if i {
		r := math.Pow(float64(a.(int64)), float64(b.(int64)))
		if float64(int64(r)) == r {
			return int64(r), info, true
		}
		return r, info, true
	} else {
		return math.Pow(a.(float64), b.(float64)), info, true
	}
}

var state = `
state: {}
`
var stub = `
unused: (( input ))
ages:
  alice: (( pow(2,5) ))
  bob: (( alice + 1 ))
`

var template = `
state:
  <<<: (( &state ))
  random: (( rand("[:alnum:]", 10) )) 
ages: (( &temporary ))

example:
  name: (( input ))  # direct reference to additional values 
  sum: (( sum[ages|0|s,k,v|->s + v] ))
  int: (( pow(2,4) ))
  float: 2.1
  pow: (( pow(1.1e1,2.1) ))
`

func Error(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	values := map[string]interface{}{}
	values["input"] = "this is an input"

	functions := spiffing.NewFunctions()
	functions.RegisterFunction("pow", func_pow)

	spiff, err := spiffing.New().WithFunctions(functions).WithValues(values)
	Error(err)
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
	newstate, err := spiff.Marshal(spiff.DetermineState(result))
	Error(err)
	fmt.Printf("==== new state ===\n")
	fmt.Printf("%s\n", string(newstate))
	fmt.Printf("==== result ===\n")
	fmt.Printf("%s\n", string(b))
}
