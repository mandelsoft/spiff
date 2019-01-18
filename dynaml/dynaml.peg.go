package dynaml

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleDynaml
	rulePrefer
	ruleMarkedExpression
	ruleSubsequentMarker
	ruleMarker
	ruleMarkerExpression
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
	ruleChainedRef
	ruleChainedDynRef
	ruleSlice
	ruleChainedCall
	ruleStartArguments
	ruleExpressionList
	ruleNextExpression
	ruleProjection
	ruleProjectionValue
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
	ruleStartList
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
	ruleIP
	rulews
	rulereq_ws
	ruleAction0

	rulePre
	ruleIn
	ruleSuf
)

var rul3s = [...]string{
	"Unknown",
	"Dynaml",
	"Prefer",
	"MarkedExpression",
	"SubsequentMarker",
	"Marker",
	"MarkerExpression",
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
	"ChainedRef",
	"ChainedDynRef",
	"Slice",
	"ChainedCall",
	"StartArguments",
	"ExpressionList",
	"NextExpression",
	"Projection",
	"ProjectionValue",
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
	"StartList",
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
	"IP",
	"ws",
	"req_ws",
	"Action0",

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

func (node *node32) Print(buffer string) {
	node.print(0, buffer)
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
		for i := range states {
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
							write(token32{pegRule: ruleIn, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre, begin: a.begin, end: b.begin}, true)
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
					write(token32{pegRule: ruleSuf, begin: b.end, end: a.end}, true)
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
	for i := range tokens {
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
	rules  [76]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	Pretty bool
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
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
	p   *DynamlGrammar
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *DynamlGrammar) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *DynamlGrammar) Highlighter() {
	p.tokenTree.PrintSyntax()
}

func (p *DynamlGrammar) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for token := range p.tokenTree.Tokens() {
		switch token.pegRule {

		case ruleAction0:

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *DynamlGrammar) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
		p.buffer = append(p.buffer, endSymbol)
	}

	var tree tokenTree = &tokens32{tree: make([]token32, math.MaxInt16)}
	var max token32
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
		return &parseError{p, max}
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
		if begin != position && position > max.end {
			max = token32{rule, begin, position, depth}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
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
		/* 2 MarkedExpression <- <(ws Marker (req_ws SubsequentMarker)* ws MarkerExpression? ws)> */
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
					if !_rules[ruleMarkerExpression]() {
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
		/* 4 Marker <- <('&' (('t' 'e' 'm' 'p' 'l' 'a' 't' 'e') / ('t' 'e' 'm' 'p' 'o' 'r' 'a' 'r' 'y') / ('l' 'o' 'c' 'a' 'l') / ('s' 't' 'a' 't' 'e')))> */
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
						goto l20
					}
					position++
					if buffer[position] != rune('e') {
						goto l20
					}
					position++
					if buffer[position] != rune('m') {
						goto l20
					}
					position++
					if buffer[position] != rune('p') {
						goto l20
					}
					position++
					if buffer[position] != rune('o') {
						goto l20
					}
					position++
					if buffer[position] != rune('r') {
						goto l20
					}
					position++
					if buffer[position] != rune('a') {
						goto l20
					}
					position++
					if buffer[position] != rune('r') {
						goto l20
					}
					position++
					if buffer[position] != rune('y') {
						goto l20
					}
					position++
					goto l18
				l20:
					position, tokenIndex, depth = position18, tokenIndex18, depth18
					if buffer[position] != rune('l') {
						goto l21
					}
					position++
					if buffer[position] != rune('o') {
						goto l21
					}
					position++
					if buffer[position] != rune('c') {
						goto l21
					}
					position++
					if buffer[position] != rune('a') {
						goto l21
					}
					position++
					if buffer[position] != rune('l') {
						goto l21
					}
					position++
					goto l18
				l21:
					position, tokenIndex, depth = position18, tokenIndex18, depth18
					if buffer[position] != rune('s') {
						goto l16
					}
					position++
					if buffer[position] != rune('t') {
						goto l16
					}
					position++
					if buffer[position] != rune('a') {
						goto l16
					}
					position++
					if buffer[position] != rune('t') {
						goto l16
					}
					position++
					if buffer[position] != rune('e') {
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
		/* 5 MarkerExpression <- <Grouped> */
		func() bool {
			position22, tokenIndex22, depth22 := position, tokenIndex, depth
			{
				position23 := position
				depth++
				if !_rules[ruleGrouped]() {
					goto l22
				}
				depth--
				add(ruleMarkerExpression, position23)
			}
			return true
		l22:
			position, tokenIndex, depth = position22, tokenIndex22, depth22
			return false
		},
		/* 6 Expression <- <(ws (LambdaExpr / Level7) ws)> */
		func() bool {
			position24, tokenIndex24, depth24 := position, tokenIndex, depth
			{
				position25 := position
				depth++
				if !_rules[rulews]() {
					goto l24
				}
				{
					position26, tokenIndex26, depth26 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l27
					}
					goto l26
				l27:
					position, tokenIndex, depth = position26, tokenIndex26, depth26
					if !_rules[ruleLevel7]() {
						goto l24
					}
				}
			l26:
				if !_rules[rulews]() {
					goto l24
				}
				depth--
				add(ruleExpression, position25)
			}
			return true
		l24:
			position, tokenIndex, depth = position24, tokenIndex24, depth24
			return false
		},
		/* 7 Level7 <- <(Level6 (req_ws Or)*)> */
		func() bool {
			position28, tokenIndex28, depth28 := position, tokenIndex, depth
			{
				position29 := position
				depth++
				if !_rules[ruleLevel6]() {
					goto l28
				}
			l30:
				{
					position31, tokenIndex31, depth31 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l31
					}
					if !_rules[ruleOr]() {
						goto l31
					}
					goto l30
				l31:
					position, tokenIndex, depth = position31, tokenIndex31, depth31
				}
				depth--
				add(ruleLevel7, position29)
			}
			return true
		l28:
			position, tokenIndex, depth = position28, tokenIndex28, depth28
			return false
		},
		/* 8 Or <- <('|' '|' req_ws Level6)> */
		func() bool {
			position32, tokenIndex32, depth32 := position, tokenIndex, depth
			{
				position33 := position
				depth++
				if buffer[position] != rune('|') {
					goto l32
				}
				position++
				if buffer[position] != rune('|') {
					goto l32
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l32
				}
				if !_rules[ruleLevel6]() {
					goto l32
				}
				depth--
				add(ruleOr, position33)
			}
			return true
		l32:
			position, tokenIndex, depth = position32, tokenIndex32, depth32
			return false
		},
		/* 9 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position34, tokenIndex34, depth34 := position, tokenIndex, depth
			{
				position35 := position
				depth++
				{
					position36, tokenIndex36, depth36 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l37
					}
					goto l36
				l37:
					position, tokenIndex, depth = position36, tokenIndex36, depth36
					if !_rules[ruleLevel5]() {
						goto l34
					}
				}
			l36:
				depth--
				add(ruleLevel6, position35)
			}
			return true
		l34:
			position, tokenIndex, depth = position34, tokenIndex34, depth34
			return false
		},
		/* 10 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position38, tokenIndex38, depth38 := position, tokenIndex, depth
			{
				position39 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l38
				}
				if !_rules[rulews]() {
					goto l38
				}
				if buffer[position] != rune('?') {
					goto l38
				}
				position++
				if !_rules[ruleExpression]() {
					goto l38
				}
				if buffer[position] != rune(':') {
					goto l38
				}
				position++
				if !_rules[ruleExpression]() {
					goto l38
				}
				depth--
				add(ruleConditional, position39)
			}
			return true
		l38:
			position, tokenIndex, depth = position38, tokenIndex38, depth38
			return false
		},
		/* 11 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l40
				}
			l42:
				{
					position43, tokenIndex43, depth43 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l43
					}
					goto l42
				l43:
					position, tokenIndex, depth = position43, tokenIndex43, depth43
				}
				depth--
				add(ruleLevel5, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 12 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position44, tokenIndex44, depth44 := position, tokenIndex, depth
			{
				position45 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l44
				}
				if !_rules[ruleLevel4]() {
					goto l44
				}
				depth--
				add(ruleConcatenation, position45)
			}
			return true
		l44:
			position, tokenIndex, depth = position44, tokenIndex44, depth44
			return false
		},
		/* 13 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position46, tokenIndex46, depth46 := position, tokenIndex, depth
			{
				position47 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l46
				}
			l48:
				{
					position49, tokenIndex49, depth49 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l49
					}
					{
						position50, tokenIndex50, depth50 := position, tokenIndex, depth
						if !_rules[ruleLogOr]() {
							goto l51
						}
						goto l50
					l51:
						position, tokenIndex, depth = position50, tokenIndex50, depth50
						if !_rules[ruleLogAnd]() {
							goto l49
						}
					}
				l50:
					goto l48
				l49:
					position, tokenIndex, depth = position49, tokenIndex49, depth49
				}
				depth--
				add(ruleLevel4, position47)
			}
			return true
		l46:
			position, tokenIndex, depth = position46, tokenIndex46, depth46
			return false
		},
		/* 14 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position52, tokenIndex52, depth52 := position, tokenIndex, depth
			{
				position53 := position
				depth++
				if buffer[position] != rune('-') {
					goto l52
				}
				position++
				if buffer[position] != rune('o') {
					goto l52
				}
				position++
				if buffer[position] != rune('r') {
					goto l52
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l52
				}
				if !_rules[ruleLevel3]() {
					goto l52
				}
				depth--
				add(ruleLogOr, position53)
			}
			return true
		l52:
			position, tokenIndex, depth = position52, tokenIndex52, depth52
			return false
		},
		/* 15 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position54, tokenIndex54, depth54 := position, tokenIndex, depth
			{
				position55 := position
				depth++
				if buffer[position] != rune('-') {
					goto l54
				}
				position++
				if buffer[position] != rune('a') {
					goto l54
				}
				position++
				if buffer[position] != rune('n') {
					goto l54
				}
				position++
				if buffer[position] != rune('d') {
					goto l54
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l54
				}
				if !_rules[ruleLevel3]() {
					goto l54
				}
				depth--
				add(ruleLogAnd, position55)
			}
			return true
		l54:
			position, tokenIndex, depth = position54, tokenIndex54, depth54
			return false
		},
		/* 16 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position56, tokenIndex56, depth56 := position, tokenIndex, depth
			{
				position57 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l56
				}
			l58:
				{
					position59, tokenIndex59, depth59 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l59
					}
					if !_rules[ruleComparison]() {
						goto l59
					}
					goto l58
				l59:
					position, tokenIndex, depth = position59, tokenIndex59, depth59
				}
				depth--
				add(ruleLevel3, position57)
			}
			return true
		l56:
			position, tokenIndex, depth = position56, tokenIndex56, depth56
			return false
		},
		/* 17 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position60, tokenIndex60, depth60 := position, tokenIndex, depth
			{
				position61 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l60
				}
				if !_rules[rulereq_ws]() {
					goto l60
				}
				if !_rules[ruleLevel2]() {
					goto l60
				}
				depth--
				add(ruleComparison, position61)
			}
			return true
		l60:
			position, tokenIndex, depth = position60, tokenIndex60, depth60
			return false
		},
		/* 18 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position62, tokenIndex62, depth62 := position, tokenIndex, depth
			{
				position63 := position
				depth++
				{
					position64, tokenIndex64, depth64 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l65
					}
					position++
					if buffer[position] != rune('=') {
						goto l65
					}
					position++
					goto l64
				l65:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
					if buffer[position] != rune('!') {
						goto l66
					}
					position++
					if buffer[position] != rune('=') {
						goto l66
					}
					position++
					goto l64
				l66:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
					if buffer[position] != rune('<') {
						goto l67
					}
					position++
					if buffer[position] != rune('=') {
						goto l67
					}
					position++
					goto l64
				l67:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
					if buffer[position] != rune('>') {
						goto l68
					}
					position++
					if buffer[position] != rune('=') {
						goto l68
					}
					position++
					goto l64
				l68:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
					if buffer[position] != rune('>') {
						goto l69
					}
					position++
					goto l64
				l69:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
					if buffer[position] != rune('<') {
						goto l70
					}
					position++
					goto l64
				l70:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
					if buffer[position] != rune('>') {
						goto l62
					}
					position++
				}
			l64:
				depth--
				add(ruleCompareOp, position63)
			}
			return true
		l62:
			position, tokenIndex, depth = position62, tokenIndex62, depth62
			return false
		},
		/* 19 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position71, tokenIndex71, depth71 := position, tokenIndex, depth
			{
				position72 := position
				depth++
				if !_rules[ruleLevel1]() {
					goto l71
				}
			l73:
				{
					position74, tokenIndex74, depth74 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l74
					}
					{
						position75, tokenIndex75, depth75 := position, tokenIndex, depth
						if !_rules[ruleAddition]() {
							goto l76
						}
						goto l75
					l76:
						position, tokenIndex, depth = position75, tokenIndex75, depth75
						if !_rules[ruleSubtraction]() {
							goto l74
						}
					}
				l75:
					goto l73
				l74:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
				}
				depth--
				add(ruleLevel2, position72)
			}
			return true
		l71:
			position, tokenIndex, depth = position71, tokenIndex71, depth71
			return false
		},
		/* 20 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position77, tokenIndex77, depth77 := position, tokenIndex, depth
			{
				position78 := position
				depth++
				if buffer[position] != rune('+') {
					goto l77
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l77
				}
				if !_rules[ruleLevel1]() {
					goto l77
				}
				depth--
				add(ruleAddition, position78)
			}
			return true
		l77:
			position, tokenIndex, depth = position77, tokenIndex77, depth77
			return false
		},
		/* 21 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position79, tokenIndex79, depth79 := position, tokenIndex, depth
			{
				position80 := position
				depth++
				if buffer[position] != rune('-') {
					goto l79
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l79
				}
				if !_rules[ruleLevel1]() {
					goto l79
				}
				depth--
				add(ruleSubtraction, position80)
			}
			return true
		l79:
			position, tokenIndex, depth = position79, tokenIndex79, depth79
			return false
		},
		/* 22 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position81, tokenIndex81, depth81 := position, tokenIndex, depth
			{
				position82 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l81
				}
			l83:
				{
					position84, tokenIndex84, depth84 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l84
					}
					{
						position85, tokenIndex85, depth85 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l86
						}
						goto l85
					l86:
						position, tokenIndex, depth = position85, tokenIndex85, depth85
						if !_rules[ruleDivision]() {
							goto l87
						}
						goto l85
					l87:
						position, tokenIndex, depth = position85, tokenIndex85, depth85
						if !_rules[ruleModulo]() {
							goto l84
						}
					}
				l85:
					goto l83
				l84:
					position, tokenIndex, depth = position84, tokenIndex84, depth84
				}
				depth--
				add(ruleLevel1, position82)
			}
			return true
		l81:
			position, tokenIndex, depth = position81, tokenIndex81, depth81
			return false
		},
		/* 23 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position88, tokenIndex88, depth88 := position, tokenIndex, depth
			{
				position89 := position
				depth++
				if buffer[position] != rune('*') {
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
				add(ruleMultiplication, position89)
			}
			return true
		l88:
			position, tokenIndex, depth = position88, tokenIndex88, depth88
			return false
		},
		/* 24 Division <- <('/' req_ws Level0)> */
		func() bool {
			position90, tokenIndex90, depth90 := position, tokenIndex, depth
			{
				position91 := position
				depth++
				if buffer[position] != rune('/') {
					goto l90
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l90
				}
				if !_rules[ruleLevel0]() {
					goto l90
				}
				depth--
				add(ruleDivision, position91)
			}
			return true
		l90:
			position, tokenIndex, depth = position90, tokenIndex90, depth90
			return false
		},
		/* 25 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position92, tokenIndex92, depth92 := position, tokenIndex, depth
			{
				position93 := position
				depth++
				if buffer[position] != rune('%') {
					goto l92
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l92
				}
				if !_rules[ruleLevel0]() {
					goto l92
				}
				depth--
				add(ruleModulo, position93)
			}
			return true
		l92:
			position, tokenIndex, depth = position92, tokenIndex92, depth92
			return false
		},
		/* 26 Level0 <- <(IP / String / Integer / Boolean / Undefined / Nil / Not / Substitution / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position94, tokenIndex94, depth94 := position, tokenIndex, depth
			{
				position95 := position
				depth++
				{
					position96, tokenIndex96, depth96 := position, tokenIndex, depth
					if !_rules[ruleIP]() {
						goto l97
					}
					goto l96
				l97:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
					if !_rules[ruleString]() {
						goto l98
					}
					goto l96
				l98:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
					if !_rules[ruleInteger]() {
						goto l99
					}
					goto l96
				l99:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
					if !_rules[ruleBoolean]() {
						goto l100
					}
					goto l96
				l100:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
					if !_rules[ruleUndefined]() {
						goto l101
					}
					goto l96
				l101:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
					if !_rules[ruleNil]() {
						goto l102
					}
					goto l96
				l102:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
					if !_rules[ruleNot]() {
						goto l103
					}
					goto l96
				l103:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
					if !_rules[ruleSubstitution]() {
						goto l104
					}
					goto l96
				l104:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
					if !_rules[ruleMerge]() {
						goto l105
					}
					goto l96
				l105:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
					if !_rules[ruleAuto]() {
						goto l106
					}
					goto l96
				l106:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
					if !_rules[ruleLambda]() {
						goto l107
					}
					goto l96
				l107:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
					if !_rules[ruleChained]() {
						goto l94
					}
				}
			l96:
				depth--
				add(ruleLevel0, position95)
			}
			return true
		l94:
			position, tokenIndex, depth = position94, tokenIndex94, depth94
			return false
		},
		/* 27 Chained <- <((Mapping / Sum / List / Map / Range / Grouped / Reference) ChainedQualifiedExpression*)> */
		func() bool {
			position108, tokenIndex108, depth108 := position, tokenIndex, depth
			{
				position109 := position
				depth++
				{
					position110, tokenIndex110, depth110 := position, tokenIndex, depth
					if !_rules[ruleMapping]() {
						goto l111
					}
					goto l110
				l111:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleSum]() {
						goto l112
					}
					goto l110
				l112:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleList]() {
						goto l113
					}
					goto l110
				l113:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleMap]() {
						goto l114
					}
					goto l110
				l114:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleRange]() {
						goto l115
					}
					goto l110
				l115:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleGrouped]() {
						goto l116
					}
					goto l110
				l116:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleReference]() {
						goto l108
					}
				}
			l110:
			l117:
				{
					position118, tokenIndex118, depth118 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l118
					}
					goto l117
				l118:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
				}
				depth--
				add(ruleChained, position109)
			}
			return true
		l108:
			position, tokenIndex, depth = position108, tokenIndex108, depth108
			return false
		},
		/* 28 ChainedQualifiedExpression <- <(ChainedCall / ('.' (ChainedRef / ChainedDynRef / Projection)))> */
		func() bool {
			position119, tokenIndex119, depth119 := position, tokenIndex, depth
			{
				position120 := position
				depth++
				{
					position121, tokenIndex121, depth121 := position, tokenIndex, depth
					if !_rules[ruleChainedCall]() {
						goto l122
					}
					goto l121
				l122:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
					if buffer[position] != rune('.') {
						goto l119
					}
					position++
					{
						position123, tokenIndex123, depth123 := position, tokenIndex, depth
						if !_rules[ruleChainedRef]() {
							goto l124
						}
						goto l123
					l124:
						position, tokenIndex, depth = position123, tokenIndex123, depth123
						if !_rules[ruleChainedDynRef]() {
							goto l125
						}
						goto l123
					l125:
						position, tokenIndex, depth = position123, tokenIndex123, depth123
						if !_rules[ruleProjection]() {
							goto l119
						}
					}
				l123:
				}
			l121:
				depth--
				add(ruleChainedQualifiedExpression, position120)
			}
			return true
		l119:
			position, tokenIndex, depth = position119, tokenIndex119, depth119
			return false
		},
		/* 29 ChainedRef <- <((Key / Index) FollowUpRef)> */
		func() bool {
			position126, tokenIndex126, depth126 := position, tokenIndex, depth
			{
				position127 := position
				depth++
				{
					position128, tokenIndex128, depth128 := position, tokenIndex, depth
					if !_rules[ruleKey]() {
						goto l129
					}
					goto l128
				l129:
					position, tokenIndex, depth = position128, tokenIndex128, depth128
					if !_rules[ruleIndex]() {
						goto l126
					}
				}
			l128:
				if !_rules[ruleFollowUpRef]() {
					goto l126
				}
				depth--
				add(ruleChainedRef, position127)
			}
			return true
		l126:
			position, tokenIndex, depth = position126, tokenIndex126, depth126
			return false
		},
		/* 30 ChainedDynRef <- <('[' Expression ']')> */
		func() bool {
			position130, tokenIndex130, depth130 := position, tokenIndex, depth
			{
				position131 := position
				depth++
				if buffer[position] != rune('[') {
					goto l130
				}
				position++
				if !_rules[ruleExpression]() {
					goto l130
				}
				if buffer[position] != rune(']') {
					goto l130
				}
				position++
				depth--
				add(ruleChainedDynRef, position131)
			}
			return true
		l130:
			position, tokenIndex, depth = position130, tokenIndex130, depth130
			return false
		},
		/* 31 Slice <- <Range> */
		func() bool {
			position132, tokenIndex132, depth132 := position, tokenIndex, depth
			{
				position133 := position
				depth++
				if !_rules[ruleRange]() {
					goto l132
				}
				depth--
				add(ruleSlice, position133)
			}
			return true
		l132:
			position, tokenIndex, depth = position132, tokenIndex132, depth132
			return false
		},
		/* 32 ChainedCall <- <(StartArguments ExpressionList? ')')> */
		func() bool {
			position134, tokenIndex134, depth134 := position, tokenIndex, depth
			{
				position135 := position
				depth++
				if !_rules[ruleStartArguments]() {
					goto l134
				}
				{
					position136, tokenIndex136, depth136 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l136
					}
					goto l137
				l136:
					position, tokenIndex, depth = position136, tokenIndex136, depth136
				}
			l137:
				if buffer[position] != rune(')') {
					goto l134
				}
				position++
				depth--
				add(ruleChainedCall, position135)
			}
			return true
		l134:
			position, tokenIndex, depth = position134, tokenIndex134, depth134
			return false
		},
		/* 33 StartArguments <- <('(' ws)> */
		func() bool {
			position138, tokenIndex138, depth138 := position, tokenIndex, depth
			{
				position139 := position
				depth++
				if buffer[position] != rune('(') {
					goto l138
				}
				position++
				if !_rules[rulews]() {
					goto l138
				}
				depth--
				add(ruleStartArguments, position139)
			}
			return true
		l138:
			position, tokenIndex, depth = position138, tokenIndex138, depth138
			return false
		},
		/* 34 ExpressionList <- <(NextExpression (',' NextExpression)*)> */
		func() bool {
			position140, tokenIndex140, depth140 := position, tokenIndex, depth
			{
				position141 := position
				depth++
				if !_rules[ruleNextExpression]() {
					goto l140
				}
			l142:
				{
					position143, tokenIndex143, depth143 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l143
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l143
					}
					goto l142
				l143:
					position, tokenIndex, depth = position143, tokenIndex143, depth143
				}
				depth--
				add(ruleExpressionList, position141)
			}
			return true
		l140:
			position, tokenIndex, depth = position140, tokenIndex140, depth140
			return false
		},
		/* 35 NextExpression <- <Expression> */
		func() bool {
			position144, tokenIndex144, depth144 := position, tokenIndex, depth
			{
				position145 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l144
				}
				depth--
				add(ruleNextExpression, position145)
			}
			return true
		l144:
			position, tokenIndex, depth = position144, tokenIndex144, depth144
			return false
		},
		/* 36 Projection <- <((('[' '*' ']') / Slice) ProjectionValue ChainedQualifiedExpression*)> */
		func() bool {
			position146, tokenIndex146, depth146 := position, tokenIndex, depth
			{
				position147 := position
				depth++
				{
					position148, tokenIndex148, depth148 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l149
					}
					position++
					if buffer[position] != rune('*') {
						goto l149
					}
					position++
					if buffer[position] != rune(']') {
						goto l149
					}
					position++
					goto l148
				l149:
					position, tokenIndex, depth = position148, tokenIndex148, depth148
					if !_rules[ruleSlice]() {
						goto l146
					}
				}
			l148:
				if !_rules[ruleProjectionValue]() {
					goto l146
				}
			l150:
				{
					position151, tokenIndex151, depth151 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l151
					}
					goto l150
				l151:
					position, tokenIndex, depth = position151, tokenIndex151, depth151
				}
				depth--
				add(ruleProjection, position147)
			}
			return true
		l146:
			position, tokenIndex, depth = position146, tokenIndex146, depth146
			return false
		},
		/* 37 ProjectionValue <- <Action0> */
		func() bool {
			position152, tokenIndex152, depth152 := position, tokenIndex, depth
			{
				position153 := position
				depth++
				if !_rules[ruleAction0]() {
					goto l152
				}
				depth--
				add(ruleProjectionValue, position153)
			}
			return true
		l152:
			position, tokenIndex, depth = position152, tokenIndex152, depth152
			return false
		},
		/* 38 Substitution <- <('*' Level0)> */
		func() bool {
			position154, tokenIndex154, depth154 := position, tokenIndex, depth
			{
				position155 := position
				depth++
				if buffer[position] != rune('*') {
					goto l154
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l154
				}
				depth--
				add(ruleSubstitution, position155)
			}
			return true
		l154:
			position, tokenIndex, depth = position154, tokenIndex154, depth154
			return false
		},
		/* 39 Not <- <('!' ws Level0)> */
		func() bool {
			position156, tokenIndex156, depth156 := position, tokenIndex, depth
			{
				position157 := position
				depth++
				if buffer[position] != rune('!') {
					goto l156
				}
				position++
				if !_rules[rulews]() {
					goto l156
				}
				if !_rules[ruleLevel0]() {
					goto l156
				}
				depth--
				add(ruleNot, position157)
			}
			return true
		l156:
			position, tokenIndex, depth = position156, tokenIndex156, depth156
			return false
		},
		/* 40 Grouped <- <('(' Expression ')')> */
		func() bool {
			position158, tokenIndex158, depth158 := position, tokenIndex, depth
			{
				position159 := position
				depth++
				if buffer[position] != rune('(') {
					goto l158
				}
				position++
				if !_rules[ruleExpression]() {
					goto l158
				}
				if buffer[position] != rune(')') {
					goto l158
				}
				position++
				depth--
				add(ruleGrouped, position159)
			}
			return true
		l158:
			position, tokenIndex, depth = position158, tokenIndex158, depth158
			return false
		},
		/* 41 Range <- <('[' Expression ('.' '.') Expression ']')> */
		func() bool {
			position160, tokenIndex160, depth160 := position, tokenIndex, depth
			{
				position161 := position
				depth++
				if buffer[position] != rune('[') {
					goto l160
				}
				position++
				if !_rules[ruleExpression]() {
					goto l160
				}
				if buffer[position] != rune('.') {
					goto l160
				}
				position++
				if buffer[position] != rune('.') {
					goto l160
				}
				position++
				if !_rules[ruleExpression]() {
					goto l160
				}
				if buffer[position] != rune(']') {
					goto l160
				}
				position++
				depth--
				add(ruleRange, position161)
			}
			return true
		l160:
			position, tokenIndex, depth = position160, tokenIndex160, depth160
			return false
		},
		/* 42 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position162, tokenIndex162, depth162 := position, tokenIndex, depth
			{
				position163 := position
				depth++
				{
					position164, tokenIndex164, depth164 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l164
					}
					position++
					goto l165
				l164:
					position, tokenIndex, depth = position164, tokenIndex164, depth164
				}
			l165:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l162
				}
				position++
			l166:
				{
					position167, tokenIndex167, depth167 := position, tokenIndex, depth
					{
						position168, tokenIndex168, depth168 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l169
						}
						position++
						goto l168
					l169:
						position, tokenIndex, depth = position168, tokenIndex168, depth168
						if buffer[position] != rune('_') {
							goto l167
						}
						position++
					}
				l168:
					goto l166
				l167:
					position, tokenIndex, depth = position167, tokenIndex167, depth167
				}
				depth--
				add(ruleInteger, position163)
			}
			return true
		l162:
			position, tokenIndex, depth = position162, tokenIndex162, depth162
			return false
		},
		/* 43 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position170, tokenIndex170, depth170 := position, tokenIndex, depth
			{
				position171 := position
				depth++
				if buffer[position] != rune('"') {
					goto l170
				}
				position++
			l172:
				{
					position173, tokenIndex173, depth173 := position, tokenIndex, depth
					{
						position174, tokenIndex174, depth174 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l175
						}
						position++
						if buffer[position] != rune('"') {
							goto l175
						}
						position++
						goto l174
					l175:
						position, tokenIndex, depth = position174, tokenIndex174, depth174
						{
							position176, tokenIndex176, depth176 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l176
							}
							position++
							goto l173
						l176:
							position, tokenIndex, depth = position176, tokenIndex176, depth176
						}
						if !matchDot() {
							goto l173
						}
					}
				l174:
					goto l172
				l173:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
				}
				if buffer[position] != rune('"') {
					goto l170
				}
				position++
				depth--
				add(ruleString, position171)
			}
			return true
		l170:
			position, tokenIndex, depth = position170, tokenIndex170, depth170
			return false
		},
		/* 44 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position177, tokenIndex177, depth177 := position, tokenIndex, depth
			{
				position178 := position
				depth++
				{
					position179, tokenIndex179, depth179 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l180
					}
					position++
					if buffer[position] != rune('r') {
						goto l180
					}
					position++
					if buffer[position] != rune('u') {
						goto l180
					}
					position++
					if buffer[position] != rune('e') {
						goto l180
					}
					position++
					goto l179
				l180:
					position, tokenIndex, depth = position179, tokenIndex179, depth179
					if buffer[position] != rune('f') {
						goto l177
					}
					position++
					if buffer[position] != rune('a') {
						goto l177
					}
					position++
					if buffer[position] != rune('l') {
						goto l177
					}
					position++
					if buffer[position] != rune('s') {
						goto l177
					}
					position++
					if buffer[position] != rune('e') {
						goto l177
					}
					position++
				}
			l179:
				depth--
				add(ruleBoolean, position178)
			}
			return true
		l177:
			position, tokenIndex, depth = position177, tokenIndex177, depth177
			return false
		},
		/* 45 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position181, tokenIndex181, depth181 := position, tokenIndex, depth
			{
				position182 := position
				depth++
				{
					position183, tokenIndex183, depth183 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l184
					}
					position++
					if buffer[position] != rune('i') {
						goto l184
					}
					position++
					if buffer[position] != rune('l') {
						goto l184
					}
					position++
					goto l183
				l184:
					position, tokenIndex, depth = position183, tokenIndex183, depth183
					if buffer[position] != rune('~') {
						goto l181
					}
					position++
				}
			l183:
				depth--
				add(ruleNil, position182)
			}
			return true
		l181:
			position, tokenIndex, depth = position181, tokenIndex181, depth181
			return false
		},
		/* 46 Undefined <- <('~' '~')> */
		func() bool {
			position185, tokenIndex185, depth185 := position, tokenIndex, depth
			{
				position186 := position
				depth++
				if buffer[position] != rune('~') {
					goto l185
				}
				position++
				if buffer[position] != rune('~') {
					goto l185
				}
				position++
				depth--
				add(ruleUndefined, position186)
			}
			return true
		l185:
			position, tokenIndex, depth = position185, tokenIndex185, depth185
			return false
		},
		/* 47 List <- <(StartList ExpressionList? ']')> */
		func() bool {
			position187, tokenIndex187, depth187 := position, tokenIndex, depth
			{
				position188 := position
				depth++
				if !_rules[ruleStartList]() {
					goto l187
				}
				{
					position189, tokenIndex189, depth189 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l189
					}
					goto l190
				l189:
					position, tokenIndex, depth = position189, tokenIndex189, depth189
				}
			l190:
				if buffer[position] != rune(']') {
					goto l187
				}
				position++
				depth--
				add(ruleList, position188)
			}
			return true
		l187:
			position, tokenIndex, depth = position187, tokenIndex187, depth187
			return false
		},
		/* 48 StartList <- <'['> */
		func() bool {
			position191, tokenIndex191, depth191 := position, tokenIndex, depth
			{
				position192 := position
				depth++
				if buffer[position] != rune('[') {
					goto l191
				}
				position++
				depth--
				add(ruleStartList, position192)
			}
			return true
		l191:
			position, tokenIndex, depth = position191, tokenIndex191, depth191
			return false
		},
		/* 49 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position193, tokenIndex193, depth193 := position, tokenIndex, depth
			{
				position194 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l193
				}
				if !_rules[rulews]() {
					goto l193
				}
				{
					position195, tokenIndex195, depth195 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l195
					}
					goto l196
				l195:
					position, tokenIndex, depth = position195, tokenIndex195, depth195
				}
			l196:
				if buffer[position] != rune('}') {
					goto l193
				}
				position++
				depth--
				add(ruleMap, position194)
			}
			return true
		l193:
			position, tokenIndex, depth = position193, tokenIndex193, depth193
			return false
		},
		/* 50 CreateMap <- <'{'> */
		func() bool {
			position197, tokenIndex197, depth197 := position, tokenIndex, depth
			{
				position198 := position
				depth++
				if buffer[position] != rune('{') {
					goto l197
				}
				position++
				depth--
				add(ruleCreateMap, position198)
			}
			return true
		l197:
			position, tokenIndex, depth = position197, tokenIndex197, depth197
			return false
		},
		/* 51 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position199, tokenIndex199, depth199 := position, tokenIndex, depth
			{
				position200 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l199
				}
			l201:
				{
					position202, tokenIndex202, depth202 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l202
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l202
					}
					goto l201
				l202:
					position, tokenIndex, depth = position202, tokenIndex202, depth202
				}
				depth--
				add(ruleAssignments, position200)
			}
			return true
		l199:
			position, tokenIndex, depth = position199, tokenIndex199, depth199
			return false
		},
		/* 52 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position203, tokenIndex203, depth203 := position, tokenIndex, depth
			{
				position204 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l203
				}
				if buffer[position] != rune('=') {
					goto l203
				}
				position++
				if !_rules[ruleExpression]() {
					goto l203
				}
				depth--
				add(ruleAssignment, position204)
			}
			return true
		l203:
			position, tokenIndex, depth = position203, tokenIndex203, depth203
			return false
		},
		/* 53 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position205, tokenIndex205, depth205 := position, tokenIndex, depth
			{
				position206 := position
				depth++
				{
					position207, tokenIndex207, depth207 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l208
					}
					goto l207
				l208:
					position, tokenIndex, depth = position207, tokenIndex207, depth207
					if !_rules[ruleSimpleMerge]() {
						goto l205
					}
				}
			l207:
				depth--
				add(ruleMerge, position206)
			}
			return true
		l205:
			position, tokenIndex, depth = position205, tokenIndex205, depth205
			return false
		},
		/* 54 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position209, tokenIndex209, depth209 := position, tokenIndex, depth
			{
				position210 := position
				depth++
				if buffer[position] != rune('m') {
					goto l209
				}
				position++
				if buffer[position] != rune('e') {
					goto l209
				}
				position++
				if buffer[position] != rune('r') {
					goto l209
				}
				position++
				if buffer[position] != rune('g') {
					goto l209
				}
				position++
				if buffer[position] != rune('e') {
					goto l209
				}
				position++
				{
					position211, tokenIndex211, depth211 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l211
					}
					if !_rules[ruleRequired]() {
						goto l211
					}
					goto l209
				l211:
					position, tokenIndex, depth = position211, tokenIndex211, depth211
				}
				{
					position212, tokenIndex212, depth212 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l212
					}
					{
						position214, tokenIndex214, depth214 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l215
						}
						goto l214
					l215:
						position, tokenIndex, depth = position214, tokenIndex214, depth214
						if !_rules[ruleOn]() {
							goto l212
						}
					}
				l214:
					goto l213
				l212:
					position, tokenIndex, depth = position212, tokenIndex212, depth212
				}
			l213:
				if !_rules[rulereq_ws]() {
					goto l209
				}
				if !_rules[ruleReference]() {
					goto l209
				}
				depth--
				add(ruleRefMerge, position210)
			}
			return true
		l209:
			position, tokenIndex, depth = position209, tokenIndex209, depth209
			return false
		},
		/* 55 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' !'(' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position216, tokenIndex216, depth216 := position, tokenIndex, depth
			{
				position217 := position
				depth++
				if buffer[position] != rune('m') {
					goto l216
				}
				position++
				if buffer[position] != rune('e') {
					goto l216
				}
				position++
				if buffer[position] != rune('r') {
					goto l216
				}
				position++
				if buffer[position] != rune('g') {
					goto l216
				}
				position++
				if buffer[position] != rune('e') {
					goto l216
				}
				position++
				{
					position218, tokenIndex218, depth218 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l218
					}
					position++
					goto l216
				l218:
					position, tokenIndex, depth = position218, tokenIndex218, depth218
				}
				{
					position219, tokenIndex219, depth219 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l219
					}
					{
						position221, tokenIndex221, depth221 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l222
						}
						goto l221
					l222:
						position, tokenIndex, depth = position221, tokenIndex221, depth221
						if !_rules[ruleRequired]() {
							goto l223
						}
						goto l221
					l223:
						position, tokenIndex, depth = position221, tokenIndex221, depth221
						if !_rules[ruleOn]() {
							goto l219
						}
					}
				l221:
					goto l220
				l219:
					position, tokenIndex, depth = position219, tokenIndex219, depth219
				}
			l220:
				depth--
				add(ruleSimpleMerge, position217)
			}
			return true
		l216:
			position, tokenIndex, depth = position216, tokenIndex216, depth216
			return false
		},
		/* 56 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position224, tokenIndex224, depth224 := position, tokenIndex, depth
			{
				position225 := position
				depth++
				if buffer[position] != rune('r') {
					goto l224
				}
				position++
				if buffer[position] != rune('e') {
					goto l224
				}
				position++
				if buffer[position] != rune('p') {
					goto l224
				}
				position++
				if buffer[position] != rune('l') {
					goto l224
				}
				position++
				if buffer[position] != rune('a') {
					goto l224
				}
				position++
				if buffer[position] != rune('c') {
					goto l224
				}
				position++
				if buffer[position] != rune('e') {
					goto l224
				}
				position++
				depth--
				add(ruleReplace, position225)
			}
			return true
		l224:
			position, tokenIndex, depth = position224, tokenIndex224, depth224
			return false
		},
		/* 57 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position226, tokenIndex226, depth226 := position, tokenIndex, depth
			{
				position227 := position
				depth++
				if buffer[position] != rune('r') {
					goto l226
				}
				position++
				if buffer[position] != rune('e') {
					goto l226
				}
				position++
				if buffer[position] != rune('q') {
					goto l226
				}
				position++
				if buffer[position] != rune('u') {
					goto l226
				}
				position++
				if buffer[position] != rune('i') {
					goto l226
				}
				position++
				if buffer[position] != rune('r') {
					goto l226
				}
				position++
				if buffer[position] != rune('e') {
					goto l226
				}
				position++
				if buffer[position] != rune('d') {
					goto l226
				}
				position++
				depth--
				add(ruleRequired, position227)
			}
			return true
		l226:
			position, tokenIndex, depth = position226, tokenIndex226, depth226
			return false
		},
		/* 58 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position228, tokenIndex228, depth228 := position, tokenIndex, depth
			{
				position229 := position
				depth++
				if buffer[position] != rune('o') {
					goto l228
				}
				position++
				if buffer[position] != rune('n') {
					goto l228
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l228
				}
				if !_rules[ruleName]() {
					goto l228
				}
				depth--
				add(ruleOn, position229)
			}
			return true
		l228:
			position, tokenIndex, depth = position228, tokenIndex228, depth228
			return false
		},
		/* 59 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position230, tokenIndex230, depth230 := position, tokenIndex, depth
			{
				position231 := position
				depth++
				if buffer[position] != rune('a') {
					goto l230
				}
				position++
				if buffer[position] != rune('u') {
					goto l230
				}
				position++
				if buffer[position] != rune('t') {
					goto l230
				}
				position++
				if buffer[position] != rune('o') {
					goto l230
				}
				position++
				depth--
				add(ruleAuto, position231)
			}
			return true
		l230:
			position, tokenIndex, depth = position230, tokenIndex230, depth230
			return false
		},
		/* 60 Mapping <- <('m' 'a' 'p' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position232, tokenIndex232, depth232 := position, tokenIndex, depth
			{
				position233 := position
				depth++
				if buffer[position] != rune('m') {
					goto l232
				}
				position++
				if buffer[position] != rune('a') {
					goto l232
				}
				position++
				if buffer[position] != rune('p') {
					goto l232
				}
				position++
				if buffer[position] != rune('[') {
					goto l232
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l232
				}
				{
					position234, tokenIndex234, depth234 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l235
					}
					goto l234
				l235:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
					if buffer[position] != rune('|') {
						goto l232
					}
					position++
					if !_rules[ruleExpression]() {
						goto l232
					}
				}
			l234:
				if buffer[position] != rune(']') {
					goto l232
				}
				position++
				depth--
				add(ruleMapping, position233)
			}
			return true
		l232:
			position, tokenIndex, depth = position232, tokenIndex232, depth232
			return false
		},
		/* 61 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position236, tokenIndex236, depth236 := position, tokenIndex, depth
			{
				position237 := position
				depth++
				if buffer[position] != rune('s') {
					goto l236
				}
				position++
				if buffer[position] != rune('u') {
					goto l236
				}
				position++
				if buffer[position] != rune('m') {
					goto l236
				}
				position++
				if buffer[position] != rune('[') {
					goto l236
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l236
				}
				if buffer[position] != rune('|') {
					goto l236
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l236
				}
				{
					position238, tokenIndex238, depth238 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l239
					}
					goto l238
				l239:
					position, tokenIndex, depth = position238, tokenIndex238, depth238
					if buffer[position] != rune('|') {
						goto l236
					}
					position++
					if !_rules[ruleExpression]() {
						goto l236
					}
				}
			l238:
				if buffer[position] != rune(']') {
					goto l236
				}
				position++
				depth--
				add(ruleSum, position237)
			}
			return true
		l236:
			position, tokenIndex, depth = position236, tokenIndex236, depth236
			return false
		},
		/* 62 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position240, tokenIndex240, depth240 := position, tokenIndex, depth
			{
				position241 := position
				depth++
				if buffer[position] != rune('l') {
					goto l240
				}
				position++
				if buffer[position] != rune('a') {
					goto l240
				}
				position++
				if buffer[position] != rune('m') {
					goto l240
				}
				position++
				if buffer[position] != rune('b') {
					goto l240
				}
				position++
				if buffer[position] != rune('d') {
					goto l240
				}
				position++
				if buffer[position] != rune('a') {
					goto l240
				}
				position++
				{
					position242, tokenIndex242, depth242 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l243
					}
					goto l242
				l243:
					position, tokenIndex, depth = position242, tokenIndex242, depth242
					if !_rules[ruleLambdaExpr]() {
						goto l240
					}
				}
			l242:
				depth--
				add(ruleLambda, position241)
			}
			return true
		l240:
			position, tokenIndex, depth = position240, tokenIndex240, depth240
			return false
		},
		/* 63 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position244, tokenIndex244, depth244 := position, tokenIndex, depth
			{
				position245 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l244
				}
				if !_rules[ruleExpression]() {
					goto l244
				}
				depth--
				add(ruleLambdaRef, position245)
			}
			return true
		l244:
			position, tokenIndex, depth = position244, tokenIndex244, depth244
			return false
		},
		/* 64 LambdaExpr <- <(ws '|' ws Name NextName* ws '|' ws ('-' '>') Expression)> */
		func() bool {
			position246, tokenIndex246, depth246 := position, tokenIndex, depth
			{
				position247 := position
				depth++
				if !_rules[rulews]() {
					goto l246
				}
				if buffer[position] != rune('|') {
					goto l246
				}
				position++
				if !_rules[rulews]() {
					goto l246
				}
				if !_rules[ruleName]() {
					goto l246
				}
			l248:
				{
					position249, tokenIndex249, depth249 := position, tokenIndex, depth
					if !_rules[ruleNextName]() {
						goto l249
					}
					goto l248
				l249:
					position, tokenIndex, depth = position249, tokenIndex249, depth249
				}
				if !_rules[rulews]() {
					goto l246
				}
				if buffer[position] != rune('|') {
					goto l246
				}
				position++
				if !_rules[rulews]() {
					goto l246
				}
				if buffer[position] != rune('-') {
					goto l246
				}
				position++
				if buffer[position] != rune('>') {
					goto l246
				}
				position++
				if !_rules[ruleExpression]() {
					goto l246
				}
				depth--
				add(ruleLambdaExpr, position247)
			}
			return true
		l246:
			position, tokenIndex, depth = position246, tokenIndex246, depth246
			return false
		},
		/* 65 NextName <- <(ws ',' ws Name)> */
		func() bool {
			position250, tokenIndex250, depth250 := position, tokenIndex, depth
			{
				position251 := position
				depth++
				if !_rules[rulews]() {
					goto l250
				}
				if buffer[position] != rune(',') {
					goto l250
				}
				position++
				if !_rules[rulews]() {
					goto l250
				}
				if !_rules[ruleName]() {
					goto l250
				}
				depth--
				add(ruleNextName, position251)
			}
			return true
		l250:
			position, tokenIndex, depth = position250, tokenIndex250, depth250
			return false
		},
		/* 66 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position252, tokenIndex252, depth252 := position, tokenIndex, depth
			{
				position253 := position
				depth++
				{
					position256, tokenIndex256, depth256 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l257
					}
					position++
					goto l256
				l257:
					position, tokenIndex, depth = position256, tokenIndex256, depth256
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l258
					}
					position++
					goto l256
				l258:
					position, tokenIndex, depth = position256, tokenIndex256, depth256
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l259
					}
					position++
					goto l256
				l259:
					position, tokenIndex, depth = position256, tokenIndex256, depth256
					if buffer[position] != rune('_') {
						goto l252
					}
					position++
				}
			l256:
			l254:
				{
					position255, tokenIndex255, depth255 := position, tokenIndex, depth
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
							goto l255
						}
						position++
					}
				l260:
					goto l254
				l255:
					position, tokenIndex, depth = position255, tokenIndex255, depth255
				}
				depth--
				add(ruleName, position253)
			}
			return true
		l252:
			position, tokenIndex, depth = position252, tokenIndex252, depth252
			return false
		},
		/* 67 Reference <- <('.'? Key FollowUpRef)> */
		func() bool {
			position264, tokenIndex264, depth264 := position, tokenIndex, depth
			{
				position265 := position
				depth++
				{
					position266, tokenIndex266, depth266 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l266
					}
					position++
					goto l267
				l266:
					position, tokenIndex, depth = position266, tokenIndex266, depth266
				}
			l267:
				if !_rules[ruleKey]() {
					goto l264
				}
				if !_rules[ruleFollowUpRef]() {
					goto l264
				}
				depth--
				add(ruleReference, position265)
			}
			return true
		l264:
			position, tokenIndex, depth = position264, tokenIndex264, depth264
			return false
		},
		/* 68 FollowUpRef <- <('.' (Key / Index))*> */
		func() bool {
			{
				position269 := position
				depth++
			l270:
				{
					position271, tokenIndex271, depth271 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l271
					}
					position++
					{
						position272, tokenIndex272, depth272 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l273
						}
						goto l272
					l273:
						position, tokenIndex, depth = position272, tokenIndex272, depth272
						if !_rules[ruleIndex]() {
							goto l271
						}
					}
				l272:
					goto l270
				l271:
					position, tokenIndex, depth = position271, tokenIndex271, depth271
				}
				depth--
				add(ruleFollowUpRef, position269)
			}
			return true
		},
		/* 69 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position274, tokenIndex274, depth274 := position, tokenIndex, depth
			{
				position275 := position
				depth++
				{
					position276, tokenIndex276, depth276 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l277
					}
					position++
					goto l276
				l277:
					position, tokenIndex, depth = position276, tokenIndex276, depth276
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l278
					}
					position++
					goto l276
				l278:
					position, tokenIndex, depth = position276, tokenIndex276, depth276
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l279
					}
					position++
					goto l276
				l279:
					position, tokenIndex, depth = position276, tokenIndex276, depth276
					if buffer[position] != rune('_') {
						goto l274
					}
					position++
				}
			l276:
			l280:
				{
					position281, tokenIndex281, depth281 := position, tokenIndex, depth
					{
						position282, tokenIndex282, depth282 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l283
						}
						position++
						goto l282
					l283:
						position, tokenIndex, depth = position282, tokenIndex282, depth282
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l284
						}
						position++
						goto l282
					l284:
						position, tokenIndex, depth = position282, tokenIndex282, depth282
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l285
						}
						position++
						goto l282
					l285:
						position, tokenIndex, depth = position282, tokenIndex282, depth282
						if buffer[position] != rune('_') {
							goto l286
						}
						position++
						goto l282
					l286:
						position, tokenIndex, depth = position282, tokenIndex282, depth282
						if buffer[position] != rune('-') {
							goto l281
						}
						position++
					}
				l282:
					goto l280
				l281:
					position, tokenIndex, depth = position281, tokenIndex281, depth281
				}
				{
					position287, tokenIndex287, depth287 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l287
					}
					position++
					{
						position289, tokenIndex289, depth289 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l290
						}
						position++
						goto l289
					l290:
						position, tokenIndex, depth = position289, tokenIndex289, depth289
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l291
						}
						position++
						goto l289
					l291:
						position, tokenIndex, depth = position289, tokenIndex289, depth289
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l292
						}
						position++
						goto l289
					l292:
						position, tokenIndex, depth = position289, tokenIndex289, depth289
						if buffer[position] != rune('_') {
							goto l287
						}
						position++
					}
				l289:
				l293:
					{
						position294, tokenIndex294, depth294 := position, tokenIndex, depth
						{
							position295, tokenIndex295, depth295 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l296
							}
							position++
							goto l295
						l296:
							position, tokenIndex, depth = position295, tokenIndex295, depth295
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l297
							}
							position++
							goto l295
						l297:
							position, tokenIndex, depth = position295, tokenIndex295, depth295
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l298
							}
							position++
							goto l295
						l298:
							position, tokenIndex, depth = position295, tokenIndex295, depth295
							if buffer[position] != rune('_') {
								goto l299
							}
							position++
							goto l295
						l299:
							position, tokenIndex, depth = position295, tokenIndex295, depth295
							if buffer[position] != rune('-') {
								goto l294
							}
							position++
						}
					l295:
						goto l293
					l294:
						position, tokenIndex, depth = position294, tokenIndex294, depth294
					}
					goto l288
				l287:
					position, tokenIndex, depth = position287, tokenIndex287, depth287
				}
			l288:
				depth--
				add(ruleKey, position275)
			}
			return true
		l274:
			position, tokenIndex, depth = position274, tokenIndex274, depth274
			return false
		},
		/* 70 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position300, tokenIndex300, depth300 := position, tokenIndex, depth
			{
				position301 := position
				depth++
				if buffer[position] != rune('[') {
					goto l300
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l300
				}
				position++
			l302:
				{
					position303, tokenIndex303, depth303 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l303
					}
					position++
					goto l302
				l303:
					position, tokenIndex, depth = position303, tokenIndex303, depth303
				}
				if buffer[position] != rune(']') {
					goto l300
				}
				position++
				depth--
				add(ruleIndex, position301)
			}
			return true
		l300:
			position, tokenIndex, depth = position300, tokenIndex300, depth300
			return false
		},
		/* 71 IP <- <([0-9]+ '.' [0-9]+ '.' [0-9]+ '.' [0-9]+)> */
		func() bool {
			position304, tokenIndex304, depth304 := position, tokenIndex, depth
			{
				position305 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l304
				}
				position++
			l306:
				{
					position307, tokenIndex307, depth307 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l307
					}
					position++
					goto l306
				l307:
					position, tokenIndex, depth = position307, tokenIndex307, depth307
				}
				if buffer[position] != rune('.') {
					goto l304
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l304
				}
				position++
			l308:
				{
					position309, tokenIndex309, depth309 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l309
					}
					position++
					goto l308
				l309:
					position, tokenIndex, depth = position309, tokenIndex309, depth309
				}
				if buffer[position] != rune('.') {
					goto l304
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l304
				}
				position++
			l310:
				{
					position311, tokenIndex311, depth311 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l311
					}
					position++
					goto l310
				l311:
					position, tokenIndex, depth = position311, tokenIndex311, depth311
				}
				if buffer[position] != rune('.') {
					goto l304
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l304
				}
				position++
			l312:
				{
					position313, tokenIndex313, depth313 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l313
					}
					position++
					goto l312
				l313:
					position, tokenIndex, depth = position313, tokenIndex313, depth313
				}
				depth--
				add(ruleIP, position305)
			}
			return true
		l304:
			position, tokenIndex, depth = position304, tokenIndex304, depth304
			return false
		},
		/* 72 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position315 := position
				depth++
			l316:
				{
					position317, tokenIndex317, depth317 := position, tokenIndex, depth
					{
						position318, tokenIndex318, depth318 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l319
						}
						position++
						goto l318
					l319:
						position, tokenIndex, depth = position318, tokenIndex318, depth318
						if buffer[position] != rune('\t') {
							goto l320
						}
						position++
						goto l318
					l320:
						position, tokenIndex, depth = position318, tokenIndex318, depth318
						if buffer[position] != rune('\n') {
							goto l321
						}
						position++
						goto l318
					l321:
						position, tokenIndex, depth = position318, tokenIndex318, depth318
						if buffer[position] != rune('\r') {
							goto l317
						}
						position++
					}
				l318:
					goto l316
				l317:
					position, tokenIndex, depth = position317, tokenIndex317, depth317
				}
				depth--
				add(rulews, position315)
			}
			return true
		},
		/* 73 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position322, tokenIndex322, depth322 := position, tokenIndex, depth
			{
				position323 := position
				depth++
				{
					position326, tokenIndex326, depth326 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l327
					}
					position++
					goto l326
				l327:
					position, tokenIndex, depth = position326, tokenIndex326, depth326
					if buffer[position] != rune('\t') {
						goto l328
					}
					position++
					goto l326
				l328:
					position, tokenIndex, depth = position326, tokenIndex326, depth326
					if buffer[position] != rune('\n') {
						goto l329
					}
					position++
					goto l326
				l329:
					position, tokenIndex, depth = position326, tokenIndex326, depth326
					if buffer[position] != rune('\r') {
						goto l322
					}
					position++
				}
			l326:
			l324:
				{
					position325, tokenIndex325, depth325 := position, tokenIndex, depth
					{
						position330, tokenIndex330, depth330 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l331
						}
						position++
						goto l330
					l331:
						position, tokenIndex, depth = position330, tokenIndex330, depth330
						if buffer[position] != rune('\t') {
							goto l332
						}
						position++
						goto l330
					l332:
						position, tokenIndex, depth = position330, tokenIndex330, depth330
						if buffer[position] != rune('\n') {
							goto l333
						}
						position++
						goto l330
					l333:
						position, tokenIndex, depth = position330, tokenIndex330, depth330
						if buffer[position] != rune('\r') {
							goto l325
						}
						position++
					}
				l330:
					goto l324
				l325:
					position, tokenIndex, depth = position325, tokenIndex325, depth325
				}
				depth--
				add(rulereq_ws, position323)
			}
			return true
		l322:
			position, tokenIndex, depth = position322, tokenIndex322, depth322
			return false
		},
		/* 75 Action0 <- <{}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
	}
	p.rules = _rules
}
