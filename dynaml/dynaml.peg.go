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
	rules  [107]func() bool
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
		/* 4 Marker <- <('&' (('t' 'e' 'm' 'p' 'l' 'a' 't' 'e') / ('t' 'e' 'm' 'p' 'o' 'r' 'a' 'r' 'y') / ('l' 'o' 'c' 'a' 'l') / ('i' 'n' 'j' 'e' 'c' 't') / ('s' 't' 'a' 't' 'e') / ('d' 'e' 'f' 'a' 'u' 'l' 't') / TagMarker))> */
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
			position25, tokenIndex25, depth25 := position, tokenIndex, depth
			{
				position26 := position
				depth++
				if buffer[position] != rune('t') {
					goto l25
				}
				position++
				if buffer[position] != rune('a') {
					goto l25
				}
				position++
				if buffer[position] != rune('g') {
					goto l25
				}
				position++
				if buffer[position] != rune(':') {
					goto l25
				}
				position++
				{
					position27, tokenIndex27, depth27 := position, tokenIndex, depth
					if buffer[position] != rune('*') {
						goto l27
					}
					position++
					goto l28
				l27:
					position, tokenIndex, depth = position27, tokenIndex27, depth27
				}
			l28:
				if !_rules[ruleTag]() {
					goto l25
				}
				depth--
				add(ruleTagMarker, position26)
			}
			return true
		l25:
			position, tokenIndex, depth = position25, tokenIndex25, depth25
			return false
		},
		/* 6 MarkerExpression <- <Grouped> */
		func() bool {
			position29, tokenIndex29, depth29 := position, tokenIndex, depth
			{
				position30 := position
				depth++
				if !_rules[ruleGrouped]() {
					goto l29
				}
				depth--
				add(ruleMarkerExpression, position30)
			}
			return true
		l29:
			position, tokenIndex, depth = position29, tokenIndex29, depth29
			return false
		},
		/* 7 Expression <- <((Scoped / LambdaExpr / Level7) ws)> */
		func() bool {
			position31, tokenIndex31, depth31 := position, tokenIndex, depth
			{
				position32 := position
				depth++
				{
					position33, tokenIndex33, depth33 := position, tokenIndex, depth
					if !_rules[ruleScoped]() {
						goto l34
					}
					goto l33
				l34:
					position, tokenIndex, depth = position33, tokenIndex33, depth33
					if !_rules[ruleLambdaExpr]() {
						goto l35
					}
					goto l33
				l35:
					position, tokenIndex, depth = position33, tokenIndex33, depth33
					if !_rules[ruleLevel7]() {
						goto l31
					}
				}
			l33:
				if !_rules[rulews]() {
					goto l31
				}
				depth--
				add(ruleExpression, position32)
			}
			return true
		l31:
			position, tokenIndex, depth = position31, tokenIndex31, depth31
			return false
		},
		/* 8 Scoped <- <(ws Scope ws Expression)> */
		func() bool {
			position36, tokenIndex36, depth36 := position, tokenIndex, depth
			{
				position37 := position
				depth++
				if !_rules[rulews]() {
					goto l36
				}
				if !_rules[ruleScope]() {
					goto l36
				}
				if !_rules[rulews]() {
					goto l36
				}
				if !_rules[ruleExpression]() {
					goto l36
				}
				depth--
				add(ruleScoped, position37)
			}
			return true
		l36:
			position, tokenIndex, depth = position36, tokenIndex36, depth36
			return false
		},
		/* 9 Scope <- <(CreateScope ws Assignments? ')')> */
		func() bool {
			position38, tokenIndex38, depth38 := position, tokenIndex, depth
			{
				position39 := position
				depth++
				if !_rules[ruleCreateScope]() {
					goto l38
				}
				if !_rules[rulews]() {
					goto l38
				}
				{
					position40, tokenIndex40, depth40 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l40
					}
					goto l41
				l40:
					position, tokenIndex, depth = position40, tokenIndex40, depth40
				}
			l41:
				if buffer[position] != rune(')') {
					goto l38
				}
				position++
				depth--
				add(ruleScope, position39)
			}
			return true
		l38:
			position, tokenIndex, depth = position38, tokenIndex38, depth38
			return false
		},
		/* 10 CreateScope <- <'('> */
		func() bool {
			position42, tokenIndex42, depth42 := position, tokenIndex, depth
			{
				position43 := position
				depth++
				if buffer[position] != rune('(') {
					goto l42
				}
				position++
				depth--
				add(ruleCreateScope, position43)
			}
			return true
		l42:
			position, tokenIndex, depth = position42, tokenIndex42, depth42
			return false
		},
		/* 11 Level7 <- <(ws Level6 (req_ws Or)*)> */
		func() bool {
			position44, tokenIndex44, depth44 := position, tokenIndex, depth
			{
				position45 := position
				depth++
				if !_rules[rulews]() {
					goto l44
				}
				if !_rules[ruleLevel6]() {
					goto l44
				}
			l46:
				{
					position47, tokenIndex47, depth47 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l47
					}
					if !_rules[ruleOr]() {
						goto l47
					}
					goto l46
				l47:
					position, tokenIndex, depth = position47, tokenIndex47, depth47
				}
				depth--
				add(ruleLevel7, position45)
			}
			return true
		l44:
			position, tokenIndex, depth = position44, tokenIndex44, depth44
			return false
		},
		/* 12 Or <- <(OrOp req_ws Level6)> */
		func() bool {
			position48, tokenIndex48, depth48 := position, tokenIndex, depth
			{
				position49 := position
				depth++
				if !_rules[ruleOrOp]() {
					goto l48
				}
				if !_rules[rulereq_ws]() {
					goto l48
				}
				if !_rules[ruleLevel6]() {
					goto l48
				}
				depth--
				add(ruleOr, position49)
			}
			return true
		l48:
			position, tokenIndex, depth = position48, tokenIndex48, depth48
			return false
		},
		/* 13 OrOp <- <(('|' '|') / ('/' '/'))> */
		func() bool {
			position50, tokenIndex50, depth50 := position, tokenIndex, depth
			{
				position51 := position
				depth++
				{
					position52, tokenIndex52, depth52 := position, tokenIndex, depth
					if buffer[position] != rune('|') {
						goto l53
					}
					position++
					if buffer[position] != rune('|') {
						goto l53
					}
					position++
					goto l52
				l53:
					position, tokenIndex, depth = position52, tokenIndex52, depth52
					if buffer[position] != rune('/') {
						goto l50
					}
					position++
					if buffer[position] != rune('/') {
						goto l50
					}
					position++
				}
			l52:
				depth--
				add(ruleOrOp, position51)
			}
			return true
		l50:
			position, tokenIndex, depth = position50, tokenIndex50, depth50
			return false
		},
		/* 14 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position54, tokenIndex54, depth54 := position, tokenIndex, depth
			{
				position55 := position
				depth++
				{
					position56, tokenIndex56, depth56 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l57
					}
					goto l56
				l57:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
					if !_rules[ruleLevel5]() {
						goto l54
					}
				}
			l56:
				depth--
				add(ruleLevel6, position55)
			}
			return true
		l54:
			position, tokenIndex, depth = position54, tokenIndex54, depth54
			return false
		},
		/* 15 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position58, tokenIndex58, depth58 := position, tokenIndex, depth
			{
				position59 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l58
				}
				if !_rules[rulews]() {
					goto l58
				}
				if buffer[position] != rune('?') {
					goto l58
				}
				position++
				if !_rules[ruleExpression]() {
					goto l58
				}
				if buffer[position] != rune(':') {
					goto l58
				}
				position++
				if !_rules[ruleExpression]() {
					goto l58
				}
				depth--
				add(ruleConditional, position59)
			}
			return true
		l58:
			position, tokenIndex, depth = position58, tokenIndex58, depth58
			return false
		},
		/* 16 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position60, tokenIndex60, depth60 := position, tokenIndex, depth
			{
				position61 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l60
				}
			l62:
				{
					position63, tokenIndex63, depth63 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l63
					}
					goto l62
				l63:
					position, tokenIndex, depth = position63, tokenIndex63, depth63
				}
				depth--
				add(ruleLevel5, position61)
			}
			return true
		l60:
			position, tokenIndex, depth = position60, tokenIndex60, depth60
			return false
		},
		/* 17 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position64, tokenIndex64, depth64 := position, tokenIndex, depth
			{
				position65 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l64
				}
				if !_rules[ruleLevel4]() {
					goto l64
				}
				depth--
				add(ruleConcatenation, position65)
			}
			return true
		l64:
			position, tokenIndex, depth = position64, tokenIndex64, depth64
			return false
		},
		/* 18 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position66, tokenIndex66, depth66 := position, tokenIndex, depth
			{
				position67 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l66
				}
			l68:
				{
					position69, tokenIndex69, depth69 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l69
					}
					{
						position70, tokenIndex70, depth70 := position, tokenIndex, depth
						if !_rules[ruleLogOr]() {
							goto l71
						}
						goto l70
					l71:
						position, tokenIndex, depth = position70, tokenIndex70, depth70
						if !_rules[ruleLogAnd]() {
							goto l69
						}
					}
				l70:
					goto l68
				l69:
					position, tokenIndex, depth = position69, tokenIndex69, depth69
				}
				depth--
				add(ruleLevel4, position67)
			}
			return true
		l66:
			position, tokenIndex, depth = position66, tokenIndex66, depth66
			return false
		},
		/* 19 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position72, tokenIndex72, depth72 := position, tokenIndex, depth
			{
				position73 := position
				depth++
				if buffer[position] != rune('-') {
					goto l72
				}
				position++
				if buffer[position] != rune('o') {
					goto l72
				}
				position++
				if buffer[position] != rune('r') {
					goto l72
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l72
				}
				if !_rules[ruleLevel3]() {
					goto l72
				}
				depth--
				add(ruleLogOr, position73)
			}
			return true
		l72:
			position, tokenIndex, depth = position72, tokenIndex72, depth72
			return false
		},
		/* 20 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position74, tokenIndex74, depth74 := position, tokenIndex, depth
			{
				position75 := position
				depth++
				if buffer[position] != rune('-') {
					goto l74
				}
				position++
				if buffer[position] != rune('a') {
					goto l74
				}
				position++
				if buffer[position] != rune('n') {
					goto l74
				}
				position++
				if buffer[position] != rune('d') {
					goto l74
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l74
				}
				if !_rules[ruleLevel3]() {
					goto l74
				}
				depth--
				add(ruleLogAnd, position75)
			}
			return true
		l74:
			position, tokenIndex, depth = position74, tokenIndex74, depth74
			return false
		},
		/* 21 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position76, tokenIndex76, depth76 := position, tokenIndex, depth
			{
				position77 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l76
				}
			l78:
				{
					position79, tokenIndex79, depth79 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l79
					}
					if !_rules[ruleComparison]() {
						goto l79
					}
					goto l78
				l79:
					position, tokenIndex, depth = position79, tokenIndex79, depth79
				}
				depth--
				add(ruleLevel3, position77)
			}
			return true
		l76:
			position, tokenIndex, depth = position76, tokenIndex76, depth76
			return false
		},
		/* 22 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position80, tokenIndex80, depth80 := position, tokenIndex, depth
			{
				position81 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l80
				}
				if !_rules[rulereq_ws]() {
					goto l80
				}
				if !_rules[ruleLevel2]() {
					goto l80
				}
				depth--
				add(ruleComparison, position81)
			}
			return true
		l80:
			position, tokenIndex, depth = position80, tokenIndex80, depth80
			return false
		},
		/* 23 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position82, tokenIndex82, depth82 := position, tokenIndex, depth
			{
				position83 := position
				depth++
				{
					position84, tokenIndex84, depth84 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l85
					}
					position++
					if buffer[position] != rune('=') {
						goto l85
					}
					position++
					goto l84
				l85:
					position, tokenIndex, depth = position84, tokenIndex84, depth84
					if buffer[position] != rune('!') {
						goto l86
					}
					position++
					if buffer[position] != rune('=') {
						goto l86
					}
					position++
					goto l84
				l86:
					position, tokenIndex, depth = position84, tokenIndex84, depth84
					if buffer[position] != rune('<') {
						goto l87
					}
					position++
					if buffer[position] != rune('=') {
						goto l87
					}
					position++
					goto l84
				l87:
					position, tokenIndex, depth = position84, tokenIndex84, depth84
					if buffer[position] != rune('>') {
						goto l88
					}
					position++
					if buffer[position] != rune('=') {
						goto l88
					}
					position++
					goto l84
				l88:
					position, tokenIndex, depth = position84, tokenIndex84, depth84
					if buffer[position] != rune('>') {
						goto l89
					}
					position++
					goto l84
				l89:
					position, tokenIndex, depth = position84, tokenIndex84, depth84
					if buffer[position] != rune('<') {
						goto l90
					}
					position++
					goto l84
				l90:
					position, tokenIndex, depth = position84, tokenIndex84, depth84
					if buffer[position] != rune('>') {
						goto l82
					}
					position++
				}
			l84:
				depth--
				add(ruleCompareOp, position83)
			}
			return true
		l82:
			position, tokenIndex, depth = position82, tokenIndex82, depth82
			return false
		},
		/* 24 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position91, tokenIndex91, depth91 := position, tokenIndex, depth
			{
				position92 := position
				depth++
				if !_rules[ruleLevel1]() {
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
						if !_rules[ruleAddition]() {
							goto l96
						}
						goto l95
					l96:
						position, tokenIndex, depth = position95, tokenIndex95, depth95
						if !_rules[ruleSubtraction]() {
							goto l94
						}
					}
				l95:
					goto l93
				l94:
					position, tokenIndex, depth = position94, tokenIndex94, depth94
				}
				depth--
				add(ruleLevel2, position92)
			}
			return true
		l91:
			position, tokenIndex, depth = position91, tokenIndex91, depth91
			return false
		},
		/* 25 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position97, tokenIndex97, depth97 := position, tokenIndex, depth
			{
				position98 := position
				depth++
				if buffer[position] != rune('+') {
					goto l97
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l97
				}
				if !_rules[ruleLevel1]() {
					goto l97
				}
				depth--
				add(ruleAddition, position98)
			}
			return true
		l97:
			position, tokenIndex, depth = position97, tokenIndex97, depth97
			return false
		},
		/* 26 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position99, tokenIndex99, depth99 := position, tokenIndex, depth
			{
				position100 := position
				depth++
				if buffer[position] != rune('-') {
					goto l99
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l99
				}
				if !_rules[ruleLevel1]() {
					goto l99
				}
				depth--
				add(ruleSubtraction, position100)
			}
			return true
		l99:
			position, tokenIndex, depth = position99, tokenIndex99, depth99
			return false
		},
		/* 27 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position101, tokenIndex101, depth101 := position, tokenIndex, depth
			{
				position102 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l101
				}
			l103:
				{
					position104, tokenIndex104, depth104 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l104
					}
					{
						position105, tokenIndex105, depth105 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l106
						}
						goto l105
					l106:
						position, tokenIndex, depth = position105, tokenIndex105, depth105
						if !_rules[ruleDivision]() {
							goto l107
						}
						goto l105
					l107:
						position, tokenIndex, depth = position105, tokenIndex105, depth105
						if !_rules[ruleModulo]() {
							goto l104
						}
					}
				l105:
					goto l103
				l104:
					position, tokenIndex, depth = position104, tokenIndex104, depth104
				}
				depth--
				add(ruleLevel1, position102)
			}
			return true
		l101:
			position, tokenIndex, depth = position101, tokenIndex101, depth101
			return false
		},
		/* 28 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position108, tokenIndex108, depth108 := position, tokenIndex, depth
			{
				position109 := position
				depth++
				if buffer[position] != rune('*') {
					goto l108
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l108
				}
				if !_rules[ruleLevel0]() {
					goto l108
				}
				depth--
				add(ruleMultiplication, position109)
			}
			return true
		l108:
			position, tokenIndex, depth = position108, tokenIndex108, depth108
			return false
		},
		/* 29 Division <- <('/' req_ws Level0)> */
		func() bool {
			position110, tokenIndex110, depth110 := position, tokenIndex, depth
			{
				position111 := position
				depth++
				if buffer[position] != rune('/') {
					goto l110
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l110
				}
				if !_rules[ruleLevel0]() {
					goto l110
				}
				depth--
				add(ruleDivision, position111)
			}
			return true
		l110:
			position, tokenIndex, depth = position110, tokenIndex110, depth110
			return false
		},
		/* 30 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position112, tokenIndex112, depth112 := position, tokenIndex, depth
			{
				position113 := position
				depth++
				if buffer[position] != rune('%') {
					goto l112
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l112
				}
				if !_rules[ruleLevel0]() {
					goto l112
				}
				depth--
				add(ruleModulo, position113)
			}
			return true
		l112:
			position, tokenIndex, depth = position112, tokenIndex112, depth112
			return false
		},
		/* 31 Level0 <- <(IP / String / Number / Boolean / Undefined / Nil / Symbol / Not / Substitution / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position114, tokenIndex114, depth114 := position, tokenIndex, depth
			{
				position115 := position
				depth++
				{
					position116, tokenIndex116, depth116 := position, tokenIndex, depth
					if !_rules[ruleIP]() {
						goto l117
					}
					goto l116
				l117:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					if !_rules[ruleString]() {
						goto l118
					}
					goto l116
				l118:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					if !_rules[ruleNumber]() {
						goto l119
					}
					goto l116
				l119:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					if !_rules[ruleBoolean]() {
						goto l120
					}
					goto l116
				l120:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					if !_rules[ruleUndefined]() {
						goto l121
					}
					goto l116
				l121:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					if !_rules[ruleNil]() {
						goto l122
					}
					goto l116
				l122:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					if !_rules[ruleSymbol]() {
						goto l123
					}
					goto l116
				l123:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					if !_rules[ruleNot]() {
						goto l124
					}
					goto l116
				l124:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					if !_rules[ruleSubstitution]() {
						goto l125
					}
					goto l116
				l125:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					if !_rules[ruleMerge]() {
						goto l126
					}
					goto l116
				l126:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					if !_rules[ruleAuto]() {
						goto l127
					}
					goto l116
				l127:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					if !_rules[ruleLambda]() {
						goto l128
					}
					goto l116
				l128:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					if !_rules[ruleChained]() {
						goto l114
					}
				}
			l116:
				depth--
				add(ruleLevel0, position115)
			}
			return true
		l114:
			position, tokenIndex, depth = position114, tokenIndex114, depth114
			return false
		},
		/* 32 Chained <- <((MapMapping / Sync / Catch / Mapping / MapSelection / Selection / Sum / List / Map / Range / Grouped / Reference) ChainedQualifiedExpression*)> */
		func() bool {
			position129, tokenIndex129, depth129 := position, tokenIndex, depth
			{
				position130 := position
				depth++
				{
					position131, tokenIndex131, depth131 := position, tokenIndex, depth
					if !_rules[ruleMapMapping]() {
						goto l132
					}
					goto l131
				l132:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					if !_rules[ruleSync]() {
						goto l133
					}
					goto l131
				l133:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					if !_rules[ruleCatch]() {
						goto l134
					}
					goto l131
				l134:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					if !_rules[ruleMapping]() {
						goto l135
					}
					goto l131
				l135:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					if !_rules[ruleMapSelection]() {
						goto l136
					}
					goto l131
				l136:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					if !_rules[ruleSelection]() {
						goto l137
					}
					goto l131
				l137:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					if !_rules[ruleSum]() {
						goto l138
					}
					goto l131
				l138:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					if !_rules[ruleList]() {
						goto l139
					}
					goto l131
				l139:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					if !_rules[ruleMap]() {
						goto l140
					}
					goto l131
				l140:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					if !_rules[ruleRange]() {
						goto l141
					}
					goto l131
				l141:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					if !_rules[ruleGrouped]() {
						goto l142
					}
					goto l131
				l142:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
					if !_rules[ruleReference]() {
						goto l129
					}
				}
			l131:
			l143:
				{
					position144, tokenIndex144, depth144 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l144
					}
					goto l143
				l144:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
				}
				depth--
				add(ruleChained, position130)
			}
			return true
		l129:
			position, tokenIndex, depth = position129, tokenIndex129, depth129
			return false
		},
		/* 33 ChainedQualifiedExpression <- <(ChainedCall / Currying / ChainedRef / ChainedDynRef / Projection)> */
		func() bool {
			position145, tokenIndex145, depth145 := position, tokenIndex, depth
			{
				position146 := position
				depth++
				{
					position147, tokenIndex147, depth147 := position, tokenIndex, depth
					if !_rules[ruleChainedCall]() {
						goto l148
					}
					goto l147
				l148:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
					if !_rules[ruleCurrying]() {
						goto l149
					}
					goto l147
				l149:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
					if !_rules[ruleChainedRef]() {
						goto l150
					}
					goto l147
				l150:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
					if !_rules[ruleChainedDynRef]() {
						goto l151
					}
					goto l147
				l151:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
					if !_rules[ruleProjection]() {
						goto l145
					}
				}
			l147:
				depth--
				add(ruleChainedQualifiedExpression, position146)
			}
			return true
		l145:
			position, tokenIndex, depth = position145, tokenIndex145, depth145
			return false
		},
		/* 34 ChainedRef <- <(PathComponent FollowUpRef)> */
		func() bool {
			position152, tokenIndex152, depth152 := position, tokenIndex, depth
			{
				position153 := position
				depth++
				if !_rules[rulePathComponent]() {
					goto l152
				}
				if !_rules[ruleFollowUpRef]() {
					goto l152
				}
				depth--
				add(ruleChainedRef, position153)
			}
			return true
		l152:
			position, tokenIndex, depth = position152, tokenIndex152, depth152
			return false
		},
		/* 35 ChainedDynRef <- <('.'? '[' Expression ']')> */
		func() bool {
			position154, tokenIndex154, depth154 := position, tokenIndex, depth
			{
				position155 := position
				depth++
				{
					position156, tokenIndex156, depth156 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l156
					}
					position++
					goto l157
				l156:
					position, tokenIndex, depth = position156, tokenIndex156, depth156
				}
			l157:
				if buffer[position] != rune('[') {
					goto l154
				}
				position++
				if !_rules[ruleExpression]() {
					goto l154
				}
				if buffer[position] != rune(']') {
					goto l154
				}
				position++
				depth--
				add(ruleChainedDynRef, position155)
			}
			return true
		l154:
			position, tokenIndex, depth = position154, tokenIndex154, depth154
			return false
		},
		/* 36 Slice <- <Range> */
		func() bool {
			position158, tokenIndex158, depth158 := position, tokenIndex, depth
			{
				position159 := position
				depth++
				if !_rules[ruleRange]() {
					goto l158
				}
				depth--
				add(ruleSlice, position159)
			}
			return true
		l158:
			position, tokenIndex, depth = position158, tokenIndex158, depth158
			return false
		},
		/* 37 Currying <- <('*' ChainedCall)> */
		func() bool {
			position160, tokenIndex160, depth160 := position, tokenIndex, depth
			{
				position161 := position
				depth++
				if buffer[position] != rune('*') {
					goto l160
				}
				position++
				if !_rules[ruleChainedCall]() {
					goto l160
				}
				depth--
				add(ruleCurrying, position161)
			}
			return true
		l160:
			position, tokenIndex, depth = position160, tokenIndex160, depth160
			return false
		},
		/* 38 ChainedCall <- <(StartArguments NameArgumentList? ')')> */
		func() bool {
			position162, tokenIndex162, depth162 := position, tokenIndex, depth
			{
				position163 := position
				depth++
				if !_rules[ruleStartArguments]() {
					goto l162
				}
				{
					position164, tokenIndex164, depth164 := position, tokenIndex, depth
					if !_rules[ruleNameArgumentList]() {
						goto l164
					}
					goto l165
				l164:
					position, tokenIndex, depth = position164, tokenIndex164, depth164
				}
			l165:
				if buffer[position] != rune(')') {
					goto l162
				}
				position++
				depth--
				add(ruleChainedCall, position163)
			}
			return true
		l162:
			position, tokenIndex, depth = position162, tokenIndex162, depth162
			return false
		},
		/* 39 StartArguments <- <('(' ws)> */
		func() bool {
			position166, tokenIndex166, depth166 := position, tokenIndex, depth
			{
				position167 := position
				depth++
				if buffer[position] != rune('(') {
					goto l166
				}
				position++
				if !_rules[rulews]() {
					goto l166
				}
				depth--
				add(ruleStartArguments, position167)
			}
			return true
		l166:
			position, tokenIndex, depth = position166, tokenIndex166, depth166
			return false
		},
		/* 40 NameArgumentList <- <(((NextNameArgument (',' NextNameArgument)*) / NextExpression) (',' NextExpression)*)> */
		func() bool {
			position168, tokenIndex168, depth168 := position, tokenIndex, depth
			{
				position169 := position
				depth++
				{
					position170, tokenIndex170, depth170 := position, tokenIndex, depth
					if !_rules[ruleNextNameArgument]() {
						goto l171
					}
				l172:
					{
						position173, tokenIndex173, depth173 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l173
						}
						position++
						if !_rules[ruleNextNameArgument]() {
							goto l173
						}
						goto l172
					l173:
						position, tokenIndex, depth = position173, tokenIndex173, depth173
					}
					goto l170
				l171:
					position, tokenIndex, depth = position170, tokenIndex170, depth170
					if !_rules[ruleNextExpression]() {
						goto l168
					}
				}
			l170:
			l174:
				{
					position175, tokenIndex175, depth175 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l175
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l175
					}
					goto l174
				l175:
					position, tokenIndex, depth = position175, tokenIndex175, depth175
				}
				depth--
				add(ruleNameArgumentList, position169)
			}
			return true
		l168:
			position, tokenIndex, depth = position168, tokenIndex168, depth168
			return false
		},
		/* 41 NextNameArgument <- <(ws Name ws '=' ws Expression ws)> */
		func() bool {
			position176, tokenIndex176, depth176 := position, tokenIndex, depth
			{
				position177 := position
				depth++
				if !_rules[rulews]() {
					goto l176
				}
				if !_rules[ruleName]() {
					goto l176
				}
				if !_rules[rulews]() {
					goto l176
				}
				if buffer[position] != rune('=') {
					goto l176
				}
				position++
				if !_rules[rulews]() {
					goto l176
				}
				if !_rules[ruleExpression]() {
					goto l176
				}
				if !_rules[rulews]() {
					goto l176
				}
				depth--
				add(ruleNextNameArgument, position177)
			}
			return true
		l176:
			position, tokenIndex, depth = position176, tokenIndex176, depth176
			return false
		},
		/* 42 ExpressionList <- <(NextExpression (',' NextExpression)*)> */
		func() bool {
			position178, tokenIndex178, depth178 := position, tokenIndex, depth
			{
				position179 := position
				depth++
				if !_rules[ruleNextExpression]() {
					goto l178
				}
			l180:
				{
					position181, tokenIndex181, depth181 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l181
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l181
					}
					goto l180
				l181:
					position, tokenIndex, depth = position181, tokenIndex181, depth181
				}
				depth--
				add(ruleExpressionList, position179)
			}
			return true
		l178:
			position, tokenIndex, depth = position178, tokenIndex178, depth178
			return false
		},
		/* 43 NextExpression <- <(Expression ListExpansion?)> */
		func() bool {
			position182, tokenIndex182, depth182 := position, tokenIndex, depth
			{
				position183 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l182
				}
				{
					position184, tokenIndex184, depth184 := position, tokenIndex, depth
					if !_rules[ruleListExpansion]() {
						goto l184
					}
					goto l185
				l184:
					position, tokenIndex, depth = position184, tokenIndex184, depth184
				}
			l185:
				depth--
				add(ruleNextExpression, position183)
			}
			return true
		l182:
			position, tokenIndex, depth = position182, tokenIndex182, depth182
			return false
		},
		/* 44 ListExpansion <- <('.' '.' '.' ws)> */
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
				if buffer[position] != rune('.') {
					goto l186
				}
				position++
				if !_rules[rulews]() {
					goto l186
				}
				depth--
				add(ruleListExpansion, position187)
			}
			return true
		l186:
			position, tokenIndex, depth = position186, tokenIndex186, depth186
			return false
		},
		/* 45 Projection <- <('.'? (('[' '*' ']') / Slice) ProjectionValue ChainedQualifiedExpression*)> */
		func() bool {
			position188, tokenIndex188, depth188 := position, tokenIndex, depth
			{
				position189 := position
				depth++
				{
					position190, tokenIndex190, depth190 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l190
					}
					position++
					goto l191
				l190:
					position, tokenIndex, depth = position190, tokenIndex190, depth190
				}
			l191:
				{
					position192, tokenIndex192, depth192 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l193
					}
					position++
					if buffer[position] != rune('*') {
						goto l193
					}
					position++
					if buffer[position] != rune(']') {
						goto l193
					}
					position++
					goto l192
				l193:
					position, tokenIndex, depth = position192, tokenIndex192, depth192
					if !_rules[ruleSlice]() {
						goto l188
					}
				}
			l192:
				if !_rules[ruleProjectionValue]() {
					goto l188
				}
			l194:
				{
					position195, tokenIndex195, depth195 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l195
					}
					goto l194
				l195:
					position, tokenIndex, depth = position195, tokenIndex195, depth195
				}
				depth--
				add(ruleProjection, position189)
			}
			return true
		l188:
			position, tokenIndex, depth = position188, tokenIndex188, depth188
			return false
		},
		/* 46 ProjectionValue <- <Action0> */
		func() bool {
			position196, tokenIndex196, depth196 := position, tokenIndex, depth
			{
				position197 := position
				depth++
				if !_rules[ruleAction0]() {
					goto l196
				}
				depth--
				add(ruleProjectionValue, position197)
			}
			return true
		l196:
			position, tokenIndex, depth = position196, tokenIndex196, depth196
			return false
		},
		/* 47 Substitution <- <('*' Level0)> */
		func() bool {
			position198, tokenIndex198, depth198 := position, tokenIndex, depth
			{
				position199 := position
				depth++
				if buffer[position] != rune('*') {
					goto l198
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l198
				}
				depth--
				add(ruleSubstitution, position199)
			}
			return true
		l198:
			position, tokenIndex, depth = position198, tokenIndex198, depth198
			return false
		},
		/* 48 Not <- <('!' ws Level0)> */
		func() bool {
			position200, tokenIndex200, depth200 := position, tokenIndex, depth
			{
				position201 := position
				depth++
				if buffer[position] != rune('!') {
					goto l200
				}
				position++
				if !_rules[rulews]() {
					goto l200
				}
				if !_rules[ruleLevel0]() {
					goto l200
				}
				depth--
				add(ruleNot, position201)
			}
			return true
		l200:
			position, tokenIndex, depth = position200, tokenIndex200, depth200
			return false
		},
		/* 49 Grouped <- <('(' Expression ')')> */
		func() bool {
			position202, tokenIndex202, depth202 := position, tokenIndex, depth
			{
				position203 := position
				depth++
				if buffer[position] != rune('(') {
					goto l202
				}
				position++
				if !_rules[ruleExpression]() {
					goto l202
				}
				if buffer[position] != rune(')') {
					goto l202
				}
				position++
				depth--
				add(ruleGrouped, position203)
			}
			return true
		l202:
			position, tokenIndex, depth = position202, tokenIndex202, depth202
			return false
		},
		/* 50 Range <- <(StartRange Expression? RangeOp Expression? ']')> */
		func() bool {
			position204, tokenIndex204, depth204 := position, tokenIndex, depth
			{
				position205 := position
				depth++
				if !_rules[ruleStartRange]() {
					goto l204
				}
				{
					position206, tokenIndex206, depth206 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l206
					}
					goto l207
				l206:
					position, tokenIndex, depth = position206, tokenIndex206, depth206
				}
			l207:
				if !_rules[ruleRangeOp]() {
					goto l204
				}
				{
					position208, tokenIndex208, depth208 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l208
					}
					goto l209
				l208:
					position, tokenIndex, depth = position208, tokenIndex208, depth208
				}
			l209:
				if buffer[position] != rune(']') {
					goto l204
				}
				position++
				depth--
				add(ruleRange, position205)
			}
			return true
		l204:
			position, tokenIndex, depth = position204, tokenIndex204, depth204
			return false
		},
		/* 51 StartRange <- <'['> */
		func() bool {
			position210, tokenIndex210, depth210 := position, tokenIndex, depth
			{
				position211 := position
				depth++
				if buffer[position] != rune('[') {
					goto l210
				}
				position++
				depth--
				add(ruleStartRange, position211)
			}
			return true
		l210:
			position, tokenIndex, depth = position210, tokenIndex210, depth210
			return false
		},
		/* 52 RangeOp <- <('.' '.')> */
		func() bool {
			position212, tokenIndex212, depth212 := position, tokenIndex, depth
			{
				position213 := position
				depth++
				if buffer[position] != rune('.') {
					goto l212
				}
				position++
				if buffer[position] != rune('.') {
					goto l212
				}
				position++
				depth--
				add(ruleRangeOp, position213)
			}
			return true
		l212:
			position, tokenIndex, depth = position212, tokenIndex212, depth212
			return false
		},
		/* 53 Number <- <('-'? [0-9] ([0-9] / '_')* ('.' [0-9] [0-9]*)? (('e' / 'E') '-'? [0-9] [0-9]*)? !(':' ':'))> */
		func() bool {
			position214, tokenIndex214, depth214 := position, tokenIndex, depth
			{
				position215 := position
				depth++
				{
					position216, tokenIndex216, depth216 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l216
					}
					position++
					goto l217
				l216:
					position, tokenIndex, depth = position216, tokenIndex216, depth216
				}
			l217:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l214
				}
				position++
			l218:
				{
					position219, tokenIndex219, depth219 := position, tokenIndex, depth
					{
						position220, tokenIndex220, depth220 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l221
						}
						position++
						goto l220
					l221:
						position, tokenIndex, depth = position220, tokenIndex220, depth220
						if buffer[position] != rune('_') {
							goto l219
						}
						position++
					}
				l220:
					goto l218
				l219:
					position, tokenIndex, depth = position219, tokenIndex219, depth219
				}
				{
					position222, tokenIndex222, depth222 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l222
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l222
					}
					position++
				l224:
					{
						position225, tokenIndex225, depth225 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l225
						}
						position++
						goto l224
					l225:
						position, tokenIndex, depth = position225, tokenIndex225, depth225
					}
					goto l223
				l222:
					position, tokenIndex, depth = position222, tokenIndex222, depth222
				}
			l223:
				{
					position226, tokenIndex226, depth226 := position, tokenIndex, depth
					{
						position228, tokenIndex228, depth228 := position, tokenIndex, depth
						if buffer[position] != rune('e') {
							goto l229
						}
						position++
						goto l228
					l229:
						position, tokenIndex, depth = position228, tokenIndex228, depth228
						if buffer[position] != rune('E') {
							goto l226
						}
						position++
					}
				l228:
					{
						position230, tokenIndex230, depth230 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l230
						}
						position++
						goto l231
					l230:
						position, tokenIndex, depth = position230, tokenIndex230, depth230
					}
				l231:
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l226
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
					goto l227
				l226:
					position, tokenIndex, depth = position226, tokenIndex226, depth226
				}
			l227:
				{
					position234, tokenIndex234, depth234 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l234
					}
					position++
					if buffer[position] != rune(':') {
						goto l234
					}
					position++
					goto l214
				l234:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
				}
				depth--
				add(ruleNumber, position215)
			}
			return true
		l214:
			position, tokenIndex, depth = position214, tokenIndex214, depth214
			return false
		},
		/* 54 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position235, tokenIndex235, depth235 := position, tokenIndex, depth
			{
				position236 := position
				depth++
				if buffer[position] != rune('"') {
					goto l235
				}
				position++
			l237:
				{
					position238, tokenIndex238, depth238 := position, tokenIndex, depth
					{
						position239, tokenIndex239, depth239 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l240
						}
						position++
						if buffer[position] != rune('"') {
							goto l240
						}
						position++
						goto l239
					l240:
						position, tokenIndex, depth = position239, tokenIndex239, depth239
						{
							position241, tokenIndex241, depth241 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l241
							}
							position++
							goto l238
						l241:
							position, tokenIndex, depth = position241, tokenIndex241, depth241
						}
						if !matchDot() {
							goto l238
						}
					}
				l239:
					goto l237
				l238:
					position, tokenIndex, depth = position238, tokenIndex238, depth238
				}
				if buffer[position] != rune('"') {
					goto l235
				}
				position++
				depth--
				add(ruleString, position236)
			}
			return true
		l235:
			position, tokenIndex, depth = position235, tokenIndex235, depth235
			return false
		},
		/* 55 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position242, tokenIndex242, depth242 := position, tokenIndex, depth
			{
				position243 := position
				depth++
				{
					position244, tokenIndex244, depth244 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l245
					}
					position++
					if buffer[position] != rune('r') {
						goto l245
					}
					position++
					if buffer[position] != rune('u') {
						goto l245
					}
					position++
					if buffer[position] != rune('e') {
						goto l245
					}
					position++
					goto l244
				l245:
					position, tokenIndex, depth = position244, tokenIndex244, depth244
					if buffer[position] != rune('f') {
						goto l242
					}
					position++
					if buffer[position] != rune('a') {
						goto l242
					}
					position++
					if buffer[position] != rune('l') {
						goto l242
					}
					position++
					if buffer[position] != rune('s') {
						goto l242
					}
					position++
					if buffer[position] != rune('e') {
						goto l242
					}
					position++
				}
			l244:
				depth--
				add(ruleBoolean, position243)
			}
			return true
		l242:
			position, tokenIndex, depth = position242, tokenIndex242, depth242
			return false
		},
		/* 56 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position246, tokenIndex246, depth246 := position, tokenIndex, depth
			{
				position247 := position
				depth++
				{
					position248, tokenIndex248, depth248 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l249
					}
					position++
					if buffer[position] != rune('i') {
						goto l249
					}
					position++
					if buffer[position] != rune('l') {
						goto l249
					}
					position++
					goto l248
				l249:
					position, tokenIndex, depth = position248, tokenIndex248, depth248
					if buffer[position] != rune('~') {
						goto l246
					}
					position++
				}
			l248:
				depth--
				add(ruleNil, position247)
			}
			return true
		l246:
			position, tokenIndex, depth = position246, tokenIndex246, depth246
			return false
		},
		/* 57 Undefined <- <('~' '~')> */
		func() bool {
			position250, tokenIndex250, depth250 := position, tokenIndex, depth
			{
				position251 := position
				depth++
				if buffer[position] != rune('~') {
					goto l250
				}
				position++
				if buffer[position] != rune('~') {
					goto l250
				}
				position++
				depth--
				add(ruleUndefined, position251)
			}
			return true
		l250:
			position, tokenIndex, depth = position250, tokenIndex250, depth250
			return false
		},
		/* 58 Symbol <- <('$' Name)> */
		func() bool {
			position252, tokenIndex252, depth252 := position, tokenIndex, depth
			{
				position253 := position
				depth++
				if buffer[position] != rune('$') {
					goto l252
				}
				position++
				if !_rules[ruleName]() {
					goto l252
				}
				depth--
				add(ruleSymbol, position253)
			}
			return true
		l252:
			position, tokenIndex, depth = position252, tokenIndex252, depth252
			return false
		},
		/* 59 List <- <(StartList ExpressionList? ']')> */
		func() bool {
			position254, tokenIndex254, depth254 := position, tokenIndex, depth
			{
				position255 := position
				depth++
				if !_rules[ruleStartList]() {
					goto l254
				}
				{
					position256, tokenIndex256, depth256 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l256
					}
					goto l257
				l256:
					position, tokenIndex, depth = position256, tokenIndex256, depth256
				}
			l257:
				if buffer[position] != rune(']') {
					goto l254
				}
				position++
				depth--
				add(ruleList, position255)
			}
			return true
		l254:
			position, tokenIndex, depth = position254, tokenIndex254, depth254
			return false
		},
		/* 60 StartList <- <('[' ws)> */
		func() bool {
			position258, tokenIndex258, depth258 := position, tokenIndex, depth
			{
				position259 := position
				depth++
				if buffer[position] != rune('[') {
					goto l258
				}
				position++
				if !_rules[rulews]() {
					goto l258
				}
				depth--
				add(ruleStartList, position259)
			}
			return true
		l258:
			position, tokenIndex, depth = position258, tokenIndex258, depth258
			return false
		},
		/* 61 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position260, tokenIndex260, depth260 := position, tokenIndex, depth
			{
				position261 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l260
				}
				if !_rules[rulews]() {
					goto l260
				}
				{
					position262, tokenIndex262, depth262 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l262
					}
					goto l263
				l262:
					position, tokenIndex, depth = position262, tokenIndex262, depth262
				}
			l263:
				if buffer[position] != rune('}') {
					goto l260
				}
				position++
				depth--
				add(ruleMap, position261)
			}
			return true
		l260:
			position, tokenIndex, depth = position260, tokenIndex260, depth260
			return false
		},
		/* 62 CreateMap <- <'{'> */
		func() bool {
			position264, tokenIndex264, depth264 := position, tokenIndex, depth
			{
				position265 := position
				depth++
				if buffer[position] != rune('{') {
					goto l264
				}
				position++
				depth--
				add(ruleCreateMap, position265)
			}
			return true
		l264:
			position, tokenIndex, depth = position264, tokenIndex264, depth264
			return false
		},
		/* 63 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position266, tokenIndex266, depth266 := position, tokenIndex, depth
			{
				position267 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l266
				}
			l268:
				{
					position269, tokenIndex269, depth269 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l269
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l269
					}
					goto l268
				l269:
					position, tokenIndex, depth = position269, tokenIndex269, depth269
				}
				depth--
				add(ruleAssignments, position267)
			}
			return true
		l266:
			position, tokenIndex, depth = position266, tokenIndex266, depth266
			return false
		},
		/* 64 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position270, tokenIndex270, depth270 := position, tokenIndex, depth
			{
				position271 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l270
				}
				if buffer[position] != rune('=') {
					goto l270
				}
				position++
				if !_rules[ruleExpression]() {
					goto l270
				}
				depth--
				add(ruleAssignment, position271)
			}
			return true
		l270:
			position, tokenIndex, depth = position270, tokenIndex270, depth270
			return false
		},
		/* 65 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position272, tokenIndex272, depth272 := position, tokenIndex, depth
			{
				position273 := position
				depth++
				{
					position274, tokenIndex274, depth274 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l275
					}
					goto l274
				l275:
					position, tokenIndex, depth = position274, tokenIndex274, depth274
					if !_rules[ruleSimpleMerge]() {
						goto l272
					}
				}
			l274:
				depth--
				add(ruleMerge, position273)
			}
			return true
		l272:
			position, tokenIndex, depth = position272, tokenIndex272, depth272
			return false
		},
		/* 66 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position276, tokenIndex276, depth276 := position, tokenIndex, depth
			{
				position277 := position
				depth++
				if buffer[position] != rune('m') {
					goto l276
				}
				position++
				if buffer[position] != rune('e') {
					goto l276
				}
				position++
				if buffer[position] != rune('r') {
					goto l276
				}
				position++
				if buffer[position] != rune('g') {
					goto l276
				}
				position++
				if buffer[position] != rune('e') {
					goto l276
				}
				position++
				{
					position278, tokenIndex278, depth278 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l278
					}
					if !_rules[ruleRequired]() {
						goto l278
					}
					goto l276
				l278:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
				}
				{
					position279, tokenIndex279, depth279 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l279
					}
					{
						position281, tokenIndex281, depth281 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l282
						}
						goto l281
					l282:
						position, tokenIndex, depth = position281, tokenIndex281, depth281
						if !_rules[ruleOn]() {
							goto l279
						}
					}
				l281:
					goto l280
				l279:
					position, tokenIndex, depth = position279, tokenIndex279, depth279
				}
			l280:
				if !_rules[rulereq_ws]() {
					goto l276
				}
				if !_rules[ruleReference]() {
					goto l276
				}
				depth--
				add(ruleRefMerge, position277)
			}
			return true
		l276:
			position, tokenIndex, depth = position276, tokenIndex276, depth276
			return false
		},
		/* 67 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' !'(' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position283, tokenIndex283, depth283 := position, tokenIndex, depth
			{
				position284 := position
				depth++
				if buffer[position] != rune('m') {
					goto l283
				}
				position++
				if buffer[position] != rune('e') {
					goto l283
				}
				position++
				if buffer[position] != rune('r') {
					goto l283
				}
				position++
				if buffer[position] != rune('g') {
					goto l283
				}
				position++
				if buffer[position] != rune('e') {
					goto l283
				}
				position++
				{
					position285, tokenIndex285, depth285 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l285
					}
					position++
					goto l283
				l285:
					position, tokenIndex, depth = position285, tokenIndex285, depth285
				}
				{
					position286, tokenIndex286, depth286 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l286
					}
					{
						position288, tokenIndex288, depth288 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l289
						}
						goto l288
					l289:
						position, tokenIndex, depth = position288, tokenIndex288, depth288
						if !_rules[ruleRequired]() {
							goto l290
						}
						goto l288
					l290:
						position, tokenIndex, depth = position288, tokenIndex288, depth288
						if !_rules[ruleOn]() {
							goto l286
						}
					}
				l288:
					goto l287
				l286:
					position, tokenIndex, depth = position286, tokenIndex286, depth286
				}
			l287:
				depth--
				add(ruleSimpleMerge, position284)
			}
			return true
		l283:
			position, tokenIndex, depth = position283, tokenIndex283, depth283
			return false
		},
		/* 68 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position291, tokenIndex291, depth291 := position, tokenIndex, depth
			{
				position292 := position
				depth++
				if buffer[position] != rune('r') {
					goto l291
				}
				position++
				if buffer[position] != rune('e') {
					goto l291
				}
				position++
				if buffer[position] != rune('p') {
					goto l291
				}
				position++
				if buffer[position] != rune('l') {
					goto l291
				}
				position++
				if buffer[position] != rune('a') {
					goto l291
				}
				position++
				if buffer[position] != rune('c') {
					goto l291
				}
				position++
				if buffer[position] != rune('e') {
					goto l291
				}
				position++
				depth--
				add(ruleReplace, position292)
			}
			return true
		l291:
			position, tokenIndex, depth = position291, tokenIndex291, depth291
			return false
		},
		/* 69 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position293, tokenIndex293, depth293 := position, tokenIndex, depth
			{
				position294 := position
				depth++
				if buffer[position] != rune('r') {
					goto l293
				}
				position++
				if buffer[position] != rune('e') {
					goto l293
				}
				position++
				if buffer[position] != rune('q') {
					goto l293
				}
				position++
				if buffer[position] != rune('u') {
					goto l293
				}
				position++
				if buffer[position] != rune('i') {
					goto l293
				}
				position++
				if buffer[position] != rune('r') {
					goto l293
				}
				position++
				if buffer[position] != rune('e') {
					goto l293
				}
				position++
				if buffer[position] != rune('d') {
					goto l293
				}
				position++
				depth--
				add(ruleRequired, position294)
			}
			return true
		l293:
			position, tokenIndex, depth = position293, tokenIndex293, depth293
			return false
		},
		/* 70 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position295, tokenIndex295, depth295 := position, tokenIndex, depth
			{
				position296 := position
				depth++
				if buffer[position] != rune('o') {
					goto l295
				}
				position++
				if buffer[position] != rune('n') {
					goto l295
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l295
				}
				if !_rules[ruleName]() {
					goto l295
				}
				depth--
				add(ruleOn, position296)
			}
			return true
		l295:
			position, tokenIndex, depth = position295, tokenIndex295, depth295
			return false
		},
		/* 71 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position297, tokenIndex297, depth297 := position, tokenIndex, depth
			{
				position298 := position
				depth++
				if buffer[position] != rune('a') {
					goto l297
				}
				position++
				if buffer[position] != rune('u') {
					goto l297
				}
				position++
				if buffer[position] != rune('t') {
					goto l297
				}
				position++
				if buffer[position] != rune('o') {
					goto l297
				}
				position++
				depth--
				add(ruleAuto, position298)
			}
			return true
		l297:
			position, tokenIndex, depth = position297, tokenIndex297, depth297
			return false
		},
		/* 72 Default <- <Action1> */
		func() bool {
			position299, tokenIndex299, depth299 := position, tokenIndex, depth
			{
				position300 := position
				depth++
				if !_rules[ruleAction1]() {
					goto l299
				}
				depth--
				add(ruleDefault, position300)
			}
			return true
		l299:
			position, tokenIndex, depth = position299, tokenIndex299, depth299
			return false
		},
		/* 73 Sync <- <('s' 'y' 'n' 'c' '[' Level7 ((((LambdaExpr LambdaExt) / (LambdaOrExpr LambdaOrExpr)) (('|' Expression) / Default)) / (LambdaOrExpr Default Default)) ']')> */
		func() bool {
			position301, tokenIndex301, depth301 := position, tokenIndex, depth
			{
				position302 := position
				depth++
				if buffer[position] != rune('s') {
					goto l301
				}
				position++
				if buffer[position] != rune('y') {
					goto l301
				}
				position++
				if buffer[position] != rune('n') {
					goto l301
				}
				position++
				if buffer[position] != rune('c') {
					goto l301
				}
				position++
				if buffer[position] != rune('[') {
					goto l301
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l301
				}
				{
					position303, tokenIndex303, depth303 := position, tokenIndex, depth
					{
						position305, tokenIndex305, depth305 := position, tokenIndex, depth
						if !_rules[ruleLambdaExpr]() {
							goto l306
						}
						if !_rules[ruleLambdaExt]() {
							goto l306
						}
						goto l305
					l306:
						position, tokenIndex, depth = position305, tokenIndex305, depth305
						if !_rules[ruleLambdaOrExpr]() {
							goto l304
						}
						if !_rules[ruleLambdaOrExpr]() {
							goto l304
						}
					}
				l305:
					{
						position307, tokenIndex307, depth307 := position, tokenIndex, depth
						if buffer[position] != rune('|') {
							goto l308
						}
						position++
						if !_rules[ruleExpression]() {
							goto l308
						}
						goto l307
					l308:
						position, tokenIndex, depth = position307, tokenIndex307, depth307
						if !_rules[ruleDefault]() {
							goto l304
						}
					}
				l307:
					goto l303
				l304:
					position, tokenIndex, depth = position303, tokenIndex303, depth303
					if !_rules[ruleLambdaOrExpr]() {
						goto l301
					}
					if !_rules[ruleDefault]() {
						goto l301
					}
					if !_rules[ruleDefault]() {
						goto l301
					}
				}
			l303:
				if buffer[position] != rune(']') {
					goto l301
				}
				position++
				depth--
				add(ruleSync, position302)
			}
			return true
		l301:
			position, tokenIndex, depth = position301, tokenIndex301, depth301
			return false
		},
		/* 74 LambdaExt <- <(',' Expression)> */
		func() bool {
			position309, tokenIndex309, depth309 := position, tokenIndex, depth
			{
				position310 := position
				depth++
				if buffer[position] != rune(',') {
					goto l309
				}
				position++
				if !_rules[ruleExpression]() {
					goto l309
				}
				depth--
				add(ruleLambdaExt, position310)
			}
			return true
		l309:
			position, tokenIndex, depth = position309, tokenIndex309, depth309
			return false
		},
		/* 75 LambdaOrExpr <- <(LambdaExpr / ('|' Expression))> */
		func() bool {
			position311, tokenIndex311, depth311 := position, tokenIndex, depth
			{
				position312 := position
				depth++
				{
					position313, tokenIndex313, depth313 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l314
					}
					goto l313
				l314:
					position, tokenIndex, depth = position313, tokenIndex313, depth313
					if buffer[position] != rune('|') {
						goto l311
					}
					position++
					if !_rules[ruleExpression]() {
						goto l311
					}
				}
			l313:
				depth--
				add(ruleLambdaOrExpr, position312)
			}
			return true
		l311:
			position, tokenIndex, depth = position311, tokenIndex311, depth311
			return false
		},
		/* 76 Catch <- <('c' 'a' 't' 'c' 'h' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position315, tokenIndex315, depth315 := position, tokenIndex, depth
			{
				position316 := position
				depth++
				if buffer[position] != rune('c') {
					goto l315
				}
				position++
				if buffer[position] != rune('a') {
					goto l315
				}
				position++
				if buffer[position] != rune('t') {
					goto l315
				}
				position++
				if buffer[position] != rune('c') {
					goto l315
				}
				position++
				if buffer[position] != rune('h') {
					goto l315
				}
				position++
				if buffer[position] != rune('[') {
					goto l315
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l315
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l315
				}
				if buffer[position] != rune(']') {
					goto l315
				}
				position++
				depth--
				add(ruleCatch, position316)
			}
			return true
		l315:
			position, tokenIndex, depth = position315, tokenIndex315, depth315
			return false
		},
		/* 77 MapMapping <- <('m' 'a' 'p' '{' Level7 LambdaOrExpr '}')> */
		func() bool {
			position317, tokenIndex317, depth317 := position, tokenIndex, depth
			{
				position318 := position
				depth++
				if buffer[position] != rune('m') {
					goto l317
				}
				position++
				if buffer[position] != rune('a') {
					goto l317
				}
				position++
				if buffer[position] != rune('p') {
					goto l317
				}
				position++
				if buffer[position] != rune('{') {
					goto l317
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l317
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l317
				}
				if buffer[position] != rune('}') {
					goto l317
				}
				position++
				depth--
				add(ruleMapMapping, position318)
			}
			return true
		l317:
			position, tokenIndex, depth = position317, tokenIndex317, depth317
			return false
		},
		/* 78 Mapping <- <('m' 'a' 'p' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position319, tokenIndex319, depth319 := position, tokenIndex, depth
			{
				position320 := position
				depth++
				if buffer[position] != rune('m') {
					goto l319
				}
				position++
				if buffer[position] != rune('a') {
					goto l319
				}
				position++
				if buffer[position] != rune('p') {
					goto l319
				}
				position++
				if buffer[position] != rune('[') {
					goto l319
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l319
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l319
				}
				if buffer[position] != rune(']') {
					goto l319
				}
				position++
				depth--
				add(ruleMapping, position320)
			}
			return true
		l319:
			position, tokenIndex, depth = position319, tokenIndex319, depth319
			return false
		},
		/* 79 MapSelection <- <('s' 'e' 'l' 'e' 'c' 't' '{' Level7 LambdaOrExpr '}')> */
		func() bool {
			position321, tokenIndex321, depth321 := position, tokenIndex, depth
			{
				position322 := position
				depth++
				if buffer[position] != rune('s') {
					goto l321
				}
				position++
				if buffer[position] != rune('e') {
					goto l321
				}
				position++
				if buffer[position] != rune('l') {
					goto l321
				}
				position++
				if buffer[position] != rune('e') {
					goto l321
				}
				position++
				if buffer[position] != rune('c') {
					goto l321
				}
				position++
				if buffer[position] != rune('t') {
					goto l321
				}
				position++
				if buffer[position] != rune('{') {
					goto l321
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l321
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l321
				}
				if buffer[position] != rune('}') {
					goto l321
				}
				position++
				depth--
				add(ruleMapSelection, position322)
			}
			return true
		l321:
			position, tokenIndex, depth = position321, tokenIndex321, depth321
			return false
		},
		/* 80 Selection <- <('s' 'e' 'l' 'e' 'c' 't' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position323, tokenIndex323, depth323 := position, tokenIndex, depth
			{
				position324 := position
				depth++
				if buffer[position] != rune('s') {
					goto l323
				}
				position++
				if buffer[position] != rune('e') {
					goto l323
				}
				position++
				if buffer[position] != rune('l') {
					goto l323
				}
				position++
				if buffer[position] != rune('e') {
					goto l323
				}
				position++
				if buffer[position] != rune('c') {
					goto l323
				}
				position++
				if buffer[position] != rune('t') {
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
				add(ruleSelection, position324)
			}
			return true
		l323:
			position, tokenIndex, depth = position323, tokenIndex323, depth323
			return false
		},
		/* 81 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 LambdaOrExpr ']')> */
		func() bool {
			position325, tokenIndex325, depth325 := position, tokenIndex, depth
			{
				position326 := position
				depth++
				if buffer[position] != rune('s') {
					goto l325
				}
				position++
				if buffer[position] != rune('u') {
					goto l325
				}
				position++
				if buffer[position] != rune('m') {
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
				if buffer[position] != rune('|') {
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
				add(ruleSum, position326)
			}
			return true
		l325:
			position, tokenIndex, depth = position325, tokenIndex325, depth325
			return false
		},
		/* 82 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position327, tokenIndex327, depth327 := position, tokenIndex, depth
			{
				position328 := position
				depth++
				if buffer[position] != rune('l') {
					goto l327
				}
				position++
				if buffer[position] != rune('a') {
					goto l327
				}
				position++
				if buffer[position] != rune('m') {
					goto l327
				}
				position++
				if buffer[position] != rune('b') {
					goto l327
				}
				position++
				if buffer[position] != rune('d') {
					goto l327
				}
				position++
				if buffer[position] != rune('a') {
					goto l327
				}
				position++
				{
					position329, tokenIndex329, depth329 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l330
					}
					goto l329
				l330:
					position, tokenIndex, depth = position329, tokenIndex329, depth329
					if !_rules[ruleLambdaExpr]() {
						goto l327
					}
				}
			l329:
				depth--
				add(ruleLambda, position328)
			}
			return true
		l327:
			position, tokenIndex, depth = position327, tokenIndex327, depth327
			return false
		},
		/* 83 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position331, tokenIndex331, depth331 := position, tokenIndex, depth
			{
				position332 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l331
				}
				if !_rules[ruleExpression]() {
					goto l331
				}
				depth--
				add(ruleLambdaRef, position332)
			}
			return true
		l331:
			position, tokenIndex, depth = position331, tokenIndex331, depth331
			return false
		},
		/* 84 LambdaExpr <- <(ws Params ws ('-' '>') Expression)> */
		func() bool {
			position333, tokenIndex333, depth333 := position, tokenIndex, depth
			{
				position334 := position
				depth++
				if !_rules[rulews]() {
					goto l333
				}
				if !_rules[ruleParams]() {
					goto l333
				}
				if !_rules[rulews]() {
					goto l333
				}
				if buffer[position] != rune('-') {
					goto l333
				}
				position++
				if buffer[position] != rune('>') {
					goto l333
				}
				position++
				if !_rules[ruleExpression]() {
					goto l333
				}
				depth--
				add(ruleLambdaExpr, position334)
			}
			return true
		l333:
			position, tokenIndex, depth = position333, tokenIndex333, depth333
			return false
		},
		/* 85 Params <- <('|' StartParams ws Names? '|')> */
		func() bool {
			position335, tokenIndex335, depth335 := position, tokenIndex, depth
			{
				position336 := position
				depth++
				if buffer[position] != rune('|') {
					goto l335
				}
				position++
				if !_rules[ruleStartParams]() {
					goto l335
				}
				if !_rules[rulews]() {
					goto l335
				}
				{
					position337, tokenIndex337, depth337 := position, tokenIndex, depth
					if !_rules[ruleNames]() {
						goto l337
					}
					goto l338
				l337:
					position, tokenIndex, depth = position337, tokenIndex337, depth337
				}
			l338:
				if buffer[position] != rune('|') {
					goto l335
				}
				position++
				depth--
				add(ruleParams, position336)
			}
			return true
		l335:
			position, tokenIndex, depth = position335, tokenIndex335, depth335
			return false
		},
		/* 86 StartParams <- <Action2> */
		func() bool {
			position339, tokenIndex339, depth339 := position, tokenIndex, depth
			{
				position340 := position
				depth++
				if !_rules[ruleAction2]() {
					goto l339
				}
				depth--
				add(ruleStartParams, position340)
			}
			return true
		l339:
			position, tokenIndex, depth = position339, tokenIndex339, depth339
			return false
		},
		/* 87 Names <- <(NextName (',' NextName)* DefaultValue? (',' NextName DefaultValue)* VarParams?)> */
		func() bool {
			position341, tokenIndex341, depth341 := position, tokenIndex, depth
			{
				position342 := position
				depth++
				if !_rules[ruleNextName]() {
					goto l341
				}
			l343:
				{
					position344, tokenIndex344, depth344 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l344
					}
					position++
					if !_rules[ruleNextName]() {
						goto l344
					}
					goto l343
				l344:
					position, tokenIndex, depth = position344, tokenIndex344, depth344
				}
				{
					position345, tokenIndex345, depth345 := position, tokenIndex, depth
					if !_rules[ruleDefaultValue]() {
						goto l345
					}
					goto l346
				l345:
					position, tokenIndex, depth = position345, tokenIndex345, depth345
				}
			l346:
			l347:
				{
					position348, tokenIndex348, depth348 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l348
					}
					position++
					if !_rules[ruleNextName]() {
						goto l348
					}
					if !_rules[ruleDefaultValue]() {
						goto l348
					}
					goto l347
				l348:
					position, tokenIndex, depth = position348, tokenIndex348, depth348
				}
				{
					position349, tokenIndex349, depth349 := position, tokenIndex, depth
					if !_rules[ruleVarParams]() {
						goto l349
					}
					goto l350
				l349:
					position, tokenIndex, depth = position349, tokenIndex349, depth349
				}
			l350:
				depth--
				add(ruleNames, position342)
			}
			return true
		l341:
			position, tokenIndex, depth = position341, tokenIndex341, depth341
			return false
		},
		/* 88 NextName <- <(ws Name ws)> */
		func() bool {
			position351, tokenIndex351, depth351 := position, tokenIndex, depth
			{
				position352 := position
				depth++
				if !_rules[rulews]() {
					goto l351
				}
				if !_rules[ruleName]() {
					goto l351
				}
				if !_rules[rulews]() {
					goto l351
				}
				depth--
				add(ruleNextName, position352)
			}
			return true
		l351:
			position, tokenIndex, depth = position351, tokenIndex351, depth351
			return false
		},
		/* 89 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position353, tokenIndex353, depth353 := position, tokenIndex, depth
			{
				position354 := position
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
						goto l353
					}
					position++
				}
			l357:
			l355:
				{
					position356, tokenIndex356, depth356 := position, tokenIndex, depth
					{
						position361, tokenIndex361, depth361 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l362
						}
						position++
						goto l361
					l362:
						position, tokenIndex, depth = position361, tokenIndex361, depth361
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l363
						}
						position++
						goto l361
					l363:
						position, tokenIndex, depth = position361, tokenIndex361, depth361
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l364
						}
						position++
						goto l361
					l364:
						position, tokenIndex, depth = position361, tokenIndex361, depth361
						if buffer[position] != rune('_') {
							goto l356
						}
						position++
					}
				l361:
					goto l355
				l356:
					position, tokenIndex, depth = position356, tokenIndex356, depth356
				}
				depth--
				add(ruleName, position354)
			}
			return true
		l353:
			position, tokenIndex, depth = position353, tokenIndex353, depth353
			return false
		},
		/* 90 DefaultValue <- <('=' Expression)> */
		func() bool {
			position365, tokenIndex365, depth365 := position, tokenIndex, depth
			{
				position366 := position
				depth++
				if buffer[position] != rune('=') {
					goto l365
				}
				position++
				if !_rules[ruleExpression]() {
					goto l365
				}
				depth--
				add(ruleDefaultValue, position366)
			}
			return true
		l365:
			position, tokenIndex, depth = position365, tokenIndex365, depth365
			return false
		},
		/* 91 VarParams <- <('.' '.' '.' ws)> */
		func() bool {
			position367, tokenIndex367, depth367 := position, tokenIndex, depth
			{
				position368 := position
				depth++
				if buffer[position] != rune('.') {
					goto l367
				}
				position++
				if buffer[position] != rune('.') {
					goto l367
				}
				position++
				if buffer[position] != rune('.') {
					goto l367
				}
				position++
				if !_rules[rulews]() {
					goto l367
				}
				depth--
				add(ruleVarParams, position368)
			}
			return true
		l367:
			position, tokenIndex, depth = position367, tokenIndex367, depth367
			return false
		},
		/* 92 Reference <- <(((TagPrefix ('.' / Key)) / ('.'? Key)) FollowUpRef)> */
		func() bool {
			position369, tokenIndex369, depth369 := position, tokenIndex, depth
			{
				position370 := position
				depth++
				{
					position371, tokenIndex371, depth371 := position, tokenIndex, depth
					if !_rules[ruleTagPrefix]() {
						goto l372
					}
					{
						position373, tokenIndex373, depth373 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l374
						}
						position++
						goto l373
					l374:
						position, tokenIndex, depth = position373, tokenIndex373, depth373
						if !_rules[ruleKey]() {
							goto l372
						}
					}
				l373:
					goto l371
				l372:
					position, tokenIndex, depth = position371, tokenIndex371, depth371
					{
						position375, tokenIndex375, depth375 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l375
						}
						position++
						goto l376
					l375:
						position, tokenIndex, depth = position375, tokenIndex375, depth375
					}
				l376:
					if !_rules[ruleKey]() {
						goto l369
					}
				}
			l371:
				if !_rules[ruleFollowUpRef]() {
					goto l369
				}
				depth--
				add(ruleReference, position370)
			}
			return true
		l369:
			position, tokenIndex, depth = position369, tokenIndex369, depth369
			return false
		},
		/* 93 TagPrefix <- <((('d' 'o' 'c' ('.' / ':') '-'? [0-9]+) / Tag) (':' ':'))> */
		func() bool {
			position377, tokenIndex377, depth377 := position, tokenIndex, depth
			{
				position378 := position
				depth++
				{
					position379, tokenIndex379, depth379 := position, tokenIndex, depth
					if buffer[position] != rune('d') {
						goto l380
					}
					position++
					if buffer[position] != rune('o') {
						goto l380
					}
					position++
					if buffer[position] != rune('c') {
						goto l380
					}
					position++
					{
						position381, tokenIndex381, depth381 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l382
						}
						position++
						goto l381
					l382:
						position, tokenIndex, depth = position381, tokenIndex381, depth381
						if buffer[position] != rune(':') {
							goto l380
						}
						position++
					}
				l381:
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
						goto l380
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
					goto l379
				l380:
					position, tokenIndex, depth = position379, tokenIndex379, depth379
					if !_rules[ruleTag]() {
						goto l377
					}
				}
			l379:
				if buffer[position] != rune(':') {
					goto l377
				}
				position++
				if buffer[position] != rune(':') {
					goto l377
				}
				position++
				depth--
				add(ruleTagPrefix, position378)
			}
			return true
		l377:
			position, tokenIndex, depth = position377, tokenIndex377, depth377
			return false
		},
		/* 94 Tag <- <(TagComponent (('.' / ':') TagComponent)*)> */
		func() bool {
			position387, tokenIndex387, depth387 := position, tokenIndex, depth
			{
				position388 := position
				depth++
				if !_rules[ruleTagComponent]() {
					goto l387
				}
			l389:
				{
					position390, tokenIndex390, depth390 := position, tokenIndex, depth
					{
						position391, tokenIndex391, depth391 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l392
						}
						position++
						goto l391
					l392:
						position, tokenIndex, depth = position391, tokenIndex391, depth391
						if buffer[position] != rune(':') {
							goto l390
						}
						position++
					}
				l391:
					if !_rules[ruleTagComponent]() {
						goto l390
					}
					goto l389
				l390:
					position, tokenIndex, depth = position390, tokenIndex390, depth390
				}
				depth--
				add(ruleTag, position388)
			}
			return true
		l387:
			position, tokenIndex, depth = position387, tokenIndex387, depth387
			return false
		},
		/* 95 TagComponent <- <(([a-z] / [A-Z] / '_') ([a-z] / [A-Z] / [0-9] / '_')*)> */
		func() bool {
			position393, tokenIndex393, depth393 := position, tokenIndex, depth
			{
				position394 := position
				depth++
				{
					position395, tokenIndex395, depth395 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l396
					}
					position++
					goto l395
				l396:
					position, tokenIndex, depth = position395, tokenIndex395, depth395
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l397
					}
					position++
					goto l395
				l397:
					position, tokenIndex, depth = position395, tokenIndex395, depth395
					if buffer[position] != rune('_') {
						goto l393
					}
					position++
				}
			l395:
			l398:
				{
					position399, tokenIndex399, depth399 := position, tokenIndex, depth
					{
						position400, tokenIndex400, depth400 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l401
						}
						position++
						goto l400
					l401:
						position, tokenIndex, depth = position400, tokenIndex400, depth400
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l402
						}
						position++
						goto l400
					l402:
						position, tokenIndex, depth = position400, tokenIndex400, depth400
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l403
						}
						position++
						goto l400
					l403:
						position, tokenIndex, depth = position400, tokenIndex400, depth400
						if buffer[position] != rune('_') {
							goto l399
						}
						position++
					}
				l400:
					goto l398
				l399:
					position, tokenIndex, depth = position399, tokenIndex399, depth399
				}
				depth--
				add(ruleTagComponent, position394)
			}
			return true
		l393:
			position, tokenIndex, depth = position393, tokenIndex393, depth393
			return false
		},
		/* 96 FollowUpRef <- <PathComponent*> */
		func() bool {
			{
				position405 := position
				depth++
			l406:
				{
					position407, tokenIndex407, depth407 := position, tokenIndex, depth
					if !_rules[rulePathComponent]() {
						goto l407
					}
					goto l406
				l407:
					position, tokenIndex, depth = position407, tokenIndex407, depth407
				}
				depth--
				add(ruleFollowUpRef, position405)
			}
			return true
		},
		/* 97 PathComponent <- <(('.' Key) / ('.'? Index))> */
		func() bool {
			position408, tokenIndex408, depth408 := position, tokenIndex, depth
			{
				position409 := position
				depth++
				{
					position410, tokenIndex410, depth410 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l411
					}
					position++
					if !_rules[ruleKey]() {
						goto l411
					}
					goto l410
				l411:
					position, tokenIndex, depth = position410, tokenIndex410, depth410
					{
						position412, tokenIndex412, depth412 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l412
						}
						position++
						goto l413
					l412:
						position, tokenIndex, depth = position412, tokenIndex412, depth412
					}
				l413:
					if !_rules[ruleIndex]() {
						goto l408
					}
				}
			l410:
				depth--
				add(rulePathComponent, position409)
			}
			return true
		l408:
			position, tokenIndex, depth = position408, tokenIndex408, depth408
			return false
		},
		/* 98 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position414, tokenIndex414, depth414 := position, tokenIndex, depth
			{
				position415 := position
				depth++
				{
					position416, tokenIndex416, depth416 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l417
					}
					position++
					goto l416
				l417:
					position, tokenIndex, depth = position416, tokenIndex416, depth416
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l418
					}
					position++
					goto l416
				l418:
					position, tokenIndex, depth = position416, tokenIndex416, depth416
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l419
					}
					position++
					goto l416
				l419:
					position, tokenIndex, depth = position416, tokenIndex416, depth416
					if buffer[position] != rune('_') {
						goto l414
					}
					position++
				}
			l416:
			l420:
				{
					position421, tokenIndex421, depth421 := position, tokenIndex, depth
					{
						position422, tokenIndex422, depth422 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l423
						}
						position++
						goto l422
					l423:
						position, tokenIndex, depth = position422, tokenIndex422, depth422
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l424
						}
						position++
						goto l422
					l424:
						position, tokenIndex, depth = position422, tokenIndex422, depth422
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l425
						}
						position++
						goto l422
					l425:
						position, tokenIndex, depth = position422, tokenIndex422, depth422
						if buffer[position] != rune('_') {
							goto l426
						}
						position++
						goto l422
					l426:
						position, tokenIndex, depth = position422, tokenIndex422, depth422
						if buffer[position] != rune('-') {
							goto l421
						}
						position++
					}
				l422:
					goto l420
				l421:
					position, tokenIndex, depth = position421, tokenIndex421, depth421
				}
				{
					position427, tokenIndex427, depth427 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l427
					}
					position++
					{
						position429, tokenIndex429, depth429 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l430
						}
						position++
						goto l429
					l430:
						position, tokenIndex, depth = position429, tokenIndex429, depth429
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l431
						}
						position++
						goto l429
					l431:
						position, tokenIndex, depth = position429, tokenIndex429, depth429
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l432
						}
						position++
						goto l429
					l432:
						position, tokenIndex, depth = position429, tokenIndex429, depth429
						if buffer[position] != rune('_') {
							goto l427
						}
						position++
					}
				l429:
				l433:
					{
						position434, tokenIndex434, depth434 := position, tokenIndex, depth
						{
							position435, tokenIndex435, depth435 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l436
							}
							position++
							goto l435
						l436:
							position, tokenIndex, depth = position435, tokenIndex435, depth435
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l437
							}
							position++
							goto l435
						l437:
							position, tokenIndex, depth = position435, tokenIndex435, depth435
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l438
							}
							position++
							goto l435
						l438:
							position, tokenIndex, depth = position435, tokenIndex435, depth435
							if buffer[position] != rune('_') {
								goto l439
							}
							position++
							goto l435
						l439:
							position, tokenIndex, depth = position435, tokenIndex435, depth435
							if buffer[position] != rune('-') {
								goto l434
							}
							position++
						}
					l435:
						goto l433
					l434:
						position, tokenIndex, depth = position434, tokenIndex434, depth434
					}
					goto l428
				l427:
					position, tokenIndex, depth = position427, tokenIndex427, depth427
				}
			l428:
				depth--
				add(ruleKey, position415)
			}
			return true
		l414:
			position, tokenIndex, depth = position414, tokenIndex414, depth414
			return false
		},
		/* 99 Index <- <('[' '-'? [0-9]+ ']')> */
		func() bool {
			position440, tokenIndex440, depth440 := position, tokenIndex, depth
			{
				position441 := position
				depth++
				if buffer[position] != rune('[') {
					goto l440
				}
				position++
				{
					position442, tokenIndex442, depth442 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l442
					}
					position++
					goto l443
				l442:
					position, tokenIndex, depth = position442, tokenIndex442, depth442
				}
			l443:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l440
				}
				position++
			l444:
				{
					position445, tokenIndex445, depth445 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l445
					}
					position++
					goto l444
				l445:
					position, tokenIndex, depth = position445, tokenIndex445, depth445
				}
				if buffer[position] != rune(']') {
					goto l440
				}
				position++
				depth--
				add(ruleIndex, position441)
			}
			return true
		l440:
			position, tokenIndex, depth = position440, tokenIndex440, depth440
			return false
		},
		/* 100 IP <- <([0-9]+ '.' [0-9]+ '.' [0-9]+ '.' [0-9]+)> */
		func() bool {
			position446, tokenIndex446, depth446 := position, tokenIndex, depth
			{
				position447 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l446
				}
				position++
			l448:
				{
					position449, tokenIndex449, depth449 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l449
					}
					position++
					goto l448
				l449:
					position, tokenIndex, depth = position449, tokenIndex449, depth449
				}
				if buffer[position] != rune('.') {
					goto l446
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l446
				}
				position++
			l450:
				{
					position451, tokenIndex451, depth451 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l451
					}
					position++
					goto l450
				l451:
					position, tokenIndex, depth = position451, tokenIndex451, depth451
				}
				if buffer[position] != rune('.') {
					goto l446
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l446
				}
				position++
			l452:
				{
					position453, tokenIndex453, depth453 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l453
					}
					position++
					goto l452
				l453:
					position, tokenIndex, depth = position453, tokenIndex453, depth453
				}
				if buffer[position] != rune('.') {
					goto l446
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l446
				}
				position++
			l454:
				{
					position455, tokenIndex455, depth455 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l455
					}
					position++
					goto l454
				l455:
					position, tokenIndex, depth = position455, tokenIndex455, depth455
				}
				depth--
				add(ruleIP, position447)
			}
			return true
		l446:
			position, tokenIndex, depth = position446, tokenIndex446, depth446
			return false
		},
		/* 101 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position457 := position
				depth++
			l458:
				{
					position459, tokenIndex459, depth459 := position, tokenIndex, depth
					{
						position460, tokenIndex460, depth460 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l461
						}
						position++
						goto l460
					l461:
						position, tokenIndex, depth = position460, tokenIndex460, depth460
						if buffer[position] != rune('\t') {
							goto l462
						}
						position++
						goto l460
					l462:
						position, tokenIndex, depth = position460, tokenIndex460, depth460
						if buffer[position] != rune('\n') {
							goto l463
						}
						position++
						goto l460
					l463:
						position, tokenIndex, depth = position460, tokenIndex460, depth460
						if buffer[position] != rune('\r') {
							goto l459
						}
						position++
					}
				l460:
					goto l458
				l459:
					position, tokenIndex, depth = position459, tokenIndex459, depth459
				}
				depth--
				add(rulews, position457)
			}
			return true
		},
		/* 102 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position464, tokenIndex464, depth464 := position, tokenIndex, depth
			{
				position465 := position
				depth++
				{
					position468, tokenIndex468, depth468 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l469
					}
					position++
					goto l468
				l469:
					position, tokenIndex, depth = position468, tokenIndex468, depth468
					if buffer[position] != rune('\t') {
						goto l470
					}
					position++
					goto l468
				l470:
					position, tokenIndex, depth = position468, tokenIndex468, depth468
					if buffer[position] != rune('\n') {
						goto l471
					}
					position++
					goto l468
				l471:
					position, tokenIndex, depth = position468, tokenIndex468, depth468
					if buffer[position] != rune('\r') {
						goto l464
					}
					position++
				}
			l468:
			l466:
				{
					position467, tokenIndex467, depth467 := position, tokenIndex, depth
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
							goto l467
						}
						position++
					}
				l472:
					goto l466
				l467:
					position, tokenIndex, depth = position467, tokenIndex467, depth467
				}
				depth--
				add(rulereq_ws, position465)
			}
			return true
		l464:
			position, tokenIndex, depth = position464, tokenIndex464, depth464
			return false
		},
		/* 104 Action0 <- <{}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 105 Action1 <- <{}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 106 Action2 <- <{}> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
	}
	p.rules = _rules
}
