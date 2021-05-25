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
		/* 5 TagMarker <- <('t' 'a' 'g' ':' '*'? TagName (':' TagName)*)> */
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
				if !_rules[ruleTagName]() {
					goto l25
				}
			l29:
				{
					position30, tokenIndex30, depth30 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l30
					}
					position++
					if !_rules[ruleTagName]() {
						goto l30
					}
					goto l29
				l30:
					position, tokenIndex, depth = position30, tokenIndex30, depth30
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
			position31, tokenIndex31, depth31 := position, tokenIndex, depth
			{
				position32 := position
				depth++
				if !_rules[ruleGrouped]() {
					goto l31
				}
				depth--
				add(ruleMarkerExpression, position32)
			}
			return true
		l31:
			position, tokenIndex, depth = position31, tokenIndex31, depth31
			return false
		},
		/* 7 Expression <- <((Scoped / LambdaExpr / Level7) ws)> */
		func() bool {
			position33, tokenIndex33, depth33 := position, tokenIndex, depth
			{
				position34 := position
				depth++
				{
					position35, tokenIndex35, depth35 := position, tokenIndex, depth
					if !_rules[ruleScoped]() {
						goto l36
					}
					goto l35
				l36:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !_rules[ruleLambdaExpr]() {
						goto l37
					}
					goto l35
				l37:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !_rules[ruleLevel7]() {
						goto l33
					}
				}
			l35:
				if !_rules[rulews]() {
					goto l33
				}
				depth--
				add(ruleExpression, position34)
			}
			return true
		l33:
			position, tokenIndex, depth = position33, tokenIndex33, depth33
			return false
		},
		/* 8 Scoped <- <(ws Scope ws Expression)> */
		func() bool {
			position38, tokenIndex38, depth38 := position, tokenIndex, depth
			{
				position39 := position
				depth++
				if !_rules[rulews]() {
					goto l38
				}
				if !_rules[ruleScope]() {
					goto l38
				}
				if !_rules[rulews]() {
					goto l38
				}
				if !_rules[ruleExpression]() {
					goto l38
				}
				depth--
				add(ruleScoped, position39)
			}
			return true
		l38:
			position, tokenIndex, depth = position38, tokenIndex38, depth38
			return false
		},
		/* 9 Scope <- <(CreateScope ws Assignments? ')')> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				if !_rules[ruleCreateScope]() {
					goto l40
				}
				if !_rules[rulews]() {
					goto l40
				}
				{
					position42, tokenIndex42, depth42 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l42
					}
					goto l43
				l42:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
				}
			l43:
				if buffer[position] != rune(')') {
					goto l40
				}
				position++
				depth--
				add(ruleScope, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 10 CreateScope <- <'('> */
		func() bool {
			position44, tokenIndex44, depth44 := position, tokenIndex, depth
			{
				position45 := position
				depth++
				if buffer[position] != rune('(') {
					goto l44
				}
				position++
				depth--
				add(ruleCreateScope, position45)
			}
			return true
		l44:
			position, tokenIndex, depth = position44, tokenIndex44, depth44
			return false
		},
		/* 11 Level7 <- <(ws Level6 (req_ws Or)*)> */
		func() bool {
			position46, tokenIndex46, depth46 := position, tokenIndex, depth
			{
				position47 := position
				depth++
				if !_rules[rulews]() {
					goto l46
				}
				if !_rules[ruleLevel6]() {
					goto l46
				}
			l48:
				{
					position49, tokenIndex49, depth49 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l49
					}
					if !_rules[ruleOr]() {
						goto l49
					}
					goto l48
				l49:
					position, tokenIndex, depth = position49, tokenIndex49, depth49
				}
				depth--
				add(ruleLevel7, position47)
			}
			return true
		l46:
			position, tokenIndex, depth = position46, tokenIndex46, depth46
			return false
		},
		/* 12 Or <- <(OrOp req_ws Level6)> */
		func() bool {
			position50, tokenIndex50, depth50 := position, tokenIndex, depth
			{
				position51 := position
				depth++
				if !_rules[ruleOrOp]() {
					goto l50
				}
				if !_rules[rulereq_ws]() {
					goto l50
				}
				if !_rules[ruleLevel6]() {
					goto l50
				}
				depth--
				add(ruleOr, position51)
			}
			return true
		l50:
			position, tokenIndex, depth = position50, tokenIndex50, depth50
			return false
		},
		/* 13 OrOp <- <(('|' '|') / ('/' '/'))> */
		func() bool {
			position52, tokenIndex52, depth52 := position, tokenIndex, depth
			{
				position53 := position
				depth++
				{
					position54, tokenIndex54, depth54 := position, tokenIndex, depth
					if buffer[position] != rune('|') {
						goto l55
					}
					position++
					if buffer[position] != rune('|') {
						goto l55
					}
					position++
					goto l54
				l55:
					position, tokenIndex, depth = position54, tokenIndex54, depth54
					if buffer[position] != rune('/') {
						goto l52
					}
					position++
					if buffer[position] != rune('/') {
						goto l52
					}
					position++
				}
			l54:
				depth--
				add(ruleOrOp, position53)
			}
			return true
		l52:
			position, tokenIndex, depth = position52, tokenIndex52, depth52
			return false
		},
		/* 14 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position56, tokenIndex56, depth56 := position, tokenIndex, depth
			{
				position57 := position
				depth++
				{
					position58, tokenIndex58, depth58 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l59
					}
					goto l58
				l59:
					position, tokenIndex, depth = position58, tokenIndex58, depth58
					if !_rules[ruleLevel5]() {
						goto l56
					}
				}
			l58:
				depth--
				add(ruleLevel6, position57)
			}
			return true
		l56:
			position, tokenIndex, depth = position56, tokenIndex56, depth56
			return false
		},
		/* 15 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position60, tokenIndex60, depth60 := position, tokenIndex, depth
			{
				position61 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l60
				}
				if !_rules[rulews]() {
					goto l60
				}
				if buffer[position] != rune('?') {
					goto l60
				}
				position++
				if !_rules[ruleExpression]() {
					goto l60
				}
				if buffer[position] != rune(':') {
					goto l60
				}
				position++
				if !_rules[ruleExpression]() {
					goto l60
				}
				depth--
				add(ruleConditional, position61)
			}
			return true
		l60:
			position, tokenIndex, depth = position60, tokenIndex60, depth60
			return false
		},
		/* 16 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position62, tokenIndex62, depth62 := position, tokenIndex, depth
			{
				position63 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l62
				}
			l64:
				{
					position65, tokenIndex65, depth65 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l65
					}
					goto l64
				l65:
					position, tokenIndex, depth = position65, tokenIndex65, depth65
				}
				depth--
				add(ruleLevel5, position63)
			}
			return true
		l62:
			position, tokenIndex, depth = position62, tokenIndex62, depth62
			return false
		},
		/* 17 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position66, tokenIndex66, depth66 := position, tokenIndex, depth
			{
				position67 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l66
				}
				if !_rules[ruleLevel4]() {
					goto l66
				}
				depth--
				add(ruleConcatenation, position67)
			}
			return true
		l66:
			position, tokenIndex, depth = position66, tokenIndex66, depth66
			return false
		},
		/* 18 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position68, tokenIndex68, depth68 := position, tokenIndex, depth
			{
				position69 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l68
				}
			l70:
				{
					position71, tokenIndex71, depth71 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l71
					}
					{
						position72, tokenIndex72, depth72 := position, tokenIndex, depth
						if !_rules[ruleLogOr]() {
							goto l73
						}
						goto l72
					l73:
						position, tokenIndex, depth = position72, tokenIndex72, depth72
						if !_rules[ruleLogAnd]() {
							goto l71
						}
					}
				l72:
					goto l70
				l71:
					position, tokenIndex, depth = position71, tokenIndex71, depth71
				}
				depth--
				add(ruleLevel4, position69)
			}
			return true
		l68:
			position, tokenIndex, depth = position68, tokenIndex68, depth68
			return false
		},
		/* 19 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position74, tokenIndex74, depth74 := position, tokenIndex, depth
			{
				position75 := position
				depth++
				if buffer[position] != rune('-') {
					goto l74
				}
				position++
				if buffer[position] != rune('o') {
					goto l74
				}
				position++
				if buffer[position] != rune('r') {
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
				add(ruleLogOr, position75)
			}
			return true
		l74:
			position, tokenIndex, depth = position74, tokenIndex74, depth74
			return false
		},
		/* 20 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position76, tokenIndex76, depth76 := position, tokenIndex, depth
			{
				position77 := position
				depth++
				if buffer[position] != rune('-') {
					goto l76
				}
				position++
				if buffer[position] != rune('a') {
					goto l76
				}
				position++
				if buffer[position] != rune('n') {
					goto l76
				}
				position++
				if buffer[position] != rune('d') {
					goto l76
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l76
				}
				if !_rules[ruleLevel3]() {
					goto l76
				}
				depth--
				add(ruleLogAnd, position77)
			}
			return true
		l76:
			position, tokenIndex, depth = position76, tokenIndex76, depth76
			return false
		},
		/* 21 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position78, tokenIndex78, depth78 := position, tokenIndex, depth
			{
				position79 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l78
				}
			l80:
				{
					position81, tokenIndex81, depth81 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l81
					}
					if !_rules[ruleComparison]() {
						goto l81
					}
					goto l80
				l81:
					position, tokenIndex, depth = position81, tokenIndex81, depth81
				}
				depth--
				add(ruleLevel3, position79)
			}
			return true
		l78:
			position, tokenIndex, depth = position78, tokenIndex78, depth78
			return false
		},
		/* 22 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position82, tokenIndex82, depth82 := position, tokenIndex, depth
			{
				position83 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l82
				}
				if !_rules[rulereq_ws]() {
					goto l82
				}
				if !_rules[ruleLevel2]() {
					goto l82
				}
				depth--
				add(ruleComparison, position83)
			}
			return true
		l82:
			position, tokenIndex, depth = position82, tokenIndex82, depth82
			return false
		},
		/* 23 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position84, tokenIndex84, depth84 := position, tokenIndex, depth
			{
				position85 := position
				depth++
				{
					position86, tokenIndex86, depth86 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l87
					}
					position++
					if buffer[position] != rune('=') {
						goto l87
					}
					position++
					goto l86
				l87:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
					if buffer[position] != rune('!') {
						goto l88
					}
					position++
					if buffer[position] != rune('=') {
						goto l88
					}
					position++
					goto l86
				l88:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
					if buffer[position] != rune('<') {
						goto l89
					}
					position++
					if buffer[position] != rune('=') {
						goto l89
					}
					position++
					goto l86
				l89:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
					if buffer[position] != rune('>') {
						goto l90
					}
					position++
					if buffer[position] != rune('=') {
						goto l90
					}
					position++
					goto l86
				l90:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
					if buffer[position] != rune('>') {
						goto l91
					}
					position++
					goto l86
				l91:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
					if buffer[position] != rune('<') {
						goto l92
					}
					position++
					goto l86
				l92:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
					if buffer[position] != rune('>') {
						goto l84
					}
					position++
				}
			l86:
				depth--
				add(ruleCompareOp, position85)
			}
			return true
		l84:
			position, tokenIndex, depth = position84, tokenIndex84, depth84
			return false
		},
		/* 24 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position93, tokenIndex93, depth93 := position, tokenIndex, depth
			{
				position94 := position
				depth++
				if !_rules[ruleLevel1]() {
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
						if !_rules[ruleAddition]() {
							goto l98
						}
						goto l97
					l98:
						position, tokenIndex, depth = position97, tokenIndex97, depth97
						if !_rules[ruleSubtraction]() {
							goto l96
						}
					}
				l97:
					goto l95
				l96:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
				}
				depth--
				add(ruleLevel2, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 25 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position99, tokenIndex99, depth99 := position, tokenIndex, depth
			{
				position100 := position
				depth++
				if buffer[position] != rune('+') {
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
				add(ruleAddition, position100)
			}
			return true
		l99:
			position, tokenIndex, depth = position99, tokenIndex99, depth99
			return false
		},
		/* 26 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position101, tokenIndex101, depth101 := position, tokenIndex, depth
			{
				position102 := position
				depth++
				if buffer[position] != rune('-') {
					goto l101
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l101
				}
				if !_rules[ruleLevel1]() {
					goto l101
				}
				depth--
				add(ruleSubtraction, position102)
			}
			return true
		l101:
			position, tokenIndex, depth = position101, tokenIndex101, depth101
			return false
		},
		/* 27 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position103, tokenIndex103, depth103 := position, tokenIndex, depth
			{
				position104 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l103
				}
			l105:
				{
					position106, tokenIndex106, depth106 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l106
					}
					{
						position107, tokenIndex107, depth107 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l108
						}
						goto l107
					l108:
						position, tokenIndex, depth = position107, tokenIndex107, depth107
						if !_rules[ruleDivision]() {
							goto l109
						}
						goto l107
					l109:
						position, tokenIndex, depth = position107, tokenIndex107, depth107
						if !_rules[ruleModulo]() {
							goto l106
						}
					}
				l107:
					goto l105
				l106:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
				}
				depth--
				add(ruleLevel1, position104)
			}
			return true
		l103:
			position, tokenIndex, depth = position103, tokenIndex103, depth103
			return false
		},
		/* 28 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position110, tokenIndex110, depth110 := position, tokenIndex, depth
			{
				position111 := position
				depth++
				if buffer[position] != rune('*') {
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
				add(ruleMultiplication, position111)
			}
			return true
		l110:
			position, tokenIndex, depth = position110, tokenIndex110, depth110
			return false
		},
		/* 29 Division <- <('/' req_ws Level0)> */
		func() bool {
			position112, tokenIndex112, depth112 := position, tokenIndex, depth
			{
				position113 := position
				depth++
				if buffer[position] != rune('/') {
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
				add(ruleDivision, position113)
			}
			return true
		l112:
			position, tokenIndex, depth = position112, tokenIndex112, depth112
			return false
		},
		/* 30 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position114, tokenIndex114, depth114 := position, tokenIndex, depth
			{
				position115 := position
				depth++
				if buffer[position] != rune('%') {
					goto l114
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l114
				}
				if !_rules[ruleLevel0]() {
					goto l114
				}
				depth--
				add(ruleModulo, position115)
			}
			return true
		l114:
			position, tokenIndex, depth = position114, tokenIndex114, depth114
			return false
		},
		/* 31 Level0 <- <(IP / String / Number / Boolean / Undefined / Nil / Symbol / Not / Substitution / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position116, tokenIndex116, depth116 := position, tokenIndex, depth
			{
				position117 := position
				depth++
				{
					position118, tokenIndex118, depth118 := position, tokenIndex, depth
					if !_rules[ruleIP]() {
						goto l119
					}
					goto l118
				l119:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
					if !_rules[ruleString]() {
						goto l120
					}
					goto l118
				l120:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
					if !_rules[ruleNumber]() {
						goto l121
					}
					goto l118
				l121:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
					if !_rules[ruleBoolean]() {
						goto l122
					}
					goto l118
				l122:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
					if !_rules[ruleUndefined]() {
						goto l123
					}
					goto l118
				l123:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
					if !_rules[ruleNil]() {
						goto l124
					}
					goto l118
				l124:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
					if !_rules[ruleSymbol]() {
						goto l125
					}
					goto l118
				l125:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
					if !_rules[ruleNot]() {
						goto l126
					}
					goto l118
				l126:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
					if !_rules[ruleSubstitution]() {
						goto l127
					}
					goto l118
				l127:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
					if !_rules[ruleMerge]() {
						goto l128
					}
					goto l118
				l128:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
					if !_rules[ruleAuto]() {
						goto l129
					}
					goto l118
				l129:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
					if !_rules[ruleLambda]() {
						goto l130
					}
					goto l118
				l130:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
					if !_rules[ruleChained]() {
						goto l116
					}
				}
			l118:
				depth--
				add(ruleLevel0, position117)
			}
			return true
		l116:
			position, tokenIndex, depth = position116, tokenIndex116, depth116
			return false
		},
		/* 32 Chained <- <((MapMapping / Sync / Catch / Mapping / MapSelection / Selection / Sum / List / Map / Range / Grouped / Reference) ChainedQualifiedExpression*)> */
		func() bool {
			position131, tokenIndex131, depth131 := position, tokenIndex, depth
			{
				position132 := position
				depth++
				{
					position133, tokenIndex133, depth133 := position, tokenIndex, depth
					if !_rules[ruleMapMapping]() {
						goto l134
					}
					goto l133
				l134:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
					if !_rules[ruleSync]() {
						goto l135
					}
					goto l133
				l135:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
					if !_rules[ruleCatch]() {
						goto l136
					}
					goto l133
				l136:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
					if !_rules[ruleMapping]() {
						goto l137
					}
					goto l133
				l137:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
					if !_rules[ruleMapSelection]() {
						goto l138
					}
					goto l133
				l138:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
					if !_rules[ruleSelection]() {
						goto l139
					}
					goto l133
				l139:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
					if !_rules[ruleSum]() {
						goto l140
					}
					goto l133
				l140:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
					if !_rules[ruleList]() {
						goto l141
					}
					goto l133
				l141:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
					if !_rules[ruleMap]() {
						goto l142
					}
					goto l133
				l142:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
					if !_rules[ruleRange]() {
						goto l143
					}
					goto l133
				l143:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
					if !_rules[ruleGrouped]() {
						goto l144
					}
					goto l133
				l144:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
					if !_rules[ruleReference]() {
						goto l131
					}
				}
			l133:
			l145:
				{
					position146, tokenIndex146, depth146 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l146
					}
					goto l145
				l146:
					position, tokenIndex, depth = position146, tokenIndex146, depth146
				}
				depth--
				add(ruleChained, position132)
			}
			return true
		l131:
			position, tokenIndex, depth = position131, tokenIndex131, depth131
			return false
		},
		/* 33 ChainedQualifiedExpression <- <(ChainedCall / Currying / ChainedRef / ChainedDynRef / Projection)> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				{
					position149, tokenIndex149, depth149 := position, tokenIndex, depth
					if !_rules[ruleChainedCall]() {
						goto l150
					}
					goto l149
				l150:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
					if !_rules[ruleCurrying]() {
						goto l151
					}
					goto l149
				l151:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
					if !_rules[ruleChainedRef]() {
						goto l152
					}
					goto l149
				l152:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
					if !_rules[ruleChainedDynRef]() {
						goto l153
					}
					goto l149
				l153:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
					if !_rules[ruleProjection]() {
						goto l147
					}
				}
			l149:
				depth--
				add(ruleChainedQualifiedExpression, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 34 ChainedRef <- <(PathComponent FollowUpRef)> */
		func() bool {
			position154, tokenIndex154, depth154 := position, tokenIndex, depth
			{
				position155 := position
				depth++
				if !_rules[rulePathComponent]() {
					goto l154
				}
				if !_rules[ruleFollowUpRef]() {
					goto l154
				}
				depth--
				add(ruleChainedRef, position155)
			}
			return true
		l154:
			position, tokenIndex, depth = position154, tokenIndex154, depth154
			return false
		},
		/* 35 ChainedDynRef <- <('.'? '[' Expression ']')> */
		func() bool {
			position156, tokenIndex156, depth156 := position, tokenIndex, depth
			{
				position157 := position
				depth++
				{
					position158, tokenIndex158, depth158 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l158
					}
					position++
					goto l159
				l158:
					position, tokenIndex, depth = position158, tokenIndex158, depth158
				}
			l159:
				if buffer[position] != rune('[') {
					goto l156
				}
				position++
				if !_rules[ruleExpression]() {
					goto l156
				}
				if buffer[position] != rune(']') {
					goto l156
				}
				position++
				depth--
				add(ruleChainedDynRef, position157)
			}
			return true
		l156:
			position, tokenIndex, depth = position156, tokenIndex156, depth156
			return false
		},
		/* 36 Slice <- <Range> */
		func() bool {
			position160, tokenIndex160, depth160 := position, tokenIndex, depth
			{
				position161 := position
				depth++
				if !_rules[ruleRange]() {
					goto l160
				}
				depth--
				add(ruleSlice, position161)
			}
			return true
		l160:
			position, tokenIndex, depth = position160, tokenIndex160, depth160
			return false
		},
		/* 37 Currying <- <('*' ChainedCall)> */
		func() bool {
			position162, tokenIndex162, depth162 := position, tokenIndex, depth
			{
				position163 := position
				depth++
				if buffer[position] != rune('*') {
					goto l162
				}
				position++
				if !_rules[ruleChainedCall]() {
					goto l162
				}
				depth--
				add(ruleCurrying, position163)
			}
			return true
		l162:
			position, tokenIndex, depth = position162, tokenIndex162, depth162
			return false
		},
		/* 38 ChainedCall <- <(StartArguments NameArgumentList? ')')> */
		func() bool {
			position164, tokenIndex164, depth164 := position, tokenIndex, depth
			{
				position165 := position
				depth++
				if !_rules[ruleStartArguments]() {
					goto l164
				}
				{
					position166, tokenIndex166, depth166 := position, tokenIndex, depth
					if !_rules[ruleNameArgumentList]() {
						goto l166
					}
					goto l167
				l166:
					position, tokenIndex, depth = position166, tokenIndex166, depth166
				}
			l167:
				if buffer[position] != rune(')') {
					goto l164
				}
				position++
				depth--
				add(ruleChainedCall, position165)
			}
			return true
		l164:
			position, tokenIndex, depth = position164, tokenIndex164, depth164
			return false
		},
		/* 39 StartArguments <- <('(' ws)> */
		func() bool {
			position168, tokenIndex168, depth168 := position, tokenIndex, depth
			{
				position169 := position
				depth++
				if buffer[position] != rune('(') {
					goto l168
				}
				position++
				if !_rules[rulews]() {
					goto l168
				}
				depth--
				add(ruleStartArguments, position169)
			}
			return true
		l168:
			position, tokenIndex, depth = position168, tokenIndex168, depth168
			return false
		},
		/* 40 NameArgumentList <- <(((NextNameArgument (',' NextNameArgument)*) / NextExpression) (',' NextExpression)*)> */
		func() bool {
			position170, tokenIndex170, depth170 := position, tokenIndex, depth
			{
				position171 := position
				depth++
				{
					position172, tokenIndex172, depth172 := position, tokenIndex, depth
					if !_rules[ruleNextNameArgument]() {
						goto l173
					}
				l174:
					{
						position175, tokenIndex175, depth175 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l175
						}
						position++
						if !_rules[ruleNextNameArgument]() {
							goto l175
						}
						goto l174
					l175:
						position, tokenIndex, depth = position175, tokenIndex175, depth175
					}
					goto l172
				l173:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
					if !_rules[ruleNextExpression]() {
						goto l170
					}
				}
			l172:
			l176:
				{
					position177, tokenIndex177, depth177 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l177
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l177
					}
					goto l176
				l177:
					position, tokenIndex, depth = position177, tokenIndex177, depth177
				}
				depth--
				add(ruleNameArgumentList, position171)
			}
			return true
		l170:
			position, tokenIndex, depth = position170, tokenIndex170, depth170
			return false
		},
		/* 41 NextNameArgument <- <(ws Name ws '=' ws Expression ws)> */
		func() bool {
			position178, tokenIndex178, depth178 := position, tokenIndex, depth
			{
				position179 := position
				depth++
				if !_rules[rulews]() {
					goto l178
				}
				if !_rules[ruleName]() {
					goto l178
				}
				if !_rules[rulews]() {
					goto l178
				}
				if buffer[position] != rune('=') {
					goto l178
				}
				position++
				if !_rules[rulews]() {
					goto l178
				}
				if !_rules[ruleExpression]() {
					goto l178
				}
				if !_rules[rulews]() {
					goto l178
				}
				depth--
				add(ruleNextNameArgument, position179)
			}
			return true
		l178:
			position, tokenIndex, depth = position178, tokenIndex178, depth178
			return false
		},
		/* 42 ExpressionList <- <(NextExpression (',' NextExpression)*)> */
		func() bool {
			position180, tokenIndex180, depth180 := position, tokenIndex, depth
			{
				position181 := position
				depth++
				if !_rules[ruleNextExpression]() {
					goto l180
				}
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
				add(ruleExpressionList, position181)
			}
			return true
		l180:
			position, tokenIndex, depth = position180, tokenIndex180, depth180
			return false
		},
		/* 43 NextExpression <- <(Expression ListExpansion?)> */
		func() bool {
			position184, tokenIndex184, depth184 := position, tokenIndex, depth
			{
				position185 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l184
				}
				{
					position186, tokenIndex186, depth186 := position, tokenIndex, depth
					if !_rules[ruleListExpansion]() {
						goto l186
					}
					goto l187
				l186:
					position, tokenIndex, depth = position186, tokenIndex186, depth186
				}
			l187:
				depth--
				add(ruleNextExpression, position185)
			}
			return true
		l184:
			position, tokenIndex, depth = position184, tokenIndex184, depth184
			return false
		},
		/* 44 ListExpansion <- <('.' '.' '.' ws)> */
		func() bool {
			position188, tokenIndex188, depth188 := position, tokenIndex, depth
			{
				position189 := position
				depth++
				if buffer[position] != rune('.') {
					goto l188
				}
				position++
				if buffer[position] != rune('.') {
					goto l188
				}
				position++
				if buffer[position] != rune('.') {
					goto l188
				}
				position++
				if !_rules[rulews]() {
					goto l188
				}
				depth--
				add(ruleListExpansion, position189)
			}
			return true
		l188:
			position, tokenIndex, depth = position188, tokenIndex188, depth188
			return false
		},
		/* 45 Projection <- <('.'? (('[' '*' ']') / Slice) ProjectionValue ChainedQualifiedExpression*)> */
		func() bool {
			position190, tokenIndex190, depth190 := position, tokenIndex, depth
			{
				position191 := position
				depth++
				{
					position192, tokenIndex192, depth192 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l192
					}
					position++
					goto l193
				l192:
					position, tokenIndex, depth = position192, tokenIndex192, depth192
				}
			l193:
				{
					position194, tokenIndex194, depth194 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l195
					}
					position++
					if buffer[position] != rune('*') {
						goto l195
					}
					position++
					if buffer[position] != rune(']') {
						goto l195
					}
					position++
					goto l194
				l195:
					position, tokenIndex, depth = position194, tokenIndex194, depth194
					if !_rules[ruleSlice]() {
						goto l190
					}
				}
			l194:
				if !_rules[ruleProjectionValue]() {
					goto l190
				}
			l196:
				{
					position197, tokenIndex197, depth197 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l197
					}
					goto l196
				l197:
					position, tokenIndex, depth = position197, tokenIndex197, depth197
				}
				depth--
				add(ruleProjection, position191)
			}
			return true
		l190:
			position, tokenIndex, depth = position190, tokenIndex190, depth190
			return false
		},
		/* 46 ProjectionValue <- <Action0> */
		func() bool {
			position198, tokenIndex198, depth198 := position, tokenIndex, depth
			{
				position199 := position
				depth++
				if !_rules[ruleAction0]() {
					goto l198
				}
				depth--
				add(ruleProjectionValue, position199)
			}
			return true
		l198:
			position, tokenIndex, depth = position198, tokenIndex198, depth198
			return false
		},
		/* 47 Substitution <- <('*' Level0)> */
		func() bool {
			position200, tokenIndex200, depth200 := position, tokenIndex, depth
			{
				position201 := position
				depth++
				if buffer[position] != rune('*') {
					goto l200
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l200
				}
				depth--
				add(ruleSubstitution, position201)
			}
			return true
		l200:
			position, tokenIndex, depth = position200, tokenIndex200, depth200
			return false
		},
		/* 48 Not <- <('!' ws Level0)> */
		func() bool {
			position202, tokenIndex202, depth202 := position, tokenIndex, depth
			{
				position203 := position
				depth++
				if buffer[position] != rune('!') {
					goto l202
				}
				position++
				if !_rules[rulews]() {
					goto l202
				}
				if !_rules[ruleLevel0]() {
					goto l202
				}
				depth--
				add(ruleNot, position203)
			}
			return true
		l202:
			position, tokenIndex, depth = position202, tokenIndex202, depth202
			return false
		},
		/* 49 Grouped <- <('(' Expression ')')> */
		func() bool {
			position204, tokenIndex204, depth204 := position, tokenIndex, depth
			{
				position205 := position
				depth++
				if buffer[position] != rune('(') {
					goto l204
				}
				position++
				if !_rules[ruleExpression]() {
					goto l204
				}
				if buffer[position] != rune(')') {
					goto l204
				}
				position++
				depth--
				add(ruleGrouped, position205)
			}
			return true
		l204:
			position, tokenIndex, depth = position204, tokenIndex204, depth204
			return false
		},
		/* 50 Range <- <(StartRange Expression? RangeOp Expression? ']')> */
		func() bool {
			position206, tokenIndex206, depth206 := position, tokenIndex, depth
			{
				position207 := position
				depth++
				if !_rules[ruleStartRange]() {
					goto l206
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
				if !_rules[ruleRangeOp]() {
					goto l206
				}
				{
					position210, tokenIndex210, depth210 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l210
					}
					goto l211
				l210:
					position, tokenIndex, depth = position210, tokenIndex210, depth210
				}
			l211:
				if buffer[position] != rune(']') {
					goto l206
				}
				position++
				depth--
				add(ruleRange, position207)
			}
			return true
		l206:
			position, tokenIndex, depth = position206, tokenIndex206, depth206
			return false
		},
		/* 51 StartRange <- <'['> */
		func() bool {
			position212, tokenIndex212, depth212 := position, tokenIndex, depth
			{
				position213 := position
				depth++
				if buffer[position] != rune('[') {
					goto l212
				}
				position++
				depth--
				add(ruleStartRange, position213)
			}
			return true
		l212:
			position, tokenIndex, depth = position212, tokenIndex212, depth212
			return false
		},
		/* 52 RangeOp <- <('.' '.')> */
		func() bool {
			position214, tokenIndex214, depth214 := position, tokenIndex, depth
			{
				position215 := position
				depth++
				if buffer[position] != rune('.') {
					goto l214
				}
				position++
				if buffer[position] != rune('.') {
					goto l214
				}
				position++
				depth--
				add(ruleRangeOp, position215)
			}
			return true
		l214:
			position, tokenIndex, depth = position214, tokenIndex214, depth214
			return false
		},
		/* 53 Number <- <('-'? [0-9] ([0-9] / '_')* ('.' [0-9] [0-9]*)? (('e' / 'E') '-'? [0-9] [0-9]*)? !(':' ':'))> */
		func() bool {
			position216, tokenIndex216, depth216 := position, tokenIndex, depth
			{
				position217 := position
				depth++
				{
					position218, tokenIndex218, depth218 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l218
					}
					position++
					goto l219
				l218:
					position, tokenIndex, depth = position218, tokenIndex218, depth218
				}
			l219:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l216
				}
				position++
			l220:
				{
					position221, tokenIndex221, depth221 := position, tokenIndex, depth
					{
						position222, tokenIndex222, depth222 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l223
						}
						position++
						goto l222
					l223:
						position, tokenIndex, depth = position222, tokenIndex222, depth222
						if buffer[position] != rune('_') {
							goto l221
						}
						position++
					}
				l222:
					goto l220
				l221:
					position, tokenIndex, depth = position221, tokenIndex221, depth221
				}
				{
					position224, tokenIndex224, depth224 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l224
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l224
					}
					position++
				l226:
					{
						position227, tokenIndex227, depth227 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l227
						}
						position++
						goto l226
					l227:
						position, tokenIndex, depth = position227, tokenIndex227, depth227
					}
					goto l225
				l224:
					position, tokenIndex, depth = position224, tokenIndex224, depth224
				}
			l225:
				{
					position228, tokenIndex228, depth228 := position, tokenIndex, depth
					{
						position230, tokenIndex230, depth230 := position, tokenIndex, depth
						if buffer[position] != rune('e') {
							goto l231
						}
						position++
						goto l230
					l231:
						position, tokenIndex, depth = position230, tokenIndex230, depth230
						if buffer[position] != rune('E') {
							goto l228
						}
						position++
					}
				l230:
					{
						position232, tokenIndex232, depth232 := position, tokenIndex, depth
						if buffer[position] != rune('-') {
							goto l232
						}
						position++
						goto l233
					l232:
						position, tokenIndex, depth = position232, tokenIndex232, depth232
					}
				l233:
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l228
					}
					position++
				l234:
					{
						position235, tokenIndex235, depth235 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l235
						}
						position++
						goto l234
					l235:
						position, tokenIndex, depth = position235, tokenIndex235, depth235
					}
					goto l229
				l228:
					position, tokenIndex, depth = position228, tokenIndex228, depth228
				}
			l229:
				{
					position236, tokenIndex236, depth236 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l236
					}
					position++
					if buffer[position] != rune(':') {
						goto l236
					}
					position++
					goto l216
				l236:
					position, tokenIndex, depth = position236, tokenIndex236, depth236
				}
				depth--
				add(ruleNumber, position217)
			}
			return true
		l216:
			position, tokenIndex, depth = position216, tokenIndex216, depth216
			return false
		},
		/* 54 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position237, tokenIndex237, depth237 := position, tokenIndex, depth
			{
				position238 := position
				depth++
				if buffer[position] != rune('"') {
					goto l237
				}
				position++
			l239:
				{
					position240, tokenIndex240, depth240 := position, tokenIndex, depth
					{
						position241, tokenIndex241, depth241 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l242
						}
						position++
						if buffer[position] != rune('"') {
							goto l242
						}
						position++
						goto l241
					l242:
						position, tokenIndex, depth = position241, tokenIndex241, depth241
						{
							position243, tokenIndex243, depth243 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l243
							}
							position++
							goto l240
						l243:
							position, tokenIndex, depth = position243, tokenIndex243, depth243
						}
						if !matchDot() {
							goto l240
						}
					}
				l241:
					goto l239
				l240:
					position, tokenIndex, depth = position240, tokenIndex240, depth240
				}
				if buffer[position] != rune('"') {
					goto l237
				}
				position++
				depth--
				add(ruleString, position238)
			}
			return true
		l237:
			position, tokenIndex, depth = position237, tokenIndex237, depth237
			return false
		},
		/* 55 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position244, tokenIndex244, depth244 := position, tokenIndex, depth
			{
				position245 := position
				depth++
				{
					position246, tokenIndex246, depth246 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l247
					}
					position++
					if buffer[position] != rune('r') {
						goto l247
					}
					position++
					if buffer[position] != rune('u') {
						goto l247
					}
					position++
					if buffer[position] != rune('e') {
						goto l247
					}
					position++
					goto l246
				l247:
					position, tokenIndex, depth = position246, tokenIndex246, depth246
					if buffer[position] != rune('f') {
						goto l244
					}
					position++
					if buffer[position] != rune('a') {
						goto l244
					}
					position++
					if buffer[position] != rune('l') {
						goto l244
					}
					position++
					if buffer[position] != rune('s') {
						goto l244
					}
					position++
					if buffer[position] != rune('e') {
						goto l244
					}
					position++
				}
			l246:
				depth--
				add(ruleBoolean, position245)
			}
			return true
		l244:
			position, tokenIndex, depth = position244, tokenIndex244, depth244
			return false
		},
		/* 56 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position248, tokenIndex248, depth248 := position, tokenIndex, depth
			{
				position249 := position
				depth++
				{
					position250, tokenIndex250, depth250 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l251
					}
					position++
					if buffer[position] != rune('i') {
						goto l251
					}
					position++
					if buffer[position] != rune('l') {
						goto l251
					}
					position++
					goto l250
				l251:
					position, tokenIndex, depth = position250, tokenIndex250, depth250
					if buffer[position] != rune('~') {
						goto l248
					}
					position++
				}
			l250:
				depth--
				add(ruleNil, position249)
			}
			return true
		l248:
			position, tokenIndex, depth = position248, tokenIndex248, depth248
			return false
		},
		/* 57 Undefined <- <('~' '~')> */
		func() bool {
			position252, tokenIndex252, depth252 := position, tokenIndex, depth
			{
				position253 := position
				depth++
				if buffer[position] != rune('~') {
					goto l252
				}
				position++
				if buffer[position] != rune('~') {
					goto l252
				}
				position++
				depth--
				add(ruleUndefined, position253)
			}
			return true
		l252:
			position, tokenIndex, depth = position252, tokenIndex252, depth252
			return false
		},
		/* 58 Symbol <- <('$' Name)> */
		func() bool {
			position254, tokenIndex254, depth254 := position, tokenIndex, depth
			{
				position255 := position
				depth++
				if buffer[position] != rune('$') {
					goto l254
				}
				position++
				if !_rules[ruleName]() {
					goto l254
				}
				depth--
				add(ruleSymbol, position255)
			}
			return true
		l254:
			position, tokenIndex, depth = position254, tokenIndex254, depth254
			return false
		},
		/* 59 List <- <(StartList ExpressionList? ']')> */
		func() bool {
			position256, tokenIndex256, depth256 := position, tokenIndex, depth
			{
				position257 := position
				depth++
				if !_rules[ruleStartList]() {
					goto l256
				}
				{
					position258, tokenIndex258, depth258 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l258
					}
					goto l259
				l258:
					position, tokenIndex, depth = position258, tokenIndex258, depth258
				}
			l259:
				if buffer[position] != rune(']') {
					goto l256
				}
				position++
				depth--
				add(ruleList, position257)
			}
			return true
		l256:
			position, tokenIndex, depth = position256, tokenIndex256, depth256
			return false
		},
		/* 60 StartList <- <('[' ws)> */
		func() bool {
			position260, tokenIndex260, depth260 := position, tokenIndex, depth
			{
				position261 := position
				depth++
				if buffer[position] != rune('[') {
					goto l260
				}
				position++
				if !_rules[rulews]() {
					goto l260
				}
				depth--
				add(ruleStartList, position261)
			}
			return true
		l260:
			position, tokenIndex, depth = position260, tokenIndex260, depth260
			return false
		},
		/* 61 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position262, tokenIndex262, depth262 := position, tokenIndex, depth
			{
				position263 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l262
				}
				if !_rules[rulews]() {
					goto l262
				}
				{
					position264, tokenIndex264, depth264 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l264
					}
					goto l265
				l264:
					position, tokenIndex, depth = position264, tokenIndex264, depth264
				}
			l265:
				if buffer[position] != rune('}') {
					goto l262
				}
				position++
				depth--
				add(ruleMap, position263)
			}
			return true
		l262:
			position, tokenIndex, depth = position262, tokenIndex262, depth262
			return false
		},
		/* 62 CreateMap <- <'{'> */
		func() bool {
			position266, tokenIndex266, depth266 := position, tokenIndex, depth
			{
				position267 := position
				depth++
				if buffer[position] != rune('{') {
					goto l266
				}
				position++
				depth--
				add(ruleCreateMap, position267)
			}
			return true
		l266:
			position, tokenIndex, depth = position266, tokenIndex266, depth266
			return false
		},
		/* 63 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position268, tokenIndex268, depth268 := position, tokenIndex, depth
			{
				position269 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l268
				}
			l270:
				{
					position271, tokenIndex271, depth271 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l271
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l271
					}
					goto l270
				l271:
					position, tokenIndex, depth = position271, tokenIndex271, depth271
				}
				depth--
				add(ruleAssignments, position269)
			}
			return true
		l268:
			position, tokenIndex, depth = position268, tokenIndex268, depth268
			return false
		},
		/* 64 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position272, tokenIndex272, depth272 := position, tokenIndex, depth
			{
				position273 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l272
				}
				if buffer[position] != rune('=') {
					goto l272
				}
				position++
				if !_rules[ruleExpression]() {
					goto l272
				}
				depth--
				add(ruleAssignment, position273)
			}
			return true
		l272:
			position, tokenIndex, depth = position272, tokenIndex272, depth272
			return false
		},
		/* 65 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position274, tokenIndex274, depth274 := position, tokenIndex, depth
			{
				position275 := position
				depth++
				{
					position276, tokenIndex276, depth276 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l277
					}
					goto l276
				l277:
					position, tokenIndex, depth = position276, tokenIndex276, depth276
					if !_rules[ruleSimpleMerge]() {
						goto l274
					}
				}
			l276:
				depth--
				add(ruleMerge, position275)
			}
			return true
		l274:
			position, tokenIndex, depth = position274, tokenIndex274, depth274
			return false
		},
		/* 66 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position278, tokenIndex278, depth278 := position, tokenIndex, depth
			{
				position279 := position
				depth++
				if buffer[position] != rune('m') {
					goto l278
				}
				position++
				if buffer[position] != rune('e') {
					goto l278
				}
				position++
				if buffer[position] != rune('r') {
					goto l278
				}
				position++
				if buffer[position] != rune('g') {
					goto l278
				}
				position++
				if buffer[position] != rune('e') {
					goto l278
				}
				position++
				{
					position280, tokenIndex280, depth280 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l280
					}
					if !_rules[ruleRequired]() {
						goto l280
					}
					goto l278
				l280:
					position, tokenIndex, depth = position280, tokenIndex280, depth280
				}
				{
					position281, tokenIndex281, depth281 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l281
					}
					{
						position283, tokenIndex283, depth283 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l284
						}
						goto l283
					l284:
						position, tokenIndex, depth = position283, tokenIndex283, depth283
						if !_rules[ruleOn]() {
							goto l281
						}
					}
				l283:
					goto l282
				l281:
					position, tokenIndex, depth = position281, tokenIndex281, depth281
				}
			l282:
				if !_rules[rulereq_ws]() {
					goto l278
				}
				if !_rules[ruleReference]() {
					goto l278
				}
				depth--
				add(ruleRefMerge, position279)
			}
			return true
		l278:
			position, tokenIndex, depth = position278, tokenIndex278, depth278
			return false
		},
		/* 67 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' !'(' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position285, tokenIndex285, depth285 := position, tokenIndex, depth
			{
				position286 := position
				depth++
				if buffer[position] != rune('m') {
					goto l285
				}
				position++
				if buffer[position] != rune('e') {
					goto l285
				}
				position++
				if buffer[position] != rune('r') {
					goto l285
				}
				position++
				if buffer[position] != rune('g') {
					goto l285
				}
				position++
				if buffer[position] != rune('e') {
					goto l285
				}
				position++
				{
					position287, tokenIndex287, depth287 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l287
					}
					position++
					goto l285
				l287:
					position, tokenIndex, depth = position287, tokenIndex287, depth287
				}
				{
					position288, tokenIndex288, depth288 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l288
					}
					{
						position290, tokenIndex290, depth290 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l291
						}
						goto l290
					l291:
						position, tokenIndex, depth = position290, tokenIndex290, depth290
						if !_rules[ruleRequired]() {
							goto l292
						}
						goto l290
					l292:
						position, tokenIndex, depth = position290, tokenIndex290, depth290
						if !_rules[ruleOn]() {
							goto l288
						}
					}
				l290:
					goto l289
				l288:
					position, tokenIndex, depth = position288, tokenIndex288, depth288
				}
			l289:
				depth--
				add(ruleSimpleMerge, position286)
			}
			return true
		l285:
			position, tokenIndex, depth = position285, tokenIndex285, depth285
			return false
		},
		/* 68 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
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
				if buffer[position] != rune('p') {
					goto l293
				}
				position++
				if buffer[position] != rune('l') {
					goto l293
				}
				position++
				if buffer[position] != rune('a') {
					goto l293
				}
				position++
				if buffer[position] != rune('c') {
					goto l293
				}
				position++
				if buffer[position] != rune('e') {
					goto l293
				}
				position++
				depth--
				add(ruleReplace, position294)
			}
			return true
		l293:
			position, tokenIndex, depth = position293, tokenIndex293, depth293
			return false
		},
		/* 69 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position295, tokenIndex295, depth295 := position, tokenIndex, depth
			{
				position296 := position
				depth++
				if buffer[position] != rune('r') {
					goto l295
				}
				position++
				if buffer[position] != rune('e') {
					goto l295
				}
				position++
				if buffer[position] != rune('q') {
					goto l295
				}
				position++
				if buffer[position] != rune('u') {
					goto l295
				}
				position++
				if buffer[position] != rune('i') {
					goto l295
				}
				position++
				if buffer[position] != rune('r') {
					goto l295
				}
				position++
				if buffer[position] != rune('e') {
					goto l295
				}
				position++
				if buffer[position] != rune('d') {
					goto l295
				}
				position++
				depth--
				add(ruleRequired, position296)
			}
			return true
		l295:
			position, tokenIndex, depth = position295, tokenIndex295, depth295
			return false
		},
		/* 70 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position297, tokenIndex297, depth297 := position, tokenIndex, depth
			{
				position298 := position
				depth++
				if buffer[position] != rune('o') {
					goto l297
				}
				position++
				if buffer[position] != rune('n') {
					goto l297
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l297
				}
				if !_rules[ruleName]() {
					goto l297
				}
				depth--
				add(ruleOn, position298)
			}
			return true
		l297:
			position, tokenIndex, depth = position297, tokenIndex297, depth297
			return false
		},
		/* 71 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position299, tokenIndex299, depth299 := position, tokenIndex, depth
			{
				position300 := position
				depth++
				if buffer[position] != rune('a') {
					goto l299
				}
				position++
				if buffer[position] != rune('u') {
					goto l299
				}
				position++
				if buffer[position] != rune('t') {
					goto l299
				}
				position++
				if buffer[position] != rune('o') {
					goto l299
				}
				position++
				depth--
				add(ruleAuto, position300)
			}
			return true
		l299:
			position, tokenIndex, depth = position299, tokenIndex299, depth299
			return false
		},
		/* 72 Default <- <Action1> */
		func() bool {
			position301, tokenIndex301, depth301 := position, tokenIndex, depth
			{
				position302 := position
				depth++
				if !_rules[ruleAction1]() {
					goto l301
				}
				depth--
				add(ruleDefault, position302)
			}
			return true
		l301:
			position, tokenIndex, depth = position301, tokenIndex301, depth301
			return false
		},
		/* 73 Sync <- <('s' 'y' 'n' 'c' '[' Level7 ((((LambdaExpr LambdaExt) / (LambdaOrExpr LambdaOrExpr)) (('|' Expression) / Default)) / (LambdaOrExpr Default Default)) ']')> */
		func() bool {
			position303, tokenIndex303, depth303 := position, tokenIndex, depth
			{
				position304 := position
				depth++
				if buffer[position] != rune('s') {
					goto l303
				}
				position++
				if buffer[position] != rune('y') {
					goto l303
				}
				position++
				if buffer[position] != rune('n') {
					goto l303
				}
				position++
				if buffer[position] != rune('c') {
					goto l303
				}
				position++
				if buffer[position] != rune('[') {
					goto l303
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l303
				}
				{
					position305, tokenIndex305, depth305 := position, tokenIndex, depth
					{
						position307, tokenIndex307, depth307 := position, tokenIndex, depth
						if !_rules[ruleLambdaExpr]() {
							goto l308
						}
						if !_rules[ruleLambdaExt]() {
							goto l308
						}
						goto l307
					l308:
						position, tokenIndex, depth = position307, tokenIndex307, depth307
						if !_rules[ruleLambdaOrExpr]() {
							goto l306
						}
						if !_rules[ruleLambdaOrExpr]() {
							goto l306
						}
					}
				l307:
					{
						position309, tokenIndex309, depth309 := position, tokenIndex, depth
						if buffer[position] != rune('|') {
							goto l310
						}
						position++
						if !_rules[ruleExpression]() {
							goto l310
						}
						goto l309
					l310:
						position, tokenIndex, depth = position309, tokenIndex309, depth309
						if !_rules[ruleDefault]() {
							goto l306
						}
					}
				l309:
					goto l305
				l306:
					position, tokenIndex, depth = position305, tokenIndex305, depth305
					if !_rules[ruleLambdaOrExpr]() {
						goto l303
					}
					if !_rules[ruleDefault]() {
						goto l303
					}
					if !_rules[ruleDefault]() {
						goto l303
					}
				}
			l305:
				if buffer[position] != rune(']') {
					goto l303
				}
				position++
				depth--
				add(ruleSync, position304)
			}
			return true
		l303:
			position, tokenIndex, depth = position303, tokenIndex303, depth303
			return false
		},
		/* 74 LambdaExt <- <(',' Expression)> */
		func() bool {
			position311, tokenIndex311, depth311 := position, tokenIndex, depth
			{
				position312 := position
				depth++
				if buffer[position] != rune(',') {
					goto l311
				}
				position++
				if !_rules[ruleExpression]() {
					goto l311
				}
				depth--
				add(ruleLambdaExt, position312)
			}
			return true
		l311:
			position, tokenIndex, depth = position311, tokenIndex311, depth311
			return false
		},
		/* 75 LambdaOrExpr <- <(LambdaExpr / ('|' Expression))> */
		func() bool {
			position313, tokenIndex313, depth313 := position, tokenIndex, depth
			{
				position314 := position
				depth++
				{
					position315, tokenIndex315, depth315 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l316
					}
					goto l315
				l316:
					position, tokenIndex, depth = position315, tokenIndex315, depth315
					if buffer[position] != rune('|') {
						goto l313
					}
					position++
					if !_rules[ruleExpression]() {
						goto l313
					}
				}
			l315:
				depth--
				add(ruleLambdaOrExpr, position314)
			}
			return true
		l313:
			position, tokenIndex, depth = position313, tokenIndex313, depth313
			return false
		},
		/* 76 Catch <- <('c' 'a' 't' 'c' 'h' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position317, tokenIndex317, depth317 := position, tokenIndex, depth
			{
				position318 := position
				depth++
				if buffer[position] != rune('c') {
					goto l317
				}
				position++
				if buffer[position] != rune('a') {
					goto l317
				}
				position++
				if buffer[position] != rune('t') {
					goto l317
				}
				position++
				if buffer[position] != rune('c') {
					goto l317
				}
				position++
				if buffer[position] != rune('h') {
					goto l317
				}
				position++
				if buffer[position] != rune('[') {
					goto l317
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l317
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l317
				}
				if buffer[position] != rune(']') {
					goto l317
				}
				position++
				depth--
				add(ruleCatch, position318)
			}
			return true
		l317:
			position, tokenIndex, depth = position317, tokenIndex317, depth317
			return false
		},
		/* 77 MapMapping <- <('m' 'a' 'p' '{' Level7 LambdaOrExpr '}')> */
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
				if buffer[position] != rune('{') {
					goto l319
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l319
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l319
				}
				if buffer[position] != rune('}') {
					goto l319
				}
				position++
				depth--
				add(ruleMapMapping, position320)
			}
			return true
		l319:
			position, tokenIndex, depth = position319, tokenIndex319, depth319
			return false
		},
		/* 78 Mapping <- <('m' 'a' 'p' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position321, tokenIndex321, depth321 := position, tokenIndex, depth
			{
				position322 := position
				depth++
				if buffer[position] != rune('m') {
					goto l321
				}
				position++
				if buffer[position] != rune('a') {
					goto l321
				}
				position++
				if buffer[position] != rune('p') {
					goto l321
				}
				position++
				if buffer[position] != rune('[') {
					goto l321
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l321
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l321
				}
				if buffer[position] != rune(']') {
					goto l321
				}
				position++
				depth--
				add(ruleMapping, position322)
			}
			return true
		l321:
			position, tokenIndex, depth = position321, tokenIndex321, depth321
			return false
		},
		/* 79 MapSelection <- <('s' 'e' 'l' 'e' 'c' 't' '{' Level7 LambdaOrExpr '}')> */
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
				if buffer[position] != rune('{') {
					goto l323
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l323
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l323
				}
				if buffer[position] != rune('}') {
					goto l323
				}
				position++
				depth--
				add(ruleMapSelection, position324)
			}
			return true
		l323:
			position, tokenIndex, depth = position323, tokenIndex323, depth323
			return false
		},
		/* 80 Selection <- <('s' 'e' 'l' 'e' 'c' 't' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position325, tokenIndex325, depth325 := position, tokenIndex, depth
			{
				position326 := position
				depth++
				if buffer[position] != rune('s') {
					goto l325
				}
				position++
				if buffer[position] != rune('e') {
					goto l325
				}
				position++
				if buffer[position] != rune('l') {
					goto l325
				}
				position++
				if buffer[position] != rune('e') {
					goto l325
				}
				position++
				if buffer[position] != rune('c') {
					goto l325
				}
				position++
				if buffer[position] != rune('t') {
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
				add(ruleSelection, position326)
			}
			return true
		l325:
			position, tokenIndex, depth = position325, tokenIndex325, depth325
			return false
		},
		/* 81 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 LambdaOrExpr ']')> */
		func() bool {
			position327, tokenIndex327, depth327 := position, tokenIndex, depth
			{
				position328 := position
				depth++
				if buffer[position] != rune('s') {
					goto l327
				}
				position++
				if buffer[position] != rune('u') {
					goto l327
				}
				position++
				if buffer[position] != rune('m') {
					goto l327
				}
				position++
				if buffer[position] != rune('[') {
					goto l327
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l327
				}
				if buffer[position] != rune('|') {
					goto l327
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l327
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l327
				}
				if buffer[position] != rune(']') {
					goto l327
				}
				position++
				depth--
				add(ruleSum, position328)
			}
			return true
		l327:
			position, tokenIndex, depth = position327, tokenIndex327, depth327
			return false
		},
		/* 82 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position329, tokenIndex329, depth329 := position, tokenIndex, depth
			{
				position330 := position
				depth++
				if buffer[position] != rune('l') {
					goto l329
				}
				position++
				if buffer[position] != rune('a') {
					goto l329
				}
				position++
				if buffer[position] != rune('m') {
					goto l329
				}
				position++
				if buffer[position] != rune('b') {
					goto l329
				}
				position++
				if buffer[position] != rune('d') {
					goto l329
				}
				position++
				if buffer[position] != rune('a') {
					goto l329
				}
				position++
				{
					position331, tokenIndex331, depth331 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l332
					}
					goto l331
				l332:
					position, tokenIndex, depth = position331, tokenIndex331, depth331
					if !_rules[ruleLambdaExpr]() {
						goto l329
					}
				}
			l331:
				depth--
				add(ruleLambda, position330)
			}
			return true
		l329:
			position, tokenIndex, depth = position329, tokenIndex329, depth329
			return false
		},
		/* 83 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position333, tokenIndex333, depth333 := position, tokenIndex, depth
			{
				position334 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l333
				}
				if !_rules[ruleExpression]() {
					goto l333
				}
				depth--
				add(ruleLambdaRef, position334)
			}
			return true
		l333:
			position, tokenIndex, depth = position333, tokenIndex333, depth333
			return false
		},
		/* 84 LambdaExpr <- <(ws Params ws ('-' '>') Expression)> */
		func() bool {
			position335, tokenIndex335, depth335 := position, tokenIndex, depth
			{
				position336 := position
				depth++
				if !_rules[rulews]() {
					goto l335
				}
				if !_rules[ruleParams]() {
					goto l335
				}
				if !_rules[rulews]() {
					goto l335
				}
				if buffer[position] != rune('-') {
					goto l335
				}
				position++
				if buffer[position] != rune('>') {
					goto l335
				}
				position++
				if !_rules[ruleExpression]() {
					goto l335
				}
				depth--
				add(ruleLambdaExpr, position336)
			}
			return true
		l335:
			position, tokenIndex, depth = position335, tokenIndex335, depth335
			return false
		},
		/* 85 Params <- <('|' StartParams ws Names? '|')> */
		func() bool {
			position337, tokenIndex337, depth337 := position, tokenIndex, depth
			{
				position338 := position
				depth++
				if buffer[position] != rune('|') {
					goto l337
				}
				position++
				if !_rules[ruleStartParams]() {
					goto l337
				}
				if !_rules[rulews]() {
					goto l337
				}
				{
					position339, tokenIndex339, depth339 := position, tokenIndex, depth
					if !_rules[ruleNames]() {
						goto l339
					}
					goto l340
				l339:
					position, tokenIndex, depth = position339, tokenIndex339, depth339
				}
			l340:
				if buffer[position] != rune('|') {
					goto l337
				}
				position++
				depth--
				add(ruleParams, position338)
			}
			return true
		l337:
			position, tokenIndex, depth = position337, tokenIndex337, depth337
			return false
		},
		/* 86 StartParams <- <Action2> */
		func() bool {
			position341, tokenIndex341, depth341 := position, tokenIndex, depth
			{
				position342 := position
				depth++
				if !_rules[ruleAction2]() {
					goto l341
				}
				depth--
				add(ruleStartParams, position342)
			}
			return true
		l341:
			position, tokenIndex, depth = position341, tokenIndex341, depth341
			return false
		},
		/* 87 Names <- <(NextName (',' NextName)* DefaultValue? (',' NextName DefaultValue)* VarParams?)> */
		func() bool {
			position343, tokenIndex343, depth343 := position, tokenIndex, depth
			{
				position344 := position
				depth++
				if !_rules[ruleNextName]() {
					goto l343
				}
			l345:
				{
					position346, tokenIndex346, depth346 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l346
					}
					position++
					if !_rules[ruleNextName]() {
						goto l346
					}
					goto l345
				l346:
					position, tokenIndex, depth = position346, tokenIndex346, depth346
				}
				{
					position347, tokenIndex347, depth347 := position, tokenIndex, depth
					if !_rules[ruleDefaultValue]() {
						goto l347
					}
					goto l348
				l347:
					position, tokenIndex, depth = position347, tokenIndex347, depth347
				}
			l348:
			l349:
				{
					position350, tokenIndex350, depth350 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l350
					}
					position++
					if !_rules[ruleNextName]() {
						goto l350
					}
					if !_rules[ruleDefaultValue]() {
						goto l350
					}
					goto l349
				l350:
					position, tokenIndex, depth = position350, tokenIndex350, depth350
				}
				{
					position351, tokenIndex351, depth351 := position, tokenIndex, depth
					if !_rules[ruleVarParams]() {
						goto l351
					}
					goto l352
				l351:
					position, tokenIndex, depth = position351, tokenIndex351, depth351
				}
			l352:
				depth--
				add(ruleNames, position344)
			}
			return true
		l343:
			position, tokenIndex, depth = position343, tokenIndex343, depth343
			return false
		},
		/* 88 NextName <- <(ws Name ws)> */
		func() bool {
			position353, tokenIndex353, depth353 := position, tokenIndex, depth
			{
				position354 := position
				depth++
				if !_rules[rulews]() {
					goto l353
				}
				if !_rules[ruleName]() {
					goto l353
				}
				if !_rules[rulews]() {
					goto l353
				}
				depth--
				add(ruleNextName, position354)
			}
			return true
		l353:
			position, tokenIndex, depth = position353, tokenIndex353, depth353
			return false
		},
		/* 89 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position355, tokenIndex355, depth355 := position, tokenIndex, depth
			{
				position356 := position
				depth++
				{
					position359, tokenIndex359, depth359 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l360
					}
					position++
					goto l359
				l360:
					position, tokenIndex, depth = position359, tokenIndex359, depth359
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l361
					}
					position++
					goto l359
				l361:
					position, tokenIndex, depth = position359, tokenIndex359, depth359
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l362
					}
					position++
					goto l359
				l362:
					position, tokenIndex, depth = position359, tokenIndex359, depth359
					if buffer[position] != rune('_') {
						goto l355
					}
					position++
				}
			l359:
			l357:
				{
					position358, tokenIndex358, depth358 := position, tokenIndex, depth
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
							goto l358
						}
						position++
					}
				l363:
					goto l357
				l358:
					position, tokenIndex, depth = position358, tokenIndex358, depth358
				}
				depth--
				add(ruleName, position356)
			}
			return true
		l355:
			position, tokenIndex, depth = position355, tokenIndex355, depth355
			return false
		},
		/* 90 DefaultValue <- <('=' Expression)> */
		func() bool {
			position367, tokenIndex367, depth367 := position, tokenIndex, depth
			{
				position368 := position
				depth++
				if buffer[position] != rune('=') {
					goto l367
				}
				position++
				if !_rules[ruleExpression]() {
					goto l367
				}
				depth--
				add(ruleDefaultValue, position368)
			}
			return true
		l367:
			position, tokenIndex, depth = position367, tokenIndex367, depth367
			return false
		},
		/* 91 VarParams <- <('.' '.' '.' ws)> */
		func() bool {
			position369, tokenIndex369, depth369 := position, tokenIndex, depth
			{
				position370 := position
				depth++
				if buffer[position] != rune('.') {
					goto l369
				}
				position++
				if buffer[position] != rune('.') {
					goto l369
				}
				position++
				if buffer[position] != rune('.') {
					goto l369
				}
				position++
				if !_rules[rulews]() {
					goto l369
				}
				depth--
				add(ruleVarParams, position370)
			}
			return true
		l369:
			position, tokenIndex, depth = position369, tokenIndex369, depth369
			return false
		},
		/* 92 Reference <- <(((Tag ('.' / Key)) / ('.'? Key)) FollowUpRef)> */
		func() bool {
			position371, tokenIndex371, depth371 := position, tokenIndex, depth
			{
				position372 := position
				depth++
				{
					position373, tokenIndex373, depth373 := position, tokenIndex, depth
					if !_rules[ruleTag]() {
						goto l374
					}
					{
						position375, tokenIndex375, depth375 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l376
						}
						position++
						goto l375
					l376:
						position, tokenIndex, depth = position375, tokenIndex375, depth375
						if !_rules[ruleKey]() {
							goto l374
						}
					}
				l375:
					goto l373
				l374:
					position, tokenIndex, depth = position373, tokenIndex373, depth373
					{
						position377, tokenIndex377, depth377 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l377
						}
						position++
						goto l378
					l377:
						position, tokenIndex, depth = position377, tokenIndex377, depth377
					}
				l378:
					if !_rules[ruleKey]() {
						goto l371
					}
				}
			l373:
				if !_rules[ruleFollowUpRef]() {
					goto l371
				}
				depth--
				add(ruleReference, position372)
			}
			return true
		l371:
			position, tokenIndex, depth = position371, tokenIndex371, depth371
			return false
		},
		/* 93 Tag <- <((('d' 'o' 'c' ':' '-'? [0-9]+) / (TagName (':' TagName)*)) (':' ':'))> */
		func() bool {
			position379, tokenIndex379, depth379 := position, tokenIndex, depth
			{
				position380 := position
				depth++
				{
					position381, tokenIndex381, depth381 := position, tokenIndex, depth
					if buffer[position] != rune('d') {
						goto l382
					}
					position++
					if buffer[position] != rune('o') {
						goto l382
					}
					position++
					if buffer[position] != rune('c') {
						goto l382
					}
					position++
					if buffer[position] != rune(':') {
						goto l382
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
						goto l382
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
					goto l381
				l382:
					position, tokenIndex, depth = position381, tokenIndex381, depth381
					if !_rules[ruleTagName]() {
						goto l379
					}
				l387:
					{
						position388, tokenIndex388, depth388 := position, tokenIndex, depth
						if buffer[position] != rune(':') {
							goto l388
						}
						position++
						if !_rules[ruleTagName]() {
							goto l388
						}
						goto l387
					l388:
						position, tokenIndex, depth = position388, tokenIndex388, depth388
					}
				}
			l381:
				if buffer[position] != rune(':') {
					goto l379
				}
				position++
				if buffer[position] != rune(':') {
					goto l379
				}
				position++
				depth--
				add(ruleTag, position380)
			}
			return true
		l379:
			position, tokenIndex, depth = position379, tokenIndex379, depth379
			return false
		},
		/* 94 TagName <- <(([a-z] / [A-Z] / '_') ([a-z] / [A-Z] / [0-9] / '_')*)> */
		func() bool {
			position389, tokenIndex389, depth389 := position, tokenIndex, depth
			{
				position390 := position
				depth++
				{
					position391, tokenIndex391, depth391 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l392
					}
					position++
					goto l391
				l392:
					position, tokenIndex, depth = position391, tokenIndex391, depth391
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l393
					}
					position++
					goto l391
				l393:
					position, tokenIndex, depth = position391, tokenIndex391, depth391
					if buffer[position] != rune('_') {
						goto l389
					}
					position++
				}
			l391:
			l394:
				{
					position395, tokenIndex395, depth395 := position, tokenIndex, depth
					{
						position396, tokenIndex396, depth396 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l397
						}
						position++
						goto l396
					l397:
						position, tokenIndex, depth = position396, tokenIndex396, depth396
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l398
						}
						position++
						goto l396
					l398:
						position, tokenIndex, depth = position396, tokenIndex396, depth396
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l399
						}
						position++
						goto l396
					l399:
						position, tokenIndex, depth = position396, tokenIndex396, depth396
						if buffer[position] != rune('_') {
							goto l395
						}
						position++
					}
				l396:
					goto l394
				l395:
					position, tokenIndex, depth = position395, tokenIndex395, depth395
				}
				depth--
				add(ruleTagName, position390)
			}
			return true
		l389:
			position, tokenIndex, depth = position389, tokenIndex389, depth389
			return false
		},
		/* 95 FollowUpRef <- <PathComponent*> */
		func() bool {
			{
				position401 := position
				depth++
			l402:
				{
					position403, tokenIndex403, depth403 := position, tokenIndex, depth
					if !_rules[rulePathComponent]() {
						goto l403
					}
					goto l402
				l403:
					position, tokenIndex, depth = position403, tokenIndex403, depth403
				}
				depth--
				add(ruleFollowUpRef, position401)
			}
			return true
		},
		/* 96 PathComponent <- <(('.' Key) / ('.'? Index))> */
		func() bool {
			position404, tokenIndex404, depth404 := position, tokenIndex, depth
			{
				position405 := position
				depth++
				{
					position406, tokenIndex406, depth406 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l407
					}
					position++
					if !_rules[ruleKey]() {
						goto l407
					}
					goto l406
				l407:
					position, tokenIndex, depth = position406, tokenIndex406, depth406
					{
						position408, tokenIndex408, depth408 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l408
						}
						position++
						goto l409
					l408:
						position, tokenIndex, depth = position408, tokenIndex408, depth408
					}
				l409:
					if !_rules[ruleIndex]() {
						goto l404
					}
				}
			l406:
				depth--
				add(rulePathComponent, position405)
			}
			return true
		l404:
			position, tokenIndex, depth = position404, tokenIndex404, depth404
			return false
		},
		/* 97 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position410, tokenIndex410, depth410 := position, tokenIndex, depth
			{
				position411 := position
				depth++
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
						goto l410
					}
					position++
				}
			l412:
			l416:
				{
					position417, tokenIndex417, depth417 := position, tokenIndex, depth
					{
						position418, tokenIndex418, depth418 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l419
						}
						position++
						goto l418
					l419:
						position, tokenIndex, depth = position418, tokenIndex418, depth418
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l420
						}
						position++
						goto l418
					l420:
						position, tokenIndex, depth = position418, tokenIndex418, depth418
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l421
						}
						position++
						goto l418
					l421:
						position, tokenIndex, depth = position418, tokenIndex418, depth418
						if buffer[position] != rune('_') {
							goto l422
						}
						position++
						goto l418
					l422:
						position, tokenIndex, depth = position418, tokenIndex418, depth418
						if buffer[position] != rune('-') {
							goto l417
						}
						position++
					}
				l418:
					goto l416
				l417:
					position, tokenIndex, depth = position417, tokenIndex417, depth417
				}
				{
					position423, tokenIndex423, depth423 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l423
					}
					position++
					{
						position425, tokenIndex425, depth425 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l426
						}
						position++
						goto l425
					l426:
						position, tokenIndex, depth = position425, tokenIndex425, depth425
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l427
						}
						position++
						goto l425
					l427:
						position, tokenIndex, depth = position425, tokenIndex425, depth425
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l428
						}
						position++
						goto l425
					l428:
						position, tokenIndex, depth = position425, tokenIndex425, depth425
						if buffer[position] != rune('_') {
							goto l423
						}
						position++
					}
				l425:
				l429:
					{
						position430, tokenIndex430, depth430 := position, tokenIndex, depth
						{
							position431, tokenIndex431, depth431 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l432
							}
							position++
							goto l431
						l432:
							position, tokenIndex, depth = position431, tokenIndex431, depth431
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l433
							}
							position++
							goto l431
						l433:
							position, tokenIndex, depth = position431, tokenIndex431, depth431
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l434
							}
							position++
							goto l431
						l434:
							position, tokenIndex, depth = position431, tokenIndex431, depth431
							if buffer[position] != rune('_') {
								goto l435
							}
							position++
							goto l431
						l435:
							position, tokenIndex, depth = position431, tokenIndex431, depth431
							if buffer[position] != rune('-') {
								goto l430
							}
							position++
						}
					l431:
						goto l429
					l430:
						position, tokenIndex, depth = position430, tokenIndex430, depth430
					}
					goto l424
				l423:
					position, tokenIndex, depth = position423, tokenIndex423, depth423
				}
			l424:
				depth--
				add(ruleKey, position411)
			}
			return true
		l410:
			position, tokenIndex, depth = position410, tokenIndex410, depth410
			return false
		},
		/* 98 Index <- <('[' '-'? [0-9]+ ']')> */
		func() bool {
			position436, tokenIndex436, depth436 := position, tokenIndex, depth
			{
				position437 := position
				depth++
				if buffer[position] != rune('[') {
					goto l436
				}
				position++
				{
					position438, tokenIndex438, depth438 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l438
					}
					position++
					goto l439
				l438:
					position, tokenIndex, depth = position438, tokenIndex438, depth438
				}
			l439:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l436
				}
				position++
			l440:
				{
					position441, tokenIndex441, depth441 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l441
					}
					position++
					goto l440
				l441:
					position, tokenIndex, depth = position441, tokenIndex441, depth441
				}
				if buffer[position] != rune(']') {
					goto l436
				}
				position++
				depth--
				add(ruleIndex, position437)
			}
			return true
		l436:
			position, tokenIndex, depth = position436, tokenIndex436, depth436
			return false
		},
		/* 99 IP <- <([0-9]+ '.' [0-9]+ '.' [0-9]+ '.' [0-9]+)> */
		func() bool {
			position442, tokenIndex442, depth442 := position, tokenIndex, depth
			{
				position443 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l442
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
				if buffer[position] != rune('.') {
					goto l442
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l442
				}
				position++
			l446:
				{
					position447, tokenIndex447, depth447 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l447
					}
					position++
					goto l446
				l447:
					position, tokenIndex, depth = position447, tokenIndex447, depth447
				}
				if buffer[position] != rune('.') {
					goto l442
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l442
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
					goto l442
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l442
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
				depth--
				add(ruleIP, position443)
			}
			return true
		l442:
			position, tokenIndex, depth = position442, tokenIndex442, depth442
			return false
		},
		/* 100 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position453 := position
				depth++
			l454:
				{
					position455, tokenIndex455, depth455 := position, tokenIndex, depth
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
							goto l455
						}
						position++
					}
				l456:
					goto l454
				l455:
					position, tokenIndex, depth = position455, tokenIndex455, depth455
				}
				depth--
				add(rulews, position453)
			}
			return true
		},
		/* 101 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position460, tokenIndex460, depth460 := position, tokenIndex, depth
			{
				position461 := position
				depth++
				{
					position464, tokenIndex464, depth464 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l465
					}
					position++
					goto l464
				l465:
					position, tokenIndex, depth = position464, tokenIndex464, depth464
					if buffer[position] != rune('\t') {
						goto l466
					}
					position++
					goto l464
				l466:
					position, tokenIndex, depth = position464, tokenIndex464, depth464
					if buffer[position] != rune('\n') {
						goto l467
					}
					position++
					goto l464
				l467:
					position, tokenIndex, depth = position464, tokenIndex464, depth464
					if buffer[position] != rune('\r') {
						goto l460
					}
					position++
				}
			l464:
			l462:
				{
					position463, tokenIndex463, depth463 := position, tokenIndex, depth
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
							goto l463
						}
						position++
					}
				l468:
					goto l462
				l463:
					position, tokenIndex, depth = position463, tokenIndex463, depth463
				}
				depth--
				add(rulereq_ws, position461)
			}
			return true
		l460:
			position, tokenIndex, depth = position460, tokenIndex460, depth460
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
