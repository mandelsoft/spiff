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
	ruleSelection
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
	"Selection",
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
	rules  [77]func() bool
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
		/* 4 Marker <- <('&' (('t' 'e' 'm' 'p' 'l' 'a' 't' 'e') / ('t' 'e' 'm' 'p' 'o' 'r' 'a' 'r' 'y') / ('l' 'o' 'c' 'a' 'l') / ('i' 'n' 'j' 'e' 'c' 't') / ('s' 't' 'a' 't' 'e')))> */
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
					if buffer[position] != rune('i') {
						goto l22
					}
					position++
					if buffer[position] != rune('n') {
						goto l22
					}
					position++
					if buffer[position] != rune('j') {
						goto l22
					}
					position++
					if buffer[position] != rune('e') {
						goto l22
					}
					position++
					if buffer[position] != rune('c') {
						goto l22
					}
					position++
					if buffer[position] != rune('t') {
						goto l22
					}
					position++
					goto l18
				l22:
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
			position23, tokenIndex23, depth23 := position, tokenIndex, depth
			{
				position24 := position
				depth++
				if !_rules[ruleGrouped]() {
					goto l23
				}
				depth--
				add(ruleMarkerExpression, position24)
			}
			return true
		l23:
			position, tokenIndex, depth = position23, tokenIndex23, depth23
			return false
		},
		/* 6 Expression <- <(ws (LambdaExpr / Level7) ws)> */
		func() bool {
			position25, tokenIndex25, depth25 := position, tokenIndex, depth
			{
				position26 := position
				depth++
				if !_rules[rulews]() {
					goto l25
				}
				{
					position27, tokenIndex27, depth27 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l28
					}
					goto l27
				l28:
					position, tokenIndex, depth = position27, tokenIndex27, depth27
					if !_rules[ruleLevel7]() {
						goto l25
					}
				}
			l27:
				if !_rules[rulews]() {
					goto l25
				}
				depth--
				add(ruleExpression, position26)
			}
			return true
		l25:
			position, tokenIndex, depth = position25, tokenIndex25, depth25
			return false
		},
		/* 7 Level7 <- <(Level6 (req_ws Or)*)> */
		func() bool {
			position29, tokenIndex29, depth29 := position, tokenIndex, depth
			{
				position30 := position
				depth++
				if !_rules[ruleLevel6]() {
					goto l29
				}
			l31:
				{
					position32, tokenIndex32, depth32 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l32
					}
					if !_rules[ruleOr]() {
						goto l32
					}
					goto l31
				l32:
					position, tokenIndex, depth = position32, tokenIndex32, depth32
				}
				depth--
				add(ruleLevel7, position30)
			}
			return true
		l29:
			position, tokenIndex, depth = position29, tokenIndex29, depth29
			return false
		},
		/* 8 Or <- <('|' '|' req_ws Level6)> */
		func() bool {
			position33, tokenIndex33, depth33 := position, tokenIndex, depth
			{
				position34 := position
				depth++
				if buffer[position] != rune('|') {
					goto l33
				}
				position++
				if buffer[position] != rune('|') {
					goto l33
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l33
				}
				if !_rules[ruleLevel6]() {
					goto l33
				}
				depth--
				add(ruleOr, position34)
			}
			return true
		l33:
			position, tokenIndex, depth = position33, tokenIndex33, depth33
			return false
		},
		/* 9 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position35, tokenIndex35, depth35 := position, tokenIndex, depth
			{
				position36 := position
				depth++
				{
					position37, tokenIndex37, depth37 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l38
					}
					goto l37
				l38:
					position, tokenIndex, depth = position37, tokenIndex37, depth37
					if !_rules[ruleLevel5]() {
						goto l35
					}
				}
			l37:
				depth--
				add(ruleLevel6, position36)
			}
			return true
		l35:
			position, tokenIndex, depth = position35, tokenIndex35, depth35
			return false
		},
		/* 10 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position39, tokenIndex39, depth39 := position, tokenIndex, depth
			{
				position40 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l39
				}
				if !_rules[rulews]() {
					goto l39
				}
				if buffer[position] != rune('?') {
					goto l39
				}
				position++
				if !_rules[ruleExpression]() {
					goto l39
				}
				if buffer[position] != rune(':') {
					goto l39
				}
				position++
				if !_rules[ruleExpression]() {
					goto l39
				}
				depth--
				add(ruleConditional, position40)
			}
			return true
		l39:
			position, tokenIndex, depth = position39, tokenIndex39, depth39
			return false
		},
		/* 11 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position41, tokenIndex41, depth41 := position, tokenIndex, depth
			{
				position42 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l41
				}
			l43:
				{
					position44, tokenIndex44, depth44 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l44
					}
					goto l43
				l44:
					position, tokenIndex, depth = position44, tokenIndex44, depth44
				}
				depth--
				add(ruleLevel5, position42)
			}
			return true
		l41:
			position, tokenIndex, depth = position41, tokenIndex41, depth41
			return false
		},
		/* 12 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position45, tokenIndex45, depth45 := position, tokenIndex, depth
			{
				position46 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l45
				}
				if !_rules[ruleLevel4]() {
					goto l45
				}
				depth--
				add(ruleConcatenation, position46)
			}
			return true
		l45:
			position, tokenIndex, depth = position45, tokenIndex45, depth45
			return false
		},
		/* 13 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position47, tokenIndex47, depth47 := position, tokenIndex, depth
			{
				position48 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l47
				}
			l49:
				{
					position50, tokenIndex50, depth50 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l50
					}
					{
						position51, tokenIndex51, depth51 := position, tokenIndex, depth
						if !_rules[ruleLogOr]() {
							goto l52
						}
						goto l51
					l52:
						position, tokenIndex, depth = position51, tokenIndex51, depth51
						if !_rules[ruleLogAnd]() {
							goto l50
						}
					}
				l51:
					goto l49
				l50:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
				}
				depth--
				add(ruleLevel4, position48)
			}
			return true
		l47:
			position, tokenIndex, depth = position47, tokenIndex47, depth47
			return false
		},
		/* 14 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position53, tokenIndex53, depth53 := position, tokenIndex, depth
			{
				position54 := position
				depth++
				if buffer[position] != rune('-') {
					goto l53
				}
				position++
				if buffer[position] != rune('o') {
					goto l53
				}
				position++
				if buffer[position] != rune('r') {
					goto l53
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l53
				}
				if !_rules[ruleLevel3]() {
					goto l53
				}
				depth--
				add(ruleLogOr, position54)
			}
			return true
		l53:
			position, tokenIndex, depth = position53, tokenIndex53, depth53
			return false
		},
		/* 15 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position55, tokenIndex55, depth55 := position, tokenIndex, depth
			{
				position56 := position
				depth++
				if buffer[position] != rune('-') {
					goto l55
				}
				position++
				if buffer[position] != rune('a') {
					goto l55
				}
				position++
				if buffer[position] != rune('n') {
					goto l55
				}
				position++
				if buffer[position] != rune('d') {
					goto l55
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l55
				}
				if !_rules[ruleLevel3]() {
					goto l55
				}
				depth--
				add(ruleLogAnd, position56)
			}
			return true
		l55:
			position, tokenIndex, depth = position55, tokenIndex55, depth55
			return false
		},
		/* 16 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position57, tokenIndex57, depth57 := position, tokenIndex, depth
			{
				position58 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l57
				}
			l59:
				{
					position60, tokenIndex60, depth60 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l60
					}
					if !_rules[ruleComparison]() {
						goto l60
					}
					goto l59
				l60:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
				}
				depth--
				add(ruleLevel3, position58)
			}
			return true
		l57:
			position, tokenIndex, depth = position57, tokenIndex57, depth57
			return false
		},
		/* 17 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position61, tokenIndex61, depth61 := position, tokenIndex, depth
			{
				position62 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l61
				}
				if !_rules[rulereq_ws]() {
					goto l61
				}
				if !_rules[ruleLevel2]() {
					goto l61
				}
				depth--
				add(ruleComparison, position62)
			}
			return true
		l61:
			position, tokenIndex, depth = position61, tokenIndex61, depth61
			return false
		},
		/* 18 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position63, tokenIndex63, depth63 := position, tokenIndex, depth
			{
				position64 := position
				depth++
				{
					position65, tokenIndex65, depth65 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l66
					}
					position++
					if buffer[position] != rune('=') {
						goto l66
					}
					position++
					goto l65
				l66:
					position, tokenIndex, depth = position65, tokenIndex65, depth65
					if buffer[position] != rune('!') {
						goto l67
					}
					position++
					if buffer[position] != rune('=') {
						goto l67
					}
					position++
					goto l65
				l67:
					position, tokenIndex, depth = position65, tokenIndex65, depth65
					if buffer[position] != rune('<') {
						goto l68
					}
					position++
					if buffer[position] != rune('=') {
						goto l68
					}
					position++
					goto l65
				l68:
					position, tokenIndex, depth = position65, tokenIndex65, depth65
					if buffer[position] != rune('>') {
						goto l69
					}
					position++
					if buffer[position] != rune('=') {
						goto l69
					}
					position++
					goto l65
				l69:
					position, tokenIndex, depth = position65, tokenIndex65, depth65
					if buffer[position] != rune('>') {
						goto l70
					}
					position++
					goto l65
				l70:
					position, tokenIndex, depth = position65, tokenIndex65, depth65
					if buffer[position] != rune('<') {
						goto l71
					}
					position++
					goto l65
				l71:
					position, tokenIndex, depth = position65, tokenIndex65, depth65
					if buffer[position] != rune('>') {
						goto l63
					}
					position++
				}
			l65:
				depth--
				add(ruleCompareOp, position64)
			}
			return true
		l63:
			position, tokenIndex, depth = position63, tokenIndex63, depth63
			return false
		},
		/* 19 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position72, tokenIndex72, depth72 := position, tokenIndex, depth
			{
				position73 := position
				depth++
				if !_rules[ruleLevel1]() {
					goto l72
				}
			l74:
				{
					position75, tokenIndex75, depth75 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l75
					}
					{
						position76, tokenIndex76, depth76 := position, tokenIndex, depth
						if !_rules[ruleAddition]() {
							goto l77
						}
						goto l76
					l77:
						position, tokenIndex, depth = position76, tokenIndex76, depth76
						if !_rules[ruleSubtraction]() {
							goto l75
						}
					}
				l76:
					goto l74
				l75:
					position, tokenIndex, depth = position75, tokenIndex75, depth75
				}
				depth--
				add(ruleLevel2, position73)
			}
			return true
		l72:
			position, tokenIndex, depth = position72, tokenIndex72, depth72
			return false
		},
		/* 20 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position78, tokenIndex78, depth78 := position, tokenIndex, depth
			{
				position79 := position
				depth++
				if buffer[position] != rune('+') {
					goto l78
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l78
				}
				if !_rules[ruleLevel1]() {
					goto l78
				}
				depth--
				add(ruleAddition, position79)
			}
			return true
		l78:
			position, tokenIndex, depth = position78, tokenIndex78, depth78
			return false
		},
		/* 21 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position80, tokenIndex80, depth80 := position, tokenIndex, depth
			{
				position81 := position
				depth++
				if buffer[position] != rune('-') {
					goto l80
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l80
				}
				if !_rules[ruleLevel1]() {
					goto l80
				}
				depth--
				add(ruleSubtraction, position81)
			}
			return true
		l80:
			position, tokenIndex, depth = position80, tokenIndex80, depth80
			return false
		},
		/* 22 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position82, tokenIndex82, depth82 := position, tokenIndex, depth
			{
				position83 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l82
				}
			l84:
				{
					position85, tokenIndex85, depth85 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l85
					}
					{
						position86, tokenIndex86, depth86 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l87
						}
						goto l86
					l87:
						position, tokenIndex, depth = position86, tokenIndex86, depth86
						if !_rules[ruleDivision]() {
							goto l88
						}
						goto l86
					l88:
						position, tokenIndex, depth = position86, tokenIndex86, depth86
						if !_rules[ruleModulo]() {
							goto l85
						}
					}
				l86:
					goto l84
				l85:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
				}
				depth--
				add(ruleLevel1, position83)
			}
			return true
		l82:
			position, tokenIndex, depth = position82, tokenIndex82, depth82
			return false
		},
		/* 23 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position89, tokenIndex89, depth89 := position, tokenIndex, depth
			{
				position90 := position
				depth++
				if buffer[position] != rune('*') {
					goto l89
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l89
				}
				if !_rules[ruleLevel0]() {
					goto l89
				}
				depth--
				add(ruleMultiplication, position90)
			}
			return true
		l89:
			position, tokenIndex, depth = position89, tokenIndex89, depth89
			return false
		},
		/* 24 Division <- <('/' req_ws Level0)> */
		func() bool {
			position91, tokenIndex91, depth91 := position, tokenIndex, depth
			{
				position92 := position
				depth++
				if buffer[position] != rune('/') {
					goto l91
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l91
				}
				if !_rules[ruleLevel0]() {
					goto l91
				}
				depth--
				add(ruleDivision, position92)
			}
			return true
		l91:
			position, tokenIndex, depth = position91, tokenIndex91, depth91
			return false
		},
		/* 25 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position93, tokenIndex93, depth93 := position, tokenIndex, depth
			{
				position94 := position
				depth++
				if buffer[position] != rune('%') {
					goto l93
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l93
				}
				if !_rules[ruleLevel0]() {
					goto l93
				}
				depth--
				add(ruleModulo, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 26 Level0 <- <(IP / String / Integer / Boolean / Undefined / Nil / Not / Substitution / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position95, tokenIndex95, depth95 := position, tokenIndex, depth
			{
				position96 := position
				depth++
				{
					position97, tokenIndex97, depth97 := position, tokenIndex, depth
					if !_rules[ruleIP]() {
						goto l98
					}
					goto l97
				l98:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
					if !_rules[ruleString]() {
						goto l99
					}
					goto l97
				l99:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
					if !_rules[ruleInteger]() {
						goto l100
					}
					goto l97
				l100:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
					if !_rules[ruleBoolean]() {
						goto l101
					}
					goto l97
				l101:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
					if !_rules[ruleUndefined]() {
						goto l102
					}
					goto l97
				l102:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
					if !_rules[ruleNil]() {
						goto l103
					}
					goto l97
				l103:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
					if !_rules[ruleNot]() {
						goto l104
					}
					goto l97
				l104:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
					if !_rules[ruleSubstitution]() {
						goto l105
					}
					goto l97
				l105:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
					if !_rules[ruleMerge]() {
						goto l106
					}
					goto l97
				l106:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
					if !_rules[ruleAuto]() {
						goto l107
					}
					goto l97
				l107:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
					if !_rules[ruleLambda]() {
						goto l108
					}
					goto l97
				l108:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
					if !_rules[ruleChained]() {
						goto l95
					}
				}
			l97:
				depth--
				add(ruleLevel0, position96)
			}
			return true
		l95:
			position, tokenIndex, depth = position95, tokenIndex95, depth95
			return false
		},
		/* 27 Chained <- <((Mapping / Selection / Sum / List / Map / Range / Grouped / Reference) ChainedQualifiedExpression*)> */
		func() bool {
			position109, tokenIndex109, depth109 := position, tokenIndex, depth
			{
				position110 := position
				depth++
				{
					position111, tokenIndex111, depth111 := position, tokenIndex, depth
					if !_rules[ruleMapping]() {
						goto l112
					}
					goto l111
				l112:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleSelection]() {
						goto l113
					}
					goto l111
				l113:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleSum]() {
						goto l114
					}
					goto l111
				l114:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleList]() {
						goto l115
					}
					goto l111
				l115:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleMap]() {
						goto l116
					}
					goto l111
				l116:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleRange]() {
						goto l117
					}
					goto l111
				l117:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleGrouped]() {
						goto l118
					}
					goto l111
				l118:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleReference]() {
						goto l109
					}
				}
			l111:
			l119:
				{
					position120, tokenIndex120, depth120 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l120
					}
					goto l119
				l120:
					position, tokenIndex, depth = position120, tokenIndex120, depth120
				}
				depth--
				add(ruleChained, position110)
			}
			return true
		l109:
			position, tokenIndex, depth = position109, tokenIndex109, depth109
			return false
		},
		/* 28 ChainedQualifiedExpression <- <(ChainedCall / ('.' (ChainedRef / ChainedDynRef / Projection)))> */
		func() bool {
			position121, tokenIndex121, depth121 := position, tokenIndex, depth
			{
				position122 := position
				depth++
				{
					position123, tokenIndex123, depth123 := position, tokenIndex, depth
					if !_rules[ruleChainedCall]() {
						goto l124
					}
					goto l123
				l124:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if buffer[position] != rune('.') {
						goto l121
					}
					position++
					{
						position125, tokenIndex125, depth125 := position, tokenIndex, depth
						if !_rules[ruleChainedRef]() {
							goto l126
						}
						goto l125
					l126:
						position, tokenIndex, depth = position125, tokenIndex125, depth125
						if !_rules[ruleChainedDynRef]() {
							goto l127
						}
						goto l125
					l127:
						position, tokenIndex, depth = position125, tokenIndex125, depth125
						if !_rules[ruleProjection]() {
							goto l121
						}
					}
				l125:
				}
			l123:
				depth--
				add(ruleChainedQualifiedExpression, position122)
			}
			return true
		l121:
			position, tokenIndex, depth = position121, tokenIndex121, depth121
			return false
		},
		/* 29 ChainedRef <- <((Key / Index) FollowUpRef)> */
		func() bool {
			position128, tokenIndex128, depth128 := position, tokenIndex, depth
			{
				position129 := position
				depth++
				{
					position130, tokenIndex130, depth130 := position, tokenIndex, depth
					if !_rules[ruleKey]() {
						goto l131
					}
					goto l130
				l131:
					position, tokenIndex, depth = position130, tokenIndex130, depth130
					if !_rules[ruleIndex]() {
						goto l128
					}
				}
			l130:
				if !_rules[ruleFollowUpRef]() {
					goto l128
				}
				depth--
				add(ruleChainedRef, position129)
			}
			return true
		l128:
			position, tokenIndex, depth = position128, tokenIndex128, depth128
			return false
		},
		/* 30 ChainedDynRef <- <('[' Expression ']')> */
		func() bool {
			position132, tokenIndex132, depth132 := position, tokenIndex, depth
			{
				position133 := position
				depth++
				if buffer[position] != rune('[') {
					goto l132
				}
				position++
				if !_rules[ruleExpression]() {
					goto l132
				}
				if buffer[position] != rune(']') {
					goto l132
				}
				position++
				depth--
				add(ruleChainedDynRef, position133)
			}
			return true
		l132:
			position, tokenIndex, depth = position132, tokenIndex132, depth132
			return false
		},
		/* 31 Slice <- <Range> */
		func() bool {
			position134, tokenIndex134, depth134 := position, tokenIndex, depth
			{
				position135 := position
				depth++
				if !_rules[ruleRange]() {
					goto l134
				}
				depth--
				add(ruleSlice, position135)
			}
			return true
		l134:
			position, tokenIndex, depth = position134, tokenIndex134, depth134
			return false
		},
		/* 32 ChainedCall <- <(StartArguments ExpressionList? ')')> */
		func() bool {
			position136, tokenIndex136, depth136 := position, tokenIndex, depth
			{
				position137 := position
				depth++
				if !_rules[ruleStartArguments]() {
					goto l136
				}
				{
					position138, tokenIndex138, depth138 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l138
					}
					goto l139
				l138:
					position, tokenIndex, depth = position138, tokenIndex138, depth138
				}
			l139:
				if buffer[position] != rune(')') {
					goto l136
				}
				position++
				depth--
				add(ruleChainedCall, position137)
			}
			return true
		l136:
			position, tokenIndex, depth = position136, tokenIndex136, depth136
			return false
		},
		/* 33 StartArguments <- <('(' ws)> */
		func() bool {
			position140, tokenIndex140, depth140 := position, tokenIndex, depth
			{
				position141 := position
				depth++
				if buffer[position] != rune('(') {
					goto l140
				}
				position++
				if !_rules[rulews]() {
					goto l140
				}
				depth--
				add(ruleStartArguments, position141)
			}
			return true
		l140:
			position, tokenIndex, depth = position140, tokenIndex140, depth140
			return false
		},
		/* 34 ExpressionList <- <(NextExpression (',' NextExpression)*)> */
		func() bool {
			position142, tokenIndex142, depth142 := position, tokenIndex, depth
			{
				position143 := position
				depth++
				if !_rules[ruleNextExpression]() {
					goto l142
				}
			l144:
				{
					position145, tokenIndex145, depth145 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l145
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l145
					}
					goto l144
				l145:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
				}
				depth--
				add(ruleExpressionList, position143)
			}
			return true
		l142:
			position, tokenIndex, depth = position142, tokenIndex142, depth142
			return false
		},
		/* 35 NextExpression <- <Expression> */
		func() bool {
			position146, tokenIndex146, depth146 := position, tokenIndex, depth
			{
				position147 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l146
				}
				depth--
				add(ruleNextExpression, position147)
			}
			return true
		l146:
			position, tokenIndex, depth = position146, tokenIndex146, depth146
			return false
		},
		/* 36 Projection <- <((('[' '*' ']') / Slice) ProjectionValue ChainedQualifiedExpression*)> */
		func() bool {
			position148, tokenIndex148, depth148 := position, tokenIndex, depth
			{
				position149 := position
				depth++
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l151
					}
					position++
					if buffer[position] != rune('*') {
						goto l151
					}
					position++
					if buffer[position] != rune(']') {
						goto l151
					}
					position++
					goto l150
				l151:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
					if !_rules[ruleSlice]() {
						goto l148
					}
				}
			l150:
				if !_rules[ruleProjectionValue]() {
					goto l148
				}
			l152:
				{
					position153, tokenIndex153, depth153 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l153
					}
					goto l152
				l153:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
				}
				depth--
				add(ruleProjection, position149)
			}
			return true
		l148:
			position, tokenIndex, depth = position148, tokenIndex148, depth148
			return false
		},
		/* 37 ProjectionValue <- <Action0> */
		func() bool {
			position154, tokenIndex154, depth154 := position, tokenIndex, depth
			{
				position155 := position
				depth++
				if !_rules[ruleAction0]() {
					goto l154
				}
				depth--
				add(ruleProjectionValue, position155)
			}
			return true
		l154:
			position, tokenIndex, depth = position154, tokenIndex154, depth154
			return false
		},
		/* 38 Substitution <- <('*' Level0)> */
		func() bool {
			position156, tokenIndex156, depth156 := position, tokenIndex, depth
			{
				position157 := position
				depth++
				if buffer[position] != rune('*') {
					goto l156
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l156
				}
				depth--
				add(ruleSubstitution, position157)
			}
			return true
		l156:
			position, tokenIndex, depth = position156, tokenIndex156, depth156
			return false
		},
		/* 39 Not <- <('!' ws Level0)> */
		func() bool {
			position158, tokenIndex158, depth158 := position, tokenIndex, depth
			{
				position159 := position
				depth++
				if buffer[position] != rune('!') {
					goto l158
				}
				position++
				if !_rules[rulews]() {
					goto l158
				}
				if !_rules[ruleLevel0]() {
					goto l158
				}
				depth--
				add(ruleNot, position159)
			}
			return true
		l158:
			position, tokenIndex, depth = position158, tokenIndex158, depth158
			return false
		},
		/* 40 Grouped <- <('(' Expression ')')> */
		func() bool {
			position160, tokenIndex160, depth160 := position, tokenIndex, depth
			{
				position161 := position
				depth++
				if buffer[position] != rune('(') {
					goto l160
				}
				position++
				if !_rules[ruleExpression]() {
					goto l160
				}
				if buffer[position] != rune(')') {
					goto l160
				}
				position++
				depth--
				add(ruleGrouped, position161)
			}
			return true
		l160:
			position, tokenIndex, depth = position160, tokenIndex160, depth160
			return false
		},
		/* 41 Range <- <('[' Expression ('.' '.') Expression ']')> */
		func() bool {
			position162, tokenIndex162, depth162 := position, tokenIndex, depth
			{
				position163 := position
				depth++
				if buffer[position] != rune('[') {
					goto l162
				}
				position++
				if !_rules[ruleExpression]() {
					goto l162
				}
				if buffer[position] != rune('.') {
					goto l162
				}
				position++
				if buffer[position] != rune('.') {
					goto l162
				}
				position++
				if !_rules[ruleExpression]() {
					goto l162
				}
				if buffer[position] != rune(']') {
					goto l162
				}
				position++
				depth--
				add(ruleRange, position163)
			}
			return true
		l162:
			position, tokenIndex, depth = position162, tokenIndex162, depth162
			return false
		},
		/* 42 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position164, tokenIndex164, depth164 := position, tokenIndex, depth
			{
				position165 := position
				depth++
				{
					position166, tokenIndex166, depth166 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l166
					}
					position++
					goto l167
				l166:
					position, tokenIndex, depth = position166, tokenIndex166, depth166
				}
			l167:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l164
				}
				position++
			l168:
				{
					position169, tokenIndex169, depth169 := position, tokenIndex, depth
					{
						position170, tokenIndex170, depth170 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l171
						}
						position++
						goto l170
					l171:
						position, tokenIndex, depth = position170, tokenIndex170, depth170
						if buffer[position] != rune('_') {
							goto l169
						}
						position++
					}
				l170:
					goto l168
				l169:
					position, tokenIndex, depth = position169, tokenIndex169, depth169
				}
				depth--
				add(ruleInteger, position165)
			}
			return true
		l164:
			position, tokenIndex, depth = position164, tokenIndex164, depth164
			return false
		},
		/* 43 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position172, tokenIndex172, depth172 := position, tokenIndex, depth
			{
				position173 := position
				depth++
				if buffer[position] != rune('"') {
					goto l172
				}
				position++
			l174:
				{
					position175, tokenIndex175, depth175 := position, tokenIndex, depth
					{
						position176, tokenIndex176, depth176 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l177
						}
						position++
						if buffer[position] != rune('"') {
							goto l177
						}
						position++
						goto l176
					l177:
						position, tokenIndex, depth = position176, tokenIndex176, depth176
						{
							position178, tokenIndex178, depth178 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l178
							}
							position++
							goto l175
						l178:
							position, tokenIndex, depth = position178, tokenIndex178, depth178
						}
						if !matchDot() {
							goto l175
						}
					}
				l176:
					goto l174
				l175:
					position, tokenIndex, depth = position175, tokenIndex175, depth175
				}
				if buffer[position] != rune('"') {
					goto l172
				}
				position++
				depth--
				add(ruleString, position173)
			}
			return true
		l172:
			position, tokenIndex, depth = position172, tokenIndex172, depth172
			return false
		},
		/* 44 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position179, tokenIndex179, depth179 := position, tokenIndex, depth
			{
				position180 := position
				depth++
				{
					position181, tokenIndex181, depth181 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l182
					}
					position++
					if buffer[position] != rune('r') {
						goto l182
					}
					position++
					if buffer[position] != rune('u') {
						goto l182
					}
					position++
					if buffer[position] != rune('e') {
						goto l182
					}
					position++
					goto l181
				l182:
					position, tokenIndex, depth = position181, tokenIndex181, depth181
					if buffer[position] != rune('f') {
						goto l179
					}
					position++
					if buffer[position] != rune('a') {
						goto l179
					}
					position++
					if buffer[position] != rune('l') {
						goto l179
					}
					position++
					if buffer[position] != rune('s') {
						goto l179
					}
					position++
					if buffer[position] != rune('e') {
						goto l179
					}
					position++
				}
			l181:
				depth--
				add(ruleBoolean, position180)
			}
			return true
		l179:
			position, tokenIndex, depth = position179, tokenIndex179, depth179
			return false
		},
		/* 45 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position183, tokenIndex183, depth183 := position, tokenIndex, depth
			{
				position184 := position
				depth++
				{
					position185, tokenIndex185, depth185 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l186
					}
					position++
					if buffer[position] != rune('i') {
						goto l186
					}
					position++
					if buffer[position] != rune('l') {
						goto l186
					}
					position++
					goto l185
				l186:
					position, tokenIndex, depth = position185, tokenIndex185, depth185
					if buffer[position] != rune('~') {
						goto l183
					}
					position++
				}
			l185:
				depth--
				add(ruleNil, position184)
			}
			return true
		l183:
			position, tokenIndex, depth = position183, tokenIndex183, depth183
			return false
		},
		/* 46 Undefined <- <('~' '~')> */
		func() bool {
			position187, tokenIndex187, depth187 := position, tokenIndex, depth
			{
				position188 := position
				depth++
				if buffer[position] != rune('~') {
					goto l187
				}
				position++
				if buffer[position] != rune('~') {
					goto l187
				}
				position++
				depth--
				add(ruleUndefined, position188)
			}
			return true
		l187:
			position, tokenIndex, depth = position187, tokenIndex187, depth187
			return false
		},
		/* 47 List <- <(StartList ExpressionList? ']')> */
		func() bool {
			position189, tokenIndex189, depth189 := position, tokenIndex, depth
			{
				position190 := position
				depth++
				if !_rules[ruleStartList]() {
					goto l189
				}
				{
					position191, tokenIndex191, depth191 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l191
					}
					goto l192
				l191:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
				}
			l192:
				if buffer[position] != rune(']') {
					goto l189
				}
				position++
				depth--
				add(ruleList, position190)
			}
			return true
		l189:
			position, tokenIndex, depth = position189, tokenIndex189, depth189
			return false
		},
		/* 48 StartList <- <'['> */
		func() bool {
			position193, tokenIndex193, depth193 := position, tokenIndex, depth
			{
				position194 := position
				depth++
				if buffer[position] != rune('[') {
					goto l193
				}
				position++
				depth--
				add(ruleStartList, position194)
			}
			return true
		l193:
			position, tokenIndex, depth = position193, tokenIndex193, depth193
			return false
		},
		/* 49 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position195, tokenIndex195, depth195 := position, tokenIndex, depth
			{
				position196 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l195
				}
				if !_rules[rulews]() {
					goto l195
				}
				{
					position197, tokenIndex197, depth197 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l197
					}
					goto l198
				l197:
					position, tokenIndex, depth = position197, tokenIndex197, depth197
				}
			l198:
				if buffer[position] != rune('}') {
					goto l195
				}
				position++
				depth--
				add(ruleMap, position196)
			}
			return true
		l195:
			position, tokenIndex, depth = position195, tokenIndex195, depth195
			return false
		},
		/* 50 CreateMap <- <'{'> */
		func() bool {
			position199, tokenIndex199, depth199 := position, tokenIndex, depth
			{
				position200 := position
				depth++
				if buffer[position] != rune('{') {
					goto l199
				}
				position++
				depth--
				add(ruleCreateMap, position200)
			}
			return true
		l199:
			position, tokenIndex, depth = position199, tokenIndex199, depth199
			return false
		},
		/* 51 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position201, tokenIndex201, depth201 := position, tokenIndex, depth
			{
				position202 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l201
				}
			l203:
				{
					position204, tokenIndex204, depth204 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l204
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l204
					}
					goto l203
				l204:
					position, tokenIndex, depth = position204, tokenIndex204, depth204
				}
				depth--
				add(ruleAssignments, position202)
			}
			return true
		l201:
			position, tokenIndex, depth = position201, tokenIndex201, depth201
			return false
		},
		/* 52 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position205, tokenIndex205, depth205 := position, tokenIndex, depth
			{
				position206 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l205
				}
				if buffer[position] != rune('=') {
					goto l205
				}
				position++
				if !_rules[ruleExpression]() {
					goto l205
				}
				depth--
				add(ruleAssignment, position206)
			}
			return true
		l205:
			position, tokenIndex, depth = position205, tokenIndex205, depth205
			return false
		},
		/* 53 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position207, tokenIndex207, depth207 := position, tokenIndex, depth
			{
				position208 := position
				depth++
				{
					position209, tokenIndex209, depth209 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l210
					}
					goto l209
				l210:
					position, tokenIndex, depth = position209, tokenIndex209, depth209
					if !_rules[ruleSimpleMerge]() {
						goto l207
					}
				}
			l209:
				depth--
				add(ruleMerge, position208)
			}
			return true
		l207:
			position, tokenIndex, depth = position207, tokenIndex207, depth207
			return false
		},
		/* 54 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position211, tokenIndex211, depth211 := position, tokenIndex, depth
			{
				position212 := position
				depth++
				if buffer[position] != rune('m') {
					goto l211
				}
				position++
				if buffer[position] != rune('e') {
					goto l211
				}
				position++
				if buffer[position] != rune('r') {
					goto l211
				}
				position++
				if buffer[position] != rune('g') {
					goto l211
				}
				position++
				if buffer[position] != rune('e') {
					goto l211
				}
				position++
				{
					position213, tokenIndex213, depth213 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l213
					}
					if !_rules[ruleRequired]() {
						goto l213
					}
					goto l211
				l213:
					position, tokenIndex, depth = position213, tokenIndex213, depth213
				}
				{
					position214, tokenIndex214, depth214 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l214
					}
					{
						position216, tokenIndex216, depth216 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l217
						}
						goto l216
					l217:
						position, tokenIndex, depth = position216, tokenIndex216, depth216
						if !_rules[ruleOn]() {
							goto l214
						}
					}
				l216:
					goto l215
				l214:
					position, tokenIndex, depth = position214, tokenIndex214, depth214
				}
			l215:
				if !_rules[rulereq_ws]() {
					goto l211
				}
				if !_rules[ruleReference]() {
					goto l211
				}
				depth--
				add(ruleRefMerge, position212)
			}
			return true
		l211:
			position, tokenIndex, depth = position211, tokenIndex211, depth211
			return false
		},
		/* 55 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' !'(' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position218, tokenIndex218, depth218 := position, tokenIndex, depth
			{
				position219 := position
				depth++
				if buffer[position] != rune('m') {
					goto l218
				}
				position++
				if buffer[position] != rune('e') {
					goto l218
				}
				position++
				if buffer[position] != rune('r') {
					goto l218
				}
				position++
				if buffer[position] != rune('g') {
					goto l218
				}
				position++
				if buffer[position] != rune('e') {
					goto l218
				}
				position++
				{
					position220, tokenIndex220, depth220 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l220
					}
					position++
					goto l218
				l220:
					position, tokenIndex, depth = position220, tokenIndex220, depth220
				}
				{
					position221, tokenIndex221, depth221 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l221
					}
					{
						position223, tokenIndex223, depth223 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l224
						}
						goto l223
					l224:
						position, tokenIndex, depth = position223, tokenIndex223, depth223
						if !_rules[ruleRequired]() {
							goto l225
						}
						goto l223
					l225:
						position, tokenIndex, depth = position223, tokenIndex223, depth223
						if !_rules[ruleOn]() {
							goto l221
						}
					}
				l223:
					goto l222
				l221:
					position, tokenIndex, depth = position221, tokenIndex221, depth221
				}
			l222:
				depth--
				add(ruleSimpleMerge, position219)
			}
			return true
		l218:
			position, tokenIndex, depth = position218, tokenIndex218, depth218
			return false
		},
		/* 56 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
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
				if buffer[position] != rune('p') {
					goto l226
				}
				position++
				if buffer[position] != rune('l') {
					goto l226
				}
				position++
				if buffer[position] != rune('a') {
					goto l226
				}
				position++
				if buffer[position] != rune('c') {
					goto l226
				}
				position++
				if buffer[position] != rune('e') {
					goto l226
				}
				position++
				depth--
				add(ruleReplace, position227)
			}
			return true
		l226:
			position, tokenIndex, depth = position226, tokenIndex226, depth226
			return false
		},
		/* 57 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position228, tokenIndex228, depth228 := position, tokenIndex, depth
			{
				position229 := position
				depth++
				if buffer[position] != rune('r') {
					goto l228
				}
				position++
				if buffer[position] != rune('e') {
					goto l228
				}
				position++
				if buffer[position] != rune('q') {
					goto l228
				}
				position++
				if buffer[position] != rune('u') {
					goto l228
				}
				position++
				if buffer[position] != rune('i') {
					goto l228
				}
				position++
				if buffer[position] != rune('r') {
					goto l228
				}
				position++
				if buffer[position] != rune('e') {
					goto l228
				}
				position++
				if buffer[position] != rune('d') {
					goto l228
				}
				position++
				depth--
				add(ruleRequired, position229)
			}
			return true
		l228:
			position, tokenIndex, depth = position228, tokenIndex228, depth228
			return false
		},
		/* 58 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position230, tokenIndex230, depth230 := position, tokenIndex, depth
			{
				position231 := position
				depth++
				if buffer[position] != rune('o') {
					goto l230
				}
				position++
				if buffer[position] != rune('n') {
					goto l230
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l230
				}
				if !_rules[ruleName]() {
					goto l230
				}
				depth--
				add(ruleOn, position231)
			}
			return true
		l230:
			position, tokenIndex, depth = position230, tokenIndex230, depth230
			return false
		},
		/* 59 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position232, tokenIndex232, depth232 := position, tokenIndex, depth
			{
				position233 := position
				depth++
				if buffer[position] != rune('a') {
					goto l232
				}
				position++
				if buffer[position] != rune('u') {
					goto l232
				}
				position++
				if buffer[position] != rune('t') {
					goto l232
				}
				position++
				if buffer[position] != rune('o') {
					goto l232
				}
				position++
				depth--
				add(ruleAuto, position233)
			}
			return true
		l232:
			position, tokenIndex, depth = position232, tokenIndex232, depth232
			return false
		},
		/* 60 Mapping <- <('m' 'a' 'p' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position234, tokenIndex234, depth234 := position, tokenIndex, depth
			{
				position235 := position
				depth++
				if buffer[position] != rune('m') {
					goto l234
				}
				position++
				if buffer[position] != rune('a') {
					goto l234
				}
				position++
				if buffer[position] != rune('p') {
					goto l234
				}
				position++
				if buffer[position] != rune('[') {
					goto l234
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l234
				}
				{
					position236, tokenIndex236, depth236 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l237
					}
					goto l236
				l237:
					position, tokenIndex, depth = position236, tokenIndex236, depth236
					if buffer[position] != rune('|') {
						goto l234
					}
					position++
					if !_rules[ruleExpression]() {
						goto l234
					}
				}
			l236:
				if buffer[position] != rune(']') {
					goto l234
				}
				position++
				depth--
				add(ruleMapping, position235)
			}
			return true
		l234:
			position, tokenIndex, depth = position234, tokenIndex234, depth234
			return false
		},
		/* 61 Selection <- <('s' 'e' 'l' 'e' 'c' 't' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position238, tokenIndex238, depth238 := position, tokenIndex, depth
			{
				position239 := position
				depth++
				if buffer[position] != rune('s') {
					goto l238
				}
				position++
				if buffer[position] != rune('e') {
					goto l238
				}
				position++
				if buffer[position] != rune('l') {
					goto l238
				}
				position++
				if buffer[position] != rune('e') {
					goto l238
				}
				position++
				if buffer[position] != rune('c') {
					goto l238
				}
				position++
				if buffer[position] != rune('t') {
					goto l238
				}
				position++
				if buffer[position] != rune('[') {
					goto l238
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l238
				}
				{
					position240, tokenIndex240, depth240 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l241
					}
					goto l240
				l241:
					position, tokenIndex, depth = position240, tokenIndex240, depth240
					if buffer[position] != rune('|') {
						goto l238
					}
					position++
					if !_rules[ruleExpression]() {
						goto l238
					}
				}
			l240:
				if buffer[position] != rune(']') {
					goto l238
				}
				position++
				depth--
				add(ruleSelection, position239)
			}
			return true
		l238:
			position, tokenIndex, depth = position238, tokenIndex238, depth238
			return false
		},
		/* 62 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position242, tokenIndex242, depth242 := position, tokenIndex, depth
			{
				position243 := position
				depth++
				if buffer[position] != rune('s') {
					goto l242
				}
				position++
				if buffer[position] != rune('u') {
					goto l242
				}
				position++
				if buffer[position] != rune('m') {
					goto l242
				}
				position++
				if buffer[position] != rune('[') {
					goto l242
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l242
				}
				if buffer[position] != rune('|') {
					goto l242
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l242
				}
				{
					position244, tokenIndex244, depth244 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l245
					}
					goto l244
				l245:
					position, tokenIndex, depth = position244, tokenIndex244, depth244
					if buffer[position] != rune('|') {
						goto l242
					}
					position++
					if !_rules[ruleExpression]() {
						goto l242
					}
				}
			l244:
				if buffer[position] != rune(']') {
					goto l242
				}
				position++
				depth--
				add(ruleSum, position243)
			}
			return true
		l242:
			position, tokenIndex, depth = position242, tokenIndex242, depth242
			return false
		},
		/* 63 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position246, tokenIndex246, depth246 := position, tokenIndex, depth
			{
				position247 := position
				depth++
				if buffer[position] != rune('l') {
					goto l246
				}
				position++
				if buffer[position] != rune('a') {
					goto l246
				}
				position++
				if buffer[position] != rune('m') {
					goto l246
				}
				position++
				if buffer[position] != rune('b') {
					goto l246
				}
				position++
				if buffer[position] != rune('d') {
					goto l246
				}
				position++
				if buffer[position] != rune('a') {
					goto l246
				}
				position++
				{
					position248, tokenIndex248, depth248 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l249
					}
					goto l248
				l249:
					position, tokenIndex, depth = position248, tokenIndex248, depth248
					if !_rules[ruleLambdaExpr]() {
						goto l246
					}
				}
			l248:
				depth--
				add(ruleLambda, position247)
			}
			return true
		l246:
			position, tokenIndex, depth = position246, tokenIndex246, depth246
			return false
		},
		/* 64 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position250, tokenIndex250, depth250 := position, tokenIndex, depth
			{
				position251 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l250
				}
				if !_rules[ruleExpression]() {
					goto l250
				}
				depth--
				add(ruleLambdaRef, position251)
			}
			return true
		l250:
			position, tokenIndex, depth = position250, tokenIndex250, depth250
			return false
		},
		/* 65 LambdaExpr <- <(ws '|' ws Name NextName* ws '|' ws ('-' '>') Expression)> */
		func() bool {
			position252, tokenIndex252, depth252 := position, tokenIndex, depth
			{
				position253 := position
				depth++
				if !_rules[rulews]() {
					goto l252
				}
				if buffer[position] != rune('|') {
					goto l252
				}
				position++
				if !_rules[rulews]() {
					goto l252
				}
				if !_rules[ruleName]() {
					goto l252
				}
			l254:
				{
					position255, tokenIndex255, depth255 := position, tokenIndex, depth
					if !_rules[ruleNextName]() {
						goto l255
					}
					goto l254
				l255:
					position, tokenIndex, depth = position255, tokenIndex255, depth255
				}
				if !_rules[rulews]() {
					goto l252
				}
				if buffer[position] != rune('|') {
					goto l252
				}
				position++
				if !_rules[rulews]() {
					goto l252
				}
				if buffer[position] != rune('-') {
					goto l252
				}
				position++
				if buffer[position] != rune('>') {
					goto l252
				}
				position++
				if !_rules[ruleExpression]() {
					goto l252
				}
				depth--
				add(ruleLambdaExpr, position253)
			}
			return true
		l252:
			position, tokenIndex, depth = position252, tokenIndex252, depth252
			return false
		},
		/* 66 NextName <- <(ws ',' ws Name)> */
		func() bool {
			position256, tokenIndex256, depth256 := position, tokenIndex, depth
			{
				position257 := position
				depth++
				if !_rules[rulews]() {
					goto l256
				}
				if buffer[position] != rune(',') {
					goto l256
				}
				position++
				if !_rules[rulews]() {
					goto l256
				}
				if !_rules[ruleName]() {
					goto l256
				}
				depth--
				add(ruleNextName, position257)
			}
			return true
		l256:
			position, tokenIndex, depth = position256, tokenIndex256, depth256
			return false
		},
		/* 67 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position258, tokenIndex258, depth258 := position, tokenIndex, depth
			{
				position259 := position
				depth++
				{
					position262, tokenIndex262, depth262 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l263
					}
					position++
					goto l262
				l263:
					position, tokenIndex, depth = position262, tokenIndex262, depth262
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l264
					}
					position++
					goto l262
				l264:
					position, tokenIndex, depth = position262, tokenIndex262, depth262
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l265
					}
					position++
					goto l262
				l265:
					position, tokenIndex, depth = position262, tokenIndex262, depth262
					if buffer[position] != rune('_') {
						goto l258
					}
					position++
				}
			l262:
			l260:
				{
					position261, tokenIndex261, depth261 := position, tokenIndex, depth
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
							goto l261
						}
						position++
					}
				l266:
					goto l260
				l261:
					position, tokenIndex, depth = position261, tokenIndex261, depth261
				}
				depth--
				add(ruleName, position259)
			}
			return true
		l258:
			position, tokenIndex, depth = position258, tokenIndex258, depth258
			return false
		},
		/* 68 Reference <- <('.'? Key FollowUpRef)> */
		func() bool {
			position270, tokenIndex270, depth270 := position, tokenIndex, depth
			{
				position271 := position
				depth++
				{
					position272, tokenIndex272, depth272 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l272
					}
					position++
					goto l273
				l272:
					position, tokenIndex, depth = position272, tokenIndex272, depth272
				}
			l273:
				if !_rules[ruleKey]() {
					goto l270
				}
				if !_rules[ruleFollowUpRef]() {
					goto l270
				}
				depth--
				add(ruleReference, position271)
			}
			return true
		l270:
			position, tokenIndex, depth = position270, tokenIndex270, depth270
			return false
		},
		/* 69 FollowUpRef <- <('.' (Key / Index))*> */
		func() bool {
			{
				position275 := position
				depth++
			l276:
				{
					position277, tokenIndex277, depth277 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l277
					}
					position++
					{
						position278, tokenIndex278, depth278 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l279
						}
						goto l278
					l279:
						position, tokenIndex, depth = position278, tokenIndex278, depth278
						if !_rules[ruleIndex]() {
							goto l277
						}
					}
				l278:
					goto l276
				l277:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
				}
				depth--
				add(ruleFollowUpRef, position275)
			}
			return true
		},
		/* 70 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position280, tokenIndex280, depth280 := position, tokenIndex, depth
			{
				position281 := position
				depth++
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
						goto l280
					}
					position++
				}
			l282:
			l286:
				{
					position287, tokenIndex287, depth287 := position, tokenIndex, depth
					{
						position288, tokenIndex288, depth288 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l289
						}
						position++
						goto l288
					l289:
						position, tokenIndex, depth = position288, tokenIndex288, depth288
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l290
						}
						position++
						goto l288
					l290:
						position, tokenIndex, depth = position288, tokenIndex288, depth288
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l291
						}
						position++
						goto l288
					l291:
						position, tokenIndex, depth = position288, tokenIndex288, depth288
						if buffer[position] != rune('_') {
							goto l292
						}
						position++
						goto l288
					l292:
						position, tokenIndex, depth = position288, tokenIndex288, depth288
						if buffer[position] != rune('-') {
							goto l287
						}
						position++
					}
				l288:
					goto l286
				l287:
					position, tokenIndex, depth = position287, tokenIndex287, depth287
				}
				{
					position293, tokenIndex293, depth293 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l293
					}
					position++
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
							goto l293
						}
						position++
					}
				l295:
				l299:
					{
						position300, tokenIndex300, depth300 := position, tokenIndex, depth
						{
							position301, tokenIndex301, depth301 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l302
							}
							position++
							goto l301
						l302:
							position, tokenIndex, depth = position301, tokenIndex301, depth301
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l303
							}
							position++
							goto l301
						l303:
							position, tokenIndex, depth = position301, tokenIndex301, depth301
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l304
							}
							position++
							goto l301
						l304:
							position, tokenIndex, depth = position301, tokenIndex301, depth301
							if buffer[position] != rune('_') {
								goto l305
							}
							position++
							goto l301
						l305:
							position, tokenIndex, depth = position301, tokenIndex301, depth301
							if buffer[position] != rune('-') {
								goto l300
							}
							position++
						}
					l301:
						goto l299
					l300:
						position, tokenIndex, depth = position300, tokenIndex300, depth300
					}
					goto l294
				l293:
					position, tokenIndex, depth = position293, tokenIndex293, depth293
				}
			l294:
				depth--
				add(ruleKey, position281)
			}
			return true
		l280:
			position, tokenIndex, depth = position280, tokenIndex280, depth280
			return false
		},
		/* 71 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position306, tokenIndex306, depth306 := position, tokenIndex, depth
			{
				position307 := position
				depth++
				if buffer[position] != rune('[') {
					goto l306
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l306
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
				if buffer[position] != rune(']') {
					goto l306
				}
				position++
				depth--
				add(ruleIndex, position307)
			}
			return true
		l306:
			position, tokenIndex, depth = position306, tokenIndex306, depth306
			return false
		},
		/* 72 IP <- <([0-9]+ '.' [0-9]+ '.' [0-9]+ '.' [0-9]+)> */
		func() bool {
			position310, tokenIndex310, depth310 := position, tokenIndex, depth
			{
				position311 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l310
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
				if buffer[position] != rune('.') {
					goto l310
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l310
				}
				position++
			l314:
				{
					position315, tokenIndex315, depth315 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l315
					}
					position++
					goto l314
				l315:
					position, tokenIndex, depth = position315, tokenIndex315, depth315
				}
				if buffer[position] != rune('.') {
					goto l310
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l310
				}
				position++
			l316:
				{
					position317, tokenIndex317, depth317 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l317
					}
					position++
					goto l316
				l317:
					position, tokenIndex, depth = position317, tokenIndex317, depth317
				}
				if buffer[position] != rune('.') {
					goto l310
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l310
				}
				position++
			l318:
				{
					position319, tokenIndex319, depth319 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l319
					}
					position++
					goto l318
				l319:
					position, tokenIndex, depth = position319, tokenIndex319, depth319
				}
				depth--
				add(ruleIP, position311)
			}
			return true
		l310:
			position, tokenIndex, depth = position310, tokenIndex310, depth310
			return false
		},
		/* 73 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position321 := position
				depth++
			l322:
				{
					position323, tokenIndex323, depth323 := position, tokenIndex, depth
					{
						position324, tokenIndex324, depth324 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l325
						}
						position++
						goto l324
					l325:
						position, tokenIndex, depth = position324, tokenIndex324, depth324
						if buffer[position] != rune('\t') {
							goto l326
						}
						position++
						goto l324
					l326:
						position, tokenIndex, depth = position324, tokenIndex324, depth324
						if buffer[position] != rune('\n') {
							goto l327
						}
						position++
						goto l324
					l327:
						position, tokenIndex, depth = position324, tokenIndex324, depth324
						if buffer[position] != rune('\r') {
							goto l323
						}
						position++
					}
				l324:
					goto l322
				l323:
					position, tokenIndex, depth = position323, tokenIndex323, depth323
				}
				depth--
				add(rulews, position321)
			}
			return true
		},
		/* 74 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position328, tokenIndex328, depth328 := position, tokenIndex, depth
			{
				position329 := position
				depth++
				{
					position332, tokenIndex332, depth332 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l333
					}
					position++
					goto l332
				l333:
					position, tokenIndex, depth = position332, tokenIndex332, depth332
					if buffer[position] != rune('\t') {
						goto l334
					}
					position++
					goto l332
				l334:
					position, tokenIndex, depth = position332, tokenIndex332, depth332
					if buffer[position] != rune('\n') {
						goto l335
					}
					position++
					goto l332
				l335:
					position, tokenIndex, depth = position332, tokenIndex332, depth332
					if buffer[position] != rune('\r') {
						goto l328
					}
					position++
				}
			l332:
			l330:
				{
					position331, tokenIndex331, depth331 := position, tokenIndex, depth
					{
						position336, tokenIndex336, depth336 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l337
						}
						position++
						goto l336
					l337:
						position, tokenIndex, depth = position336, tokenIndex336, depth336
						if buffer[position] != rune('\t') {
							goto l338
						}
						position++
						goto l336
					l338:
						position, tokenIndex, depth = position336, tokenIndex336, depth336
						if buffer[position] != rune('\n') {
							goto l339
						}
						position++
						goto l336
					l339:
						position, tokenIndex, depth = position336, tokenIndex336, depth336
						if buffer[position] != rune('\r') {
							goto l331
						}
						position++
					}
				l336:
					goto l330
				l331:
					position, tokenIndex, depth = position331, tokenIndex331, depth331
				}
				depth--
				add(rulereq_ws, position329)
			}
			return true
		l328:
			position, tokenIndex, depth = position328, tokenIndex328, depth328
			return false
		},
		/* 76 Action0 <- <{}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
	}
	p.rules = _rules
}
