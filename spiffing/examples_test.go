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
