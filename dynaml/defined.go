package dynaml

func (e CallExpr) valid(binding Binding) (interface{}, EvaluationInfo, bool) {
	pushed := make([]Expression, len(e.Arguments))
	ok := true
	resolved := true
	valid := true
	var val interface{}

	copy(pushed, e.Arguments)
	for i := range pushed {
		val, _, ok = ResolveExpressionOrPushEvaluation(&pushed[i], &resolved, nil, binding, true)
		if resolved && !ok {
			return false, DefaultInfo(), true
		}
		valid = valid && (val != nil)
	}
	if !resolved {
		return e, DefaultInfo(), true
	}
	return valid, DefaultInfo(), ok
}

func (e CallExpr) defined(binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()
	pushed := make([]Expression, len(e.Arguments))
	ok := true
	resolved := true

	copy(pushed, e.Arguments)
	for i, _ := range pushed {
		_, info, ok = ResolveExpressionOrPushEvaluation(&pushed[i], &resolved, nil, binding, true)
		if resolved {
			if !ok {
				return false, DefaultInfo(), true
			}
			if info.Undefined {
				return false, DefaultInfo(), true
			}
		}
	}
	if !resolved {
		return e, DefaultInfo(), true
	}
	return true, DefaultInfo(), ok
}

func (e CallExpr) optional(binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()
	resolved := true

	if len(e.Arguments) != 2 {
		return info.Error("condition and template required for optional()")
	}

	a, info, ok := ResolveExpressionOrPushEvaluation(&e.Arguments[0], &resolved, &info, binding, false)
	if !resolved {
		return e, info, true
	}
	if !ok || info.Undefined || !toBool(a) {
		return UndefinedExpr{}.Evaluate(binding, false)
	}

	result, infov, ok := e.Arguments[1].Evaluate(binding, false)
	return result, infov.Join(info), ok
}
