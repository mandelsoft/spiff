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
		/* 27 Chained <- <((Mapping / Sum / List / Map / Range / Grouped / Reference) ChainedQualifiedExpression*)> */
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
					if !_rules[ruleSum]() {
						goto l113
					}
					goto l111
				l113:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleList]() {
						goto l114
					}
					goto l111
				l114:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleMap]() {
						goto l115
					}
					goto l111
				l115:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleRange]() {
						goto l116
					}
					goto l111
				l116:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleGrouped]() {
						goto l117
					}
					goto l111
				l117:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleReference]() {
						goto l109
					}
				}
			l111:
			l118:
				{
					position119, tokenIndex119, depth119 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l119
					}
					goto l118
				l119:
					position, tokenIndex, depth = position119, tokenIndex119, depth119
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
			position120, tokenIndex120, depth120 := position, tokenIndex, depth
			{
				position121 := position
				depth++
				{
					position122, tokenIndex122, depth122 := position, tokenIndex, depth
					if !_rules[ruleChainedCall]() {
						goto l123
					}
					goto l122
				l123:
					position, tokenIndex, depth = position122, tokenIndex122, depth122
					if buffer[position] != rune('.') {
						goto l120
					}
					position++
					{
						position124, tokenIndex124, depth124 := position, tokenIndex, depth
						if !_rules[ruleChainedRef]() {
							goto l125
						}
						goto l124
					l125:
						position, tokenIndex, depth = position124, tokenIndex124, depth124
						if !_rules[ruleChainedDynRef]() {
							goto l126
						}
						goto l124
					l126:
						position, tokenIndex, depth = position124, tokenIndex124, depth124
						if !_rules[ruleProjection]() {
							goto l120
						}
					}
				l124:
				}
			l122:
				depth--
				add(ruleChainedQualifiedExpression, position121)
			}
			return true
		l120:
			position, tokenIndex, depth = position120, tokenIndex120, depth120
			return false
		},
		/* 29 ChainedRef <- <((Key / Index) FollowUpRef)> */
		func() bool {
			position127, tokenIndex127, depth127 := position, tokenIndex, depth
			{
				position128 := position
				depth++
				{
					position129, tokenIndex129, depth129 := position, tokenIndex, depth
					if !_rules[ruleKey]() {
						goto l130
					}
					goto l129
				l130:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
					if !_rules[ruleIndex]() {
						goto l127
					}
				}
			l129:
				if !_rules[ruleFollowUpRef]() {
					goto l127
				}
				depth--
				add(ruleChainedRef, position128)
			}
			return true
		l127:
			position, tokenIndex, depth = position127, tokenIndex127, depth127
			return false
		},
		/* 30 ChainedDynRef <- <('[' Expression ']')> */
		func() bool {
			position131, tokenIndex131, depth131 := position, tokenIndex, depth
			{
				position132 := position
				depth++
				if buffer[position] != rune('[') {
					goto l131
				}
				position++
				if !_rules[ruleExpression]() {
					goto l131
				}
				if buffer[position] != rune(']') {
					goto l131
				}
				position++
				depth--
				add(ruleChainedDynRef, position132)
			}
			return true
		l131:
			position, tokenIndex, depth = position131, tokenIndex131, depth131
			return false
		},
		/* 31 Slice <- <Range> */
		func() bool {
			position133, tokenIndex133, depth133 := position, tokenIndex, depth
			{
				position134 := position
				depth++
				if !_rules[ruleRange]() {
					goto l133
				}
				depth--
				add(ruleSlice, position134)
			}
			return true
		l133:
			position, tokenIndex, depth = position133, tokenIndex133, depth133
			return false
		},
		/* 32 ChainedCall <- <(StartArguments ExpressionList? ')')> */
		func() bool {
			position135, tokenIndex135, depth135 := position, tokenIndex, depth
			{
				position136 := position
				depth++
				if !_rules[ruleStartArguments]() {
					goto l135
				}
				{
					position137, tokenIndex137, depth137 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l137
					}
					goto l138
				l137:
					position, tokenIndex, depth = position137, tokenIndex137, depth137
				}
			l138:
				if buffer[position] != rune(')') {
					goto l135
				}
				position++
				depth--
				add(ruleChainedCall, position136)
			}
			return true
		l135:
			position, tokenIndex, depth = position135, tokenIndex135, depth135
			return false
		},
		/* 33 StartArguments <- <('(' ws)> */
		func() bool {
			position139, tokenIndex139, depth139 := position, tokenIndex, depth
			{
				position140 := position
				depth++
				if buffer[position] != rune('(') {
					goto l139
				}
				position++
				if !_rules[rulews]() {
					goto l139
				}
				depth--
				add(ruleStartArguments, position140)
			}
			return true
		l139:
			position, tokenIndex, depth = position139, tokenIndex139, depth139
			return false
		},
		/* 34 ExpressionList <- <(NextExpression (',' NextExpression)*)> */
		func() bool {
			position141, tokenIndex141, depth141 := position, tokenIndex, depth
			{
				position142 := position
				depth++
				if !_rules[ruleNextExpression]() {
					goto l141
				}
			l143:
				{
					position144, tokenIndex144, depth144 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l144
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l144
					}
					goto l143
				l144:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
				}
				depth--
				add(ruleExpressionList, position142)
			}
			return true
		l141:
			position, tokenIndex, depth = position141, tokenIndex141, depth141
			return false
		},
		/* 35 NextExpression <- <Expression> */
		func() bool {
			position145, tokenIndex145, depth145 := position, tokenIndex, depth
			{
				position146 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l145
				}
				depth--
				add(ruleNextExpression, position146)
			}
			return true
		l145:
			position, tokenIndex, depth = position145, tokenIndex145, depth145
			return false
		},
		/* 36 Projection <- <((('[' '*' ']') / Slice) ProjectionValue ChainedQualifiedExpression*)> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				{
					position149, tokenIndex149, depth149 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l150
					}
					position++
					if buffer[position] != rune('*') {
						goto l150
					}
					position++
					if buffer[position] != rune(']') {
						goto l150
					}
					position++
					goto l149
				l150:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
					if !_rules[ruleSlice]() {
						goto l147
					}
				}
			l149:
				if !_rules[ruleProjectionValue]() {
					goto l147
				}
			l151:
				{
					position152, tokenIndex152, depth152 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l152
					}
					goto l151
				l152:
					position, tokenIndex, depth = position152, tokenIndex152, depth152
				}
				depth--
				add(ruleProjection, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 37 ProjectionValue <- <Action0> */
		func() bool {
			position153, tokenIndex153, depth153 := position, tokenIndex, depth
			{
				position154 := position
				depth++
				if !_rules[ruleAction0]() {
					goto l153
				}
				depth--
				add(ruleProjectionValue, position154)
			}
			return true
		l153:
			position, tokenIndex, depth = position153, tokenIndex153, depth153
			return false
		},
		/* 38 Substitution <- <('*' Level0)> */
		func() bool {
			position155, tokenIndex155, depth155 := position, tokenIndex, depth
			{
				position156 := position
				depth++
				if buffer[position] != rune('*') {
					goto l155
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l155
				}
				depth--
				add(ruleSubstitution, position156)
			}
			return true
		l155:
			position, tokenIndex, depth = position155, tokenIndex155, depth155
			return false
		},
		/* 39 Not <- <('!' ws Level0)> */
		func() bool {
			position157, tokenIndex157, depth157 := position, tokenIndex, depth
			{
				position158 := position
				depth++
				if buffer[position] != rune('!') {
					goto l157
				}
				position++
				if !_rules[rulews]() {
					goto l157
				}
				if !_rules[ruleLevel0]() {
					goto l157
				}
				depth--
				add(ruleNot, position158)
			}
			return true
		l157:
			position, tokenIndex, depth = position157, tokenIndex157, depth157
			return false
		},
		/* 40 Grouped <- <('(' Expression ')')> */
		func() bool {
			position159, tokenIndex159, depth159 := position, tokenIndex, depth
			{
				position160 := position
				depth++
				if buffer[position] != rune('(') {
					goto l159
				}
				position++
				if !_rules[ruleExpression]() {
					goto l159
				}
				if buffer[position] != rune(')') {
					goto l159
				}
				position++
				depth--
				add(ruleGrouped, position160)
			}
			return true
		l159:
			position, tokenIndex, depth = position159, tokenIndex159, depth159
			return false
		},
		/* 41 Range <- <('[' Expression ('.' '.') Expression ']')> */
		func() bool {
			position161, tokenIndex161, depth161 := position, tokenIndex, depth
			{
				position162 := position
				depth++
				if buffer[position] != rune('[') {
					goto l161
				}
				position++
				if !_rules[ruleExpression]() {
					goto l161
				}
				if buffer[position] != rune('.') {
					goto l161
				}
				position++
				if buffer[position] != rune('.') {
					goto l161
				}
				position++
				if !_rules[ruleExpression]() {
					goto l161
				}
				if buffer[position] != rune(']') {
					goto l161
				}
				position++
				depth--
				add(ruleRange, position162)
			}
			return true
		l161:
			position, tokenIndex, depth = position161, tokenIndex161, depth161
			return false
		},
		/* 42 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position163, tokenIndex163, depth163 := position, tokenIndex, depth
			{
				position164 := position
				depth++
				{
					position165, tokenIndex165, depth165 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l165
					}
					position++
					goto l166
				l165:
					position, tokenIndex, depth = position165, tokenIndex165, depth165
				}
			l166:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l163
				}
				position++
			l167:
				{
					position168, tokenIndex168, depth168 := position, tokenIndex, depth
					{
						position169, tokenIndex169, depth169 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l170
						}
						position++
						goto l169
					l170:
						position, tokenIndex, depth = position169, tokenIndex169, depth169
						if buffer[position] != rune('_') {
							goto l168
						}
						position++
					}
				l169:
					goto l167
				l168:
					position, tokenIndex, depth = position168, tokenIndex168, depth168
				}
				depth--
				add(ruleInteger, position164)
			}
			return true
		l163:
			position, tokenIndex, depth = position163, tokenIndex163, depth163
			return false
		},
		/* 43 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position171, tokenIndex171, depth171 := position, tokenIndex, depth
			{
				position172 := position
				depth++
				if buffer[position] != rune('"') {
					goto l171
				}
				position++
			l173:
				{
					position174, tokenIndex174, depth174 := position, tokenIndex, depth
					{
						position175, tokenIndex175, depth175 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l176
						}
						position++
						if buffer[position] != rune('"') {
							goto l176
						}
						position++
						goto l175
					l176:
						position, tokenIndex, depth = position175, tokenIndex175, depth175
						{
							position177, tokenIndex177, depth177 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l177
							}
							position++
							goto l174
						l177:
							position, tokenIndex, depth = position177, tokenIndex177, depth177
						}
						if !matchDot() {
							goto l174
						}
					}
				l175:
					goto l173
				l174:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
				}
				if buffer[position] != rune('"') {
					goto l171
				}
				position++
				depth--
				add(ruleString, position172)
			}
			return true
		l171:
			position, tokenIndex, depth = position171, tokenIndex171, depth171
			return false
		},
		/* 44 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position178, tokenIndex178, depth178 := position, tokenIndex, depth
			{
				position179 := position
				depth++
				{
					position180, tokenIndex180, depth180 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l181
					}
					position++
					if buffer[position] != rune('r') {
						goto l181
					}
					position++
					if buffer[position] != rune('u') {
						goto l181
					}
					position++
					if buffer[position] != rune('e') {
						goto l181
					}
					position++
					goto l180
				l181:
					position, tokenIndex, depth = position180, tokenIndex180, depth180
					if buffer[position] != rune('f') {
						goto l178
					}
					position++
					if buffer[position] != rune('a') {
						goto l178
					}
					position++
					if buffer[position] != rune('l') {
						goto l178
					}
					position++
					if buffer[position] != rune('s') {
						goto l178
					}
					position++
					if buffer[position] != rune('e') {
						goto l178
					}
					position++
				}
			l180:
				depth--
				add(ruleBoolean, position179)
			}
			return true
		l178:
			position, tokenIndex, depth = position178, tokenIndex178, depth178
			return false
		},
		/* 45 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position182, tokenIndex182, depth182 := position, tokenIndex, depth
			{
				position183 := position
				depth++
				{
					position184, tokenIndex184, depth184 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l185
					}
					position++
					if buffer[position] != rune('i') {
						goto l185
					}
					position++
					if buffer[position] != rune('l') {
						goto l185
					}
					position++
					goto l184
				l185:
					position, tokenIndex, depth = position184, tokenIndex184, depth184
					if buffer[position] != rune('~') {
						goto l182
					}
					position++
				}
			l184:
				depth--
				add(ruleNil, position183)
			}
			return true
		l182:
			position, tokenIndex, depth = position182, tokenIndex182, depth182
			return false
		},
		/* 46 Undefined <- <('~' '~')> */
		func() bool {
			position186, tokenIndex186, depth186 := position, tokenIndex, depth
			{
				position187 := position
				depth++
				if buffer[position] != rune('~') {
					goto l186
				}
				position++
				if buffer[position] != rune('~') {
					goto l186
				}
				position++
				depth--
				add(ruleUndefined, position187)
			}
			return true
		l186:
			position, tokenIndex, depth = position186, tokenIndex186, depth186
			return false
		},
		/* 47 List <- <(StartList ExpressionList? ']')> */
		func() bool {
			position188, tokenIndex188, depth188 := position, tokenIndex, depth
			{
				position189 := position
				depth++
				if !_rules[ruleStartList]() {
					goto l188
				}
				{
					position190, tokenIndex190, depth190 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l190
					}
					goto l191
				l190:
					position, tokenIndex, depth = position190, tokenIndex190, depth190
				}
			l191:
				if buffer[position] != rune(']') {
					goto l188
				}
				position++
				depth--
				add(ruleList, position189)
			}
			return true
		l188:
			position, tokenIndex, depth = position188, tokenIndex188, depth188
			return false
		},
		/* 48 StartList <- <'['> */
		func() bool {
			position192, tokenIndex192, depth192 := position, tokenIndex, depth
			{
				position193 := position
				depth++
				if buffer[position] != rune('[') {
					goto l192
				}
				position++
				depth--
				add(ruleStartList, position193)
			}
			return true
		l192:
			position, tokenIndex, depth = position192, tokenIndex192, depth192
			return false
		},
		/* 49 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position194, tokenIndex194, depth194 := position, tokenIndex, depth
			{
				position195 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l194
				}
				if !_rules[rulews]() {
					goto l194
				}
				{
					position196, tokenIndex196, depth196 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l196
					}
					goto l197
				l196:
					position, tokenIndex, depth = position196, tokenIndex196, depth196
				}
			l197:
				if buffer[position] != rune('}') {
					goto l194
				}
				position++
				depth--
				add(ruleMap, position195)
			}
			return true
		l194:
			position, tokenIndex, depth = position194, tokenIndex194, depth194
			return false
		},
		/* 50 CreateMap <- <'{'> */
		func() bool {
			position198, tokenIndex198, depth198 := position, tokenIndex, depth
			{
				position199 := position
				depth++
				if buffer[position] != rune('{') {
					goto l198
				}
				position++
				depth--
				add(ruleCreateMap, position199)
			}
			return true
		l198:
			position, tokenIndex, depth = position198, tokenIndex198, depth198
			return false
		},
		/* 51 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position200, tokenIndex200, depth200 := position, tokenIndex, depth
			{
				position201 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l200
				}
			l202:
				{
					position203, tokenIndex203, depth203 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l203
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l203
					}
					goto l202
				l203:
					position, tokenIndex, depth = position203, tokenIndex203, depth203
				}
				depth--
				add(ruleAssignments, position201)
			}
			return true
		l200:
			position, tokenIndex, depth = position200, tokenIndex200, depth200
			return false
		},
		/* 52 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position204, tokenIndex204, depth204 := position, tokenIndex, depth
			{
				position205 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l204
				}
				if buffer[position] != rune('=') {
					goto l204
				}
				position++
				if !_rules[ruleExpression]() {
					goto l204
				}
				depth--
				add(ruleAssignment, position205)
			}
			return true
		l204:
			position, tokenIndex, depth = position204, tokenIndex204, depth204
			return false
		},
		/* 53 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position206, tokenIndex206, depth206 := position, tokenIndex, depth
			{
				position207 := position
				depth++
				{
					position208, tokenIndex208, depth208 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l209
					}
					goto l208
				l209:
					position, tokenIndex, depth = position208, tokenIndex208, depth208
					if !_rules[ruleSimpleMerge]() {
						goto l206
					}
				}
			l208:
				depth--
				add(ruleMerge, position207)
			}
			return true
		l206:
			position, tokenIndex, depth = position206, tokenIndex206, depth206
			return false
		},
		/* 54 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position210, tokenIndex210, depth210 := position, tokenIndex, depth
			{
				position211 := position
				depth++
				if buffer[position] != rune('m') {
					goto l210
				}
				position++
				if buffer[position] != rune('e') {
					goto l210
				}
				position++
				if buffer[position] != rune('r') {
					goto l210
				}
				position++
				if buffer[position] != rune('g') {
					goto l210
				}
				position++
				if buffer[position] != rune('e') {
					goto l210
				}
				position++
				{
					position212, tokenIndex212, depth212 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l212
					}
					if !_rules[ruleRequired]() {
						goto l212
					}
					goto l210
				l212:
					position, tokenIndex, depth = position212, tokenIndex212, depth212
				}
				{
					position213, tokenIndex213, depth213 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l213
					}
					{
						position215, tokenIndex215, depth215 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l216
						}
						goto l215
					l216:
						position, tokenIndex, depth = position215, tokenIndex215, depth215
						if !_rules[ruleOn]() {
							goto l213
						}
					}
				l215:
					goto l214
				l213:
					position, tokenIndex, depth = position213, tokenIndex213, depth213
				}
			l214:
				if !_rules[rulereq_ws]() {
					goto l210
				}
				if !_rules[ruleReference]() {
					goto l210
				}
				depth--
				add(ruleRefMerge, position211)
			}
			return true
		l210:
			position, tokenIndex, depth = position210, tokenIndex210, depth210
			return false
		},
		/* 55 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' !'(' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position217, tokenIndex217, depth217 := position, tokenIndex, depth
			{
				position218 := position
				depth++
				if buffer[position] != rune('m') {
					goto l217
				}
				position++
				if buffer[position] != rune('e') {
					goto l217
				}
				position++
				if buffer[position] != rune('r') {
					goto l217
				}
				position++
				if buffer[position] != rune('g') {
					goto l217
				}
				position++
				if buffer[position] != rune('e') {
					goto l217
				}
				position++
				{
					position219, tokenIndex219, depth219 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l219
					}
					position++
					goto l217
				l219:
					position, tokenIndex, depth = position219, tokenIndex219, depth219
				}
				{
					position220, tokenIndex220, depth220 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l220
					}
					{
						position222, tokenIndex222, depth222 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l223
						}
						goto l222
					l223:
						position, tokenIndex, depth = position222, tokenIndex222, depth222
						if !_rules[ruleRequired]() {
							goto l224
						}
						goto l222
					l224:
						position, tokenIndex, depth = position222, tokenIndex222, depth222
						if !_rules[ruleOn]() {
							goto l220
						}
					}
				l222:
					goto l221
				l220:
					position, tokenIndex, depth = position220, tokenIndex220, depth220
				}
			l221:
				depth--
				add(ruleSimpleMerge, position218)
			}
			return true
		l217:
			position, tokenIndex, depth = position217, tokenIndex217, depth217
			return false
		},
		/* 56 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position225, tokenIndex225, depth225 := position, tokenIndex, depth
			{
				position226 := position
				depth++
				if buffer[position] != rune('r') {
					goto l225
				}
				position++
				if buffer[position] != rune('e') {
					goto l225
				}
				position++
				if buffer[position] != rune('p') {
					goto l225
				}
				position++
				if buffer[position] != rune('l') {
					goto l225
				}
				position++
				if buffer[position] != rune('a') {
					goto l225
				}
				position++
				if buffer[position] != rune('c') {
					goto l225
				}
				position++
				if buffer[position] != rune('e') {
					goto l225
				}
				position++
				depth--
				add(ruleReplace, position226)
			}
			return true
		l225:
			position, tokenIndex, depth = position225, tokenIndex225, depth225
			return false
		},
		/* 57 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position227, tokenIndex227, depth227 := position, tokenIndex, depth
			{
				position228 := position
				depth++
				if buffer[position] != rune('r') {
					goto l227
				}
				position++
				if buffer[position] != rune('e') {
					goto l227
				}
				position++
				if buffer[position] != rune('q') {
					goto l227
				}
				position++
				if buffer[position] != rune('u') {
					goto l227
				}
				position++
				if buffer[position] != rune('i') {
					goto l227
				}
				position++
				if buffer[position] != rune('r') {
					goto l227
				}
				position++
				if buffer[position] != rune('e') {
					goto l227
				}
				position++
				if buffer[position] != rune('d') {
					goto l227
				}
				position++
				depth--
				add(ruleRequired, position228)
			}
			return true
		l227:
			position, tokenIndex, depth = position227, tokenIndex227, depth227
			return false
		},
		/* 58 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position229, tokenIndex229, depth229 := position, tokenIndex, depth
			{
				position230 := position
				depth++
				if buffer[position] != rune('o') {
					goto l229
				}
				position++
				if buffer[position] != rune('n') {
					goto l229
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l229
				}
				if !_rules[ruleName]() {
					goto l229
				}
				depth--
				add(ruleOn, position230)
			}
			return true
		l229:
			position, tokenIndex, depth = position229, tokenIndex229, depth229
			return false
		},
		/* 59 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position231, tokenIndex231, depth231 := position, tokenIndex, depth
			{
				position232 := position
				depth++
				if buffer[position] != rune('a') {
					goto l231
				}
				position++
				if buffer[position] != rune('u') {
					goto l231
				}
				position++
				if buffer[position] != rune('t') {
					goto l231
				}
				position++
				if buffer[position] != rune('o') {
					goto l231
				}
				position++
				depth--
				add(ruleAuto, position232)
			}
			return true
		l231:
			position, tokenIndex, depth = position231, tokenIndex231, depth231
			return false
		},
		/* 60 Mapping <- <('m' 'a' 'p' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position233, tokenIndex233, depth233 := position, tokenIndex, depth
			{
				position234 := position
				depth++
				if buffer[position] != rune('m') {
					goto l233
				}
				position++
				if buffer[position] != rune('a') {
					goto l233
				}
				position++
				if buffer[position] != rune('p') {
					goto l233
				}
				position++
				if buffer[position] != rune('[') {
					goto l233
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l233
				}
				{
					position235, tokenIndex235, depth235 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l236
					}
					goto l235
				l236:
					position, tokenIndex, depth = position235, tokenIndex235, depth235
					if buffer[position] != rune('|') {
						goto l233
					}
					position++
					if !_rules[ruleExpression]() {
						goto l233
					}
				}
			l235:
				if buffer[position] != rune(']') {
					goto l233
				}
				position++
				depth--
				add(ruleMapping, position234)
			}
			return true
		l233:
			position, tokenIndex, depth = position233, tokenIndex233, depth233
			return false
		},
		/* 61 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position237, tokenIndex237, depth237 := position, tokenIndex, depth
			{
				position238 := position
				depth++
				if buffer[position] != rune('s') {
					goto l237
				}
				position++
				if buffer[position] != rune('u') {
					goto l237
				}
				position++
				if buffer[position] != rune('m') {
					goto l237
				}
				position++
				if buffer[position] != rune('[') {
					goto l237
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l237
				}
				if buffer[position] != rune('|') {
					goto l237
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l237
				}
				{
					position239, tokenIndex239, depth239 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l240
					}
					goto l239
				l240:
					position, tokenIndex, depth = position239, tokenIndex239, depth239
					if buffer[position] != rune('|') {
						goto l237
					}
					position++
					if !_rules[ruleExpression]() {
						goto l237
					}
				}
			l239:
				if buffer[position] != rune(']') {
					goto l237
				}
				position++
				depth--
				add(ruleSum, position238)
			}
			return true
		l237:
			position, tokenIndex, depth = position237, tokenIndex237, depth237
			return false
		},
		/* 62 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position241, tokenIndex241, depth241 := position, tokenIndex, depth
			{
				position242 := position
				depth++
				if buffer[position] != rune('l') {
					goto l241
				}
				position++
				if buffer[position] != rune('a') {
					goto l241
				}
				position++
				if buffer[position] != rune('m') {
					goto l241
				}
				position++
				if buffer[position] != rune('b') {
					goto l241
				}
				position++
				if buffer[position] != rune('d') {
					goto l241
				}
				position++
				if buffer[position] != rune('a') {
					goto l241
				}
				position++
				{
					position243, tokenIndex243, depth243 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l244
					}
					goto l243
				l244:
					position, tokenIndex, depth = position243, tokenIndex243, depth243
					if !_rules[ruleLambdaExpr]() {
						goto l241
					}
				}
			l243:
				depth--
				add(ruleLambda, position242)
			}
			return true
		l241:
			position, tokenIndex, depth = position241, tokenIndex241, depth241
			return false
		},
		/* 63 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position245, tokenIndex245, depth245 := position, tokenIndex, depth
			{
				position246 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l245
				}
				if !_rules[ruleExpression]() {
					goto l245
				}
				depth--
				add(ruleLambdaRef, position246)
			}
			return true
		l245:
			position, tokenIndex, depth = position245, tokenIndex245, depth245
			return false
		},
		/* 64 LambdaExpr <- <(ws '|' ws Name NextName* ws '|' ws ('-' '>') Expression)> */
		func() bool {
			position247, tokenIndex247, depth247 := position, tokenIndex, depth
			{
				position248 := position
				depth++
				if !_rules[rulews]() {
					goto l247
				}
				if buffer[position] != rune('|') {
					goto l247
				}
				position++
				if !_rules[rulews]() {
					goto l247
				}
				if !_rules[ruleName]() {
					goto l247
				}
			l249:
				{
					position250, tokenIndex250, depth250 := position, tokenIndex, depth
					if !_rules[ruleNextName]() {
						goto l250
					}
					goto l249
				l250:
					position, tokenIndex, depth = position250, tokenIndex250, depth250
				}
				if !_rules[rulews]() {
					goto l247
				}
				if buffer[position] != rune('|') {
					goto l247
				}
				position++
				if !_rules[rulews]() {
					goto l247
				}
				if buffer[position] != rune('-') {
					goto l247
				}
				position++
				if buffer[position] != rune('>') {
					goto l247
				}
				position++
				if !_rules[ruleExpression]() {
					goto l247
				}
				depth--
				add(ruleLambdaExpr, position248)
			}
			return true
		l247:
			position, tokenIndex, depth = position247, tokenIndex247, depth247
			return false
		},
		/* 65 NextName <- <(ws ',' ws Name)> */
		func() bool {
			position251, tokenIndex251, depth251 := position, tokenIndex, depth
			{
				position252 := position
				depth++
				if !_rules[rulews]() {
					goto l251
				}
				if buffer[position] != rune(',') {
					goto l251
				}
				position++
				if !_rules[rulews]() {
					goto l251
				}
				if !_rules[ruleName]() {
					goto l251
				}
				depth--
				add(ruleNextName, position252)
			}
			return true
		l251:
			position, tokenIndex, depth = position251, tokenIndex251, depth251
			return false
		},
		/* 66 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position253, tokenIndex253, depth253 := position, tokenIndex, depth
			{
				position254 := position
				depth++
				{
					position257, tokenIndex257, depth257 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l258
					}
					position++
					goto l257
				l258:
					position, tokenIndex, depth = position257, tokenIndex257, depth257
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l259
					}
					position++
					goto l257
				l259:
					position, tokenIndex, depth = position257, tokenIndex257, depth257
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l260
					}
					position++
					goto l257
				l260:
					position, tokenIndex, depth = position257, tokenIndex257, depth257
					if buffer[position] != rune('_') {
						goto l253
					}
					position++
				}
			l257:
			l255:
				{
					position256, tokenIndex256, depth256 := position, tokenIndex, depth
					{
						position261, tokenIndex261, depth261 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l262
						}
						position++
						goto l261
					l262:
						position, tokenIndex, depth = position261, tokenIndex261, depth261
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l263
						}
						position++
						goto l261
					l263:
						position, tokenIndex, depth = position261, tokenIndex261, depth261
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l264
						}
						position++
						goto l261
					l264:
						position, tokenIndex, depth = position261, tokenIndex261, depth261
						if buffer[position] != rune('_') {
							goto l256
						}
						position++
					}
				l261:
					goto l255
				l256:
					position, tokenIndex, depth = position256, tokenIndex256, depth256
				}
				depth--
				add(ruleName, position254)
			}
			return true
		l253:
			position, tokenIndex, depth = position253, tokenIndex253, depth253
			return false
		},
		/* 67 Reference <- <('.'? Key FollowUpRef)> */
		func() bool {
			position265, tokenIndex265, depth265 := position, tokenIndex, depth
			{
				position266 := position
				depth++
				{
					position267, tokenIndex267, depth267 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l267
					}
					position++
					goto l268
				l267:
					position, tokenIndex, depth = position267, tokenIndex267, depth267
				}
			l268:
				if !_rules[ruleKey]() {
					goto l265
				}
				if !_rules[ruleFollowUpRef]() {
					goto l265
				}
				depth--
				add(ruleReference, position266)
			}
			return true
		l265:
			position, tokenIndex, depth = position265, tokenIndex265, depth265
			return false
		},
		/* 68 FollowUpRef <- <('.' (Key / Index))*> */
		func() bool {
			{
				position270 := position
				depth++
			l271:
				{
					position272, tokenIndex272, depth272 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l272
					}
					position++
					{
						position273, tokenIndex273, depth273 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l274
						}
						goto l273
					l274:
						position, tokenIndex, depth = position273, tokenIndex273, depth273
						if !_rules[ruleIndex]() {
							goto l272
						}
					}
				l273:
					goto l271
				l272:
					position, tokenIndex, depth = position272, tokenIndex272, depth272
				}
				depth--
				add(ruleFollowUpRef, position270)
			}
			return true
		},
		/* 69 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position275, tokenIndex275, depth275 := position, tokenIndex, depth
			{
				position276 := position
				depth++
				{
					position277, tokenIndex277, depth277 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l278
					}
					position++
					goto l277
				l278:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l279
					}
					position++
					goto l277
				l279:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l280
					}
					position++
					goto l277
				l280:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
					if buffer[position] != rune('_') {
						goto l275
					}
					position++
				}
			l277:
			l281:
				{
					position282, tokenIndex282, depth282 := position, tokenIndex, depth
					{
						position283, tokenIndex283, depth283 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l284
						}
						position++
						goto l283
					l284:
						position, tokenIndex, depth = position283, tokenIndex283, depth283
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l285
						}
						position++
						goto l283
					l285:
						position, tokenIndex, depth = position283, tokenIndex283, depth283
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l286
						}
						position++
						goto l283
					l286:
						position, tokenIndex, depth = position283, tokenIndex283, depth283
						if buffer[position] != rune('_') {
							goto l287
						}
						position++
						goto l283
					l287:
						position, tokenIndex, depth = position283, tokenIndex283, depth283
						if buffer[position] != rune('-') {
							goto l282
						}
						position++
					}
				l283:
					goto l281
				l282:
					position, tokenIndex, depth = position282, tokenIndex282, depth282
				}
				{
					position288, tokenIndex288, depth288 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l288
					}
					position++
					{
						position290, tokenIndex290, depth290 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l291
						}
						position++
						goto l290
					l291:
						position, tokenIndex, depth = position290, tokenIndex290, depth290
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l292
						}
						position++
						goto l290
					l292:
						position, tokenIndex, depth = position290, tokenIndex290, depth290
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l293
						}
						position++
						goto l290
					l293:
						position, tokenIndex, depth = position290, tokenIndex290, depth290
						if buffer[position] != rune('_') {
							goto l288
						}
						position++
					}
				l290:
				l294:
					{
						position295, tokenIndex295, depth295 := position, tokenIndex, depth
						{
							position296, tokenIndex296, depth296 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l297
							}
							position++
							goto l296
						l297:
							position, tokenIndex, depth = position296, tokenIndex296, depth296
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l298
							}
							position++
							goto l296
						l298:
							position, tokenIndex, depth = position296, tokenIndex296, depth296
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l299
							}
							position++
							goto l296
						l299:
							position, tokenIndex, depth = position296, tokenIndex296, depth296
							if buffer[position] != rune('_') {
								goto l300
							}
							position++
							goto l296
						l300:
							position, tokenIndex, depth = position296, tokenIndex296, depth296
							if buffer[position] != rune('-') {
								goto l295
							}
							position++
						}
					l296:
						goto l294
					l295:
						position, tokenIndex, depth = position295, tokenIndex295, depth295
					}
					goto l289
				l288:
					position, tokenIndex, depth = position288, tokenIndex288, depth288
				}
			l289:
				depth--
				add(ruleKey, position276)
			}
			return true
		l275:
			position, tokenIndex, depth = position275, tokenIndex275, depth275
			return false
		},
		/* 70 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position301, tokenIndex301, depth301 := position, tokenIndex, depth
			{
				position302 := position
				depth++
				if buffer[position] != rune('[') {
					goto l301
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l301
				}
				position++
			l303:
				{
					position304, tokenIndex304, depth304 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l304
					}
					position++
					goto l303
				l304:
					position, tokenIndex, depth = position304, tokenIndex304, depth304
				}
				if buffer[position] != rune(']') {
					goto l301
				}
				position++
				depth--
				add(ruleIndex, position302)
			}
			return true
		l301:
			position, tokenIndex, depth = position301, tokenIndex301, depth301
			return false
		},
		/* 71 IP <- <([0-9]+ '.' [0-9]+ '.' [0-9]+ '.' [0-9]+)> */
		func() bool {
			position305, tokenIndex305, depth305 := position, tokenIndex, depth
			{
				position306 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l305
				}
				position++
			l307:
				{
					position308, tokenIndex308, depth308 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l308
					}
					position++
					goto l307
				l308:
					position, tokenIndex, depth = position308, tokenIndex308, depth308
				}
				if buffer[position] != rune('.') {
					goto l305
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l305
				}
				position++
			l309:
				{
					position310, tokenIndex310, depth310 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l310
					}
					position++
					goto l309
				l310:
					position, tokenIndex, depth = position310, tokenIndex310, depth310
				}
				if buffer[position] != rune('.') {
					goto l305
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l305
				}
				position++
			l311:
				{
					position312, tokenIndex312, depth312 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l312
					}
					position++
					goto l311
				l312:
					position, tokenIndex, depth = position312, tokenIndex312, depth312
				}
				if buffer[position] != rune('.') {
					goto l305
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l305
				}
				position++
			l313:
				{
					position314, tokenIndex314, depth314 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l314
					}
					position++
					goto l313
				l314:
					position, tokenIndex, depth = position314, tokenIndex314, depth314
				}
				depth--
				add(ruleIP, position306)
			}
			return true
		l305:
			position, tokenIndex, depth = position305, tokenIndex305, depth305
			return false
		},
		/* 72 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position316 := position
				depth++
			l317:
				{
					position318, tokenIndex318, depth318 := position, tokenIndex, depth
					{
						position319, tokenIndex319, depth319 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l320
						}
						position++
						goto l319
					l320:
						position, tokenIndex, depth = position319, tokenIndex319, depth319
						if buffer[position] != rune('\t') {
							goto l321
						}
						position++
						goto l319
					l321:
						position, tokenIndex, depth = position319, tokenIndex319, depth319
						if buffer[position] != rune('\n') {
							goto l322
						}
						position++
						goto l319
					l322:
						position, tokenIndex, depth = position319, tokenIndex319, depth319
						if buffer[position] != rune('\r') {
							goto l318
						}
						position++
					}
				l319:
					goto l317
				l318:
					position, tokenIndex, depth = position318, tokenIndex318, depth318
				}
				depth--
				add(rulews, position316)
			}
			return true
		},
		/* 73 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position323, tokenIndex323, depth323 := position, tokenIndex, depth
			{
				position324 := position
				depth++
				{
					position327, tokenIndex327, depth327 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l328
					}
					position++
					goto l327
				l328:
					position, tokenIndex, depth = position327, tokenIndex327, depth327
					if buffer[position] != rune('\t') {
						goto l329
					}
					position++
					goto l327
				l329:
					position, tokenIndex, depth = position327, tokenIndex327, depth327
					if buffer[position] != rune('\n') {
						goto l330
					}
					position++
					goto l327
				l330:
					position, tokenIndex, depth = position327, tokenIndex327, depth327
					if buffer[position] != rune('\r') {
						goto l323
					}
					position++
				}
			l327:
			l325:
				{
					position326, tokenIndex326, depth326 := position, tokenIndex, depth
					{
						position331, tokenIndex331, depth331 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l332
						}
						position++
						goto l331
					l332:
						position, tokenIndex, depth = position331, tokenIndex331, depth331
						if buffer[position] != rune('\t') {
							goto l333
						}
						position++
						goto l331
					l333:
						position, tokenIndex, depth = position331, tokenIndex331, depth331
						if buffer[position] != rune('\n') {
							goto l334
						}
						position++
						goto l331
					l334:
						position, tokenIndex, depth = position331, tokenIndex331, depth331
						if buffer[position] != rune('\r') {
							goto l326
						}
						position++
					}
				l331:
					goto l325
				l326:
					position, tokenIndex, depth = position326, tokenIndex326, depth326
				}
				depth--
				add(rulereq_ws, position324)
			}
			return true
		l323:
			position, tokenIndex, depth = position323, tokenIndex323, depth323
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
