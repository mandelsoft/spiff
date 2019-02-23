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
	ruleOrOp
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
	"OrOp",
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
		/* 11 Or <- <(OrOp req_ws Level6)> */
		func() bool {
			position42, tokenIndex42, depth42 := position, tokenIndex, depth
			{
				position43 := position
				depth++
				if !_rules[ruleOrOp]() {
					goto l42
				}
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
		/* 12 OrOp <- <(('|' '|') / ('/' '/'))> */
		func() bool {
			position44, tokenIndex44, depth44 := position, tokenIndex, depth
			{
				position45 := position
				depth++
				{
					position46, tokenIndex46, depth46 := position, tokenIndex, depth
					if buffer[position] != rune('|') {
						goto l47
					}
					position++
					if buffer[position] != rune('|') {
						goto l47
					}
					position++
					goto l46
				l47:
					position, tokenIndex, depth = position46, tokenIndex46, depth46
					if buffer[position] != rune('/') {
						goto l44
					}
					position++
					if buffer[position] != rune('/') {
						goto l44
					}
					position++
				}
			l46:
				depth--
				add(ruleOrOp, position45)
			}
			return true
		l44:
			position, tokenIndex, depth = position44, tokenIndex44, depth44
			return false
		},
		/* 13 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position48, tokenIndex48, depth48 := position, tokenIndex, depth
			{
				position49 := position
				depth++
				{
					position50, tokenIndex50, depth50 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l51
					}
					goto l50
				l51:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if !_rules[ruleLevel5]() {
						goto l48
					}
				}
			l50:
				depth--
				add(ruleLevel6, position49)
			}
			return true
		l48:
			position, tokenIndex, depth = position48, tokenIndex48, depth48
			return false
		},
		/* 14 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position52, tokenIndex52, depth52 := position, tokenIndex, depth
			{
				position53 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l52
				}
				if !_rules[rulews]() {
					goto l52
				}
				if buffer[position] != rune('?') {
					goto l52
				}
				position++
				if !_rules[ruleExpression]() {
					goto l52
				}
				if buffer[position] != rune(':') {
					goto l52
				}
				position++
				if !_rules[ruleExpression]() {
					goto l52
				}
				depth--
				add(ruleConditional, position53)
			}
			return true
		l52:
			position, tokenIndex, depth = position52, tokenIndex52, depth52
			return false
		},
		/* 15 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position54, tokenIndex54, depth54 := position, tokenIndex, depth
			{
				position55 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l54
				}
			l56:
				{
					position57, tokenIndex57, depth57 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l57
					}
					goto l56
				l57:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
				}
				depth--
				add(ruleLevel5, position55)
			}
			return true
		l54:
			position, tokenIndex, depth = position54, tokenIndex54, depth54
			return false
		},
		/* 16 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position58, tokenIndex58, depth58 := position, tokenIndex, depth
			{
				position59 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l58
				}
				if !_rules[ruleLevel4]() {
					goto l58
				}
				depth--
				add(ruleConcatenation, position59)
			}
			return true
		l58:
			position, tokenIndex, depth = position58, tokenIndex58, depth58
			return false
		},
		/* 17 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position60, tokenIndex60, depth60 := position, tokenIndex, depth
			{
				position61 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l60
				}
			l62:
				{
					position63, tokenIndex63, depth63 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l63
					}
					{
						position64, tokenIndex64, depth64 := position, tokenIndex, depth
						if !_rules[ruleLogOr]() {
							goto l65
						}
						goto l64
					l65:
						position, tokenIndex, depth = position64, tokenIndex64, depth64
						if !_rules[ruleLogAnd]() {
							goto l63
						}
					}
				l64:
					goto l62
				l63:
					position, tokenIndex, depth = position63, tokenIndex63, depth63
				}
				depth--
				add(ruleLevel4, position61)
			}
			return true
		l60:
			position, tokenIndex, depth = position60, tokenIndex60, depth60
			return false
		},
		/* 18 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position66, tokenIndex66, depth66 := position, tokenIndex, depth
			{
				position67 := position
				depth++
				if buffer[position] != rune('-') {
					goto l66
				}
				position++
				if buffer[position] != rune('o') {
					goto l66
				}
				position++
				if buffer[position] != rune('r') {
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
				add(ruleLogOr, position67)
			}
			return true
		l66:
			position, tokenIndex, depth = position66, tokenIndex66, depth66
			return false
		},
		/* 19 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position68, tokenIndex68, depth68 := position, tokenIndex, depth
			{
				position69 := position
				depth++
				if buffer[position] != rune('-') {
					goto l68
				}
				position++
				if buffer[position] != rune('a') {
					goto l68
				}
				position++
				if buffer[position] != rune('n') {
					goto l68
				}
				position++
				if buffer[position] != rune('d') {
					goto l68
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l68
				}
				if !_rules[ruleLevel3]() {
					goto l68
				}
				depth--
				add(ruleLogAnd, position69)
			}
			return true
		l68:
			position, tokenIndex, depth = position68, tokenIndex68, depth68
			return false
		},
		/* 20 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position70, tokenIndex70, depth70 := position, tokenIndex, depth
			{
				position71 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l70
				}
			l72:
				{
					position73, tokenIndex73, depth73 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l73
					}
					if !_rules[ruleComparison]() {
						goto l73
					}
					goto l72
				l73:
					position, tokenIndex, depth = position73, tokenIndex73, depth73
				}
				depth--
				add(ruleLevel3, position71)
			}
			return true
		l70:
			position, tokenIndex, depth = position70, tokenIndex70, depth70
			return false
		},
		/* 21 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position74, tokenIndex74, depth74 := position, tokenIndex, depth
			{
				position75 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l74
				}
				if !_rules[rulereq_ws]() {
					goto l74
				}
				if !_rules[ruleLevel2]() {
					goto l74
				}
				depth--
				add(ruleComparison, position75)
			}
			return true
		l74:
			position, tokenIndex, depth = position74, tokenIndex74, depth74
			return false
		},
		/* 22 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position76, tokenIndex76, depth76 := position, tokenIndex, depth
			{
				position77 := position
				depth++
				{
					position78, tokenIndex78, depth78 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l79
					}
					position++
					if buffer[position] != rune('=') {
						goto l79
					}
					position++
					goto l78
				l79:
					position, tokenIndex, depth = position78, tokenIndex78, depth78
					if buffer[position] != rune('!') {
						goto l80
					}
					position++
					if buffer[position] != rune('=') {
						goto l80
					}
					position++
					goto l78
				l80:
					position, tokenIndex, depth = position78, tokenIndex78, depth78
					if buffer[position] != rune('<') {
						goto l81
					}
					position++
					if buffer[position] != rune('=') {
						goto l81
					}
					position++
					goto l78
				l81:
					position, tokenIndex, depth = position78, tokenIndex78, depth78
					if buffer[position] != rune('>') {
						goto l82
					}
					position++
					if buffer[position] != rune('=') {
						goto l82
					}
					position++
					goto l78
				l82:
					position, tokenIndex, depth = position78, tokenIndex78, depth78
					if buffer[position] != rune('>') {
						goto l83
					}
					position++
					goto l78
				l83:
					position, tokenIndex, depth = position78, tokenIndex78, depth78
					if buffer[position] != rune('<') {
						goto l84
					}
					position++
					goto l78
				l84:
					position, tokenIndex, depth = position78, tokenIndex78, depth78
					if buffer[position] != rune('>') {
						goto l76
					}
					position++
				}
			l78:
				depth--
				add(ruleCompareOp, position77)
			}
			return true
		l76:
			position, tokenIndex, depth = position76, tokenIndex76, depth76
			return false
		},
		/* 23 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position85, tokenIndex85, depth85 := position, tokenIndex, depth
			{
				position86 := position
				depth++
				if !_rules[ruleLevel1]() {
					goto l85
				}
			l87:
				{
					position88, tokenIndex88, depth88 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l88
					}
					{
						position89, tokenIndex89, depth89 := position, tokenIndex, depth
						if !_rules[ruleAddition]() {
							goto l90
						}
						goto l89
					l90:
						position, tokenIndex, depth = position89, tokenIndex89, depth89
						if !_rules[ruleSubtraction]() {
							goto l88
						}
					}
				l89:
					goto l87
				l88:
					position, tokenIndex, depth = position88, tokenIndex88, depth88
				}
				depth--
				add(ruleLevel2, position86)
			}
			return true
		l85:
			position, tokenIndex, depth = position85, tokenIndex85, depth85
			return false
		},
		/* 24 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position91, tokenIndex91, depth91 := position, tokenIndex, depth
			{
				position92 := position
				depth++
				if buffer[position] != rune('+') {
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
				add(ruleAddition, position92)
			}
			return true
		l91:
			position, tokenIndex, depth = position91, tokenIndex91, depth91
			return false
		},
		/* 25 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position93, tokenIndex93, depth93 := position, tokenIndex, depth
			{
				position94 := position
				depth++
				if buffer[position] != rune('-') {
					goto l93
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l93
				}
				if !_rules[ruleLevel1]() {
					goto l93
				}
				depth--
				add(ruleSubtraction, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 26 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position95, tokenIndex95, depth95 := position, tokenIndex, depth
			{
				position96 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l95
				}
			l97:
				{
					position98, tokenIndex98, depth98 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l98
					}
					{
						position99, tokenIndex99, depth99 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l100
						}
						goto l99
					l100:
						position, tokenIndex, depth = position99, tokenIndex99, depth99
						if !_rules[ruleDivision]() {
							goto l101
						}
						goto l99
					l101:
						position, tokenIndex, depth = position99, tokenIndex99, depth99
						if !_rules[ruleModulo]() {
							goto l98
						}
					}
				l99:
					goto l97
				l98:
					position, tokenIndex, depth = position98, tokenIndex98, depth98
				}
				depth--
				add(ruleLevel1, position96)
			}
			return true
		l95:
			position, tokenIndex, depth = position95, tokenIndex95, depth95
			return false
		},
		/* 27 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position102, tokenIndex102, depth102 := position, tokenIndex, depth
			{
				position103 := position
				depth++
				if buffer[position] != rune('*') {
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
				add(ruleMultiplication, position103)
			}
			return true
		l102:
			position, tokenIndex, depth = position102, tokenIndex102, depth102
			return false
		},
		/* 28 Division <- <('/' req_ws Level0)> */
		func() bool {
			position104, tokenIndex104, depth104 := position, tokenIndex, depth
			{
				position105 := position
				depth++
				if buffer[position] != rune('/') {
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
				add(ruleDivision, position105)
			}
			return true
		l104:
			position, tokenIndex, depth = position104, tokenIndex104, depth104
			return false
		},
		/* 29 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position106, tokenIndex106, depth106 := position, tokenIndex, depth
			{
				position107 := position
				depth++
				if buffer[position] != rune('%') {
					goto l106
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l106
				}
				if !_rules[ruleLevel0]() {
					goto l106
				}
				depth--
				add(ruleModulo, position107)
			}
			return true
		l106:
			position, tokenIndex, depth = position106, tokenIndex106, depth106
			return false
		},
		/* 30 Level0 <- <(IP / String / Integer / Boolean / Undefined / Nil / Symbol / Not / Substitution / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position108, tokenIndex108, depth108 := position, tokenIndex, depth
			{
				position109 := position
				depth++
				{
					position110, tokenIndex110, depth110 := position, tokenIndex, depth
					if !_rules[ruleIP]() {
						goto l111
					}
					goto l110
				l111:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleString]() {
						goto l112
					}
					goto l110
				l112:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleInteger]() {
						goto l113
					}
					goto l110
				l113:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleBoolean]() {
						goto l114
					}
					goto l110
				l114:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleUndefined]() {
						goto l115
					}
					goto l110
				l115:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleNil]() {
						goto l116
					}
					goto l110
				l116:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleSymbol]() {
						goto l117
					}
					goto l110
				l117:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleNot]() {
						goto l118
					}
					goto l110
				l118:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleSubstitution]() {
						goto l119
					}
					goto l110
				l119:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleMerge]() {
						goto l120
					}
					goto l110
				l120:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleAuto]() {
						goto l121
					}
					goto l110
				l121:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleLambda]() {
						goto l122
					}
					goto l110
				l122:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
					if !_rules[ruleChained]() {
						goto l108
					}
				}
			l110:
				depth--
				add(ruleLevel0, position109)
			}
			return true
		l108:
			position, tokenIndex, depth = position108, tokenIndex108, depth108
			return false
		},
		/* 31 Chained <- <((MapMapping / Mapping / MapSelection / Selection / Sum / List / Map / Range / Grouped / Reference) ChainedQualifiedExpression*)> */
		func() bool {
			position123, tokenIndex123, depth123 := position, tokenIndex, depth
			{
				position124 := position
				depth++
				{
					position125, tokenIndex125, depth125 := position, tokenIndex, depth
					if !_rules[ruleMapMapping]() {
						goto l126
					}
					goto l125
				l126:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleMapping]() {
						goto l127
					}
					goto l125
				l127:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleMapSelection]() {
						goto l128
					}
					goto l125
				l128:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleSelection]() {
						goto l129
					}
					goto l125
				l129:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleSum]() {
						goto l130
					}
					goto l125
				l130:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleList]() {
						goto l131
					}
					goto l125
				l131:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleMap]() {
						goto l132
					}
					goto l125
				l132:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleRange]() {
						goto l133
					}
					goto l125
				l133:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleGrouped]() {
						goto l134
					}
					goto l125
				l134:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleReference]() {
						goto l123
					}
				}
			l125:
			l135:
				{
					position136, tokenIndex136, depth136 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l136
					}
					goto l135
				l136:
					position, tokenIndex, depth = position136, tokenIndex136, depth136
				}
				depth--
				add(ruleChained, position124)
			}
			return true
		l123:
			position, tokenIndex, depth = position123, tokenIndex123, depth123
			return false
		},
		/* 32 ChainedQualifiedExpression <- <(ChainedCall / ('.' (ChainedRef / ChainedDynRef / Projection)))> */
		func() bool {
			position137, tokenIndex137, depth137 := position, tokenIndex, depth
			{
				position138 := position
				depth++
				{
					position139, tokenIndex139, depth139 := position, tokenIndex, depth
					if !_rules[ruleChainedCall]() {
						goto l140
					}
					goto l139
				l140:
					position, tokenIndex, depth = position139, tokenIndex139, depth139
					if buffer[position] != rune('.') {
						goto l137
					}
					position++
					{
						position141, tokenIndex141, depth141 := position, tokenIndex, depth
						if !_rules[ruleChainedRef]() {
							goto l142
						}
						goto l141
					l142:
						position, tokenIndex, depth = position141, tokenIndex141, depth141
						if !_rules[ruleChainedDynRef]() {
							goto l143
						}
						goto l141
					l143:
						position, tokenIndex, depth = position141, tokenIndex141, depth141
						if !_rules[ruleProjection]() {
							goto l137
						}
					}
				l141:
				}
			l139:
				depth--
				add(ruleChainedQualifiedExpression, position138)
			}
			return true
		l137:
			position, tokenIndex, depth = position137, tokenIndex137, depth137
			return false
		},
		/* 33 ChainedRef <- <((Key / Index) FollowUpRef)> */
		func() bool {
			position144, tokenIndex144, depth144 := position, tokenIndex, depth
			{
				position145 := position
				depth++
				{
					position146, tokenIndex146, depth146 := position, tokenIndex, depth
					if !_rules[ruleKey]() {
						goto l147
					}
					goto l146
				l147:
					position, tokenIndex, depth = position146, tokenIndex146, depth146
					if !_rules[ruleIndex]() {
						goto l144
					}
				}
			l146:
				if !_rules[ruleFollowUpRef]() {
					goto l144
				}
				depth--
				add(ruleChainedRef, position145)
			}
			return true
		l144:
			position, tokenIndex, depth = position144, tokenIndex144, depth144
			return false
		},
		/* 34 ChainedDynRef <- <('[' Expression ']')> */
		func() bool {
			position148, tokenIndex148, depth148 := position, tokenIndex, depth
			{
				position149 := position
				depth++
				if buffer[position] != rune('[') {
					goto l148
				}
				position++
				if !_rules[ruleExpression]() {
					goto l148
				}
				if buffer[position] != rune(']') {
					goto l148
				}
				position++
				depth--
				add(ruleChainedDynRef, position149)
			}
			return true
		l148:
			position, tokenIndex, depth = position148, tokenIndex148, depth148
			return false
		},
		/* 35 Slice <- <Range> */
		func() bool {
			position150, tokenIndex150, depth150 := position, tokenIndex, depth
			{
				position151 := position
				depth++
				if !_rules[ruleRange]() {
					goto l150
				}
				depth--
				add(ruleSlice, position151)
			}
			return true
		l150:
			position, tokenIndex, depth = position150, tokenIndex150, depth150
			return false
		},
		/* 36 ChainedCall <- <(StartArguments ExpressionList? ')')> */
		func() bool {
			position152, tokenIndex152, depth152 := position, tokenIndex, depth
			{
				position153 := position
				depth++
				if !_rules[ruleStartArguments]() {
					goto l152
				}
				{
					position154, tokenIndex154, depth154 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l154
					}
					goto l155
				l154:
					position, tokenIndex, depth = position154, tokenIndex154, depth154
				}
			l155:
				if buffer[position] != rune(')') {
					goto l152
				}
				position++
				depth--
				add(ruleChainedCall, position153)
			}
			return true
		l152:
			position, tokenIndex, depth = position152, tokenIndex152, depth152
			return false
		},
		/* 37 StartArguments <- <('(' ws)> */
		func() bool {
			position156, tokenIndex156, depth156 := position, tokenIndex, depth
			{
				position157 := position
				depth++
				if buffer[position] != rune('(') {
					goto l156
				}
				position++
				if !_rules[rulews]() {
					goto l156
				}
				depth--
				add(ruleStartArguments, position157)
			}
			return true
		l156:
			position, tokenIndex, depth = position156, tokenIndex156, depth156
			return false
		},
		/* 38 ExpressionList <- <(NextExpression (',' NextExpression)*)> */
		func() bool {
			position158, tokenIndex158, depth158 := position, tokenIndex, depth
			{
				position159 := position
				depth++
				if !_rules[ruleNextExpression]() {
					goto l158
				}
			l160:
				{
					position161, tokenIndex161, depth161 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l161
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l161
					}
					goto l160
				l161:
					position, tokenIndex, depth = position161, tokenIndex161, depth161
				}
				depth--
				add(ruleExpressionList, position159)
			}
			return true
		l158:
			position, tokenIndex, depth = position158, tokenIndex158, depth158
			return false
		},
		/* 39 NextExpression <- <Expression> */
		func() bool {
			position162, tokenIndex162, depth162 := position, tokenIndex, depth
			{
				position163 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l162
				}
				depth--
				add(ruleNextExpression, position163)
			}
			return true
		l162:
			position, tokenIndex, depth = position162, tokenIndex162, depth162
			return false
		},
		/* 40 Projection <- <((('[' '*' ']') / Slice) ProjectionValue ChainedQualifiedExpression*)> */
		func() bool {
			position164, tokenIndex164, depth164 := position, tokenIndex, depth
			{
				position165 := position
				depth++
				{
					position166, tokenIndex166, depth166 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l167
					}
					position++
					if buffer[position] != rune('*') {
						goto l167
					}
					position++
					if buffer[position] != rune(']') {
						goto l167
					}
					position++
					goto l166
				l167:
					position, tokenIndex, depth = position166, tokenIndex166, depth166
					if !_rules[ruleSlice]() {
						goto l164
					}
				}
			l166:
				if !_rules[ruleProjectionValue]() {
					goto l164
				}
			l168:
				{
					position169, tokenIndex169, depth169 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l169
					}
					goto l168
				l169:
					position, tokenIndex, depth = position169, tokenIndex169, depth169
				}
				depth--
				add(ruleProjection, position165)
			}
			return true
		l164:
			position, tokenIndex, depth = position164, tokenIndex164, depth164
			return false
		},
		/* 41 ProjectionValue <- <Action0> */
		func() bool {
			position170, tokenIndex170, depth170 := position, tokenIndex, depth
			{
				position171 := position
				depth++
				if !_rules[ruleAction0]() {
					goto l170
				}
				depth--
				add(ruleProjectionValue, position171)
			}
			return true
		l170:
			position, tokenIndex, depth = position170, tokenIndex170, depth170
			return false
		},
		/* 42 Substitution <- <('*' Level0)> */
		func() bool {
			position172, tokenIndex172, depth172 := position, tokenIndex, depth
			{
				position173 := position
				depth++
				if buffer[position] != rune('*') {
					goto l172
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l172
				}
				depth--
				add(ruleSubstitution, position173)
			}
			return true
		l172:
			position, tokenIndex, depth = position172, tokenIndex172, depth172
			return false
		},
		/* 43 Not <- <('!' ws Level0)> */
		func() bool {
			position174, tokenIndex174, depth174 := position, tokenIndex, depth
			{
				position175 := position
				depth++
				if buffer[position] != rune('!') {
					goto l174
				}
				position++
				if !_rules[rulews]() {
					goto l174
				}
				if !_rules[ruleLevel0]() {
					goto l174
				}
				depth--
				add(ruleNot, position175)
			}
			return true
		l174:
			position, tokenIndex, depth = position174, tokenIndex174, depth174
			return false
		},
		/* 44 Grouped <- <('(' Expression ')')> */
		func() bool {
			position176, tokenIndex176, depth176 := position, tokenIndex, depth
			{
				position177 := position
				depth++
				if buffer[position] != rune('(') {
					goto l176
				}
				position++
				if !_rules[ruleExpression]() {
					goto l176
				}
				if buffer[position] != rune(')') {
					goto l176
				}
				position++
				depth--
				add(ruleGrouped, position177)
			}
			return true
		l176:
			position, tokenIndex, depth = position176, tokenIndex176, depth176
			return false
		},
		/* 45 Range <- <(StartRange Expression? RangeOp Expression? ']')> */
		func() bool {
			position178, tokenIndex178, depth178 := position, tokenIndex, depth
			{
				position179 := position
				depth++
				if !_rules[ruleStartRange]() {
					goto l178
				}
				{
					position180, tokenIndex180, depth180 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l180
					}
					goto l181
				l180:
					position, tokenIndex, depth = position180, tokenIndex180, depth180
				}
			l181:
				if !_rules[ruleRangeOp]() {
					goto l178
				}
				{
					position182, tokenIndex182, depth182 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l182
					}
					goto l183
				l182:
					position, tokenIndex, depth = position182, tokenIndex182, depth182
				}
			l183:
				if buffer[position] != rune(']') {
					goto l178
				}
				position++
				depth--
				add(ruleRange, position179)
			}
			return true
		l178:
			position, tokenIndex, depth = position178, tokenIndex178, depth178
			return false
		},
		/* 46 StartRange <- <'['> */
		func() bool {
			position184, tokenIndex184, depth184 := position, tokenIndex, depth
			{
				position185 := position
				depth++
				if buffer[position] != rune('[') {
					goto l184
				}
				position++
				depth--
				add(ruleStartRange, position185)
			}
			return true
		l184:
			position, tokenIndex, depth = position184, tokenIndex184, depth184
			return false
		},
		/* 47 RangeOp <- <('.' '.')> */
		func() bool {
			position186, tokenIndex186, depth186 := position, tokenIndex, depth
			{
				position187 := position
				depth++
				if buffer[position] != rune('.') {
					goto l186
				}
				position++
				if buffer[position] != rune('.') {
					goto l186
				}
				position++
				depth--
				add(ruleRangeOp, position187)
			}
			return true
		l186:
			position, tokenIndex, depth = position186, tokenIndex186, depth186
			return false
		},
		/* 48 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position188, tokenIndex188, depth188 := position, tokenIndex, depth
			{
				position189 := position
				depth++
				{
					position190, tokenIndex190, depth190 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l190
					}
					position++
					goto l191
				l190:
					position, tokenIndex, depth = position190, tokenIndex190, depth190
				}
			l191:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l188
				}
				position++
			l192:
				{
					position193, tokenIndex193, depth193 := position, tokenIndex, depth
					{
						position194, tokenIndex194, depth194 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l195
						}
						position++
						goto l194
					l195:
						position, tokenIndex, depth = position194, tokenIndex194, depth194
						if buffer[position] != rune('_') {
							goto l193
						}
						position++
					}
				l194:
					goto l192
				l193:
					position, tokenIndex, depth = position193, tokenIndex193, depth193
				}
				depth--
				add(ruleInteger, position189)
			}
			return true
		l188:
			position, tokenIndex, depth = position188, tokenIndex188, depth188
			return false
		},
		/* 49 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position196, tokenIndex196, depth196 := position, tokenIndex, depth
			{
				position197 := position
				depth++
				if buffer[position] != rune('"') {
					goto l196
				}
				position++
			l198:
				{
					position199, tokenIndex199, depth199 := position, tokenIndex, depth
					{
						position200, tokenIndex200, depth200 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l201
						}
						position++
						if buffer[position] != rune('"') {
							goto l201
						}
						position++
						goto l200
					l201:
						position, tokenIndex, depth = position200, tokenIndex200, depth200
						{
							position202, tokenIndex202, depth202 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l202
							}
							position++
							goto l199
						l202:
							position, tokenIndex, depth = position202, tokenIndex202, depth202
						}
						if !matchDot() {
							goto l199
						}
					}
				l200:
					goto l198
				l199:
					position, tokenIndex, depth = position199, tokenIndex199, depth199
				}
				if buffer[position] != rune('"') {
					goto l196
				}
				position++
				depth--
				add(ruleString, position197)
			}
			return true
		l196:
			position, tokenIndex, depth = position196, tokenIndex196, depth196
			return false
		},
		/* 50 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position203, tokenIndex203, depth203 := position, tokenIndex, depth
			{
				position204 := position
				depth++
				{
					position205, tokenIndex205, depth205 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l206
					}
					position++
					if buffer[position] != rune('r') {
						goto l206
					}
					position++
					if buffer[position] != rune('u') {
						goto l206
					}
					position++
					if buffer[position] != rune('e') {
						goto l206
					}
					position++
					goto l205
				l206:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
					if buffer[position] != rune('f') {
						goto l203
					}
					position++
					if buffer[position] != rune('a') {
						goto l203
					}
					position++
					if buffer[position] != rune('l') {
						goto l203
					}
					position++
					if buffer[position] != rune('s') {
						goto l203
					}
					position++
					if buffer[position] != rune('e') {
						goto l203
					}
					position++
				}
			l205:
				depth--
				add(ruleBoolean, position204)
			}
			return true
		l203:
			position, tokenIndex, depth = position203, tokenIndex203, depth203
			return false
		},
		/* 51 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position207, tokenIndex207, depth207 := position, tokenIndex, depth
			{
				position208 := position
				depth++
				{
					position209, tokenIndex209, depth209 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l210
					}
					position++
					if buffer[position] != rune('i') {
						goto l210
					}
					position++
					if buffer[position] != rune('l') {
						goto l210
					}
					position++
					goto l209
				l210:
					position, tokenIndex, depth = position209, tokenIndex209, depth209
					if buffer[position] != rune('~') {
						goto l207
					}
					position++
				}
			l209:
				depth--
				add(ruleNil, position208)
			}
			return true
		l207:
			position, tokenIndex, depth = position207, tokenIndex207, depth207
			return false
		},
		/* 52 Undefined <- <('~' '~')> */
		func() bool {
			position211, tokenIndex211, depth211 := position, tokenIndex, depth
			{
				position212 := position
				depth++
				if buffer[position] != rune('~') {
					goto l211
				}
				position++
				if buffer[position] != rune('~') {
					goto l211
				}
				position++
				depth--
				add(ruleUndefined, position212)
			}
			return true
		l211:
			position, tokenIndex, depth = position211, tokenIndex211, depth211
			return false
		},
		/* 53 Symbol <- <('$' Name)> */
		func() bool {
			position213, tokenIndex213, depth213 := position, tokenIndex, depth
			{
				position214 := position
				depth++
				if buffer[position] != rune('$') {
					goto l213
				}
				position++
				if !_rules[ruleName]() {
					goto l213
				}
				depth--
				add(ruleSymbol, position214)
			}
			return true
		l213:
			position, tokenIndex, depth = position213, tokenIndex213, depth213
			return false
		},
		/* 54 List <- <(StartList ExpressionList? ']')> */
		func() bool {
			position215, tokenIndex215, depth215 := position, tokenIndex, depth
			{
				position216 := position
				depth++
				if !_rules[ruleStartList]() {
					goto l215
				}
				{
					position217, tokenIndex217, depth217 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l217
					}
					goto l218
				l217:
					position, tokenIndex, depth = position217, tokenIndex217, depth217
				}
			l218:
				if buffer[position] != rune(']') {
					goto l215
				}
				position++
				depth--
				add(ruleList, position216)
			}
			return true
		l215:
			position, tokenIndex, depth = position215, tokenIndex215, depth215
			return false
		},
		/* 55 StartList <- <('[' ws)> */
		func() bool {
			position219, tokenIndex219, depth219 := position, tokenIndex, depth
			{
				position220 := position
				depth++
				if buffer[position] != rune('[') {
					goto l219
				}
				position++
				if !_rules[rulews]() {
					goto l219
				}
				depth--
				add(ruleStartList, position220)
			}
			return true
		l219:
			position, tokenIndex, depth = position219, tokenIndex219, depth219
			return false
		},
		/* 56 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position221, tokenIndex221, depth221 := position, tokenIndex, depth
			{
				position222 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l221
				}
				if !_rules[rulews]() {
					goto l221
				}
				{
					position223, tokenIndex223, depth223 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l223
					}
					goto l224
				l223:
					position, tokenIndex, depth = position223, tokenIndex223, depth223
				}
			l224:
				if buffer[position] != rune('}') {
					goto l221
				}
				position++
				depth--
				add(ruleMap, position222)
			}
			return true
		l221:
			position, tokenIndex, depth = position221, tokenIndex221, depth221
			return false
		},
		/* 57 CreateMap <- <'{'> */
		func() bool {
			position225, tokenIndex225, depth225 := position, tokenIndex, depth
			{
				position226 := position
				depth++
				if buffer[position] != rune('{') {
					goto l225
				}
				position++
				depth--
				add(ruleCreateMap, position226)
			}
			return true
		l225:
			position, tokenIndex, depth = position225, tokenIndex225, depth225
			return false
		},
		/* 58 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position227, tokenIndex227, depth227 := position, tokenIndex, depth
			{
				position228 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l227
				}
			l229:
				{
					position230, tokenIndex230, depth230 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l230
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l230
					}
					goto l229
				l230:
					position, tokenIndex, depth = position230, tokenIndex230, depth230
				}
				depth--
				add(ruleAssignments, position228)
			}
			return true
		l227:
			position, tokenIndex, depth = position227, tokenIndex227, depth227
			return false
		},
		/* 59 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position231, tokenIndex231, depth231 := position, tokenIndex, depth
			{
				position232 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l231
				}
				if buffer[position] != rune('=') {
					goto l231
				}
				position++
				if !_rules[ruleExpression]() {
					goto l231
				}
				depth--
				add(ruleAssignment, position232)
			}
			return true
		l231:
			position, tokenIndex, depth = position231, tokenIndex231, depth231
			return false
		},
		/* 60 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position233, tokenIndex233, depth233 := position, tokenIndex, depth
			{
				position234 := position
				depth++
				{
					position235, tokenIndex235, depth235 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l236
					}
					goto l235
				l236:
					position, tokenIndex, depth = position235, tokenIndex235, depth235
					if !_rules[ruleSimpleMerge]() {
						goto l233
					}
				}
			l235:
				depth--
				add(ruleMerge, position234)
			}
			return true
		l233:
			position, tokenIndex, depth = position233, tokenIndex233, depth233
			return false
		},
		/* 61 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position237, tokenIndex237, depth237 := position, tokenIndex, depth
			{
				position238 := position
				depth++
				if buffer[position] != rune('m') {
					goto l237
				}
				position++
				if buffer[position] != rune('e') {
					goto l237
				}
				position++
				if buffer[position] != rune('r') {
					goto l237
				}
				position++
				if buffer[position] != rune('g') {
					goto l237
				}
				position++
				if buffer[position] != rune('e') {
					goto l237
				}
				position++
				{
					position239, tokenIndex239, depth239 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l239
					}
					if !_rules[ruleRequired]() {
						goto l239
					}
					goto l237
				l239:
					position, tokenIndex, depth = position239, tokenIndex239, depth239
				}
				{
					position240, tokenIndex240, depth240 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l240
					}
					{
						position242, tokenIndex242, depth242 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l243
						}
						goto l242
					l243:
						position, tokenIndex, depth = position242, tokenIndex242, depth242
						if !_rules[ruleOn]() {
							goto l240
						}
					}
				l242:
					goto l241
				l240:
					position, tokenIndex, depth = position240, tokenIndex240, depth240
				}
			l241:
				if !_rules[rulereq_ws]() {
					goto l237
				}
				if !_rules[ruleReference]() {
					goto l237
				}
				depth--
				add(ruleRefMerge, position238)
			}
			return true
		l237:
			position, tokenIndex, depth = position237, tokenIndex237, depth237
			return false
		},
		/* 62 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' !'(' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position244, tokenIndex244, depth244 := position, tokenIndex, depth
			{
				position245 := position
				depth++
				if buffer[position] != rune('m') {
					goto l244
				}
				position++
				if buffer[position] != rune('e') {
					goto l244
				}
				position++
				if buffer[position] != rune('r') {
					goto l244
				}
				position++
				if buffer[position] != rune('g') {
					goto l244
				}
				position++
				if buffer[position] != rune('e') {
					goto l244
				}
				position++
				{
					position246, tokenIndex246, depth246 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l246
					}
					position++
					goto l244
				l246:
					position, tokenIndex, depth = position246, tokenIndex246, depth246
				}
				{
					position247, tokenIndex247, depth247 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l247
					}
					{
						position249, tokenIndex249, depth249 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l250
						}
						goto l249
					l250:
						position, tokenIndex, depth = position249, tokenIndex249, depth249
						if !_rules[ruleRequired]() {
							goto l251
						}
						goto l249
					l251:
						position, tokenIndex, depth = position249, tokenIndex249, depth249
						if !_rules[ruleOn]() {
							goto l247
						}
					}
				l249:
					goto l248
				l247:
					position, tokenIndex, depth = position247, tokenIndex247, depth247
				}
			l248:
				depth--
				add(ruleSimpleMerge, position245)
			}
			return true
		l244:
			position, tokenIndex, depth = position244, tokenIndex244, depth244
			return false
		},
		/* 63 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position252, tokenIndex252, depth252 := position, tokenIndex, depth
			{
				position253 := position
				depth++
				if buffer[position] != rune('r') {
					goto l252
				}
				position++
				if buffer[position] != rune('e') {
					goto l252
				}
				position++
				if buffer[position] != rune('p') {
					goto l252
				}
				position++
				if buffer[position] != rune('l') {
					goto l252
				}
				position++
				if buffer[position] != rune('a') {
					goto l252
				}
				position++
				if buffer[position] != rune('c') {
					goto l252
				}
				position++
				if buffer[position] != rune('e') {
					goto l252
				}
				position++
				depth--
				add(ruleReplace, position253)
			}
			return true
		l252:
			position, tokenIndex, depth = position252, tokenIndex252, depth252
			return false
		},
		/* 64 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position254, tokenIndex254, depth254 := position, tokenIndex, depth
			{
				position255 := position
				depth++
				if buffer[position] != rune('r') {
					goto l254
				}
				position++
				if buffer[position] != rune('e') {
					goto l254
				}
				position++
				if buffer[position] != rune('q') {
					goto l254
				}
				position++
				if buffer[position] != rune('u') {
					goto l254
				}
				position++
				if buffer[position] != rune('i') {
					goto l254
				}
				position++
				if buffer[position] != rune('r') {
					goto l254
				}
				position++
				if buffer[position] != rune('e') {
					goto l254
				}
				position++
				if buffer[position] != rune('d') {
					goto l254
				}
				position++
				depth--
				add(ruleRequired, position255)
			}
			return true
		l254:
			position, tokenIndex, depth = position254, tokenIndex254, depth254
			return false
		},
		/* 65 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position256, tokenIndex256, depth256 := position, tokenIndex, depth
			{
				position257 := position
				depth++
				if buffer[position] != rune('o') {
					goto l256
				}
				position++
				if buffer[position] != rune('n') {
					goto l256
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l256
				}
				if !_rules[ruleName]() {
					goto l256
				}
				depth--
				add(ruleOn, position257)
			}
			return true
		l256:
			position, tokenIndex, depth = position256, tokenIndex256, depth256
			return false
		},
		/* 66 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position258, tokenIndex258, depth258 := position, tokenIndex, depth
			{
				position259 := position
				depth++
				if buffer[position] != rune('a') {
					goto l258
				}
				position++
				if buffer[position] != rune('u') {
					goto l258
				}
				position++
				if buffer[position] != rune('t') {
					goto l258
				}
				position++
				if buffer[position] != rune('o') {
					goto l258
				}
				position++
				depth--
				add(ruleAuto, position259)
			}
			return true
		l258:
			position, tokenIndex, depth = position258, tokenIndex258, depth258
			return false
		},
		/* 67 MapMapping <- <('m' 'a' 'p' '{' Level7 (LambdaExpr / ('|' Expression)) '}')> */
		func() bool {
			position260, tokenIndex260, depth260 := position, tokenIndex, depth
			{
				position261 := position
				depth++
				if buffer[position] != rune('m') {
					goto l260
				}
				position++
				if buffer[position] != rune('a') {
					goto l260
				}
				position++
				if buffer[position] != rune('p') {
					goto l260
				}
				position++
				if buffer[position] != rune('{') {
					goto l260
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l260
				}
				{
					position262, tokenIndex262, depth262 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l263
					}
					goto l262
				l263:
					position, tokenIndex, depth = position262, tokenIndex262, depth262
					if buffer[position] != rune('|') {
						goto l260
					}
					position++
					if !_rules[ruleExpression]() {
						goto l260
					}
				}
			l262:
				if buffer[position] != rune('}') {
					goto l260
				}
				position++
				depth--
				add(ruleMapMapping, position261)
			}
			return true
		l260:
			position, tokenIndex, depth = position260, tokenIndex260, depth260
			return false
		},
		/* 68 Mapping <- <('m' 'a' 'p' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position264, tokenIndex264, depth264 := position, tokenIndex, depth
			{
				position265 := position
				depth++
				if buffer[position] != rune('m') {
					goto l264
				}
				position++
				if buffer[position] != rune('a') {
					goto l264
				}
				position++
				if buffer[position] != rune('p') {
					goto l264
				}
				position++
				if buffer[position] != rune('[') {
					goto l264
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l264
				}
				{
					position266, tokenIndex266, depth266 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l267
					}
					goto l266
				l267:
					position, tokenIndex, depth = position266, tokenIndex266, depth266
					if buffer[position] != rune('|') {
						goto l264
					}
					position++
					if !_rules[ruleExpression]() {
						goto l264
					}
				}
			l266:
				if buffer[position] != rune(']') {
					goto l264
				}
				position++
				depth--
				add(ruleMapping, position265)
			}
			return true
		l264:
			position, tokenIndex, depth = position264, tokenIndex264, depth264
			return false
		},
		/* 69 MapSelection <- <('s' 'e' 'l' 'e' 'c' 't' '{' Level7 (LambdaExpr / ('|' Expression)) '}')> */
		func() bool {
			position268, tokenIndex268, depth268 := position, tokenIndex, depth
			{
				position269 := position
				depth++
				if buffer[position] != rune('s') {
					goto l268
				}
				position++
				if buffer[position] != rune('e') {
					goto l268
				}
				position++
				if buffer[position] != rune('l') {
					goto l268
				}
				position++
				if buffer[position] != rune('e') {
					goto l268
				}
				position++
				if buffer[position] != rune('c') {
					goto l268
				}
				position++
				if buffer[position] != rune('t') {
					goto l268
				}
				position++
				if buffer[position] != rune('{') {
					goto l268
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l268
				}
				{
					position270, tokenIndex270, depth270 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l271
					}
					goto l270
				l271:
					position, tokenIndex, depth = position270, tokenIndex270, depth270
					if buffer[position] != rune('|') {
						goto l268
					}
					position++
					if !_rules[ruleExpression]() {
						goto l268
					}
				}
			l270:
				if buffer[position] != rune('}') {
					goto l268
				}
				position++
				depth--
				add(ruleMapSelection, position269)
			}
			return true
		l268:
			position, tokenIndex, depth = position268, tokenIndex268, depth268
			return false
		},
		/* 70 Selection <- <('s' 'e' 'l' 'e' 'c' 't' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position272, tokenIndex272, depth272 := position, tokenIndex, depth
			{
				position273 := position
				depth++
				if buffer[position] != rune('s') {
					goto l272
				}
				position++
				if buffer[position] != rune('e') {
					goto l272
				}
				position++
				if buffer[position] != rune('l') {
					goto l272
				}
				position++
				if buffer[position] != rune('e') {
					goto l272
				}
				position++
				if buffer[position] != rune('c') {
					goto l272
				}
				position++
				if buffer[position] != rune('t') {
					goto l272
				}
				position++
				if buffer[position] != rune('[') {
					goto l272
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l272
				}
				{
					position274, tokenIndex274, depth274 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l275
					}
					goto l274
				l275:
					position, tokenIndex, depth = position274, tokenIndex274, depth274
					if buffer[position] != rune('|') {
						goto l272
					}
					position++
					if !_rules[ruleExpression]() {
						goto l272
					}
				}
			l274:
				if buffer[position] != rune(']') {
					goto l272
				}
				position++
				depth--
				add(ruleSelection, position273)
			}
			return true
		l272:
			position, tokenIndex, depth = position272, tokenIndex272, depth272
			return false
		},
		/* 71 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position276, tokenIndex276, depth276 := position, tokenIndex, depth
			{
				position277 := position
				depth++
				if buffer[position] != rune('s') {
					goto l276
				}
				position++
				if buffer[position] != rune('u') {
					goto l276
				}
				position++
				if buffer[position] != rune('m') {
					goto l276
				}
				position++
				if buffer[position] != rune('[') {
					goto l276
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l276
				}
				if buffer[position] != rune('|') {
					goto l276
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l276
				}
				{
					position278, tokenIndex278, depth278 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l279
					}
					goto l278
				l279:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
					if buffer[position] != rune('|') {
						goto l276
					}
					position++
					if !_rules[ruleExpression]() {
						goto l276
					}
				}
			l278:
				if buffer[position] != rune(']') {
					goto l276
				}
				position++
				depth--
				add(ruleSum, position277)
			}
			return true
		l276:
			position, tokenIndex, depth = position276, tokenIndex276, depth276
			return false
		},
		/* 72 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position280, tokenIndex280, depth280 := position, tokenIndex, depth
			{
				position281 := position
				depth++
				if buffer[position] != rune('l') {
					goto l280
				}
				position++
				if buffer[position] != rune('a') {
					goto l280
				}
				position++
				if buffer[position] != rune('m') {
					goto l280
				}
				position++
				if buffer[position] != rune('b') {
					goto l280
				}
				position++
				if buffer[position] != rune('d') {
					goto l280
				}
				position++
				if buffer[position] != rune('a') {
					goto l280
				}
				position++
				{
					position282, tokenIndex282, depth282 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l283
					}
					goto l282
				l283:
					position, tokenIndex, depth = position282, tokenIndex282, depth282
					if !_rules[ruleLambdaExpr]() {
						goto l280
					}
				}
			l282:
				depth--
				add(ruleLambda, position281)
			}
			return true
		l280:
			position, tokenIndex, depth = position280, tokenIndex280, depth280
			return false
		},
		/* 73 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position284, tokenIndex284, depth284 := position, tokenIndex, depth
			{
				position285 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l284
				}
				if !_rules[ruleExpression]() {
					goto l284
				}
				depth--
				add(ruleLambdaRef, position285)
			}
			return true
		l284:
			position, tokenIndex, depth = position284, tokenIndex284, depth284
			return false
		},
		/* 74 LambdaExpr <- <(ws Params ws ('-' '>') Expression)> */
		func() bool {
			position286, tokenIndex286, depth286 := position, tokenIndex, depth
			{
				position287 := position
				depth++
				if !_rules[rulews]() {
					goto l286
				}
				if !_rules[ruleParams]() {
					goto l286
				}
				if !_rules[rulews]() {
					goto l286
				}
				if buffer[position] != rune('-') {
					goto l286
				}
				position++
				if buffer[position] != rune('>') {
					goto l286
				}
				position++
				if !_rules[ruleExpression]() {
					goto l286
				}
				depth--
				add(ruleLambdaExpr, position287)
			}
			return true
		l286:
			position, tokenIndex, depth = position286, tokenIndex286, depth286
			return false
		},
		/* 75 Params <- <('|' StartParams ws Names? ws '|')> */
		func() bool {
			position288, tokenIndex288, depth288 := position, tokenIndex, depth
			{
				position289 := position
				depth++
				if buffer[position] != rune('|') {
					goto l288
				}
				position++
				if !_rules[ruleStartParams]() {
					goto l288
				}
				if !_rules[rulews]() {
					goto l288
				}
				{
					position290, tokenIndex290, depth290 := position, tokenIndex, depth
					if !_rules[ruleNames]() {
						goto l290
					}
					goto l291
				l290:
					position, tokenIndex, depth = position290, tokenIndex290, depth290
				}
			l291:
				if !_rules[rulews]() {
					goto l288
				}
				if buffer[position] != rune('|') {
					goto l288
				}
				position++
				depth--
				add(ruleParams, position289)
			}
			return true
		l288:
			position, tokenIndex, depth = position288, tokenIndex288, depth288
			return false
		},
		/* 76 StartParams <- <Action1> */
		func() bool {
			position292, tokenIndex292, depth292 := position, tokenIndex, depth
			{
				position293 := position
				depth++
				if !_rules[ruleAction1]() {
					goto l292
				}
				depth--
				add(ruleStartParams, position293)
			}
			return true
		l292:
			position, tokenIndex, depth = position292, tokenIndex292, depth292
			return false
		},
		/* 77 Names <- <(NextName (',' NextName)*)> */
		func() bool {
			position294, tokenIndex294, depth294 := position, tokenIndex, depth
			{
				position295 := position
				depth++
				if !_rules[ruleNextName]() {
					goto l294
				}
			l296:
				{
					position297, tokenIndex297, depth297 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l297
					}
					position++
					if !_rules[ruleNextName]() {
						goto l297
					}
					goto l296
				l297:
					position, tokenIndex, depth = position297, tokenIndex297, depth297
				}
				depth--
				add(ruleNames, position295)
			}
			return true
		l294:
			position, tokenIndex, depth = position294, tokenIndex294, depth294
			return false
		},
		/* 78 NextName <- <(ws Name ws)> */
		func() bool {
			position298, tokenIndex298, depth298 := position, tokenIndex, depth
			{
				position299 := position
				depth++
				if !_rules[rulews]() {
					goto l298
				}
				if !_rules[ruleName]() {
					goto l298
				}
				if !_rules[rulews]() {
					goto l298
				}
				depth--
				add(ruleNextName, position299)
			}
			return true
		l298:
			position, tokenIndex, depth = position298, tokenIndex298, depth298
			return false
		},
		/* 79 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position300, tokenIndex300, depth300 := position, tokenIndex, depth
			{
				position301 := position
				depth++
				{
					position304, tokenIndex304, depth304 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l305
					}
					position++
					goto l304
				l305:
					position, tokenIndex, depth = position304, tokenIndex304, depth304
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l306
					}
					position++
					goto l304
				l306:
					position, tokenIndex, depth = position304, tokenIndex304, depth304
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l307
					}
					position++
					goto l304
				l307:
					position, tokenIndex, depth = position304, tokenIndex304, depth304
					if buffer[position] != rune('_') {
						goto l300
					}
					position++
				}
			l304:
			l302:
				{
					position303, tokenIndex303, depth303 := position, tokenIndex, depth
					{
						position308, tokenIndex308, depth308 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l309
						}
						position++
						goto l308
					l309:
						position, tokenIndex, depth = position308, tokenIndex308, depth308
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l310
						}
						position++
						goto l308
					l310:
						position, tokenIndex, depth = position308, tokenIndex308, depth308
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l311
						}
						position++
						goto l308
					l311:
						position, tokenIndex, depth = position308, tokenIndex308, depth308
						if buffer[position] != rune('_') {
							goto l303
						}
						position++
					}
				l308:
					goto l302
				l303:
					position, tokenIndex, depth = position303, tokenIndex303, depth303
				}
				depth--
				add(ruleName, position301)
			}
			return true
		l300:
			position, tokenIndex, depth = position300, tokenIndex300, depth300
			return false
		},
		/* 80 Reference <- <('.'? Key FollowUpRef)> */
		func() bool {
			position312, tokenIndex312, depth312 := position, tokenIndex, depth
			{
				position313 := position
				depth++
				{
					position314, tokenIndex314, depth314 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l314
					}
					position++
					goto l315
				l314:
					position, tokenIndex, depth = position314, tokenIndex314, depth314
				}
			l315:
				if !_rules[ruleKey]() {
					goto l312
				}
				if !_rules[ruleFollowUpRef]() {
					goto l312
				}
				depth--
				add(ruleReference, position313)
			}
			return true
		l312:
			position, tokenIndex, depth = position312, tokenIndex312, depth312
			return false
		},
		/* 81 FollowUpRef <- <('.' (Key / Index))*> */
		func() bool {
			{
				position317 := position
				depth++
			l318:
				{
					position319, tokenIndex319, depth319 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l319
					}
					position++
					{
						position320, tokenIndex320, depth320 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l321
						}
						goto l320
					l321:
						position, tokenIndex, depth = position320, tokenIndex320, depth320
						if !_rules[ruleIndex]() {
							goto l319
						}
					}
				l320:
					goto l318
				l319:
					position, tokenIndex, depth = position319, tokenIndex319, depth319
				}
				depth--
				add(ruleFollowUpRef, position317)
			}
			return true
		},
		/* 82 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position322, tokenIndex322, depth322 := position, tokenIndex, depth
			{
				position323 := position
				depth++
				{
					position324, tokenIndex324, depth324 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l325
					}
					position++
					goto l324
				l325:
					position, tokenIndex, depth = position324, tokenIndex324, depth324
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l326
					}
					position++
					goto l324
				l326:
					position, tokenIndex, depth = position324, tokenIndex324, depth324
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l327
					}
					position++
					goto l324
				l327:
					position, tokenIndex, depth = position324, tokenIndex324, depth324
					if buffer[position] != rune('_') {
						goto l322
					}
					position++
				}
			l324:
			l328:
				{
					position329, tokenIndex329, depth329 := position, tokenIndex, depth
					{
						position330, tokenIndex330, depth330 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l331
						}
						position++
						goto l330
					l331:
						position, tokenIndex, depth = position330, tokenIndex330, depth330
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l332
						}
						position++
						goto l330
					l332:
						position, tokenIndex, depth = position330, tokenIndex330, depth330
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l333
						}
						position++
						goto l330
					l333:
						position, tokenIndex, depth = position330, tokenIndex330, depth330
						if buffer[position] != rune('_') {
							goto l334
						}
						position++
						goto l330
					l334:
						position, tokenIndex, depth = position330, tokenIndex330, depth330
						if buffer[position] != rune('-') {
							goto l329
						}
						position++
					}
				l330:
					goto l328
				l329:
					position, tokenIndex, depth = position329, tokenIndex329, depth329
				}
				{
					position335, tokenIndex335, depth335 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l335
					}
					position++
					{
						position337, tokenIndex337, depth337 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l338
						}
						position++
						goto l337
					l338:
						position, tokenIndex, depth = position337, tokenIndex337, depth337
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l339
						}
						position++
						goto l337
					l339:
						position, tokenIndex, depth = position337, tokenIndex337, depth337
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l340
						}
						position++
						goto l337
					l340:
						position, tokenIndex, depth = position337, tokenIndex337, depth337
						if buffer[position] != rune('_') {
							goto l335
						}
						position++
					}
				l337:
				l341:
					{
						position342, tokenIndex342, depth342 := position, tokenIndex, depth
						{
							position343, tokenIndex343, depth343 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l344
							}
							position++
							goto l343
						l344:
							position, tokenIndex, depth = position343, tokenIndex343, depth343
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l345
							}
							position++
							goto l343
						l345:
							position, tokenIndex, depth = position343, tokenIndex343, depth343
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l346
							}
							position++
							goto l343
						l346:
							position, tokenIndex, depth = position343, tokenIndex343, depth343
							if buffer[position] != rune('_') {
								goto l347
							}
							position++
							goto l343
						l347:
							position, tokenIndex, depth = position343, tokenIndex343, depth343
							if buffer[position] != rune('-') {
								goto l342
							}
							position++
						}
					l343:
						goto l341
					l342:
						position, tokenIndex, depth = position342, tokenIndex342, depth342
					}
					goto l336
				l335:
					position, tokenIndex, depth = position335, tokenIndex335, depth335
				}
			l336:
				depth--
				add(ruleKey, position323)
			}
			return true
		l322:
			position, tokenIndex, depth = position322, tokenIndex322, depth322
			return false
		},
		/* 83 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position348, tokenIndex348, depth348 := position, tokenIndex, depth
			{
				position349 := position
				depth++
				if buffer[position] != rune('[') {
					goto l348
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l348
				}
				position++
			l350:
				{
					position351, tokenIndex351, depth351 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l351
					}
					position++
					goto l350
				l351:
					position, tokenIndex, depth = position351, tokenIndex351, depth351
				}
				if buffer[position] != rune(']') {
					goto l348
				}
				position++
				depth--
				add(ruleIndex, position349)
			}
			return true
		l348:
			position, tokenIndex, depth = position348, tokenIndex348, depth348
			return false
		},
		/* 84 IP <- <([0-9]+ '.' [0-9]+ '.' [0-9]+ '.' [0-9]+)> */
		func() bool {
			position352, tokenIndex352, depth352 := position, tokenIndex, depth
			{
				position353 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l352
				}
				position++
			l354:
				{
					position355, tokenIndex355, depth355 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l355
					}
					position++
					goto l354
				l355:
					position, tokenIndex, depth = position355, tokenIndex355, depth355
				}
				if buffer[position] != rune('.') {
					goto l352
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l352
				}
				position++
			l356:
				{
					position357, tokenIndex357, depth357 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l357
					}
					position++
					goto l356
				l357:
					position, tokenIndex, depth = position357, tokenIndex357, depth357
				}
				if buffer[position] != rune('.') {
					goto l352
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l352
				}
				position++
			l358:
				{
					position359, tokenIndex359, depth359 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l359
					}
					position++
					goto l358
				l359:
					position, tokenIndex, depth = position359, tokenIndex359, depth359
				}
				if buffer[position] != rune('.') {
					goto l352
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l352
				}
				position++
			l360:
				{
					position361, tokenIndex361, depth361 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l361
					}
					position++
					goto l360
				l361:
					position, tokenIndex, depth = position361, tokenIndex361, depth361
				}
				depth--
				add(ruleIP, position353)
			}
			return true
		l352:
			position, tokenIndex, depth = position352, tokenIndex352, depth352
			return false
		},
		/* 85 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position363 := position
				depth++
			l364:
				{
					position365, tokenIndex365, depth365 := position, tokenIndex, depth
					{
						position366, tokenIndex366, depth366 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l367
						}
						position++
						goto l366
					l367:
						position, tokenIndex, depth = position366, tokenIndex366, depth366
						if buffer[position] != rune('\t') {
							goto l368
						}
						position++
						goto l366
					l368:
						position, tokenIndex, depth = position366, tokenIndex366, depth366
						if buffer[position] != rune('\n') {
							goto l369
						}
						position++
						goto l366
					l369:
						position, tokenIndex, depth = position366, tokenIndex366, depth366
						if buffer[position] != rune('\r') {
							goto l365
						}
						position++
					}
				l366:
					goto l364
				l365:
					position, tokenIndex, depth = position365, tokenIndex365, depth365
				}
				depth--
				add(rulews, position363)
			}
			return true
		},
		/* 86 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position370, tokenIndex370, depth370 := position, tokenIndex, depth
			{
				position371 := position
				depth++
				{
					position374, tokenIndex374, depth374 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l375
					}
					position++
					goto l374
				l375:
					position, tokenIndex, depth = position374, tokenIndex374, depth374
					if buffer[position] != rune('\t') {
						goto l376
					}
					position++
					goto l374
				l376:
					position, tokenIndex, depth = position374, tokenIndex374, depth374
					if buffer[position] != rune('\n') {
						goto l377
					}
					position++
					goto l374
				l377:
					position, tokenIndex, depth = position374, tokenIndex374, depth374
					if buffer[position] != rune('\r') {
						goto l370
					}
					position++
				}
			l374:
			l372:
				{
					position373, tokenIndex373, depth373 := position, tokenIndex, depth
					{
						position378, tokenIndex378, depth378 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l379
						}
						position++
						goto l378
					l379:
						position, tokenIndex, depth = position378, tokenIndex378, depth378
						if buffer[position] != rune('\t') {
							goto l380
						}
						position++
						goto l378
					l380:
						position, tokenIndex, depth = position378, tokenIndex378, depth378
						if buffer[position] != rune('\n') {
							goto l381
						}
						position++
						goto l378
					l381:
						position, tokenIndex, depth = position378, tokenIndex378, depth378
						if buffer[position] != rune('\r') {
							goto l373
						}
						position++
					}
				l378:
					goto l372
				l373:
					position, tokenIndex, depth = position373, tokenIndex373, depth373
				}
				depth--
				add(rulereq_ws, position371)
			}
			return true
		l370:
			position, tokenIndex, depth = position370, tokenIndex370, depth370
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
