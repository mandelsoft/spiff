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
	ruleTag
	ruleTagName
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
	"Tag",
	"TagName",
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
	rules  [106]func() bool
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
		/* 5 TagMarker <- <('t' 'a' 'g' ':' TagName)> */
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
				if !_rules[ruleTagName]() {
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
			position27, tokenIndex27, depth27 := position, tokenIndex, depth
			{
				position28 := position
				depth++
				if !_rules[ruleGrouped]() {
					goto l27
				}
				depth--
				add(ruleMarkerExpression, position28)
			}
			return true
		l27:
			position, tokenIndex, depth = position27, tokenIndex27, depth27
			return false
		},
		/* 7 Expression <- <((Scoped / LambdaExpr / Level7) ws)> */
		func() bool {
			position29, tokenIndex29, depth29 := position, tokenIndex, depth
			{
				position30 := position
				depth++
				{
					position31, tokenIndex31, depth31 := position, tokenIndex, depth
					if !_rules[ruleScoped]() {
						goto l32
					}
					goto l31
				l32:
					position, tokenIndex, depth = position31, tokenIndex31, depth31
					if !_rules[ruleLambdaExpr]() {
						goto l33
					}
					goto l31
				l33:
					position, tokenIndex, depth = position31, tokenIndex31, depth31
					if !_rules[ruleLevel7]() {
						goto l29
					}
				}
			l31:
				if !_rules[rulews]() {
					goto l29
				}
				depth--
				add(ruleExpression, position30)
			}
			return true
		l29:
			position, tokenIndex, depth = position29, tokenIndex29, depth29
			return false
		},
		/* 8 Scoped <- <(ws Scope ws Expression)> */
		func() bool {
			position34, tokenIndex34, depth34 := position, tokenIndex, depth
			{
				position35 := position
				depth++
				if !_rules[rulews]() {
					goto l34
				}
				if !_rules[ruleScope]() {
					goto l34
				}
				if !_rules[rulews]() {
					goto l34
				}
				if !_rules[ruleExpression]() {
					goto l34
				}
				depth--
				add(ruleScoped, position35)
			}
			return true
		l34:
			position, tokenIndex, depth = position34, tokenIndex34, depth34
			return false
		},
		/* 9 Scope <- <(CreateScope ws Assignments? ')')> */
		func() bool {
			position36, tokenIndex36, depth36 := position, tokenIndex, depth
			{
				position37 := position
				depth++
				if !_rules[ruleCreateScope]() {
					goto l36
				}
				if !_rules[rulews]() {
					goto l36
				}
				{
					position38, tokenIndex38, depth38 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l38
					}
					goto l39
				l38:
					position, tokenIndex, depth = position38, tokenIndex38, depth38
				}
			l39:
				if buffer[position] != rune(')') {
					goto l36
				}
				position++
				depth--
				add(ruleScope, position37)
			}
			return true
		l36:
			position, tokenIndex, depth = position36, tokenIndex36, depth36
			return false
		},
		/* 10 CreateScope <- <'('> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				if buffer[position] != rune('(') {
					goto l40
				}
				position++
				depth--
				add(ruleCreateScope, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 11 Level7 <- <(ws Level6 (req_ws Or)*)> */
		func() bool {
			position42, tokenIndex42, depth42 := position, tokenIndex, depth
			{
				position43 := position
				depth++
				if !_rules[rulews]() {
					goto l42
				}
				if !_rules[ruleLevel6]() {
					goto l42
				}
			l44:
				{
					position45, tokenIndex45, depth45 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l45
					}
					if !_rules[ruleOr]() {
						goto l45
					}
					goto l44
				l45:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
				}
				depth--
				add(ruleLevel7, position43)
			}
			return true
		l42:
			position, tokenIndex, depth = position42, tokenIndex42, depth42
			return false
		},
		/* 12 Or <- <(OrOp req_ws Level6)> */
		func() bool {
			position46, tokenIndex46, depth46 := position, tokenIndex, depth
			{
				position47 := position
				depth++
				if !_rules[ruleOrOp]() {
					goto l46
				}
				if !_rules[rulereq_ws]() {
					goto l46
				}
				if !_rules[ruleLevel6]() {
					goto l46
				}
				depth--
				add(ruleOr, position47)
			}
			return true
		l46:
			position, tokenIndex, depth = position46, tokenIndex46, depth46
			return false
		},
		/* 13 OrOp <- <(('|' '|') / ('/' '/'))> */
		func() bool {
			position48, tokenIndex48, depth48 := position, tokenIndex, depth
			{
				position49 := position
				depth++
				{
					position50, tokenIndex50, depth50 := position, tokenIndex, depth
					if buffer[position] != rune('|') {
						goto l51
					}
					position++
					if buffer[position] != rune('|') {
						goto l51
					}
					position++
					goto l50
				l51:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if buffer[position] != rune('/') {
						goto l48
					}
					position++
					if buffer[position] != rune('/') {
						goto l48
					}
					position++
				}
			l50:
				depth--
				add(ruleOrOp, position49)
			}
			return true
		l48:
			position, tokenIndex, depth = position48, tokenIndex48, depth48
			return false
		},
		/* 14 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position52, tokenIndex52, depth52 := position, tokenIndex, depth
			{
				position53 := position
				depth++
				{
					position54, tokenIndex54, depth54 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l55
					}
					goto l54
				l55:
					position, tokenIndex, depth = position54, tokenIndex54, depth54
					if !_rules[ruleLevel5]() {
						goto l52
					}
				}
			l54:
				depth--
				add(ruleLevel6, position53)
			}
			return true
		l52:
			position, tokenIndex, depth = position52, tokenIndex52, depth52
			return false
		},
		/* 15 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position56, tokenIndex56, depth56 := position, tokenIndex, depth
			{
				position57 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l56
				}
				if !_rules[rulews]() {
					goto l56
				}
				if buffer[position] != rune('?') {
					goto l56
				}
				position++
				if !_rules[ruleExpression]() {
					goto l56
				}
				if buffer[position] != rune(':') {
					goto l56
				}
				position++
				if !_rules[ruleExpression]() {
					goto l56
				}
				depth--
				add(ruleConditional, position57)
			}
			return true
		l56:
			position, tokenIndex, depth = position56, tokenIndex56, depth56
			return false
		},
		/* 16 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position58, tokenIndex58, depth58 := position, tokenIndex, depth
			{
				position59 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l58
				}
			l60:
				{
					position61, tokenIndex61, depth61 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l61
					}
					goto l60
				l61:
					position, tokenIndex, depth = position61, tokenIndex61, depth61
				}
				depth--
				add(ruleLevel5, position59)
			}
			return true
		l58:
			position, tokenIndex, depth = position58, tokenIndex58, depth58
			return false
		},
		/* 17 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position62, tokenIndex62, depth62 := position, tokenIndex, depth
			{
				position63 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l62
				}
				if !_rules[ruleLevel4]() {
					goto l62
				}
				depth--
				add(ruleConcatenation, position63)
			}
			return true
		l62:
			position, tokenIndex, depth = position62, tokenIndex62, depth62
			return false
		},
		/* 18 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position64, tokenIndex64, depth64 := position, tokenIndex, depth
			{
				position65 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l64
				}
			l66:
				{
					position67, tokenIndex67, depth67 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l67
					}
					{
						position68, tokenIndex68, depth68 := position, tokenIndex, depth
						if !_rules[ruleLogOr]() {
							goto l69
						}
						goto l68
					l69:
						position, tokenIndex, depth = position68, tokenIndex68, depth68
						if !_rules[ruleLogAnd]() {
							goto l67
						}
					}
				l68:
					goto l66
				l67:
					position, tokenIndex, depth = position67, tokenIndex67, depth67
				}
				depth--
				add(ruleLevel4, position65)
			}
			return true
		l64:
			position, tokenIndex, depth = position64, tokenIndex64, depth64
			return false
		},
		/* 19 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position70, tokenIndex70, depth70 := position, tokenIndex, depth
			{
				position71 := position
				depth++
				if buffer[position] != rune('-') {
					goto l70
				}
				position++
				if buffer[position] != rune('o') {
					goto l70
				}
				position++
				if buffer[position] != rune('r') {
					goto l70
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l70
				}
				if !_rules[ruleLevel3]() {
					goto l70
				}
				depth--
				add(ruleLogOr, position71)
			}
			return true
		l70:
			position, tokenIndex, depth = position70, tokenIndex70, depth70
			return false
		},
		/* 20 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position72, tokenIndex72, depth72 := position, tokenIndex, depth
			{
				position73 := position
				depth++
				if buffer[position] != rune('-') {
					goto l72
				}
				position++
				if buffer[position] != rune('a') {
					goto l72
				}
				position++
				if buffer[position] != rune('n') {
					goto l72
				}
				position++
				if buffer[position] != rune('d') {
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
				add(ruleLogAnd, position73)
			}
			return true
		l72:
			position, tokenIndex, depth = position72, tokenIndex72, depth72
			return false
		},
		/* 21 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position74, tokenIndex74, depth74 := position, tokenIndex, depth
			{
				position75 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l74
				}
			l76:
				{
					position77, tokenIndex77, depth77 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l77
					}
					if !_rules[ruleComparison]() {
						goto l77
					}
					goto l76
				l77:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
				}
				depth--
				add(ruleLevel3, position75)
			}
			return true
		l74:
			position, tokenIndex, depth = position74, tokenIndex74, depth74
			return false
		},
		/* 22 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position78, tokenIndex78, depth78 := position, tokenIndex, depth
			{
				position79 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l78
				}
				if !_rules[rulereq_ws]() {
					goto l78
				}
				if !_rules[ruleLevel2]() {
					goto l78
				}
				depth--
				add(ruleComparison, position79)
			}
			return true
		l78:
			position, tokenIndex, depth = position78, tokenIndex78, depth78
			return false
		},
		/* 23 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position80, tokenIndex80, depth80 := position, tokenIndex, depth
			{
				position81 := position
				depth++
				{
					position82, tokenIndex82, depth82 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l83
					}
					position++
					if buffer[position] != rune('=') {
						goto l83
					}
					position++
					goto l82
				l83:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if buffer[position] != rune('!') {
						goto l84
					}
					position++
					if buffer[position] != rune('=') {
						goto l84
					}
					position++
					goto l82
				l84:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if buffer[position] != rune('<') {
						goto l85
					}
					position++
					if buffer[position] != rune('=') {
						goto l85
					}
					position++
					goto l82
				l85:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if buffer[position] != rune('>') {
						goto l86
					}
					position++
					if buffer[position] != rune('=') {
						goto l86
					}
					position++
					goto l82
				l86:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if buffer[position] != rune('>') {
						goto l87
					}
					position++
					goto l82
				l87:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if buffer[position] != rune('<') {
						goto l88
					}
					position++
					goto l82
				l88:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if buffer[position] != rune('>') {
						goto l80
					}
					position++
				}
			l82:
				depth--
				add(ruleCompareOp, position81)
			}
			return true
		l80:
			position, tokenIndex, depth = position80, tokenIndex80, depth80
			return false
		},
		/* 24 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position89, tokenIndex89, depth89 := position, tokenIndex, depth
			{
				position90 := position
				depth++
				if !_rules[ruleLevel1]() {
					goto l89
				}
			l91:
				{
					position92, tokenIndex92, depth92 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l92
					}
					{
						position93, tokenIndex93, depth93 := position, tokenIndex, depth
						if !_rules[ruleAddition]() {
							goto l94
						}
						goto l93
					l94:
						position, tokenIndex, depth = position93, tokenIndex93, depth93
						if !_rules[ruleSubtraction]() {
							goto l92
						}
					}
				l93:
					goto l91
				l92:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
				}
				depth--
				add(ruleLevel2, position90)
			}
			return true
		l89:
			position, tokenIndex, depth = position89, tokenIndex89, depth89
			return false
		},
		/* 25 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position95, tokenIndex95, depth95 := position, tokenIndex, depth
			{
				position96 := position
				depth++
				if buffer[position] != rune('+') {
					goto l95
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l95
				}
				if !_rules[ruleLevel1]() {
					goto l95
				}
				depth--
				add(ruleAddition, position96)
			}
			return true
		l95:
			position, tokenIndex, depth = position95, tokenIndex95, depth95
			return false
		},
		/* 26 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position97, tokenIndex97, depth97 := position, tokenIndex, depth
			{
				position98 := position
				depth++
				if buffer[position] != rune('-') {
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
				add(ruleSubtraction, position98)
			}
			return true
		l97:
			position, tokenIndex, depth = position97, tokenIndex97, depth97
			return false
		},
		/* 27 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position99, tokenIndex99, depth99 := position, tokenIndex, depth
			{
				position100 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l99
				}
			l101:
				{
					position102, tokenIndex102, depth102 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l102
					}
					{
						position103, tokenIndex103, depth103 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l104
						}
						goto l103
					l104:
						position, tokenIndex, depth = position103, tokenIndex103, depth103
						if !_rules[ruleDivision]() {
							goto l105
						}
						goto l103
					l105:
						position, tokenIndex, depth = position103, tokenIndex103, depth103
						if !_rules[ruleModulo]() {
							goto l102
						}
					}
				l103:
					goto l101
				l102:
					position, tokenIndex, depth = position102, tokenIndex102, depth102
				}
				depth--
				add(ruleLevel1, position100)
			}
			return true
		l99:
			position, tokenIndex, depth = position99, tokenIndex99, depth99
			return false
		},
		/* 28 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position106, tokenIndex106, depth106 := position, tokenIndex, depth
			{
				position107 := position
				depth++
				if buffer[position] != rune('*') {
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
				add(ruleMultiplication, position107)
			}
			return true
		l106:
			position, tokenIndex, depth = position106, tokenIndex106, depth106
			return false
		},
		/* 29 Division <- <('/' req_ws Level0)> */
		func() bool {
			position108, tokenIndex108, depth108 := position, tokenIndex, depth
			{
				position109 := position
				depth++
				if buffer[position] != rune('/') {
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
				add(ruleDivision, position109)
			}
			return true
		l108:
			position, tokenIndex, depth = position108, tokenIndex108, depth108
			return false
		},
		/* 30 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position110, tokenIndex110, depth110 := position, tokenIndex, depth
			{
				position111 := position
				depth++
				if buffer[position] != rune('%') {
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
				add(ruleModulo, position111)
			}
			return true
		l110:
			position, tokenIndex, depth = position110, tokenIndex110, depth110
			return false
		},
		/* 31 Level0 <- <(IP / String / Number / Boolean / Undefined / Nil / Symbol / Not / Substitution / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position112, tokenIndex112, depth112 := position, tokenIndex, depth
			{
				position113 := position
				depth++
				{
					position114, tokenIndex114, depth114 := position, tokenIndex, depth
					if !_rules[ruleIP]() {
						goto l115
					}
					goto l114
				l115:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if !_rules[ruleString]() {
						goto l116
					}
					goto l114
				l116:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if !_rules[ruleNumber]() {
						goto l117
					}
					goto l114
				l117:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if !_rules[ruleBoolean]() {
						goto l118
					}
					goto l114
				l118:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if !_rules[ruleUndefined]() {
						goto l119
					}
					goto l114
				l119:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if !_rules[ruleNil]() {
						goto l120
					}
					goto l114
				l120:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if !_rules[ruleSymbol]() {
						goto l121
					}
					goto l114
				l121:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if !_rules[ruleNot]() {
						goto l122
					}
					goto l114
				l122:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if !_rules[ruleSubstitution]() {
						goto l123
					}
					goto l114
				l123:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if !_rules[ruleMerge]() {
						goto l124
					}
					goto l114
				l124:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if !_rules[ruleAuto]() {
						goto l125
					}
					goto l114
				l125:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if !_rules[ruleLambda]() {
						goto l126
					}
					goto l114
				l126:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if !_rules[ruleChained]() {
						goto l112
					}
				}
			l114:
				depth--
				add(ruleLevel0, position113)
			}
			return true
		l112:
			position, tokenIndex, depth = position112, tokenIndex112, depth112
			return false
		},
		/* 32 Chained <- <((MapMapping / Sync / Catch / Mapping / MapSelection / Selection / Sum / List / Map / Range / Grouped / Reference) ChainedQualifiedExpression*)> */
		func() bool {
			position127, tokenIndex127, depth127 := position, tokenIndex, depth
			{
				position128 := position
				depth++
				{
					position129, tokenIndex129, depth129 := position, tokenIndex, depth
					if !_rules[ruleMapMapping]() {
						goto l130
					}
					goto l129
				l130:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
					if !_rules[ruleSync]() {
						goto l131
					}
					goto l129
				l131:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
					if !_rules[ruleCatch]() {
						goto l132
					}
					goto l129
				l132:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
					if !_rules[ruleMapping]() {
						goto l133
					}
					goto l129
				l133:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
					if !_rules[ruleMapSelection]() {
						goto l134
					}
					goto l129
				l134:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
					if !_rules[ruleSelection]() {
						goto l135
					}
					goto l129
				l135:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
					if !_rules[ruleSum]() {
						goto l136
					}
					goto l129
				l136:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
					if !_rules[ruleList]() {
						goto l137
					}
					goto l129
				l137:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
					if !_rules[ruleMap]() {
						goto l138
					}
					goto l129
				l138:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
					if !_rules[ruleRange]() {
						goto l139
					}
					goto l129
				l139:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
					if !_rules[ruleGrouped]() {
						goto l140
					}
					goto l129
				l140:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
					if !_rules[ruleReference]() {
						goto l127
					}
				}
			l129:
			l141:
				{
					position142, tokenIndex142, depth142 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l142
					}
					goto l141
				l142:
					position, tokenIndex, depth = position142, tokenIndex142, depth142
				}
				depth--
				add(ruleChained, position128)
			}
			return true
		l127:
			position, tokenIndex, depth = position127, tokenIndex127, depth127
			return false
		},
		/* 33 ChainedQualifiedExpression <- <(ChainedCall / Currying / ChainedRef / ChainedDynRef / Projection)> */
		func() bool {
			position143, tokenIndex143, depth143 := position, tokenIndex, depth
			{
				position144 := position
				depth++
				{
					position145, tokenIndex145, depth145 := position, tokenIndex, depth
					if !_rules[ruleChainedCall]() {
						goto l146
					}
					goto l145
				l146:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if !_rules[ruleCurrying]() {
						goto l147
					}
					goto l145
				l147:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if !_rules[ruleChainedRef]() {
						goto l148
					}
					goto l145
				l148:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if !_rules[ruleChainedDynRef]() {
						goto l149
					}
					goto l145
				l149:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if !_rules[ruleProjection]() {
						goto l143
					}
				}
			l145:
				depth--
				add(ruleChainedQualifiedExpression, position144)
			}
			return true
		l143:
			position, tokenIndex, depth = position143, tokenIndex143, depth143
			return false
		},
		/* 34 ChainedRef <- <(PathComponent FollowUpRef)> */
		func() bool {
			position150, tokenIndex150, depth150 := position, tokenIndex, depth
			{
				position151 := position
				depth++
				if !_rules[rulePathComponent]() {
					goto l150
				}
				if !_rules[ruleFollowUpRef]() {
					goto l150
				}
				depth--
				add(ruleChainedRef, position151)
			}
			return true
		l150:
			position, tokenIndex, depth = position150, tokenIndex150, depth150
			return false
		},
		/* 35 ChainedDynRef <- <('.'? '[' Expression ']')> */
		func() bool {
			position152, tokenIndex152, depth152 := position, tokenIndex, depth
			{
				position153 := position
				depth++
				{
					position154, tokenIndex154, depth154 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l154
					}
					position++
					goto l155
				l154:
					position, tokenIndex, depth = position154, tokenIndex154, depth154
				}
			l155:
				if buffer[position] != rune('[') {
					goto l152
				}
				position++
				if !_rules[ruleExpression]() {
					goto l152
				}
				if buffer[position] != rune(']') {
					goto l152
				}
				position++
				depth--
				add(ruleChainedDynRef, position153)
			}
			return true
		l152:
			position, tokenIndex, depth = position152, tokenIndex152, depth152
			return false
		},
		/* 36 Slice <- <Range> */
		func() bool {
			position156, tokenIndex156, depth156 := position, tokenIndex, depth
			{
				position157 := position
				depth++
				if !_rules[ruleRange]() {
					goto l156
				}
				depth--
				add(ruleSlice, position157)
			}
			return true
		l156:
			position, tokenIndex, depth = position156, tokenIndex156, depth156
			return false
		},
		/* 37 Currying <- <('*' ChainedCall)> */
		func() bool {
			position158, tokenIndex158, depth158 := position, tokenIndex, depth
			{
				position159 := position
				depth++
				if buffer[position] != rune('*') {
					goto l158
				}
				position++
				if !_rules[ruleChainedCall]() {
					goto l158
				}
				depth--
				add(ruleCurrying, position159)
			}
			return true
		l158:
			position, tokenIndex, depth = position158, tokenIndex158, depth158
			return false
		},
		/* 38 ChainedCall <- <(StartArguments NameArgumentList? ')')> */
		func() bool {
			position160, tokenIndex160, depth160 := position, tokenIndex, depth
			{
				position161 := position
				depth++
				if !_rules[ruleStartArguments]() {
					goto l160
				}
				{
					position162, tokenIndex162, depth162 := position, tokenIndex, depth
					if !_rules[ruleNameArgumentList]() {
						goto l162
					}
					goto l163
				l162:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
				}
			l163:
				if buffer[position] != rune(')') {
					goto l160
				}
				position++
				depth--
				add(ruleChainedCall, position161)
			}
			return true
		l160:
			position, tokenIndex, depth = position160, tokenIndex160, depth160
			return false
		},
		/* 39 StartArguments <- <('(' ws)> */
		func() bool {
			position164, tokenIndex164, depth164 := position, tokenIndex, depth
			{
				position165 := position
				depth++
				if buffer[position] != rune('(') {
					goto l164
				}
				position++
				if !_rules[rulews]() {
					goto l164
				}
				depth--
				add(ruleStartArguments, position165)
			}
			return true
		l164:
			position, tokenIndex, depth = position164, tokenIndex164, depth164
			return false
		},
		/* 40 NameArgumentList <- <(((NextNameArgument (',' NextNameArgument)*) / NextExpression) (',' NextExpression)*)> */
		func() bool {
			position166, tokenIndex166, depth166 := position, tokenIndex, depth
			{
				position167 := position
				depth++
				{
					position168, tokenIndex168, depth168 := position, tokenIndex, depth
					if !_rules[ruleNextNameArgument]() {
						goto l169
					}
				l170:
					{
						position171, tokenIndex171, depth171 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l171
						}
						position++
						if !_rules[ruleNextNameArgument]() {
							goto l171
						}
						goto l170
					l171:
						position, tokenIndex, depth = position171, tokenIndex171, depth171
					}
					goto l168
				l169:
					position, tokenIndex, depth = position168, tokenIndex168, depth168
					if !_rules[ruleNextExpression]() {
						goto l166
					}
				}
			l168:
			l172:
				{
					position173, tokenIndex173, depth173 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l173
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l173
					}
					goto l172
				l173:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
				}
				depth--
				add(ruleNameArgumentList, position167)
			}
			return true
		l166:
			position, tokenIndex, depth = position166, tokenIndex166, depth166
			return false
		},
		/* 41 NextNameArgument <- <(ws Name ws '=' ws Expression ws)> */
		func() bool {
			position174, tokenIndex174, depth174 := position, tokenIndex, depth
			{
				position175 := position
				depth++
				if !_rules[rulews]() {
					goto l174
				}
				if !_rules[ruleName]() {
					goto l174
				}
				if !_rules[rulews]() {
					goto l174
				}
				if buffer[position] != rune('=') {
					goto l174
				}
				position++
				if !_rules[rulews]() {
					goto l174
				}
				if !_rules[ruleExpression]() {
					goto l174
				}
				if !_rules[rulews]() {
					goto l174
				}
				depth--
				add(ruleNextNameArgument, position175)
			}
			return true
		l174:
			position, tokenIndex, depth = position174, tokenIndex174, depth174
			return false
		},
		/* 42 ExpressionList <- <(NextExpression (',' NextExpression)*)> */
		func() bool {
			position176, tokenIndex176, depth176 := position, tokenIndex, depth
			{
				position177 := position
				depth++
				if !_rules[ruleNextExpression]() {
					goto l176
				}
			l178:
				{
					position179, tokenIndex179, depth179 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l179
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l179
					}
					goto l178
				l179:
					position, tokenIndex, depth = position179, tokenIndex179, depth179
				}
				depth--
				add(ruleExpressionList, position177)
			}
			return true
		l176:
			position, tokenIndex, depth = position176, tokenIndex176, depth176
			return false
		},
		/* 43 NextExpression <- <(Expression ListExpansion?)> */
		func() bool {
			position180, tokenIndex180, depth180 := position, tokenIndex, depth
			{
				position181 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l180
				}
				{
					position182, tokenIndex182, depth182 := position, tokenIndex, depth
					if !_rules[ruleListExpansion]() {
						goto l182
					}
					goto l183
				l182:
					position, tokenIndex, depth = position182, tokenIndex182, depth182
				}
			l183:
				depth--
				add(ruleNextExpression, position181)
			}
			return true
		l180:
			position, tokenIndex, depth = position180, tokenIndex180, depth180
			return false
		},
		/* 44 ListExpansion <- <('.' '.' '.' ws)> */
		func() bool {
			position184, tokenIndex184, depth184 := position, tokenIndex, depth
			{
				position185 := position
				depth++
				if buffer[position] != rune('.') {
					goto l184
				}
				position++
				if buffer[position] != rune('.') {
					goto l184
				}
				position++
				if buffer[position] != rune('.') {
					goto l184
				}
				position++
				if !_rules[rulews]() {
					goto l184
				}
				depth--
				add(ruleListExpansion, position185)
			}
			return true
		l184:
			position, tokenIndex, depth = position184, tokenIndex184, depth184
			return false
		},
		/* 45 Projection <- <('.'? (('[' '*' ']') / Slice) ProjectionValue ChainedQualifiedExpression*)> */
		func() bool {
			position186, tokenIndex186, depth186 := position, tokenIndex, depth
			{
				position187 := position
				depth++
				{
					position188, tokenIndex188, depth188 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l188
					}
					position++
					goto l189
				l188:
					position, tokenIndex, depth = position188, tokenIndex188, depth188
				}
			l189:
				{
					position190, tokenIndex190, depth190 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l191
					}
					position++
					if buffer[position] != rune('*') {
						goto l191
					}
					position++
					if buffer[position] != rune(']') {
						goto l191
					}
					position++
					goto l190
				l191:
					position, tokenIndex, depth = position190, tokenIndex190, depth190
					if !_rules[ruleSlice]() {
						goto l186
					}
				}
			l190:
				if !_rules[ruleProjectionValue]() {
					goto l186
				}
			l192:
				{
					position193, tokenIndex193, depth193 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l193
					}
					goto l192
				l193:
					position, tokenIndex, depth = position193, tokenIndex193, depth193
				}
				depth--
				add(ruleProjection, position187)
			}
			return true
		l186:
			position, tokenIndex, depth = position186, tokenIndex186, depth186
			return false
		},
		/* 46 ProjectionValue <- <Action0> */
		func() bool {
			position194, tokenIndex194, depth194 := position, tokenIndex, depth
			{
				position195 := position
				depth++
				if !_rules[ruleAction0]() {
					goto l194
				}
				depth--
				add(ruleProjectionValue, position195)
			}
			return true
		l194:
			position, tokenIndex, depth = position194, tokenIndex194, depth194
			return false
		},
		/* 47 Substitution <- <('*' Level0)> */
		func() bool {
			position196, tokenIndex196, depth196 := position, tokenIndex, depth
			{
				position197 := position
				depth++
				if buffer[position] != rune('*') {
					goto l196
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l196
				}
				depth--
				add(ruleSubstitution, position197)
			}
			return true
		l196:
			position, tokenIndex, depth = position196, tokenIndex196, depth196
			return false
		},
		/* 48 Not <- <('!' ws Level0)> */
		func() bool {
			position198, tokenIndex198, depth198 := position, tokenIndex, depth
			{
				position199 := position
				depth++
				if buffer[position] != rune('!') {
					goto l198
				}
				position++
				if !_rules[rulews]() {
					goto l198
				}
				if !_rules[ruleLevel0]() {
					goto l198
				}
				depth--
				add(ruleNot, position199)
			}
			return true
		l198:
			position, tokenIndex, depth = position198, tokenIndex198, depth198
			return false
		},
		/* 49 Grouped <- <('(' Expression ')')> */
		func() bool {
			position200, tokenIndex200, depth200 := position, tokenIndex, depth
			{
				position201 := position
				depth++
				if buffer[position] != rune('(') {
					goto l200
				}
				position++
				if !_rules[ruleExpression]() {
					goto l200
				}
				if buffer[position] != rune(')') {
					goto l200
				}
				position++
				depth--
				add(ruleGrouped, position201)
			}
			return true
		l200:
			position, tokenIndex, depth = position200, tokenIndex200, depth200
			return false
		},
		/* 50 Range <- <(StartRange Expression? RangeOp Expression? ']')> */
		func() bool {
			position202, tokenIndex202, depth202 := position, tokenIndex, depth
			{
				position203 := position
				depth++
				if !_rules[ruleStartRange]() {
					goto l202
				}
				{
					position204, tokenIndex204, depth204 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l204
					}
					goto l205
				l204:
					position, tokenIndex, depth = position204, tokenIndex204, depth204
				}
			l205:
				if !_rules[ruleRangeOp]() {
					goto l202
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
				if buffer[position] != rune(']') {
					goto l202
				}
				position++
				depth--
				add(ruleRange, position203)
			}
			return true
		l202:
			position, tokenIndex, depth = position202, tokenIndex202, depth202
			return false
		},
		/* 51 StartRange <- <'['> */
		func() bool {
			position208, tokenIndex208, depth208 := position, tokenIndex, depth
			{
				position209 := position
				depth++
				if buffer[position] != rune('[') {
					goto l208
				}
				position++
				depth--
				add(ruleStartRange, position209)
			}
			return true
		l208:
			position, tokenIndex, depth = position208, tokenIndex208, depth208
			return false
		},
		/* 52 RangeOp <- <('.' '.')> */
		func() bool {
			position210, tokenIndex210, depth210 := position, tokenIndex, depth
			{
				position211 := position
				depth++
				if buffer[position] != rune('.') {
					goto l210
				}
				position++
				if buffer[position] != rune('.') {
					goto l210
				}
				position++
				depth--
				add(ruleRangeOp, position211)
			}
			return true
		l210:
			position, tokenIndex, depth = position210, tokenIndex210, depth210
			return false
		},
		/* 53 Number <- <('-'? [0-9] ([0-9] / '_')* ('.' [0-9] [0-9]*)? (('e' / 'E') '-'? [0-9] [0-9]*)?)> */
		func() bool {
			position212, tokenIndex212, depth212 := position, tokenIndex, depth
			{
				position213 := position
				depth++
				{
					position214, tokenIndex214, depth214 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l214
					}
					position++
					goto l215
				l214:
					position, tokenIndex, depth = position214, tokenIndex214, depth214
				}
			l215:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l212
				}
				position++
			l216:
				{
					position217, tokenIndex217, depth217 := position, tokenIndex, depth
					{
						position218, tokenIndex218, depth218 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l219
						}
						position++
						goto l218
					l219:
						position, tokenIndex, depth = position218, tokenIndex218, depth218
						if buffer[position] != rune('_') {
							goto l217
						}
						position++
					}
				l218:
					goto l216
				l217:
					position, tokenIndex, depth = position217, tokenIndex217, depth217
				}
				{
					position220, tokenIndex220, depth220 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l220
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l220
					}
					position++
				l222:
					{
						position223, tokenIndex223, depth223 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l223
						}
						position++
						goto l222
					l223:
						position, tokenIndex, depth = position223, tokenIndex223, depth223
					}
					goto l221
				l220:
					position, tokenIndex, depth = position220, tokenIndex220, depth220
				}
			l221:
				{
					position224, tokenIndex224, depth224 := position, tokenIndex, depth
					{
						position226, tokenIndex226, depth226 := position, tokenIndex, depth
						if buffer[position] != rune('e') {
							goto l227
						}
						position++
						goto l226
					l227:
						position, tokenIndex, depth = position226, tokenIndex226, depth226
						if buffer[position] != rune('E') {
							goto l224
						}
						position++
					}
				l226:
					{
						position228, tokenIndex228, depth228 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l228
						}
						position++
						goto l229
					l228:
						position, tokenIndex, depth = position228, tokenIndex228, depth228
					}
				l229:
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l224
					}
					position++
				l230:
					{
						position231, tokenIndex231, depth231 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l231
						}
						position++
						goto l230
					l231:
						position, tokenIndex, depth = position231, tokenIndex231, depth231
					}
					goto l225
				l224:
					position, tokenIndex, depth = position224, tokenIndex224, depth224
				}
			l225:
				depth--
				add(ruleNumber, position213)
			}
			return true
		l212:
			position, tokenIndex, depth = position212, tokenIndex212, depth212
			return false
		},
		/* 54 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position232, tokenIndex232, depth232 := position, tokenIndex, depth
			{
				position233 := position
				depth++
				if buffer[position] != rune('"') {
					goto l232
				}
				position++
			l234:
				{
					position235, tokenIndex235, depth235 := position, tokenIndex, depth
					{
						position236, tokenIndex236, depth236 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l237
						}
						position++
						if buffer[position] != rune('"') {
							goto l237
						}
						position++
						goto l236
					l237:
						position, tokenIndex, depth = position236, tokenIndex236, depth236
						{
							position238, tokenIndex238, depth238 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l238
							}
							position++
							goto l235
						l238:
							position, tokenIndex, depth = position238, tokenIndex238, depth238
						}
						if !matchDot() {
							goto l235
						}
					}
				l236:
					goto l234
				l235:
					position, tokenIndex, depth = position235, tokenIndex235, depth235
				}
				if buffer[position] != rune('"') {
					goto l232
				}
				position++
				depth--
				add(ruleString, position233)
			}
			return true
		l232:
			position, tokenIndex, depth = position232, tokenIndex232, depth232
			return false
		},
		/* 55 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position239, tokenIndex239, depth239 := position, tokenIndex, depth
			{
				position240 := position
				depth++
				{
					position241, tokenIndex241, depth241 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l242
					}
					position++
					if buffer[position] != rune('r') {
						goto l242
					}
					position++
					if buffer[position] != rune('u') {
						goto l242
					}
					position++
					if buffer[position] != rune('e') {
						goto l242
					}
					position++
					goto l241
				l242:
					position, tokenIndex, depth = position241, tokenIndex241, depth241
					if buffer[position] != rune('f') {
						goto l239
					}
					position++
					if buffer[position] != rune('a') {
						goto l239
					}
					position++
					if buffer[position] != rune('l') {
						goto l239
					}
					position++
					if buffer[position] != rune('s') {
						goto l239
					}
					position++
					if buffer[position] != rune('e') {
						goto l239
					}
					position++
				}
			l241:
				depth--
				add(ruleBoolean, position240)
			}
			return true
		l239:
			position, tokenIndex, depth = position239, tokenIndex239, depth239
			return false
		},
		/* 56 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position243, tokenIndex243, depth243 := position, tokenIndex, depth
			{
				position244 := position
				depth++
				{
					position245, tokenIndex245, depth245 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l246
					}
					position++
					if buffer[position] != rune('i') {
						goto l246
					}
					position++
					if buffer[position] != rune('l') {
						goto l246
					}
					position++
					goto l245
				l246:
					position, tokenIndex, depth = position245, tokenIndex245, depth245
					if buffer[position] != rune('~') {
						goto l243
					}
					position++
				}
			l245:
				depth--
				add(ruleNil, position244)
			}
			return true
		l243:
			position, tokenIndex, depth = position243, tokenIndex243, depth243
			return false
		},
		/* 57 Undefined <- <('~' '~')> */
		func() bool {
			position247, tokenIndex247, depth247 := position, tokenIndex, depth
			{
				position248 := position
				depth++
				if buffer[position] != rune('~') {
					goto l247
				}
				position++
				if buffer[position] != rune('~') {
					goto l247
				}
				position++
				depth--
				add(ruleUndefined, position248)
			}
			return true
		l247:
			position, tokenIndex, depth = position247, tokenIndex247, depth247
			return false
		},
		/* 58 Symbol <- <('$' Name)> */
		func() bool {
			position249, tokenIndex249, depth249 := position, tokenIndex, depth
			{
				position250 := position
				depth++
				if buffer[position] != rune('$') {
					goto l249
				}
				position++
				if !_rules[ruleName]() {
					goto l249
				}
				depth--
				add(ruleSymbol, position250)
			}
			return true
		l249:
			position, tokenIndex, depth = position249, tokenIndex249, depth249
			return false
		},
		/* 59 List <- <(StartList ExpressionList? ']')> */
		func() bool {
			position251, tokenIndex251, depth251 := position, tokenIndex, depth
			{
				position252 := position
				depth++
				if !_rules[ruleStartList]() {
					goto l251
				}
				{
					position253, tokenIndex253, depth253 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l253
					}
					goto l254
				l253:
					position, tokenIndex, depth = position253, tokenIndex253, depth253
				}
			l254:
				if buffer[position] != rune(']') {
					goto l251
				}
				position++
				depth--
				add(ruleList, position252)
			}
			return true
		l251:
			position, tokenIndex, depth = position251, tokenIndex251, depth251
			return false
		},
		/* 60 StartList <- <('[' ws)> */
		func() bool {
			position255, tokenIndex255, depth255 := position, tokenIndex, depth
			{
				position256 := position
				depth++
				if buffer[position] != rune('[') {
					goto l255
				}
				position++
				if !_rules[rulews]() {
					goto l255
				}
				depth--
				add(ruleStartList, position256)
			}
			return true
		l255:
			position, tokenIndex, depth = position255, tokenIndex255, depth255
			return false
		},
		/* 61 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position257, tokenIndex257, depth257 := position, tokenIndex, depth
			{
				position258 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l257
				}
				if !_rules[rulews]() {
					goto l257
				}
				{
					position259, tokenIndex259, depth259 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l259
					}
					goto l260
				l259:
					position, tokenIndex, depth = position259, tokenIndex259, depth259
				}
			l260:
				if buffer[position] != rune('}') {
					goto l257
				}
				position++
				depth--
				add(ruleMap, position258)
			}
			return true
		l257:
			position, tokenIndex, depth = position257, tokenIndex257, depth257
			return false
		},
		/* 62 CreateMap <- <'{'> */
		func() bool {
			position261, tokenIndex261, depth261 := position, tokenIndex, depth
			{
				position262 := position
				depth++
				if buffer[position] != rune('{') {
					goto l261
				}
				position++
				depth--
				add(ruleCreateMap, position262)
			}
			return true
		l261:
			position, tokenIndex, depth = position261, tokenIndex261, depth261
			return false
		},
		/* 63 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position263, tokenIndex263, depth263 := position, tokenIndex, depth
			{
				position264 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l263
				}
			l265:
				{
					position266, tokenIndex266, depth266 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l266
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l266
					}
					goto l265
				l266:
					position, tokenIndex, depth = position266, tokenIndex266, depth266
				}
				depth--
				add(ruleAssignments, position264)
			}
			return true
		l263:
			position, tokenIndex, depth = position263, tokenIndex263, depth263
			return false
		},
		/* 64 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position267, tokenIndex267, depth267 := position, tokenIndex, depth
			{
				position268 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l267
				}
				if buffer[position] != rune('=') {
					goto l267
				}
				position++
				if !_rules[ruleExpression]() {
					goto l267
				}
				depth--
				add(ruleAssignment, position268)
			}
			return true
		l267:
			position, tokenIndex, depth = position267, tokenIndex267, depth267
			return false
		},
		/* 65 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position269, tokenIndex269, depth269 := position, tokenIndex, depth
			{
				position270 := position
				depth++
				{
					position271, tokenIndex271, depth271 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l272
					}
					goto l271
				l272:
					position, tokenIndex, depth = position271, tokenIndex271, depth271
					if !_rules[ruleSimpleMerge]() {
						goto l269
					}
				}
			l271:
				depth--
				add(ruleMerge, position270)
			}
			return true
		l269:
			position, tokenIndex, depth = position269, tokenIndex269, depth269
			return false
		},
		/* 66 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position273, tokenIndex273, depth273 := position, tokenIndex, depth
			{
				position274 := position
				depth++
				if buffer[position] != rune('m') {
					goto l273
				}
				position++
				if buffer[position] != rune('e') {
					goto l273
				}
				position++
				if buffer[position] != rune('r') {
					goto l273
				}
				position++
				if buffer[position] != rune('g') {
					goto l273
				}
				position++
				if buffer[position] != rune('e') {
					goto l273
				}
				position++
				{
					position275, tokenIndex275, depth275 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l275
					}
					if !_rules[ruleRequired]() {
						goto l275
					}
					goto l273
				l275:
					position, tokenIndex, depth = position275, tokenIndex275, depth275
				}
				{
					position276, tokenIndex276, depth276 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l276
					}
					{
						position278, tokenIndex278, depth278 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l279
						}
						goto l278
					l279:
						position, tokenIndex, depth = position278, tokenIndex278, depth278
						if !_rules[ruleOn]() {
							goto l276
						}
					}
				l278:
					goto l277
				l276:
					position, tokenIndex, depth = position276, tokenIndex276, depth276
				}
			l277:
				if !_rules[rulereq_ws]() {
					goto l273
				}
				if !_rules[ruleReference]() {
					goto l273
				}
				depth--
				add(ruleRefMerge, position274)
			}
			return true
		l273:
			position, tokenIndex, depth = position273, tokenIndex273, depth273
			return false
		},
		/* 67 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' !'(' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position280, tokenIndex280, depth280 := position, tokenIndex, depth
			{
				position281 := position
				depth++
				if buffer[position] != rune('m') {
					goto l280
				}
				position++
				if buffer[position] != rune('e') {
					goto l280
				}
				position++
				if buffer[position] != rune('r') {
					goto l280
				}
				position++
				if buffer[position] != rune('g') {
					goto l280
				}
				position++
				if buffer[position] != rune('e') {
					goto l280
				}
				position++
				{
					position282, tokenIndex282, depth282 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l282
					}
					position++
					goto l280
				l282:
					position, tokenIndex, depth = position282, tokenIndex282, depth282
				}
				{
					position283, tokenIndex283, depth283 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l283
					}
					{
						position285, tokenIndex285, depth285 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l286
						}
						goto l285
					l286:
						position, tokenIndex, depth = position285, tokenIndex285, depth285
						if !_rules[ruleRequired]() {
							goto l287
						}
						goto l285
					l287:
						position, tokenIndex, depth = position285, tokenIndex285, depth285
						if !_rules[ruleOn]() {
							goto l283
						}
					}
				l285:
					goto l284
				l283:
					position, tokenIndex, depth = position283, tokenIndex283, depth283
				}
			l284:
				depth--
				add(ruleSimpleMerge, position281)
			}
			return true
		l280:
			position, tokenIndex, depth = position280, tokenIndex280, depth280
			return false
		},
		/* 68 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position288, tokenIndex288, depth288 := position, tokenIndex, depth
			{
				position289 := position
				depth++
				if buffer[position] != rune('r') {
					goto l288
				}
				position++
				if buffer[position] != rune('e') {
					goto l288
				}
				position++
				if buffer[position] != rune('p') {
					goto l288
				}
				position++
				if buffer[position] != rune('l') {
					goto l288
				}
				position++
				if buffer[position] != rune('a') {
					goto l288
				}
				position++
				if buffer[position] != rune('c') {
					goto l288
				}
				position++
				if buffer[position] != rune('e') {
					goto l288
				}
				position++
				depth--
				add(ruleReplace, position289)
			}
			return true
		l288:
			position, tokenIndex, depth = position288, tokenIndex288, depth288
			return false
		},
		/* 69 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position290, tokenIndex290, depth290 := position, tokenIndex, depth
			{
				position291 := position
				depth++
				if buffer[position] != rune('r') {
					goto l290
				}
				position++
				if buffer[position] != rune('e') {
					goto l290
				}
				position++
				if buffer[position] != rune('q') {
					goto l290
				}
				position++
				if buffer[position] != rune('u') {
					goto l290
				}
				position++
				if buffer[position] != rune('i') {
					goto l290
				}
				position++
				if buffer[position] != rune('r') {
					goto l290
				}
				position++
				if buffer[position] != rune('e') {
					goto l290
				}
				position++
				if buffer[position] != rune('d') {
					goto l290
				}
				position++
				depth--
				add(ruleRequired, position291)
			}
			return true
		l290:
			position, tokenIndex, depth = position290, tokenIndex290, depth290
			return false
		},
		/* 70 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position292, tokenIndex292, depth292 := position, tokenIndex, depth
			{
				position293 := position
				depth++
				if buffer[position] != rune('o') {
					goto l292
				}
				position++
				if buffer[position] != rune('n') {
					goto l292
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l292
				}
				if !_rules[ruleName]() {
					goto l292
				}
				depth--
				add(ruleOn, position293)
			}
			return true
		l292:
			position, tokenIndex, depth = position292, tokenIndex292, depth292
			return false
		},
		/* 71 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position294, tokenIndex294, depth294 := position, tokenIndex, depth
			{
				position295 := position
				depth++
				if buffer[position] != rune('a') {
					goto l294
				}
				position++
				if buffer[position] != rune('u') {
					goto l294
				}
				position++
				if buffer[position] != rune('t') {
					goto l294
				}
				position++
				if buffer[position] != rune('o') {
					goto l294
				}
				position++
				depth--
				add(ruleAuto, position295)
			}
			return true
		l294:
			position, tokenIndex, depth = position294, tokenIndex294, depth294
			return false
		},
		/* 72 Default <- <Action1> */
		func() bool {
			position296, tokenIndex296, depth296 := position, tokenIndex, depth
			{
				position297 := position
				depth++
				if !_rules[ruleAction1]() {
					goto l296
				}
				depth--
				add(ruleDefault, position297)
			}
			return true
		l296:
			position, tokenIndex, depth = position296, tokenIndex296, depth296
			return false
		},
		/* 73 Sync <- <('s' 'y' 'n' 'c' '[' Level7 ((((LambdaExpr LambdaExt) / (LambdaOrExpr LambdaOrExpr)) (('|' Expression) / Default)) / (LambdaOrExpr Default Default)) ']')> */
		func() bool {
			position298, tokenIndex298, depth298 := position, tokenIndex, depth
			{
				position299 := position
				depth++
				if buffer[position] != rune('s') {
					goto l298
				}
				position++
				if buffer[position] != rune('y') {
					goto l298
				}
				position++
				if buffer[position] != rune('n') {
					goto l298
				}
				position++
				if buffer[position] != rune('c') {
					goto l298
				}
				position++
				if buffer[position] != rune('[') {
					goto l298
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l298
				}
				{
					position300, tokenIndex300, depth300 := position, tokenIndex, depth
					{
						position302, tokenIndex302, depth302 := position, tokenIndex, depth
						if !_rules[ruleLambdaExpr]() {
							goto l303
						}
						if !_rules[ruleLambdaExt]() {
							goto l303
						}
						goto l302
					l303:
						position, tokenIndex, depth = position302, tokenIndex302, depth302
						if !_rules[ruleLambdaOrExpr]() {
							goto l301
						}
						if !_rules[ruleLambdaOrExpr]() {
							goto l301
						}
					}
				l302:
					{
						position304, tokenIndex304, depth304 := position, tokenIndex, depth
						if buffer[position] != rune('|') {
							goto l305
						}
						position++
						if !_rules[ruleExpression]() {
							goto l305
						}
						goto l304
					l305:
						position, tokenIndex, depth = position304, tokenIndex304, depth304
						if !_rules[ruleDefault]() {
							goto l301
						}
					}
				l304:
					goto l300
				l301:
					position, tokenIndex, depth = position300, tokenIndex300, depth300
					if !_rules[ruleLambdaOrExpr]() {
						goto l298
					}
					if !_rules[ruleDefault]() {
						goto l298
					}
					if !_rules[ruleDefault]() {
						goto l298
					}
				}
			l300:
				if buffer[position] != rune(']') {
					goto l298
				}
				position++
				depth--
				add(ruleSync, position299)
			}
			return true
		l298:
			position, tokenIndex, depth = position298, tokenIndex298, depth298
			return false
		},
		/* 74 LambdaExt <- <(',' Expression)> */
		func() bool {
			position306, tokenIndex306, depth306 := position, tokenIndex, depth
			{
				position307 := position
				depth++
				if buffer[position] != rune(',') {
					goto l306
				}
				position++
				if !_rules[ruleExpression]() {
					goto l306
				}
				depth--
				add(ruleLambdaExt, position307)
			}
			return true
		l306:
			position, tokenIndex, depth = position306, tokenIndex306, depth306
			return false
		},
		/* 75 LambdaOrExpr <- <(LambdaExpr / ('|' Expression))> */
		func() bool {
			position308, tokenIndex308, depth308 := position, tokenIndex, depth
			{
				position309 := position
				depth++
				{
					position310, tokenIndex310, depth310 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l311
					}
					goto l310
				l311:
					position, tokenIndex, depth = position310, tokenIndex310, depth310
					if buffer[position] != rune('|') {
						goto l308
					}
					position++
					if !_rules[ruleExpression]() {
						goto l308
					}
				}
			l310:
				depth--
				add(ruleLambdaOrExpr, position309)
			}
			return true
		l308:
			position, tokenIndex, depth = position308, tokenIndex308, depth308
			return false
		},
		/* 76 Catch <- <('c' 'a' 't' 'c' 'h' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position312, tokenIndex312, depth312 := position, tokenIndex, depth
			{
				position313 := position
				depth++
				if buffer[position] != rune('c') {
					goto l312
				}
				position++
				if buffer[position] != rune('a') {
					goto l312
				}
				position++
				if buffer[position] != rune('t') {
					goto l312
				}
				position++
				if buffer[position] != rune('c') {
					goto l312
				}
				position++
				if buffer[position] != rune('h') {
					goto l312
				}
				position++
				if buffer[position] != rune('[') {
					goto l312
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l312
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l312
				}
				if buffer[position] != rune(']') {
					goto l312
				}
				position++
				depth--
				add(ruleCatch, position313)
			}
			return true
		l312:
			position, tokenIndex, depth = position312, tokenIndex312, depth312
			return false
		},
		/* 77 MapMapping <- <('m' 'a' 'p' '{' Level7 LambdaOrExpr '}')> */
		func() bool {
			position314, tokenIndex314, depth314 := position, tokenIndex, depth
			{
				position315 := position
				depth++
				if buffer[position] != rune('m') {
					goto l314
				}
				position++
				if buffer[position] != rune('a') {
					goto l314
				}
				position++
				if buffer[position] != rune('p') {
					goto l314
				}
				position++
				if buffer[position] != rune('{') {
					goto l314
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l314
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l314
				}
				if buffer[position] != rune('}') {
					goto l314
				}
				position++
				depth--
				add(ruleMapMapping, position315)
			}
			return true
		l314:
			position, tokenIndex, depth = position314, tokenIndex314, depth314
			return false
		},
		/* 78 Mapping <- <('m' 'a' 'p' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position316, tokenIndex316, depth316 := position, tokenIndex, depth
			{
				position317 := position
				depth++
				if buffer[position] != rune('m') {
					goto l316
				}
				position++
				if buffer[position] != rune('a') {
					goto l316
				}
				position++
				if buffer[position] != rune('p') {
					goto l316
				}
				position++
				if buffer[position] != rune('[') {
					goto l316
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l316
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l316
				}
				if buffer[position] != rune(']') {
					goto l316
				}
				position++
				depth--
				add(ruleMapping, position317)
			}
			return true
		l316:
			position, tokenIndex, depth = position316, tokenIndex316, depth316
			return false
		},
		/* 79 MapSelection <- <('s' 'e' 'l' 'e' 'c' 't' '{' Level7 LambdaOrExpr '}')> */
		func() bool {
			position318, tokenIndex318, depth318 := position, tokenIndex, depth
			{
				position319 := position
				depth++
				if buffer[position] != rune('s') {
					goto l318
				}
				position++
				if buffer[position] != rune('e') {
					goto l318
				}
				position++
				if buffer[position] != rune('l') {
					goto l318
				}
				position++
				if buffer[position] != rune('e') {
					goto l318
				}
				position++
				if buffer[position] != rune('c') {
					goto l318
				}
				position++
				if buffer[position] != rune('t') {
					goto l318
				}
				position++
				if buffer[position] != rune('{') {
					goto l318
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l318
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l318
				}
				if buffer[position] != rune('}') {
					goto l318
				}
				position++
				depth--
				add(ruleMapSelection, position319)
			}
			return true
		l318:
			position, tokenIndex, depth = position318, tokenIndex318, depth318
			return false
		},
		/* 80 Selection <- <('s' 'e' 'l' 'e' 'c' 't' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position320, tokenIndex320, depth320 := position, tokenIndex, depth
			{
				position321 := position
				depth++
				if buffer[position] != rune('s') {
					goto l320
				}
				position++
				if buffer[position] != rune('e') {
					goto l320
				}
				position++
				if buffer[position] != rune('l') {
					goto l320
				}
				position++
				if buffer[position] != rune('e') {
					goto l320
				}
				position++
				if buffer[position] != rune('c') {
					goto l320
				}
				position++
				if buffer[position] != rune('t') {
					goto l320
				}
				position++
				if buffer[position] != rune('[') {
					goto l320
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l320
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l320
				}
				if buffer[position] != rune(']') {
					goto l320
				}
				position++
				depth--
				add(ruleSelection, position321)
			}
			return true
		l320:
			position, tokenIndex, depth = position320, tokenIndex320, depth320
			return false
		},
		/* 81 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 LambdaOrExpr ']')> */
		func() bool {
			position322, tokenIndex322, depth322 := position, tokenIndex, depth
			{
				position323 := position
				depth++
				if buffer[position] != rune('s') {
					goto l322
				}
				position++
				if buffer[position] != rune('u') {
					goto l322
				}
				position++
				if buffer[position] != rune('m') {
					goto l322
				}
				position++
				if buffer[position] != rune('[') {
					goto l322
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l322
				}
				if buffer[position] != rune('|') {
					goto l322
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l322
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l322
				}
				if buffer[position] != rune(']') {
					goto l322
				}
				position++
				depth--
				add(ruleSum, position323)
			}
			return true
		l322:
			position, tokenIndex, depth = position322, tokenIndex322, depth322
			return false
		},
		/* 82 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position324, tokenIndex324, depth324 := position, tokenIndex, depth
			{
				position325 := position
				depth++
				if buffer[position] != rune('l') {
					goto l324
				}
				position++
				if buffer[position] != rune('a') {
					goto l324
				}
				position++
				if buffer[position] != rune('m') {
					goto l324
				}
				position++
				if buffer[position] != rune('b') {
					goto l324
				}
				position++
				if buffer[position] != rune('d') {
					goto l324
				}
				position++
				if buffer[position] != rune('a') {
					goto l324
				}
				position++
				{
					position326, tokenIndex326, depth326 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l327
					}
					goto l326
				l327:
					position, tokenIndex, depth = position326, tokenIndex326, depth326
					if !_rules[ruleLambdaExpr]() {
						goto l324
					}
				}
			l326:
				depth--
				add(ruleLambda, position325)
			}
			return true
		l324:
			position, tokenIndex, depth = position324, tokenIndex324, depth324
			return false
		},
		/* 83 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position328, tokenIndex328, depth328 := position, tokenIndex, depth
			{
				position329 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l328
				}
				if !_rules[ruleExpression]() {
					goto l328
				}
				depth--
				add(ruleLambdaRef, position329)
			}
			return true
		l328:
			position, tokenIndex, depth = position328, tokenIndex328, depth328
			return false
		},
		/* 84 LambdaExpr <- <(ws Params ws ('-' '>') Expression)> */
		func() bool {
			position330, tokenIndex330, depth330 := position, tokenIndex, depth
			{
				position331 := position
				depth++
				if !_rules[rulews]() {
					goto l330
				}
				if !_rules[ruleParams]() {
					goto l330
				}
				if !_rules[rulews]() {
					goto l330
				}
				if buffer[position] != rune('-') {
					goto l330
				}
				position++
				if buffer[position] != rune('>') {
					goto l330
				}
				position++
				if !_rules[ruleExpression]() {
					goto l330
				}
				depth--
				add(ruleLambdaExpr, position331)
			}
			return true
		l330:
			position, tokenIndex, depth = position330, tokenIndex330, depth330
			return false
		},
		/* 85 Params <- <('|' StartParams ws Names? '|')> */
		func() bool {
			position332, tokenIndex332, depth332 := position, tokenIndex, depth
			{
				position333 := position
				depth++
				if buffer[position] != rune('|') {
					goto l332
				}
				position++
				if !_rules[ruleStartParams]() {
					goto l332
				}
				if !_rules[rulews]() {
					goto l332
				}
				{
					position334, tokenIndex334, depth334 := position, tokenIndex, depth
					if !_rules[ruleNames]() {
						goto l334
					}
					goto l335
				l334:
					position, tokenIndex, depth = position334, tokenIndex334, depth334
				}
			l335:
				if buffer[position] != rune('|') {
					goto l332
				}
				position++
				depth--
				add(ruleParams, position333)
			}
			return true
		l332:
			position, tokenIndex, depth = position332, tokenIndex332, depth332
			return false
		},
		/* 86 StartParams <- <Action2> */
		func() bool {
			position336, tokenIndex336, depth336 := position, tokenIndex, depth
			{
				position337 := position
				depth++
				if !_rules[ruleAction2]() {
					goto l336
				}
				depth--
				add(ruleStartParams, position337)
			}
			return true
		l336:
			position, tokenIndex, depth = position336, tokenIndex336, depth336
			return false
		},
		/* 87 Names <- <(NextName (',' NextName)* DefaultValue? (',' NextName DefaultValue)* VarParams?)> */
		func() bool {
			position338, tokenIndex338, depth338 := position, tokenIndex, depth
			{
				position339 := position
				depth++
				if !_rules[ruleNextName]() {
					goto l338
				}
			l340:
				{
					position341, tokenIndex341, depth341 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l341
					}
					position++
					if !_rules[ruleNextName]() {
						goto l341
					}
					goto l340
				l341:
					position, tokenIndex, depth = position341, tokenIndex341, depth341
				}
				{
					position342, tokenIndex342, depth342 := position, tokenIndex, depth
					if !_rules[ruleDefaultValue]() {
						goto l342
					}
					goto l343
				l342:
					position, tokenIndex, depth = position342, tokenIndex342, depth342
				}
			l343:
			l344:
				{
					position345, tokenIndex345, depth345 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l345
					}
					position++
					if !_rules[ruleNextName]() {
						goto l345
					}
					if !_rules[ruleDefaultValue]() {
						goto l345
					}
					goto l344
				l345:
					position, tokenIndex, depth = position345, tokenIndex345, depth345
				}
				{
					position346, tokenIndex346, depth346 := position, tokenIndex, depth
					if !_rules[ruleVarParams]() {
						goto l346
					}
					goto l347
				l346:
					position, tokenIndex, depth = position346, tokenIndex346, depth346
				}
			l347:
				depth--
				add(ruleNames, position339)
			}
			return true
		l338:
			position, tokenIndex, depth = position338, tokenIndex338, depth338
			return false
		},
		/* 88 NextName <- <(ws Name ws)> */
		func() bool {
			position348, tokenIndex348, depth348 := position, tokenIndex, depth
			{
				position349 := position
				depth++
				if !_rules[rulews]() {
					goto l348
				}
				if !_rules[ruleName]() {
					goto l348
				}
				if !_rules[rulews]() {
					goto l348
				}
				depth--
				add(ruleNextName, position349)
			}
			return true
		l348:
			position, tokenIndex, depth = position348, tokenIndex348, depth348
			return false
		},
		/* 89 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position350, tokenIndex350, depth350 := position, tokenIndex, depth
			{
				position351 := position
				depth++
				{
					position354, tokenIndex354, depth354 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l355
					}
					position++
					goto l354
				l355:
					position, tokenIndex, depth = position354, tokenIndex354, depth354
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l356
					}
					position++
					goto l354
				l356:
					position, tokenIndex, depth = position354, tokenIndex354, depth354
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l357
					}
					position++
					goto l354
				l357:
					position, tokenIndex, depth = position354, tokenIndex354, depth354
					if buffer[position] != rune('_') {
						goto l350
					}
					position++
				}
			l354:
			l352:
				{
					position353, tokenIndex353, depth353 := position, tokenIndex, depth
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
							goto l353
						}
						position++
					}
				l358:
					goto l352
				l353:
					position, tokenIndex, depth = position353, tokenIndex353, depth353
				}
				depth--
				add(ruleName, position351)
			}
			return true
		l350:
			position, tokenIndex, depth = position350, tokenIndex350, depth350
			return false
		},
		/* 90 DefaultValue <- <('=' Expression)> */
		func() bool {
			position362, tokenIndex362, depth362 := position, tokenIndex, depth
			{
				position363 := position
				depth++
				if buffer[position] != rune('=') {
					goto l362
				}
				position++
				if !_rules[ruleExpression]() {
					goto l362
				}
				depth--
				add(ruleDefaultValue, position363)
			}
			return true
		l362:
			position, tokenIndex, depth = position362, tokenIndex362, depth362
			return false
		},
		/* 91 VarParams <- <('.' '.' '.' ws)> */
		func() bool {
			position364, tokenIndex364, depth364 := position, tokenIndex, depth
			{
				position365 := position
				depth++
				if buffer[position] != rune('.') {
					goto l364
				}
				position++
				if buffer[position] != rune('.') {
					goto l364
				}
				position++
				if buffer[position] != rune('.') {
					goto l364
				}
				position++
				if !_rules[rulews]() {
					goto l364
				}
				depth--
				add(ruleVarParams, position365)
			}
			return true
		l364:
			position, tokenIndex, depth = position364, tokenIndex364, depth364
			return false
		},
		/* 92 Reference <- <(((Tag ('.' / Key)) / ('.'? Key)) FollowUpRef)> */
		func() bool {
			position366, tokenIndex366, depth366 := position, tokenIndex, depth
			{
				position367 := position
				depth++
				{
					position368, tokenIndex368, depth368 := position, tokenIndex, depth
					if !_rules[ruleTag]() {
						goto l369
					}
					{
						position370, tokenIndex370, depth370 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l371
						}
						position++
						goto l370
					l371:
						position, tokenIndex, depth = position370, tokenIndex370, depth370
						if !_rules[ruleKey]() {
							goto l369
						}
					}
				l370:
					goto l368
				l369:
					position, tokenIndex, depth = position368, tokenIndex368, depth368
					{
						position372, tokenIndex372, depth372 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l372
						}
						position++
						goto l373
					l372:
						position, tokenIndex, depth = position372, tokenIndex372, depth372
					}
				l373:
					if !_rules[ruleKey]() {
						goto l366
					}
				}
			l368:
				if !_rules[ruleFollowUpRef]() {
					goto l366
				}
				depth--
				add(ruleReference, position367)
			}
			return true
		l366:
			position, tokenIndex, depth = position366, tokenIndex366, depth366
			return false
		},
		/* 93 Tag <- <(TagName (':' ':'))> */
		func() bool {
			position374, tokenIndex374, depth374 := position, tokenIndex, depth
			{
				position375 := position
				depth++
				if !_rules[ruleTagName]() {
					goto l374
				}
				if buffer[position] != rune(':') {
					goto l374
				}
				position++
				if buffer[position] != rune(':') {
					goto l374
				}
				position++
				depth--
				add(ruleTag, position375)
			}
			return true
		l374:
			position, tokenIndex, depth = position374, tokenIndex374, depth374
			return false
		},
		/* 94 TagName <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position376, tokenIndex376, depth376 := position, tokenIndex, depth
			{
				position377 := position
				depth++
				{
					position380, tokenIndex380, depth380 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l381
					}
					position++
					goto l380
				l381:
					position, tokenIndex, depth = position380, tokenIndex380, depth380
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l382
					}
					position++
					goto l380
				l382:
					position, tokenIndex, depth = position380, tokenIndex380, depth380
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l383
					}
					position++
					goto l380
				l383:
					position, tokenIndex, depth = position380, tokenIndex380, depth380
					if buffer[position] != rune('_') {
						goto l376
					}
					position++
				}
			l380:
			l378:
				{
					position379, tokenIndex379, depth379 := position, tokenIndex, depth
					{
						position384, tokenIndex384, depth384 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l385
						}
						position++
						goto l384
					l385:
						position, tokenIndex, depth = position384, tokenIndex384, depth384
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l386
						}
						position++
						goto l384
					l386:
						position, tokenIndex, depth = position384, tokenIndex384, depth384
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l387
						}
						position++
						goto l384
					l387:
						position, tokenIndex, depth = position384, tokenIndex384, depth384
						if buffer[position] != rune('_') {
							goto l379
						}
						position++
					}
				l384:
					goto l378
				l379:
					position, tokenIndex, depth = position379, tokenIndex379, depth379
				}
				depth--
				add(ruleTagName, position377)
			}
			return true
		l376:
			position, tokenIndex, depth = position376, tokenIndex376, depth376
			return false
		},
		/* 95 FollowUpRef <- <PathComponent*> */
		func() bool {
			{
				position389 := position
				depth++
			l390:
				{
					position391, tokenIndex391, depth391 := position, tokenIndex, depth
					if !_rules[rulePathComponent]() {
						goto l391
					}
					goto l390
				l391:
					position, tokenIndex, depth = position391, tokenIndex391, depth391
				}
				depth--
				add(ruleFollowUpRef, position389)
			}
			return true
		},
		/* 96 PathComponent <- <(('.' Key) / ('.'? Index))> */
		func() bool {
			position392, tokenIndex392, depth392 := position, tokenIndex, depth
			{
				position393 := position
				depth++
				{
					position394, tokenIndex394, depth394 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l395
					}
					position++
					if !_rules[ruleKey]() {
						goto l395
					}
					goto l394
				l395:
					position, tokenIndex, depth = position394, tokenIndex394, depth394
					{
						position396, tokenIndex396, depth396 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l396
						}
						position++
						goto l397
					l396:
						position, tokenIndex, depth = position396, tokenIndex396, depth396
					}
				l397:
					if !_rules[ruleIndex]() {
						goto l392
					}
				}
			l394:
				depth--
				add(rulePathComponent, position393)
			}
			return true
		l392:
			position, tokenIndex, depth = position392, tokenIndex392, depth392
			return false
		},
		/* 97 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position398, tokenIndex398, depth398 := position, tokenIndex, depth
			{
				position399 := position
				depth++
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
						goto l398
					}
					position++
				}
			l400:
			l404:
				{
					position405, tokenIndex405, depth405 := position, tokenIndex, depth
					{
						position406, tokenIndex406, depth406 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l407
						}
						position++
						goto l406
					l407:
						position, tokenIndex, depth = position406, tokenIndex406, depth406
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l408
						}
						position++
						goto l406
					l408:
						position, tokenIndex, depth = position406, tokenIndex406, depth406
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l409
						}
						position++
						goto l406
					l409:
						position, tokenIndex, depth = position406, tokenIndex406, depth406
						if buffer[position] != rune('_') {
							goto l410
						}
						position++
						goto l406
					l410:
						position, tokenIndex, depth = position406, tokenIndex406, depth406
						if buffer[position] != rune('-') {
							goto l405
						}
						position++
					}
				l406:
					goto l404
				l405:
					position, tokenIndex, depth = position405, tokenIndex405, depth405
				}
				{
					position411, tokenIndex411, depth411 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l411
					}
					position++
					{
						position413, tokenIndex413, depth413 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l414
						}
						position++
						goto l413
					l414:
						position, tokenIndex, depth = position413, tokenIndex413, depth413
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l415
						}
						position++
						goto l413
					l415:
						position, tokenIndex, depth = position413, tokenIndex413, depth413
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l416
						}
						position++
						goto l413
					l416:
						position, tokenIndex, depth = position413, tokenIndex413, depth413
						if buffer[position] != rune('_') {
							goto l411
						}
						position++
					}
				l413:
				l417:
					{
						position418, tokenIndex418, depth418 := position, tokenIndex, depth
						{
							position419, tokenIndex419, depth419 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l420
							}
							position++
							goto l419
						l420:
							position, tokenIndex, depth = position419, tokenIndex419, depth419
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l421
							}
							position++
							goto l419
						l421:
							position, tokenIndex, depth = position419, tokenIndex419, depth419
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l422
							}
							position++
							goto l419
						l422:
							position, tokenIndex, depth = position419, tokenIndex419, depth419
							if buffer[position] != rune('_') {
								goto l423
							}
							position++
							goto l419
						l423:
							position, tokenIndex, depth = position419, tokenIndex419, depth419
							if buffer[position] != rune('-') {
								goto l418
							}
							position++
						}
					l419:
						goto l417
					l418:
						position, tokenIndex, depth = position418, tokenIndex418, depth418
					}
					goto l412
				l411:
					position, tokenIndex, depth = position411, tokenIndex411, depth411
				}
			l412:
				depth--
				add(ruleKey, position399)
			}
			return true
		l398:
			position, tokenIndex, depth = position398, tokenIndex398, depth398
			return false
		},
		/* 98 Index <- <('[' '-'? [0-9]+ ']')> */
		func() bool {
			position424, tokenIndex424, depth424 := position, tokenIndex, depth
			{
				position425 := position
				depth++
				if buffer[position] != rune('[') {
					goto l424
				}
				position++
				{
					position426, tokenIndex426, depth426 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l426
					}
					position++
					goto l427
				l426:
					position, tokenIndex, depth = position426, tokenIndex426, depth426
				}
			l427:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l424
				}
				position++
			l428:
				{
					position429, tokenIndex429, depth429 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l429
					}
					position++
					goto l428
				l429:
					position, tokenIndex, depth = position429, tokenIndex429, depth429
				}
				if buffer[position] != rune(']') {
					goto l424
				}
				position++
				depth--
				add(ruleIndex, position425)
			}
			return true
		l424:
			position, tokenIndex, depth = position424, tokenIndex424, depth424
			return false
		},
		/* 99 IP <- <([0-9]+ '.' [0-9]+ '.' [0-9]+ '.' [0-9]+)> */
		func() bool {
			position430, tokenIndex430, depth430 := position, tokenIndex, depth
			{
				position431 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l430
				}
				position++
			l432:
				{
					position433, tokenIndex433, depth433 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l433
					}
					position++
					goto l432
				l433:
					position, tokenIndex, depth = position433, tokenIndex433, depth433
				}
				if buffer[position] != rune('.') {
					goto l430
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l430
				}
				position++
			l434:
				{
					position435, tokenIndex435, depth435 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l435
					}
					position++
					goto l434
				l435:
					position, tokenIndex, depth = position435, tokenIndex435, depth435
				}
				if buffer[position] != rune('.') {
					goto l430
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l430
				}
				position++
			l436:
				{
					position437, tokenIndex437, depth437 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l437
					}
					position++
					goto l436
				l437:
					position, tokenIndex, depth = position437, tokenIndex437, depth437
				}
				if buffer[position] != rune('.') {
					goto l430
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l430
				}
				position++
			l438:
				{
					position439, tokenIndex439, depth439 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l439
					}
					position++
					goto l438
				l439:
					position, tokenIndex, depth = position439, tokenIndex439, depth439
				}
				depth--
				add(ruleIP, position431)
			}
			return true
		l430:
			position, tokenIndex, depth = position430, tokenIndex430, depth430
			return false
		},
		/* 100 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position441 := position
				depth++
			l442:
				{
					position443, tokenIndex443, depth443 := position, tokenIndex, depth
					{
						position444, tokenIndex444, depth444 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l445
						}
						position++
						goto l444
					l445:
						position, tokenIndex, depth = position444, tokenIndex444, depth444
						if buffer[position] != rune('\t') {
							goto l446
						}
						position++
						goto l444
					l446:
						position, tokenIndex, depth = position444, tokenIndex444, depth444
						if buffer[position] != rune('\n') {
							goto l447
						}
						position++
						goto l444
					l447:
						position, tokenIndex, depth = position444, tokenIndex444, depth444
						if buffer[position] != rune('\r') {
							goto l443
						}
						position++
					}
				l444:
					goto l442
				l443:
					position, tokenIndex, depth = position443, tokenIndex443, depth443
				}
				depth--
				add(rulews, position441)
			}
			return true
		},
		/* 101 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position448, tokenIndex448, depth448 := position, tokenIndex, depth
			{
				position449 := position
				depth++
				{
					position452, tokenIndex452, depth452 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l453
					}
					position++
					goto l452
				l453:
					position, tokenIndex, depth = position452, tokenIndex452, depth452
					if buffer[position] != rune('\t') {
						goto l454
					}
					position++
					goto l452
				l454:
					position, tokenIndex, depth = position452, tokenIndex452, depth452
					if buffer[position] != rune('\n') {
						goto l455
					}
					position++
					goto l452
				l455:
					position, tokenIndex, depth = position452, tokenIndex452, depth452
					if buffer[position] != rune('\r') {
						goto l448
					}
					position++
				}
			l452:
			l450:
				{
					position451, tokenIndex451, depth451 := position, tokenIndex, depth
					{
						position456, tokenIndex456, depth456 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l457
						}
						position++
						goto l456
					l457:
						position, tokenIndex, depth = position456, tokenIndex456, depth456
						if buffer[position] != rune('\t') {
							goto l458
						}
						position++
						goto l456
					l458:
						position, tokenIndex, depth = position456, tokenIndex456, depth456
						if buffer[position] != rune('\n') {
							goto l459
						}
						position++
						goto l456
					l459:
						position, tokenIndex, depth = position456, tokenIndex456, depth456
						if buffer[position] != rune('\r') {
							goto l451
						}
						position++
					}
				l456:
					goto l450
				l451:
					position, tokenIndex, depth = position451, tokenIndex451, depth451
				}
				depth--
				add(rulereq_ws, position449)
			}
			return true
		l448:
			position, tokenIndex, depth = position448, tokenIndex448, depth448
			return false
		},
		/* 103 Action0 <- <{}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 104 Action1 <- <{}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 105 Action2 <- <{}> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
	}
	p.rules = _rules
}
