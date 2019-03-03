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
	ruleDefault
	ruleSync
	ruleLambdaExt
	ruleLambdaOrExpr
	ruleCatch
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
	rulePathComponent
	ruleKey
	ruleIndex
	ruleIP
	rulews
	rulereq_ws
	ruleAction0
	ruleAction1
	ruleAction2

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
	"Default",
	"Sync",
	"LambdaExt",
	"LambdaOrExpr",
	"Catch",
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
	"PathComponent",
	"Key",
	"Index",
	"IP",
	"ws",
	"req_ws",
	"Action0",
	"Action1",
	"Action2",

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
	rules  [97]func() bool
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

		case ruleAction2:

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
		/* 6 Expression <- <((Scoped / LambdaExpr / Level7) ws)> */
		func() bool {
			position25, tokenIndex25, depth25 := position, tokenIndex, depth
			{
				position26 := position
				depth++
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
		/* 7 Scoped <- <(ws Scope ws Expression)> */
		func() bool {
			position30, tokenIndex30, depth30 := position, tokenIndex, depth
			{
				position31 := position
				depth++
				if !_rules[rulews]() {
					goto l30
				}
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
		/* 10 Level7 <- <(ws Level6 (req_ws Or)*)> */
		func() bool {
			position38, tokenIndex38, depth38 := position, tokenIndex, depth
			{
				position39 := position
				depth++
				if !_rules[rulews]() {
					goto l38
				}
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
		/* 31 Chained <- <((MapMapping / Sync / Catch / Mapping / MapSelection / Selection / Sum / List / Map / Range / Grouped / Reference) ChainedQualifiedExpression*)> */
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
					if !_rules[ruleSync]() {
						goto l127
					}
					goto l125
				l127:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleCatch]() {
						goto l128
					}
					goto l125
				l128:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleMapping]() {
						goto l129
					}
					goto l125
				l129:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleMapSelection]() {
						goto l130
					}
					goto l125
				l130:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleSelection]() {
						goto l131
					}
					goto l125
				l131:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleSum]() {
						goto l132
					}
					goto l125
				l132:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleList]() {
						goto l133
					}
					goto l125
				l133:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleMap]() {
						goto l134
					}
					goto l125
				l134:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleRange]() {
						goto l135
					}
					goto l125
				l135:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleGrouped]() {
						goto l136
					}
					goto l125
				l136:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
					if !_rules[ruleReference]() {
						goto l123
					}
				}
			l125:
			l137:
				{
					position138, tokenIndex138, depth138 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l138
					}
					goto l137
				l138:
					position, tokenIndex, depth = position138, tokenIndex138, depth138
				}
				depth--
				add(ruleChained, position124)
			}
			return true
		l123:
			position, tokenIndex, depth = position123, tokenIndex123, depth123
			return false
		},
		/* 32 ChainedQualifiedExpression <- <(ChainedCall / ChainedRef / ChainedDynRef / Projection)> */
		func() bool {
			position139, tokenIndex139, depth139 := position, tokenIndex, depth
			{
				position140 := position
				depth++
				{
					position141, tokenIndex141, depth141 := position, tokenIndex, depth
					if !_rules[ruleChainedCall]() {
						goto l142
					}
					goto l141
				l142:
					position, tokenIndex, depth = position141, tokenIndex141, depth141
					if !_rules[ruleChainedRef]() {
						goto l143
					}
					goto l141
				l143:
					position, tokenIndex, depth = position141, tokenIndex141, depth141
					if !_rules[ruleChainedDynRef]() {
						goto l144
					}
					goto l141
				l144:
					position, tokenIndex, depth = position141, tokenIndex141, depth141
					if !_rules[ruleProjection]() {
						goto l139
					}
				}
			l141:
				depth--
				add(ruleChainedQualifiedExpression, position140)
			}
			return true
		l139:
			position, tokenIndex, depth = position139, tokenIndex139, depth139
			return false
		},
		/* 33 ChainedRef <- <(PathComponent FollowUpRef)> */
		func() bool {
			position145, tokenIndex145, depth145 := position, tokenIndex, depth
			{
				position146 := position
				depth++
				if !_rules[rulePathComponent]() {
					goto l145
				}
				if !_rules[ruleFollowUpRef]() {
					goto l145
				}
				depth--
				add(ruleChainedRef, position146)
			}
			return true
		l145:
			position, tokenIndex, depth = position145, tokenIndex145, depth145
			return false
		},
		/* 34 ChainedDynRef <- <('.'? '[' Expression ']')> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				{
					position149, tokenIndex149, depth149 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l149
					}
					position++
					goto l150
				l149:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
				}
			l150:
				if buffer[position] != rune('[') {
					goto l147
				}
				position++
				if !_rules[ruleExpression]() {
					goto l147
				}
				if buffer[position] != rune(']') {
					goto l147
				}
				position++
				depth--
				add(ruleChainedDynRef, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 35 Slice <- <Range> */
		func() bool {
			position151, tokenIndex151, depth151 := position, tokenIndex, depth
			{
				position152 := position
				depth++
				if !_rules[ruleRange]() {
					goto l151
				}
				depth--
				add(ruleSlice, position152)
			}
			return true
		l151:
			position, tokenIndex, depth = position151, tokenIndex151, depth151
			return false
		},
		/* 36 ChainedCall <- <(StartArguments ExpressionList? ')')> */
		func() bool {
			position153, tokenIndex153, depth153 := position, tokenIndex, depth
			{
				position154 := position
				depth++
				if !_rules[ruleStartArguments]() {
					goto l153
				}
				{
					position155, tokenIndex155, depth155 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l155
					}
					goto l156
				l155:
					position, tokenIndex, depth = position155, tokenIndex155, depth155
				}
			l156:
				if buffer[position] != rune(')') {
					goto l153
				}
				position++
				depth--
				add(ruleChainedCall, position154)
			}
			return true
		l153:
			position, tokenIndex, depth = position153, tokenIndex153, depth153
			return false
		},
		/* 37 StartArguments <- <('(' ws)> */
		func() bool {
			position157, tokenIndex157, depth157 := position, tokenIndex, depth
			{
				position158 := position
				depth++
				if buffer[position] != rune('(') {
					goto l157
				}
				position++
				if !_rules[rulews]() {
					goto l157
				}
				depth--
				add(ruleStartArguments, position158)
			}
			return true
		l157:
			position, tokenIndex, depth = position157, tokenIndex157, depth157
			return false
		},
		/* 38 ExpressionList <- <(NextExpression (',' NextExpression)*)> */
		func() bool {
			position159, tokenIndex159, depth159 := position, tokenIndex, depth
			{
				position160 := position
				depth++
				if !_rules[ruleNextExpression]() {
					goto l159
				}
			l161:
				{
					position162, tokenIndex162, depth162 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l162
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l162
					}
					goto l161
				l162:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
				}
				depth--
				add(ruleExpressionList, position160)
			}
			return true
		l159:
			position, tokenIndex, depth = position159, tokenIndex159, depth159
			return false
		},
		/* 39 NextExpression <- <Expression> */
		func() bool {
			position163, tokenIndex163, depth163 := position, tokenIndex, depth
			{
				position164 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l163
				}
				depth--
				add(ruleNextExpression, position164)
			}
			return true
		l163:
			position, tokenIndex, depth = position163, tokenIndex163, depth163
			return false
		},
		/* 40 Projection <- <('.'? (('[' '*' ']') / Slice) ProjectionValue ChainedQualifiedExpression*)> */
		func() bool {
			position165, tokenIndex165, depth165 := position, tokenIndex, depth
			{
				position166 := position
				depth++
				{
					position167, tokenIndex167, depth167 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l167
					}
					position++
					goto l168
				l167:
					position, tokenIndex, depth = position167, tokenIndex167, depth167
				}
			l168:
				{
					position169, tokenIndex169, depth169 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l170
					}
					position++
					if buffer[position] != rune('*') {
						goto l170
					}
					position++
					if buffer[position] != rune(']') {
						goto l170
					}
					position++
					goto l169
				l170:
					position, tokenIndex, depth = position169, tokenIndex169, depth169
					if !_rules[ruleSlice]() {
						goto l165
					}
				}
			l169:
				if !_rules[ruleProjectionValue]() {
					goto l165
				}
			l171:
				{
					position172, tokenIndex172, depth172 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l172
					}
					goto l171
				l172:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
				}
				depth--
				add(ruleProjection, position166)
			}
			return true
		l165:
			position, tokenIndex, depth = position165, tokenIndex165, depth165
			return false
		},
		/* 41 ProjectionValue <- <Action0> */
		func() bool {
			position173, tokenIndex173, depth173 := position, tokenIndex, depth
			{
				position174 := position
				depth++
				if !_rules[ruleAction0]() {
					goto l173
				}
				depth--
				add(ruleProjectionValue, position174)
			}
			return true
		l173:
			position, tokenIndex, depth = position173, tokenIndex173, depth173
			return false
		},
		/* 42 Substitution <- <('*' Level0)> */
		func() bool {
			position175, tokenIndex175, depth175 := position, tokenIndex, depth
			{
				position176 := position
				depth++
				if buffer[position] != rune('*') {
					goto l175
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l175
				}
				depth--
				add(ruleSubstitution, position176)
			}
			return true
		l175:
			position, tokenIndex, depth = position175, tokenIndex175, depth175
			return false
		},
		/* 43 Not <- <('!' ws Level0)> */
		func() bool {
			position177, tokenIndex177, depth177 := position, tokenIndex, depth
			{
				position178 := position
				depth++
				if buffer[position] != rune('!') {
					goto l177
				}
				position++
				if !_rules[rulews]() {
					goto l177
				}
				if !_rules[ruleLevel0]() {
					goto l177
				}
				depth--
				add(ruleNot, position178)
			}
			return true
		l177:
			position, tokenIndex, depth = position177, tokenIndex177, depth177
			return false
		},
		/* 44 Grouped <- <('(' Expression ')')> */
		func() bool {
			position179, tokenIndex179, depth179 := position, tokenIndex, depth
			{
				position180 := position
				depth++
				if buffer[position] != rune('(') {
					goto l179
				}
				position++
				if !_rules[ruleExpression]() {
					goto l179
				}
				if buffer[position] != rune(')') {
					goto l179
				}
				position++
				depth--
				add(ruleGrouped, position180)
			}
			return true
		l179:
			position, tokenIndex, depth = position179, tokenIndex179, depth179
			return false
		},
		/* 45 Range <- <(StartRange Expression? RangeOp Expression? ']')> */
		func() bool {
			position181, tokenIndex181, depth181 := position, tokenIndex, depth
			{
				position182 := position
				depth++
				if !_rules[ruleStartRange]() {
					goto l181
				}
				{
					position183, tokenIndex183, depth183 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l183
					}
					goto l184
				l183:
					position, tokenIndex, depth = position183, tokenIndex183, depth183
				}
			l184:
				if !_rules[ruleRangeOp]() {
					goto l181
				}
				{
					position185, tokenIndex185, depth185 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l185
					}
					goto l186
				l185:
					position, tokenIndex, depth = position185, tokenIndex185, depth185
				}
			l186:
				if buffer[position] != rune(']') {
					goto l181
				}
				position++
				depth--
				add(ruleRange, position182)
			}
			return true
		l181:
			position, tokenIndex, depth = position181, tokenIndex181, depth181
			return false
		},
		/* 46 StartRange <- <'['> */
		func() bool {
			position187, tokenIndex187, depth187 := position, tokenIndex, depth
			{
				position188 := position
				depth++
				if buffer[position] != rune('[') {
					goto l187
				}
				position++
				depth--
				add(ruleStartRange, position188)
			}
			return true
		l187:
			position, tokenIndex, depth = position187, tokenIndex187, depth187
			return false
		},
		/* 47 RangeOp <- <('.' '.')> */
		func() bool {
			position189, tokenIndex189, depth189 := position, tokenIndex, depth
			{
				position190 := position
				depth++
				if buffer[position] != rune('.') {
					goto l189
				}
				position++
				if buffer[position] != rune('.') {
					goto l189
				}
				position++
				depth--
				add(ruleRangeOp, position190)
			}
			return true
		l189:
			position, tokenIndex, depth = position189, tokenIndex189, depth189
			return false
		},
		/* 48 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position191, tokenIndex191, depth191 := position, tokenIndex, depth
			{
				position192 := position
				depth++
				{
					position193, tokenIndex193, depth193 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l193
					}
					position++
					goto l194
				l193:
					position, tokenIndex, depth = position193, tokenIndex193, depth193
				}
			l194:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l191
				}
				position++
			l195:
				{
					position196, tokenIndex196, depth196 := position, tokenIndex, depth
					{
						position197, tokenIndex197, depth197 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l198
						}
						position++
						goto l197
					l198:
						position, tokenIndex, depth = position197, tokenIndex197, depth197
						if buffer[position] != rune('_') {
							goto l196
						}
						position++
					}
				l197:
					goto l195
				l196:
					position, tokenIndex, depth = position196, tokenIndex196, depth196
				}
				depth--
				add(ruleInteger, position192)
			}
			return true
		l191:
			position, tokenIndex, depth = position191, tokenIndex191, depth191
			return false
		},
		/* 49 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position199, tokenIndex199, depth199 := position, tokenIndex, depth
			{
				position200 := position
				depth++
				if buffer[position] != rune('"') {
					goto l199
				}
				position++
			l201:
				{
					position202, tokenIndex202, depth202 := position, tokenIndex, depth
					{
						position203, tokenIndex203, depth203 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l204
						}
						position++
						if buffer[position] != rune('"') {
							goto l204
						}
						position++
						goto l203
					l204:
						position, tokenIndex, depth = position203, tokenIndex203, depth203
						{
							position205, tokenIndex205, depth205 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l205
							}
							position++
							goto l202
						l205:
							position, tokenIndex, depth = position205, tokenIndex205, depth205
						}
						if !matchDot() {
							goto l202
						}
					}
				l203:
					goto l201
				l202:
					position, tokenIndex, depth = position202, tokenIndex202, depth202
				}
				if buffer[position] != rune('"') {
					goto l199
				}
				position++
				depth--
				add(ruleString, position200)
			}
			return true
		l199:
			position, tokenIndex, depth = position199, tokenIndex199, depth199
			return false
		},
		/* 50 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position206, tokenIndex206, depth206 := position, tokenIndex, depth
			{
				position207 := position
				depth++
				{
					position208, tokenIndex208, depth208 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l209
					}
					position++
					if buffer[position] != rune('r') {
						goto l209
					}
					position++
					if buffer[position] != rune('u') {
						goto l209
					}
					position++
					if buffer[position] != rune('e') {
						goto l209
					}
					position++
					goto l208
				l209:
					position, tokenIndex, depth = position208, tokenIndex208, depth208
					if buffer[position] != rune('f') {
						goto l206
					}
					position++
					if buffer[position] != rune('a') {
						goto l206
					}
					position++
					if buffer[position] != rune('l') {
						goto l206
					}
					position++
					if buffer[position] != rune('s') {
						goto l206
					}
					position++
					if buffer[position] != rune('e') {
						goto l206
					}
					position++
				}
			l208:
				depth--
				add(ruleBoolean, position207)
			}
			return true
		l206:
			position, tokenIndex, depth = position206, tokenIndex206, depth206
			return false
		},
		/* 51 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position210, tokenIndex210, depth210 := position, tokenIndex, depth
			{
				position211 := position
				depth++
				{
					position212, tokenIndex212, depth212 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l213
					}
					position++
					if buffer[position] != rune('i') {
						goto l213
					}
					position++
					if buffer[position] != rune('l') {
						goto l213
					}
					position++
					goto l212
				l213:
					position, tokenIndex, depth = position212, tokenIndex212, depth212
					if buffer[position] != rune('~') {
						goto l210
					}
					position++
				}
			l212:
				depth--
				add(ruleNil, position211)
			}
			return true
		l210:
			position, tokenIndex, depth = position210, tokenIndex210, depth210
			return false
		},
		/* 52 Undefined <- <('~' '~')> */
		func() bool {
			position214, tokenIndex214, depth214 := position, tokenIndex, depth
			{
				position215 := position
				depth++
				if buffer[position] != rune('~') {
					goto l214
				}
				position++
				if buffer[position] != rune('~') {
					goto l214
				}
				position++
				depth--
				add(ruleUndefined, position215)
			}
			return true
		l214:
			position, tokenIndex, depth = position214, tokenIndex214, depth214
			return false
		},
		/* 53 Symbol <- <('$' Name)> */
		func() bool {
			position216, tokenIndex216, depth216 := position, tokenIndex, depth
			{
				position217 := position
				depth++
				if buffer[position] != rune('$') {
					goto l216
				}
				position++
				if !_rules[ruleName]() {
					goto l216
				}
				depth--
				add(ruleSymbol, position217)
			}
			return true
		l216:
			position, tokenIndex, depth = position216, tokenIndex216, depth216
			return false
		},
		/* 54 List <- <(StartList ExpressionList? ']')> */
		func() bool {
			position218, tokenIndex218, depth218 := position, tokenIndex, depth
			{
				position219 := position
				depth++
				if !_rules[ruleStartList]() {
					goto l218
				}
				{
					position220, tokenIndex220, depth220 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l220
					}
					goto l221
				l220:
					position, tokenIndex, depth = position220, tokenIndex220, depth220
				}
			l221:
				if buffer[position] != rune(']') {
					goto l218
				}
				position++
				depth--
				add(ruleList, position219)
			}
			return true
		l218:
			position, tokenIndex, depth = position218, tokenIndex218, depth218
			return false
		},
		/* 55 StartList <- <('[' ws)> */
		func() bool {
			position222, tokenIndex222, depth222 := position, tokenIndex, depth
			{
				position223 := position
				depth++
				if buffer[position] != rune('[') {
					goto l222
				}
				position++
				if !_rules[rulews]() {
					goto l222
				}
				depth--
				add(ruleStartList, position223)
			}
			return true
		l222:
			position, tokenIndex, depth = position222, tokenIndex222, depth222
			return false
		},
		/* 56 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position224, tokenIndex224, depth224 := position, tokenIndex, depth
			{
				position225 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l224
				}
				if !_rules[rulews]() {
					goto l224
				}
				{
					position226, tokenIndex226, depth226 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l226
					}
					goto l227
				l226:
					position, tokenIndex, depth = position226, tokenIndex226, depth226
				}
			l227:
				if buffer[position] != rune('}') {
					goto l224
				}
				position++
				depth--
				add(ruleMap, position225)
			}
			return true
		l224:
			position, tokenIndex, depth = position224, tokenIndex224, depth224
			return false
		},
		/* 57 CreateMap <- <'{'> */
		func() bool {
			position228, tokenIndex228, depth228 := position, tokenIndex, depth
			{
				position229 := position
				depth++
				if buffer[position] != rune('{') {
					goto l228
				}
				position++
				depth--
				add(ruleCreateMap, position229)
			}
			return true
		l228:
			position, tokenIndex, depth = position228, tokenIndex228, depth228
			return false
		},
		/* 58 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position230, tokenIndex230, depth230 := position, tokenIndex, depth
			{
				position231 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l230
				}
			l232:
				{
					position233, tokenIndex233, depth233 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l233
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l233
					}
					goto l232
				l233:
					position, tokenIndex, depth = position233, tokenIndex233, depth233
				}
				depth--
				add(ruleAssignments, position231)
			}
			return true
		l230:
			position, tokenIndex, depth = position230, tokenIndex230, depth230
			return false
		},
		/* 59 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position234, tokenIndex234, depth234 := position, tokenIndex, depth
			{
				position235 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l234
				}
				if buffer[position] != rune('=') {
					goto l234
				}
				position++
				if !_rules[ruleExpression]() {
					goto l234
				}
				depth--
				add(ruleAssignment, position235)
			}
			return true
		l234:
			position, tokenIndex, depth = position234, tokenIndex234, depth234
			return false
		},
		/* 60 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position236, tokenIndex236, depth236 := position, tokenIndex, depth
			{
				position237 := position
				depth++
				{
					position238, tokenIndex238, depth238 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l239
					}
					goto l238
				l239:
					position, tokenIndex, depth = position238, tokenIndex238, depth238
					if !_rules[ruleSimpleMerge]() {
						goto l236
					}
				}
			l238:
				depth--
				add(ruleMerge, position237)
			}
			return true
		l236:
			position, tokenIndex, depth = position236, tokenIndex236, depth236
			return false
		},
		/* 61 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position240, tokenIndex240, depth240 := position, tokenIndex, depth
			{
				position241 := position
				depth++
				if buffer[position] != rune('m') {
					goto l240
				}
				position++
				if buffer[position] != rune('e') {
					goto l240
				}
				position++
				if buffer[position] != rune('r') {
					goto l240
				}
				position++
				if buffer[position] != rune('g') {
					goto l240
				}
				position++
				if buffer[position] != rune('e') {
					goto l240
				}
				position++
				{
					position242, tokenIndex242, depth242 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l242
					}
					if !_rules[ruleRequired]() {
						goto l242
					}
					goto l240
				l242:
					position, tokenIndex, depth = position242, tokenIndex242, depth242
				}
				{
					position243, tokenIndex243, depth243 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l243
					}
					{
						position245, tokenIndex245, depth245 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l246
						}
						goto l245
					l246:
						position, tokenIndex, depth = position245, tokenIndex245, depth245
						if !_rules[ruleOn]() {
							goto l243
						}
					}
				l245:
					goto l244
				l243:
					position, tokenIndex, depth = position243, tokenIndex243, depth243
				}
			l244:
				if !_rules[rulereq_ws]() {
					goto l240
				}
				if !_rules[ruleReference]() {
					goto l240
				}
				depth--
				add(ruleRefMerge, position241)
			}
			return true
		l240:
			position, tokenIndex, depth = position240, tokenIndex240, depth240
			return false
		},
		/* 62 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' !'(' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position247, tokenIndex247, depth247 := position, tokenIndex, depth
			{
				position248 := position
				depth++
				if buffer[position] != rune('m') {
					goto l247
				}
				position++
				if buffer[position] != rune('e') {
					goto l247
				}
				position++
				if buffer[position] != rune('r') {
					goto l247
				}
				position++
				if buffer[position] != rune('g') {
					goto l247
				}
				position++
				if buffer[position] != rune('e') {
					goto l247
				}
				position++
				{
					position249, tokenIndex249, depth249 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l249
					}
					position++
					goto l247
				l249:
					position, tokenIndex, depth = position249, tokenIndex249, depth249
				}
				{
					position250, tokenIndex250, depth250 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l250
					}
					{
						position252, tokenIndex252, depth252 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l253
						}
						goto l252
					l253:
						position, tokenIndex, depth = position252, tokenIndex252, depth252
						if !_rules[ruleRequired]() {
							goto l254
						}
						goto l252
					l254:
						position, tokenIndex, depth = position252, tokenIndex252, depth252
						if !_rules[ruleOn]() {
							goto l250
						}
					}
				l252:
					goto l251
				l250:
					position, tokenIndex, depth = position250, tokenIndex250, depth250
				}
			l251:
				depth--
				add(ruleSimpleMerge, position248)
			}
			return true
		l247:
			position, tokenIndex, depth = position247, tokenIndex247, depth247
			return false
		},
		/* 63 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position255, tokenIndex255, depth255 := position, tokenIndex, depth
			{
				position256 := position
				depth++
				if buffer[position] != rune('r') {
					goto l255
				}
				position++
				if buffer[position] != rune('e') {
					goto l255
				}
				position++
				if buffer[position] != rune('p') {
					goto l255
				}
				position++
				if buffer[position] != rune('l') {
					goto l255
				}
				position++
				if buffer[position] != rune('a') {
					goto l255
				}
				position++
				if buffer[position] != rune('c') {
					goto l255
				}
				position++
				if buffer[position] != rune('e') {
					goto l255
				}
				position++
				depth--
				add(ruleReplace, position256)
			}
			return true
		l255:
			position, tokenIndex, depth = position255, tokenIndex255, depth255
			return false
		},
		/* 64 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position257, tokenIndex257, depth257 := position, tokenIndex, depth
			{
				position258 := position
				depth++
				if buffer[position] != rune('r') {
					goto l257
				}
				position++
				if buffer[position] != rune('e') {
					goto l257
				}
				position++
				if buffer[position] != rune('q') {
					goto l257
				}
				position++
				if buffer[position] != rune('u') {
					goto l257
				}
				position++
				if buffer[position] != rune('i') {
					goto l257
				}
				position++
				if buffer[position] != rune('r') {
					goto l257
				}
				position++
				if buffer[position] != rune('e') {
					goto l257
				}
				position++
				if buffer[position] != rune('d') {
					goto l257
				}
				position++
				depth--
				add(ruleRequired, position258)
			}
			return true
		l257:
			position, tokenIndex, depth = position257, tokenIndex257, depth257
			return false
		},
		/* 65 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position259, tokenIndex259, depth259 := position, tokenIndex, depth
			{
				position260 := position
				depth++
				if buffer[position] != rune('o') {
					goto l259
				}
				position++
				if buffer[position] != rune('n') {
					goto l259
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l259
				}
				if !_rules[ruleName]() {
					goto l259
				}
				depth--
				add(ruleOn, position260)
			}
			return true
		l259:
			position, tokenIndex, depth = position259, tokenIndex259, depth259
			return false
		},
		/* 66 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position261, tokenIndex261, depth261 := position, tokenIndex, depth
			{
				position262 := position
				depth++
				if buffer[position] != rune('a') {
					goto l261
				}
				position++
				if buffer[position] != rune('u') {
					goto l261
				}
				position++
				if buffer[position] != rune('t') {
					goto l261
				}
				position++
				if buffer[position] != rune('o') {
					goto l261
				}
				position++
				depth--
				add(ruleAuto, position262)
			}
			return true
		l261:
			position, tokenIndex, depth = position261, tokenIndex261, depth261
			return false
		},
		/* 67 Default <- <Action1> */
		func() bool {
			position263, tokenIndex263, depth263 := position, tokenIndex, depth
			{
				position264 := position
				depth++
				if !_rules[ruleAction1]() {
					goto l263
				}
				depth--
				add(ruleDefault, position264)
			}
			return true
		l263:
			position, tokenIndex, depth = position263, tokenIndex263, depth263
			return false
		},
		/* 68 Sync <- <('s' 'y' 'n' 'c' '[' Level7 ((((LambdaExpr LambdaExt) / (LambdaOrExpr LambdaOrExpr)) (('|' Expression) / Default)) / (LambdaOrExpr Default Default)) ']')> */
		func() bool {
			position265, tokenIndex265, depth265 := position, tokenIndex, depth
			{
				position266 := position
				depth++
				if buffer[position] != rune('s') {
					goto l265
				}
				position++
				if buffer[position] != rune('y') {
					goto l265
				}
				position++
				if buffer[position] != rune('n') {
					goto l265
				}
				position++
				if buffer[position] != rune('c') {
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
					{
						position269, tokenIndex269, depth269 := position, tokenIndex, depth
						if !_rules[ruleLambdaExpr]() {
							goto l270
						}
						if !_rules[ruleLambdaExt]() {
							goto l270
						}
						goto l269
					l270:
						position, tokenIndex, depth = position269, tokenIndex269, depth269
						if !_rules[ruleLambdaOrExpr]() {
							goto l268
						}
						if !_rules[ruleLambdaOrExpr]() {
							goto l268
						}
					}
				l269:
					{
						position271, tokenIndex271, depth271 := position, tokenIndex, depth
						if buffer[position] != rune('|') {
							goto l272
						}
						position++
						if !_rules[ruleExpression]() {
							goto l272
						}
						goto l271
					l272:
						position, tokenIndex, depth = position271, tokenIndex271, depth271
						if !_rules[ruleDefault]() {
							goto l268
						}
					}
				l271:
					goto l267
				l268:
					position, tokenIndex, depth = position267, tokenIndex267, depth267
					if !_rules[ruleLambdaOrExpr]() {
						goto l265
					}
					if !_rules[ruleDefault]() {
						goto l265
					}
					if !_rules[ruleDefault]() {
						goto l265
					}
				}
			l267:
				if buffer[position] != rune(']') {
					goto l265
				}
				position++
				depth--
				add(ruleSync, position266)
			}
			return true
		l265:
			position, tokenIndex, depth = position265, tokenIndex265, depth265
			return false
		},
		/* 69 LambdaExt <- <(',' Expression)> */
		func() bool {
			position273, tokenIndex273, depth273 := position, tokenIndex, depth
			{
				position274 := position
				depth++
				if buffer[position] != rune(',') {
					goto l273
				}
				position++
				if !_rules[ruleExpression]() {
					goto l273
				}
				depth--
				add(ruleLambdaExt, position274)
			}
			return true
		l273:
			position, tokenIndex, depth = position273, tokenIndex273, depth273
			return false
		},
		/* 70 LambdaOrExpr <- <(LambdaExpr / ('|' Expression))> */
		func() bool {
			position275, tokenIndex275, depth275 := position, tokenIndex, depth
			{
				position276 := position
				depth++
				{
					position277, tokenIndex277, depth277 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l278
					}
					goto l277
				l278:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
					if buffer[position] != rune('|') {
						goto l275
					}
					position++
					if !_rules[ruleExpression]() {
						goto l275
					}
				}
			l277:
				depth--
				add(ruleLambdaOrExpr, position276)
			}
			return true
		l275:
			position, tokenIndex, depth = position275, tokenIndex275, depth275
			return false
		},
		/* 71 Catch <- <('c' 'a' 't' 'c' 'h' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position279, tokenIndex279, depth279 := position, tokenIndex, depth
			{
				position280 := position
				depth++
				if buffer[position] != rune('c') {
					goto l279
				}
				position++
				if buffer[position] != rune('a') {
					goto l279
				}
				position++
				if buffer[position] != rune('t') {
					goto l279
				}
				position++
				if buffer[position] != rune('c') {
					goto l279
				}
				position++
				if buffer[position] != rune('h') {
					goto l279
				}
				position++
				if buffer[position] != rune('[') {
					goto l279
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l279
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l279
				}
				if buffer[position] != rune(']') {
					goto l279
				}
				position++
				depth--
				add(ruleCatch, position280)
			}
			return true
		l279:
			position, tokenIndex, depth = position279, tokenIndex279, depth279
			return false
		},
		/* 72 MapMapping <- <('m' 'a' 'p' '{' Level7 LambdaOrExpr '}')> */
		func() bool {
			position281, tokenIndex281, depth281 := position, tokenIndex, depth
			{
				position282 := position
				depth++
				if buffer[position] != rune('m') {
					goto l281
				}
				position++
				if buffer[position] != rune('a') {
					goto l281
				}
				position++
				if buffer[position] != rune('p') {
					goto l281
				}
				position++
				if buffer[position] != rune('{') {
					goto l281
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l281
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l281
				}
				if buffer[position] != rune('}') {
					goto l281
				}
				position++
				depth--
				add(ruleMapMapping, position282)
			}
			return true
		l281:
			position, tokenIndex, depth = position281, tokenIndex281, depth281
			return false
		},
		/* 73 Mapping <- <('m' 'a' 'p' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position283, tokenIndex283, depth283 := position, tokenIndex, depth
			{
				position284 := position
				depth++
				if buffer[position] != rune('m') {
					goto l283
				}
				position++
				if buffer[position] != rune('a') {
					goto l283
				}
				position++
				if buffer[position] != rune('p') {
					goto l283
				}
				position++
				if buffer[position] != rune('[') {
					goto l283
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l283
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l283
				}
				if buffer[position] != rune(']') {
					goto l283
				}
				position++
				depth--
				add(ruleMapping, position284)
			}
			return true
		l283:
			position, tokenIndex, depth = position283, tokenIndex283, depth283
			return false
		},
		/* 74 MapSelection <- <('s' 'e' 'l' 'e' 'c' 't' '{' Level7 LambdaOrExpr '}')> */
		func() bool {
			position285, tokenIndex285, depth285 := position, tokenIndex, depth
			{
				position286 := position
				depth++
				if buffer[position] != rune('s') {
					goto l285
				}
				position++
				if buffer[position] != rune('e') {
					goto l285
				}
				position++
				if buffer[position] != rune('l') {
					goto l285
				}
				position++
				if buffer[position] != rune('e') {
					goto l285
				}
				position++
				if buffer[position] != rune('c') {
					goto l285
				}
				position++
				if buffer[position] != rune('t') {
					goto l285
				}
				position++
				if buffer[position] != rune('{') {
					goto l285
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l285
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l285
				}
				if buffer[position] != rune('}') {
					goto l285
				}
				position++
				depth--
				add(ruleMapSelection, position286)
			}
			return true
		l285:
			position, tokenIndex, depth = position285, tokenIndex285, depth285
			return false
		},
		/* 75 Selection <- <('s' 'e' 'l' 'e' 'c' 't' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position287, tokenIndex287, depth287 := position, tokenIndex, depth
			{
				position288 := position
				depth++
				if buffer[position] != rune('s') {
					goto l287
				}
				position++
				if buffer[position] != rune('e') {
					goto l287
				}
				position++
				if buffer[position] != rune('l') {
					goto l287
				}
				position++
				if buffer[position] != rune('e') {
					goto l287
				}
				position++
				if buffer[position] != rune('c') {
					goto l287
				}
				position++
				if buffer[position] != rune('t') {
					goto l287
				}
				position++
				if buffer[position] != rune('[') {
					goto l287
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l287
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l287
				}
				if buffer[position] != rune(']') {
					goto l287
				}
				position++
				depth--
				add(ruleSelection, position288)
			}
			return true
		l287:
			position, tokenIndex, depth = position287, tokenIndex287, depth287
			return false
		},
		/* 76 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 LambdaOrExpr ']')> */
		func() bool {
			position289, tokenIndex289, depth289 := position, tokenIndex, depth
			{
				position290 := position
				depth++
				if buffer[position] != rune('s') {
					goto l289
				}
				position++
				if buffer[position] != rune('u') {
					goto l289
				}
				position++
				if buffer[position] != rune('m') {
					goto l289
				}
				position++
				if buffer[position] != rune('[') {
					goto l289
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l289
				}
				if buffer[position] != rune('|') {
					goto l289
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l289
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l289
				}
				if buffer[position] != rune(']') {
					goto l289
				}
				position++
				depth--
				add(ruleSum, position290)
			}
			return true
		l289:
			position, tokenIndex, depth = position289, tokenIndex289, depth289
			return false
		},
		/* 77 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position291, tokenIndex291, depth291 := position, tokenIndex, depth
			{
				position292 := position
				depth++
				if buffer[position] != rune('l') {
					goto l291
				}
				position++
				if buffer[position] != rune('a') {
					goto l291
				}
				position++
				if buffer[position] != rune('m') {
					goto l291
				}
				position++
				if buffer[position] != rune('b') {
					goto l291
				}
				position++
				if buffer[position] != rune('d') {
					goto l291
				}
				position++
				if buffer[position] != rune('a') {
					goto l291
				}
				position++
				{
					position293, tokenIndex293, depth293 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l294
					}
					goto l293
				l294:
					position, tokenIndex, depth = position293, tokenIndex293, depth293
					if !_rules[ruleLambdaExpr]() {
						goto l291
					}
				}
			l293:
				depth--
				add(ruleLambda, position292)
			}
			return true
		l291:
			position, tokenIndex, depth = position291, tokenIndex291, depth291
			return false
		},
		/* 78 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position295, tokenIndex295, depth295 := position, tokenIndex, depth
			{
				position296 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l295
				}
				if !_rules[ruleExpression]() {
					goto l295
				}
				depth--
				add(ruleLambdaRef, position296)
			}
			return true
		l295:
			position, tokenIndex, depth = position295, tokenIndex295, depth295
			return false
		},
		/* 79 LambdaExpr <- <(ws Params ws ('-' '>') Expression)> */
		func() bool {
			position297, tokenIndex297, depth297 := position, tokenIndex, depth
			{
				position298 := position
				depth++
				if !_rules[rulews]() {
					goto l297
				}
				if !_rules[ruleParams]() {
					goto l297
				}
				if !_rules[rulews]() {
					goto l297
				}
				if buffer[position] != rune('-') {
					goto l297
				}
				position++
				if buffer[position] != rune('>') {
					goto l297
				}
				position++
				if !_rules[ruleExpression]() {
					goto l297
				}
				depth--
				add(ruleLambdaExpr, position298)
			}
			return true
		l297:
			position, tokenIndex, depth = position297, tokenIndex297, depth297
			return false
		},
		/* 80 Params <- <('|' StartParams ws Names? ws '|')> */
		func() bool {
			position299, tokenIndex299, depth299 := position, tokenIndex, depth
			{
				position300 := position
				depth++
				if buffer[position] != rune('|') {
					goto l299
				}
				position++
				if !_rules[ruleStartParams]() {
					goto l299
				}
				if !_rules[rulews]() {
					goto l299
				}
				{
					position301, tokenIndex301, depth301 := position, tokenIndex, depth
					if !_rules[ruleNames]() {
						goto l301
					}
					goto l302
				l301:
					position, tokenIndex, depth = position301, tokenIndex301, depth301
				}
			l302:
				if !_rules[rulews]() {
					goto l299
				}
				if buffer[position] != rune('|') {
					goto l299
				}
				position++
				depth--
				add(ruleParams, position300)
			}
			return true
		l299:
			position, tokenIndex, depth = position299, tokenIndex299, depth299
			return false
		},
		/* 81 StartParams <- <Action2> */
		func() bool {
			position303, tokenIndex303, depth303 := position, tokenIndex, depth
			{
				position304 := position
				depth++
				if !_rules[ruleAction2]() {
					goto l303
				}
				depth--
				add(ruleStartParams, position304)
			}
			return true
		l303:
			position, tokenIndex, depth = position303, tokenIndex303, depth303
			return false
		},
		/* 82 Names <- <(NextName (',' NextName)*)> */
		func() bool {
			position305, tokenIndex305, depth305 := position, tokenIndex, depth
			{
				position306 := position
				depth++
				if !_rules[ruleNextName]() {
					goto l305
				}
			l307:
				{
					position308, tokenIndex308, depth308 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l308
					}
					position++
					if !_rules[ruleNextName]() {
						goto l308
					}
					goto l307
				l308:
					position, tokenIndex, depth = position308, tokenIndex308, depth308
				}
				depth--
				add(ruleNames, position306)
			}
			return true
		l305:
			position, tokenIndex, depth = position305, tokenIndex305, depth305
			return false
		},
		/* 83 NextName <- <(ws Name ws)> */
		func() bool {
			position309, tokenIndex309, depth309 := position, tokenIndex, depth
			{
				position310 := position
				depth++
				if !_rules[rulews]() {
					goto l309
				}
				if !_rules[ruleName]() {
					goto l309
				}
				if !_rules[rulews]() {
					goto l309
				}
				depth--
				add(ruleNextName, position310)
			}
			return true
		l309:
			position, tokenIndex, depth = position309, tokenIndex309, depth309
			return false
		},
		/* 84 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position311, tokenIndex311, depth311 := position, tokenIndex, depth
			{
				position312 := position
				depth++
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
						goto l311
					}
					position++
				}
			l315:
			l313:
				{
					position314, tokenIndex314, depth314 := position, tokenIndex, depth
					{
						position319, tokenIndex319, depth319 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l320
						}
						position++
						goto l319
					l320:
						position, tokenIndex, depth = position319, tokenIndex319, depth319
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l321
						}
						position++
						goto l319
					l321:
						position, tokenIndex, depth = position319, tokenIndex319, depth319
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l322
						}
						position++
						goto l319
					l322:
						position, tokenIndex, depth = position319, tokenIndex319, depth319
						if buffer[position] != rune('_') {
							goto l314
						}
						position++
					}
				l319:
					goto l313
				l314:
					position, tokenIndex, depth = position314, tokenIndex314, depth314
				}
				depth--
				add(ruleName, position312)
			}
			return true
		l311:
			position, tokenIndex, depth = position311, tokenIndex311, depth311
			return false
		},
		/* 85 Reference <- <('.'? Key FollowUpRef)> */
		func() bool {
			position323, tokenIndex323, depth323 := position, tokenIndex, depth
			{
				position324 := position
				depth++
				{
					position325, tokenIndex325, depth325 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l325
					}
					position++
					goto l326
				l325:
					position, tokenIndex, depth = position325, tokenIndex325, depth325
				}
			l326:
				if !_rules[ruleKey]() {
					goto l323
				}
				if !_rules[ruleFollowUpRef]() {
					goto l323
				}
				depth--
				add(ruleReference, position324)
			}
			return true
		l323:
			position, tokenIndex, depth = position323, tokenIndex323, depth323
			return false
		},
		/* 86 FollowUpRef <- <PathComponent*> */
		func() bool {
			{
				position328 := position
				depth++
			l329:
				{
					position330, tokenIndex330, depth330 := position, tokenIndex, depth
					if !_rules[rulePathComponent]() {
						goto l330
					}
					goto l329
				l330:
					position, tokenIndex, depth = position330, tokenIndex330, depth330
				}
				depth--
				add(ruleFollowUpRef, position328)
			}
			return true
		},
		/* 87 PathComponent <- <(('.' Key) / ('.'? Index))> */
		func() bool {
			position331, tokenIndex331, depth331 := position, tokenIndex, depth
			{
				position332 := position
				depth++
				{
					position333, tokenIndex333, depth333 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l334
					}
					position++
					if !_rules[ruleKey]() {
						goto l334
					}
					goto l333
				l334:
					position, tokenIndex, depth = position333, tokenIndex333, depth333
					{
						position335, tokenIndex335, depth335 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l335
						}
						position++
						goto l336
					l335:
						position, tokenIndex, depth = position335, tokenIndex335, depth335
					}
				l336:
					if !_rules[ruleIndex]() {
						goto l331
					}
				}
			l333:
				depth--
				add(rulePathComponent, position332)
			}
			return true
		l331:
			position, tokenIndex, depth = position331, tokenIndex331, depth331
			return false
		},
		/* 88 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position337, tokenIndex337, depth337 := position, tokenIndex, depth
			{
				position338 := position
				depth++
				{
					position339, tokenIndex339, depth339 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l340
					}
					position++
					goto l339
				l340:
					position, tokenIndex, depth = position339, tokenIndex339, depth339
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l341
					}
					position++
					goto l339
				l341:
					position, tokenIndex, depth = position339, tokenIndex339, depth339
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l342
					}
					position++
					goto l339
				l342:
					position, tokenIndex, depth = position339, tokenIndex339, depth339
					if buffer[position] != rune('_') {
						goto l337
					}
					position++
				}
			l339:
			l343:
				{
					position344, tokenIndex344, depth344 := position, tokenIndex, depth
					{
						position345, tokenIndex345, depth345 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l346
						}
						position++
						goto l345
					l346:
						position, tokenIndex, depth = position345, tokenIndex345, depth345
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l347
						}
						position++
						goto l345
					l347:
						position, tokenIndex, depth = position345, tokenIndex345, depth345
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l348
						}
						position++
						goto l345
					l348:
						position, tokenIndex, depth = position345, tokenIndex345, depth345
						if buffer[position] != rune('_') {
							goto l349
						}
						position++
						goto l345
					l349:
						position, tokenIndex, depth = position345, tokenIndex345, depth345
						if buffer[position] != rune('-') {
							goto l344
						}
						position++
					}
				l345:
					goto l343
				l344:
					position, tokenIndex, depth = position344, tokenIndex344, depth344
				}
				{
					position350, tokenIndex350, depth350 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l350
					}
					position++
					{
						position352, tokenIndex352, depth352 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l353
						}
						position++
						goto l352
					l353:
						position, tokenIndex, depth = position352, tokenIndex352, depth352
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l354
						}
						position++
						goto l352
					l354:
						position, tokenIndex, depth = position352, tokenIndex352, depth352
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l355
						}
						position++
						goto l352
					l355:
						position, tokenIndex, depth = position352, tokenIndex352, depth352
						if buffer[position] != rune('_') {
							goto l350
						}
						position++
					}
				l352:
				l356:
					{
						position357, tokenIndex357, depth357 := position, tokenIndex, depth
						{
							position358, tokenIndex358, depth358 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l359
							}
							position++
							goto l358
						l359:
							position, tokenIndex, depth = position358, tokenIndex358, depth358
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l360
							}
							position++
							goto l358
						l360:
							position, tokenIndex, depth = position358, tokenIndex358, depth358
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l361
							}
							position++
							goto l358
						l361:
							position, tokenIndex, depth = position358, tokenIndex358, depth358
							if buffer[position] != rune('_') {
								goto l362
							}
							position++
							goto l358
						l362:
							position, tokenIndex, depth = position358, tokenIndex358, depth358
							if buffer[position] != rune('-') {
								goto l357
							}
							position++
						}
					l358:
						goto l356
					l357:
						position, tokenIndex, depth = position357, tokenIndex357, depth357
					}
					goto l351
				l350:
					position, tokenIndex, depth = position350, tokenIndex350, depth350
				}
			l351:
				depth--
				add(ruleKey, position338)
			}
			return true
		l337:
			position, tokenIndex, depth = position337, tokenIndex337, depth337
			return false
		},
		/* 89 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position363, tokenIndex363, depth363 := position, tokenIndex, depth
			{
				position364 := position
				depth++
				if buffer[position] != rune('[') {
					goto l363
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l363
				}
				position++
			l365:
				{
					position366, tokenIndex366, depth366 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l366
					}
					position++
					goto l365
				l366:
					position, tokenIndex, depth = position366, tokenIndex366, depth366
				}
				if buffer[position] != rune(']') {
					goto l363
				}
				position++
				depth--
				add(ruleIndex, position364)
			}
			return true
		l363:
			position, tokenIndex, depth = position363, tokenIndex363, depth363
			return false
		},
		/* 90 IP <- <([0-9]+ '.' [0-9]+ '.' [0-9]+ '.' [0-9]+)> */
		func() bool {
			position367, tokenIndex367, depth367 := position, tokenIndex, depth
			{
				position368 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l367
				}
				position++
			l369:
				{
					position370, tokenIndex370, depth370 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l370
					}
					position++
					goto l369
				l370:
					position, tokenIndex, depth = position370, tokenIndex370, depth370
				}
				if buffer[position] != rune('.') {
					goto l367
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l367
				}
				position++
			l371:
				{
					position372, tokenIndex372, depth372 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l372
					}
					position++
					goto l371
				l372:
					position, tokenIndex, depth = position372, tokenIndex372, depth372
				}
				if buffer[position] != rune('.') {
					goto l367
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l367
				}
				position++
			l373:
				{
					position374, tokenIndex374, depth374 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l374
					}
					position++
					goto l373
				l374:
					position, tokenIndex, depth = position374, tokenIndex374, depth374
				}
				if buffer[position] != rune('.') {
					goto l367
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l367
				}
				position++
			l375:
				{
					position376, tokenIndex376, depth376 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l376
					}
					position++
					goto l375
				l376:
					position, tokenIndex, depth = position376, tokenIndex376, depth376
				}
				depth--
				add(ruleIP, position368)
			}
			return true
		l367:
			position, tokenIndex, depth = position367, tokenIndex367, depth367
			return false
		},
		/* 91 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position378 := position
				depth++
			l379:
				{
					position380, tokenIndex380, depth380 := position, tokenIndex, depth
					{
						position381, tokenIndex381, depth381 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l382
						}
						position++
						goto l381
					l382:
						position, tokenIndex, depth = position381, tokenIndex381, depth381
						if buffer[position] != rune('\t') {
							goto l383
						}
						position++
						goto l381
					l383:
						position, tokenIndex, depth = position381, tokenIndex381, depth381
						if buffer[position] != rune('\n') {
							goto l384
						}
						position++
						goto l381
					l384:
						position, tokenIndex, depth = position381, tokenIndex381, depth381
						if buffer[position] != rune('\r') {
							goto l380
						}
						position++
					}
				l381:
					goto l379
				l380:
					position, tokenIndex, depth = position380, tokenIndex380, depth380
				}
				depth--
				add(rulews, position378)
			}
			return true
		},
		/* 92 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position385, tokenIndex385, depth385 := position, tokenIndex, depth
			{
				position386 := position
				depth++
				{
					position389, tokenIndex389, depth389 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l390
					}
					position++
					goto l389
				l390:
					position, tokenIndex, depth = position389, tokenIndex389, depth389
					if buffer[position] != rune('\t') {
						goto l391
					}
					position++
					goto l389
				l391:
					position, tokenIndex, depth = position389, tokenIndex389, depth389
					if buffer[position] != rune('\n') {
						goto l392
					}
					position++
					goto l389
				l392:
					position, tokenIndex, depth = position389, tokenIndex389, depth389
					if buffer[position] != rune('\r') {
						goto l385
					}
					position++
				}
			l389:
			l387:
				{
					position388, tokenIndex388, depth388 := position, tokenIndex, depth
					{
						position393, tokenIndex393, depth393 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l394
						}
						position++
						goto l393
					l394:
						position, tokenIndex, depth = position393, tokenIndex393, depth393
						if buffer[position] != rune('\t') {
							goto l395
						}
						position++
						goto l393
					l395:
						position, tokenIndex, depth = position393, tokenIndex393, depth393
						if buffer[position] != rune('\n') {
							goto l396
						}
						position++
						goto l393
					l396:
						position, tokenIndex, depth = position393, tokenIndex393, depth393
						if buffer[position] != rune('\r') {
							goto l388
						}
						position++
					}
				l393:
					goto l387
				l388:
					position, tokenIndex, depth = position388, tokenIndex388, depth388
				}
				depth--
				add(rulereq_ws, position386)
			}
			return true
		l385:
			position, tokenIndex, depth = position385, tokenIndex385, depth385
			return false
		},
		/* 94 Action0 <- <{}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 95 Action1 <- <{}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 96 Action2 <- <{}> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
	}
	p.rules = _rules
}
