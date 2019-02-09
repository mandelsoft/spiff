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
	ruleScoped
	ruleScope
	ruleCreateScope
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
	ruleSymbol
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
	"Scoped",
	"Scope",
	"CreateScope",
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
	"Symbol",
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
	rules  [81]func() bool
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
		/* 6 Expression <- <(ws (Scoped / LambdaExpr / Level7) ws)> */
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
					if !_rules[ruleScoped]() {
						goto l28
					}
					goto l27
				l28:
					position, tokenIndex, depth = position27, tokenIndex27, depth27
					if !_rules[ruleLambdaExpr]() {
						goto l29
					}
					goto l27
				l29:
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
		/* 7 Scoped <- <(Scope ws (LambdaExpr / Level7))> */
		func() bool {
			position30, tokenIndex30, depth30 := position, tokenIndex, depth
			{
				position31 := position
				depth++
				if !_rules[ruleScope]() {
					goto l30
				}
				if !_rules[rulews]() {
					goto l30
				}
				{
					position32, tokenIndex32, depth32 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l33
					}
					goto l32
				l33:
					position, tokenIndex, depth = position32, tokenIndex32, depth32
					if !_rules[ruleLevel7]() {
						goto l30
					}
				}
			l32:
				depth--
				add(ruleScoped, position31)
			}
			return true
		l30:
			position, tokenIndex, depth = position30, tokenIndex30, depth30
			return false
		},
		/* 8 Scope <- <(CreateScope ws Assignments? ')')> */
		func() bool {
			position34, tokenIndex34, depth34 := position, tokenIndex, depth
			{
				position35 := position
				depth++
				if !_rules[ruleCreateScope]() {
					goto l34
				}
				if !_rules[rulews]() {
					goto l34
				}
				{
					position36, tokenIndex36, depth36 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l36
					}
					goto l37
				l36:
					position, tokenIndex, depth = position36, tokenIndex36, depth36
				}
			l37:
				if buffer[position] != rune(')') {
					goto l34
				}
				position++
				depth--
				add(ruleScope, position35)
			}
			return true
		l34:
			position, tokenIndex, depth = position34, tokenIndex34, depth34
			return false
		},
		/* 9 CreateScope <- <'('> */
		func() bool {
			position38, tokenIndex38, depth38 := position, tokenIndex, depth
			{
				position39 := position
				depth++
				if buffer[position] != rune('(') {
					goto l38
				}
				position++
				depth--
				add(ruleCreateScope, position39)
			}
			return true
		l38:
			position, tokenIndex, depth = position38, tokenIndex38, depth38
			return false
		},
		/* 10 Level7 <- <(Level6 (req_ws Or)*)> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				if !_rules[ruleLevel6]() {
					goto l40
				}
			l42:
				{
					position43, tokenIndex43, depth43 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l43
					}
					if !_rules[ruleOr]() {
						goto l43
					}
					goto l42
				l43:
					position, tokenIndex, depth = position43, tokenIndex43, depth43
				}
				depth--
				add(ruleLevel7, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 11 Or <- <('|' '|' req_ws Level6)> */
		func() bool {
			position44, tokenIndex44, depth44 := position, tokenIndex, depth
			{
				position45 := position
				depth++
				if buffer[position] != rune('|') {
					goto l44
				}
				position++
				if buffer[position] != rune('|') {
					goto l44
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l44
				}
				if !_rules[ruleLevel6]() {
					goto l44
				}
				depth--
				add(ruleOr, position45)
			}
			return true
		l44:
			position, tokenIndex, depth = position44, tokenIndex44, depth44
			return false
		},
		/* 12 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position46, tokenIndex46, depth46 := position, tokenIndex, depth
			{
				position47 := position
				depth++
				{
					position48, tokenIndex48, depth48 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l49
					}
					goto l48
				l49:
					position, tokenIndex, depth = position48, tokenIndex48, depth48
					if !_rules[ruleLevel5]() {
						goto l46
					}
				}
			l48:
				depth--
				add(ruleLevel6, position47)
			}
			return true
		l46:
			position, tokenIndex, depth = position46, tokenIndex46, depth46
			return false
		},
		/* 13 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position50, tokenIndex50, depth50 := position, tokenIndex, depth
			{
				position51 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l50
				}
				if !_rules[rulews]() {
					goto l50
				}
				if buffer[position] != rune('?') {
					goto l50
				}
				position++
				if !_rules[ruleExpression]() {
					goto l50
				}
				if buffer[position] != rune(':') {
					goto l50
				}
				position++
				if !_rules[ruleExpression]() {
					goto l50
				}
				depth--
				add(ruleConditional, position51)
			}
			return true
		l50:
			position, tokenIndex, depth = position50, tokenIndex50, depth50
			return false
		},
		/* 14 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position52, tokenIndex52, depth52 := position, tokenIndex, depth
			{
				position53 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l52
				}
			l54:
				{
					position55, tokenIndex55, depth55 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l55
					}
					goto l54
				l55:
					position, tokenIndex, depth = position55, tokenIndex55, depth55
				}
				depth--
				add(ruleLevel5, position53)
			}
			return true
		l52:
			position, tokenIndex, depth = position52, tokenIndex52, depth52
			return false
		},
		/* 15 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position56, tokenIndex56, depth56 := position, tokenIndex, depth
			{
				position57 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l56
				}
				if !_rules[ruleLevel4]() {
					goto l56
				}
				depth--
				add(ruleConcatenation, position57)
			}
			return true
		l56:
			position, tokenIndex, depth = position56, tokenIndex56, depth56
			return false
		},
		/* 16 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position58, tokenIndex58, depth58 := position, tokenIndex, depth
			{
				position59 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l58
				}
			l60:
				{
					position61, tokenIndex61, depth61 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l61
					}
					{
						position62, tokenIndex62, depth62 := position, tokenIndex, depth
						if !_rules[ruleLogOr]() {
							goto l63
						}
						goto l62
					l63:
						position, tokenIndex, depth = position62, tokenIndex62, depth62
						if !_rules[ruleLogAnd]() {
							goto l61
						}
					}
				l62:
					goto l60
				l61:
					position, tokenIndex, depth = position61, tokenIndex61, depth61
				}
				depth--
				add(ruleLevel4, position59)
			}
			return true
		l58:
			position, tokenIndex, depth = position58, tokenIndex58, depth58
			return false
		},
		/* 17 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position64, tokenIndex64, depth64 := position, tokenIndex, depth
			{
				position65 := position
				depth++
				if buffer[position] != rune('-') {
					goto l64
				}
				position++
				if buffer[position] != rune('o') {
					goto l64
				}
				position++
				if buffer[position] != rune('r') {
					goto l64
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l64
				}
				if !_rules[ruleLevel3]() {
					goto l64
				}
				depth--
				add(ruleLogOr, position65)
			}
			return true
		l64:
			position, tokenIndex, depth = position64, tokenIndex64, depth64
			return false
		},
		/* 18 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position66, tokenIndex66, depth66 := position, tokenIndex, depth
			{
				position67 := position
				depth++
				if buffer[position] != rune('-') {
					goto l66
				}
				position++
				if buffer[position] != rune('a') {
					goto l66
				}
				position++
				if buffer[position] != rune('n') {
					goto l66
				}
				position++
				if buffer[position] != rune('d') {
					goto l66
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l66
				}
				if !_rules[ruleLevel3]() {
					goto l66
				}
				depth--
				add(ruleLogAnd, position67)
			}
			return true
		l66:
			position, tokenIndex, depth = position66, tokenIndex66, depth66
			return false
		},
		/* 19 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position68, tokenIndex68, depth68 := position, tokenIndex, depth
			{
				position69 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l68
				}
			l70:
				{
					position71, tokenIndex71, depth71 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l71
					}
					if !_rules[ruleComparison]() {
						goto l71
					}
					goto l70
				l71:
					position, tokenIndex, depth = position71, tokenIndex71, depth71
				}
				depth--
				add(ruleLevel3, position69)
			}
			return true
		l68:
			position, tokenIndex, depth = position68, tokenIndex68, depth68
			return false
		},
		/* 20 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position72, tokenIndex72, depth72 := position, tokenIndex, depth
			{
				position73 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l72
				}
				if !_rules[rulereq_ws]() {
					goto l72
				}
				if !_rules[ruleLevel2]() {
					goto l72
				}
				depth--
				add(ruleComparison, position73)
			}
			return true
		l72:
			position, tokenIndex, depth = position72, tokenIndex72, depth72
			return false
		},
		/* 21 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position74, tokenIndex74, depth74 := position, tokenIndex, depth
			{
				position75 := position
				depth++
				{
					position76, tokenIndex76, depth76 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l77
					}
					position++
					if buffer[position] != rune('=') {
						goto l77
					}
					position++
					goto l76
				l77:
					position, tokenIndex, depth = position76, tokenIndex76, depth76
					if buffer[position] != rune('!') {
						goto l78
					}
					position++
					if buffer[position] != rune('=') {
						goto l78
					}
					position++
					goto l76
				l78:
					position, tokenIndex, depth = position76, tokenIndex76, depth76
					if buffer[position] != rune('<') {
						goto l79
					}
					position++
					if buffer[position] != rune('=') {
						goto l79
					}
					position++
					goto l76
				l79:
					position, tokenIndex, depth = position76, tokenIndex76, depth76
					if buffer[position] != rune('>') {
						goto l80
					}
					position++
					if buffer[position] != rune('=') {
						goto l80
					}
					position++
					goto l76
				l80:
					position, tokenIndex, depth = position76, tokenIndex76, depth76
					if buffer[position] != rune('>') {
						goto l81
					}
					position++
					goto l76
				l81:
					position, tokenIndex, depth = position76, tokenIndex76, depth76
					if buffer[position] != rune('<') {
						goto l82
					}
					position++
					goto l76
				l82:
					position, tokenIndex, depth = position76, tokenIndex76, depth76
					if buffer[position] != rune('>') {
						goto l74
					}
					position++
				}
			l76:
				depth--
				add(ruleCompareOp, position75)
			}
			return true
		l74:
			position, tokenIndex, depth = position74, tokenIndex74, depth74
			return false
		},
		/* 22 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position83, tokenIndex83, depth83 := position, tokenIndex, depth
			{
				position84 := position
				depth++
				if !_rules[ruleLevel1]() {
					goto l83
				}
			l85:
				{
					position86, tokenIndex86, depth86 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l86
					}
					{
						position87, tokenIndex87, depth87 := position, tokenIndex, depth
						if !_rules[ruleAddition]() {
							goto l88
						}
						goto l87
					l88:
						position, tokenIndex, depth = position87, tokenIndex87, depth87
						if !_rules[ruleSubtraction]() {
							goto l86
						}
					}
				l87:
					goto l85
				l86:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
				}
				depth--
				add(ruleLevel2, position84)
			}
			return true
		l83:
			position, tokenIndex, depth = position83, tokenIndex83, depth83
			return false
		},
		/* 23 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position89, tokenIndex89, depth89 := position, tokenIndex, depth
			{
				position90 := position
				depth++
				if buffer[position] != rune('+') {
					goto l89
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l89
				}
				if !_rules[ruleLevel1]() {
					goto l89
				}
				depth--
				add(ruleAddition, position90)
			}
			return true
		l89:
			position, tokenIndex, depth = position89, tokenIndex89, depth89
			return false
		},
		/* 24 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position91, tokenIndex91, depth91 := position, tokenIndex, depth
			{
				position92 := position
				depth++
				if buffer[position] != rune('-') {
					goto l91
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l91
				}
				if !_rules[ruleLevel1]() {
					goto l91
				}
				depth--
				add(ruleSubtraction, position92)
			}
			return true
		l91:
			position, tokenIndex, depth = position91, tokenIndex91, depth91
			return false
		},
		/* 25 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position93, tokenIndex93, depth93 := position, tokenIndex, depth
			{
				position94 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l93
				}
			l95:
				{
					position96, tokenIndex96, depth96 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l96
					}
					{
						position97, tokenIndex97, depth97 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l98
						}
						goto l97
					l98:
						position, tokenIndex, depth = position97, tokenIndex97, depth97
						if !_rules[ruleDivision]() {
							goto l99
						}
						goto l97
					l99:
						position, tokenIndex, depth = position97, tokenIndex97, depth97
						if !_rules[ruleModulo]() {
							goto l96
						}
					}
				l97:
					goto l95
				l96:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
				}
				depth--
				add(ruleLevel1, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 26 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position100, tokenIndex100, depth100 := position, tokenIndex, depth
			{
				position101 := position
				depth++
				if buffer[position] != rune('*') {
					goto l100
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l100
				}
				if !_rules[ruleLevel0]() {
					goto l100
				}
				depth--
				add(ruleMultiplication, position101)
			}
			return true
		l100:
			position, tokenIndex, depth = position100, tokenIndex100, depth100
			return false
		},
		/* 27 Division <- <('/' req_ws Level0)> */
		func() bool {
			position102, tokenIndex102, depth102 := position, tokenIndex, depth
			{
				position103 := position
				depth++
				if buffer[position] != rune('/') {
					goto l102
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l102
				}
				if !_rules[ruleLevel0]() {
					goto l102
				}
				depth--
				add(ruleDivision, position103)
			}
			return true
		l102:
			position, tokenIndex, depth = position102, tokenIndex102, depth102
			return false
		},
		/* 28 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position104, tokenIndex104, depth104 := position, tokenIndex, depth
			{
				position105 := position
				depth++
				if buffer[position] != rune('%') {
					goto l104
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l104
				}
				if !_rules[ruleLevel0]() {
					goto l104
				}
				depth--
				add(ruleModulo, position105)
			}
			return true
		l104:
			position, tokenIndex, depth = position104, tokenIndex104, depth104
			return false
		},
		/* 29 Level0 <- <(IP / String / Integer / Boolean / Undefined / Nil / Symbol / Not / Substitution / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position106, tokenIndex106, depth106 := position, tokenIndex, depth
			{
				position107 := position
				depth++
				{
					position108, tokenIndex108, depth108 := position, tokenIndex, depth
					if !_rules[ruleIP]() {
						goto l109
					}
					goto l108
				l109:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if !_rules[ruleString]() {
						goto l110
					}
					goto l108
				l110:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if !_rules[ruleInteger]() {
						goto l111
					}
					goto l108
				l111:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if !_rules[ruleBoolean]() {
						goto l112
					}
					goto l108
				l112:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if !_rules[ruleUndefined]() {
						goto l113
					}
					goto l108
				l113:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if !_rules[ruleNil]() {
						goto l114
					}
					goto l108
				l114:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if !_rules[ruleSymbol]() {
						goto l115
					}
					goto l108
				l115:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if !_rules[ruleNot]() {
						goto l116
					}
					goto l108
				l116:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if !_rules[ruleSubstitution]() {
						goto l117
					}
					goto l108
				l117:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if !_rules[ruleMerge]() {
						goto l118
					}
					goto l108
				l118:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if !_rules[ruleAuto]() {
						goto l119
					}
					goto l108
				l119:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if !_rules[ruleLambda]() {
						goto l120
					}
					goto l108
				l120:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
					if !_rules[ruleChained]() {
						goto l106
					}
				}
			l108:
				depth--
				add(ruleLevel0, position107)
			}
			return true
		l106:
			position, tokenIndex, depth = position106, tokenIndex106, depth106
			return false
		},
		/* 30 Chained <- <((Mapping / Selection / Sum / List / Map / Range / Grouped / Reference) ChainedQualifiedExpression*)> */
		func() bool {
			position121, tokenIndex121, depth121 := position, tokenIndex, depth
			{
				position122 := position
				depth++
				{
					position123, tokenIndex123, depth123 := position, tokenIndex, depth
					if !_rules[ruleMapping]() {
						goto l124
					}
					goto l123
				l124:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if !_rules[ruleSelection]() {
						goto l125
					}
					goto l123
				l125:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if !_rules[ruleSum]() {
						goto l126
					}
					goto l123
				l126:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if !_rules[ruleList]() {
						goto l127
					}
					goto l123
				l127:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if !_rules[ruleMap]() {
						goto l128
					}
					goto l123
				l128:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if !_rules[ruleRange]() {
						goto l129
					}
					goto l123
				l129:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if !_rules[ruleGrouped]() {
						goto l130
					}
					goto l123
				l130:
					position, tokenIndex, depth = position123, tokenIndex123, depth123
					if !_rules[ruleReference]() {
						goto l121
					}
				}
			l123:
			l131:
				{
					position132, tokenIndex132, depth132 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l132
					}
					goto l131
				l132:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
				}
				depth--
				add(ruleChained, position122)
			}
			return true
		l121:
			position, tokenIndex, depth = position121, tokenIndex121, depth121
			return false
		},
		/* 31 ChainedQualifiedExpression <- <(ChainedCall / ('.' (ChainedRef / ChainedDynRef / Projection)))> */
		func() bool {
			position133, tokenIndex133, depth133 := position, tokenIndex, depth
			{
				position134 := position
				depth++
				{
					position135, tokenIndex135, depth135 := position, tokenIndex, depth
					if !_rules[ruleChainedCall]() {
						goto l136
					}
					goto l135
				l136:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
					if buffer[position] != rune('.') {
						goto l133
					}
					position++
					{
						position137, tokenIndex137, depth137 := position, tokenIndex, depth
						if !_rules[ruleChainedRef]() {
							goto l138
						}
						goto l137
					l138:
						position, tokenIndex, depth = position137, tokenIndex137, depth137
						if !_rules[ruleChainedDynRef]() {
							goto l139
						}
						goto l137
					l139:
						position, tokenIndex, depth = position137, tokenIndex137, depth137
						if !_rules[ruleProjection]() {
							goto l133
						}
					}
				l137:
				}
			l135:
				depth--
				add(ruleChainedQualifiedExpression, position134)
			}
			return true
		l133:
			position, tokenIndex, depth = position133, tokenIndex133, depth133
			return false
		},
		/* 32 ChainedRef <- <((Key / Index) FollowUpRef)> */
		func() bool {
			position140, tokenIndex140, depth140 := position, tokenIndex, depth
			{
				position141 := position
				depth++
				{
					position142, tokenIndex142, depth142 := position, tokenIndex, depth
					if !_rules[ruleKey]() {
						goto l143
					}
					goto l142
				l143:
					position, tokenIndex, depth = position142, tokenIndex142, depth142
					if !_rules[ruleIndex]() {
						goto l140
					}
				}
			l142:
				if !_rules[ruleFollowUpRef]() {
					goto l140
				}
				depth--
				add(ruleChainedRef, position141)
			}
			return true
		l140:
			position, tokenIndex, depth = position140, tokenIndex140, depth140
			return false
		},
		/* 33 ChainedDynRef <- <('[' Expression ']')> */
		func() bool {
			position144, tokenIndex144, depth144 := position, tokenIndex, depth
			{
				position145 := position
				depth++
				if buffer[position] != rune('[') {
					goto l144
				}
				position++
				if !_rules[ruleExpression]() {
					goto l144
				}
				if buffer[position] != rune(']') {
					goto l144
				}
				position++
				depth--
				add(ruleChainedDynRef, position145)
			}
			return true
		l144:
			position, tokenIndex, depth = position144, tokenIndex144, depth144
			return false
		},
		/* 34 Slice <- <Range> */
		func() bool {
			position146, tokenIndex146, depth146 := position, tokenIndex, depth
			{
				position147 := position
				depth++
				if !_rules[ruleRange]() {
					goto l146
				}
				depth--
				add(ruleSlice, position147)
			}
			return true
		l146:
			position, tokenIndex, depth = position146, tokenIndex146, depth146
			return false
		},
		/* 35 ChainedCall <- <(StartArguments ExpressionList? ')')> */
		func() bool {
			position148, tokenIndex148, depth148 := position, tokenIndex, depth
			{
				position149 := position
				depth++
				if !_rules[ruleStartArguments]() {
					goto l148
				}
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l150
					}
					goto l151
				l150:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
				}
			l151:
				if buffer[position] != rune(')') {
					goto l148
				}
				position++
				depth--
				add(ruleChainedCall, position149)
			}
			return true
		l148:
			position, tokenIndex, depth = position148, tokenIndex148, depth148
			return false
		},
		/* 36 StartArguments <- <('(' ws)> */
		func() bool {
			position152, tokenIndex152, depth152 := position, tokenIndex, depth
			{
				position153 := position
				depth++
				if buffer[position] != rune('(') {
					goto l152
				}
				position++
				if !_rules[rulews]() {
					goto l152
				}
				depth--
				add(ruleStartArguments, position153)
			}
			return true
		l152:
			position, tokenIndex, depth = position152, tokenIndex152, depth152
			return false
		},
		/* 37 ExpressionList <- <(NextExpression (',' NextExpression)*)> */
		func() bool {
			position154, tokenIndex154, depth154 := position, tokenIndex, depth
			{
				position155 := position
				depth++
				if !_rules[ruleNextExpression]() {
					goto l154
				}
			l156:
				{
					position157, tokenIndex157, depth157 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l157
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l157
					}
					goto l156
				l157:
					position, tokenIndex, depth = position157, tokenIndex157, depth157
				}
				depth--
				add(ruleExpressionList, position155)
			}
			return true
		l154:
			position, tokenIndex, depth = position154, tokenIndex154, depth154
			return false
		},
		/* 38 NextExpression <- <Expression> */
		func() bool {
			position158, tokenIndex158, depth158 := position, tokenIndex, depth
			{
				position159 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l158
				}
				depth--
				add(ruleNextExpression, position159)
			}
			return true
		l158:
			position, tokenIndex, depth = position158, tokenIndex158, depth158
			return false
		},
		/* 39 Projection <- <((('[' '*' ']') / Slice) ProjectionValue ChainedQualifiedExpression*)> */
		func() bool {
			position160, tokenIndex160, depth160 := position, tokenIndex, depth
			{
				position161 := position
				depth++
				{
					position162, tokenIndex162, depth162 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l163
					}
					position++
					if buffer[position] != rune('*') {
						goto l163
					}
					position++
					if buffer[position] != rune(']') {
						goto l163
					}
					position++
					goto l162
				l163:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if !_rules[ruleSlice]() {
						goto l160
					}
				}
			l162:
				if !_rules[ruleProjectionValue]() {
					goto l160
				}
			l164:
				{
					position165, tokenIndex165, depth165 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l165
					}
					goto l164
				l165:
					position, tokenIndex, depth = position165, tokenIndex165, depth165
				}
				depth--
				add(ruleProjection, position161)
			}
			return true
		l160:
			position, tokenIndex, depth = position160, tokenIndex160, depth160
			return false
		},
		/* 40 ProjectionValue <- <Action0> */
		func() bool {
			position166, tokenIndex166, depth166 := position, tokenIndex, depth
			{
				position167 := position
				depth++
				if !_rules[ruleAction0]() {
					goto l166
				}
				depth--
				add(ruleProjectionValue, position167)
			}
			return true
		l166:
			position, tokenIndex, depth = position166, tokenIndex166, depth166
			return false
		},
		/* 41 Substitution <- <('*' Level0)> */
		func() bool {
			position168, tokenIndex168, depth168 := position, tokenIndex, depth
			{
				position169 := position
				depth++
				if buffer[position] != rune('*') {
					goto l168
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l168
				}
				depth--
				add(ruleSubstitution, position169)
			}
			return true
		l168:
			position, tokenIndex, depth = position168, tokenIndex168, depth168
			return false
		},
		/* 42 Not <- <('!' ws Level0)> */
		func() bool {
			position170, tokenIndex170, depth170 := position, tokenIndex, depth
			{
				position171 := position
				depth++
				if buffer[position] != rune('!') {
					goto l170
				}
				position++
				if !_rules[rulews]() {
					goto l170
				}
				if !_rules[ruleLevel0]() {
					goto l170
				}
				depth--
				add(ruleNot, position171)
			}
			return true
		l170:
			position, tokenIndex, depth = position170, tokenIndex170, depth170
			return false
		},
		/* 43 Grouped <- <('(' Expression ')')> */
		func() bool {
			position172, tokenIndex172, depth172 := position, tokenIndex, depth
			{
				position173 := position
				depth++
				if buffer[position] != rune('(') {
					goto l172
				}
				position++
				if !_rules[ruleExpression]() {
					goto l172
				}
				if buffer[position] != rune(')') {
					goto l172
				}
				position++
				depth--
				add(ruleGrouped, position173)
			}
			return true
		l172:
			position, tokenIndex, depth = position172, tokenIndex172, depth172
			return false
		},
		/* 44 Range <- <('[' Expression ('.' '.') Expression ']')> */
		func() bool {
			position174, tokenIndex174, depth174 := position, tokenIndex, depth
			{
				position175 := position
				depth++
				if buffer[position] != rune('[') {
					goto l174
				}
				position++
				if !_rules[ruleExpression]() {
					goto l174
				}
				if buffer[position] != rune('.') {
					goto l174
				}
				position++
				if buffer[position] != rune('.') {
					goto l174
				}
				position++
				if !_rules[ruleExpression]() {
					goto l174
				}
				if buffer[position] != rune(']') {
					goto l174
				}
				position++
				depth--
				add(ruleRange, position175)
			}
			return true
		l174:
			position, tokenIndex, depth = position174, tokenIndex174, depth174
			return false
		},
		/* 45 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position176, tokenIndex176, depth176 := position, tokenIndex, depth
			{
				position177 := position
				depth++
				{
					position178, tokenIndex178, depth178 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l178
					}
					position++
					goto l179
				l178:
					position, tokenIndex, depth = position178, tokenIndex178, depth178
				}
			l179:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l176
				}
				position++
			l180:
				{
					position181, tokenIndex181, depth181 := position, tokenIndex, depth
					{
						position182, tokenIndex182, depth182 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l183
						}
						position++
						goto l182
					l183:
						position, tokenIndex, depth = position182, tokenIndex182, depth182
						if buffer[position] != rune('_') {
							goto l181
						}
						position++
					}
				l182:
					goto l180
				l181:
					position, tokenIndex, depth = position181, tokenIndex181, depth181
				}
				depth--
				add(ruleInteger, position177)
			}
			return true
		l176:
			position, tokenIndex, depth = position176, tokenIndex176, depth176
			return false
		},
		/* 46 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position184, tokenIndex184, depth184 := position, tokenIndex, depth
			{
				position185 := position
				depth++
				if buffer[position] != rune('"') {
					goto l184
				}
				position++
			l186:
				{
					position187, tokenIndex187, depth187 := position, tokenIndex, depth
					{
						position188, tokenIndex188, depth188 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l189
						}
						position++
						if buffer[position] != rune('"') {
							goto l189
						}
						position++
						goto l188
					l189:
						position, tokenIndex, depth = position188, tokenIndex188, depth188
						{
							position190, tokenIndex190, depth190 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l190
							}
							position++
							goto l187
						l190:
							position, tokenIndex, depth = position190, tokenIndex190, depth190
						}
						if !matchDot() {
							goto l187
						}
					}
				l188:
					goto l186
				l187:
					position, tokenIndex, depth = position187, tokenIndex187, depth187
				}
				if buffer[position] != rune('"') {
					goto l184
				}
				position++
				depth--
				add(ruleString, position185)
			}
			return true
		l184:
			position, tokenIndex, depth = position184, tokenIndex184, depth184
			return false
		},
		/* 47 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position191, tokenIndex191, depth191 := position, tokenIndex, depth
			{
				position192 := position
				depth++
				{
					position193, tokenIndex193, depth193 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l194
					}
					position++
					if buffer[position] != rune('r') {
						goto l194
					}
					position++
					if buffer[position] != rune('u') {
						goto l194
					}
					position++
					if buffer[position] != rune('e') {
						goto l194
					}
					position++
					goto l193
				l194:
					position, tokenIndex, depth = position193, tokenIndex193, depth193
					if buffer[position] != rune('f') {
						goto l191
					}
					position++
					if buffer[position] != rune('a') {
						goto l191
					}
					position++
					if buffer[position] != rune('l') {
						goto l191
					}
					position++
					if buffer[position] != rune('s') {
						goto l191
					}
					position++
					if buffer[position] != rune('e') {
						goto l191
					}
					position++
				}
			l193:
				depth--
				add(ruleBoolean, position192)
			}
			return true
		l191:
			position, tokenIndex, depth = position191, tokenIndex191, depth191
			return false
		},
		/* 48 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position195, tokenIndex195, depth195 := position, tokenIndex, depth
			{
				position196 := position
				depth++
				{
					position197, tokenIndex197, depth197 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l198
					}
					position++
					if buffer[position] != rune('i') {
						goto l198
					}
					position++
					if buffer[position] != rune('l') {
						goto l198
					}
					position++
					goto l197
				l198:
					position, tokenIndex, depth = position197, tokenIndex197, depth197
					if buffer[position] != rune('~') {
						goto l195
					}
					position++
				}
			l197:
				depth--
				add(ruleNil, position196)
			}
			return true
		l195:
			position, tokenIndex, depth = position195, tokenIndex195, depth195
			return false
		},
		/* 49 Undefined <- <('~' '~')> */
		func() bool {
			position199, tokenIndex199, depth199 := position, tokenIndex, depth
			{
				position200 := position
				depth++
				if buffer[position] != rune('~') {
					goto l199
				}
				position++
				if buffer[position] != rune('~') {
					goto l199
				}
				position++
				depth--
				add(ruleUndefined, position200)
			}
			return true
		l199:
			position, tokenIndex, depth = position199, tokenIndex199, depth199
			return false
		},
		/* 50 Symbol <- <('$' Name)> */
		func() bool {
			position201, tokenIndex201, depth201 := position, tokenIndex, depth
			{
				position202 := position
				depth++
				if buffer[position] != rune('$') {
					goto l201
				}
				position++
				if !_rules[ruleName]() {
					goto l201
				}
				depth--
				add(ruleSymbol, position202)
			}
			return true
		l201:
			position, tokenIndex, depth = position201, tokenIndex201, depth201
			return false
		},
		/* 51 List <- <(StartList ExpressionList? ']')> */
		func() bool {
			position203, tokenIndex203, depth203 := position, tokenIndex, depth
			{
				position204 := position
				depth++
				if !_rules[ruleStartList]() {
					goto l203
				}
				{
					position205, tokenIndex205, depth205 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l205
					}
					goto l206
				l205:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
				}
			l206:
				if buffer[position] != rune(']') {
					goto l203
				}
				position++
				depth--
				add(ruleList, position204)
			}
			return true
		l203:
			position, tokenIndex, depth = position203, tokenIndex203, depth203
			return false
		},
		/* 52 StartList <- <'['> */
		func() bool {
			position207, tokenIndex207, depth207 := position, tokenIndex, depth
			{
				position208 := position
				depth++
				if buffer[position] != rune('[') {
					goto l207
				}
				position++
				depth--
				add(ruleStartList, position208)
			}
			return true
		l207:
			position, tokenIndex, depth = position207, tokenIndex207, depth207
			return false
		},
		/* 53 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position209, tokenIndex209, depth209 := position, tokenIndex, depth
			{
				position210 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l209
				}
				if !_rules[rulews]() {
					goto l209
				}
				{
					position211, tokenIndex211, depth211 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l211
					}
					goto l212
				l211:
					position, tokenIndex, depth = position211, tokenIndex211, depth211
				}
			l212:
				if buffer[position] != rune('}') {
					goto l209
				}
				position++
				depth--
				add(ruleMap, position210)
			}
			return true
		l209:
			position, tokenIndex, depth = position209, tokenIndex209, depth209
			return false
		},
		/* 54 CreateMap <- <'{'> */
		func() bool {
			position213, tokenIndex213, depth213 := position, tokenIndex, depth
			{
				position214 := position
				depth++
				if buffer[position] != rune('{') {
					goto l213
				}
				position++
				depth--
				add(ruleCreateMap, position214)
			}
			return true
		l213:
			position, tokenIndex, depth = position213, tokenIndex213, depth213
			return false
		},
		/* 55 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position215, tokenIndex215, depth215 := position, tokenIndex, depth
			{
				position216 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l215
				}
			l217:
				{
					position218, tokenIndex218, depth218 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l218
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l218
					}
					goto l217
				l218:
					position, tokenIndex, depth = position218, tokenIndex218, depth218
				}
				depth--
				add(ruleAssignments, position216)
			}
			return true
		l215:
			position, tokenIndex, depth = position215, tokenIndex215, depth215
			return false
		},
		/* 56 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position219, tokenIndex219, depth219 := position, tokenIndex, depth
			{
				position220 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l219
				}
				if buffer[position] != rune('=') {
					goto l219
				}
				position++
				if !_rules[ruleExpression]() {
					goto l219
				}
				depth--
				add(ruleAssignment, position220)
			}
			return true
		l219:
			position, tokenIndex, depth = position219, tokenIndex219, depth219
			return false
		},
		/* 57 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position221, tokenIndex221, depth221 := position, tokenIndex, depth
			{
				position222 := position
				depth++
				{
					position223, tokenIndex223, depth223 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l224
					}
					goto l223
				l224:
					position, tokenIndex, depth = position223, tokenIndex223, depth223
					if !_rules[ruleSimpleMerge]() {
						goto l221
					}
				}
			l223:
				depth--
				add(ruleMerge, position222)
			}
			return true
		l221:
			position, tokenIndex, depth = position221, tokenIndex221, depth221
			return false
		},
		/* 58 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position225, tokenIndex225, depth225 := position, tokenIndex, depth
			{
				position226 := position
				depth++
				if buffer[position] != rune('m') {
					goto l225
				}
				position++
				if buffer[position] != rune('e') {
					goto l225
				}
				position++
				if buffer[position] != rune('r') {
					goto l225
				}
				position++
				if buffer[position] != rune('g') {
					goto l225
				}
				position++
				if buffer[position] != rune('e') {
					goto l225
				}
				position++
				{
					position227, tokenIndex227, depth227 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l227
					}
					if !_rules[ruleRequired]() {
						goto l227
					}
					goto l225
				l227:
					position, tokenIndex, depth = position227, tokenIndex227, depth227
				}
				{
					position228, tokenIndex228, depth228 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l228
					}
					{
						position230, tokenIndex230, depth230 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l231
						}
						goto l230
					l231:
						position, tokenIndex, depth = position230, tokenIndex230, depth230
						if !_rules[ruleOn]() {
							goto l228
						}
					}
				l230:
					goto l229
				l228:
					position, tokenIndex, depth = position228, tokenIndex228, depth228
				}
			l229:
				if !_rules[rulereq_ws]() {
					goto l225
				}
				if !_rules[ruleReference]() {
					goto l225
				}
				depth--
				add(ruleRefMerge, position226)
			}
			return true
		l225:
			position, tokenIndex, depth = position225, tokenIndex225, depth225
			return false
		},
		/* 59 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' !'(' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position232, tokenIndex232, depth232 := position, tokenIndex, depth
			{
				position233 := position
				depth++
				if buffer[position] != rune('m') {
					goto l232
				}
				position++
				if buffer[position] != rune('e') {
					goto l232
				}
				position++
				if buffer[position] != rune('r') {
					goto l232
				}
				position++
				if buffer[position] != rune('g') {
					goto l232
				}
				position++
				if buffer[position] != rune('e') {
					goto l232
				}
				position++
				{
					position234, tokenIndex234, depth234 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l234
					}
					position++
					goto l232
				l234:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
				}
				{
					position235, tokenIndex235, depth235 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l235
					}
					{
						position237, tokenIndex237, depth237 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l238
						}
						goto l237
					l238:
						position, tokenIndex, depth = position237, tokenIndex237, depth237
						if !_rules[ruleRequired]() {
							goto l239
						}
						goto l237
					l239:
						position, tokenIndex, depth = position237, tokenIndex237, depth237
						if !_rules[ruleOn]() {
							goto l235
						}
					}
				l237:
					goto l236
				l235:
					position, tokenIndex, depth = position235, tokenIndex235, depth235
				}
			l236:
				depth--
				add(ruleSimpleMerge, position233)
			}
			return true
		l232:
			position, tokenIndex, depth = position232, tokenIndex232, depth232
			return false
		},
		/* 60 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position240, tokenIndex240, depth240 := position, tokenIndex, depth
			{
				position241 := position
				depth++
				if buffer[position] != rune('r') {
					goto l240
				}
				position++
				if buffer[position] != rune('e') {
					goto l240
				}
				position++
				if buffer[position] != rune('p') {
					goto l240
				}
				position++
				if buffer[position] != rune('l') {
					goto l240
				}
				position++
				if buffer[position] != rune('a') {
					goto l240
				}
				position++
				if buffer[position] != rune('c') {
					goto l240
				}
				position++
				if buffer[position] != rune('e') {
					goto l240
				}
				position++
				depth--
				add(ruleReplace, position241)
			}
			return true
		l240:
			position, tokenIndex, depth = position240, tokenIndex240, depth240
			return false
		},
		/* 61 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position242, tokenIndex242, depth242 := position, tokenIndex, depth
			{
				position243 := position
				depth++
				if buffer[position] != rune('r') {
					goto l242
				}
				position++
				if buffer[position] != rune('e') {
					goto l242
				}
				position++
				if buffer[position] != rune('q') {
					goto l242
				}
				position++
				if buffer[position] != rune('u') {
					goto l242
				}
				position++
				if buffer[position] != rune('i') {
					goto l242
				}
				position++
				if buffer[position] != rune('r') {
					goto l242
				}
				position++
				if buffer[position] != rune('e') {
					goto l242
				}
				position++
				if buffer[position] != rune('d') {
					goto l242
				}
				position++
				depth--
				add(ruleRequired, position243)
			}
			return true
		l242:
			position, tokenIndex, depth = position242, tokenIndex242, depth242
			return false
		},
		/* 62 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position244, tokenIndex244, depth244 := position, tokenIndex, depth
			{
				position245 := position
				depth++
				if buffer[position] != rune('o') {
					goto l244
				}
				position++
				if buffer[position] != rune('n') {
					goto l244
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l244
				}
				if !_rules[ruleName]() {
					goto l244
				}
				depth--
				add(ruleOn, position245)
			}
			return true
		l244:
			position, tokenIndex, depth = position244, tokenIndex244, depth244
			return false
		},
		/* 63 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position246, tokenIndex246, depth246 := position, tokenIndex, depth
			{
				position247 := position
				depth++
				if buffer[position] != rune('a') {
					goto l246
				}
				position++
				if buffer[position] != rune('u') {
					goto l246
				}
				position++
				if buffer[position] != rune('t') {
					goto l246
				}
				position++
				if buffer[position] != rune('o') {
					goto l246
				}
				position++
				depth--
				add(ruleAuto, position247)
			}
			return true
		l246:
			position, tokenIndex, depth = position246, tokenIndex246, depth246
			return false
		},
		/* 64 Mapping <- <('m' 'a' 'p' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position248, tokenIndex248, depth248 := position, tokenIndex, depth
			{
				position249 := position
				depth++
				if buffer[position] != rune('m') {
					goto l248
				}
				position++
				if buffer[position] != rune('a') {
					goto l248
				}
				position++
				if buffer[position] != rune('p') {
					goto l248
				}
				position++
				if buffer[position] != rune('[') {
					goto l248
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l248
				}
				{
					position250, tokenIndex250, depth250 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l251
					}
					goto l250
				l251:
					position, tokenIndex, depth = position250, tokenIndex250, depth250
					if buffer[position] != rune('|') {
						goto l248
					}
					position++
					if !_rules[ruleExpression]() {
						goto l248
					}
				}
			l250:
				if buffer[position] != rune(']') {
					goto l248
				}
				position++
				depth--
				add(ruleMapping, position249)
			}
			return true
		l248:
			position, tokenIndex, depth = position248, tokenIndex248, depth248
			return false
		},
		/* 65 Selection <- <('s' 'e' 'l' 'e' 'c' 't' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position252, tokenIndex252, depth252 := position, tokenIndex, depth
			{
				position253 := position
				depth++
				if buffer[position] != rune('s') {
					goto l252
				}
				position++
				if buffer[position] != rune('e') {
					goto l252
				}
				position++
				if buffer[position] != rune('l') {
					goto l252
				}
				position++
				if buffer[position] != rune('e') {
					goto l252
				}
				position++
				if buffer[position] != rune('c') {
					goto l252
				}
				position++
				if buffer[position] != rune('t') {
					goto l252
				}
				position++
				if buffer[position] != rune('[') {
					goto l252
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l252
				}
				{
					position254, tokenIndex254, depth254 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l255
					}
					goto l254
				l255:
					position, tokenIndex, depth = position254, tokenIndex254, depth254
					if buffer[position] != rune('|') {
						goto l252
					}
					position++
					if !_rules[ruleExpression]() {
						goto l252
					}
				}
			l254:
				if buffer[position] != rune(']') {
					goto l252
				}
				position++
				depth--
				add(ruleSelection, position253)
			}
			return true
		l252:
			position, tokenIndex, depth = position252, tokenIndex252, depth252
			return false
		},
		/* 66 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position256, tokenIndex256, depth256 := position, tokenIndex, depth
			{
				position257 := position
				depth++
				if buffer[position] != rune('s') {
					goto l256
				}
				position++
				if buffer[position] != rune('u') {
					goto l256
				}
				position++
				if buffer[position] != rune('m') {
					goto l256
				}
				position++
				if buffer[position] != rune('[') {
					goto l256
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l256
				}
				if buffer[position] != rune('|') {
					goto l256
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l256
				}
				{
					position258, tokenIndex258, depth258 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l259
					}
					goto l258
				l259:
					position, tokenIndex, depth = position258, tokenIndex258, depth258
					if buffer[position] != rune('|') {
						goto l256
					}
					position++
					if !_rules[ruleExpression]() {
						goto l256
					}
				}
			l258:
				if buffer[position] != rune(']') {
					goto l256
				}
				position++
				depth--
				add(ruleSum, position257)
			}
			return true
		l256:
			position, tokenIndex, depth = position256, tokenIndex256, depth256
			return false
		},
		/* 67 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position260, tokenIndex260, depth260 := position, tokenIndex, depth
			{
				position261 := position
				depth++
				if buffer[position] != rune('l') {
					goto l260
				}
				position++
				if buffer[position] != rune('a') {
					goto l260
				}
				position++
				if buffer[position] != rune('m') {
					goto l260
				}
				position++
				if buffer[position] != rune('b') {
					goto l260
				}
				position++
				if buffer[position] != rune('d') {
					goto l260
				}
				position++
				if buffer[position] != rune('a') {
					goto l260
				}
				position++
				{
					position262, tokenIndex262, depth262 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l263
					}
					goto l262
				l263:
					position, tokenIndex, depth = position262, tokenIndex262, depth262
					if !_rules[ruleLambdaExpr]() {
						goto l260
					}
				}
			l262:
				depth--
				add(ruleLambda, position261)
			}
			return true
		l260:
			position, tokenIndex, depth = position260, tokenIndex260, depth260
			return false
		},
		/* 68 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position264, tokenIndex264, depth264 := position, tokenIndex, depth
			{
				position265 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l264
				}
				if !_rules[ruleExpression]() {
					goto l264
				}
				depth--
				add(ruleLambdaRef, position265)
			}
			return true
		l264:
			position, tokenIndex, depth = position264, tokenIndex264, depth264
			return false
		},
		/* 69 LambdaExpr <- <(ws '|' ws Name NextName* ws '|' ws ('-' '>') Expression)> */
		func() bool {
			position266, tokenIndex266, depth266 := position, tokenIndex, depth
			{
				position267 := position
				depth++
				if !_rules[rulews]() {
					goto l266
				}
				if buffer[position] != rune('|') {
					goto l266
				}
				position++
				if !_rules[rulews]() {
					goto l266
				}
				if !_rules[ruleName]() {
					goto l266
				}
			l268:
				{
					position269, tokenIndex269, depth269 := position, tokenIndex, depth
					if !_rules[ruleNextName]() {
						goto l269
					}
					goto l268
				l269:
					position, tokenIndex, depth = position269, tokenIndex269, depth269
				}
				if !_rules[rulews]() {
					goto l266
				}
				if buffer[position] != rune('|') {
					goto l266
				}
				position++
				if !_rules[rulews]() {
					goto l266
				}
				if buffer[position] != rune('-') {
					goto l266
				}
				position++
				if buffer[position] != rune('>') {
					goto l266
				}
				position++
				if !_rules[ruleExpression]() {
					goto l266
				}
				depth--
				add(ruleLambdaExpr, position267)
			}
			return true
		l266:
			position, tokenIndex, depth = position266, tokenIndex266, depth266
			return false
		},
		/* 70 NextName <- <(ws ',' ws Name)> */
		func() bool {
			position270, tokenIndex270, depth270 := position, tokenIndex, depth
			{
				position271 := position
				depth++
				if !_rules[rulews]() {
					goto l270
				}
				if buffer[position] != rune(',') {
					goto l270
				}
				position++
				if !_rules[rulews]() {
					goto l270
				}
				if !_rules[ruleName]() {
					goto l270
				}
				depth--
				add(ruleNextName, position271)
			}
			return true
		l270:
			position, tokenIndex, depth = position270, tokenIndex270, depth270
			return false
		},
		/* 71 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position272, tokenIndex272, depth272 := position, tokenIndex, depth
			{
				position273 := position
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
						goto l272
					}
					position++
				}
			l276:
			l274:
				{
					position275, tokenIndex275, depth275 := position, tokenIndex, depth
					{
						position280, tokenIndex280, depth280 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l281
						}
						position++
						goto l280
					l281:
						position, tokenIndex, depth = position280, tokenIndex280, depth280
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l282
						}
						position++
						goto l280
					l282:
						position, tokenIndex, depth = position280, tokenIndex280, depth280
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l283
						}
						position++
						goto l280
					l283:
						position, tokenIndex, depth = position280, tokenIndex280, depth280
						if buffer[position] != rune('_') {
							goto l275
						}
						position++
					}
				l280:
					goto l274
				l275:
					position, tokenIndex, depth = position275, tokenIndex275, depth275
				}
				depth--
				add(ruleName, position273)
			}
			return true
		l272:
			position, tokenIndex, depth = position272, tokenIndex272, depth272
			return false
		},
		/* 72 Reference <- <('.'? Key FollowUpRef)> */
		func() bool {
			position284, tokenIndex284, depth284 := position, tokenIndex, depth
			{
				position285 := position
				depth++
				{
					position286, tokenIndex286, depth286 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l286
					}
					position++
					goto l287
				l286:
					position, tokenIndex, depth = position286, tokenIndex286, depth286
				}
			l287:
				if !_rules[ruleKey]() {
					goto l284
				}
				if !_rules[ruleFollowUpRef]() {
					goto l284
				}
				depth--
				add(ruleReference, position285)
			}
			return true
		l284:
			position, tokenIndex, depth = position284, tokenIndex284, depth284
			return false
		},
		/* 73 FollowUpRef <- <('.' (Key / Index))*> */
		func() bool {
			{
				position289 := position
				depth++
			l290:
				{
					position291, tokenIndex291, depth291 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l291
					}
					position++
					{
						position292, tokenIndex292, depth292 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l293
						}
						goto l292
					l293:
						position, tokenIndex, depth = position292, tokenIndex292, depth292
						if !_rules[ruleIndex]() {
							goto l291
						}
					}
				l292:
					goto l290
				l291:
					position, tokenIndex, depth = position291, tokenIndex291, depth291
				}
				depth--
				add(ruleFollowUpRef, position289)
			}
			return true
		},
		/* 74 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position294, tokenIndex294, depth294 := position, tokenIndex, depth
			{
				position295 := position
				depth++
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
						goto l294
					}
					position++
				}
			l296:
			l300:
				{
					position301, tokenIndex301, depth301 := position, tokenIndex, depth
					{
						position302, tokenIndex302, depth302 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l303
						}
						position++
						goto l302
					l303:
						position, tokenIndex, depth = position302, tokenIndex302, depth302
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l304
						}
						position++
						goto l302
					l304:
						position, tokenIndex, depth = position302, tokenIndex302, depth302
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l305
						}
						position++
						goto l302
					l305:
						position, tokenIndex, depth = position302, tokenIndex302, depth302
						if buffer[position] != rune('_') {
							goto l306
						}
						position++
						goto l302
					l306:
						position, tokenIndex, depth = position302, tokenIndex302, depth302
						if buffer[position] != rune('-') {
							goto l301
						}
						position++
					}
				l302:
					goto l300
				l301:
					position, tokenIndex, depth = position301, tokenIndex301, depth301
				}
				{
					position307, tokenIndex307, depth307 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l307
					}
					position++
					{
						position309, tokenIndex309, depth309 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l310
						}
						position++
						goto l309
					l310:
						position, tokenIndex, depth = position309, tokenIndex309, depth309
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l311
						}
						position++
						goto l309
					l311:
						position, tokenIndex, depth = position309, tokenIndex309, depth309
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l312
						}
						position++
						goto l309
					l312:
						position, tokenIndex, depth = position309, tokenIndex309, depth309
						if buffer[position] != rune('_') {
							goto l307
						}
						position++
					}
				l309:
				l313:
					{
						position314, tokenIndex314, depth314 := position, tokenIndex, depth
						{
							position315, tokenIndex315, depth315 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l316
							}
							position++
							goto l315
						l316:
							position, tokenIndex, depth = position315, tokenIndex315, depth315
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l317
							}
							position++
							goto l315
						l317:
							position, tokenIndex, depth = position315, tokenIndex315, depth315
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l318
							}
							position++
							goto l315
						l318:
							position, tokenIndex, depth = position315, tokenIndex315, depth315
							if buffer[position] != rune('_') {
								goto l319
							}
							position++
							goto l315
						l319:
							position, tokenIndex, depth = position315, tokenIndex315, depth315
							if buffer[position] != rune('-') {
								goto l314
							}
							position++
						}
					l315:
						goto l313
					l314:
						position, tokenIndex, depth = position314, tokenIndex314, depth314
					}
					goto l308
				l307:
					position, tokenIndex, depth = position307, tokenIndex307, depth307
				}
			l308:
				depth--
				add(ruleKey, position295)
			}
			return true
		l294:
			position, tokenIndex, depth = position294, tokenIndex294, depth294
			return false
		},
		/* 75 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position320, tokenIndex320, depth320 := position, tokenIndex, depth
			{
				position321 := position
				depth++
				if buffer[position] != rune('[') {
					goto l320
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l320
				}
				position++
			l322:
				{
					position323, tokenIndex323, depth323 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l323
					}
					position++
					goto l322
				l323:
					position, tokenIndex, depth = position323, tokenIndex323, depth323
				}
				if buffer[position] != rune(']') {
					goto l320
				}
				position++
				depth--
				add(ruleIndex, position321)
			}
			return true
		l320:
			position, tokenIndex, depth = position320, tokenIndex320, depth320
			return false
		},
		/* 76 IP <- <([0-9]+ '.' [0-9]+ '.' [0-9]+ '.' [0-9]+)> */
		func() bool {
			position324, tokenIndex324, depth324 := position, tokenIndex, depth
			{
				position325 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l324
				}
				position++
			l326:
				{
					position327, tokenIndex327, depth327 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l327
					}
					position++
					goto l326
				l327:
					position, tokenIndex, depth = position327, tokenIndex327, depth327
				}
				if buffer[position] != rune('.') {
					goto l324
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l324
				}
				position++
			l328:
				{
					position329, tokenIndex329, depth329 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l329
					}
					position++
					goto l328
				l329:
					position, tokenIndex, depth = position329, tokenIndex329, depth329
				}
				if buffer[position] != rune('.') {
					goto l324
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l324
				}
				position++
			l330:
				{
					position331, tokenIndex331, depth331 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l331
					}
					position++
					goto l330
				l331:
					position, tokenIndex, depth = position331, tokenIndex331, depth331
				}
				if buffer[position] != rune('.') {
					goto l324
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l324
				}
				position++
			l332:
				{
					position333, tokenIndex333, depth333 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l333
					}
					position++
					goto l332
				l333:
					position, tokenIndex, depth = position333, tokenIndex333, depth333
				}
				depth--
				add(ruleIP, position325)
			}
			return true
		l324:
			position, tokenIndex, depth = position324, tokenIndex324, depth324
			return false
		},
		/* 77 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position335 := position
				depth++
			l336:
				{
					position337, tokenIndex337, depth337 := position, tokenIndex, depth
					{
						position338, tokenIndex338, depth338 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l339
						}
						position++
						goto l338
					l339:
						position, tokenIndex, depth = position338, tokenIndex338, depth338
						if buffer[position] != rune('\t') {
							goto l340
						}
						position++
						goto l338
					l340:
						position, tokenIndex, depth = position338, tokenIndex338, depth338
						if buffer[position] != rune('\n') {
							goto l341
						}
						position++
						goto l338
					l341:
						position, tokenIndex, depth = position338, tokenIndex338, depth338
						if buffer[position] != rune('\r') {
							goto l337
						}
						position++
					}
				l338:
					goto l336
				l337:
					position, tokenIndex, depth = position337, tokenIndex337, depth337
				}
				depth--
				add(rulews, position335)
			}
			return true
		},
		/* 78 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position342, tokenIndex342, depth342 := position, tokenIndex, depth
			{
				position343 := position
				depth++
				{
					position346, tokenIndex346, depth346 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l347
					}
					position++
					goto l346
				l347:
					position, tokenIndex, depth = position346, tokenIndex346, depth346
					if buffer[position] != rune('\t') {
						goto l348
					}
					position++
					goto l346
				l348:
					position, tokenIndex, depth = position346, tokenIndex346, depth346
					if buffer[position] != rune('\n') {
						goto l349
					}
					position++
					goto l346
				l349:
					position, tokenIndex, depth = position346, tokenIndex346, depth346
					if buffer[position] != rune('\r') {
						goto l342
					}
					position++
				}
			l346:
			l344:
				{
					position345, tokenIndex345, depth345 := position, tokenIndex, depth
					{
						position350, tokenIndex350, depth350 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l351
						}
						position++
						goto l350
					l351:
						position, tokenIndex, depth = position350, tokenIndex350, depth350
						if buffer[position] != rune('\t') {
							goto l352
						}
						position++
						goto l350
					l352:
						position, tokenIndex, depth = position350, tokenIndex350, depth350
						if buffer[position] != rune('\n') {
							goto l353
						}
						position++
						goto l350
					l353:
						position, tokenIndex, depth = position350, tokenIndex350, depth350
						if buffer[position] != rune('\r') {
							goto l345
						}
						position++
					}
				l350:
					goto l344
				l345:
					position, tokenIndex, depth = position345, tokenIndex345, depth345
				}
				depth--
				add(rulereq_ws, position343)
			}
			return true
		l342:
			position, tokenIndex, depth = position342, tokenIndex342, depth342
			return false
		},
		/* 80 Action0 <- <{}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
	}
	p.rules = _rules
}
