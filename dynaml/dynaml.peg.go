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
	ruleStartRange
	ruleRangeOp
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
	ruleNone
	ruleOn
	ruleAuto
	ruleMapMapping
	ruleMapping
	ruleMapSelection
	ruleSelection
	ruleSum
	ruleLambda
	ruleLambdaRef
	ruleLambdaExpr
	ruleParams
	ruleStartParams
	ruleNames
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
	ruleAction1

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
	"StartRange",
	"RangeOp",
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
	"None",
	"On",
	"Auto",
	"MapMapping",
	"Mapping",
	"MapSelection",
	"Selection",
	"Sum",
	"Lambda",
	"LambdaRef",
	"LambdaExpr",
	"Params",
	"StartParams",
	"Names",
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
	"Action1",

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
	rules  [90]func() bool
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

		case ruleAction1:

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
		/* 7 Scoped <- <(Scope ws Expression)> */
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
				if !_rules[ruleExpression]() {
					goto l30
				}
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
			position32, tokenIndex32, depth32 := position, tokenIndex, depth
			{
				position33 := position
				depth++
				if !_rules[ruleCreateScope]() {
					goto l32
				}
				if !_rules[rulews]() {
					goto l32
				}
				{
					position34, tokenIndex34, depth34 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l34
					}
					goto l35
				l34:
					position, tokenIndex, depth = position34, tokenIndex34, depth34
				}
			l35:
				if buffer[position] != rune(')') {
					goto l32
				}
				position++
				depth--
				add(ruleScope, position33)
			}
			return true
		l32:
			position, tokenIndex, depth = position32, tokenIndex32, depth32
			return false
		},
		/* 9 CreateScope <- <'('> */
		func() bool {
			position36, tokenIndex36, depth36 := position, tokenIndex, depth
			{
				position37 := position
				depth++
				if buffer[position] != rune('(') {
					goto l36
				}
				position++
				depth--
				add(ruleCreateScope, position37)
			}
			return true
		l36:
			position, tokenIndex, depth = position36, tokenIndex36, depth36
			return false
		},
		/* 10 Level7 <- <(Level6 (req_ws Or)*)> */
		func() bool {
			position38, tokenIndex38, depth38 := position, tokenIndex, depth
			{
				position39 := position
				depth++
				if !_rules[ruleLevel6]() {
					goto l38
				}
			l40:
				{
					position41, tokenIndex41, depth41 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l41
					}
					if !_rules[ruleOr]() {
						goto l41
					}
					goto l40
				l41:
					position, tokenIndex, depth = position41, tokenIndex41, depth41
				}
				depth--
				add(ruleLevel7, position39)
			}
			return true
		l38:
			position, tokenIndex, depth = position38, tokenIndex38, depth38
			return false
		},
		/* 11 Or <- <('|' '|' req_ws Level6)> */
		func() bool {
			position42, tokenIndex42, depth42 := position, tokenIndex, depth
			{
				position43 := position
				depth++
				if buffer[position] != rune('|') {
					goto l42
				}
				position++
				if buffer[position] != rune('|') {
					goto l42
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l42
				}
				if !_rules[ruleLevel6]() {
					goto l42
				}
				depth--
				add(ruleOr, position43)
			}
			return true
		l42:
			position, tokenIndex, depth = position42, tokenIndex42, depth42
			return false
		},
		/* 12 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position44, tokenIndex44, depth44 := position, tokenIndex, depth
			{
				position45 := position
				depth++
				{
					position46, tokenIndex46, depth46 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l47
					}
					goto l46
				l47:
					position, tokenIndex, depth = position46, tokenIndex46, depth46
					if !_rules[ruleLevel5]() {
						goto l44
					}
				}
			l46:
				depth--
				add(ruleLevel6, position45)
			}
			return true
		l44:
			position, tokenIndex, depth = position44, tokenIndex44, depth44
			return false
		},
		/* 13 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position48, tokenIndex48, depth48 := position, tokenIndex, depth
			{
				position49 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l48
				}
				if !_rules[rulews]() {
					goto l48
				}
				if buffer[position] != rune('?') {
					goto l48
				}
				position++
				if !_rules[ruleExpression]() {
					goto l48
				}
				if buffer[position] != rune(':') {
					goto l48
				}
				position++
				if !_rules[ruleExpression]() {
					goto l48
				}
				depth--
				add(ruleConditional, position49)
			}
			return true
		l48:
			position, tokenIndex, depth = position48, tokenIndex48, depth48
			return false
		},
		/* 14 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position50, tokenIndex50, depth50 := position, tokenIndex, depth
			{
				position51 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l50
				}
			l52:
				{
					position53, tokenIndex53, depth53 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l53
					}
					goto l52
				l53:
					position, tokenIndex, depth = position53, tokenIndex53, depth53
				}
				depth--
				add(ruleLevel5, position51)
			}
			return true
		l50:
			position, tokenIndex, depth = position50, tokenIndex50, depth50
			return false
		},
		/* 15 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position54, tokenIndex54, depth54 := position, tokenIndex, depth
			{
				position55 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l54
				}
				if !_rules[ruleLevel4]() {
					goto l54
				}
				depth--
				add(ruleConcatenation, position55)
			}
			return true
		l54:
			position, tokenIndex, depth = position54, tokenIndex54, depth54
			return false
		},
		/* 16 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position56, tokenIndex56, depth56 := position, tokenIndex, depth
			{
				position57 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l56
				}
			l58:
				{
					position59, tokenIndex59, depth59 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l59
					}
					{
						position60, tokenIndex60, depth60 := position, tokenIndex, depth
						if !_rules[ruleLogOr]() {
							goto l61
						}
						goto l60
					l61:
						position, tokenIndex, depth = position60, tokenIndex60, depth60
						if !_rules[ruleLogAnd]() {
							goto l59
						}
					}
				l60:
					goto l58
				l59:
					position, tokenIndex, depth = position59, tokenIndex59, depth59
				}
				depth--
				add(ruleLevel4, position57)
			}
			return true
		l56:
			position, tokenIndex, depth = position56, tokenIndex56, depth56
			return false
		},
		/* 17 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position62, tokenIndex62, depth62 := position, tokenIndex, depth
			{
				position63 := position
				depth++
				if buffer[position] != rune('-') {
					goto l62
				}
				position++
				if buffer[position] != rune('o') {
					goto l62
				}
				position++
				if buffer[position] != rune('r') {
					goto l62
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l62
				}
				if !_rules[ruleLevel3]() {
					goto l62
				}
				depth--
				add(ruleLogOr, position63)
			}
			return true
		l62:
			position, tokenIndex, depth = position62, tokenIndex62, depth62
			return false
		},
		/* 18 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position64, tokenIndex64, depth64 := position, tokenIndex, depth
			{
				position65 := position
				depth++
				if buffer[position] != rune('-') {
					goto l64
				}
				position++
				if buffer[position] != rune('a') {
					goto l64
				}
				position++
				if buffer[position] != rune('n') {
					goto l64
				}
				position++
				if buffer[position] != rune('d') {
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
				add(ruleLogAnd, position65)
			}
			return true
		l64:
			position, tokenIndex, depth = position64, tokenIndex64, depth64
			return false
		},
		/* 19 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position66, tokenIndex66, depth66 := position, tokenIndex, depth
			{
				position67 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l66
				}
			l68:
				{
					position69, tokenIndex69, depth69 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l69
					}
					if !_rules[ruleComparison]() {
						goto l69
					}
					goto l68
				l69:
					position, tokenIndex, depth = position69, tokenIndex69, depth69
				}
				depth--
				add(ruleLevel3, position67)
			}
			return true
		l66:
			position, tokenIndex, depth = position66, tokenIndex66, depth66
			return false
		},
		/* 20 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position70, tokenIndex70, depth70 := position, tokenIndex, depth
			{
				position71 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l70
				}
				if !_rules[rulereq_ws]() {
					goto l70
				}
				if !_rules[ruleLevel2]() {
					goto l70
				}
				depth--
				add(ruleComparison, position71)
			}
			return true
		l70:
			position, tokenIndex, depth = position70, tokenIndex70, depth70
			return false
		},
		/* 21 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position72, tokenIndex72, depth72 := position, tokenIndex, depth
			{
				position73 := position
				depth++
				{
					position74, tokenIndex74, depth74 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l75
					}
					position++
					if buffer[position] != rune('=') {
						goto l75
					}
					position++
					goto l74
				l75:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if buffer[position] != rune('!') {
						goto l76
					}
					position++
					if buffer[position] != rune('=') {
						goto l76
					}
					position++
					goto l74
				l76:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if buffer[position] != rune('<') {
						goto l77
					}
					position++
					if buffer[position] != rune('=') {
						goto l77
					}
					position++
					goto l74
				l77:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if buffer[position] != rune('>') {
						goto l78
					}
					position++
					if buffer[position] != rune('=') {
						goto l78
					}
					position++
					goto l74
				l78:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if buffer[position] != rune('>') {
						goto l79
					}
					position++
					goto l74
				l79:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if buffer[position] != rune('<') {
						goto l80
					}
					position++
					goto l74
				l80:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
					if buffer[position] != rune('>') {
						goto l72
					}
					position++
				}
			l74:
				depth--
				add(ruleCompareOp, position73)
			}
			return true
		l72:
			position, tokenIndex, depth = position72, tokenIndex72, depth72
			return false
		},
		/* 22 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position81, tokenIndex81, depth81 := position, tokenIndex, depth
			{
				position82 := position
				depth++
				if !_rules[ruleLevel1]() {
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
						if !_rules[ruleAddition]() {
							goto l86
						}
						goto l85
					l86:
						position, tokenIndex, depth = position85, tokenIndex85, depth85
						if !_rules[ruleSubtraction]() {
							goto l84
						}
					}
				l85:
					goto l83
				l84:
					position, tokenIndex, depth = position84, tokenIndex84, depth84
				}
				depth--
				add(ruleLevel2, position82)
			}
			return true
		l81:
			position, tokenIndex, depth = position81, tokenIndex81, depth81
			return false
		},
		/* 23 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position87, tokenIndex87, depth87 := position, tokenIndex, depth
			{
				position88 := position
				depth++
				if buffer[position] != rune('+') {
					goto l87
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l87
				}
				if !_rules[ruleLevel1]() {
					goto l87
				}
				depth--
				add(ruleAddition, position88)
			}
			return true
		l87:
			position, tokenIndex, depth = position87, tokenIndex87, depth87
			return false
		},
		/* 24 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position89, tokenIndex89, depth89 := position, tokenIndex, depth
			{
				position90 := position
				depth++
				if buffer[position] != rune('-') {
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
				add(ruleSubtraction, position90)
			}
			return true
		l89:
			position, tokenIndex, depth = position89, tokenIndex89, depth89
			return false
		},
		/* 25 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position91, tokenIndex91, depth91 := position, tokenIndex, depth
			{
				position92 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l91
				}
			l93:
				{
					position94, tokenIndex94, depth94 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l94
					}
					{
						position95, tokenIndex95, depth95 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l96
						}
						goto l95
					l96:
						position, tokenIndex, depth = position95, tokenIndex95, depth95
						if !_rules[ruleDivision]() {
							goto l97
						}
						goto l95
					l97:
						position, tokenIndex, depth = position95, tokenIndex95, depth95
						if !_rules[ruleModulo]() {
							goto l94
						}
					}
				l95:
					goto l93
				l94:
					position, tokenIndex, depth = position94, tokenIndex94, depth94
				}
				depth--
				add(ruleLevel1, position92)
			}
			return true
		l91:
			position, tokenIndex, depth = position91, tokenIndex91, depth91
			return false
		},
		/* 26 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position98, tokenIndex98, depth98 := position, tokenIndex, depth
			{
				position99 := position
				depth++
				if buffer[position] != rune('*') {
					goto l98
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l98
				}
				if !_rules[ruleLevel0]() {
					goto l98
				}
				depth--
				add(ruleMultiplication, position99)
			}
			return true
		l98:
			position, tokenIndex, depth = position98, tokenIndex98, depth98
			return false
		},
		/* 27 Division <- <('/' req_ws Level0)> */
		func() bool {
			position100, tokenIndex100, depth100 := position, tokenIndex, depth
			{
				position101 := position
				depth++
				if buffer[position] != rune('/') {
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
				add(ruleDivision, position101)
			}
			return true
		l100:
			position, tokenIndex, depth = position100, tokenIndex100, depth100
			return false
		},
		/* 28 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position102, tokenIndex102, depth102 := position, tokenIndex, depth
			{
				position103 := position
				depth++
				if buffer[position] != rune('%') {
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
				add(ruleModulo, position103)
			}
			return true
		l102:
			position, tokenIndex, depth = position102, tokenIndex102, depth102
			return false
		},
		/* 29 Level0 <- <(IP / String / Integer / Boolean / Undefined / Nil / Symbol / Not / Substitution / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position104, tokenIndex104, depth104 := position, tokenIndex, depth
			{
				position105 := position
				depth++
				{
					position106, tokenIndex106, depth106 := position, tokenIndex, depth
					if !_rules[ruleIP]() {
						goto l107
					}
					goto l106
				l107:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleString]() {
						goto l108
					}
					goto l106
				l108:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleInteger]() {
						goto l109
					}
					goto l106
				l109:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleBoolean]() {
						goto l110
					}
					goto l106
				l110:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleUndefined]() {
						goto l111
					}
					goto l106
				l111:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleNil]() {
						goto l112
					}
					goto l106
				l112:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleSymbol]() {
						goto l113
					}
					goto l106
				l113:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleNot]() {
						goto l114
					}
					goto l106
				l114:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleSubstitution]() {
						goto l115
					}
					goto l106
				l115:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleMerge]() {
						goto l116
					}
					goto l106
				l116:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleAuto]() {
						goto l117
					}
					goto l106
				l117:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleLambda]() {
						goto l118
					}
					goto l106
				l118:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleChained]() {
						goto l104
					}
				}
			l106:
				depth--
				add(ruleLevel0, position105)
			}
			return true
		l104:
			position, tokenIndex, depth = position104, tokenIndex104, depth104
			return false
		},
		/* 30 Chained <- <((MapMapping / Mapping / MapSelection / Selection / Sum / List / Map / Range / Grouped / Reference) ChainedQualifiedExpression*)> */
		func() bool {
			position119, tokenIndex119, depth119 := position, tokenIndex, depth
			{
				position120 := position
				depth++
				{
					position121, tokenIndex121, depth121 := position, tokenIndex, depth
					if !_rules[ruleMapMapping]() {
						goto l122
					}
					goto l121
				l122:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
					if !_rules[ruleMapping]() {
						goto l123
					}
					goto l121
				l123:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
					if !_rules[ruleMapSelection]() {
						goto l124
					}
					goto l121
				l124:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
					if !_rules[ruleSelection]() {
						goto l125
					}
					goto l121
				l125:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
					if !_rules[ruleSum]() {
						goto l126
					}
					goto l121
				l126:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
					if !_rules[ruleList]() {
						goto l127
					}
					goto l121
				l127:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
					if !_rules[ruleMap]() {
						goto l128
					}
					goto l121
				l128:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
					if !_rules[ruleRange]() {
						goto l129
					}
					goto l121
				l129:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
					if !_rules[ruleGrouped]() {
						goto l130
					}
					goto l121
				l130:
					position, tokenIndex, depth = position121, tokenIndex121, depth121
					if !_rules[ruleReference]() {
						goto l119
					}
				}
			l121:
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
				add(ruleChained, position120)
			}
			return true
		l119:
			position, tokenIndex, depth = position119, tokenIndex119, depth119
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
		/* 44 Range <- <(StartRange Expression? RangeOp Expression? ']')> */
		func() bool {
			position174, tokenIndex174, depth174 := position, tokenIndex, depth
			{
				position175 := position
				depth++
				if !_rules[ruleStartRange]() {
					goto l174
				}
				{
					position176, tokenIndex176, depth176 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l176
					}
					goto l177
				l176:
					position, tokenIndex, depth = position176, tokenIndex176, depth176
				}
			l177:
				if !_rules[ruleRangeOp]() {
					goto l174
				}
				{
					position178, tokenIndex178, depth178 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l178
					}
					goto l179
				l178:
					position, tokenIndex, depth = position178, tokenIndex178, depth178
				}
			l179:
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
		/* 45 StartRange <- <'['> */
		func() bool {
			position180, tokenIndex180, depth180 := position, tokenIndex, depth
			{
				position181 := position
				depth++
				if buffer[position] != rune('[') {
					goto l180
				}
				position++
				depth--
				add(ruleStartRange, position181)
			}
			return true
		l180:
			position, tokenIndex, depth = position180, tokenIndex180, depth180
			return false
		},
		/* 46 RangeOp <- <('.' '.')> */
		func() bool {
			position182, tokenIndex182, depth182 := position, tokenIndex, depth
			{
				position183 := position
				depth++
				if buffer[position] != rune('.') {
					goto l182
				}
				position++
				if buffer[position] != rune('.') {
					goto l182
				}
				position++
				depth--
				add(ruleRangeOp, position183)
			}
			return true
		l182:
			position, tokenIndex, depth = position182, tokenIndex182, depth182
			return false
		},
		/* 47 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position184, tokenIndex184, depth184 := position, tokenIndex, depth
			{
				position185 := position
				depth++
				{
					position186, tokenIndex186, depth186 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l186
					}
					position++
					goto l187
				l186:
					position, tokenIndex, depth = position186, tokenIndex186, depth186
				}
			l187:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l184
				}
				position++
			l188:
				{
					position189, tokenIndex189, depth189 := position, tokenIndex, depth
					{
						position190, tokenIndex190, depth190 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l191
						}
						position++
						goto l190
					l191:
						position, tokenIndex, depth = position190, tokenIndex190, depth190
						if buffer[position] != rune('_') {
							goto l189
						}
						position++
					}
				l190:
					goto l188
				l189:
					position, tokenIndex, depth = position189, tokenIndex189, depth189
				}
				depth--
				add(ruleInteger, position185)
			}
			return true
		l184:
			position, tokenIndex, depth = position184, tokenIndex184, depth184
			return false
		},
		/* 48 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position192, tokenIndex192, depth192 := position, tokenIndex, depth
			{
				position193 := position
				depth++
				if buffer[position] != rune('"') {
					goto l192
				}
				position++
			l194:
				{
					position195, tokenIndex195, depth195 := position, tokenIndex, depth
					{
						position196, tokenIndex196, depth196 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l197
						}
						position++
						if buffer[position] != rune('"') {
							goto l197
						}
						position++
						goto l196
					l197:
						position, tokenIndex, depth = position196, tokenIndex196, depth196
						{
							position198, tokenIndex198, depth198 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l198
							}
							position++
							goto l195
						l198:
							position, tokenIndex, depth = position198, tokenIndex198, depth198
						}
						if !matchDot() {
							goto l195
						}
					}
				l196:
					goto l194
				l195:
					position, tokenIndex, depth = position195, tokenIndex195, depth195
				}
				if buffer[position] != rune('"') {
					goto l192
				}
				position++
				depth--
				add(ruleString, position193)
			}
			return true
		l192:
			position, tokenIndex, depth = position192, tokenIndex192, depth192
			return false
		},
		/* 49 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position199, tokenIndex199, depth199 := position, tokenIndex, depth
			{
				position200 := position
				depth++
				{
					position201, tokenIndex201, depth201 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l202
					}
					position++
					if buffer[position] != rune('r') {
						goto l202
					}
					position++
					if buffer[position] != rune('u') {
						goto l202
					}
					position++
					if buffer[position] != rune('e') {
						goto l202
					}
					position++
					goto l201
				l202:
					position, tokenIndex, depth = position201, tokenIndex201, depth201
					if buffer[position] != rune('f') {
						goto l199
					}
					position++
					if buffer[position] != rune('a') {
						goto l199
					}
					position++
					if buffer[position] != rune('l') {
						goto l199
					}
					position++
					if buffer[position] != rune('s') {
						goto l199
					}
					position++
					if buffer[position] != rune('e') {
						goto l199
					}
					position++
				}
			l201:
				depth--
				add(ruleBoolean, position200)
			}
			return true
		l199:
			position, tokenIndex, depth = position199, tokenIndex199, depth199
			return false
		},
		/* 50 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position203, tokenIndex203, depth203 := position, tokenIndex, depth
			{
				position204 := position
				depth++
				{
					position205, tokenIndex205, depth205 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l206
					}
					position++
					if buffer[position] != rune('i') {
						goto l206
					}
					position++
					if buffer[position] != rune('l') {
						goto l206
					}
					position++
					goto l205
				l206:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
					if buffer[position] != rune('~') {
						goto l203
					}
					position++
				}
			l205:
				depth--
				add(ruleNil, position204)
			}
			return true
		l203:
			position, tokenIndex, depth = position203, tokenIndex203, depth203
			return false
		},
		/* 51 Undefined <- <('~' '~')> */
		func() bool {
			position207, tokenIndex207, depth207 := position, tokenIndex, depth
			{
				position208 := position
				depth++
				if buffer[position] != rune('~') {
					goto l207
				}
				position++
				if buffer[position] != rune('~') {
					goto l207
				}
				position++
				depth--
				add(ruleUndefined, position208)
			}
			return true
		l207:
			position, tokenIndex, depth = position207, tokenIndex207, depth207
			return false
		},
		/* 52 Symbol <- <('$' Name)> */
		func() bool {
			position209, tokenIndex209, depth209 := position, tokenIndex, depth
			{
				position210 := position
				depth++
				if buffer[position] != rune('$') {
					goto l209
				}
				position++
				if !_rules[ruleName]() {
					goto l209
				}
				depth--
				add(ruleSymbol, position210)
			}
			return true
		l209:
			position, tokenIndex, depth = position209, tokenIndex209, depth209
			return false
		},
		/* 53 List <- <(StartList ExpressionList? ']')> */
		func() bool {
			position211, tokenIndex211, depth211 := position, tokenIndex, depth
			{
				position212 := position
				depth++
				if !_rules[ruleStartList]() {
					goto l211
				}
				{
					position213, tokenIndex213, depth213 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l213
					}
					goto l214
				l213:
					position, tokenIndex, depth = position213, tokenIndex213, depth213
				}
			l214:
				if buffer[position] != rune(']') {
					goto l211
				}
				position++
				depth--
				add(ruleList, position212)
			}
			return true
		l211:
			position, tokenIndex, depth = position211, tokenIndex211, depth211
			return false
		},
		/* 54 StartList <- <('[' ws)> */
		func() bool {
			position215, tokenIndex215, depth215 := position, tokenIndex, depth
			{
				position216 := position
				depth++
				if buffer[position] != rune('[') {
					goto l215
				}
				position++
				if !_rules[rulews]() {
					goto l215
				}
				depth--
				add(ruleStartList, position216)
			}
			return true
		l215:
			position, tokenIndex, depth = position215, tokenIndex215, depth215
			return false
		},
		/* 55 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position217, tokenIndex217, depth217 := position, tokenIndex, depth
			{
				position218 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l217
				}
				if !_rules[rulews]() {
					goto l217
				}
				{
					position219, tokenIndex219, depth219 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l219
					}
					goto l220
				l219:
					position, tokenIndex, depth = position219, tokenIndex219, depth219
				}
			l220:
				if buffer[position] != rune('}') {
					goto l217
				}
				position++
				depth--
				add(ruleMap, position218)
			}
			return true
		l217:
			position, tokenIndex, depth = position217, tokenIndex217, depth217
			return false
		},
		/* 56 CreateMap <- <'{'> */
		func() bool {
			position221, tokenIndex221, depth221 := position, tokenIndex, depth
			{
				position222 := position
				depth++
				if buffer[position] != rune('{') {
					goto l221
				}
				position++
				depth--
				add(ruleCreateMap, position222)
			}
			return true
		l221:
			position, tokenIndex, depth = position221, tokenIndex221, depth221
			return false
		},
		/* 57 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position223, tokenIndex223, depth223 := position, tokenIndex, depth
			{
				position224 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l223
				}
			l225:
				{
					position226, tokenIndex226, depth226 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l226
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l226
					}
					goto l225
				l226:
					position, tokenIndex, depth = position226, tokenIndex226, depth226
				}
				depth--
				add(ruleAssignments, position224)
			}
			return true
		l223:
			position, tokenIndex, depth = position223, tokenIndex223, depth223
			return false
		},
		/* 58 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position227, tokenIndex227, depth227 := position, tokenIndex, depth
			{
				position228 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l227
				}
				if buffer[position] != rune('=') {
					goto l227
				}
				position++
				if !_rules[ruleExpression]() {
					goto l227
				}
				depth--
				add(ruleAssignment, position228)
			}
			return true
		l227:
			position, tokenIndex, depth = position227, tokenIndex227, depth227
			return false
		},
		/* 59 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position229, tokenIndex229, depth229 := position, tokenIndex, depth
			{
				position230 := position
				depth++
				{
					position231, tokenIndex231, depth231 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l232
					}
					goto l231
				l232:
					position, tokenIndex, depth = position231, tokenIndex231, depth231
					if !_rules[ruleSimpleMerge]() {
						goto l229
					}
				}
			l231:
				depth--
				add(ruleMerge, position230)
			}
			return true
		l229:
			position, tokenIndex, depth = position229, tokenIndex229, depth229
			return false
		},
		/* 60 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws (Required / None)) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position233, tokenIndex233, depth233 := position, tokenIndex, depth
			{
				position234 := position
				depth++
				if buffer[position] != rune('m') {
					goto l233
				}
				position++
				if buffer[position] != rune('e') {
					goto l233
				}
				position++
				if buffer[position] != rune('r') {
					goto l233
				}
				position++
				if buffer[position] != rune('g') {
					goto l233
				}
				position++
				if buffer[position] != rune('e') {
					goto l233
				}
				position++
				{
					position235, tokenIndex235, depth235 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l235
					}
					{
						position236, tokenIndex236, depth236 := position, tokenIndex, depth
						if !_rules[ruleRequired]() {
							goto l237
						}
						goto l236
					l237:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						if !_rules[ruleNone]() {
							goto l235
						}
					}
				l236:
					goto l233
				l235:
					position, tokenIndex, depth = position235, tokenIndex235, depth235
				}
				{
					position238, tokenIndex238, depth238 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l238
					}
					{
						position240, tokenIndex240, depth240 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l241
						}
						goto l240
					l241:
						position, tokenIndex, depth = position240, tokenIndex240, depth240
						if !_rules[ruleOn]() {
							goto l238
						}
					}
				l240:
					goto l239
				l238:
					position, tokenIndex, depth = position238, tokenIndex238, depth238
				}
			l239:
				if !_rules[rulereq_ws]() {
					goto l233
				}
				if !_rules[ruleReference]() {
					goto l233
				}
				depth--
				add(ruleRefMerge, position234)
			}
			return true
		l233:
			position, tokenIndex, depth = position233, tokenIndex233, depth233
			return false
		},
		/* 61 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' !'(' (req_ws (Replace / Required / On / None))?)> */
		func() bool {
			position242, tokenIndex242, depth242 := position, tokenIndex, depth
			{
				position243 := position
				depth++
				if buffer[position] != rune('m') {
					goto l242
				}
				position++
				if buffer[position] != rune('e') {
					goto l242
				}
				position++
				if buffer[position] != rune('r') {
					goto l242
				}
				position++
				if buffer[position] != rune('g') {
					goto l242
				}
				position++
				if buffer[position] != rune('e') {
					goto l242
				}
				position++
				{
					position244, tokenIndex244, depth244 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l244
					}
					position++
					goto l242
				l244:
					position, tokenIndex, depth = position244, tokenIndex244, depth244
				}
				{
					position245, tokenIndex245, depth245 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l245
					}
					{
						position247, tokenIndex247, depth247 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l248
						}
						goto l247
					l248:
						position, tokenIndex, depth = position247, tokenIndex247, depth247
						if !_rules[ruleRequired]() {
							goto l249
						}
						goto l247
					l249:
						position, tokenIndex, depth = position247, tokenIndex247, depth247
						if !_rules[ruleOn]() {
							goto l250
						}
						goto l247
					l250:
						position, tokenIndex, depth = position247, tokenIndex247, depth247
						if !_rules[ruleNone]() {
							goto l245
						}
					}
				l247:
					goto l246
				l245:
					position, tokenIndex, depth = position245, tokenIndex245, depth245
				}
			l246:
				depth--
				add(ruleSimpleMerge, position243)
			}
			return true
		l242:
			position, tokenIndex, depth = position242, tokenIndex242, depth242
			return false
		},
		/* 62 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position251, tokenIndex251, depth251 := position, tokenIndex, depth
			{
				position252 := position
				depth++
				if buffer[position] != rune('r') {
					goto l251
				}
				position++
				if buffer[position] != rune('e') {
					goto l251
				}
				position++
				if buffer[position] != rune('p') {
					goto l251
				}
				position++
				if buffer[position] != rune('l') {
					goto l251
				}
				position++
				if buffer[position] != rune('a') {
					goto l251
				}
				position++
				if buffer[position] != rune('c') {
					goto l251
				}
				position++
				if buffer[position] != rune('e') {
					goto l251
				}
				position++
				depth--
				add(ruleReplace, position252)
			}
			return true
		l251:
			position, tokenIndex, depth = position251, tokenIndex251, depth251
			return false
		},
		/* 63 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position253, tokenIndex253, depth253 := position, tokenIndex, depth
			{
				position254 := position
				depth++
				if buffer[position] != rune('r') {
					goto l253
				}
				position++
				if buffer[position] != rune('e') {
					goto l253
				}
				position++
				if buffer[position] != rune('q') {
					goto l253
				}
				position++
				if buffer[position] != rune('u') {
					goto l253
				}
				position++
				if buffer[position] != rune('i') {
					goto l253
				}
				position++
				if buffer[position] != rune('r') {
					goto l253
				}
				position++
				if buffer[position] != rune('e') {
					goto l253
				}
				position++
				if buffer[position] != rune('d') {
					goto l253
				}
				position++
				depth--
				add(ruleRequired, position254)
			}
			return true
		l253:
			position, tokenIndex, depth = position253, tokenIndex253, depth253
			return false
		},
		/* 64 None <- <('n' 'o' 'n' 'e')> */
		func() bool {
			position255, tokenIndex255, depth255 := position, tokenIndex, depth
			{
				position256 := position
				depth++
				if buffer[position] != rune('n') {
					goto l255
				}
				position++
				if buffer[position] != rune('o') {
					goto l255
				}
				position++
				if buffer[position] != rune('n') {
					goto l255
				}
				position++
				if buffer[position] != rune('e') {
					goto l255
				}
				position++
				depth--
				add(ruleNone, position256)
			}
			return true
		l255:
			position, tokenIndex, depth = position255, tokenIndex255, depth255
			return false
		},
		/* 65 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position257, tokenIndex257, depth257 := position, tokenIndex, depth
			{
				position258 := position
				depth++
				if buffer[position] != rune('o') {
					goto l257
				}
				position++
				if buffer[position] != rune('n') {
					goto l257
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l257
				}
				if !_rules[ruleName]() {
					goto l257
				}
				depth--
				add(ruleOn, position258)
			}
			return true
		l257:
			position, tokenIndex, depth = position257, tokenIndex257, depth257
			return false
		},
		/* 66 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position259, tokenIndex259, depth259 := position, tokenIndex, depth
			{
				position260 := position
				depth++
				if buffer[position] != rune('a') {
					goto l259
				}
				position++
				if buffer[position] != rune('u') {
					goto l259
				}
				position++
				if buffer[position] != rune('t') {
					goto l259
				}
				position++
				if buffer[position] != rune('o') {
					goto l259
				}
				position++
				depth--
				add(ruleAuto, position260)
			}
			return true
		l259:
			position, tokenIndex, depth = position259, tokenIndex259, depth259
			return false
		},
		/* 67 MapMapping <- <('m' 'a' 'p' '{' Level7 (LambdaExpr / ('|' Expression)) '}')> */
		func() bool {
			position261, tokenIndex261, depth261 := position, tokenIndex, depth
			{
				position262 := position
				depth++
				if buffer[position] != rune('m') {
					goto l261
				}
				position++
				if buffer[position] != rune('a') {
					goto l261
				}
				position++
				if buffer[position] != rune('p') {
					goto l261
				}
				position++
				if buffer[position] != rune('{') {
					goto l261
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l261
				}
				{
					position263, tokenIndex263, depth263 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l264
					}
					goto l263
				l264:
					position, tokenIndex, depth = position263, tokenIndex263, depth263
					if buffer[position] != rune('|') {
						goto l261
					}
					position++
					if !_rules[ruleExpression]() {
						goto l261
					}
				}
			l263:
				if buffer[position] != rune('}') {
					goto l261
				}
				position++
				depth--
				add(ruleMapMapping, position262)
			}
			return true
		l261:
			position, tokenIndex, depth = position261, tokenIndex261, depth261
			return false
		},
		/* 68 Mapping <- <('m' 'a' 'p' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position265, tokenIndex265, depth265 := position, tokenIndex, depth
			{
				position266 := position
				depth++
				if buffer[position] != rune('m') {
					goto l265
				}
				position++
				if buffer[position] != rune('a') {
					goto l265
				}
				position++
				if buffer[position] != rune('p') {
					goto l265
				}
				position++
				if buffer[position] != rune('[') {
					goto l265
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l265
				}
				{
					position267, tokenIndex267, depth267 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l268
					}
					goto l267
				l268:
					position, tokenIndex, depth = position267, tokenIndex267, depth267
					if buffer[position] != rune('|') {
						goto l265
					}
					position++
					if !_rules[ruleExpression]() {
						goto l265
					}
				}
			l267:
				if buffer[position] != rune(']') {
					goto l265
				}
				position++
				depth--
				add(ruleMapping, position266)
			}
			return true
		l265:
			position, tokenIndex, depth = position265, tokenIndex265, depth265
			return false
		},
		/* 69 MapSelection <- <('s' 'e' 'l' 'e' 'c' 't' '{' Level7 (LambdaExpr / ('|' Expression)) '}')> */
		func() bool {
			position269, tokenIndex269, depth269 := position, tokenIndex, depth
			{
				position270 := position
				depth++
				if buffer[position] != rune('s') {
					goto l269
				}
				position++
				if buffer[position] != rune('e') {
					goto l269
				}
				position++
				if buffer[position] != rune('l') {
					goto l269
				}
				position++
				if buffer[position] != rune('e') {
					goto l269
				}
				position++
				if buffer[position] != rune('c') {
					goto l269
				}
				position++
				if buffer[position] != rune('t') {
					goto l269
				}
				position++
				if buffer[position] != rune('{') {
					goto l269
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l269
				}
				{
					position271, tokenIndex271, depth271 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l272
					}
					goto l271
				l272:
					position, tokenIndex, depth = position271, tokenIndex271, depth271
					if buffer[position] != rune('|') {
						goto l269
					}
					position++
					if !_rules[ruleExpression]() {
						goto l269
					}
				}
			l271:
				if buffer[position] != rune('}') {
					goto l269
				}
				position++
				depth--
				add(ruleMapSelection, position270)
			}
			return true
		l269:
			position, tokenIndex, depth = position269, tokenIndex269, depth269
			return false
		},
		/* 70 Selection <- <('s' 'e' 'l' 'e' 'c' 't' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position273, tokenIndex273, depth273 := position, tokenIndex, depth
			{
				position274 := position
				depth++
				if buffer[position] != rune('s') {
					goto l273
				}
				position++
				if buffer[position] != rune('e') {
					goto l273
				}
				position++
				if buffer[position] != rune('l') {
					goto l273
				}
				position++
				if buffer[position] != rune('e') {
					goto l273
				}
				position++
				if buffer[position] != rune('c') {
					goto l273
				}
				position++
				if buffer[position] != rune('t') {
					goto l273
				}
				position++
				if buffer[position] != rune('[') {
					goto l273
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l273
				}
				{
					position275, tokenIndex275, depth275 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l276
					}
					goto l275
				l276:
					position, tokenIndex, depth = position275, tokenIndex275, depth275
					if buffer[position] != rune('|') {
						goto l273
					}
					position++
					if !_rules[ruleExpression]() {
						goto l273
					}
				}
			l275:
				if buffer[position] != rune(']') {
					goto l273
				}
				position++
				depth--
				add(ruleSelection, position274)
			}
			return true
		l273:
			position, tokenIndex, depth = position273, tokenIndex273, depth273
			return false
		},
		/* 71 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position277, tokenIndex277, depth277 := position, tokenIndex, depth
			{
				position278 := position
				depth++
				if buffer[position] != rune('s') {
					goto l277
				}
				position++
				if buffer[position] != rune('u') {
					goto l277
				}
				position++
				if buffer[position] != rune('m') {
					goto l277
				}
				position++
				if buffer[position] != rune('[') {
					goto l277
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l277
				}
				if buffer[position] != rune('|') {
					goto l277
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l277
				}
				{
					position279, tokenIndex279, depth279 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l280
					}
					goto l279
				l280:
					position, tokenIndex, depth = position279, tokenIndex279, depth279
					if buffer[position] != rune('|') {
						goto l277
					}
					position++
					if !_rules[ruleExpression]() {
						goto l277
					}
				}
			l279:
				if buffer[position] != rune(']') {
					goto l277
				}
				position++
				depth--
				add(ruleSum, position278)
			}
			return true
		l277:
			position, tokenIndex, depth = position277, tokenIndex277, depth277
			return false
		},
		/* 72 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position281, tokenIndex281, depth281 := position, tokenIndex, depth
			{
				position282 := position
				depth++
				if buffer[position] != rune('l') {
					goto l281
				}
				position++
				if buffer[position] != rune('a') {
					goto l281
				}
				position++
				if buffer[position] != rune('m') {
					goto l281
				}
				position++
				if buffer[position] != rune('b') {
					goto l281
				}
				position++
				if buffer[position] != rune('d') {
					goto l281
				}
				position++
				if buffer[position] != rune('a') {
					goto l281
				}
				position++
				{
					position283, tokenIndex283, depth283 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l284
					}
					goto l283
				l284:
					position, tokenIndex, depth = position283, tokenIndex283, depth283
					if !_rules[ruleLambdaExpr]() {
						goto l281
					}
				}
			l283:
				depth--
				add(ruleLambda, position282)
			}
			return true
		l281:
			position, tokenIndex, depth = position281, tokenIndex281, depth281
			return false
		},
		/* 73 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position285, tokenIndex285, depth285 := position, tokenIndex, depth
			{
				position286 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l285
				}
				if !_rules[ruleExpression]() {
					goto l285
				}
				depth--
				add(ruleLambdaRef, position286)
			}
			return true
		l285:
			position, tokenIndex, depth = position285, tokenIndex285, depth285
			return false
		},
		/* 74 LambdaExpr <- <(ws Params ws ('-' '>') Expression)> */
		func() bool {
			position287, tokenIndex287, depth287 := position, tokenIndex, depth
			{
				position288 := position
				depth++
				if !_rules[rulews]() {
					goto l287
				}
				if !_rules[ruleParams]() {
					goto l287
				}
				if !_rules[rulews]() {
					goto l287
				}
				if buffer[position] != rune('-') {
					goto l287
				}
				position++
				if buffer[position] != rune('>') {
					goto l287
				}
				position++
				if !_rules[ruleExpression]() {
					goto l287
				}
				depth--
				add(ruleLambdaExpr, position288)
			}
			return true
		l287:
			position, tokenIndex, depth = position287, tokenIndex287, depth287
			return false
		},
		/* 75 Params <- <('|' StartParams ws Names? ws '|')> */
		func() bool {
			position289, tokenIndex289, depth289 := position, tokenIndex, depth
			{
				position290 := position
				depth++
				if buffer[position] != rune('|') {
					goto l289
				}
				position++
				if !_rules[ruleStartParams]() {
					goto l289
				}
				if !_rules[rulews]() {
					goto l289
				}
				{
					position291, tokenIndex291, depth291 := position, tokenIndex, depth
					if !_rules[ruleNames]() {
						goto l291
					}
					goto l292
				l291:
					position, tokenIndex, depth = position291, tokenIndex291, depth291
				}
			l292:
				if !_rules[rulews]() {
					goto l289
				}
				if buffer[position] != rune('|') {
					goto l289
				}
				position++
				depth--
				add(ruleParams, position290)
			}
			return true
		l289:
			position, tokenIndex, depth = position289, tokenIndex289, depth289
			return false
		},
		/* 76 StartParams <- <Action1> */
		func() bool {
			position293, tokenIndex293, depth293 := position, tokenIndex, depth
			{
				position294 := position
				depth++
				if !_rules[ruleAction1]() {
					goto l293
				}
				depth--
				add(ruleStartParams, position294)
			}
			return true
		l293:
			position, tokenIndex, depth = position293, tokenIndex293, depth293
			return false
		},
		/* 77 Names <- <(NextName (',' NextName)*)> */
		func() bool {
			position295, tokenIndex295, depth295 := position, tokenIndex, depth
			{
				position296 := position
				depth++
				if !_rules[ruleNextName]() {
					goto l295
				}
			l297:
				{
					position298, tokenIndex298, depth298 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l298
					}
					position++
					if !_rules[ruleNextName]() {
						goto l298
					}
					goto l297
				l298:
					position, tokenIndex, depth = position298, tokenIndex298, depth298
				}
				depth--
				add(ruleNames, position296)
			}
			return true
		l295:
			position, tokenIndex, depth = position295, tokenIndex295, depth295
			return false
		},
		/* 78 NextName <- <(ws Name ws)> */
		func() bool {
			position299, tokenIndex299, depth299 := position, tokenIndex, depth
			{
				position300 := position
				depth++
				if !_rules[rulews]() {
					goto l299
				}
				if !_rules[ruleName]() {
					goto l299
				}
				if !_rules[rulews]() {
					goto l299
				}
				depth--
				add(ruleNextName, position300)
			}
			return true
		l299:
			position, tokenIndex, depth = position299, tokenIndex299, depth299
			return false
		},
		/* 79 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position301, tokenIndex301, depth301 := position, tokenIndex, depth
			{
				position302 := position
				depth++
				{
					position305, tokenIndex305, depth305 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l306
					}
					position++
					goto l305
				l306:
					position, tokenIndex, depth = position305, tokenIndex305, depth305
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l307
					}
					position++
					goto l305
				l307:
					position, tokenIndex, depth = position305, tokenIndex305, depth305
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l308
					}
					position++
					goto l305
				l308:
					position, tokenIndex, depth = position305, tokenIndex305, depth305
					if buffer[position] != rune('_') {
						goto l301
					}
					position++
				}
			l305:
			l303:
				{
					position304, tokenIndex304, depth304 := position, tokenIndex, depth
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
							goto l304
						}
						position++
					}
				l309:
					goto l303
				l304:
					position, tokenIndex, depth = position304, tokenIndex304, depth304
				}
				depth--
				add(ruleName, position302)
			}
			return true
		l301:
			position, tokenIndex, depth = position301, tokenIndex301, depth301
			return false
		},
		/* 80 Reference <- <('.'? Key FollowUpRef)> */
		func() bool {
			position313, tokenIndex313, depth313 := position, tokenIndex, depth
			{
				position314 := position
				depth++
				{
					position315, tokenIndex315, depth315 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l315
					}
					position++
					goto l316
				l315:
					position, tokenIndex, depth = position315, tokenIndex315, depth315
				}
			l316:
				if !_rules[ruleKey]() {
					goto l313
				}
				if !_rules[ruleFollowUpRef]() {
					goto l313
				}
				depth--
				add(ruleReference, position314)
			}
			return true
		l313:
			position, tokenIndex, depth = position313, tokenIndex313, depth313
			return false
		},
		/* 81 FollowUpRef <- <('.' (Key / Index))*> */
		func() bool {
			{
				position318 := position
				depth++
			l319:
				{
					position320, tokenIndex320, depth320 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l320
					}
					position++
					{
						position321, tokenIndex321, depth321 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l322
						}
						goto l321
					l322:
						position, tokenIndex, depth = position321, tokenIndex321, depth321
						if !_rules[ruleIndex]() {
							goto l320
						}
					}
				l321:
					goto l319
				l320:
					position, tokenIndex, depth = position320, tokenIndex320, depth320
				}
				depth--
				add(ruleFollowUpRef, position318)
			}
			return true
		},
		/* 82 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position323, tokenIndex323, depth323 := position, tokenIndex, depth
			{
				position324 := position
				depth++
				{
					position325, tokenIndex325, depth325 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l326
					}
					position++
					goto l325
				l326:
					position, tokenIndex, depth = position325, tokenIndex325, depth325
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l327
					}
					position++
					goto l325
				l327:
					position, tokenIndex, depth = position325, tokenIndex325, depth325
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l328
					}
					position++
					goto l325
				l328:
					position, tokenIndex, depth = position325, tokenIndex325, depth325
					if buffer[position] != rune('_') {
						goto l323
					}
					position++
				}
			l325:
			l329:
				{
					position330, tokenIndex330, depth330 := position, tokenIndex, depth
					{
						position331, tokenIndex331, depth331 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l332
						}
						position++
						goto l331
					l332:
						position, tokenIndex, depth = position331, tokenIndex331, depth331
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l333
						}
						position++
						goto l331
					l333:
						position, tokenIndex, depth = position331, tokenIndex331, depth331
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l334
						}
						position++
						goto l331
					l334:
						position, tokenIndex, depth = position331, tokenIndex331, depth331
						if buffer[position] != rune('_') {
							goto l335
						}
						position++
						goto l331
					l335:
						position, tokenIndex, depth = position331, tokenIndex331, depth331
						if buffer[position] != rune('-') {
							goto l330
						}
						position++
					}
				l331:
					goto l329
				l330:
					position, tokenIndex, depth = position330, tokenIndex330, depth330
				}
				{
					position336, tokenIndex336, depth336 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l336
					}
					position++
					{
						position338, tokenIndex338, depth338 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l339
						}
						position++
						goto l338
					l339:
						position, tokenIndex, depth = position338, tokenIndex338, depth338
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l340
						}
						position++
						goto l338
					l340:
						position, tokenIndex, depth = position338, tokenIndex338, depth338
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l341
						}
						position++
						goto l338
					l341:
						position, tokenIndex, depth = position338, tokenIndex338, depth338
						if buffer[position] != rune('_') {
							goto l336
						}
						position++
					}
				l338:
				l342:
					{
						position343, tokenIndex343, depth343 := position, tokenIndex, depth
						{
							position344, tokenIndex344, depth344 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l345
							}
							position++
							goto l344
						l345:
							position, tokenIndex, depth = position344, tokenIndex344, depth344
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l346
							}
							position++
							goto l344
						l346:
							position, tokenIndex, depth = position344, tokenIndex344, depth344
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l347
							}
							position++
							goto l344
						l347:
							position, tokenIndex, depth = position344, tokenIndex344, depth344
							if buffer[position] != rune('_') {
								goto l348
							}
							position++
							goto l344
						l348:
							position, tokenIndex, depth = position344, tokenIndex344, depth344
							if buffer[position] != rune('-') {
								goto l343
							}
							position++
						}
					l344:
						goto l342
					l343:
						position, tokenIndex, depth = position343, tokenIndex343, depth343
					}
					goto l337
				l336:
					position, tokenIndex, depth = position336, tokenIndex336, depth336
				}
			l337:
				depth--
				add(ruleKey, position324)
			}
			return true
		l323:
			position, tokenIndex, depth = position323, tokenIndex323, depth323
			return false
		},
		/* 83 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position349, tokenIndex349, depth349 := position, tokenIndex, depth
			{
				position350 := position
				depth++
				if buffer[position] != rune('[') {
					goto l349
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l349
				}
				position++
			l351:
				{
					position352, tokenIndex352, depth352 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l352
					}
					position++
					goto l351
				l352:
					position, tokenIndex, depth = position352, tokenIndex352, depth352
				}
				if buffer[position] != rune(']') {
					goto l349
				}
				position++
				depth--
				add(ruleIndex, position350)
			}
			return true
		l349:
			position, tokenIndex, depth = position349, tokenIndex349, depth349
			return false
		},
		/* 84 IP <- <([0-9]+ '.' [0-9]+ '.' [0-9]+ '.' [0-9]+)> */
		func() bool {
			position353, tokenIndex353, depth353 := position, tokenIndex, depth
			{
				position354 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l353
				}
				position++
			l355:
				{
					position356, tokenIndex356, depth356 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l356
					}
					position++
					goto l355
				l356:
					position, tokenIndex, depth = position356, tokenIndex356, depth356
				}
				if buffer[position] != rune('.') {
					goto l353
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l353
				}
				position++
			l357:
				{
					position358, tokenIndex358, depth358 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l358
					}
					position++
					goto l357
				l358:
					position, tokenIndex, depth = position358, tokenIndex358, depth358
				}
				if buffer[position] != rune('.') {
					goto l353
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l353
				}
				position++
			l359:
				{
					position360, tokenIndex360, depth360 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l360
					}
					position++
					goto l359
				l360:
					position, tokenIndex, depth = position360, tokenIndex360, depth360
				}
				if buffer[position] != rune('.') {
					goto l353
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l353
				}
				position++
			l361:
				{
					position362, tokenIndex362, depth362 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l362
					}
					position++
					goto l361
				l362:
					position, tokenIndex, depth = position362, tokenIndex362, depth362
				}
				depth--
				add(ruleIP, position354)
			}
			return true
		l353:
			position, tokenIndex, depth = position353, tokenIndex353, depth353
			return false
		},
		/* 85 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position364 := position
				depth++
			l365:
				{
					position366, tokenIndex366, depth366 := position, tokenIndex, depth
					{
						position367, tokenIndex367, depth367 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l368
						}
						position++
						goto l367
					l368:
						position, tokenIndex, depth = position367, tokenIndex367, depth367
						if buffer[position] != rune('\t') {
							goto l369
						}
						position++
						goto l367
					l369:
						position, tokenIndex, depth = position367, tokenIndex367, depth367
						if buffer[position] != rune('\n') {
							goto l370
						}
						position++
						goto l367
					l370:
						position, tokenIndex, depth = position367, tokenIndex367, depth367
						if buffer[position] != rune('\r') {
							goto l366
						}
						position++
					}
				l367:
					goto l365
				l366:
					position, tokenIndex, depth = position366, tokenIndex366, depth366
				}
				depth--
				add(rulews, position364)
			}
			return true
		},
		/* 86 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position371, tokenIndex371, depth371 := position, tokenIndex, depth
			{
				position372 := position
				depth++
				{
					position375, tokenIndex375, depth375 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l376
					}
					position++
					goto l375
				l376:
					position, tokenIndex, depth = position375, tokenIndex375, depth375
					if buffer[position] != rune('\t') {
						goto l377
					}
					position++
					goto l375
				l377:
					position, tokenIndex, depth = position375, tokenIndex375, depth375
					if buffer[position] != rune('\n') {
						goto l378
					}
					position++
					goto l375
				l378:
					position, tokenIndex, depth = position375, tokenIndex375, depth375
					if buffer[position] != rune('\r') {
						goto l371
					}
					position++
				}
			l375:
			l373:
				{
					position374, tokenIndex374, depth374 := position, tokenIndex, depth
					{
						position379, tokenIndex379, depth379 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l380
						}
						position++
						goto l379
					l380:
						position, tokenIndex, depth = position379, tokenIndex379, depth379
						if buffer[position] != rune('\t') {
							goto l381
						}
						position++
						goto l379
					l381:
						position, tokenIndex, depth = position379, tokenIndex379, depth379
						if buffer[position] != rune('\n') {
							goto l382
						}
						position++
						goto l379
					l382:
						position, tokenIndex, depth = position379, tokenIndex379, depth379
						if buffer[position] != rune('\r') {
							goto l374
						}
						position++
					}
				l379:
					goto l373
				l374:
					position, tokenIndex, depth = position374, tokenIndex374, depth374
				}
				depth--
				add(rulereq_ws, position372)
			}
			return true
		l371:
			position, tokenIndex, depth = position371, tokenIndex371, depth371
			return false
		},
		/* 88 Action0 <- <{}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 89 Action1 <- <{}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
	}
	p.rules = _rules
}
