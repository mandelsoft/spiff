package dynaml

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleDynaml
	rulePrefer
	ruleMarkedExpression
	ruleSubsequentMarker
	ruleMarker
	ruleExpression
	ruleLevel7
	ruleOr
	ruleLevel6
	ruleConditional
	ruleLevel5
	ruleConcatenation
	ruleLevel4
	ruleLogOr
	ruleLogAnd
	ruleLevel3
	ruleComparison
	ruleCompareOp
	ruleLevel2
	ruleAddition
	ruleSubtraction
	ruleLevel1
	ruleMultiplication
	ruleDivision
	ruleModulo
	ruleLevel0
	ruleChained
	ruleChainedQualifiedExpression
	ruleChainedCall
	ruleArguments
	ruleNextExpression
	ruleSubstitution
	ruleNot
	ruleGrouped
	ruleRange
	ruleInteger
	ruleString
	ruleBoolean
	ruleNil
	ruleUndefined
	ruleList
	ruleContents
	ruleMap
	ruleCreateMap
	ruleAssignments
	ruleAssignment
	ruleMerge
	ruleRefMerge
	ruleSimpleMerge
	ruleReplace
	ruleRequired
	ruleOn
	ruleAuto
	ruleMapping
	ruleSum
	ruleLambda
	ruleLambdaRef
	ruleLambdaExpr
	ruleNextName
	ruleName
	ruleReference
	ruleFollowUpRef
	ruleKey
	ruleIndex
	rulews
	rulereq_ws

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"Dynaml",
	"Prefer",
	"MarkedExpression",
	"SubsequentMarker",
	"Marker",
	"Expression",
	"Level7",
	"Or",
	"Level6",
	"Conditional",
	"Level5",
	"Concatenation",
	"Level4",
	"LogOr",
	"LogAnd",
	"Level3",
	"Comparison",
	"CompareOp",
	"Level2",
	"Addition",
	"Subtraction",
	"Level1",
	"Multiplication",
	"Division",
	"Modulo",
	"Level0",
	"Chained",
	"ChainedQualifiedExpression",
	"ChainedCall",
	"Arguments",
	"NextExpression",
	"Substitution",
	"Not",
	"Grouped",
	"Range",
	"Integer",
	"String",
	"Boolean",
	"Nil",
	"Undefined",
	"List",
	"Contents",
	"Map",
	"CreateMap",
	"Assignments",
	"Assignment",
	"Merge",
	"RefMerge",
	"SimpleMerge",
	"Replace",
	"Required",
	"On",
	"Auto",
	"Mapping",
	"Sum",
	"Lambda",
	"LambdaRef",
	"LambdaExpr",
	"NextName",
	"Name",
	"Reference",
	"FollowUpRef",
	"Key",
	"Index",
	"ws",
	"req_ws",

	"Pre_",
	"_In_",
	"_Suf",
}

type tokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule pegRule, begin, end, next uint32, depth int)
	Expand(index int) tokenTree
	Tokens() <-chan token32
	AST() *node32
	Error() []token32
	trim(length int)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (ast *node32) Print(buffer string) {
	ast.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next uint32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = uint32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, uint32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token32{pegRule: rule, begin: uint32(begin), end: uint32(end), next: uint32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

/*func (t *tokens16) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2 * len(tree))
		for i, v := range tree {
			expanded[i] = v.getToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}*/

func (t *tokens32) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type DynamlGrammar struct {
	Buffer string
	buffer []rune
	rules  [67]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range []rune(buffer) {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p *DynamlGrammar
}

func (e *parseError) Error() string {
	tokens, error := e.p.tokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
	}

	return error
}

func (p *DynamlGrammar) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *DynamlGrammar) Highlighter() {
	p.tokenTree.PrintSyntax()
}

func (p *DynamlGrammar) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != end_symbol {
		p.buffer = append(p.buffer, end_symbol)
	}

	var tree tokenTree = &tokens32{tree: make([]token32, math.MaxInt16)}
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokenTree = tree
		if matches {
			p.tokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin uint32) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
	}

	matchDot := func() bool {
		if buffer[position] != end_symbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Dynaml <- <((Prefer / MarkedExpression / Expression) !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					if !_rules[rulePrefer]() {
						goto l3
					}
					goto l2
				l3:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					if !_rules[ruleMarkedExpression]() {
						goto l4
					}
					goto l2
				l4:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
					if !_rules[ruleExpression]() {
						goto l0
					}
				}
			l2:
				{
					position5, tokenIndex5, depth5 := position, tokenIndex, depth
					if !matchDot() {
						goto l5
					}
					goto l0
				l5:
					position, tokenIndex, depth = position5, tokenIndex5, depth5
				}
				depth--
				add(ruleDynaml, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Prefer <- <(ws ('p' 'r' 'e' 'f' 'e' 'r') req_ws Expression)> */
		func() bool {
			position6, tokenIndex6, depth6 := position, tokenIndex, depth
			{
				position7 := position
				depth++
				if !_rules[rulews]() {
					goto l6
				}
				if buffer[position] != rune('p') {
					goto l6
				}
				position++
				if buffer[position] != rune('r') {
					goto l6
				}
				position++
				if buffer[position] != rune('e') {
					goto l6
				}
				position++
				if buffer[position] != rune('f') {
					goto l6
				}
				position++
				if buffer[position] != rune('e') {
					goto l6
				}
				position++
				if buffer[position] != rune('r') {
					goto l6
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l6
				}
				if !_rules[ruleExpression]() {
					goto l6
				}
				depth--
				add(rulePrefer, position7)
			}
			return true
		l6:
			position, tokenIndex, depth = position6, tokenIndex6, depth6
			return false
		},
		/* 2 MarkedExpression <- <(ws Marker (req_ws SubsequentMarker)* ws Grouped? ws)> */
		func() bool {
			position8, tokenIndex8, depth8 := position, tokenIndex, depth
			{
				position9 := position
				depth++
				if !_rules[rulews]() {
					goto l8
				}
				if !_rules[ruleMarker]() {
					goto l8
				}
			l10:
				{
					position11, tokenIndex11, depth11 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l11
					}
					if !_rules[ruleSubsequentMarker]() {
						goto l11
					}
					goto l10
				l11:
					position, tokenIndex, depth = position11, tokenIndex11, depth11
				}
				if !_rules[rulews]() {
					goto l8
				}
				{
					position12, tokenIndex12, depth12 := position, tokenIndex, depth
					if !_rules[ruleGrouped]() {
						goto l12
					}
					goto l13
				l12:
					position, tokenIndex, depth = position12, tokenIndex12, depth12
				}
			l13:
				if !_rules[rulews]() {
					goto l8
				}
				depth--
				add(ruleMarkedExpression, position9)
			}
			return true
		l8:
			position, tokenIndex, depth = position8, tokenIndex8, depth8
			return false
		},
		/* 3 SubsequentMarker <- <Marker> */
		func() bool {
			position14, tokenIndex14, depth14 := position, tokenIndex, depth
			{
				position15 := position
				depth++
				if !_rules[ruleMarker]() {
					goto l14
				}
				depth--
				add(ruleSubsequentMarker, position15)
			}
			return true
		l14:
			position, tokenIndex, depth = position14, tokenIndex14, depth14
			return false
		},
		/* 4 Marker <- <('&' (('t' 'e' 'm' 'p' 'l' 'a' 't' 'e') / ('t' 'e' 'm' 'p' 'o' 'r' 'a' 'r' 'y')))> */
		func() bool {
			position16, tokenIndex16, depth16 := position, tokenIndex, depth
			{
				position17 := position
				depth++
				if buffer[position] != rune('&') {
					goto l16
				}
				position++
				{
					position18, tokenIndex18, depth18 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l19
					}
					position++
					if buffer[position] != rune('e') {
						goto l19
					}
					position++
					if buffer[position] != rune('m') {
						goto l19
					}
					position++
					if buffer[position] != rune('p') {
						goto l19
					}
					position++
					if buffer[position] != rune('l') {
						goto l19
					}
					position++
					if buffer[position] != rune('a') {
						goto l19
					}
					position++
					if buffer[position] != rune('t') {
						goto l19
					}
					position++
					if buffer[position] != rune('e') {
						goto l19
					}
					position++
					goto l18
				l19:
					position, tokenIndex, depth = position18, tokenIndex18, depth18
					if buffer[position] != rune('t') {
						goto l16
					}
					position++
					if buffer[position] != rune('e') {
						goto l16
					}
					position++
					if buffer[position] != rune('m') {
						goto l16
					}
					position++
					if buffer[position] != rune('p') {
						goto l16
					}
					position++
					if buffer[position] != rune('o') {
						goto l16
					}
					position++
					if buffer[position] != rune('r') {
						goto l16
					}
					position++
					if buffer[position] != rune('a') {
						goto l16
					}
					position++
					if buffer[position] != rune('r') {
						goto l16
					}
					position++
					if buffer[position] != rune('y') {
						goto l16
					}
					position++
				}
			l18:
				depth--
				add(ruleMarker, position17)
			}
			return true
		l16:
			position, tokenIndex, depth = position16, tokenIndex16, depth16
			return false
		},
		/* 5 Expression <- <(ws (LambdaExpr / Level7) ws)> */
		func() bool {
			position20, tokenIndex20, depth20 := position, tokenIndex, depth
			{
				position21 := position
				depth++
				if !_rules[rulews]() {
					goto l20
				}
				{
					position22, tokenIndex22, depth22 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l23
					}
					goto l22
				l23:
					position, tokenIndex, depth = position22, tokenIndex22, depth22
					if !_rules[ruleLevel7]() {
						goto l20
					}
				}
			l22:
				if !_rules[rulews]() {
					goto l20
				}
				depth--
				add(ruleExpression, position21)
			}
			return true
		l20:
			position, tokenIndex, depth = position20, tokenIndex20, depth20
			return false
		},
		/* 6 Level7 <- <(Level6 (req_ws Or)*)> */
		func() bool {
			position24, tokenIndex24, depth24 := position, tokenIndex, depth
			{
				position25 := position
				depth++
				if !_rules[ruleLevel6]() {
					goto l24
				}
			l26:
				{
					position27, tokenIndex27, depth27 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l27
					}
					if !_rules[ruleOr]() {
						goto l27
					}
					goto l26
				l27:
					position, tokenIndex, depth = position27, tokenIndex27, depth27
				}
				depth--
				add(ruleLevel7, position25)
			}
			return true
		l24:
			position, tokenIndex, depth = position24, tokenIndex24, depth24
			return false
		},
		/* 7 Or <- <('|' '|' req_ws Level6)> */
		func() bool {
			position28, tokenIndex28, depth28 := position, tokenIndex, depth
			{
				position29 := position
				depth++
				if buffer[position] != rune('|') {
					goto l28
				}
				position++
				if buffer[position] != rune('|') {
					goto l28
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l28
				}
				if !_rules[ruleLevel6]() {
					goto l28
				}
				depth--
				add(ruleOr, position29)
			}
			return true
		l28:
			position, tokenIndex, depth = position28, tokenIndex28, depth28
			return false
		},
		/* 8 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position30, tokenIndex30, depth30 := position, tokenIndex, depth
			{
				position31 := position
				depth++
				{
					position32, tokenIndex32, depth32 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l33
					}
					goto l32
				l33:
					position, tokenIndex, depth = position32, tokenIndex32, depth32
					if !_rules[ruleLevel5]() {
						goto l30
					}
				}
			l32:
				depth--
				add(ruleLevel6, position31)
			}
			return true
		l30:
			position, tokenIndex, depth = position30, tokenIndex30, depth30
			return false
		},
		/* 9 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position34, tokenIndex34, depth34 := position, tokenIndex, depth
			{
				position35 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l34
				}
				if !_rules[rulews]() {
					goto l34
				}
				if buffer[position] != rune('?') {
					goto l34
				}
				position++
				if !_rules[ruleExpression]() {
					goto l34
				}
				if buffer[position] != rune(':') {
					goto l34
				}
				position++
				if !_rules[ruleExpression]() {
					goto l34
				}
				depth--
				add(ruleConditional, position35)
			}
			return true
		l34:
			position, tokenIndex, depth = position34, tokenIndex34, depth34
			return false
		},
		/* 10 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position36, tokenIndex36, depth36 := position, tokenIndex, depth
			{
				position37 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l36
				}
			l38:
				{
					position39, tokenIndex39, depth39 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l39
					}
					goto l38
				l39:
					position, tokenIndex, depth = position39, tokenIndex39, depth39
				}
				depth--
				add(ruleLevel5, position37)
			}
			return true
		l36:
			position, tokenIndex, depth = position36, tokenIndex36, depth36
			return false
		},
		/* 11 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l40
				}
				if !_rules[ruleLevel4]() {
					goto l40
				}
				depth--
				add(ruleConcatenation, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 12 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position42, tokenIndex42, depth42 := position, tokenIndex, depth
			{
				position43 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l42
				}
			l44:
				{
					position45, tokenIndex45, depth45 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l45
					}
					{
						position46, tokenIndex46, depth46 := position, tokenIndex, depth
						if !_rules[ruleLogOr]() {
							goto l47
						}
						goto l46
					l47:
						position, tokenIndex, depth = position46, tokenIndex46, depth46
						if !_rules[ruleLogAnd]() {
							goto l45
						}
					}
				l46:
					goto l44
				l45:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
				}
				depth--
				add(ruleLevel4, position43)
			}
			return true
		l42:
			position, tokenIndex, depth = position42, tokenIndex42, depth42
			return false
		},
		/* 13 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position48, tokenIndex48, depth48 := position, tokenIndex, depth
			{
				position49 := position
				depth++
				if buffer[position] != rune('-') {
					goto l48
				}
				position++
				if buffer[position] != rune('o') {
					goto l48
				}
				position++
				if buffer[position] != rune('r') {
					goto l48
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l48
				}
				if !_rules[ruleLevel3]() {
					goto l48
				}
				depth--
				add(ruleLogOr, position49)
			}
			return true
		l48:
			position, tokenIndex, depth = position48, tokenIndex48, depth48
			return false
		},
		/* 14 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position50, tokenIndex50, depth50 := position, tokenIndex, depth
			{
				position51 := position
				depth++
				if buffer[position] != rune('-') {
					goto l50
				}
				position++
				if buffer[position] != rune('a') {
					goto l50
				}
				position++
				if buffer[position] != rune('n') {
					goto l50
				}
				position++
				if buffer[position] != rune('d') {
					goto l50
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l50
				}
				if !_rules[ruleLevel3]() {
					goto l50
				}
				depth--
				add(ruleLogAnd, position51)
			}
			return true
		l50:
			position, tokenIndex, depth = position50, tokenIndex50, depth50
			return false
		},
		/* 15 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position52, tokenIndex52, depth52 := position, tokenIndex, depth
			{
				position53 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l52
				}
			l54:
				{
					position55, tokenIndex55, depth55 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l55
					}
					if !_rules[ruleComparison]() {
						goto l55
					}
					goto l54
				l55:
					position, tokenIndex, depth = position55, tokenIndex55, depth55
				}
				depth--
				add(ruleLevel3, position53)
			}
			return true
		l52:
			position, tokenIndex, depth = position52, tokenIndex52, depth52
			return false
		},
		/* 16 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position56, tokenIndex56, depth56 := position, tokenIndex, depth
			{
				position57 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l56
				}
				if !_rules[rulereq_ws]() {
					goto l56
				}
				if !_rules[ruleLevel2]() {
					goto l56
				}
				depth--
				add(ruleComparison, position57)
			}
			return true
		l56:
			position, tokenIndex, depth = position56, tokenIndex56, depth56
			return false
		},
		/* 17 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position58, tokenIndex58, depth58 := position, tokenIndex, depth
			{
				position59 := position
				depth++
				{
					position60, tokenIndex60, depth60 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l61
					}
					position++
					if buffer[position] != rune('=') {
						goto l61
					}
					position++
					goto l60
				l61:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
					if buffer[position] != rune('!') {
						goto l62
					}
					position++
					if buffer[position] != rune('=') {
						goto l62
					}
					position++
					goto l60
				l62:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
					if buffer[position] != rune('<') {
						goto l63
					}
					position++
					if buffer[position] != rune('=') {
						goto l63
					}
					position++
					goto l60
				l63:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
					if buffer[position] != rune('>') {
						goto l64
					}
					position++
					if buffer[position] != rune('=') {
						goto l64
					}
					position++
					goto l60
				l64:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
					if buffer[position] != rune('>') {
						goto l65
					}
					position++
					goto l60
				l65:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
					if buffer[position] != rune('<') {
						goto l66
					}
					position++
					goto l60
				l66:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
					if buffer[position] != rune('>') {
						goto l58
					}
					position++
				}
			l60:
				depth--
				add(ruleCompareOp, position59)
			}
			return true
		l58:
			position, tokenIndex, depth = position58, tokenIndex58, depth58
			return false
		},
		/* 18 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position67, tokenIndex67, depth67 := position, tokenIndex, depth
			{
				position68 := position
				depth++
				if !_rules[ruleLevel1]() {
					goto l67
				}
			l69:
				{
					position70, tokenIndex70, depth70 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l70
					}
					{
						position71, tokenIndex71, depth71 := position, tokenIndex, depth
						if !_rules[ruleAddition]() {
							goto l72
						}
						goto l71
					l72:
						position, tokenIndex, depth = position71, tokenIndex71, depth71
						if !_rules[ruleSubtraction]() {
							goto l70
						}
					}
				l71:
					goto l69
				l70:
					position, tokenIndex, depth = position70, tokenIndex70, depth70
				}
				depth--
				add(ruleLevel2, position68)
			}
			return true
		l67:
			position, tokenIndex, depth = position67, tokenIndex67, depth67
			return false
		},
		/* 19 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position73, tokenIndex73, depth73 := position, tokenIndex, depth
			{
				position74 := position
				depth++
				if buffer[position] != rune('+') {
					goto l73
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l73
				}
				if !_rules[ruleLevel1]() {
					goto l73
				}
				depth--
				add(ruleAddition, position74)
			}
			return true
		l73:
			position, tokenIndex, depth = position73, tokenIndex73, depth73
			return false
		},
		/* 20 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position75, tokenIndex75, depth75 := position, tokenIndex, depth
			{
				position76 := position
				depth++
				if buffer[position] != rune('-') {
					goto l75
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l75
				}
				if !_rules[ruleLevel1]() {
					goto l75
				}
				depth--
				add(ruleSubtraction, position76)
			}
			return true
		l75:
			position, tokenIndex, depth = position75, tokenIndex75, depth75
			return false
		},
		/* 21 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position77, tokenIndex77, depth77 := position, tokenIndex, depth
			{
				position78 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l77
				}
			l79:
				{
					position80, tokenIndex80, depth80 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l80
					}
					{
						position81, tokenIndex81, depth81 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l82
						}
						goto l81
					l82:
						position, tokenIndex, depth = position81, tokenIndex81, depth81
						if !_rules[ruleDivision]() {
							goto l83
						}
						goto l81
					l83:
						position, tokenIndex, depth = position81, tokenIndex81, depth81
						if !_rules[ruleModulo]() {
							goto l80
						}
					}
				l81:
					goto l79
				l80:
					position, tokenIndex, depth = position80, tokenIndex80, depth80
				}
				depth--
				add(ruleLevel1, position78)
			}
			return true
		l77:
			position, tokenIndex, depth = position77, tokenIndex77, depth77
			return false
		},
		/* 22 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position84, tokenIndex84, depth84 := position, tokenIndex, depth
			{
				position85 := position
				depth++
				if buffer[position] != rune('*') {
					goto l84
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l84
				}
				if !_rules[ruleLevel0]() {
					goto l84
				}
				depth--
				add(ruleMultiplication, position85)
			}
			return true
		l84:
			position, tokenIndex, depth = position84, tokenIndex84, depth84
			return false
		},
		/* 23 Division <- <('/' req_ws Level0)> */
		func() bool {
			position86, tokenIndex86, depth86 := position, tokenIndex, depth
			{
				position87 := position
				depth++
				if buffer[position] != rune('/') {
					goto l86
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l86
				}
				if !_rules[ruleLevel0]() {
					goto l86
				}
				depth--
				add(ruleDivision, position87)
			}
			return true
		l86:
			position, tokenIndex, depth = position86, tokenIndex86, depth86
			return false
		},
		/* 24 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position88, tokenIndex88, depth88 := position, tokenIndex, depth
			{
				position89 := position
				depth++
				if buffer[position] != rune('%') {
					goto l88
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l88
				}
				if !_rules[ruleLevel0]() {
					goto l88
				}
				depth--
				add(ruleModulo, position89)
			}
			return true
		l88:
			position, tokenIndex, depth = position88, tokenIndex88, depth88
			return false
		},
		/* 25 Level0 <- <(String / Integer / Boolean / Undefined / Nil / Not / Substitution / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position90, tokenIndex90, depth90 := position, tokenIndex, depth
			{
				position91 := position
				depth++
				{
					position92, tokenIndex92, depth92 := position, tokenIndex, depth
					if !_rules[ruleString]() {
						goto l93
					}
					goto l92
				l93:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleInteger]() {
						goto l94
					}
					goto l92
				l94:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleBoolean]() {
						goto l95
					}
					goto l92
				l95:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleUndefined]() {
						goto l96
					}
					goto l92
				l96:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleNil]() {
						goto l97
					}
					goto l92
				l97:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleNot]() {
						goto l98
					}
					goto l92
				l98:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleSubstitution]() {
						goto l99
					}
					goto l92
				l99:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleMerge]() {
						goto l100
					}
					goto l92
				l100:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleAuto]() {
						goto l101
					}
					goto l92
				l101:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleLambda]() {
						goto l102
					}
					goto l92
				l102:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleChained]() {
						goto l90
					}
				}
			l92:
				depth--
				add(ruleLevel0, position91)
			}
			return true
		l90:
			position, tokenIndex, depth = position90, tokenIndex90, depth90
			return false
		},
		/* 26 Chained <- <((Mapping / Sum / List / Map / Range / ((Grouped / Reference) ChainedCall*)) (ChainedQualifiedExpression ChainedCall+)* ChainedQualifiedExpression?)> */
		func() bool {
			position103, tokenIndex103, depth103 := position, tokenIndex, depth
			{
				position104 := position
				depth++
				{
					position105, tokenIndex105, depth105 := position, tokenIndex, depth
					if !_rules[ruleMapping]() {
						goto l106
					}
					goto l105
				l106:
					position, tokenIndex, depth = position105, tokenIndex105, depth105
					if !_rules[ruleSum]() {
						goto l107
					}
					goto l105
				l107:
					position, tokenIndex, depth = position105, tokenIndex105, depth105
					if !_rules[ruleList]() {
						goto l108
					}
					goto l105
				l108:
					position, tokenIndex, depth = position105, tokenIndex105, depth105
					if !_rules[ruleMap]() {
						goto l109
					}
					goto l105
				l109:
					position, tokenIndex, depth = position105, tokenIndex105, depth105
					if !_rules[ruleRange]() {
						goto l110
					}
					goto l105
				l110:
					position, tokenIndex, depth = position105, tokenIndex105, depth105
					{
						position111, tokenIndex111, depth111 := position, tokenIndex, depth
						if !_rules[ruleGrouped]() {
							goto l112
						}
						goto l111
					l112:
						position, tokenIndex, depth = position111, tokenIndex111, depth111
						if !_rules[ruleReference]() {
							goto l103
						}
					}
				l111:
				l113:
					{
						position114, tokenIndex114, depth114 := position, tokenIndex, depth
						if !_rules[ruleChainedCall]() {
							goto l114
						}
						goto l113
					l114:
						position, tokenIndex, depth = position114, tokenIndex114, depth114
					}
				}
			l105:
			l115:
				{
					position116, tokenIndex116, depth116 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l116
					}
					if !_rules[ruleChainedCall]() {
						goto l116
					}
				l117:
					{
						position118, tokenIndex118, depth118 := position, tokenIndex, depth
						if !_rules[ruleChainedCall]() {
							goto l118
						}
						goto l117
					l118:
						position, tokenIndex, depth = position118, tokenIndex118, depth118
					}
					goto l115
				l116:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
				}
				{
					position119, tokenIndex119, depth119 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l119
					}
					goto l120
				l119:
					position, tokenIndex, depth = position119, tokenIndex119, depth119
				}
			l120:
				depth--
				add(ruleChained, position104)
			}
			return true
		l103:
			position, tokenIndex, depth = position103, tokenIndex103, depth103
			return false
		},
		/* 27 ChainedQualifiedExpression <- <('.' FollowUpRef)> */
		func() bool {
			position121, tokenIndex121, depth121 := position, tokenIndex, depth
			{
				position122 := position
				depth++
				if buffer[position] != rune('.') {
					goto l121
				}
				position++
				if !_rules[ruleFollowUpRef]() {
					goto l121
				}
				depth--
				add(ruleChainedQualifiedExpression, position122)
			}
			return true
		l121:
			position, tokenIndex, depth = position121, tokenIndex121, depth121
			return false
		},
		/* 28 ChainedCall <- <('(' Arguments ')')> */
		func() bool {
			position123, tokenIndex123, depth123 := position, tokenIndex, depth
			{
				position124 := position
				depth++
				if buffer[position] != rune('(') {
					goto l123
				}
				position++
				if !_rules[ruleArguments]() {
					goto l123
				}
				if buffer[position] != rune(')') {
					goto l123
				}
				position++
				depth--
				add(ruleChainedCall, position124)
			}
			return true
		l123:
			position, tokenIndex, depth = position123, tokenIndex123, depth123
			return false
		},
		/* 29 Arguments <- <(Expression NextExpression*)> */
		func() bool {
			position125, tokenIndex125, depth125 := position, tokenIndex, depth
			{
				position126 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l125
				}
			l127:
				{
					position128, tokenIndex128, depth128 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l128
					}
					goto l127
				l128:
					position, tokenIndex, depth = position128, tokenIndex128, depth128
				}
				depth--
				add(ruleArguments, position126)
			}
			return true
		l125:
			position, tokenIndex, depth = position125, tokenIndex125, depth125
			return false
		},
		/* 30 NextExpression <- <(',' Expression)> */
		func() bool {
			position129, tokenIndex129, depth129 := position, tokenIndex, depth
			{
				position130 := position
				depth++
				if buffer[position] != rune(',') {
					goto l129
				}
				position++
				if !_rules[ruleExpression]() {
					goto l129
				}
				depth--
				add(ruleNextExpression, position130)
			}
			return true
		l129:
			position, tokenIndex, depth = position129, tokenIndex129, depth129
			return false
		},
		/* 31 Substitution <- <('*' Level0)> */
		func() bool {
			position131, tokenIndex131, depth131 := position, tokenIndex, depth
			{
				position132 := position
				depth++
				if buffer[position] != rune('*') {
					goto l131
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l131
				}
				depth--
				add(ruleSubstitution, position132)
			}
			return true
		l131:
			position, tokenIndex, depth = position131, tokenIndex131, depth131
			return false
		},
		/* 32 Not <- <('!' ws Level0)> */
		func() bool {
			position133, tokenIndex133, depth133 := position, tokenIndex, depth
			{
				position134 := position
				depth++
				if buffer[position] != rune('!') {
					goto l133
				}
				position++
				if !_rules[rulews]() {
					goto l133
				}
				if !_rules[ruleLevel0]() {
					goto l133
				}
				depth--
				add(ruleNot, position134)
			}
			return true
		l133:
			position, tokenIndex, depth = position133, tokenIndex133, depth133
			return false
		},
		/* 33 Grouped <- <('(' Expression ')')> */
		func() bool {
			position135, tokenIndex135, depth135 := position, tokenIndex, depth
			{
				position136 := position
				depth++
				if buffer[position] != rune('(') {
					goto l135
				}
				position++
				if !_rules[ruleExpression]() {
					goto l135
				}
				if buffer[position] != rune(')') {
					goto l135
				}
				position++
				depth--
				add(ruleGrouped, position136)
			}
			return true
		l135:
			position, tokenIndex, depth = position135, tokenIndex135, depth135
			return false
		},
		/* 34 Range <- <('[' Expression ('.' '.') Expression ']')> */
		func() bool {
			position137, tokenIndex137, depth137 := position, tokenIndex, depth
			{
				position138 := position
				depth++
				if buffer[position] != rune('[') {
					goto l137
				}
				position++
				if !_rules[ruleExpression]() {
					goto l137
				}
				if buffer[position] != rune('.') {
					goto l137
				}
				position++
				if buffer[position] != rune('.') {
					goto l137
				}
				position++
				if !_rules[ruleExpression]() {
					goto l137
				}
				if buffer[position] != rune(']') {
					goto l137
				}
				position++
				depth--
				add(ruleRange, position138)
			}
			return true
		l137:
			position, tokenIndex, depth = position137, tokenIndex137, depth137
			return false
		},
		/* 35 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position139, tokenIndex139, depth139 := position, tokenIndex, depth
			{
				position140 := position
				depth++
				{
					position141, tokenIndex141, depth141 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l141
					}
					position++
					goto l142
				l141:
					position, tokenIndex, depth = position141, tokenIndex141, depth141
				}
			l142:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l139
				}
				position++
			l143:
				{
					position144, tokenIndex144, depth144 := position, tokenIndex, depth
					{
						position145, tokenIndex145, depth145 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l146
						}
						position++
						goto l145
					l146:
						position, tokenIndex, depth = position145, tokenIndex145, depth145
						if buffer[position] != rune('_') {
							goto l144
						}
						position++
					}
				l145:
					goto l143
				l144:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
				}
				depth--
				add(ruleInteger, position140)
			}
			return true
		l139:
			position, tokenIndex, depth = position139, tokenIndex139, depth139
			return false
		},
		/* 36 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				if buffer[position] != rune('"') {
					goto l147
				}
				position++
			l149:
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					{
						position151, tokenIndex151, depth151 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l152
						}
						position++
						if buffer[position] != rune('"') {
							goto l152
						}
						position++
						goto l151
					l152:
						position, tokenIndex, depth = position151, tokenIndex151, depth151
						{
							position153, tokenIndex153, depth153 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l153
							}
							position++
							goto l150
						l153:
							position, tokenIndex, depth = position153, tokenIndex153, depth153
						}
						if !matchDot() {
							goto l150
						}
					}
				l151:
					goto l149
				l150:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
				}
				if buffer[position] != rune('"') {
					goto l147
				}
				position++
				depth--
				add(ruleString, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 37 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position154, tokenIndex154, depth154 := position, tokenIndex, depth
			{
				position155 := position
				depth++
				{
					position156, tokenIndex156, depth156 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l157
					}
					position++
					if buffer[position] != rune('r') {
						goto l157
					}
					position++
					if buffer[position] != rune('u') {
						goto l157
					}
					position++
					if buffer[position] != rune('e') {
						goto l157
					}
					position++
					goto l156
				l157:
					position, tokenIndex, depth = position156, tokenIndex156, depth156
					if buffer[position] != rune('f') {
						goto l154
					}
					position++
					if buffer[position] != rune('a') {
						goto l154
					}
					position++
					if buffer[position] != rune('l') {
						goto l154
					}
					position++
					if buffer[position] != rune('s') {
						goto l154
					}
					position++
					if buffer[position] != rune('e') {
						goto l154
					}
					position++
				}
			l156:
				depth--
				add(ruleBoolean, position155)
			}
			return true
		l154:
			position, tokenIndex, depth = position154, tokenIndex154, depth154
			return false
		},
		/* 38 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position158, tokenIndex158, depth158 := position, tokenIndex, depth
			{
				position159 := position
				depth++
				{
					position160, tokenIndex160, depth160 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l161
					}
					position++
					if buffer[position] != rune('i') {
						goto l161
					}
					position++
					if buffer[position] != rune('l') {
						goto l161
					}
					position++
					goto l160
				l161:
					position, tokenIndex, depth = position160, tokenIndex160, depth160
					if buffer[position] != rune('~') {
						goto l158
					}
					position++
				}
			l160:
				depth--
				add(ruleNil, position159)
			}
			return true
		l158:
			position, tokenIndex, depth = position158, tokenIndex158, depth158
			return false
		},
		/* 39 Undefined <- <('~' '~')> */
		func() bool {
			position162, tokenIndex162, depth162 := position, tokenIndex, depth
			{
				position163 := position
				depth++
				if buffer[position] != rune('~') {
					goto l162
				}
				position++
				if buffer[position] != rune('~') {
					goto l162
				}
				position++
				depth--
				add(ruleUndefined, position163)
			}
			return true
		l162:
			position, tokenIndex, depth = position162, tokenIndex162, depth162
			return false
		},
		/* 40 List <- <('[' Contents? ']')> */
		func() bool {
			position164, tokenIndex164, depth164 := position, tokenIndex, depth
			{
				position165 := position
				depth++
				if buffer[position] != rune('[') {
					goto l164
				}
				position++
				{
					position166, tokenIndex166, depth166 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l166
					}
					goto l167
				l166:
					position, tokenIndex, depth = position166, tokenIndex166, depth166
				}
			l167:
				if buffer[position] != rune(']') {
					goto l164
				}
				position++
				depth--
				add(ruleList, position165)
			}
			return true
		l164:
			position, tokenIndex, depth = position164, tokenIndex164, depth164
			return false
		},
		/* 41 Contents <- <(Expression NextExpression*)> */
		func() bool {
			position168, tokenIndex168, depth168 := position, tokenIndex, depth
			{
				position169 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l168
				}
			l170:
				{
					position171, tokenIndex171, depth171 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l171
					}
					goto l170
				l171:
					position, tokenIndex, depth = position171, tokenIndex171, depth171
				}
				depth--
				add(ruleContents, position169)
			}
			return true
		l168:
			position, tokenIndex, depth = position168, tokenIndex168, depth168
			return false
		},
		/* 42 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position172, tokenIndex172, depth172 := position, tokenIndex, depth
			{
				position173 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l172
				}
				if !_rules[rulews]() {
					goto l172
				}
				{
					position174, tokenIndex174, depth174 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l174
					}
					goto l175
				l174:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
				}
			l175:
				if buffer[position] != rune('}') {
					goto l172
				}
				position++
				depth--
				add(ruleMap, position173)
			}
			return true
		l172:
			position, tokenIndex, depth = position172, tokenIndex172, depth172
			return false
		},
		/* 43 CreateMap <- <'{'> */
		func() bool {
			position176, tokenIndex176, depth176 := position, tokenIndex, depth
			{
				position177 := position
				depth++
				if buffer[position] != rune('{') {
					goto l176
				}
				position++
				depth--
				add(ruleCreateMap, position177)
			}
			return true
		l176:
			position, tokenIndex, depth = position176, tokenIndex176, depth176
			return false
		},
		/* 44 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position178, tokenIndex178, depth178 := position, tokenIndex, depth
			{
				position179 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l178
				}
			l180:
				{
					position181, tokenIndex181, depth181 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l181
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l181
					}
					goto l180
				l181:
					position, tokenIndex, depth = position181, tokenIndex181, depth181
				}
				depth--
				add(ruleAssignments, position179)
			}
			return true
		l178:
			position, tokenIndex, depth = position178, tokenIndex178, depth178
			return false
		},
		/* 45 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position182, tokenIndex182, depth182 := position, tokenIndex, depth
			{
				position183 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l182
				}
				if buffer[position] != rune('=') {
					goto l182
				}
				position++
				if !_rules[ruleExpression]() {
					goto l182
				}
				depth--
				add(ruleAssignment, position183)
			}
			return true
		l182:
			position, tokenIndex, depth = position182, tokenIndex182, depth182
			return false
		},
		/* 46 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position184, tokenIndex184, depth184 := position, tokenIndex, depth
			{
				position185 := position
				depth++
				{
					position186, tokenIndex186, depth186 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l187
					}
					goto l186
				l187:
					position, tokenIndex, depth = position186, tokenIndex186, depth186
					if !_rules[ruleSimpleMerge]() {
						goto l184
					}
				}
			l186:
				depth--
				add(ruleMerge, position185)
			}
			return true
		l184:
			position, tokenIndex, depth = position184, tokenIndex184, depth184
			return false
		},
		/* 47 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position188, tokenIndex188, depth188 := position, tokenIndex, depth
			{
				position189 := position
				depth++
				if buffer[position] != rune('m') {
					goto l188
				}
				position++
				if buffer[position] != rune('e') {
					goto l188
				}
				position++
				if buffer[position] != rune('r') {
					goto l188
				}
				position++
				if buffer[position] != rune('g') {
					goto l188
				}
				position++
				if buffer[position] != rune('e') {
					goto l188
				}
				position++
				{
					position190, tokenIndex190, depth190 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l190
					}
					if !_rules[ruleRequired]() {
						goto l190
					}
					goto l188
				l190:
					position, tokenIndex, depth = position190, tokenIndex190, depth190
				}
				{
					position191, tokenIndex191, depth191 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l191
					}
					{
						position193, tokenIndex193, depth193 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l194
						}
						goto l193
					l194:
						position, tokenIndex, depth = position193, tokenIndex193, depth193
						if !_rules[ruleOn]() {
							goto l191
						}
					}
				l193:
					goto l192
				l191:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
				}
			l192:
				if !_rules[rulereq_ws]() {
					goto l188
				}
				if !_rules[ruleReference]() {
					goto l188
				}
				depth--
				add(ruleRefMerge, position189)
			}
			return true
		l188:
			position, tokenIndex, depth = position188, tokenIndex188, depth188
			return false
		},
		/* 48 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position195, tokenIndex195, depth195 := position, tokenIndex, depth
			{
				position196 := position
				depth++
				if buffer[position] != rune('m') {
					goto l195
				}
				position++
				if buffer[position] != rune('e') {
					goto l195
				}
				position++
				if buffer[position] != rune('r') {
					goto l195
				}
				position++
				if buffer[position] != rune('g') {
					goto l195
				}
				position++
				if buffer[position] != rune('e') {
					goto l195
				}
				position++
				{
					position197, tokenIndex197, depth197 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l197
					}
					{
						position199, tokenIndex199, depth199 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l200
						}
						goto l199
					l200:
						position, tokenIndex, depth = position199, tokenIndex199, depth199
						if !_rules[ruleRequired]() {
							goto l201
						}
						goto l199
					l201:
						position, tokenIndex, depth = position199, tokenIndex199, depth199
						if !_rules[ruleOn]() {
							goto l197
						}
					}
				l199:
					goto l198
				l197:
					position, tokenIndex, depth = position197, tokenIndex197, depth197
				}
			l198:
				depth--
				add(ruleSimpleMerge, position196)
			}
			return true
		l195:
			position, tokenIndex, depth = position195, tokenIndex195, depth195
			return false
		},
		/* 49 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position202, tokenIndex202, depth202 := position, tokenIndex, depth
			{
				position203 := position
				depth++
				if buffer[position] != rune('r') {
					goto l202
				}
				position++
				if buffer[position] != rune('e') {
					goto l202
				}
				position++
				if buffer[position] != rune('p') {
					goto l202
				}
				position++
				if buffer[position] != rune('l') {
					goto l202
				}
				position++
				if buffer[position] != rune('a') {
					goto l202
				}
				position++
				if buffer[position] != rune('c') {
					goto l202
				}
				position++
				if buffer[position] != rune('e') {
					goto l202
				}
				position++
				depth--
				add(ruleReplace, position203)
			}
			return true
		l202:
			position, tokenIndex, depth = position202, tokenIndex202, depth202
			return false
		},
		/* 50 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position204, tokenIndex204, depth204 := position, tokenIndex, depth
			{
				position205 := position
				depth++
				if buffer[position] != rune('r') {
					goto l204
				}
				position++
				if buffer[position] != rune('e') {
					goto l204
				}
				position++
				if buffer[position] != rune('q') {
					goto l204
				}
				position++
				if buffer[position] != rune('u') {
					goto l204
				}
				position++
				if buffer[position] != rune('i') {
					goto l204
				}
				position++
				if buffer[position] != rune('r') {
					goto l204
				}
				position++
				if buffer[position] != rune('e') {
					goto l204
				}
				position++
				if buffer[position] != rune('d') {
					goto l204
				}
				position++
				depth--
				add(ruleRequired, position205)
			}
			return true
		l204:
			position, tokenIndex, depth = position204, tokenIndex204, depth204
			return false
		},
		/* 51 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position206, tokenIndex206, depth206 := position, tokenIndex, depth
			{
				position207 := position
				depth++
				if buffer[position] != rune('o') {
					goto l206
				}
				position++
				if buffer[position] != rune('n') {
					goto l206
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l206
				}
				if !_rules[ruleName]() {
					goto l206
				}
				depth--
				add(ruleOn, position207)
			}
			return true
		l206:
			position, tokenIndex, depth = position206, tokenIndex206, depth206
			return false
		},
		/* 52 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position208, tokenIndex208, depth208 := position, tokenIndex, depth
			{
				position209 := position
				depth++
				if buffer[position] != rune('a') {
					goto l208
				}
				position++
				if buffer[position] != rune('u') {
					goto l208
				}
				position++
				if buffer[position] != rune('t') {
					goto l208
				}
				position++
				if buffer[position] != rune('o') {
					goto l208
				}
				position++
				depth--
				add(ruleAuto, position209)
			}
			return true
		l208:
			position, tokenIndex, depth = position208, tokenIndex208, depth208
			return false
		},
		/* 53 Mapping <- <('m' 'a' 'p' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position210, tokenIndex210, depth210 := position, tokenIndex, depth
			{
				position211 := position
				depth++
				if buffer[position] != rune('m') {
					goto l210
				}
				position++
				if buffer[position] != rune('a') {
					goto l210
				}
				position++
				if buffer[position] != rune('p') {
					goto l210
				}
				position++
				if buffer[position] != rune('[') {
					goto l210
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l210
				}
				{
					position212, tokenIndex212, depth212 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l213
					}
					goto l212
				l213:
					position, tokenIndex, depth = position212, tokenIndex212, depth212
					if buffer[position] != rune('|') {
						goto l210
					}
					position++
					if !_rules[ruleExpression]() {
						goto l210
					}
				}
			l212:
				if buffer[position] != rune(']') {
					goto l210
				}
				position++
				depth--
				add(ruleMapping, position211)
			}
			return true
		l210:
			position, tokenIndex, depth = position210, tokenIndex210, depth210
			return false
		},
		/* 54 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position214, tokenIndex214, depth214 := position, tokenIndex, depth
			{
				position215 := position
				depth++
				if buffer[position] != rune('s') {
					goto l214
				}
				position++
				if buffer[position] != rune('u') {
					goto l214
				}
				position++
				if buffer[position] != rune('m') {
					goto l214
				}
				position++
				if buffer[position] != rune('[') {
					goto l214
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l214
				}
				if buffer[position] != rune('|') {
					goto l214
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l214
				}
				{
					position216, tokenIndex216, depth216 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l217
					}
					goto l216
				l217:
					position, tokenIndex, depth = position216, tokenIndex216, depth216
					if buffer[position] != rune('|') {
						goto l214
					}
					position++
					if !_rules[ruleExpression]() {
						goto l214
					}
				}
			l216:
				if buffer[position] != rune(']') {
					goto l214
				}
				position++
				depth--
				add(ruleSum, position215)
			}
			return true
		l214:
			position, tokenIndex, depth = position214, tokenIndex214, depth214
			return false
		},
		/* 55 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position218, tokenIndex218, depth218 := position, tokenIndex, depth
			{
				position219 := position
				depth++
				if buffer[position] != rune('l') {
					goto l218
				}
				position++
				if buffer[position] != rune('a') {
					goto l218
				}
				position++
				if buffer[position] != rune('m') {
					goto l218
				}
				position++
				if buffer[position] != rune('b') {
					goto l218
				}
				position++
				if buffer[position] != rune('d') {
					goto l218
				}
				position++
				if buffer[position] != rune('a') {
					goto l218
				}
				position++
				{
					position220, tokenIndex220, depth220 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l221
					}
					goto l220
				l221:
					position, tokenIndex, depth = position220, tokenIndex220, depth220
					if !_rules[ruleLambdaExpr]() {
						goto l218
					}
				}
			l220:
				depth--
				add(ruleLambda, position219)
			}
			return true
		l218:
			position, tokenIndex, depth = position218, tokenIndex218, depth218
			return false
		},
		/* 56 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position222, tokenIndex222, depth222 := position, tokenIndex, depth
			{
				position223 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l222
				}
				if !_rules[ruleExpression]() {
					goto l222
				}
				depth--
				add(ruleLambdaRef, position223)
			}
			return true
		l222:
			position, tokenIndex, depth = position222, tokenIndex222, depth222
			return false
		},
		/* 57 LambdaExpr <- <(ws '|' ws Name NextName* ws '|' ws ('-' '>') Expression)> */
		func() bool {
			position224, tokenIndex224, depth224 := position, tokenIndex, depth
			{
				position225 := position
				depth++
				if !_rules[rulews]() {
					goto l224
				}
				if buffer[position] != rune('|') {
					goto l224
				}
				position++
				if !_rules[rulews]() {
					goto l224
				}
				if !_rules[ruleName]() {
					goto l224
				}
			l226:
				{
					position227, tokenIndex227, depth227 := position, tokenIndex, depth
					if !_rules[ruleNextName]() {
						goto l227
					}
					goto l226
				l227:
					position, tokenIndex, depth = position227, tokenIndex227, depth227
				}
				if !_rules[rulews]() {
					goto l224
				}
				if buffer[position] != rune('|') {
					goto l224
				}
				position++
				if !_rules[rulews]() {
					goto l224
				}
				if buffer[position] != rune('-') {
					goto l224
				}
				position++
				if buffer[position] != rune('>') {
					goto l224
				}
				position++
				if !_rules[ruleExpression]() {
					goto l224
				}
				depth--
				add(ruleLambdaExpr, position225)
			}
			return true
		l224:
			position, tokenIndex, depth = position224, tokenIndex224, depth224
			return false
		},
		/* 58 NextName <- <(ws ',' ws Name)> */
		func() bool {
			position228, tokenIndex228, depth228 := position, tokenIndex, depth
			{
				position229 := position
				depth++
				if !_rules[rulews]() {
					goto l228
				}
				if buffer[position] != rune(',') {
					goto l228
				}
				position++
				if !_rules[rulews]() {
					goto l228
				}
				if !_rules[ruleName]() {
					goto l228
				}
				depth--
				add(ruleNextName, position229)
			}
			return true
		l228:
			position, tokenIndex, depth = position228, tokenIndex228, depth228
			return false
		},
		/* 59 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position230, tokenIndex230, depth230 := position, tokenIndex, depth
			{
				position231 := position
				depth++
				{
					position234, tokenIndex234, depth234 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l235
					}
					position++
					goto l234
				l235:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l236
					}
					position++
					goto l234
				l236:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l237
					}
					position++
					goto l234
				l237:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
					if buffer[position] != rune('_') {
						goto l230
					}
					position++
				}
			l234:
			l232:
				{
					position233, tokenIndex233, depth233 := position, tokenIndex, depth
					{
						position238, tokenIndex238, depth238 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l239
						}
						position++
						goto l238
					l239:
						position, tokenIndex, depth = position238, tokenIndex238, depth238
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l240
						}
						position++
						goto l238
					l240:
						position, tokenIndex, depth = position238, tokenIndex238, depth238
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l241
						}
						position++
						goto l238
					l241:
						position, tokenIndex, depth = position238, tokenIndex238, depth238
						if buffer[position] != rune('_') {
							goto l233
						}
						position++
					}
				l238:
					goto l232
				l233:
					position, tokenIndex, depth = position233, tokenIndex233, depth233
				}
				depth--
				add(ruleName, position231)
			}
			return true
		l230:
			position, tokenIndex, depth = position230, tokenIndex230, depth230
			return false
		},
		/* 60 Reference <- <('.'? Key ('.' (Key / Index))*)> */
		func() bool {
			position242, tokenIndex242, depth242 := position, tokenIndex, depth
			{
				position243 := position
				depth++
				{
					position244, tokenIndex244, depth244 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l244
					}
					position++
					goto l245
				l244:
					position, tokenIndex, depth = position244, tokenIndex244, depth244
				}
			l245:
				if !_rules[ruleKey]() {
					goto l242
				}
			l246:
				{
					position247, tokenIndex247, depth247 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l247
					}
					position++
					{
						position248, tokenIndex248, depth248 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l249
						}
						goto l248
					l249:
						position, tokenIndex, depth = position248, tokenIndex248, depth248
						if !_rules[ruleIndex]() {
							goto l247
						}
					}
				l248:
					goto l246
				l247:
					position, tokenIndex, depth = position247, tokenIndex247, depth247
				}
				depth--
				add(ruleReference, position243)
			}
			return true
		l242:
			position, tokenIndex, depth = position242, tokenIndex242, depth242
			return false
		},
		/* 61 FollowUpRef <- <((Key / Index) ('.' (Key / Index))*)> */
		func() bool {
			position250, tokenIndex250, depth250 := position, tokenIndex, depth
			{
				position251 := position
				depth++
				{
					position252, tokenIndex252, depth252 := position, tokenIndex, depth
					if !_rules[ruleKey]() {
						goto l253
					}
					goto l252
				l253:
					position, tokenIndex, depth = position252, tokenIndex252, depth252
					if !_rules[ruleIndex]() {
						goto l250
					}
				}
			l252:
			l254:
				{
					position255, tokenIndex255, depth255 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l255
					}
					position++
					{
						position256, tokenIndex256, depth256 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l257
						}
						goto l256
					l257:
						position, tokenIndex, depth = position256, tokenIndex256, depth256
						if !_rules[ruleIndex]() {
							goto l255
						}
					}
				l256:
					goto l254
				l255:
					position, tokenIndex, depth = position255, tokenIndex255, depth255
				}
				depth--
				add(ruleFollowUpRef, position251)
			}
			return true
		l250:
			position, tokenIndex, depth = position250, tokenIndex250, depth250
			return false
		},
		/* 62 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position258, tokenIndex258, depth258 := position, tokenIndex, depth
			{
				position259 := position
				depth++
				{
					position260, tokenIndex260, depth260 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l261
					}
					position++
					goto l260
				l261:
					position, tokenIndex, depth = position260, tokenIndex260, depth260
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l262
					}
					position++
					goto l260
				l262:
					position, tokenIndex, depth = position260, tokenIndex260, depth260
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l263
					}
					position++
					goto l260
				l263:
					position, tokenIndex, depth = position260, tokenIndex260, depth260
					if buffer[position] != rune('_') {
						goto l258
					}
					position++
				}
			l260:
			l264:
				{
					position265, tokenIndex265, depth265 := position, tokenIndex, depth
					{
						position266, tokenIndex266, depth266 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l267
						}
						position++
						goto l266
					l267:
						position, tokenIndex, depth = position266, tokenIndex266, depth266
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l268
						}
						position++
						goto l266
					l268:
						position, tokenIndex, depth = position266, tokenIndex266, depth266
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l269
						}
						position++
						goto l266
					l269:
						position, tokenIndex, depth = position266, tokenIndex266, depth266
						if buffer[position] != rune('_') {
							goto l270
						}
						position++
						goto l266
					l270:
						position, tokenIndex, depth = position266, tokenIndex266, depth266
						if buffer[position] != rune('-') {
							goto l265
						}
						position++
					}
				l266:
					goto l264
				l265:
					position, tokenIndex, depth = position265, tokenIndex265, depth265
				}
				{
					position271, tokenIndex271, depth271 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l271
					}
					position++
					{
						position273, tokenIndex273, depth273 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l274
						}
						position++
						goto l273
					l274:
						position, tokenIndex, depth = position273, tokenIndex273, depth273
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l275
						}
						position++
						goto l273
					l275:
						position, tokenIndex, depth = position273, tokenIndex273, depth273
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l276
						}
						position++
						goto l273
					l276:
						position, tokenIndex, depth = position273, tokenIndex273, depth273
						if buffer[position] != rune('_') {
							goto l271
						}
						position++
					}
				l273:
				l277:
					{
						position278, tokenIndex278, depth278 := position, tokenIndex, depth
						{
							position279, tokenIndex279, depth279 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l280
							}
							position++
							goto l279
						l280:
							position, tokenIndex, depth = position279, tokenIndex279, depth279
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l281
							}
							position++
							goto l279
						l281:
							position, tokenIndex, depth = position279, tokenIndex279, depth279
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l282
							}
							position++
							goto l279
						l282:
							position, tokenIndex, depth = position279, tokenIndex279, depth279
							if buffer[position] != rune('_') {
								goto l283
							}
							position++
							goto l279
						l283:
							position, tokenIndex, depth = position279, tokenIndex279, depth279
							if buffer[position] != rune('-') {
								goto l278
							}
							position++
						}
					l279:
						goto l277
					l278:
						position, tokenIndex, depth = position278, tokenIndex278, depth278
					}
					goto l272
				l271:
					position, tokenIndex, depth = position271, tokenIndex271, depth271
				}
			l272:
				depth--
				add(ruleKey, position259)
			}
			return true
		l258:
			position, tokenIndex, depth = position258, tokenIndex258, depth258
			return false
		},
		/* 63 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position284, tokenIndex284, depth284 := position, tokenIndex, depth
			{
				position285 := position
				depth++
				if buffer[position] != rune('[') {
					goto l284
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l284
				}
				position++
			l286:
				{
					position287, tokenIndex287, depth287 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l287
					}
					position++
					goto l286
				l287:
					position, tokenIndex, depth = position287, tokenIndex287, depth287
				}
				if buffer[position] != rune(']') {
					goto l284
				}
				position++
				depth--
				add(ruleIndex, position285)
			}
			return true
		l284:
			position, tokenIndex, depth = position284, tokenIndex284, depth284
			return false
		},
		/* 64 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position289 := position
				depth++
			l290:
				{
					position291, tokenIndex291, depth291 := position, tokenIndex, depth
					{
						position292, tokenIndex292, depth292 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l293
						}
						position++
						goto l292
					l293:
						position, tokenIndex, depth = position292, tokenIndex292, depth292
						if buffer[position] != rune('\t') {
							goto l294
						}
						position++
						goto l292
					l294:
						position, tokenIndex, depth = position292, tokenIndex292, depth292
						if buffer[position] != rune('\n') {
							goto l295
						}
						position++
						goto l292
					l295:
						position, tokenIndex, depth = position292, tokenIndex292, depth292
						if buffer[position] != rune('\r') {
							goto l291
						}
						position++
					}
				l292:
					goto l290
				l291:
					position, tokenIndex, depth = position291, tokenIndex291, depth291
				}
				depth--
				add(rulews, position289)
			}
			return true
		},
		/* 65 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position296, tokenIndex296, depth296 := position, tokenIndex, depth
			{
				position297 := position
				depth++
				{
					position300, tokenIndex300, depth300 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l301
					}
					position++
					goto l300
				l301:
					position, tokenIndex, depth = position300, tokenIndex300, depth300
					if buffer[position] != rune('\t') {
						goto l302
					}
					position++
					goto l300
				l302:
					position, tokenIndex, depth = position300, tokenIndex300, depth300
					if buffer[position] != rune('\n') {
						goto l303
					}
					position++
					goto l300
				l303:
					position, tokenIndex, depth = position300, tokenIndex300, depth300
					if buffer[position] != rune('\r') {
						goto l296
					}
					position++
				}
			l300:
			l298:
				{
					position299, tokenIndex299, depth299 := position, tokenIndex, depth
					{
						position304, tokenIndex304, depth304 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l305
						}
						position++
						goto l304
					l305:
						position, tokenIndex, depth = position304, tokenIndex304, depth304
						if buffer[position] != rune('\t') {
							goto l306
						}
						position++
						goto l304
					l306:
						position, tokenIndex, depth = position304, tokenIndex304, depth304
						if buffer[position] != rune('\n') {
							goto l307
						}
						position++
						goto l304
					l307:
						position, tokenIndex, depth = position304, tokenIndex304, depth304
						if buffer[position] != rune('\r') {
							goto l299
						}
						position++
					}
				l304:
					goto l298
				l299:
					position, tokenIndex, depth = position299, tokenIndex299, depth299
				}
				depth--
				add(rulereq_ws, position297)
			}
			return true
		l296:
			position, tokenIndex, depth = position296, tokenIndex296, depth296
			return false
		},
	}
	p.rules = _rules
}
