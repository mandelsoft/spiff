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
	ruleCurrying
	ruleChainedCall
	ruleStartArguments
	ruleExpressionList
	ruleNextExpression
	ruleVarArgs
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
	ruleDefaultValue
	ruleVarParams
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
	"Currying",
	"ChainedCall",
	"StartArguments",
	"ExpressionList",
	"NextExpression",
	"VarArgs",
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
	"DefaultValue",
	"VarParams",
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
	rules  [101]func() bool
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
		/* 4 Marker <- <('&' (('t' 'e' 'm' 'p' 'l' 'a' 't' 'e') / ('t' 'e' 'm' 'p' 'o' 'r' 'a' 'r' 'y') / ('l' 'o' 'c' 'a' 'l') / ('i' 'n' 'j' 'e' 'c' 't') / ('s' 't' 'a' 't' 'e') / ('d' 'e' 'f' 'a' 'u' 'l' 't')))> */
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
						goto l23
					}
					position++
					if buffer[position] != rune('t') {
						goto l23
					}
					position++
					if buffer[position] != rune('a') {
						goto l23
					}
					position++
					if buffer[position] != rune('t') {
						goto l23
					}
					position++
					if buffer[position] != rune('e') {
						goto l23
					}
					position++
					goto l18
				l23:
					position, tokenIndex, depth = position18, tokenIndex18, depth18
					if buffer[position] != rune('d') {
						goto l16
					}
					position++
					if buffer[position] != rune('e') {
						goto l16
					}
					position++
					if buffer[position] != rune('f') {
						goto l16
					}
					position++
					if buffer[position] != rune('a') {
						goto l16
					}
					position++
					if buffer[position] != rune('u') {
						goto l16
					}
					position++
					if buffer[position] != rune('l') {
						goto l16
					}
					position++
					if buffer[position] != rune('t') {
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
			position24, tokenIndex24, depth24 := position, tokenIndex, depth
			{
				position25 := position
				depth++
				if !_rules[ruleGrouped]() {
					goto l24
				}
				depth--
				add(ruleMarkerExpression, position25)
			}
			return true
		l24:
			position, tokenIndex, depth = position24, tokenIndex24, depth24
			return false
		},
		/* 6 Expression <- <((Scoped / LambdaExpr / Level7) ws)> */
		func() bool {
			position26, tokenIndex26, depth26 := position, tokenIndex, depth
			{
				position27 := position
				depth++
				{
					position28, tokenIndex28, depth28 := position, tokenIndex, depth
					if !_rules[ruleScoped]() {
						goto l29
					}
					goto l28
				l29:
					position, tokenIndex, depth = position28, tokenIndex28, depth28
					if !_rules[ruleLambdaExpr]() {
						goto l30
					}
					goto l28
				l30:
					position, tokenIndex, depth = position28, tokenIndex28, depth28
					if !_rules[ruleLevel7]() {
						goto l26
					}
				}
			l28:
				if !_rules[rulews]() {
					goto l26
				}
				depth--
				add(ruleExpression, position27)
			}
			return true
		l26:
			position, tokenIndex, depth = position26, tokenIndex26, depth26
			return false
		},
		/* 7 Scoped <- <(ws Scope ws Expression)> */
		func() bool {
			position31, tokenIndex31, depth31 := position, tokenIndex, depth
			{
				position32 := position
				depth++
				if !_rules[rulews]() {
					goto l31
				}
				if !_rules[ruleScope]() {
					goto l31
				}
				if !_rules[rulews]() {
					goto l31
				}
				if !_rules[ruleExpression]() {
					goto l31
				}
				depth--
				add(ruleScoped, position32)
			}
			return true
		l31:
			position, tokenIndex, depth = position31, tokenIndex31, depth31
			return false
		},
		/* 8 Scope <- <(CreateScope ws Assignments? ')')> */
		func() bool {
			position33, tokenIndex33, depth33 := position, tokenIndex, depth
			{
				position34 := position
				depth++
				if !_rules[ruleCreateScope]() {
					goto l33
				}
				if !_rules[rulews]() {
					goto l33
				}
				{
					position35, tokenIndex35, depth35 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l35
					}
					goto l36
				l35:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
				}
			l36:
				if buffer[position] != rune(')') {
					goto l33
				}
				position++
				depth--
				add(ruleScope, position34)
			}
			return true
		l33:
			position, tokenIndex, depth = position33, tokenIndex33, depth33
			return false
		},
		/* 9 CreateScope <- <'('> */
		func() bool {
			position37, tokenIndex37, depth37 := position, tokenIndex, depth
			{
				position38 := position
				depth++
				if buffer[position] != rune('(') {
					goto l37
				}
				position++
				depth--
				add(ruleCreateScope, position38)
			}
			return true
		l37:
			position, tokenIndex, depth = position37, tokenIndex37, depth37
			return false
		},
		/* 10 Level7 <- <(ws Level6 (req_ws Or)*)> */
		func() bool {
			position39, tokenIndex39, depth39 := position, tokenIndex, depth
			{
				position40 := position
				depth++
				if !_rules[rulews]() {
					goto l39
				}
				if !_rules[ruleLevel6]() {
					goto l39
				}
			l41:
				{
					position42, tokenIndex42, depth42 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l42
					}
					if !_rules[ruleOr]() {
						goto l42
					}
					goto l41
				l42:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
				}
				depth--
				add(ruleLevel7, position40)
			}
			return true
		l39:
			position, tokenIndex, depth = position39, tokenIndex39, depth39
			return false
		},
		/* 11 Or <- <(OrOp req_ws Level6)> */
		func() bool {
			position43, tokenIndex43, depth43 := position, tokenIndex, depth
			{
				position44 := position
				depth++
				if !_rules[ruleOrOp]() {
					goto l43
				}
				if !_rules[rulereq_ws]() {
					goto l43
				}
				if !_rules[ruleLevel6]() {
					goto l43
				}
				depth--
				add(ruleOr, position44)
			}
			return true
		l43:
			position, tokenIndex, depth = position43, tokenIndex43, depth43
			return false
		},
		/* 12 OrOp <- <(('|' '|') / ('/' '/'))> */
		func() bool {
			position45, tokenIndex45, depth45 := position, tokenIndex, depth
			{
				position46 := position
				depth++
				{
					position47, tokenIndex47, depth47 := position, tokenIndex, depth
					if buffer[position] != rune('|') {
						goto l48
					}
					position++
					if buffer[position] != rune('|') {
						goto l48
					}
					position++
					goto l47
				l48:
					position, tokenIndex, depth = position47, tokenIndex47, depth47
					if buffer[position] != rune('/') {
						goto l45
					}
					position++
					if buffer[position] != rune('/') {
						goto l45
					}
					position++
				}
			l47:
				depth--
				add(ruleOrOp, position46)
			}
			return true
		l45:
			position, tokenIndex, depth = position45, tokenIndex45, depth45
			return false
		},
		/* 13 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position49, tokenIndex49, depth49 := position, tokenIndex, depth
			{
				position50 := position
				depth++
				{
					position51, tokenIndex51, depth51 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l52
					}
					goto l51
				l52:
					position, tokenIndex, depth = position51, tokenIndex51, depth51
					if !_rules[ruleLevel5]() {
						goto l49
					}
				}
			l51:
				depth--
				add(ruleLevel6, position50)
			}
			return true
		l49:
			position, tokenIndex, depth = position49, tokenIndex49, depth49
			return false
		},
		/* 14 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position53, tokenIndex53, depth53 := position, tokenIndex, depth
			{
				position54 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l53
				}
				if !_rules[rulews]() {
					goto l53
				}
				if buffer[position] != rune('?') {
					goto l53
				}
				position++
				if !_rules[ruleExpression]() {
					goto l53
				}
				if buffer[position] != rune(':') {
					goto l53
				}
				position++
				if !_rules[ruleExpression]() {
					goto l53
				}
				depth--
				add(ruleConditional, position54)
			}
			return true
		l53:
			position, tokenIndex, depth = position53, tokenIndex53, depth53
			return false
		},
		/* 15 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position55, tokenIndex55, depth55 := position, tokenIndex, depth
			{
				position56 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l55
				}
			l57:
				{
					position58, tokenIndex58, depth58 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l58
					}
					goto l57
				l58:
					position, tokenIndex, depth = position58, tokenIndex58, depth58
				}
				depth--
				add(ruleLevel5, position56)
			}
			return true
		l55:
			position, tokenIndex, depth = position55, tokenIndex55, depth55
			return false
		},
		/* 16 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position59, tokenIndex59, depth59 := position, tokenIndex, depth
			{
				position60 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l59
				}
				if !_rules[ruleLevel4]() {
					goto l59
				}
				depth--
				add(ruleConcatenation, position60)
			}
			return true
		l59:
			position, tokenIndex, depth = position59, tokenIndex59, depth59
			return false
		},
		/* 17 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position61, tokenIndex61, depth61 := position, tokenIndex, depth
			{
				position62 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l61
				}
			l63:
				{
					position64, tokenIndex64, depth64 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l64
					}
					{
						position65, tokenIndex65, depth65 := position, tokenIndex, depth
						if !_rules[ruleLogOr]() {
							goto l66
						}
						goto l65
					l66:
						position, tokenIndex, depth = position65, tokenIndex65, depth65
						if !_rules[ruleLogAnd]() {
							goto l64
						}
					}
				l65:
					goto l63
				l64:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
				}
				depth--
				add(ruleLevel4, position62)
			}
			return true
		l61:
			position, tokenIndex, depth = position61, tokenIndex61, depth61
			return false
		},
		/* 18 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position67, tokenIndex67, depth67 := position, tokenIndex, depth
			{
				position68 := position
				depth++
				if buffer[position] != rune('-') {
					goto l67
				}
				position++
				if buffer[position] != rune('o') {
					goto l67
				}
				position++
				if buffer[position] != rune('r') {
					goto l67
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l67
				}
				if !_rules[ruleLevel3]() {
					goto l67
				}
				depth--
				add(ruleLogOr, position68)
			}
			return true
		l67:
			position, tokenIndex, depth = position67, tokenIndex67, depth67
			return false
		},
		/* 19 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position69, tokenIndex69, depth69 := position, tokenIndex, depth
			{
				position70 := position
				depth++
				if buffer[position] != rune('-') {
					goto l69
				}
				position++
				if buffer[position] != rune('a') {
					goto l69
				}
				position++
				if buffer[position] != rune('n') {
					goto l69
				}
				position++
				if buffer[position] != rune('d') {
					goto l69
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l69
				}
				if !_rules[ruleLevel3]() {
					goto l69
				}
				depth--
				add(ruleLogAnd, position70)
			}
			return true
		l69:
			position, tokenIndex, depth = position69, tokenIndex69, depth69
			return false
		},
		/* 20 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position71, tokenIndex71, depth71 := position, tokenIndex, depth
			{
				position72 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l71
				}
			l73:
				{
					position74, tokenIndex74, depth74 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l74
					}
					if !_rules[ruleComparison]() {
						goto l74
					}
					goto l73
				l74:
					position, tokenIndex, depth = position74, tokenIndex74, depth74
				}
				depth--
				add(ruleLevel3, position72)
			}
			return true
		l71:
			position, tokenIndex, depth = position71, tokenIndex71, depth71
			return false
		},
		/* 21 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position75, tokenIndex75, depth75 := position, tokenIndex, depth
			{
				position76 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l75
				}
				if !_rules[rulereq_ws]() {
					goto l75
				}
				if !_rules[ruleLevel2]() {
					goto l75
				}
				depth--
				add(ruleComparison, position76)
			}
			return true
		l75:
			position, tokenIndex, depth = position75, tokenIndex75, depth75
			return false
		},
		/* 22 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position77, tokenIndex77, depth77 := position, tokenIndex, depth
			{
				position78 := position
				depth++
				{
					position79, tokenIndex79, depth79 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l80
					}
					position++
					if buffer[position] != rune('=') {
						goto l80
					}
					position++
					goto l79
				l80:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if buffer[position] != rune('!') {
						goto l81
					}
					position++
					if buffer[position] != rune('=') {
						goto l81
					}
					position++
					goto l79
				l81:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if buffer[position] != rune('<') {
						goto l82
					}
					position++
					if buffer[position] != rune('=') {
						goto l82
					}
					position++
					goto l79
				l82:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if buffer[position] != rune('>') {
						goto l83
					}
					position++
					if buffer[position] != rune('=') {
						goto l83
					}
					position++
					goto l79
				l83:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if buffer[position] != rune('>') {
						goto l84
					}
					position++
					goto l79
				l84:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if buffer[position] != rune('<') {
						goto l85
					}
					position++
					goto l79
				l85:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
					if buffer[position] != rune('>') {
						goto l77
					}
					position++
				}
			l79:
				depth--
				add(ruleCompareOp, position78)
			}
			return true
		l77:
			position, tokenIndex, depth = position77, tokenIndex77, depth77
			return false
		},
		/* 23 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position86, tokenIndex86, depth86 := position, tokenIndex, depth
			{
				position87 := position
				depth++
				if !_rules[ruleLevel1]() {
					goto l86
				}
			l88:
				{
					position89, tokenIndex89, depth89 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l89
					}
					{
						position90, tokenIndex90, depth90 := position, tokenIndex, depth
						if !_rules[ruleAddition]() {
							goto l91
						}
						goto l90
					l91:
						position, tokenIndex, depth = position90, tokenIndex90, depth90
						if !_rules[ruleSubtraction]() {
							goto l89
						}
					}
				l90:
					goto l88
				l89:
					position, tokenIndex, depth = position89, tokenIndex89, depth89
				}
				depth--
				add(ruleLevel2, position87)
			}
			return true
		l86:
			position, tokenIndex, depth = position86, tokenIndex86, depth86
			return false
		},
		/* 24 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position92, tokenIndex92, depth92 := position, tokenIndex, depth
			{
				position93 := position
				depth++
				if buffer[position] != rune('+') {
					goto l92
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l92
				}
				if !_rules[ruleLevel1]() {
					goto l92
				}
				depth--
				add(ruleAddition, position93)
			}
			return true
		l92:
			position, tokenIndex, depth = position92, tokenIndex92, depth92
			return false
		},
		/* 25 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position94, tokenIndex94, depth94 := position, tokenIndex, depth
			{
				position95 := position
				depth++
				if buffer[position] != rune('-') {
					goto l94
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l94
				}
				if !_rules[ruleLevel1]() {
					goto l94
				}
				depth--
				add(ruleSubtraction, position95)
			}
			return true
		l94:
			position, tokenIndex, depth = position94, tokenIndex94, depth94
			return false
		},
		/* 26 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position96, tokenIndex96, depth96 := position, tokenIndex, depth
			{
				position97 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l96
				}
			l98:
				{
					position99, tokenIndex99, depth99 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l99
					}
					{
						position100, tokenIndex100, depth100 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l101
						}
						goto l100
					l101:
						position, tokenIndex, depth = position100, tokenIndex100, depth100
						if !_rules[ruleDivision]() {
							goto l102
						}
						goto l100
					l102:
						position, tokenIndex, depth = position100, tokenIndex100, depth100
						if !_rules[ruleModulo]() {
							goto l99
						}
					}
				l100:
					goto l98
				l99:
					position, tokenIndex, depth = position99, tokenIndex99, depth99
				}
				depth--
				add(ruleLevel1, position97)
			}
			return true
		l96:
			position, tokenIndex, depth = position96, tokenIndex96, depth96
			return false
		},
		/* 27 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position103, tokenIndex103, depth103 := position, tokenIndex, depth
			{
				position104 := position
				depth++
				if buffer[position] != rune('*') {
					goto l103
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l103
				}
				if !_rules[ruleLevel0]() {
					goto l103
				}
				depth--
				add(ruleMultiplication, position104)
			}
			return true
		l103:
			position, tokenIndex, depth = position103, tokenIndex103, depth103
			return false
		},
		/* 28 Division <- <('/' req_ws Level0)> */
		func() bool {
			position105, tokenIndex105, depth105 := position, tokenIndex, depth
			{
				position106 := position
				depth++
				if buffer[position] != rune('/') {
					goto l105
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l105
				}
				if !_rules[ruleLevel0]() {
					goto l105
				}
				depth--
				add(ruleDivision, position106)
			}
			return true
		l105:
			position, tokenIndex, depth = position105, tokenIndex105, depth105
			return false
		},
		/* 29 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position107, tokenIndex107, depth107 := position, tokenIndex, depth
			{
				position108 := position
				depth++
				if buffer[position] != rune('%') {
					goto l107
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l107
				}
				if !_rules[ruleLevel0]() {
					goto l107
				}
				depth--
				add(ruleModulo, position108)
			}
			return true
		l107:
			position, tokenIndex, depth = position107, tokenIndex107, depth107
			return false
		},
		/* 30 Level0 <- <(IP / String / Integer / Boolean / Undefined / Nil / Symbol / Not / Substitution / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position109, tokenIndex109, depth109 := position, tokenIndex, depth
			{
				position110 := position
				depth++
				{
					position111, tokenIndex111, depth111 := position, tokenIndex, depth
					if !_rules[ruleIP]() {
						goto l112
					}
					goto l111
				l112:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleString]() {
						goto l113
					}
					goto l111
				l113:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleInteger]() {
						goto l114
					}
					goto l111
				l114:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleBoolean]() {
						goto l115
					}
					goto l111
				l115:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleUndefined]() {
						goto l116
					}
					goto l111
				l116:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleNil]() {
						goto l117
					}
					goto l111
				l117:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleSymbol]() {
						goto l118
					}
					goto l111
				l118:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleNot]() {
						goto l119
					}
					goto l111
				l119:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleSubstitution]() {
						goto l120
					}
					goto l111
				l120:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleMerge]() {
						goto l121
					}
					goto l111
				l121:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleAuto]() {
						goto l122
					}
					goto l111
				l122:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleLambda]() {
						goto l123
					}
					goto l111
				l123:
					position, tokenIndex, depth = position111, tokenIndex111, depth111
					if !_rules[ruleChained]() {
						goto l109
					}
				}
			l111:
				depth--
				add(ruleLevel0, position110)
			}
			return true
		l109:
			position, tokenIndex, depth = position109, tokenIndex109, depth109
			return false
		},
		/* 31 Chained <- <((MapMapping / Sync / Catch / Mapping / MapSelection / Selection / Sum / List / Map / Range / Grouped / Reference) ChainedQualifiedExpression*)> */
		func() bool {
			position124, tokenIndex124, depth124 := position, tokenIndex, depth
			{
				position125 := position
				depth++
				{
					position126, tokenIndex126, depth126 := position, tokenIndex, depth
					if !_rules[ruleMapMapping]() {
						goto l127
					}
					goto l126
				l127:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
					if !_rules[ruleSync]() {
						goto l128
					}
					goto l126
				l128:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
					if !_rules[ruleCatch]() {
						goto l129
					}
					goto l126
				l129:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
					if !_rules[ruleMapping]() {
						goto l130
					}
					goto l126
				l130:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
					if !_rules[ruleMapSelection]() {
						goto l131
					}
					goto l126
				l131:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
					if !_rules[ruleSelection]() {
						goto l132
					}
					goto l126
				l132:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
					if !_rules[ruleSum]() {
						goto l133
					}
					goto l126
				l133:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
					if !_rules[ruleList]() {
						goto l134
					}
					goto l126
				l134:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
					if !_rules[ruleMap]() {
						goto l135
					}
					goto l126
				l135:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
					if !_rules[ruleRange]() {
						goto l136
					}
					goto l126
				l136:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
					if !_rules[ruleGrouped]() {
						goto l137
					}
					goto l126
				l137:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
					if !_rules[ruleReference]() {
						goto l124
					}
				}
			l126:
			l138:
				{
					position139, tokenIndex139, depth139 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l139
					}
					goto l138
				l139:
					position, tokenIndex, depth = position139, tokenIndex139, depth139
				}
				depth--
				add(ruleChained, position125)
			}
			return true
		l124:
			position, tokenIndex, depth = position124, tokenIndex124, depth124
			return false
		},
		/* 32 ChainedQualifiedExpression <- <(ChainedCall / Currying / ChainedRef / ChainedDynRef / Projection)> */
		func() bool {
			position140, tokenIndex140, depth140 := position, tokenIndex, depth
			{
				position141 := position
				depth++
				{
					position142, tokenIndex142, depth142 := position, tokenIndex, depth
					if !_rules[ruleChainedCall]() {
						goto l143
					}
					goto l142
				l143:
					position, tokenIndex, depth = position142, tokenIndex142, depth142
					if !_rules[ruleCurrying]() {
						goto l144
					}
					goto l142
				l144:
					position, tokenIndex, depth = position142, tokenIndex142, depth142
					if !_rules[ruleChainedRef]() {
						goto l145
					}
					goto l142
				l145:
					position, tokenIndex, depth = position142, tokenIndex142, depth142
					if !_rules[ruleChainedDynRef]() {
						goto l146
					}
					goto l142
				l146:
					position, tokenIndex, depth = position142, tokenIndex142, depth142
					if !_rules[ruleProjection]() {
						goto l140
					}
				}
			l142:
				depth--
				add(ruleChainedQualifiedExpression, position141)
			}
			return true
		l140:
			position, tokenIndex, depth = position140, tokenIndex140, depth140
			return false
		},
		/* 33 ChainedRef <- <(PathComponent FollowUpRef)> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				if !_rules[rulePathComponent]() {
					goto l147
				}
				if !_rules[ruleFollowUpRef]() {
					goto l147
				}
				depth--
				add(ruleChainedRef, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 34 ChainedDynRef <- <('.'? '[' Expression ']')> */
		func() bool {
			position149, tokenIndex149, depth149 := position, tokenIndex, depth
			{
				position150 := position
				depth++
				{
					position151, tokenIndex151, depth151 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l151
					}
					position++
					goto l152
				l151:
					position, tokenIndex, depth = position151, tokenIndex151, depth151
				}
			l152:
				if buffer[position] != rune('[') {
					goto l149
				}
				position++
				if !_rules[ruleExpression]() {
					goto l149
				}
				if buffer[position] != rune(']') {
					goto l149
				}
				position++
				depth--
				add(ruleChainedDynRef, position150)
			}
			return true
		l149:
			position, tokenIndex, depth = position149, tokenIndex149, depth149
			return false
		},
		/* 35 Slice <- <Range> */
		func() bool {
			position153, tokenIndex153, depth153 := position, tokenIndex, depth
			{
				position154 := position
				depth++
				if !_rules[ruleRange]() {
					goto l153
				}
				depth--
				add(ruleSlice, position154)
			}
			return true
		l153:
			position, tokenIndex, depth = position153, tokenIndex153, depth153
			return false
		},
		/* 36 Currying <- <('*' ChainedCall)> */
		func() bool {
			position155, tokenIndex155, depth155 := position, tokenIndex, depth
			{
				position156 := position
				depth++
				if buffer[position] != rune('*') {
					goto l155
				}
				position++
				if !_rules[ruleChainedCall]() {
					goto l155
				}
				depth--
				add(ruleCurrying, position156)
			}
			return true
		l155:
			position, tokenIndex, depth = position155, tokenIndex155, depth155
			return false
		},
		/* 37 ChainedCall <- <(StartArguments ExpressionList? ')')> */
		func() bool {
			position157, tokenIndex157, depth157 := position, tokenIndex, depth
			{
				position158 := position
				depth++
				if !_rules[ruleStartArguments]() {
					goto l157
				}
				{
					position159, tokenIndex159, depth159 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l159
					}
					goto l160
				l159:
					position, tokenIndex, depth = position159, tokenIndex159, depth159
				}
			l160:
				if buffer[position] != rune(')') {
					goto l157
				}
				position++
				depth--
				add(ruleChainedCall, position158)
			}
			return true
		l157:
			position, tokenIndex, depth = position157, tokenIndex157, depth157
			return false
		},
		/* 38 StartArguments <- <('(' ws)> */
		func() bool {
			position161, tokenIndex161, depth161 := position, tokenIndex, depth
			{
				position162 := position
				depth++
				if buffer[position] != rune('(') {
					goto l161
				}
				position++
				if !_rules[rulews]() {
					goto l161
				}
				depth--
				add(ruleStartArguments, position162)
			}
			return true
		l161:
			position, tokenIndex, depth = position161, tokenIndex161, depth161
			return false
		},
		/* 39 ExpressionList <- <(NextExpression (',' NextExpression)*)> */
		func() bool {
			position163, tokenIndex163, depth163 := position, tokenIndex, depth
			{
				position164 := position
				depth++
				if !_rules[ruleNextExpression]() {
					goto l163
				}
			l165:
				{
					position166, tokenIndex166, depth166 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l166
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l166
					}
					goto l165
				l166:
					position, tokenIndex, depth = position166, tokenIndex166, depth166
				}
				depth--
				add(ruleExpressionList, position164)
			}
			return true
		l163:
			position, tokenIndex, depth = position163, tokenIndex163, depth163
			return false
		},
		/* 40 NextExpression <- <(Expression VarArgs?)> */
		func() bool {
			position167, tokenIndex167, depth167 := position, tokenIndex, depth
			{
				position168 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l167
				}
				{
					position169, tokenIndex169, depth169 := position, tokenIndex, depth
					if !_rules[ruleVarArgs]() {
						goto l169
					}
					goto l170
				l169:
					position, tokenIndex, depth = position169, tokenIndex169, depth169
				}
			l170:
				depth--
				add(ruleNextExpression, position168)
			}
			return true
		l167:
			position, tokenIndex, depth = position167, tokenIndex167, depth167
			return false
		},
		/* 41 VarArgs <- <('.' '.' '.' ws)> */
		func() bool {
			position171, tokenIndex171, depth171 := position, tokenIndex, depth
			{
				position172 := position
				depth++
				if buffer[position] != rune('.') {
					goto l171
				}
				position++
				if buffer[position] != rune('.') {
					goto l171
				}
				position++
				if buffer[position] != rune('.') {
					goto l171
				}
				position++
				if !_rules[rulews]() {
					goto l171
				}
				depth--
				add(ruleVarArgs, position172)
			}
			return true
		l171:
			position, tokenIndex, depth = position171, tokenIndex171, depth171
			return false
		},
		/* 42 Projection <- <('.'? (('[' '*' ']') / Slice) ProjectionValue ChainedQualifiedExpression*)> */
		func() bool {
			position173, tokenIndex173, depth173 := position, tokenIndex, depth
			{
				position174 := position
				depth++
				{
					position175, tokenIndex175, depth175 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l175
					}
					position++
					goto l176
				l175:
					position, tokenIndex, depth = position175, tokenIndex175, depth175
				}
			l176:
				{
					position177, tokenIndex177, depth177 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l178
					}
					position++
					if buffer[position] != rune('*') {
						goto l178
					}
					position++
					if buffer[position] != rune(']') {
						goto l178
					}
					position++
					goto l177
				l178:
					position, tokenIndex, depth = position177, tokenIndex177, depth177
					if !_rules[ruleSlice]() {
						goto l173
					}
				}
			l177:
				if !_rules[ruleProjectionValue]() {
					goto l173
				}
			l179:
				{
					position180, tokenIndex180, depth180 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l180
					}
					goto l179
				l180:
					position, tokenIndex, depth = position180, tokenIndex180, depth180
				}
				depth--
				add(ruleProjection, position174)
			}
			return true
		l173:
			position, tokenIndex, depth = position173, tokenIndex173, depth173
			return false
		},
		/* 43 ProjectionValue <- <Action0> */
		func() bool {
			position181, tokenIndex181, depth181 := position, tokenIndex, depth
			{
				position182 := position
				depth++
				if !_rules[ruleAction0]() {
					goto l181
				}
				depth--
				add(ruleProjectionValue, position182)
			}
			return true
		l181:
			position, tokenIndex, depth = position181, tokenIndex181, depth181
			return false
		},
		/* 44 Substitution <- <('*' Level0)> */
		func() bool {
			position183, tokenIndex183, depth183 := position, tokenIndex, depth
			{
				position184 := position
				depth++
				if buffer[position] != rune('*') {
					goto l183
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l183
				}
				depth--
				add(ruleSubstitution, position184)
			}
			return true
		l183:
			position, tokenIndex, depth = position183, tokenIndex183, depth183
			return false
		},
		/* 45 Not <- <('!' ws Level0)> */
		func() bool {
			position185, tokenIndex185, depth185 := position, tokenIndex, depth
			{
				position186 := position
				depth++
				if buffer[position] != rune('!') {
					goto l185
				}
				position++
				if !_rules[rulews]() {
					goto l185
				}
				if !_rules[ruleLevel0]() {
					goto l185
				}
				depth--
				add(ruleNot, position186)
			}
			return true
		l185:
			position, tokenIndex, depth = position185, tokenIndex185, depth185
			return false
		},
		/* 46 Grouped <- <('(' Expression ')')> */
		func() bool {
			position187, tokenIndex187, depth187 := position, tokenIndex, depth
			{
				position188 := position
				depth++
				if buffer[position] != rune('(') {
					goto l187
				}
				position++
				if !_rules[ruleExpression]() {
					goto l187
				}
				if buffer[position] != rune(')') {
					goto l187
				}
				position++
				depth--
				add(ruleGrouped, position188)
			}
			return true
		l187:
			position, tokenIndex, depth = position187, tokenIndex187, depth187
			return false
		},
		/* 47 Range <- <(StartRange Expression? RangeOp Expression? ']')> */
		func() bool {
			position189, tokenIndex189, depth189 := position, tokenIndex, depth
			{
				position190 := position
				depth++
				if !_rules[ruleStartRange]() {
					goto l189
				}
				{
					position191, tokenIndex191, depth191 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l191
					}
					goto l192
				l191:
					position, tokenIndex, depth = position191, tokenIndex191, depth191
				}
			l192:
				if !_rules[ruleRangeOp]() {
					goto l189
				}
				{
					position193, tokenIndex193, depth193 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l193
					}
					goto l194
				l193:
					position, tokenIndex, depth = position193, tokenIndex193, depth193
				}
			l194:
				if buffer[position] != rune(']') {
					goto l189
				}
				position++
				depth--
				add(ruleRange, position190)
			}
			return true
		l189:
			position, tokenIndex, depth = position189, tokenIndex189, depth189
			return false
		},
		/* 48 StartRange <- <'['> */
		func() bool {
			position195, tokenIndex195, depth195 := position, tokenIndex, depth
			{
				position196 := position
				depth++
				if buffer[position] != rune('[') {
					goto l195
				}
				position++
				depth--
				add(ruleStartRange, position196)
			}
			return true
		l195:
			position, tokenIndex, depth = position195, tokenIndex195, depth195
			return false
		},
		/* 49 RangeOp <- <('.' '.')> */
		func() bool {
			position197, tokenIndex197, depth197 := position, tokenIndex, depth
			{
				position198 := position
				depth++
				if buffer[position] != rune('.') {
					goto l197
				}
				position++
				if buffer[position] != rune('.') {
					goto l197
				}
				position++
				depth--
				add(ruleRangeOp, position198)
			}
			return true
		l197:
			position, tokenIndex, depth = position197, tokenIndex197, depth197
			return false
		},
		/* 50 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position199, tokenIndex199, depth199 := position, tokenIndex, depth
			{
				position200 := position
				depth++
				{
					position201, tokenIndex201, depth201 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l201
					}
					position++
					goto l202
				l201:
					position, tokenIndex, depth = position201, tokenIndex201, depth201
				}
			l202:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l199
				}
				position++
			l203:
				{
					position204, tokenIndex204, depth204 := position, tokenIndex, depth
					{
						position205, tokenIndex205, depth205 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l206
						}
						position++
						goto l205
					l206:
						position, tokenIndex, depth = position205, tokenIndex205, depth205
						if buffer[position] != rune('_') {
							goto l204
						}
						position++
					}
				l205:
					goto l203
				l204:
					position, tokenIndex, depth = position204, tokenIndex204, depth204
				}
				depth--
				add(ruleInteger, position200)
			}
			return true
		l199:
			position, tokenIndex, depth = position199, tokenIndex199, depth199
			return false
		},
		/* 51 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position207, tokenIndex207, depth207 := position, tokenIndex, depth
			{
				position208 := position
				depth++
				if buffer[position] != rune('"') {
					goto l207
				}
				position++
			l209:
				{
					position210, tokenIndex210, depth210 := position, tokenIndex, depth
					{
						position211, tokenIndex211, depth211 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l212
						}
						position++
						if buffer[position] != rune('"') {
							goto l212
						}
						position++
						goto l211
					l212:
						position, tokenIndex, depth = position211, tokenIndex211, depth211
						{
							position213, tokenIndex213, depth213 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l213
							}
							position++
							goto l210
						l213:
							position, tokenIndex, depth = position213, tokenIndex213, depth213
						}
						if !matchDot() {
							goto l210
						}
					}
				l211:
					goto l209
				l210:
					position, tokenIndex, depth = position210, tokenIndex210, depth210
				}
				if buffer[position] != rune('"') {
					goto l207
				}
				position++
				depth--
				add(ruleString, position208)
			}
			return true
		l207:
			position, tokenIndex, depth = position207, tokenIndex207, depth207
			return false
		},
		/* 52 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position214, tokenIndex214, depth214 := position, tokenIndex, depth
			{
				position215 := position
				depth++
				{
					position216, tokenIndex216, depth216 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l217
					}
					position++
					if buffer[position] != rune('r') {
						goto l217
					}
					position++
					if buffer[position] != rune('u') {
						goto l217
					}
					position++
					if buffer[position] != rune('e') {
						goto l217
					}
					position++
					goto l216
				l217:
					position, tokenIndex, depth = position216, tokenIndex216, depth216
					if buffer[position] != rune('f') {
						goto l214
					}
					position++
					if buffer[position] != rune('a') {
						goto l214
					}
					position++
					if buffer[position] != rune('l') {
						goto l214
					}
					position++
					if buffer[position] != rune('s') {
						goto l214
					}
					position++
					if buffer[position] != rune('e') {
						goto l214
					}
					position++
				}
			l216:
				depth--
				add(ruleBoolean, position215)
			}
			return true
		l214:
			position, tokenIndex, depth = position214, tokenIndex214, depth214
			return false
		},
		/* 53 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position218, tokenIndex218, depth218 := position, tokenIndex, depth
			{
				position219 := position
				depth++
				{
					position220, tokenIndex220, depth220 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l221
					}
					position++
					if buffer[position] != rune('i') {
						goto l221
					}
					position++
					if buffer[position] != rune('l') {
						goto l221
					}
					position++
					goto l220
				l221:
					position, tokenIndex, depth = position220, tokenIndex220, depth220
					if buffer[position] != rune('~') {
						goto l218
					}
					position++
				}
			l220:
				depth--
				add(ruleNil, position219)
			}
			return true
		l218:
			position, tokenIndex, depth = position218, tokenIndex218, depth218
			return false
		},
		/* 54 Undefined <- <('~' '~')> */
		func() bool {
			position222, tokenIndex222, depth222 := position, tokenIndex, depth
			{
				position223 := position
				depth++
				if buffer[position] != rune('~') {
					goto l222
				}
				position++
				if buffer[position] != rune('~') {
					goto l222
				}
				position++
				depth--
				add(ruleUndefined, position223)
			}
			return true
		l222:
			position, tokenIndex, depth = position222, tokenIndex222, depth222
			return false
		},
		/* 55 Symbol <- <('$' Name)> */
		func() bool {
			position224, tokenIndex224, depth224 := position, tokenIndex, depth
			{
				position225 := position
				depth++
				if buffer[position] != rune('$') {
					goto l224
				}
				position++
				if !_rules[ruleName]() {
					goto l224
				}
				depth--
				add(ruleSymbol, position225)
			}
			return true
		l224:
			position, tokenIndex, depth = position224, tokenIndex224, depth224
			return false
		},
		/* 56 List <- <(StartList ExpressionList? ']')> */
		func() bool {
			position226, tokenIndex226, depth226 := position, tokenIndex, depth
			{
				position227 := position
				depth++
				if !_rules[ruleStartList]() {
					goto l226
				}
				{
					position228, tokenIndex228, depth228 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l228
					}
					goto l229
				l228:
					position, tokenIndex, depth = position228, tokenIndex228, depth228
				}
			l229:
				if buffer[position] != rune(']') {
					goto l226
				}
				position++
				depth--
				add(ruleList, position227)
			}
			return true
		l226:
			position, tokenIndex, depth = position226, tokenIndex226, depth226
			return false
		},
		/* 57 StartList <- <('[' ws)> */
		func() bool {
			position230, tokenIndex230, depth230 := position, tokenIndex, depth
			{
				position231 := position
				depth++
				if buffer[position] != rune('[') {
					goto l230
				}
				position++
				if !_rules[rulews]() {
					goto l230
				}
				depth--
				add(ruleStartList, position231)
			}
			return true
		l230:
			position, tokenIndex, depth = position230, tokenIndex230, depth230
			return false
		},
		/* 58 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position232, tokenIndex232, depth232 := position, tokenIndex, depth
			{
				position233 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l232
				}
				if !_rules[rulews]() {
					goto l232
				}
				{
					position234, tokenIndex234, depth234 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l234
					}
					goto l235
				l234:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
				}
			l235:
				if buffer[position] != rune('}') {
					goto l232
				}
				position++
				depth--
				add(ruleMap, position233)
			}
			return true
		l232:
			position, tokenIndex, depth = position232, tokenIndex232, depth232
			return false
		},
		/* 59 CreateMap <- <'{'> */
		func() bool {
			position236, tokenIndex236, depth236 := position, tokenIndex, depth
			{
				position237 := position
				depth++
				if buffer[position] != rune('{') {
					goto l236
				}
				position++
				depth--
				add(ruleCreateMap, position237)
			}
			return true
		l236:
			position, tokenIndex, depth = position236, tokenIndex236, depth236
			return false
		},
		/* 60 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position238, tokenIndex238, depth238 := position, tokenIndex, depth
			{
				position239 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l238
				}
			l240:
				{
					position241, tokenIndex241, depth241 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l241
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l241
					}
					goto l240
				l241:
					position, tokenIndex, depth = position241, tokenIndex241, depth241
				}
				depth--
				add(ruleAssignments, position239)
			}
			return true
		l238:
			position, tokenIndex, depth = position238, tokenIndex238, depth238
			return false
		},
		/* 61 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position242, tokenIndex242, depth242 := position, tokenIndex, depth
			{
				position243 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l242
				}
				if buffer[position] != rune('=') {
					goto l242
				}
				position++
				if !_rules[ruleExpression]() {
					goto l242
				}
				depth--
				add(ruleAssignment, position243)
			}
			return true
		l242:
			position, tokenIndex, depth = position242, tokenIndex242, depth242
			return false
		},
		/* 62 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position244, tokenIndex244, depth244 := position, tokenIndex, depth
			{
				position245 := position
				depth++
				{
					position246, tokenIndex246, depth246 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l247
					}
					goto l246
				l247:
					position, tokenIndex, depth = position246, tokenIndex246, depth246
					if !_rules[ruleSimpleMerge]() {
						goto l244
					}
				}
			l246:
				depth--
				add(ruleMerge, position245)
			}
			return true
		l244:
			position, tokenIndex, depth = position244, tokenIndex244, depth244
			return false
		},
		/* 63 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position248, tokenIndex248, depth248 := position, tokenIndex, depth
			{
				position249 := position
				depth++
				if buffer[position] != rune('m') {
					goto l248
				}
				position++
				if buffer[position] != rune('e') {
					goto l248
				}
				position++
				if buffer[position] != rune('r') {
					goto l248
				}
				position++
				if buffer[position] != rune('g') {
					goto l248
				}
				position++
				if buffer[position] != rune('e') {
					goto l248
				}
				position++
				{
					position250, tokenIndex250, depth250 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l250
					}
					if !_rules[ruleRequired]() {
						goto l250
					}
					goto l248
				l250:
					position, tokenIndex, depth = position250, tokenIndex250, depth250
				}
				{
					position251, tokenIndex251, depth251 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l251
					}
					{
						position253, tokenIndex253, depth253 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l254
						}
						goto l253
					l254:
						position, tokenIndex, depth = position253, tokenIndex253, depth253
						if !_rules[ruleOn]() {
							goto l251
						}
					}
				l253:
					goto l252
				l251:
					position, tokenIndex, depth = position251, tokenIndex251, depth251
				}
			l252:
				if !_rules[rulereq_ws]() {
					goto l248
				}
				if !_rules[ruleReference]() {
					goto l248
				}
				depth--
				add(ruleRefMerge, position249)
			}
			return true
		l248:
			position, tokenIndex, depth = position248, tokenIndex248, depth248
			return false
		},
		/* 64 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' !'(' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position255, tokenIndex255, depth255 := position, tokenIndex, depth
			{
				position256 := position
				depth++
				if buffer[position] != rune('m') {
					goto l255
				}
				position++
				if buffer[position] != rune('e') {
					goto l255
				}
				position++
				if buffer[position] != rune('r') {
					goto l255
				}
				position++
				if buffer[position] != rune('g') {
					goto l255
				}
				position++
				if buffer[position] != rune('e') {
					goto l255
				}
				position++
				{
					position257, tokenIndex257, depth257 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l257
					}
					position++
					goto l255
				l257:
					position, tokenIndex, depth = position257, tokenIndex257, depth257
				}
				{
					position258, tokenIndex258, depth258 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l258
					}
					{
						position260, tokenIndex260, depth260 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l261
						}
						goto l260
					l261:
						position, tokenIndex, depth = position260, tokenIndex260, depth260
						if !_rules[ruleRequired]() {
							goto l262
						}
						goto l260
					l262:
						position, tokenIndex, depth = position260, tokenIndex260, depth260
						if !_rules[ruleOn]() {
							goto l258
						}
					}
				l260:
					goto l259
				l258:
					position, tokenIndex, depth = position258, tokenIndex258, depth258
				}
			l259:
				depth--
				add(ruleSimpleMerge, position256)
			}
			return true
		l255:
			position, tokenIndex, depth = position255, tokenIndex255, depth255
			return false
		},
		/* 65 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position263, tokenIndex263, depth263 := position, tokenIndex, depth
			{
				position264 := position
				depth++
				if buffer[position] != rune('r') {
					goto l263
				}
				position++
				if buffer[position] != rune('e') {
					goto l263
				}
				position++
				if buffer[position] != rune('p') {
					goto l263
				}
				position++
				if buffer[position] != rune('l') {
					goto l263
				}
				position++
				if buffer[position] != rune('a') {
					goto l263
				}
				position++
				if buffer[position] != rune('c') {
					goto l263
				}
				position++
				if buffer[position] != rune('e') {
					goto l263
				}
				position++
				depth--
				add(ruleReplace, position264)
			}
			return true
		l263:
			position, tokenIndex, depth = position263, tokenIndex263, depth263
			return false
		},
		/* 66 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position265, tokenIndex265, depth265 := position, tokenIndex, depth
			{
				position266 := position
				depth++
				if buffer[position] != rune('r') {
					goto l265
				}
				position++
				if buffer[position] != rune('e') {
					goto l265
				}
				position++
				if buffer[position] != rune('q') {
					goto l265
				}
				position++
				if buffer[position] != rune('u') {
					goto l265
				}
				position++
				if buffer[position] != rune('i') {
					goto l265
				}
				position++
				if buffer[position] != rune('r') {
					goto l265
				}
				position++
				if buffer[position] != rune('e') {
					goto l265
				}
				position++
				if buffer[position] != rune('d') {
					goto l265
				}
				position++
				depth--
				add(ruleRequired, position266)
			}
			return true
		l265:
			position, tokenIndex, depth = position265, tokenIndex265, depth265
			return false
		},
		/* 67 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position267, tokenIndex267, depth267 := position, tokenIndex, depth
			{
				position268 := position
				depth++
				if buffer[position] != rune('o') {
					goto l267
				}
				position++
				if buffer[position] != rune('n') {
					goto l267
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l267
				}
				if !_rules[ruleName]() {
					goto l267
				}
				depth--
				add(ruleOn, position268)
			}
			return true
		l267:
			position, tokenIndex, depth = position267, tokenIndex267, depth267
			return false
		},
		/* 68 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position269, tokenIndex269, depth269 := position, tokenIndex, depth
			{
				position270 := position
				depth++
				if buffer[position] != rune('a') {
					goto l269
				}
				position++
				if buffer[position] != rune('u') {
					goto l269
				}
				position++
				if buffer[position] != rune('t') {
					goto l269
				}
				position++
				if buffer[position] != rune('o') {
					goto l269
				}
				position++
				depth--
				add(ruleAuto, position270)
			}
			return true
		l269:
			position, tokenIndex, depth = position269, tokenIndex269, depth269
			return false
		},
		/* 69 Default <- <Action1> */
		func() bool {
			position271, tokenIndex271, depth271 := position, tokenIndex, depth
			{
				position272 := position
				depth++
				if !_rules[ruleAction1]() {
					goto l271
				}
				depth--
				add(ruleDefault, position272)
			}
			return true
		l271:
			position, tokenIndex, depth = position271, tokenIndex271, depth271
			return false
		},
		/* 70 Sync <- <('s' 'y' 'n' 'c' '[' Level7 ((((LambdaExpr LambdaExt) / (LambdaOrExpr LambdaOrExpr)) (('|' Expression) / Default)) / (LambdaOrExpr Default Default)) ']')> */
		func() bool {
			position273, tokenIndex273, depth273 := position, tokenIndex, depth
			{
				position274 := position
				depth++
				if buffer[position] != rune('s') {
					goto l273
				}
				position++
				if buffer[position] != rune('y') {
					goto l273
				}
				position++
				if buffer[position] != rune('n') {
					goto l273
				}
				position++
				if buffer[position] != rune('c') {
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
					{
						position277, tokenIndex277, depth277 := position, tokenIndex, depth
						if !_rules[ruleLambdaExpr]() {
							goto l278
						}
						if !_rules[ruleLambdaExt]() {
							goto l278
						}
						goto l277
					l278:
						position, tokenIndex, depth = position277, tokenIndex277, depth277
						if !_rules[ruleLambdaOrExpr]() {
							goto l276
						}
						if !_rules[ruleLambdaOrExpr]() {
							goto l276
						}
					}
				l277:
					{
						position279, tokenIndex279, depth279 := position, tokenIndex, depth
						if buffer[position] != rune('|') {
							goto l280
						}
						position++
						if !_rules[ruleExpression]() {
							goto l280
						}
						goto l279
					l280:
						position, tokenIndex, depth = position279, tokenIndex279, depth279
						if !_rules[ruleDefault]() {
							goto l276
						}
					}
				l279:
					goto l275
				l276:
					position, tokenIndex, depth = position275, tokenIndex275, depth275
					if !_rules[ruleLambdaOrExpr]() {
						goto l273
					}
					if !_rules[ruleDefault]() {
						goto l273
					}
					if !_rules[ruleDefault]() {
						goto l273
					}
				}
			l275:
				if buffer[position] != rune(']') {
					goto l273
				}
				position++
				depth--
				add(ruleSync, position274)
			}
			return true
		l273:
			position, tokenIndex, depth = position273, tokenIndex273, depth273
			return false
		},
		/* 71 LambdaExt <- <(',' Expression)> */
		func() bool {
			position281, tokenIndex281, depth281 := position, tokenIndex, depth
			{
				position282 := position
				depth++
				if buffer[position] != rune(',') {
					goto l281
				}
				position++
				if !_rules[ruleExpression]() {
					goto l281
				}
				depth--
				add(ruleLambdaExt, position282)
			}
			return true
		l281:
			position, tokenIndex, depth = position281, tokenIndex281, depth281
			return false
		},
		/* 72 LambdaOrExpr <- <(LambdaExpr / ('|' Expression))> */
		func() bool {
			position283, tokenIndex283, depth283 := position, tokenIndex, depth
			{
				position284 := position
				depth++
				{
					position285, tokenIndex285, depth285 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l286
					}
					goto l285
				l286:
					position, tokenIndex, depth = position285, tokenIndex285, depth285
					if buffer[position] != rune('|') {
						goto l283
					}
					position++
					if !_rules[ruleExpression]() {
						goto l283
					}
				}
			l285:
				depth--
				add(ruleLambdaOrExpr, position284)
			}
			return true
		l283:
			position, tokenIndex, depth = position283, tokenIndex283, depth283
			return false
		},
		/* 73 Catch <- <('c' 'a' 't' 'c' 'h' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position287, tokenIndex287, depth287 := position, tokenIndex, depth
			{
				position288 := position
				depth++
				if buffer[position] != rune('c') {
					goto l287
				}
				position++
				if buffer[position] != rune('a') {
					goto l287
				}
				position++
				if buffer[position] != rune('t') {
					goto l287
				}
				position++
				if buffer[position] != rune('c') {
					goto l287
				}
				position++
				if buffer[position] != rune('h') {
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
				add(ruleCatch, position288)
			}
			return true
		l287:
			position, tokenIndex, depth = position287, tokenIndex287, depth287
			return false
		},
		/* 74 MapMapping <- <('m' 'a' 'p' '{' Level7 LambdaOrExpr '}')> */
		func() bool {
			position289, tokenIndex289, depth289 := position, tokenIndex, depth
			{
				position290 := position
				depth++
				if buffer[position] != rune('m') {
					goto l289
				}
				position++
				if buffer[position] != rune('a') {
					goto l289
				}
				position++
				if buffer[position] != rune('p') {
					goto l289
				}
				position++
				if buffer[position] != rune('{') {
					goto l289
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l289
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l289
				}
				if buffer[position] != rune('}') {
					goto l289
				}
				position++
				depth--
				add(ruleMapMapping, position290)
			}
			return true
		l289:
			position, tokenIndex, depth = position289, tokenIndex289, depth289
			return false
		},
		/* 75 Mapping <- <('m' 'a' 'p' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position291, tokenIndex291, depth291 := position, tokenIndex, depth
			{
				position292 := position
				depth++
				if buffer[position] != rune('m') {
					goto l291
				}
				position++
				if buffer[position] != rune('a') {
					goto l291
				}
				position++
				if buffer[position] != rune('p') {
					goto l291
				}
				position++
				if buffer[position] != rune('[') {
					goto l291
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l291
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l291
				}
				if buffer[position] != rune(']') {
					goto l291
				}
				position++
				depth--
				add(ruleMapping, position292)
			}
			return true
		l291:
			position, tokenIndex, depth = position291, tokenIndex291, depth291
			return false
		},
		/* 76 MapSelection <- <('s' 'e' 'l' 'e' 'c' 't' '{' Level7 LambdaOrExpr '}')> */
		func() bool {
			position293, tokenIndex293, depth293 := position, tokenIndex, depth
			{
				position294 := position
				depth++
				if buffer[position] != rune('s') {
					goto l293
				}
				position++
				if buffer[position] != rune('e') {
					goto l293
				}
				position++
				if buffer[position] != rune('l') {
					goto l293
				}
				position++
				if buffer[position] != rune('e') {
					goto l293
				}
				position++
				if buffer[position] != rune('c') {
					goto l293
				}
				position++
				if buffer[position] != rune('t') {
					goto l293
				}
				position++
				if buffer[position] != rune('{') {
					goto l293
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l293
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l293
				}
				if buffer[position] != rune('}') {
					goto l293
				}
				position++
				depth--
				add(ruleMapSelection, position294)
			}
			return true
		l293:
			position, tokenIndex, depth = position293, tokenIndex293, depth293
			return false
		},
		/* 77 Selection <- <('s' 'e' 'l' 'e' 'c' 't' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position295, tokenIndex295, depth295 := position, tokenIndex, depth
			{
				position296 := position
				depth++
				if buffer[position] != rune('s') {
					goto l295
				}
				position++
				if buffer[position] != rune('e') {
					goto l295
				}
				position++
				if buffer[position] != rune('l') {
					goto l295
				}
				position++
				if buffer[position] != rune('e') {
					goto l295
				}
				position++
				if buffer[position] != rune('c') {
					goto l295
				}
				position++
				if buffer[position] != rune('t') {
					goto l295
				}
				position++
				if buffer[position] != rune('[') {
					goto l295
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l295
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l295
				}
				if buffer[position] != rune(']') {
					goto l295
				}
				position++
				depth--
				add(ruleSelection, position296)
			}
			return true
		l295:
			position, tokenIndex, depth = position295, tokenIndex295, depth295
			return false
		},
		/* 78 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 LambdaOrExpr ']')> */
		func() bool {
			position297, tokenIndex297, depth297 := position, tokenIndex, depth
			{
				position298 := position
				depth++
				if buffer[position] != rune('s') {
					goto l297
				}
				position++
				if buffer[position] != rune('u') {
					goto l297
				}
				position++
				if buffer[position] != rune('m') {
					goto l297
				}
				position++
				if buffer[position] != rune('[') {
					goto l297
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l297
				}
				if buffer[position] != rune('|') {
					goto l297
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l297
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l297
				}
				if buffer[position] != rune(']') {
					goto l297
				}
				position++
				depth--
				add(ruleSum, position298)
			}
			return true
		l297:
			position, tokenIndex, depth = position297, tokenIndex297, depth297
			return false
		},
		/* 79 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position299, tokenIndex299, depth299 := position, tokenIndex, depth
			{
				position300 := position
				depth++
				if buffer[position] != rune('l') {
					goto l299
				}
				position++
				if buffer[position] != rune('a') {
					goto l299
				}
				position++
				if buffer[position] != rune('m') {
					goto l299
				}
				position++
				if buffer[position] != rune('b') {
					goto l299
				}
				position++
				if buffer[position] != rune('d') {
					goto l299
				}
				position++
				if buffer[position] != rune('a') {
					goto l299
				}
				position++
				{
					position301, tokenIndex301, depth301 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l302
					}
					goto l301
				l302:
					position, tokenIndex, depth = position301, tokenIndex301, depth301
					if !_rules[ruleLambdaExpr]() {
						goto l299
					}
				}
			l301:
				depth--
				add(ruleLambda, position300)
			}
			return true
		l299:
			position, tokenIndex, depth = position299, tokenIndex299, depth299
			return false
		},
		/* 80 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position303, tokenIndex303, depth303 := position, tokenIndex, depth
			{
				position304 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l303
				}
				if !_rules[ruleExpression]() {
					goto l303
				}
				depth--
				add(ruleLambdaRef, position304)
			}
			return true
		l303:
			position, tokenIndex, depth = position303, tokenIndex303, depth303
			return false
		},
		/* 81 LambdaExpr <- <(ws Params ws ('-' '>') Expression)> */
		func() bool {
			position305, tokenIndex305, depth305 := position, tokenIndex, depth
			{
				position306 := position
				depth++
				if !_rules[rulews]() {
					goto l305
				}
				if !_rules[ruleParams]() {
					goto l305
				}
				if !_rules[rulews]() {
					goto l305
				}
				if buffer[position] != rune('-') {
					goto l305
				}
				position++
				if buffer[position] != rune('>') {
					goto l305
				}
				position++
				if !_rules[ruleExpression]() {
					goto l305
				}
				depth--
				add(ruleLambdaExpr, position306)
			}
			return true
		l305:
			position, tokenIndex, depth = position305, tokenIndex305, depth305
			return false
		},
		/* 82 Params <- <('|' StartParams ws Names? '|')> */
		func() bool {
			position307, tokenIndex307, depth307 := position, tokenIndex, depth
			{
				position308 := position
				depth++
				if buffer[position] != rune('|') {
					goto l307
				}
				position++
				if !_rules[ruleStartParams]() {
					goto l307
				}
				if !_rules[rulews]() {
					goto l307
				}
				{
					position309, tokenIndex309, depth309 := position, tokenIndex, depth
					if !_rules[ruleNames]() {
						goto l309
					}
					goto l310
				l309:
					position, tokenIndex, depth = position309, tokenIndex309, depth309
				}
			l310:
				if buffer[position] != rune('|') {
					goto l307
				}
				position++
				depth--
				add(ruleParams, position308)
			}
			return true
		l307:
			position, tokenIndex, depth = position307, tokenIndex307, depth307
			return false
		},
		/* 83 StartParams <- <Action2> */
		func() bool {
			position311, tokenIndex311, depth311 := position, tokenIndex, depth
			{
				position312 := position
				depth++
				if !_rules[ruleAction2]() {
					goto l311
				}
				depth--
				add(ruleStartParams, position312)
			}
			return true
		l311:
			position, tokenIndex, depth = position311, tokenIndex311, depth311
			return false
		},
		/* 84 Names <- <(NextName (',' NextName)* DefaultValue? (',' NextName DefaultValue)* VarParams?)> */
		func() bool {
			position313, tokenIndex313, depth313 := position, tokenIndex, depth
			{
				position314 := position
				depth++
				if !_rules[ruleNextName]() {
					goto l313
				}
			l315:
				{
					position316, tokenIndex316, depth316 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l316
					}
					position++
					if !_rules[ruleNextName]() {
						goto l316
					}
					goto l315
				l316:
					position, tokenIndex, depth = position316, tokenIndex316, depth316
				}
				{
					position317, tokenIndex317, depth317 := position, tokenIndex, depth
					if !_rules[ruleDefaultValue]() {
						goto l317
					}
					goto l318
				l317:
					position, tokenIndex, depth = position317, tokenIndex317, depth317
				}
			l318:
			l319:
				{
					position320, tokenIndex320, depth320 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l320
					}
					position++
					if !_rules[ruleNextName]() {
						goto l320
					}
					if !_rules[ruleDefaultValue]() {
						goto l320
					}
					goto l319
				l320:
					position, tokenIndex, depth = position320, tokenIndex320, depth320
				}
				{
					position321, tokenIndex321, depth321 := position, tokenIndex, depth
					if !_rules[ruleVarParams]() {
						goto l321
					}
					goto l322
				l321:
					position, tokenIndex, depth = position321, tokenIndex321, depth321
				}
			l322:
				depth--
				add(ruleNames, position314)
			}
			return true
		l313:
			position, tokenIndex, depth = position313, tokenIndex313, depth313
			return false
		},
		/* 85 NextName <- <(ws Name ws)> */
		func() bool {
			position323, tokenIndex323, depth323 := position, tokenIndex, depth
			{
				position324 := position
				depth++
				if !_rules[rulews]() {
					goto l323
				}
				if !_rules[ruleName]() {
					goto l323
				}
				if !_rules[rulews]() {
					goto l323
				}
				depth--
				add(ruleNextName, position324)
			}
			return true
		l323:
			position, tokenIndex, depth = position323, tokenIndex323, depth323
			return false
		},
		/* 86 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position325, tokenIndex325, depth325 := position, tokenIndex, depth
			{
				position326 := position
				depth++
				{
					position329, tokenIndex329, depth329 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l330
					}
					position++
					goto l329
				l330:
					position, tokenIndex, depth = position329, tokenIndex329, depth329
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l331
					}
					position++
					goto l329
				l331:
					position, tokenIndex, depth = position329, tokenIndex329, depth329
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l332
					}
					position++
					goto l329
				l332:
					position, tokenIndex, depth = position329, tokenIndex329, depth329
					if buffer[position] != rune('_') {
						goto l325
					}
					position++
				}
			l329:
			l327:
				{
					position328, tokenIndex328, depth328 := position, tokenIndex, depth
					{
						position333, tokenIndex333, depth333 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l334
						}
						position++
						goto l333
					l334:
						position, tokenIndex, depth = position333, tokenIndex333, depth333
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l335
						}
						position++
						goto l333
					l335:
						position, tokenIndex, depth = position333, tokenIndex333, depth333
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l336
						}
						position++
						goto l333
					l336:
						position, tokenIndex, depth = position333, tokenIndex333, depth333
						if buffer[position] != rune('_') {
							goto l328
						}
						position++
					}
				l333:
					goto l327
				l328:
					position, tokenIndex, depth = position328, tokenIndex328, depth328
				}
				depth--
				add(ruleName, position326)
			}
			return true
		l325:
			position, tokenIndex, depth = position325, tokenIndex325, depth325
			return false
		},
		/* 87 DefaultValue <- <('=' Expression)> */
		func() bool {
			position337, tokenIndex337, depth337 := position, tokenIndex, depth
			{
				position338 := position
				depth++
				if buffer[position] != rune('=') {
					goto l337
				}
				position++
				if !_rules[ruleExpression]() {
					goto l337
				}
				depth--
				add(ruleDefaultValue, position338)
			}
			return true
		l337:
			position, tokenIndex, depth = position337, tokenIndex337, depth337
			return false
		},
		/* 88 VarParams <- <('.' '.' '.' ws)> */
		func() bool {
			position339, tokenIndex339, depth339 := position, tokenIndex, depth
			{
				position340 := position
				depth++
				if buffer[position] != rune('.') {
					goto l339
				}
				position++
				if buffer[position] != rune('.') {
					goto l339
				}
				position++
				if buffer[position] != rune('.') {
					goto l339
				}
				position++
				if !_rules[rulews]() {
					goto l339
				}
				depth--
				add(ruleVarParams, position340)
			}
			return true
		l339:
			position, tokenIndex, depth = position339, tokenIndex339, depth339
			return false
		},
		/* 89 Reference <- <('.'? Key FollowUpRef)> */
		func() bool {
			position341, tokenIndex341, depth341 := position, tokenIndex, depth
			{
				position342 := position
				depth++
				{
					position343, tokenIndex343, depth343 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l343
					}
					position++
					goto l344
				l343:
					position, tokenIndex, depth = position343, tokenIndex343, depth343
				}
			l344:
				if !_rules[ruleKey]() {
					goto l341
				}
				if !_rules[ruleFollowUpRef]() {
					goto l341
				}
				depth--
				add(ruleReference, position342)
			}
			return true
		l341:
			position, tokenIndex, depth = position341, tokenIndex341, depth341
			return false
		},
		/* 90 FollowUpRef <- <PathComponent*> */
		func() bool {
			{
				position346 := position
				depth++
			l347:
				{
					position348, tokenIndex348, depth348 := position, tokenIndex, depth
					if !_rules[rulePathComponent]() {
						goto l348
					}
					goto l347
				l348:
					position, tokenIndex, depth = position348, tokenIndex348, depth348
				}
				depth--
				add(ruleFollowUpRef, position346)
			}
			return true
		},
		/* 91 PathComponent <- <(('.' Key) / ('.'? Index))> */
		func() bool {
			position349, tokenIndex349, depth349 := position, tokenIndex, depth
			{
				position350 := position
				depth++
				{
					position351, tokenIndex351, depth351 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l352
					}
					position++
					if !_rules[ruleKey]() {
						goto l352
					}
					goto l351
				l352:
					position, tokenIndex, depth = position351, tokenIndex351, depth351
					{
						position353, tokenIndex353, depth353 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l353
						}
						position++
						goto l354
					l353:
						position, tokenIndex, depth = position353, tokenIndex353, depth353
					}
				l354:
					if !_rules[ruleIndex]() {
						goto l349
					}
				}
			l351:
				depth--
				add(rulePathComponent, position350)
			}
			return true
		l349:
			position, tokenIndex, depth = position349, tokenIndex349, depth349
			return false
		},
		/* 92 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position355, tokenIndex355, depth355 := position, tokenIndex, depth
			{
				position356 := position
				depth++
				{
					position357, tokenIndex357, depth357 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l358
					}
					position++
					goto l357
				l358:
					position, tokenIndex, depth = position357, tokenIndex357, depth357
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l359
					}
					position++
					goto l357
				l359:
					position, tokenIndex, depth = position357, tokenIndex357, depth357
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l360
					}
					position++
					goto l357
				l360:
					position, tokenIndex, depth = position357, tokenIndex357, depth357
					if buffer[position] != rune('_') {
						goto l355
					}
					position++
				}
			l357:
			l361:
				{
					position362, tokenIndex362, depth362 := position, tokenIndex, depth
					{
						position363, tokenIndex363, depth363 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l364
						}
						position++
						goto l363
					l364:
						position, tokenIndex, depth = position363, tokenIndex363, depth363
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l365
						}
						position++
						goto l363
					l365:
						position, tokenIndex, depth = position363, tokenIndex363, depth363
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l366
						}
						position++
						goto l363
					l366:
						position, tokenIndex, depth = position363, tokenIndex363, depth363
						if buffer[position] != rune('_') {
							goto l367
						}
						position++
						goto l363
					l367:
						position, tokenIndex, depth = position363, tokenIndex363, depth363
						if buffer[position] != rune('-') {
							goto l362
						}
						position++
					}
				l363:
					goto l361
				l362:
					position, tokenIndex, depth = position362, tokenIndex362, depth362
				}
				{
					position368, tokenIndex368, depth368 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l368
					}
					position++
					{
						position370, tokenIndex370, depth370 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l371
						}
						position++
						goto l370
					l371:
						position, tokenIndex, depth = position370, tokenIndex370, depth370
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l372
						}
						position++
						goto l370
					l372:
						position, tokenIndex, depth = position370, tokenIndex370, depth370
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l373
						}
						position++
						goto l370
					l373:
						position, tokenIndex, depth = position370, tokenIndex370, depth370
						if buffer[position] != rune('_') {
							goto l368
						}
						position++
					}
				l370:
				l374:
					{
						position375, tokenIndex375, depth375 := position, tokenIndex, depth
						{
							position376, tokenIndex376, depth376 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l377
							}
							position++
							goto l376
						l377:
							position, tokenIndex, depth = position376, tokenIndex376, depth376
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l378
							}
							position++
							goto l376
						l378:
							position, tokenIndex, depth = position376, tokenIndex376, depth376
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l379
							}
							position++
							goto l376
						l379:
							position, tokenIndex, depth = position376, tokenIndex376, depth376
							if buffer[position] != rune('_') {
								goto l380
							}
							position++
							goto l376
						l380:
							position, tokenIndex, depth = position376, tokenIndex376, depth376
							if buffer[position] != rune('-') {
								goto l375
							}
							position++
						}
					l376:
						goto l374
					l375:
						position, tokenIndex, depth = position375, tokenIndex375, depth375
					}
					goto l369
				l368:
					position, tokenIndex, depth = position368, tokenIndex368, depth368
				}
			l369:
				depth--
				add(ruleKey, position356)
			}
			return true
		l355:
			position, tokenIndex, depth = position355, tokenIndex355, depth355
			return false
		},
		/* 93 Index <- <('[' '-'? [0-9]+ ']')> */
		func() bool {
			position381, tokenIndex381, depth381 := position, tokenIndex, depth
			{
				position382 := position
				depth++
				if buffer[position] != rune('[') {
					goto l381
				}
				position++
				{
					position383, tokenIndex383, depth383 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l383
					}
					position++
					goto l384
				l383:
					position, tokenIndex, depth = position383, tokenIndex383, depth383
				}
			l384:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l381
				}
				position++
			l385:
				{
					position386, tokenIndex386, depth386 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l386
					}
					position++
					goto l385
				l386:
					position, tokenIndex, depth = position386, tokenIndex386, depth386
				}
				if buffer[position] != rune(']') {
					goto l381
				}
				position++
				depth--
				add(ruleIndex, position382)
			}
			return true
		l381:
			position, tokenIndex, depth = position381, tokenIndex381, depth381
			return false
		},
		/* 94 IP <- <([0-9]+ '.' [0-9]+ '.' [0-9]+ '.' [0-9]+)> */
		func() bool {
			position387, tokenIndex387, depth387 := position, tokenIndex, depth
			{
				position388 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l387
				}
				position++
			l389:
				{
					position390, tokenIndex390, depth390 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l390
					}
					position++
					goto l389
				l390:
					position, tokenIndex, depth = position390, tokenIndex390, depth390
				}
				if buffer[position] != rune('.') {
					goto l387
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l387
				}
				position++
			l391:
				{
					position392, tokenIndex392, depth392 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l392
					}
					position++
					goto l391
				l392:
					position, tokenIndex, depth = position392, tokenIndex392, depth392
				}
				if buffer[position] != rune('.') {
					goto l387
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l387
				}
				position++
			l393:
				{
					position394, tokenIndex394, depth394 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l394
					}
					position++
					goto l393
				l394:
					position, tokenIndex, depth = position394, tokenIndex394, depth394
				}
				if buffer[position] != rune('.') {
					goto l387
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l387
				}
				position++
			l395:
				{
					position396, tokenIndex396, depth396 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l396
					}
					position++
					goto l395
				l396:
					position, tokenIndex, depth = position396, tokenIndex396, depth396
				}
				depth--
				add(ruleIP, position388)
			}
			return true
		l387:
			position, tokenIndex, depth = position387, tokenIndex387, depth387
			return false
		},
		/* 95 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position398 := position
				depth++
			l399:
				{
					position400, tokenIndex400, depth400 := position, tokenIndex, depth
					{
						position401, tokenIndex401, depth401 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l402
						}
						position++
						goto l401
					l402:
						position, tokenIndex, depth = position401, tokenIndex401, depth401
						if buffer[position] != rune('\t') {
							goto l403
						}
						position++
						goto l401
					l403:
						position, tokenIndex, depth = position401, tokenIndex401, depth401
						if buffer[position] != rune('\n') {
							goto l404
						}
						position++
						goto l401
					l404:
						position, tokenIndex, depth = position401, tokenIndex401, depth401
						if buffer[position] != rune('\r') {
							goto l400
						}
						position++
					}
				l401:
					goto l399
				l400:
					position, tokenIndex, depth = position400, tokenIndex400, depth400
				}
				depth--
				add(rulews, position398)
			}
			return true
		},
		/* 96 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position405, tokenIndex405, depth405 := position, tokenIndex, depth
			{
				position406 := position
				depth++
				{
					position409, tokenIndex409, depth409 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l410
					}
					position++
					goto l409
				l410:
					position, tokenIndex, depth = position409, tokenIndex409, depth409
					if buffer[position] != rune('\t') {
						goto l411
					}
					position++
					goto l409
				l411:
					position, tokenIndex, depth = position409, tokenIndex409, depth409
					if buffer[position] != rune('\n') {
						goto l412
					}
					position++
					goto l409
				l412:
					position, tokenIndex, depth = position409, tokenIndex409, depth409
					if buffer[position] != rune('\r') {
						goto l405
					}
					position++
				}
			l409:
			l407:
				{
					position408, tokenIndex408, depth408 := position, tokenIndex, depth
					{
						position413, tokenIndex413, depth413 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l414
						}
						position++
						goto l413
					l414:
						position, tokenIndex, depth = position413, tokenIndex413, depth413
						if buffer[position] != rune('\t') {
							goto l415
						}
						position++
						goto l413
					l415:
						position, tokenIndex, depth = position413, tokenIndex413, depth413
						if buffer[position] != rune('\n') {
							goto l416
						}
						position++
						goto l413
					l416:
						position, tokenIndex, depth = position413, tokenIndex413, depth413
						if buffer[position] != rune('\r') {
							goto l408
						}
						position++
					}
				l413:
					goto l407
				l408:
					position, tokenIndex, depth = position408, tokenIndex408, depth408
				}
				depth--
				add(rulereq_ws, position406)
			}
			return true
		l405:
			position, tokenIndex, depth = position405, tokenIndex405, depth405
			return false
		},
		/* 98 Action0 <- <{}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 99 Action1 <- <{}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 100 Action2 <- <{}> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
	}
	p.rules = _rules
}
