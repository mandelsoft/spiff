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
	ruleTagMarker
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
	ruleTopIndex
	ruleIndices
	ruleSlice
	ruleCurrying
	ruleChainedCall
	ruleStartArguments
	ruleNameArgumentList
	ruleNextNameArgument
	ruleExpressionList
	ruleNextExpression
	ruleListExpansion
	ruleProjection
	ruleProjectionValue
	ruleSubstitution
	ruleNot
	ruleGrouped
	ruleRange
	ruleStartRange
	ruleRangeOp
	ruleNumber
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
	ruleFilterList
	ruleFilterMap
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
	ruleTagPrefix
	ruleTag
	ruleTagComponent
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
	"TagMarker",
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
	"TopIndex",
	"Indices",
	"Slice",
	"Currying",
	"ChainedCall",
	"StartArguments",
	"NameArgumentList",
	"NextNameArgument",
	"ExpressionList",
	"NextExpression",
	"ListExpansion",
	"Projection",
	"ProjectionValue",
	"Substitution",
	"Not",
	"Grouped",
	"Range",
	"StartRange",
	"RangeOp",
	"Number",
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
	"FilterList",
	"FilterMap",
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
	"TagPrefix",
	"Tag",
	"TagComponent",
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
	rules  [111]func() bool
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
		/* 4 Marker <- <('&' (('t' 'e' 'm' 'p' 'l' 'a' 't' 'e') / ('t' 'e' 'm' 'p' 'o' 'r' 'a' 'r' 'y') / ('l' 'o' 'c' 'a' 'l') / ('i' 'n' 'j' 'e' 'c' 't') / ('s' 't' 'a' 't' 'e') / ('d' 'e' 'f' 'a' 'u' 'l' 't') / ('d' 'y' 'n' 'a' 'm' 'i' 'c') / TagMarker))> */
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
						goto l24
					}
					position++
					if buffer[position] != rune('e') {
						goto l24
					}
					position++
					if buffer[position] != rune('f') {
						goto l24
					}
					position++
					if buffer[position] != rune('a') {
						goto l24
					}
					position++
					if buffer[position] != rune('u') {
						goto l24
					}
					position++
					if buffer[position] != rune('l') {
						goto l24
					}
					position++
					if buffer[position] != rune('t') {
						goto l24
					}
					position++
					goto l18
				l24:
					position, tokenIndex, depth = position18, tokenIndex18, depth18
					if buffer[position] != rune('d') {
						goto l25
					}
					position++
					if buffer[position] != rune('y') {
						goto l25
					}
					position++
					if buffer[position] != rune('n') {
						goto l25
					}
					position++
					if buffer[position] != rune('a') {
						goto l25
					}
					position++
					if buffer[position] != rune('m') {
						goto l25
					}
					position++
					if buffer[position] != rune('i') {
						goto l25
					}
					position++
					if buffer[position] != rune('c') {
						goto l25
					}
					position++
					goto l18
				l25:
					position, tokenIndex, depth = position18, tokenIndex18, depth18
					if !_rules[ruleTagMarker]() {
						goto l16
					}
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
		/* 5 TagMarker <- <('t' 'a' 'g' ':' '*'? Tag)> */
		func() bool {
			position26, tokenIndex26, depth26 := position, tokenIndex, depth
			{
				position27 := position
				depth++
				if buffer[position] != rune('t') {
					goto l26
				}
				position++
				if buffer[position] != rune('a') {
					goto l26
				}
				position++
				if buffer[position] != rune('g') {
					goto l26
				}
				position++
				if buffer[position] != rune(':') {
					goto l26
				}
				position++
				{
					position28, tokenIndex28, depth28 := position, tokenIndex, depth
					if buffer[position] != rune('*') {
						goto l28
					}
					position++
					goto l29
				l28:
					position, tokenIndex, depth = position28, tokenIndex28, depth28
				}
			l29:
				if !_rules[ruleTag]() {
					goto l26
				}
				depth--
				add(ruleTagMarker, position27)
			}
			return true
		l26:
			position, tokenIndex, depth = position26, tokenIndex26, depth26
			return false
		},
		/* 6 MarkerExpression <- <Grouped> */
		func() bool {
			position30, tokenIndex30, depth30 := position, tokenIndex, depth
			{
				position31 := position
				depth++
				if !_rules[ruleGrouped]() {
					goto l30
				}
				depth--
				add(ruleMarkerExpression, position31)
			}
			return true
		l30:
			position, tokenIndex, depth = position30, tokenIndex30, depth30
			return false
		},
		/* 7 Expression <- <((Scoped / LambdaExpr / Level7) ws)> */
		func() bool {
			position32, tokenIndex32, depth32 := position, tokenIndex, depth
			{
				position33 := position
				depth++
				{
					position34, tokenIndex34, depth34 := position, tokenIndex, depth
					if !_rules[ruleScoped]() {
						goto l35
					}
					goto l34
				l35:
					position, tokenIndex, depth = position34, tokenIndex34, depth34
					if !_rules[ruleLambdaExpr]() {
						goto l36
					}
					goto l34
				l36:
					position, tokenIndex, depth = position34, tokenIndex34, depth34
					if !_rules[ruleLevel7]() {
						goto l32
					}
				}
			l34:
				if !_rules[rulews]() {
					goto l32
				}
				depth--
				add(ruleExpression, position33)
			}
			return true
		l32:
			position, tokenIndex, depth = position32, tokenIndex32, depth32
			return false
		},
		/* 8 Scoped <- <(ws Scope ws Expression)> */
		func() bool {
			position37, tokenIndex37, depth37 := position, tokenIndex, depth
			{
				position38 := position
				depth++
				if !_rules[rulews]() {
					goto l37
				}
				if !_rules[ruleScope]() {
					goto l37
				}
				if !_rules[rulews]() {
					goto l37
				}
				if !_rules[ruleExpression]() {
					goto l37
				}
				depth--
				add(ruleScoped, position38)
			}
			return true
		l37:
			position, tokenIndex, depth = position37, tokenIndex37, depth37
			return false
		},
		/* 9 Scope <- <(CreateScope ws Assignments? ')')> */
		func() bool {
			position39, tokenIndex39, depth39 := position, tokenIndex, depth
			{
				position40 := position
				depth++
				if !_rules[ruleCreateScope]() {
					goto l39
				}
				if !_rules[rulews]() {
					goto l39
				}
				{
					position41, tokenIndex41, depth41 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l41
					}
					goto l42
				l41:
					position, tokenIndex, depth = position41, tokenIndex41, depth41
				}
			l42:
				if buffer[position] != rune(')') {
					goto l39
				}
				position++
				depth--
				add(ruleScope, position40)
			}
			return true
		l39:
			position, tokenIndex, depth = position39, tokenIndex39, depth39
			return false
		},
		/* 10 CreateScope <- <'('> */
		func() bool {
			position43, tokenIndex43, depth43 := position, tokenIndex, depth
			{
				position44 := position
				depth++
				if buffer[position] != rune('(') {
					goto l43
				}
				position++
				depth--
				add(ruleCreateScope, position44)
			}
			return true
		l43:
			position, tokenIndex, depth = position43, tokenIndex43, depth43
			return false
		},
		/* 11 Level7 <- <(ws Level6 (req_ws Or)*)> */
		func() bool {
			position45, tokenIndex45, depth45 := position, tokenIndex, depth
			{
				position46 := position
				depth++
				if !_rules[rulews]() {
					goto l45
				}
				if !_rules[ruleLevel6]() {
					goto l45
				}
			l47:
				{
					position48, tokenIndex48, depth48 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l48
					}
					if !_rules[ruleOr]() {
						goto l48
					}
					goto l47
				l48:
					position, tokenIndex, depth = position48, tokenIndex48, depth48
				}
				depth--
				add(ruleLevel7, position46)
			}
			return true
		l45:
			position, tokenIndex, depth = position45, tokenIndex45, depth45
			return false
		},
		/* 12 Or <- <(OrOp req_ws Level6)> */
		func() bool {
			position49, tokenIndex49, depth49 := position, tokenIndex, depth
			{
				position50 := position
				depth++
				if !_rules[ruleOrOp]() {
					goto l49
				}
				if !_rules[rulereq_ws]() {
					goto l49
				}
				if !_rules[ruleLevel6]() {
					goto l49
				}
				depth--
				add(ruleOr, position50)
			}
			return true
		l49:
			position, tokenIndex, depth = position49, tokenIndex49, depth49
			return false
		},
		/* 13 OrOp <- <(('|' '|') / ('/' '/'))> */
		func() bool {
			position51, tokenIndex51, depth51 := position, tokenIndex, depth
			{
				position52 := position
				depth++
				{
					position53, tokenIndex53, depth53 := position, tokenIndex, depth
					if buffer[position] != rune('|') {
						goto l54
					}
					position++
					if buffer[position] != rune('|') {
						goto l54
					}
					position++
					goto l53
				l54:
					position, tokenIndex, depth = position53, tokenIndex53, depth53
					if buffer[position] != rune('/') {
						goto l51
					}
					position++
					if buffer[position] != rune('/') {
						goto l51
					}
					position++
				}
			l53:
				depth--
				add(ruleOrOp, position52)
			}
			return true
		l51:
			position, tokenIndex, depth = position51, tokenIndex51, depth51
			return false
		},
		/* 14 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position55, tokenIndex55, depth55 := position, tokenIndex, depth
			{
				position56 := position
				depth++
				{
					position57, tokenIndex57, depth57 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l58
					}
					goto l57
				l58:
					position, tokenIndex, depth = position57, tokenIndex57, depth57
					if !_rules[ruleLevel5]() {
						goto l55
					}
				}
			l57:
				depth--
				add(ruleLevel6, position56)
			}
			return true
		l55:
			position, tokenIndex, depth = position55, tokenIndex55, depth55
			return false
		},
		/* 15 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position59, tokenIndex59, depth59 := position, tokenIndex, depth
			{
				position60 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l59
				}
				if !_rules[rulews]() {
					goto l59
				}
				if buffer[position] != rune('?') {
					goto l59
				}
				position++
				if !_rules[ruleExpression]() {
					goto l59
				}
				if buffer[position] != rune(':') {
					goto l59
				}
				position++
				if !_rules[ruleExpression]() {
					goto l59
				}
				depth--
				add(ruleConditional, position60)
			}
			return true
		l59:
			position, tokenIndex, depth = position59, tokenIndex59, depth59
			return false
		},
		/* 16 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position61, tokenIndex61, depth61 := position, tokenIndex, depth
			{
				position62 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l61
				}
			l63:
				{
					position64, tokenIndex64, depth64 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l64
					}
					goto l63
				l64:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
				}
				depth--
				add(ruleLevel5, position62)
			}
			return true
		l61:
			position, tokenIndex, depth = position61, tokenIndex61, depth61
			return false
		},
		/* 17 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position65, tokenIndex65, depth65 := position, tokenIndex, depth
			{
				position66 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l65
				}
				if !_rules[ruleLevel4]() {
					goto l65
				}
				depth--
				add(ruleConcatenation, position66)
			}
			return true
		l65:
			position, tokenIndex, depth = position65, tokenIndex65, depth65
			return false
		},
		/* 18 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position67, tokenIndex67, depth67 := position, tokenIndex, depth
			{
				position68 := position
				depth++
				if !_rules[ruleLevel3]() {
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
						if !_rules[ruleLogOr]() {
							goto l72
						}
						goto l71
					l72:
						position, tokenIndex, depth = position71, tokenIndex71, depth71
						if !_rules[ruleLogAnd]() {
							goto l70
						}
					}
				l71:
					goto l69
				l70:
					position, tokenIndex, depth = position70, tokenIndex70, depth70
				}
				depth--
				add(ruleLevel4, position68)
			}
			return true
		l67:
			position, tokenIndex, depth = position67, tokenIndex67, depth67
			return false
		},
		/* 19 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position73, tokenIndex73, depth73 := position, tokenIndex, depth
			{
				position74 := position
				depth++
				if buffer[position] != rune('-') {
					goto l73
				}
				position++
				if buffer[position] != rune('o') {
					goto l73
				}
				position++
				if buffer[position] != rune('r') {
					goto l73
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l73
				}
				if !_rules[ruleLevel3]() {
					goto l73
				}
				depth--
				add(ruleLogOr, position74)
			}
			return true
		l73:
			position, tokenIndex, depth = position73, tokenIndex73, depth73
			return false
		},
		/* 20 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position75, tokenIndex75, depth75 := position, tokenIndex, depth
			{
				position76 := position
				depth++
				if buffer[position] != rune('-') {
					goto l75
				}
				position++
				if buffer[position] != rune('a') {
					goto l75
				}
				position++
				if buffer[position] != rune('n') {
					goto l75
				}
				position++
				if buffer[position] != rune('d') {
					goto l75
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l75
				}
				if !_rules[ruleLevel3]() {
					goto l75
				}
				depth--
				add(ruleLogAnd, position76)
			}
			return true
		l75:
			position, tokenIndex, depth = position75, tokenIndex75, depth75
			return false
		},
		/* 21 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position77, tokenIndex77, depth77 := position, tokenIndex, depth
			{
				position78 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l77
				}
			l79:
				{
					position80, tokenIndex80, depth80 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l80
					}
					if !_rules[ruleComparison]() {
						goto l80
					}
					goto l79
				l80:
					position, tokenIndex, depth = position80, tokenIndex80, depth80
				}
				depth--
				add(ruleLevel3, position78)
			}
			return true
		l77:
			position, tokenIndex, depth = position77, tokenIndex77, depth77
			return false
		},
		/* 22 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position81, tokenIndex81, depth81 := position, tokenIndex, depth
			{
				position82 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l81
				}
				if !_rules[rulereq_ws]() {
					goto l81
				}
				if !_rules[ruleLevel2]() {
					goto l81
				}
				depth--
				add(ruleComparison, position82)
			}
			return true
		l81:
			position, tokenIndex, depth = position81, tokenIndex81, depth81
			return false
		},
		/* 23 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position83, tokenIndex83, depth83 := position, tokenIndex, depth
			{
				position84 := position
				depth++
				{
					position85, tokenIndex85, depth85 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l86
					}
					position++
					if buffer[position] != rune('=') {
						goto l86
					}
					position++
					goto l85
				l86:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if buffer[position] != rune('!') {
						goto l87
					}
					position++
					if buffer[position] != rune('=') {
						goto l87
					}
					position++
					goto l85
				l87:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if buffer[position] != rune('<') {
						goto l88
					}
					position++
					if buffer[position] != rune('=') {
						goto l88
					}
					position++
					goto l85
				l88:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if buffer[position] != rune('>') {
						goto l89
					}
					position++
					if buffer[position] != rune('=') {
						goto l89
					}
					position++
					goto l85
				l89:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if buffer[position] != rune('>') {
						goto l90
					}
					position++
					goto l85
				l90:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if buffer[position] != rune('<') {
						goto l91
					}
					position++
					goto l85
				l91:
					position, tokenIndex, depth = position85, tokenIndex85, depth85
					if buffer[position] != rune('>') {
						goto l83
					}
					position++
				}
			l85:
				depth--
				add(ruleCompareOp, position84)
			}
			return true
		l83:
			position, tokenIndex, depth = position83, tokenIndex83, depth83
			return false
		},
		/* 24 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position92, tokenIndex92, depth92 := position, tokenIndex, depth
			{
				position93 := position
				depth++
				if !_rules[ruleLevel1]() {
					goto l92
				}
			l94:
				{
					position95, tokenIndex95, depth95 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l95
					}
					{
						position96, tokenIndex96, depth96 := position, tokenIndex, depth
						if !_rules[ruleAddition]() {
							goto l97
						}
						goto l96
					l97:
						position, tokenIndex, depth = position96, tokenIndex96, depth96
						if !_rules[ruleSubtraction]() {
							goto l95
						}
					}
				l96:
					goto l94
				l95:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
				}
				depth--
				add(ruleLevel2, position93)
			}
			return true
		l92:
			position, tokenIndex, depth = position92, tokenIndex92, depth92
			return false
		},
		/* 25 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position98, tokenIndex98, depth98 := position, tokenIndex, depth
			{
				position99 := position
				depth++
				if buffer[position] != rune('+') {
					goto l98
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l98
				}
				if !_rules[ruleLevel1]() {
					goto l98
				}
				depth--
				add(ruleAddition, position99)
			}
			return true
		l98:
			position, tokenIndex, depth = position98, tokenIndex98, depth98
			return false
		},
		/* 26 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position100, tokenIndex100, depth100 := position, tokenIndex, depth
			{
				position101 := position
				depth++
				if buffer[position] != rune('-') {
					goto l100
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l100
				}
				if !_rules[ruleLevel1]() {
					goto l100
				}
				depth--
				add(ruleSubtraction, position101)
			}
			return true
		l100:
			position, tokenIndex, depth = position100, tokenIndex100, depth100
			return false
		},
		/* 27 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position102, tokenIndex102, depth102 := position, tokenIndex, depth
			{
				position103 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l102
				}
			l104:
				{
					position105, tokenIndex105, depth105 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l105
					}
					{
						position106, tokenIndex106, depth106 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l107
						}
						goto l106
					l107:
						position, tokenIndex, depth = position106, tokenIndex106, depth106
						if !_rules[ruleDivision]() {
							goto l108
						}
						goto l106
					l108:
						position, tokenIndex, depth = position106, tokenIndex106, depth106
						if !_rules[ruleModulo]() {
							goto l105
						}
					}
				l106:
					goto l104
				l105:
					position, tokenIndex, depth = position105, tokenIndex105, depth105
				}
				depth--
				add(ruleLevel1, position103)
			}
			return true
		l102:
			position, tokenIndex, depth = position102, tokenIndex102, depth102
			return false
		},
		/* 28 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position109, tokenIndex109, depth109 := position, tokenIndex, depth
			{
				position110 := position
				depth++
				if buffer[position] != rune('*') {
					goto l109
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l109
				}
				if !_rules[ruleLevel0]() {
					goto l109
				}
				depth--
				add(ruleMultiplication, position110)
			}
			return true
		l109:
			position, tokenIndex, depth = position109, tokenIndex109, depth109
			return false
		},
		/* 29 Division <- <('/' req_ws Level0)> */
		func() bool {
			position111, tokenIndex111, depth111 := position, tokenIndex, depth
			{
				position112 := position
				depth++
				if buffer[position] != rune('/') {
					goto l111
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l111
				}
				if !_rules[ruleLevel0]() {
					goto l111
				}
				depth--
				add(ruleDivision, position112)
			}
			return true
		l111:
			position, tokenIndex, depth = position111, tokenIndex111, depth111
			return false
		},
		/* 30 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position113, tokenIndex113, depth113 := position, tokenIndex, depth
			{
				position114 := position
				depth++
				if buffer[position] != rune('%') {
					goto l113
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l113
				}
				if !_rules[ruleLevel0]() {
					goto l113
				}
				depth--
				add(ruleModulo, position114)
			}
			return true
		l113:
			position, tokenIndex, depth = position113, tokenIndex113, depth113
			return false
		},
		/* 31 Level0 <- <(IP / String / Number / Boolean / Undefined / Nil / Symbol / Not / Substitution / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position115, tokenIndex115, depth115 := position, tokenIndex, depth
			{
				position116 := position
				depth++
				{
					position117, tokenIndex117, depth117 := position, tokenIndex, depth
					if !_rules[ruleIP]() {
						goto l118
					}
					goto l117
				l118:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if !_rules[ruleString]() {
						goto l119
					}
					goto l117
				l119:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if !_rules[ruleNumber]() {
						goto l120
					}
					goto l117
				l120:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if !_rules[ruleBoolean]() {
						goto l121
					}
					goto l117
				l121:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if !_rules[ruleUndefined]() {
						goto l122
					}
					goto l117
				l122:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if !_rules[ruleNil]() {
						goto l123
					}
					goto l117
				l123:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if !_rules[ruleSymbol]() {
						goto l124
					}
					goto l117
				l124:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if !_rules[ruleNot]() {
						goto l125
					}
					goto l117
				l125:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if !_rules[ruleSubstitution]() {
						goto l126
					}
					goto l117
				l126:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if !_rules[ruleMerge]() {
						goto l127
					}
					goto l117
				l127:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if !_rules[ruleAuto]() {
						goto l128
					}
					goto l117
				l128:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if !_rules[ruleLambda]() {
						goto l129
					}
					goto l117
				l129:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if !_rules[ruleChained]() {
						goto l115
					}
				}
			l117:
				depth--
				add(ruleLevel0, position116)
			}
			return true
		l115:
			position, tokenIndex, depth = position115, tokenIndex115, depth115
			return false
		},
		/* 32 Chained <- <((MapMapping / Sync / Catch / Mapping / FilterList / FilterMap / MapSelection / Selection / Sum / List / Map / Range / Grouped / Reference / TopIndex) ChainedQualifiedExpression*)> */
		func() bool {
			position130, tokenIndex130, depth130 := position, tokenIndex, depth
			{
				position131 := position
				depth++
				{
					position132, tokenIndex132, depth132 := position, tokenIndex, depth
					if !_rules[ruleMapMapping]() {
						goto l133
					}
					goto l132
				l133:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if !_rules[ruleSync]() {
						goto l134
					}
					goto l132
				l134:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if !_rules[ruleCatch]() {
						goto l135
					}
					goto l132
				l135:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if !_rules[ruleMapping]() {
						goto l136
					}
					goto l132
				l136:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if !_rules[ruleFilterList]() {
						goto l137
					}
					goto l132
				l137:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if !_rules[ruleFilterMap]() {
						goto l138
					}
					goto l132
				l138:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if !_rules[ruleMapSelection]() {
						goto l139
					}
					goto l132
				l139:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if !_rules[ruleSelection]() {
						goto l140
					}
					goto l132
				l140:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if !_rules[ruleSum]() {
						goto l141
					}
					goto l132
				l141:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if !_rules[ruleList]() {
						goto l142
					}
					goto l132
				l142:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if !_rules[ruleMap]() {
						goto l143
					}
					goto l132
				l143:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if !_rules[ruleRange]() {
						goto l144
					}
					goto l132
				l144:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if !_rules[ruleGrouped]() {
						goto l145
					}
					goto l132
				l145:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if !_rules[ruleReference]() {
						goto l146
					}
					goto l132
				l146:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
					if !_rules[ruleTopIndex]() {
						goto l130
					}
				}
			l132:
			l147:
				{
					position148, tokenIndex148, depth148 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l148
					}
					goto l147
				l148:
					position, tokenIndex, depth = position148, tokenIndex148, depth148
				}
				depth--
				add(ruleChained, position131)
			}
			return true
		l130:
			position, tokenIndex, depth = position130, tokenIndex130, depth130
			return false
		},
		/* 33 ChainedQualifiedExpression <- <(ChainedCall / Currying / ChainedRef / ChainedDynRef / Projection)> */
		func() bool {
			position149, tokenIndex149, depth149 := position, tokenIndex, depth
			{
				position150 := position
				depth++
				{
					position151, tokenIndex151, depth151 := position, tokenIndex, depth
					if !_rules[ruleChainedCall]() {
						goto l152
					}
					goto l151
				l152:
					position, tokenIndex, depth = position151, tokenIndex151, depth151
					if !_rules[ruleCurrying]() {
						goto l153
					}
					goto l151
				l153:
					position, tokenIndex, depth = position151, tokenIndex151, depth151
					if !_rules[ruleChainedRef]() {
						goto l154
					}
					goto l151
				l154:
					position, tokenIndex, depth = position151, tokenIndex151, depth151
					if !_rules[ruleChainedDynRef]() {
						goto l155
					}
					goto l151
				l155:
					position, tokenIndex, depth = position151, tokenIndex151, depth151
					if !_rules[ruleProjection]() {
						goto l149
					}
				}
			l151:
				depth--
				add(ruleChainedQualifiedExpression, position150)
			}
			return true
		l149:
			position, tokenIndex, depth = position149, tokenIndex149, depth149
			return false
		},
		/* 34 ChainedRef <- <(PathComponent FollowUpRef)> */
		func() bool {
			position156, tokenIndex156, depth156 := position, tokenIndex, depth
			{
				position157 := position
				depth++
				if !_rules[rulePathComponent]() {
					goto l156
				}
				if !_rules[ruleFollowUpRef]() {
					goto l156
				}
				depth--
				add(ruleChainedRef, position157)
			}
			return true
		l156:
			position, tokenIndex, depth = position156, tokenIndex156, depth156
			return false
		},
		/* 35 ChainedDynRef <- <('.'? Indices)> */
		func() bool {
			position158, tokenIndex158, depth158 := position, tokenIndex, depth
			{
				position159 := position
				depth++
				{
					position160, tokenIndex160, depth160 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l160
					}
					position++
					goto l161
				l160:
					position, tokenIndex, depth = position160, tokenIndex160, depth160
				}
			l161:
				if !_rules[ruleIndices]() {
					goto l158
				}
				depth--
				add(ruleChainedDynRef, position159)
			}
			return true
		l158:
			position, tokenIndex, depth = position158, tokenIndex158, depth158
			return false
		},
		/* 36 TopIndex <- <('.' Indices)> */
		func() bool {
			position162, tokenIndex162, depth162 := position, tokenIndex, depth
			{
				position163 := position
				depth++
				if buffer[position] != rune('.') {
					goto l162
				}
				position++
				if !_rules[ruleIndices]() {
					goto l162
				}
				depth--
				add(ruleTopIndex, position163)
			}
			return true
		l162:
			position, tokenIndex, depth = position162, tokenIndex162, depth162
			return false
		},
		/* 37 Indices <- <(StartList ExpressionList ']')> */
		func() bool {
			position164, tokenIndex164, depth164 := position, tokenIndex, depth
			{
				position165 := position
				depth++
				if !_rules[ruleStartList]() {
					goto l164
				}
				if !_rules[ruleExpressionList]() {
					goto l164
				}
				if buffer[position] != rune(']') {
					goto l164
				}
				position++
				depth--
				add(ruleIndices, position165)
			}
			return true
		l164:
			position, tokenIndex, depth = position164, tokenIndex164, depth164
			return false
		},
		/* 38 Slice <- <Range> */
		func() bool {
			position166, tokenIndex166, depth166 := position, tokenIndex, depth
			{
				position167 := position
				depth++
				if !_rules[ruleRange]() {
					goto l166
				}
				depth--
				add(ruleSlice, position167)
			}
			return true
		l166:
			position, tokenIndex, depth = position166, tokenIndex166, depth166
			return false
		},
		/* 39 Currying <- <('*' ChainedCall)> */
		func() bool {
			position168, tokenIndex168, depth168 := position, tokenIndex, depth
			{
				position169 := position
				depth++
				if buffer[position] != rune('*') {
					goto l168
				}
				position++
				if !_rules[ruleChainedCall]() {
					goto l168
				}
				depth--
				add(ruleCurrying, position169)
			}
			return true
		l168:
			position, tokenIndex, depth = position168, tokenIndex168, depth168
			return false
		},
		/* 40 ChainedCall <- <(StartArguments NameArgumentList? ')')> */
		func() bool {
			position170, tokenIndex170, depth170 := position, tokenIndex, depth
			{
				position171 := position
				depth++
				if !_rules[ruleStartArguments]() {
					goto l170
				}
				{
					position172, tokenIndex172, depth172 := position, tokenIndex, depth
					if !_rules[ruleNameArgumentList]() {
						goto l172
					}
					goto l173
				l172:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
				}
			l173:
				if buffer[position] != rune(')') {
					goto l170
				}
				position++
				depth--
				add(ruleChainedCall, position171)
			}
			return true
		l170:
			position, tokenIndex, depth = position170, tokenIndex170, depth170
			return false
		},
		/* 41 StartArguments <- <('(' ws)> */
		func() bool {
			position174, tokenIndex174, depth174 := position, tokenIndex, depth
			{
				position175 := position
				depth++
				if buffer[position] != rune('(') {
					goto l174
				}
				position++
				if !_rules[rulews]() {
					goto l174
				}
				depth--
				add(ruleStartArguments, position175)
			}
			return true
		l174:
			position, tokenIndex, depth = position174, tokenIndex174, depth174
			return false
		},
		/* 42 NameArgumentList <- <(((NextNameArgument (',' NextNameArgument)*) / NextExpression) (',' NextExpression)*)> */
		func() bool {
			position176, tokenIndex176, depth176 := position, tokenIndex, depth
			{
				position177 := position
				depth++
				{
					position178, tokenIndex178, depth178 := position, tokenIndex, depth
					if !_rules[ruleNextNameArgument]() {
						goto l179
					}
				l180:
					{
						position181, tokenIndex181, depth181 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l181
						}
						position++
						if !_rules[ruleNextNameArgument]() {
							goto l181
						}
						goto l180
					l181:
						position, tokenIndex, depth = position181, tokenIndex181, depth181
					}
					goto l178
				l179:
					position, tokenIndex, depth = position178, tokenIndex178, depth178
					if !_rules[ruleNextExpression]() {
						goto l176
					}
				}
			l178:
			l182:
				{
					position183, tokenIndex183, depth183 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l183
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l183
					}
					goto l182
				l183:
					position, tokenIndex, depth = position183, tokenIndex183, depth183
				}
				depth--
				add(ruleNameArgumentList, position177)
			}
			return true
		l176:
			position, tokenIndex, depth = position176, tokenIndex176, depth176
			return false
		},
		/* 43 NextNameArgument <- <(ws Name ws '=' ws Expression ws)> */
		func() bool {
			position184, tokenIndex184, depth184 := position, tokenIndex, depth
			{
				position185 := position
				depth++
				if !_rules[rulews]() {
					goto l184
				}
				if !_rules[ruleName]() {
					goto l184
				}
				if !_rules[rulews]() {
					goto l184
				}
				if buffer[position] != rune('=') {
					goto l184
				}
				position++
				if !_rules[rulews]() {
					goto l184
				}
				if !_rules[ruleExpression]() {
					goto l184
				}
				if !_rules[rulews]() {
					goto l184
				}
				depth--
				add(ruleNextNameArgument, position185)
			}
			return true
		l184:
			position, tokenIndex, depth = position184, tokenIndex184, depth184
			return false
		},
		/* 44 ExpressionList <- <(NextExpression (',' NextExpression)*)> */
		func() bool {
			position186, tokenIndex186, depth186 := position, tokenIndex, depth
			{
				position187 := position
				depth++
				if !_rules[ruleNextExpression]() {
					goto l186
				}
			l188:
				{
					position189, tokenIndex189, depth189 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l189
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l189
					}
					goto l188
				l189:
					position, tokenIndex, depth = position189, tokenIndex189, depth189
				}
				depth--
				add(ruleExpressionList, position187)
			}
			return true
		l186:
			position, tokenIndex, depth = position186, tokenIndex186, depth186
			return false
		},
		/* 45 NextExpression <- <(Expression ListExpansion?)> */
		func() bool {
			position190, tokenIndex190, depth190 := position, tokenIndex, depth
			{
				position191 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l190
				}
				{
					position192, tokenIndex192, depth192 := position, tokenIndex, depth
					if !_rules[ruleListExpansion]() {
						goto l192
					}
					goto l193
				l192:
					position, tokenIndex, depth = position192, tokenIndex192, depth192
				}
			l193:
				depth--
				add(ruleNextExpression, position191)
			}
			return true
		l190:
			position, tokenIndex, depth = position190, tokenIndex190, depth190
			return false
		},
		/* 46 ListExpansion <- <('.' '.' '.' ws)> */
		func() bool {
			position194, tokenIndex194, depth194 := position, tokenIndex, depth
			{
				position195 := position
				depth++
				if buffer[position] != rune('.') {
					goto l194
				}
				position++
				if buffer[position] != rune('.') {
					goto l194
				}
				position++
				if buffer[position] != rune('.') {
					goto l194
				}
				position++
				if !_rules[rulews]() {
					goto l194
				}
				depth--
				add(ruleListExpansion, position195)
			}
			return true
		l194:
			position, tokenIndex, depth = position194, tokenIndex194, depth194
			return false
		},
		/* 47 Projection <- <('.'? (('[' '*' ']') / Slice) ProjectionValue ChainedQualifiedExpression*)> */
		func() bool {
			position196, tokenIndex196, depth196 := position, tokenIndex, depth
			{
				position197 := position
				depth++
				{
					position198, tokenIndex198, depth198 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l198
					}
					position++
					goto l199
				l198:
					position, tokenIndex, depth = position198, tokenIndex198, depth198
				}
			l199:
				{
					position200, tokenIndex200, depth200 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l201
					}
					position++
					if buffer[position] != rune('*') {
						goto l201
					}
					position++
					if buffer[position] != rune(']') {
						goto l201
					}
					position++
					goto l200
				l201:
					position, tokenIndex, depth = position200, tokenIndex200, depth200
					if !_rules[ruleSlice]() {
						goto l196
					}
				}
			l200:
				if !_rules[ruleProjectionValue]() {
					goto l196
				}
			l202:
				{
					position203, tokenIndex203, depth203 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l203
					}
					goto l202
				l203:
					position, tokenIndex, depth = position203, tokenIndex203, depth203
				}
				depth--
				add(ruleProjection, position197)
			}
			return true
		l196:
			position, tokenIndex, depth = position196, tokenIndex196, depth196
			return false
		},
		/* 48 ProjectionValue <- <Action0> */
		func() bool {
			position204, tokenIndex204, depth204 := position, tokenIndex, depth
			{
				position205 := position
				depth++
				if !_rules[ruleAction0]() {
					goto l204
				}
				depth--
				add(ruleProjectionValue, position205)
			}
			return true
		l204:
			position, tokenIndex, depth = position204, tokenIndex204, depth204
			return false
		},
		/* 49 Substitution <- <('*' Level0)> */
		func() bool {
			position206, tokenIndex206, depth206 := position, tokenIndex, depth
			{
				position207 := position
				depth++
				if buffer[position] != rune('*') {
					goto l206
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l206
				}
				depth--
				add(ruleSubstitution, position207)
			}
			return true
		l206:
			position, tokenIndex, depth = position206, tokenIndex206, depth206
			return false
		},
		/* 50 Not <- <('!' ws Level0)> */
		func() bool {
			position208, tokenIndex208, depth208 := position, tokenIndex, depth
			{
				position209 := position
				depth++
				if buffer[position] != rune('!') {
					goto l208
				}
				position++
				if !_rules[rulews]() {
					goto l208
				}
				if !_rules[ruleLevel0]() {
					goto l208
				}
				depth--
				add(ruleNot, position209)
			}
			return true
		l208:
			position, tokenIndex, depth = position208, tokenIndex208, depth208
			return false
		},
		/* 51 Grouped <- <('(' Expression ')')> */
		func() bool {
			position210, tokenIndex210, depth210 := position, tokenIndex, depth
			{
				position211 := position
				depth++
				if buffer[position] != rune('(') {
					goto l210
				}
				position++
				if !_rules[ruleExpression]() {
					goto l210
				}
				if buffer[position] != rune(')') {
					goto l210
				}
				position++
				depth--
				add(ruleGrouped, position211)
			}
			return true
		l210:
			position, tokenIndex, depth = position210, tokenIndex210, depth210
			return false
		},
		/* 52 Range <- <(StartRange Expression? RangeOp Expression? ']')> */
		func() bool {
			position212, tokenIndex212, depth212 := position, tokenIndex, depth
			{
				position213 := position
				depth++
				if !_rules[ruleStartRange]() {
					goto l212
				}
				{
					position214, tokenIndex214, depth214 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l214
					}
					goto l215
				l214:
					position, tokenIndex, depth = position214, tokenIndex214, depth214
				}
			l215:
				if !_rules[ruleRangeOp]() {
					goto l212
				}
				{
					position216, tokenIndex216, depth216 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l216
					}
					goto l217
				l216:
					position, tokenIndex, depth = position216, tokenIndex216, depth216
				}
			l217:
				if buffer[position] != rune(']') {
					goto l212
				}
				position++
				depth--
				add(ruleRange, position213)
			}
			return true
		l212:
			position, tokenIndex, depth = position212, tokenIndex212, depth212
			return false
		},
		/* 53 StartRange <- <'['> */
		func() bool {
			position218, tokenIndex218, depth218 := position, tokenIndex, depth
			{
				position219 := position
				depth++
				if buffer[position] != rune('[') {
					goto l218
				}
				position++
				depth--
				add(ruleStartRange, position219)
			}
			return true
		l218:
			position, tokenIndex, depth = position218, tokenIndex218, depth218
			return false
		},
		/* 54 RangeOp <- <('.' '.')> */
		func() bool {
			position220, tokenIndex220, depth220 := position, tokenIndex, depth
			{
				position221 := position
				depth++
				if buffer[position] != rune('.') {
					goto l220
				}
				position++
				if buffer[position] != rune('.') {
					goto l220
				}
				position++
				depth--
				add(ruleRangeOp, position221)
			}
			return true
		l220:
			position, tokenIndex, depth = position220, tokenIndex220, depth220
			return false
		},
		/* 55 Number <- <('-'? [0-9] ([0-9] / '_')* ('.' [0-9] [0-9]*)? (('e' / 'E') '-'? [0-9] [0-9]*)? !(':' ':'))> */
		func() bool {
			position222, tokenIndex222, depth222 := position, tokenIndex, depth
			{
				position223 := position
				depth++
				{
					position224, tokenIndex224, depth224 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l224
					}
					position++
					goto l225
				l224:
					position, tokenIndex, depth = position224, tokenIndex224, depth224
				}
			l225:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l222
				}
				position++
			l226:
				{
					position227, tokenIndex227, depth227 := position, tokenIndex, depth
					{
						position228, tokenIndex228, depth228 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l229
						}
						position++
						goto l228
					l229:
						position, tokenIndex, depth = position228, tokenIndex228, depth228
						if buffer[position] != rune('_') {
							goto l227
						}
						position++
					}
				l228:
					goto l226
				l227:
					position, tokenIndex, depth = position227, tokenIndex227, depth227
				}
				{
					position230, tokenIndex230, depth230 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l230
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l230
					}
					position++
				l232:
					{
						position233, tokenIndex233, depth233 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l233
						}
						position++
						goto l232
					l233:
						position, tokenIndex, depth = position233, tokenIndex233, depth233
					}
					goto l231
				l230:
					position, tokenIndex, depth = position230, tokenIndex230, depth230
				}
			l231:
				{
					position234, tokenIndex234, depth234 := position, tokenIndex, depth
					{
						position236, tokenIndex236, depth236 := position, tokenIndex, depth
						if buffer[position] != rune('e') {
							goto l237
						}
						position++
						goto l236
					l237:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						if buffer[position] != rune('E') {
							goto l234
						}
						position++
					}
				l236:
					{
						position238, tokenIndex238, depth238 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l238
						}
						position++
						goto l239
					l238:
						position, tokenIndex, depth = position238, tokenIndex238, depth238
					}
				l239:
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l234
					}
					position++
				l240:
					{
						position241, tokenIndex241, depth241 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l241
						}
						position++
						goto l240
					l241:
						position, tokenIndex, depth = position241, tokenIndex241, depth241
					}
					goto l235
				l234:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
				}
			l235:
				{
					position242, tokenIndex242, depth242 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l242
					}
					position++
					if buffer[position] != rune(':') {
						goto l242
					}
					position++
					goto l222
				l242:
					position, tokenIndex, depth = position242, tokenIndex242, depth242
				}
				depth--
				add(ruleNumber, position223)
			}
			return true
		l222:
			position, tokenIndex, depth = position222, tokenIndex222, depth222
			return false
		},
		/* 56 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position243, tokenIndex243, depth243 := position, tokenIndex, depth
			{
				position244 := position
				depth++
				if buffer[position] != rune('"') {
					goto l243
				}
				position++
			l245:
				{
					position246, tokenIndex246, depth246 := position, tokenIndex, depth
					{
						position247, tokenIndex247, depth247 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l248
						}
						position++
						if buffer[position] != rune('"') {
							goto l248
						}
						position++
						goto l247
					l248:
						position, tokenIndex, depth = position247, tokenIndex247, depth247
						{
							position249, tokenIndex249, depth249 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l249
							}
							position++
							goto l246
						l249:
							position, tokenIndex, depth = position249, tokenIndex249, depth249
						}
						if !matchDot() {
							goto l246
						}
					}
				l247:
					goto l245
				l246:
					position, tokenIndex, depth = position246, tokenIndex246, depth246
				}
				if buffer[position] != rune('"') {
					goto l243
				}
				position++
				depth--
				add(ruleString, position244)
			}
			return true
		l243:
			position, tokenIndex, depth = position243, tokenIndex243, depth243
			return false
		},
		/* 57 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position250, tokenIndex250, depth250 := position, tokenIndex, depth
			{
				position251 := position
				depth++
				{
					position252, tokenIndex252, depth252 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l253
					}
					position++
					if buffer[position] != rune('r') {
						goto l253
					}
					position++
					if buffer[position] != rune('u') {
						goto l253
					}
					position++
					if buffer[position] != rune('e') {
						goto l253
					}
					position++
					goto l252
				l253:
					position, tokenIndex, depth = position252, tokenIndex252, depth252
					if buffer[position] != rune('f') {
						goto l250
					}
					position++
					if buffer[position] != rune('a') {
						goto l250
					}
					position++
					if buffer[position] != rune('l') {
						goto l250
					}
					position++
					if buffer[position] != rune('s') {
						goto l250
					}
					position++
					if buffer[position] != rune('e') {
						goto l250
					}
					position++
				}
			l252:
				depth--
				add(ruleBoolean, position251)
			}
			return true
		l250:
			position, tokenIndex, depth = position250, tokenIndex250, depth250
			return false
		},
		/* 58 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position254, tokenIndex254, depth254 := position, tokenIndex, depth
			{
				position255 := position
				depth++
				{
					position256, tokenIndex256, depth256 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l257
					}
					position++
					if buffer[position] != rune('i') {
						goto l257
					}
					position++
					if buffer[position] != rune('l') {
						goto l257
					}
					position++
					goto l256
				l257:
					position, tokenIndex, depth = position256, tokenIndex256, depth256
					if buffer[position] != rune('~') {
						goto l254
					}
					position++
				}
			l256:
				depth--
				add(ruleNil, position255)
			}
			return true
		l254:
			position, tokenIndex, depth = position254, tokenIndex254, depth254
			return false
		},
		/* 59 Undefined <- <('~' '~')> */
		func() bool {
			position258, tokenIndex258, depth258 := position, tokenIndex, depth
			{
				position259 := position
				depth++
				if buffer[position] != rune('~') {
					goto l258
				}
				position++
				if buffer[position] != rune('~') {
					goto l258
				}
				position++
				depth--
				add(ruleUndefined, position259)
			}
			return true
		l258:
			position, tokenIndex, depth = position258, tokenIndex258, depth258
			return false
		},
		/* 60 Symbol <- <('$' Name)> */
		func() bool {
			position260, tokenIndex260, depth260 := position, tokenIndex, depth
			{
				position261 := position
				depth++
				if buffer[position] != rune('$') {
					goto l260
				}
				position++
				if !_rules[ruleName]() {
					goto l260
				}
				depth--
				add(ruleSymbol, position261)
			}
			return true
		l260:
			position, tokenIndex, depth = position260, tokenIndex260, depth260
			return false
		},
		/* 61 List <- <(StartList ExpressionList? ']')> */
		func() bool {
			position262, tokenIndex262, depth262 := position, tokenIndex, depth
			{
				position263 := position
				depth++
				if !_rules[ruleStartList]() {
					goto l262
				}
				{
					position264, tokenIndex264, depth264 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l264
					}
					goto l265
				l264:
					position, tokenIndex, depth = position264, tokenIndex264, depth264
				}
			l265:
				if buffer[position] != rune(']') {
					goto l262
				}
				position++
				depth--
				add(ruleList, position263)
			}
			return true
		l262:
			position, tokenIndex, depth = position262, tokenIndex262, depth262
			return false
		},
		/* 62 StartList <- <('[' ws)> */
		func() bool {
			position266, tokenIndex266, depth266 := position, tokenIndex, depth
			{
				position267 := position
				depth++
				if buffer[position] != rune('[') {
					goto l266
				}
				position++
				if !_rules[rulews]() {
					goto l266
				}
				depth--
				add(ruleStartList, position267)
			}
			return true
		l266:
			position, tokenIndex, depth = position266, tokenIndex266, depth266
			return false
		},
		/* 63 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position268, tokenIndex268, depth268 := position, tokenIndex, depth
			{
				position269 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l268
				}
				if !_rules[rulews]() {
					goto l268
				}
				{
					position270, tokenIndex270, depth270 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l270
					}
					goto l271
				l270:
					position, tokenIndex, depth = position270, tokenIndex270, depth270
				}
			l271:
				if buffer[position] != rune('}') {
					goto l268
				}
				position++
				depth--
				add(ruleMap, position269)
			}
			return true
		l268:
			position, tokenIndex, depth = position268, tokenIndex268, depth268
			return false
		},
		/* 64 CreateMap <- <'{'> */
		func() bool {
			position272, tokenIndex272, depth272 := position, tokenIndex, depth
			{
				position273 := position
				depth++
				if buffer[position] != rune('{') {
					goto l272
				}
				position++
				depth--
				add(ruleCreateMap, position273)
			}
			return true
		l272:
			position, tokenIndex, depth = position272, tokenIndex272, depth272
			return false
		},
		/* 65 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position274, tokenIndex274, depth274 := position, tokenIndex, depth
			{
				position275 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l274
				}
			l276:
				{
					position277, tokenIndex277, depth277 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l277
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l277
					}
					goto l276
				l277:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
				}
				depth--
				add(ruleAssignments, position275)
			}
			return true
		l274:
			position, tokenIndex, depth = position274, tokenIndex274, depth274
			return false
		},
		/* 66 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position278, tokenIndex278, depth278 := position, tokenIndex, depth
			{
				position279 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l278
				}
				if buffer[position] != rune('=') {
					goto l278
				}
				position++
				if !_rules[ruleExpression]() {
					goto l278
				}
				depth--
				add(ruleAssignment, position279)
			}
			return true
		l278:
			position, tokenIndex, depth = position278, tokenIndex278, depth278
			return false
		},
		/* 67 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position280, tokenIndex280, depth280 := position, tokenIndex, depth
			{
				position281 := position
				depth++
				{
					position282, tokenIndex282, depth282 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l283
					}
					goto l282
				l283:
					position, tokenIndex, depth = position282, tokenIndex282, depth282
					if !_rules[ruleSimpleMerge]() {
						goto l280
					}
				}
			l282:
				depth--
				add(ruleMerge, position281)
			}
			return true
		l280:
			position, tokenIndex, depth = position280, tokenIndex280, depth280
			return false
		},
		/* 68 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position284, tokenIndex284, depth284 := position, tokenIndex, depth
			{
				position285 := position
				depth++
				if buffer[position] != rune('m') {
					goto l284
				}
				position++
				if buffer[position] != rune('e') {
					goto l284
				}
				position++
				if buffer[position] != rune('r') {
					goto l284
				}
				position++
				if buffer[position] != rune('g') {
					goto l284
				}
				position++
				if buffer[position] != rune('e') {
					goto l284
				}
				position++
				{
					position286, tokenIndex286, depth286 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l286
					}
					if !_rules[ruleRequired]() {
						goto l286
					}
					goto l284
				l286:
					position, tokenIndex, depth = position286, tokenIndex286, depth286
				}
				{
					position287, tokenIndex287, depth287 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l287
					}
					{
						position289, tokenIndex289, depth289 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l290
						}
						goto l289
					l290:
						position, tokenIndex, depth = position289, tokenIndex289, depth289
						if !_rules[ruleOn]() {
							goto l287
						}
					}
				l289:
					goto l288
				l287:
					position, tokenIndex, depth = position287, tokenIndex287, depth287
				}
			l288:
				if !_rules[rulereq_ws]() {
					goto l284
				}
				if !_rules[ruleReference]() {
					goto l284
				}
				depth--
				add(ruleRefMerge, position285)
			}
			return true
		l284:
			position, tokenIndex, depth = position284, tokenIndex284, depth284
			return false
		},
		/* 69 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' !'(' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position291, tokenIndex291, depth291 := position, tokenIndex, depth
			{
				position292 := position
				depth++
				if buffer[position] != rune('m') {
					goto l291
				}
				position++
				if buffer[position] != rune('e') {
					goto l291
				}
				position++
				if buffer[position] != rune('r') {
					goto l291
				}
				position++
				if buffer[position] != rune('g') {
					goto l291
				}
				position++
				if buffer[position] != rune('e') {
					goto l291
				}
				position++
				{
					position293, tokenIndex293, depth293 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l293
					}
					position++
					goto l291
				l293:
					position, tokenIndex, depth = position293, tokenIndex293, depth293
				}
				{
					position294, tokenIndex294, depth294 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l294
					}
					{
						position296, tokenIndex296, depth296 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l297
						}
						goto l296
					l297:
						position, tokenIndex, depth = position296, tokenIndex296, depth296
						if !_rules[ruleRequired]() {
							goto l298
						}
						goto l296
					l298:
						position, tokenIndex, depth = position296, tokenIndex296, depth296
						if !_rules[ruleOn]() {
							goto l294
						}
					}
				l296:
					goto l295
				l294:
					position, tokenIndex, depth = position294, tokenIndex294, depth294
				}
			l295:
				depth--
				add(ruleSimpleMerge, position292)
			}
			return true
		l291:
			position, tokenIndex, depth = position291, tokenIndex291, depth291
			return false
		},
		/* 70 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position299, tokenIndex299, depth299 := position, tokenIndex, depth
			{
				position300 := position
				depth++
				if buffer[position] != rune('r') {
					goto l299
				}
				position++
				if buffer[position] != rune('e') {
					goto l299
				}
				position++
				if buffer[position] != rune('p') {
					goto l299
				}
				position++
				if buffer[position] != rune('l') {
					goto l299
				}
				position++
				if buffer[position] != rune('a') {
					goto l299
				}
				position++
				if buffer[position] != rune('c') {
					goto l299
				}
				position++
				if buffer[position] != rune('e') {
					goto l299
				}
				position++
				depth--
				add(ruleReplace, position300)
			}
			return true
		l299:
			position, tokenIndex, depth = position299, tokenIndex299, depth299
			return false
		},
		/* 71 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position301, tokenIndex301, depth301 := position, tokenIndex, depth
			{
				position302 := position
				depth++
				if buffer[position] != rune('r') {
					goto l301
				}
				position++
				if buffer[position] != rune('e') {
					goto l301
				}
				position++
				if buffer[position] != rune('q') {
					goto l301
				}
				position++
				if buffer[position] != rune('u') {
					goto l301
				}
				position++
				if buffer[position] != rune('i') {
					goto l301
				}
				position++
				if buffer[position] != rune('r') {
					goto l301
				}
				position++
				if buffer[position] != rune('e') {
					goto l301
				}
				position++
				if buffer[position] != rune('d') {
					goto l301
				}
				position++
				depth--
				add(ruleRequired, position302)
			}
			return true
		l301:
			position, tokenIndex, depth = position301, tokenIndex301, depth301
			return false
		},
		/* 72 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position303, tokenIndex303, depth303 := position, tokenIndex, depth
			{
				position304 := position
				depth++
				if buffer[position] != rune('o') {
					goto l303
				}
				position++
				if buffer[position] != rune('n') {
					goto l303
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l303
				}
				if !_rules[ruleName]() {
					goto l303
				}
				depth--
				add(ruleOn, position304)
			}
			return true
		l303:
			position, tokenIndex, depth = position303, tokenIndex303, depth303
			return false
		},
		/* 73 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position305, tokenIndex305, depth305 := position, tokenIndex, depth
			{
				position306 := position
				depth++
				if buffer[position] != rune('a') {
					goto l305
				}
				position++
				if buffer[position] != rune('u') {
					goto l305
				}
				position++
				if buffer[position] != rune('t') {
					goto l305
				}
				position++
				if buffer[position] != rune('o') {
					goto l305
				}
				position++
				depth--
				add(ruleAuto, position306)
			}
			return true
		l305:
			position, tokenIndex, depth = position305, tokenIndex305, depth305
			return false
		},
		/* 74 Default <- <Action1> */
		func() bool {
			position307, tokenIndex307, depth307 := position, tokenIndex, depth
			{
				position308 := position
				depth++
				if !_rules[ruleAction1]() {
					goto l307
				}
				depth--
				add(ruleDefault, position308)
			}
			return true
		l307:
			position, tokenIndex, depth = position307, tokenIndex307, depth307
			return false
		},
		/* 75 Sync <- <('s' 'y' 'n' 'c' '[' Level7 ((((LambdaExpr LambdaExt) / (LambdaOrExpr LambdaOrExpr)) (('|' Expression) / Default)) / (LambdaOrExpr Default Default)) ']')> */
		func() bool {
			position309, tokenIndex309, depth309 := position, tokenIndex, depth
			{
				position310 := position
				depth++
				if buffer[position] != rune('s') {
					goto l309
				}
				position++
				if buffer[position] != rune('y') {
					goto l309
				}
				position++
				if buffer[position] != rune('n') {
					goto l309
				}
				position++
				if buffer[position] != rune('c') {
					goto l309
				}
				position++
				if buffer[position] != rune('[') {
					goto l309
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l309
				}
				{
					position311, tokenIndex311, depth311 := position, tokenIndex, depth
					{
						position313, tokenIndex313, depth313 := position, tokenIndex, depth
						if !_rules[ruleLambdaExpr]() {
							goto l314
						}
						if !_rules[ruleLambdaExt]() {
							goto l314
						}
						goto l313
					l314:
						position, tokenIndex, depth = position313, tokenIndex313, depth313
						if !_rules[ruleLambdaOrExpr]() {
							goto l312
						}
						if !_rules[ruleLambdaOrExpr]() {
							goto l312
						}
					}
				l313:
					{
						position315, tokenIndex315, depth315 := position, tokenIndex, depth
						if buffer[position] != rune('|') {
							goto l316
						}
						position++
						if !_rules[ruleExpression]() {
							goto l316
						}
						goto l315
					l316:
						position, tokenIndex, depth = position315, tokenIndex315, depth315
						if !_rules[ruleDefault]() {
							goto l312
						}
					}
				l315:
					goto l311
				l312:
					position, tokenIndex, depth = position311, tokenIndex311, depth311
					if !_rules[ruleLambdaOrExpr]() {
						goto l309
					}
					if !_rules[ruleDefault]() {
						goto l309
					}
					if !_rules[ruleDefault]() {
						goto l309
					}
				}
			l311:
				if buffer[position] != rune(']') {
					goto l309
				}
				position++
				depth--
				add(ruleSync, position310)
			}
			return true
		l309:
			position, tokenIndex, depth = position309, tokenIndex309, depth309
			return false
		},
		/* 76 LambdaExt <- <(',' Expression)> */
		func() bool {
			position317, tokenIndex317, depth317 := position, tokenIndex, depth
			{
				position318 := position
				depth++
				if buffer[position] != rune(',') {
					goto l317
				}
				position++
				if !_rules[ruleExpression]() {
					goto l317
				}
				depth--
				add(ruleLambdaExt, position318)
			}
			return true
		l317:
			position, tokenIndex, depth = position317, tokenIndex317, depth317
			return false
		},
		/* 77 LambdaOrExpr <- <(LambdaExpr / ('|' Expression))> */
		func() bool {
			position319, tokenIndex319, depth319 := position, tokenIndex, depth
			{
				position320 := position
				depth++
				{
					position321, tokenIndex321, depth321 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l322
					}
					goto l321
				l322:
					position, tokenIndex, depth = position321, tokenIndex321, depth321
					if buffer[position] != rune('|') {
						goto l319
					}
					position++
					if !_rules[ruleExpression]() {
						goto l319
					}
				}
			l321:
				depth--
				add(ruleLambdaOrExpr, position320)
			}
			return true
		l319:
			position, tokenIndex, depth = position319, tokenIndex319, depth319
			return false
		},
		/* 78 Catch <- <('c' 'a' 't' 'c' 'h' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position323, tokenIndex323, depth323 := position, tokenIndex, depth
			{
				position324 := position
				depth++
				if buffer[position] != rune('c') {
					goto l323
				}
				position++
				if buffer[position] != rune('a') {
					goto l323
				}
				position++
				if buffer[position] != rune('t') {
					goto l323
				}
				position++
				if buffer[position] != rune('c') {
					goto l323
				}
				position++
				if buffer[position] != rune('h') {
					goto l323
				}
				position++
				if buffer[position] != rune('[') {
					goto l323
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l323
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l323
				}
				if buffer[position] != rune(']') {
					goto l323
				}
				position++
				depth--
				add(ruleCatch, position324)
			}
			return true
		l323:
			position, tokenIndex, depth = position323, tokenIndex323, depth323
			return false
		},
		/* 79 FilterList <- <('f' 'i' 'l' 't' 'e' 'r' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position325, tokenIndex325, depth325 := position, tokenIndex, depth
			{
				position326 := position
				depth++
				if buffer[position] != rune('f') {
					goto l325
				}
				position++
				if buffer[position] != rune('i') {
					goto l325
				}
				position++
				if buffer[position] != rune('l') {
					goto l325
				}
				position++
				if buffer[position] != rune('t') {
					goto l325
				}
				position++
				if buffer[position] != rune('e') {
					goto l325
				}
				position++
				if buffer[position] != rune('r') {
					goto l325
				}
				position++
				if buffer[position] != rune('[') {
					goto l325
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l325
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l325
				}
				if buffer[position] != rune(']') {
					goto l325
				}
				position++
				depth--
				add(ruleFilterList, position326)
			}
			return true
		l325:
			position, tokenIndex, depth = position325, tokenIndex325, depth325
			return false
		},
		/* 80 FilterMap <- <('f' 'i' 'l' 't' 'e' 'r' '{' Level7 LambdaOrExpr '}')> */
		func() bool {
			position327, tokenIndex327, depth327 := position, tokenIndex, depth
			{
				position328 := position
				depth++
				if buffer[position] != rune('f') {
					goto l327
				}
				position++
				if buffer[position] != rune('i') {
					goto l327
				}
				position++
				if buffer[position] != rune('l') {
					goto l327
				}
				position++
				if buffer[position] != rune('t') {
					goto l327
				}
				position++
				if buffer[position] != rune('e') {
					goto l327
				}
				position++
				if buffer[position] != rune('r') {
					goto l327
				}
				position++
				if buffer[position] != rune('{') {
					goto l327
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l327
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l327
				}
				if buffer[position] != rune('}') {
					goto l327
				}
				position++
				depth--
				add(ruleFilterMap, position328)
			}
			return true
		l327:
			position, tokenIndex, depth = position327, tokenIndex327, depth327
			return false
		},
		/* 81 MapMapping <- <('m' 'a' 'p' '{' Level7 LambdaOrExpr '}')> */
		func() bool {
			position329, tokenIndex329, depth329 := position, tokenIndex, depth
			{
				position330 := position
				depth++
				if buffer[position] != rune('m') {
					goto l329
				}
				position++
				if buffer[position] != rune('a') {
					goto l329
				}
				position++
				if buffer[position] != rune('p') {
					goto l329
				}
				position++
				if buffer[position] != rune('{') {
					goto l329
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l329
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l329
				}
				if buffer[position] != rune('}') {
					goto l329
				}
				position++
				depth--
				add(ruleMapMapping, position330)
			}
			return true
		l329:
			position, tokenIndex, depth = position329, tokenIndex329, depth329
			return false
		},
		/* 82 Mapping <- <('m' 'a' 'p' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position331, tokenIndex331, depth331 := position, tokenIndex, depth
			{
				position332 := position
				depth++
				if buffer[position] != rune('m') {
					goto l331
				}
				position++
				if buffer[position] != rune('a') {
					goto l331
				}
				position++
				if buffer[position] != rune('p') {
					goto l331
				}
				position++
				if buffer[position] != rune('[') {
					goto l331
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l331
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l331
				}
				if buffer[position] != rune(']') {
					goto l331
				}
				position++
				depth--
				add(ruleMapping, position332)
			}
			return true
		l331:
			position, tokenIndex, depth = position331, tokenIndex331, depth331
			return false
		},
		/* 83 MapSelection <- <('s' 'e' 'l' 'e' 'c' 't' '{' Level7 LambdaOrExpr '}')> */
		func() bool {
			position333, tokenIndex333, depth333 := position, tokenIndex, depth
			{
				position334 := position
				depth++
				if buffer[position] != rune('s') {
					goto l333
				}
				position++
				if buffer[position] != rune('e') {
					goto l333
				}
				position++
				if buffer[position] != rune('l') {
					goto l333
				}
				position++
				if buffer[position] != rune('e') {
					goto l333
				}
				position++
				if buffer[position] != rune('c') {
					goto l333
				}
				position++
				if buffer[position] != rune('t') {
					goto l333
				}
				position++
				if buffer[position] != rune('{') {
					goto l333
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l333
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l333
				}
				if buffer[position] != rune('}') {
					goto l333
				}
				position++
				depth--
				add(ruleMapSelection, position334)
			}
			return true
		l333:
			position, tokenIndex, depth = position333, tokenIndex333, depth333
			return false
		},
		/* 84 Selection <- <('s' 'e' 'l' 'e' 'c' 't' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position335, tokenIndex335, depth335 := position, tokenIndex, depth
			{
				position336 := position
				depth++
				if buffer[position] != rune('s') {
					goto l335
				}
				position++
				if buffer[position] != rune('e') {
					goto l335
				}
				position++
				if buffer[position] != rune('l') {
					goto l335
				}
				position++
				if buffer[position] != rune('e') {
					goto l335
				}
				position++
				if buffer[position] != rune('c') {
					goto l335
				}
				position++
				if buffer[position] != rune('t') {
					goto l335
				}
				position++
				if buffer[position] != rune('[') {
					goto l335
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l335
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l335
				}
				if buffer[position] != rune(']') {
					goto l335
				}
				position++
				depth--
				add(ruleSelection, position336)
			}
			return true
		l335:
			position, tokenIndex, depth = position335, tokenIndex335, depth335
			return false
		},
		/* 85 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 LambdaOrExpr ']')> */
		func() bool {
			position337, tokenIndex337, depth337 := position, tokenIndex, depth
			{
				position338 := position
				depth++
				if buffer[position] != rune('s') {
					goto l337
				}
				position++
				if buffer[position] != rune('u') {
					goto l337
				}
				position++
				if buffer[position] != rune('m') {
					goto l337
				}
				position++
				if buffer[position] != rune('[') {
					goto l337
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l337
				}
				if buffer[position] != rune('|') {
					goto l337
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l337
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l337
				}
				if buffer[position] != rune(']') {
					goto l337
				}
				position++
				depth--
				add(ruleSum, position338)
			}
			return true
		l337:
			position, tokenIndex, depth = position337, tokenIndex337, depth337
			return false
		},
		/* 86 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position339, tokenIndex339, depth339 := position, tokenIndex, depth
			{
				position340 := position
				depth++
				if buffer[position] != rune('l') {
					goto l339
				}
				position++
				if buffer[position] != rune('a') {
					goto l339
				}
				position++
				if buffer[position] != rune('m') {
					goto l339
				}
				position++
				if buffer[position] != rune('b') {
					goto l339
				}
				position++
				if buffer[position] != rune('d') {
					goto l339
				}
				position++
				if buffer[position] != rune('a') {
					goto l339
				}
				position++
				{
					position341, tokenIndex341, depth341 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l342
					}
					goto l341
				l342:
					position, tokenIndex, depth = position341, tokenIndex341, depth341
					if !_rules[ruleLambdaExpr]() {
						goto l339
					}
				}
			l341:
				depth--
				add(ruleLambda, position340)
			}
			return true
		l339:
			position, tokenIndex, depth = position339, tokenIndex339, depth339
			return false
		},
		/* 87 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position343, tokenIndex343, depth343 := position, tokenIndex, depth
			{
				position344 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l343
				}
				if !_rules[ruleExpression]() {
					goto l343
				}
				depth--
				add(ruleLambdaRef, position344)
			}
			return true
		l343:
			position, tokenIndex, depth = position343, tokenIndex343, depth343
			return false
		},
		/* 88 LambdaExpr <- <(ws Params ws ('-' '>') Expression)> */
		func() bool {
			position345, tokenIndex345, depth345 := position, tokenIndex, depth
			{
				position346 := position
				depth++
				if !_rules[rulews]() {
					goto l345
				}
				if !_rules[ruleParams]() {
					goto l345
				}
				if !_rules[rulews]() {
					goto l345
				}
				if buffer[position] != rune('-') {
					goto l345
				}
				position++
				if buffer[position] != rune('>') {
					goto l345
				}
				position++
				if !_rules[ruleExpression]() {
					goto l345
				}
				depth--
				add(ruleLambdaExpr, position346)
			}
			return true
		l345:
			position, tokenIndex, depth = position345, tokenIndex345, depth345
			return false
		},
		/* 89 Params <- <('|' StartParams ws Names? '|')> */
		func() bool {
			position347, tokenIndex347, depth347 := position, tokenIndex, depth
			{
				position348 := position
				depth++
				if buffer[position] != rune('|') {
					goto l347
				}
				position++
				if !_rules[ruleStartParams]() {
					goto l347
				}
				if !_rules[rulews]() {
					goto l347
				}
				{
					position349, tokenIndex349, depth349 := position, tokenIndex, depth
					if !_rules[ruleNames]() {
						goto l349
					}
					goto l350
				l349:
					position, tokenIndex, depth = position349, tokenIndex349, depth349
				}
			l350:
				if buffer[position] != rune('|') {
					goto l347
				}
				position++
				depth--
				add(ruleParams, position348)
			}
			return true
		l347:
			position, tokenIndex, depth = position347, tokenIndex347, depth347
			return false
		},
		/* 90 StartParams <- <Action2> */
		func() bool {
			position351, tokenIndex351, depth351 := position, tokenIndex, depth
			{
				position352 := position
				depth++
				if !_rules[ruleAction2]() {
					goto l351
				}
				depth--
				add(ruleStartParams, position352)
			}
			return true
		l351:
			position, tokenIndex, depth = position351, tokenIndex351, depth351
			return false
		},
		/* 91 Names <- <(NextName (',' NextName)* DefaultValue? (',' NextName DefaultValue)* VarParams?)> */
		func() bool {
			position353, tokenIndex353, depth353 := position, tokenIndex, depth
			{
				position354 := position
				depth++
				if !_rules[ruleNextName]() {
					goto l353
				}
			l355:
				{
					position356, tokenIndex356, depth356 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l356
					}
					position++
					if !_rules[ruleNextName]() {
						goto l356
					}
					goto l355
				l356:
					position, tokenIndex, depth = position356, tokenIndex356, depth356
				}
				{
					position357, tokenIndex357, depth357 := position, tokenIndex, depth
					if !_rules[ruleDefaultValue]() {
						goto l357
					}
					goto l358
				l357:
					position, tokenIndex, depth = position357, tokenIndex357, depth357
				}
			l358:
			l359:
				{
					position360, tokenIndex360, depth360 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l360
					}
					position++
					if !_rules[ruleNextName]() {
						goto l360
					}
					if !_rules[ruleDefaultValue]() {
						goto l360
					}
					goto l359
				l360:
					position, tokenIndex, depth = position360, tokenIndex360, depth360
				}
				{
					position361, tokenIndex361, depth361 := position, tokenIndex, depth
					if !_rules[ruleVarParams]() {
						goto l361
					}
					goto l362
				l361:
					position, tokenIndex, depth = position361, tokenIndex361, depth361
				}
			l362:
				depth--
				add(ruleNames, position354)
			}
			return true
		l353:
			position, tokenIndex, depth = position353, tokenIndex353, depth353
			return false
		},
		/* 92 NextName <- <(ws Name ws)> */
		func() bool {
			position363, tokenIndex363, depth363 := position, tokenIndex, depth
			{
				position364 := position
				depth++
				if !_rules[rulews]() {
					goto l363
				}
				if !_rules[ruleName]() {
					goto l363
				}
				if !_rules[rulews]() {
					goto l363
				}
				depth--
				add(ruleNextName, position364)
			}
			return true
		l363:
			position, tokenIndex, depth = position363, tokenIndex363, depth363
			return false
		},
		/* 93 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position365, tokenIndex365, depth365 := position, tokenIndex, depth
			{
				position366 := position
				depth++
				{
					position369, tokenIndex369, depth369 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l370
					}
					position++
					goto l369
				l370:
					position, tokenIndex, depth = position369, tokenIndex369, depth369
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l371
					}
					position++
					goto l369
				l371:
					position, tokenIndex, depth = position369, tokenIndex369, depth369
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l372
					}
					position++
					goto l369
				l372:
					position, tokenIndex, depth = position369, tokenIndex369, depth369
					if buffer[position] != rune('_') {
						goto l365
					}
					position++
				}
			l369:
			l367:
				{
					position368, tokenIndex368, depth368 := position, tokenIndex, depth
					{
						position373, tokenIndex373, depth373 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l374
						}
						position++
						goto l373
					l374:
						position, tokenIndex, depth = position373, tokenIndex373, depth373
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l375
						}
						position++
						goto l373
					l375:
						position, tokenIndex, depth = position373, tokenIndex373, depth373
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l376
						}
						position++
						goto l373
					l376:
						position, tokenIndex, depth = position373, tokenIndex373, depth373
						if buffer[position] != rune('_') {
							goto l368
						}
						position++
					}
				l373:
					goto l367
				l368:
					position, tokenIndex, depth = position368, tokenIndex368, depth368
				}
				depth--
				add(ruleName, position366)
			}
			return true
		l365:
			position, tokenIndex, depth = position365, tokenIndex365, depth365
			return false
		},
		/* 94 DefaultValue <- <('=' Expression)> */
		func() bool {
			position377, tokenIndex377, depth377 := position, tokenIndex, depth
			{
				position378 := position
				depth++
				if buffer[position] != rune('=') {
					goto l377
				}
				position++
				if !_rules[ruleExpression]() {
					goto l377
				}
				depth--
				add(ruleDefaultValue, position378)
			}
			return true
		l377:
			position, tokenIndex, depth = position377, tokenIndex377, depth377
			return false
		},
		/* 95 VarParams <- <('.' '.' '.' ws)> */
		func() bool {
			position379, tokenIndex379, depth379 := position, tokenIndex, depth
			{
				position380 := position
				depth++
				if buffer[position] != rune('.') {
					goto l379
				}
				position++
				if buffer[position] != rune('.') {
					goto l379
				}
				position++
				if buffer[position] != rune('.') {
					goto l379
				}
				position++
				if !_rules[rulews]() {
					goto l379
				}
				depth--
				add(ruleVarParams, position380)
			}
			return true
		l379:
			position, tokenIndex, depth = position379, tokenIndex379, depth379
			return false
		},
		/* 96 Reference <- <(((TagPrefix ('.' / Key)) / ('.'? Key)) FollowUpRef)> */
		func() bool {
			position381, tokenIndex381, depth381 := position, tokenIndex, depth
			{
				position382 := position
				depth++
				{
					position383, tokenIndex383, depth383 := position, tokenIndex, depth
					if !_rules[ruleTagPrefix]() {
						goto l384
					}
					{
						position385, tokenIndex385, depth385 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l386
						}
						position++
						goto l385
					l386:
						position, tokenIndex, depth = position385, tokenIndex385, depth385
						if !_rules[ruleKey]() {
							goto l384
						}
					}
				l385:
					goto l383
				l384:
					position, tokenIndex, depth = position383, tokenIndex383, depth383
					{
						position387, tokenIndex387, depth387 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l387
						}
						position++
						goto l388
					l387:
						position, tokenIndex, depth = position387, tokenIndex387, depth387
					}
				l388:
					if !_rules[ruleKey]() {
						goto l381
					}
				}
			l383:
				if !_rules[ruleFollowUpRef]() {
					goto l381
				}
				depth--
				add(ruleReference, position382)
			}
			return true
		l381:
			position, tokenIndex, depth = position381, tokenIndex381, depth381
			return false
		},
		/* 97 TagPrefix <- <((('d' 'o' 'c' ('.' / ':') '-'? [0-9]+) / Tag) (':' ':'))> */
		func() bool {
			position389, tokenIndex389, depth389 := position, tokenIndex, depth
			{
				position390 := position
				depth++
				{
					position391, tokenIndex391, depth391 := position, tokenIndex, depth
					if buffer[position] != rune('d') {
						goto l392
					}
					position++
					if buffer[position] != rune('o') {
						goto l392
					}
					position++
					if buffer[position] != rune('c') {
						goto l392
					}
					position++
					{
						position393, tokenIndex393, depth393 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l394
						}
						position++
						goto l393
					l394:
						position, tokenIndex, depth = position393, tokenIndex393, depth393
						if buffer[position] != rune(':') {
							goto l392
						}
						position++
					}
				l393:
					{
						position395, tokenIndex395, depth395 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l395
						}
						position++
						goto l396
					l395:
						position, tokenIndex, depth = position395, tokenIndex395, depth395
					}
				l396:
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l392
					}
					position++
				l397:
					{
						position398, tokenIndex398, depth398 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l398
						}
						position++
						goto l397
					l398:
						position, tokenIndex, depth = position398, tokenIndex398, depth398
					}
					goto l391
				l392:
					position, tokenIndex, depth = position391, tokenIndex391, depth391
					if !_rules[ruleTag]() {
						goto l389
					}
				}
			l391:
				if buffer[position] != rune(':') {
					goto l389
				}
				position++
				if buffer[position] != rune(':') {
					goto l389
				}
				position++
				depth--
				add(ruleTagPrefix, position390)
			}
			return true
		l389:
			position, tokenIndex, depth = position389, tokenIndex389, depth389
			return false
		},
		/* 98 Tag <- <(TagComponent (('.' / ':') TagComponent)*)> */
		func() bool {
			position399, tokenIndex399, depth399 := position, tokenIndex, depth
			{
				position400 := position
				depth++
				if !_rules[ruleTagComponent]() {
					goto l399
				}
			l401:
				{
					position402, tokenIndex402, depth402 := position, tokenIndex, depth
					{
						position403, tokenIndex403, depth403 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l404
						}
						position++
						goto l403
					l404:
						position, tokenIndex, depth = position403, tokenIndex403, depth403
						if buffer[position] != rune(':') {
							goto l402
						}
						position++
					}
				l403:
					if !_rules[ruleTagComponent]() {
						goto l402
					}
					goto l401
				l402:
					position, tokenIndex, depth = position402, tokenIndex402, depth402
				}
				depth--
				add(ruleTag, position400)
			}
			return true
		l399:
			position, tokenIndex, depth = position399, tokenIndex399, depth399
			return false
		},
		/* 99 TagComponent <- <(([a-z] / [A-Z] / '_') ([a-z] / [A-Z] / [0-9] / '_')*)> */
		func() bool {
			position405, tokenIndex405, depth405 := position, tokenIndex, depth
			{
				position406 := position
				depth++
				{
					position407, tokenIndex407, depth407 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l408
					}
					position++
					goto l407
				l408:
					position, tokenIndex, depth = position407, tokenIndex407, depth407
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l409
					}
					position++
					goto l407
				l409:
					position, tokenIndex, depth = position407, tokenIndex407, depth407
					if buffer[position] != rune('_') {
						goto l405
					}
					position++
				}
			l407:
			l410:
				{
					position411, tokenIndex411, depth411 := position, tokenIndex, depth
					{
						position412, tokenIndex412, depth412 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l413
						}
						position++
						goto l412
					l413:
						position, tokenIndex, depth = position412, tokenIndex412, depth412
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l414
						}
						position++
						goto l412
					l414:
						position, tokenIndex, depth = position412, tokenIndex412, depth412
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l415
						}
						position++
						goto l412
					l415:
						position, tokenIndex, depth = position412, tokenIndex412, depth412
						if buffer[position] != rune('_') {
							goto l411
						}
						position++
					}
				l412:
					goto l410
				l411:
					position, tokenIndex, depth = position411, tokenIndex411, depth411
				}
				depth--
				add(ruleTagComponent, position406)
			}
			return true
		l405:
			position, tokenIndex, depth = position405, tokenIndex405, depth405
			return false
		},
		/* 100 FollowUpRef <- <PathComponent*> */
		func() bool {
			{
				position417 := position
				depth++
			l418:
				{
					position419, tokenIndex419, depth419 := position, tokenIndex, depth
					if !_rules[rulePathComponent]() {
						goto l419
					}
					goto l418
				l419:
					position, tokenIndex, depth = position419, tokenIndex419, depth419
				}
				depth--
				add(ruleFollowUpRef, position417)
			}
			return true
		},
		/* 101 PathComponent <- <(('.' Key) / ('.'? Index))> */
		func() bool {
			position420, tokenIndex420, depth420 := position, tokenIndex, depth
			{
				position421 := position
				depth++
				{
					position422, tokenIndex422, depth422 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l423
					}
					position++
					if !_rules[ruleKey]() {
						goto l423
					}
					goto l422
				l423:
					position, tokenIndex, depth = position422, tokenIndex422, depth422
					{
						position424, tokenIndex424, depth424 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l424
						}
						position++
						goto l425
					l424:
						position, tokenIndex, depth = position424, tokenIndex424, depth424
					}
				l425:
					if !_rules[ruleIndex]() {
						goto l420
					}
				}
			l422:
				depth--
				add(rulePathComponent, position421)
			}
			return true
		l420:
			position, tokenIndex, depth = position420, tokenIndex420, depth420
			return false
		},
		/* 102 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position426, tokenIndex426, depth426 := position, tokenIndex, depth
			{
				position427 := position
				depth++
				{
					position428, tokenIndex428, depth428 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l429
					}
					position++
					goto l428
				l429:
					position, tokenIndex, depth = position428, tokenIndex428, depth428
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l430
					}
					position++
					goto l428
				l430:
					position, tokenIndex, depth = position428, tokenIndex428, depth428
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l431
					}
					position++
					goto l428
				l431:
					position, tokenIndex, depth = position428, tokenIndex428, depth428
					if buffer[position] != rune('_') {
						goto l426
					}
					position++
				}
			l428:
			l432:
				{
					position433, tokenIndex433, depth433 := position, tokenIndex, depth
					{
						position434, tokenIndex434, depth434 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l435
						}
						position++
						goto l434
					l435:
						position, tokenIndex, depth = position434, tokenIndex434, depth434
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l436
						}
						position++
						goto l434
					l436:
						position, tokenIndex, depth = position434, tokenIndex434, depth434
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l437
						}
						position++
						goto l434
					l437:
						position, tokenIndex, depth = position434, tokenIndex434, depth434
						if buffer[position] != rune('_') {
							goto l438
						}
						position++
						goto l434
					l438:
						position, tokenIndex, depth = position434, tokenIndex434, depth434
						if buffer[position] != rune('-') {
							goto l433
						}
						position++
					}
				l434:
					goto l432
				l433:
					position, tokenIndex, depth = position433, tokenIndex433, depth433
				}
				{
					position439, tokenIndex439, depth439 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l439
					}
					position++
					{
						position441, tokenIndex441, depth441 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l442
						}
						position++
						goto l441
					l442:
						position, tokenIndex, depth = position441, tokenIndex441, depth441
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l443
						}
						position++
						goto l441
					l443:
						position, tokenIndex, depth = position441, tokenIndex441, depth441
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l444
						}
						position++
						goto l441
					l444:
						position, tokenIndex, depth = position441, tokenIndex441, depth441
						if buffer[position] != rune('_') {
							goto l439
						}
						position++
					}
				l441:
				l445:
					{
						position446, tokenIndex446, depth446 := position, tokenIndex, depth
						{
							position447, tokenIndex447, depth447 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l448
							}
							position++
							goto l447
						l448:
							position, tokenIndex, depth = position447, tokenIndex447, depth447
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l449
							}
							position++
							goto l447
						l449:
							position, tokenIndex, depth = position447, tokenIndex447, depth447
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l450
							}
							position++
							goto l447
						l450:
							position, tokenIndex, depth = position447, tokenIndex447, depth447
							if buffer[position] != rune('_') {
								goto l451
							}
							position++
							goto l447
						l451:
							position, tokenIndex, depth = position447, tokenIndex447, depth447
							if buffer[position] != rune('-') {
								goto l446
							}
							position++
						}
					l447:
						goto l445
					l446:
						position, tokenIndex, depth = position446, tokenIndex446, depth446
					}
					goto l440
				l439:
					position, tokenIndex, depth = position439, tokenIndex439, depth439
				}
			l440:
				depth--
				add(ruleKey, position427)
			}
			return true
		l426:
			position, tokenIndex, depth = position426, tokenIndex426, depth426
			return false
		},
		/* 103 Index <- <('[' '-'? [0-9]+ ']')> */
		func() bool {
			position452, tokenIndex452, depth452 := position, tokenIndex, depth
			{
				position453 := position
				depth++
				if buffer[position] != rune('[') {
					goto l452
				}
				position++
				{
					position454, tokenIndex454, depth454 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l454
					}
					position++
					goto l455
				l454:
					position, tokenIndex, depth = position454, tokenIndex454, depth454
				}
			l455:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l452
				}
				position++
			l456:
				{
					position457, tokenIndex457, depth457 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l457
					}
					position++
					goto l456
				l457:
					position, tokenIndex, depth = position457, tokenIndex457, depth457
				}
				if buffer[position] != rune(']') {
					goto l452
				}
				position++
				depth--
				add(ruleIndex, position453)
			}
			return true
		l452:
			position, tokenIndex, depth = position452, tokenIndex452, depth452
			return false
		},
		/* 104 IP <- <([0-9]+ '.' [0-9]+ '.' [0-9]+ '.' [0-9]+)> */
		func() bool {
			position458, tokenIndex458, depth458 := position, tokenIndex, depth
			{
				position459 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l458
				}
				position++
			l460:
				{
					position461, tokenIndex461, depth461 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l461
					}
					position++
					goto l460
				l461:
					position, tokenIndex, depth = position461, tokenIndex461, depth461
				}
				if buffer[position] != rune('.') {
					goto l458
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l458
				}
				position++
			l462:
				{
					position463, tokenIndex463, depth463 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l463
					}
					position++
					goto l462
				l463:
					position, tokenIndex, depth = position463, tokenIndex463, depth463
				}
				if buffer[position] != rune('.') {
					goto l458
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l458
				}
				position++
			l464:
				{
					position465, tokenIndex465, depth465 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l465
					}
					position++
					goto l464
				l465:
					position, tokenIndex, depth = position465, tokenIndex465, depth465
				}
				if buffer[position] != rune('.') {
					goto l458
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l458
				}
				position++
			l466:
				{
					position467, tokenIndex467, depth467 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l467
					}
					position++
					goto l466
				l467:
					position, tokenIndex, depth = position467, tokenIndex467, depth467
				}
				depth--
				add(ruleIP, position459)
			}
			return true
		l458:
			position, tokenIndex, depth = position458, tokenIndex458, depth458
			return false
		},
		/* 105 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position469 := position
				depth++
			l470:
				{
					position471, tokenIndex471, depth471 := position, tokenIndex, depth
					{
						position472, tokenIndex472, depth472 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l473
						}
						position++
						goto l472
					l473:
						position, tokenIndex, depth = position472, tokenIndex472, depth472
						if buffer[position] != rune('\t') {
							goto l474
						}
						position++
						goto l472
					l474:
						position, tokenIndex, depth = position472, tokenIndex472, depth472
						if buffer[position] != rune('\n') {
							goto l475
						}
						position++
						goto l472
					l475:
						position, tokenIndex, depth = position472, tokenIndex472, depth472
						if buffer[position] != rune('\r') {
							goto l471
						}
						position++
					}
				l472:
					goto l470
				l471:
					position, tokenIndex, depth = position471, tokenIndex471, depth471
				}
				depth--
				add(rulews, position469)
			}
			return true
		},
		/* 106 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position476, tokenIndex476, depth476 := position, tokenIndex, depth
			{
				position477 := position
				depth++
				{
					position480, tokenIndex480, depth480 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l481
					}
					position++
					goto l480
				l481:
					position, tokenIndex, depth = position480, tokenIndex480, depth480
					if buffer[position] != rune('\t') {
						goto l482
					}
					position++
					goto l480
				l482:
					position, tokenIndex, depth = position480, tokenIndex480, depth480
					if buffer[position] != rune('\n') {
						goto l483
					}
					position++
					goto l480
				l483:
					position, tokenIndex, depth = position480, tokenIndex480, depth480
					if buffer[position] != rune('\r') {
						goto l476
					}
					position++
				}
			l480:
			l478:
				{
					position479, tokenIndex479, depth479 := position, tokenIndex, depth
					{
						position484, tokenIndex484, depth484 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l485
						}
						position++
						goto l484
					l485:
						position, tokenIndex, depth = position484, tokenIndex484, depth484
						if buffer[position] != rune('\t') {
							goto l486
						}
						position++
						goto l484
					l486:
						position, tokenIndex, depth = position484, tokenIndex484, depth484
						if buffer[position] != rune('\n') {
							goto l487
						}
						position++
						goto l484
					l487:
						position, tokenIndex, depth = position484, tokenIndex484, depth484
						if buffer[position] != rune('\r') {
							goto l479
						}
						position++
					}
				l484:
					goto l478
				l479:
					position, tokenIndex, depth = position479, tokenIndex479, depth479
				}
				depth--
				add(rulereq_ws, position477)
			}
			return true
		l476:
			position, tokenIndex, depth = position476, tokenIndex476, depth476
			return false
		},
		/* 108 Action0 <- <{}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 109 Action1 <- <{}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 110 Action2 <- <{}> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
	}
	p.rules = _rules
}
