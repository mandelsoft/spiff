package spiffing

import (
	"fmt"
)

func ExampleEvaluateDynamlExpression() {
	ctx, _ := New().WithValues(map[string]interface{}{
		"values": map[string]interface{}{
			"alice": 25,
			"bob":   26,
		},
	})
	result, _ := EvaluateDynamlExpression(ctx, "values.alice + values.bob")
	fmt.Printf("%s", result)
	// Output: 51
}

func ExampleEvaluateDynamlExpression_complex_data() {
	ctx, _ := New().WithValues(map[string]interface{}{
		"values": map[string]interface{}{
			"alice": 25,
			"bob":   26,
		},
	})
	result, _ := EvaluateDynamlExpression(ctx, "[values.alice, values.bob]")
	fmt.Printf("%s", result)
	// Output: - 25
	//- 26
}

func ExampleSpiff_WithInterpolation() {
	ctx := New().WithInterpolation(true)

	template, _ := ctx.Unmarshal("template", []byte(`
host: example.com
port: 8080
url:  http://(( host )):(( port ))
`))

	result, _ := ctx.Cascade(template, nil)
	out, _ := ctx.Marshal(result)
	fmt.Printf("%s", out)
	// Output: host: example.com
	//port: 8080
	//url: http://example.com:8080
}
