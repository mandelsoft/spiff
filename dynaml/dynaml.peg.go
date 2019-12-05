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
	rules  [103]func() bool
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
		/* 37 ChainedCall <- <(StartArguments NameArgumentList? ')')> */
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
					if !_rules[ruleNameArgumentList]() {
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
		/* 39 NameArgumentList <- <(((NextNameArgument (',' NextNameArgument)*) / NextExpression) (',' NextExpression)*)> */
		func() bool {
			position163, tokenIndex163, depth163 := position, tokenIndex, depth
			{
				position164 := position
				depth++
				{
					position165, tokenIndex165, depth165 := position, tokenIndex, depth
					if !_rules[ruleNextNameArgument]() {
						goto l166
					}
				l167:
					{
						position168, tokenIndex168, depth168 := position, tokenIndex, depth
						if buffer[position] != rune(',') {
							goto l168
						}
						position++
						if !_rules[ruleNextNameArgument]() {
							goto l168
						}
						goto l167
					l168:
						position, tokenIndex, depth = position168, tokenIndex168, depth168
					}
					goto l165
				l166:
					position, tokenIndex, depth = position165, tokenIndex165, depth165
					if !_rules[ruleNextExpression]() {
						goto l163
					}
				}
			l165:
			l169:
				{
					position170, tokenIndex170, depth170 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l170
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l170
					}
					goto l169
				l170:
					position, tokenIndex, depth = position170, tokenIndex170, depth170
				}
				depth--
				add(ruleNameArgumentList, position164)
			}
			return true
		l163:
			position, tokenIndex, depth = position163, tokenIndex163, depth163
			return false
		},
		/* 40 NextNameArgument <- <(ws Name ws '=' ws Expression ws)> */
		func() bool {
			position171, tokenIndex171, depth171 := position, tokenIndex, depth
			{
				position172 := position
				depth++
				if !_rules[rulews]() {
					goto l171
				}
				if !_rules[ruleName]() {
					goto l171
				}
				if !_rules[rulews]() {
					goto l171
				}
				if buffer[position] != rune('=') {
					goto l171
				}
				position++
				if !_rules[rulews]() {
					goto l171
				}
				if !_rules[ruleExpression]() {
					goto l171
				}
				if !_rules[rulews]() {
					goto l171
				}
				depth--
				add(ruleNextNameArgument, position172)
			}
			return true
		l171:
			position, tokenIndex, depth = position171, tokenIndex171, depth171
			return false
		},
		/* 41 ExpressionList <- <(NextExpression (',' NextExpression)*)> */
		func() bool {
			position173, tokenIndex173, depth173 := position, tokenIndex, depth
			{
				position174 := position
				depth++
				if !_rules[ruleNextExpression]() {
					goto l173
				}
			l175:
				{
					position176, tokenIndex176, depth176 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l176
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l176
					}
					goto l175
				l176:
					position, tokenIndex, depth = position176, tokenIndex176, depth176
				}
				depth--
				add(ruleExpressionList, position174)
			}
			return true
		l173:
			position, tokenIndex, depth = position173, tokenIndex173, depth173
			return false
		},
		/* 42 NextExpression <- <(Expression ListExpansion?)> */
		func() bool {
			position177, tokenIndex177, depth177 := position, tokenIndex, depth
			{
				position178 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l177
				}
				{
					position179, tokenIndex179, depth179 := position, tokenIndex, depth
					if !_rules[ruleListExpansion]() {
						goto l179
					}
					goto l180
				l179:
					position, tokenIndex, depth = position179, tokenIndex179, depth179
				}
			l180:
				depth--
				add(ruleNextExpression, position178)
			}
			return true
		l177:
			position, tokenIndex, depth = position177, tokenIndex177, depth177
			return false
		},
		/* 43 ListExpansion <- <('.' '.' '.' ws)> */
		func() bool {
			position181, tokenIndex181, depth181 := position, tokenIndex, depth
			{
				position182 := position
				depth++
				if buffer[position] != rune('.') {
					goto l181
				}
				position++
				if buffer[position] != rune('.') {
					goto l181
				}
				position++
				if buffer[position] != rune('.') {
					goto l181
				}
				position++
				if !_rules[rulews]() {
					goto l181
				}
				depth--
				add(ruleListExpansion, position182)
			}
			return true
		l181:
			position, tokenIndex, depth = position181, tokenIndex181, depth181
			return false
		},
		/* 44 Projection <- <('.'? (('[' '*' ']') / Slice) ProjectionValue ChainedQualifiedExpression*)> */
		func() bool {
			position183, tokenIndex183, depth183 := position, tokenIndex, depth
			{
				position184 := position
				depth++
				{
					position185, tokenIndex185, depth185 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l185
					}
					position++
					goto l186
				l185:
					position, tokenIndex, depth = position185, tokenIndex185, depth185
				}
			l186:
				{
					position187, tokenIndex187, depth187 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l188
					}
					position++
					if buffer[position] != rune('*') {
						goto l188
					}
					position++
					if buffer[position] != rune(']') {
						goto l188
					}
					position++
					goto l187
				l188:
					position, tokenIndex, depth = position187, tokenIndex187, depth187
					if !_rules[ruleSlice]() {
						goto l183
					}
				}
			l187:
				if !_rules[ruleProjectionValue]() {
					goto l183
				}
			l189:
				{
					position190, tokenIndex190, depth190 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l190
					}
					goto l189
				l190:
					position, tokenIndex, depth = position190, tokenIndex190, depth190
				}
				depth--
				add(ruleProjection, position184)
			}
			return true
		l183:
			position, tokenIndex, depth = position183, tokenIndex183, depth183
			return false
		},
		/* 45 ProjectionValue <- <Action0> */
		func() bool {
			position191, tokenIndex191, depth191 := position, tokenIndex, depth
			{
				position192 := position
				depth++
				if !_rules[ruleAction0]() {
					goto l191
				}
				depth--
				add(ruleProjectionValue, position192)
			}
			return true
		l191:
			position, tokenIndex, depth = position191, tokenIndex191, depth191
			return false
		},
		/* 46 Substitution <- <('*' Level0)> */
		func() bool {
			position193, tokenIndex193, depth193 := position, tokenIndex, depth
			{
				position194 := position
				depth++
				if buffer[position] != rune('*') {
					goto l193
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l193
				}
				depth--
				add(ruleSubstitution, position194)
			}
			return true
		l193:
			position, tokenIndex, depth = position193, tokenIndex193, depth193
			return false
		},
		/* 47 Not <- <('!' ws Level0)> */
		func() bool {
			position195, tokenIndex195, depth195 := position, tokenIndex, depth
			{
				position196 := position
				depth++
				if buffer[position] != rune('!') {
					goto l195
				}
				position++
				if !_rules[rulews]() {
					goto l195
				}
				if !_rules[ruleLevel0]() {
					goto l195
				}
				depth--
				add(ruleNot, position196)
			}
			return true
		l195:
			position, tokenIndex, depth = position195, tokenIndex195, depth195
			return false
		},
		/* 48 Grouped <- <('(' Expression ')')> */
		func() bool {
			position197, tokenIndex197, depth197 := position, tokenIndex, depth
			{
				position198 := position
				depth++
				if buffer[position] != rune('(') {
					goto l197
				}
				position++
				if !_rules[ruleExpression]() {
					goto l197
				}
				if buffer[position] != rune(')') {
					goto l197
				}
				position++
				depth--
				add(ruleGrouped, position198)
			}
			return true
		l197:
			position, tokenIndex, depth = position197, tokenIndex197, depth197
			return false
		},
		/* 49 Range <- <(StartRange Expression? RangeOp Expression? ']')> */
		func() bool {
			position199, tokenIndex199, depth199 := position, tokenIndex, depth
			{
				position200 := position
				depth++
				if !_rules[ruleStartRange]() {
					goto l199
				}
				{
					position201, tokenIndex201, depth201 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l201
					}
					goto l202
				l201:
					position, tokenIndex, depth = position201, tokenIndex201, depth201
				}
			l202:
				if !_rules[ruleRangeOp]() {
					goto l199
				}
				{
					position203, tokenIndex203, depth203 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l203
					}
					goto l204
				l203:
					position, tokenIndex, depth = position203, tokenIndex203, depth203
				}
			l204:
				if buffer[position] != rune(']') {
					goto l199
				}
				position++
				depth--
				add(ruleRange, position200)
			}
			return true
		l199:
			position, tokenIndex, depth = position199, tokenIndex199, depth199
			return false
		},
		/* 50 StartRange <- <'['> */
		func() bool {
			position205, tokenIndex205, depth205 := position, tokenIndex, depth
			{
				position206 := position
				depth++
				if buffer[position] != rune('[') {
					goto l205
				}
				position++
				depth--
				add(ruleStartRange, position206)
			}
			return true
		l205:
			position, tokenIndex, depth = position205, tokenIndex205, depth205
			return false
		},
		/* 51 RangeOp <- <('.' '.')> */
		func() bool {
			position207, tokenIndex207, depth207 := position, tokenIndex, depth
			{
				position208 := position
				depth++
				if buffer[position] != rune('.') {
					goto l207
				}
				position++
				if buffer[position] != rune('.') {
					goto l207
				}
				position++
				depth--
				add(ruleRangeOp, position208)
			}
			return true
		l207:
			position, tokenIndex, depth = position207, tokenIndex207, depth207
			return false
		},
		/* 52 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position209, tokenIndex209, depth209 := position, tokenIndex, depth
			{
				position210 := position
				depth++
				{
					position211, tokenIndex211, depth211 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l211
					}
					position++
					goto l212
				l211:
					position, tokenIndex, depth = position211, tokenIndex211, depth211
				}
			l212:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l209
				}
				position++
			l213:
				{
					position214, tokenIndex214, depth214 := position, tokenIndex, depth
					{
						position215, tokenIndex215, depth215 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l216
						}
						position++
						goto l215
					l216:
						position, tokenIndex, depth = position215, tokenIndex215, depth215
						if buffer[position] != rune('_') {
							goto l214
						}
						position++
					}
				l215:
					goto l213
				l214:
					position, tokenIndex, depth = position214, tokenIndex214, depth214
				}
				depth--
				add(ruleInteger, position210)
			}
			return true
		l209:
			position, tokenIndex, depth = position209, tokenIndex209, depth209
			return false
		},
		/* 53 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position217, tokenIndex217, depth217 := position, tokenIndex, depth
			{
				position218 := position
				depth++
				if buffer[position] != rune('"') {
					goto l217
				}
				position++
			l219:
				{
					position220, tokenIndex220, depth220 := position, tokenIndex, depth
					{
						position221, tokenIndex221, depth221 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l222
						}
						position++
						if buffer[position] != rune('"') {
							goto l222
						}
						position++
						goto l221
					l222:
						position, tokenIndex, depth = position221, tokenIndex221, depth221
						{
							position223, tokenIndex223, depth223 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l223
							}
							position++
							goto l220
						l223:
							position, tokenIndex, depth = position223, tokenIndex223, depth223
						}
						if !matchDot() {
							goto l220
						}
					}
				l221:
					goto l219
				l220:
					position, tokenIndex, depth = position220, tokenIndex220, depth220
				}
				if buffer[position] != rune('"') {
					goto l217
				}
				position++
				depth--
				add(ruleString, position218)
			}
			return true
		l217:
			position, tokenIndex, depth = position217, tokenIndex217, depth217
			return false
		},
		/* 54 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position224, tokenIndex224, depth224 := position, tokenIndex, depth
			{
				position225 := position
				depth++
				{
					position226, tokenIndex226, depth226 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l227
					}
					position++
					if buffer[position] != rune('r') {
						goto l227
					}
					position++
					if buffer[position] != rune('u') {
						goto l227
					}
					position++
					if buffer[position] != rune('e') {
						goto l227
					}
					position++
					goto l226
				l227:
					position, tokenIndex, depth = position226, tokenIndex226, depth226
					if buffer[position] != rune('f') {
						goto l224
					}
					position++
					if buffer[position] != rune('a') {
						goto l224
					}
					position++
					if buffer[position] != rune('l') {
						goto l224
					}
					position++
					if buffer[position] != rune('s') {
						goto l224
					}
					position++
					if buffer[position] != rune('e') {
						goto l224
					}
					position++
				}
			l226:
				depth--
				add(ruleBoolean, position225)
			}
			return true
		l224:
			position, tokenIndex, depth = position224, tokenIndex224, depth224
			return false
		},
		/* 55 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position228, tokenIndex228, depth228 := position, tokenIndex, depth
			{
				position229 := position
				depth++
				{
					position230, tokenIndex230, depth230 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l231
					}
					position++
					if buffer[position] != rune('i') {
						goto l231
					}
					position++
					if buffer[position] != rune('l') {
						goto l231
					}
					position++
					goto l230
				l231:
					position, tokenIndex, depth = position230, tokenIndex230, depth230
					if buffer[position] != rune('~') {
						goto l228
					}
					position++
				}
			l230:
				depth--
				add(ruleNil, position229)
			}
			return true
		l228:
			position, tokenIndex, depth = position228, tokenIndex228, depth228
			return false
		},
		/* 56 Undefined <- <('~' '~')> */
		func() bool {
			position232, tokenIndex232, depth232 := position, tokenIndex, depth
			{
				position233 := position
				depth++
				if buffer[position] != rune('~') {
					goto l232
				}
				position++
				if buffer[position] != rune('~') {
					goto l232
				}
				position++
				depth--
				add(ruleUndefined, position233)
			}
			return true
		l232:
			position, tokenIndex, depth = position232, tokenIndex232, depth232
			return false
		},
		/* 57 Symbol <- <('$' Name)> */
		func() bool {
			position234, tokenIndex234, depth234 := position, tokenIndex, depth
			{
				position235 := position
				depth++
				if buffer[position] != rune('$') {
					goto l234
				}
				position++
				if !_rules[ruleName]() {
					goto l234
				}
				depth--
				add(ruleSymbol, position235)
			}
			return true
		l234:
			position, tokenIndex, depth = position234, tokenIndex234, depth234
			return false
		},
		/* 58 List <- <(StartList ExpressionList? ']')> */
		func() bool {
			position236, tokenIndex236, depth236 := position, tokenIndex, depth
			{
				position237 := position
				depth++
				if !_rules[ruleStartList]() {
					goto l236
				}
				{
					position238, tokenIndex238, depth238 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l238
					}
					goto l239
				l238:
					position, tokenIndex, depth = position238, tokenIndex238, depth238
				}
			l239:
				if buffer[position] != rune(']') {
					goto l236
				}
				position++
				depth--
				add(ruleList, position237)
			}
			return true
		l236:
			position, tokenIndex, depth = position236, tokenIndex236, depth236
			return false
		},
		/* 59 StartList <- <('[' ws)> */
		func() bool {
			position240, tokenIndex240, depth240 := position, tokenIndex, depth
			{
				position241 := position
				depth++
				if buffer[position] != rune('[') {
					goto l240
				}
				position++
				if !_rules[rulews]() {
					goto l240
				}
				depth--
				add(ruleStartList, position241)
			}
			return true
		l240:
			position, tokenIndex, depth = position240, tokenIndex240, depth240
			return false
		},
		/* 60 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position242, tokenIndex242, depth242 := position, tokenIndex, depth
			{
				position243 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l242
				}
				if !_rules[rulews]() {
					goto l242
				}
				{
					position244, tokenIndex244, depth244 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l244
					}
					goto l245
				l244:
					position, tokenIndex, depth = position244, tokenIndex244, depth244
				}
			l245:
				if buffer[position] != rune('}') {
					goto l242
				}
				position++
				depth--
				add(ruleMap, position243)
			}
			return true
		l242:
			position, tokenIndex, depth = position242, tokenIndex242, depth242
			return false
		},
		/* 61 CreateMap <- <'{'> */
		func() bool {
			position246, tokenIndex246, depth246 := position, tokenIndex, depth
			{
				position247 := position
				depth++
				if buffer[position] != rune('{') {
					goto l246
				}
				position++
				depth--
				add(ruleCreateMap, position247)
			}
			return true
		l246:
			position, tokenIndex, depth = position246, tokenIndex246, depth246
			return false
		},
		/* 62 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position248, tokenIndex248, depth248 := position, tokenIndex, depth
			{
				position249 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l248
				}
			l250:
				{
					position251, tokenIndex251, depth251 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l251
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l251
					}
					goto l250
				l251:
					position, tokenIndex, depth = position251, tokenIndex251, depth251
				}
				depth--
				add(ruleAssignments, position249)
			}
			return true
		l248:
			position, tokenIndex, depth = position248, tokenIndex248, depth248
			return false
		},
		/* 63 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position252, tokenIndex252, depth252 := position, tokenIndex, depth
			{
				position253 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l252
				}
				if buffer[position] != rune('=') {
					goto l252
				}
				position++
				if !_rules[ruleExpression]() {
					goto l252
				}
				depth--
				add(ruleAssignment, position253)
			}
			return true
		l252:
			position, tokenIndex, depth = position252, tokenIndex252, depth252
			return false
		},
		/* 64 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position254, tokenIndex254, depth254 := position, tokenIndex, depth
			{
				position255 := position
				depth++
				{
					position256, tokenIndex256, depth256 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l257
					}
					goto l256
				l257:
					position, tokenIndex, depth = position256, tokenIndex256, depth256
					if !_rules[ruleSimpleMerge]() {
						goto l254
					}
				}
			l256:
				depth--
				add(ruleMerge, position255)
			}
			return true
		l254:
			position, tokenIndex, depth = position254, tokenIndex254, depth254
			return false
		},
		/* 65 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position258, tokenIndex258, depth258 := position, tokenIndex, depth
			{
				position259 := position
				depth++
				if buffer[position] != rune('m') {
					goto l258
				}
				position++
				if buffer[position] != rune('e') {
					goto l258
				}
				position++
				if buffer[position] != rune('r') {
					goto l258
				}
				position++
				if buffer[position] != rune('g') {
					goto l258
				}
				position++
				if buffer[position] != rune('e') {
					goto l258
				}
				position++
				{
					position260, tokenIndex260, depth260 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l260
					}
					if !_rules[ruleRequired]() {
						goto l260
					}
					goto l258
				l260:
					position, tokenIndex, depth = position260, tokenIndex260, depth260
				}
				{
					position261, tokenIndex261, depth261 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l261
					}
					{
						position263, tokenIndex263, depth263 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l264
						}
						goto l263
					l264:
						position, tokenIndex, depth = position263, tokenIndex263, depth263
						if !_rules[ruleOn]() {
							goto l261
						}
					}
				l263:
					goto l262
				l261:
					position, tokenIndex, depth = position261, tokenIndex261, depth261
				}
			l262:
				if !_rules[rulereq_ws]() {
					goto l258
				}
				if !_rules[ruleReference]() {
					goto l258
				}
				depth--
				add(ruleRefMerge, position259)
			}
			return true
		l258:
			position, tokenIndex, depth = position258, tokenIndex258, depth258
			return false
		},
		/* 66 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' !'(' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position265, tokenIndex265, depth265 := position, tokenIndex, depth
			{
				position266 := position
				depth++
				if buffer[position] != rune('m') {
					goto l265
				}
				position++
				if buffer[position] != rune('e') {
					goto l265
				}
				position++
				if buffer[position] != rune('r') {
					goto l265
				}
				position++
				if buffer[position] != rune('g') {
					goto l265
				}
				position++
				if buffer[position] != rune('e') {
					goto l265
				}
				position++
				{
					position267, tokenIndex267, depth267 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l267
					}
					position++
					goto l265
				l267:
					position, tokenIndex, depth = position267, tokenIndex267, depth267
				}
				{
					position268, tokenIndex268, depth268 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l268
					}
					{
						position270, tokenIndex270, depth270 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l271
						}
						goto l270
					l271:
						position, tokenIndex, depth = position270, tokenIndex270, depth270
						if !_rules[ruleRequired]() {
							goto l272
						}
						goto l270
					l272:
						position, tokenIndex, depth = position270, tokenIndex270, depth270
						if !_rules[ruleOn]() {
							goto l268
						}
					}
				l270:
					goto l269
				l268:
					position, tokenIndex, depth = position268, tokenIndex268, depth268
				}
			l269:
				depth--
				add(ruleSimpleMerge, position266)
			}
			return true
		l265:
			position, tokenIndex, depth = position265, tokenIndex265, depth265
			return false
		},
		/* 67 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position273, tokenIndex273, depth273 := position, tokenIndex, depth
			{
				position274 := position
				depth++
				if buffer[position] != rune('r') {
					goto l273
				}
				position++
				if buffer[position] != rune('e') {
					goto l273
				}
				position++
				if buffer[position] != rune('p') {
					goto l273
				}
				position++
				if buffer[position] != rune('l') {
					goto l273
				}
				position++
				if buffer[position] != rune('a') {
					goto l273
				}
				position++
				if buffer[position] != rune('c') {
					goto l273
				}
				position++
				if buffer[position] != rune('e') {
					goto l273
				}
				position++
				depth--
				add(ruleReplace, position274)
			}
			return true
		l273:
			position, tokenIndex, depth = position273, tokenIndex273, depth273
			return false
		},
		/* 68 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position275, tokenIndex275, depth275 := position, tokenIndex, depth
			{
				position276 := position
				depth++
				if buffer[position] != rune('r') {
					goto l275
				}
				position++
				if buffer[position] != rune('e') {
					goto l275
				}
				position++
				if buffer[position] != rune('q') {
					goto l275
				}
				position++
				if buffer[position] != rune('u') {
					goto l275
				}
				position++
				if buffer[position] != rune('i') {
					goto l275
				}
				position++
				if buffer[position] != rune('r') {
					goto l275
				}
				position++
				if buffer[position] != rune('e') {
					goto l275
				}
				position++
				if buffer[position] != rune('d') {
					goto l275
				}
				position++
				depth--
				add(ruleRequired, position276)
			}
			return true
		l275:
			position, tokenIndex, depth = position275, tokenIndex275, depth275
			return false
		},
		/* 69 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position277, tokenIndex277, depth277 := position, tokenIndex, depth
			{
				position278 := position
				depth++
				if buffer[position] != rune('o') {
					goto l277
				}
				position++
				if buffer[position] != rune('n') {
					goto l277
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l277
				}
				if !_rules[ruleName]() {
					goto l277
				}
				depth--
				add(ruleOn, position278)
			}
			return true
		l277:
			position, tokenIndex, depth = position277, tokenIndex277, depth277
			return false
		},
		/* 70 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position279, tokenIndex279, depth279 := position, tokenIndex, depth
			{
				position280 := position
				depth++
				if buffer[position] != rune('a') {
					goto l279
				}
				position++
				if buffer[position] != rune('u') {
					goto l279
				}
				position++
				if buffer[position] != rune('t') {
					goto l279
				}
				position++
				if buffer[position] != rune('o') {
					goto l279
				}
				position++
				depth--
				add(ruleAuto, position280)
			}
			return true
		l279:
			position, tokenIndex, depth = position279, tokenIndex279, depth279
			return false
		},
		/* 71 Default <- <Action1> */
		func() bool {
			position281, tokenIndex281, depth281 := position, tokenIndex, depth
			{
				position282 := position
				depth++
				if !_rules[ruleAction1]() {
					goto l281
				}
				depth--
				add(ruleDefault, position282)
			}
			return true
		l281:
			position, tokenIndex, depth = position281, tokenIndex281, depth281
			return false
		},
		/* 72 Sync <- <('s' 'y' 'n' 'c' '[' Level7 ((((LambdaExpr LambdaExt) / (LambdaOrExpr LambdaOrExpr)) (('|' Expression) / Default)) / (LambdaOrExpr Default Default)) ']')> */
		func() bool {
			position283, tokenIndex283, depth283 := position, tokenIndex, depth
			{
				position284 := position
				depth++
				if buffer[position] != rune('s') {
					goto l283
				}
				position++
				if buffer[position] != rune('y') {
					goto l283
				}
				position++
				if buffer[position] != rune('n') {
					goto l283
				}
				position++
				if buffer[position] != rune('c') {
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
				{
					position285, tokenIndex285, depth285 := position, tokenIndex, depth
					{
						position287, tokenIndex287, depth287 := position, tokenIndex, depth
						if !_rules[ruleLambdaExpr]() {
							goto l288
						}
						if !_rules[ruleLambdaExt]() {
							goto l288
						}
						goto l287
					l288:
						position, tokenIndex, depth = position287, tokenIndex287, depth287
						if !_rules[ruleLambdaOrExpr]() {
							goto l286
						}
						if !_rules[ruleLambdaOrExpr]() {
							goto l286
						}
					}
				l287:
					{
						position289, tokenIndex289, depth289 := position, tokenIndex, depth
						if buffer[position] != rune('|') {
							goto l290
						}
						position++
						if !_rules[ruleExpression]() {
							goto l290
						}
						goto l289
					l290:
						position, tokenIndex, depth = position289, tokenIndex289, depth289
						if !_rules[ruleDefault]() {
							goto l286
						}
					}
				l289:
					goto l285
				l286:
					position, tokenIndex, depth = position285, tokenIndex285, depth285
					if !_rules[ruleLambdaOrExpr]() {
						goto l283
					}
					if !_rules[ruleDefault]() {
						goto l283
					}
					if !_rules[ruleDefault]() {
						goto l283
					}
				}
			l285:
				if buffer[position] != rune(']') {
					goto l283
				}
				position++
				depth--
				add(ruleSync, position284)
			}
			return true
		l283:
			position, tokenIndex, depth = position283, tokenIndex283, depth283
			return false
		},
		/* 73 LambdaExt <- <(',' Expression)> */
		func() bool {
			position291, tokenIndex291, depth291 := position, tokenIndex, depth
			{
				position292 := position
				depth++
				if buffer[position] != rune(',') {
					goto l291
				}
				position++
				if !_rules[ruleExpression]() {
					goto l291
				}
				depth--
				add(ruleLambdaExt, position292)
			}
			return true
		l291:
			position, tokenIndex, depth = position291, tokenIndex291, depth291
			return false
		},
		/* 74 LambdaOrExpr <- <(LambdaExpr / ('|' Expression))> */
		func() bool {
			position293, tokenIndex293, depth293 := position, tokenIndex, depth
			{
				position294 := position
				depth++
				{
					position295, tokenIndex295, depth295 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l296
					}
					goto l295
				l296:
					position, tokenIndex, depth = position295, tokenIndex295, depth295
					if buffer[position] != rune('|') {
						goto l293
					}
					position++
					if !_rules[ruleExpression]() {
						goto l293
					}
				}
			l295:
				depth--
				add(ruleLambdaOrExpr, position294)
			}
			return true
		l293:
			position, tokenIndex, depth = position293, tokenIndex293, depth293
			return false
		},
		/* 75 Catch <- <('c' 'a' 't' 'c' 'h' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position297, tokenIndex297, depth297 := position, tokenIndex, depth
			{
				position298 := position
				depth++
				if buffer[position] != rune('c') {
					goto l297
				}
				position++
				if buffer[position] != rune('a') {
					goto l297
				}
				position++
				if buffer[position] != rune('t') {
					goto l297
				}
				position++
				if buffer[position] != rune('c') {
					goto l297
				}
				position++
				if buffer[position] != rune('h') {
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
				if !_rules[ruleLambdaOrExpr]() {
					goto l297
				}
				if buffer[position] != rune(']') {
					goto l297
				}
				position++
				depth--
				add(ruleCatch, position298)
			}
			return true
		l297:
			position, tokenIndex, depth = position297, tokenIndex297, depth297
			return false
		},
		/* 76 MapMapping <- <('m' 'a' 'p' '{' Level7 LambdaOrExpr '}')> */
		func() bool {
			position299, tokenIndex299, depth299 := position, tokenIndex, depth
			{
				position300 := position
				depth++
				if buffer[position] != rune('m') {
					goto l299
				}
				position++
				if buffer[position] != rune('a') {
					goto l299
				}
				position++
				if buffer[position] != rune('p') {
					goto l299
				}
				position++
				if buffer[position] != rune('{') {
					goto l299
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l299
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l299
				}
				if buffer[position] != rune('}') {
					goto l299
				}
				position++
				depth--
				add(ruleMapMapping, position300)
			}
			return true
		l299:
			position, tokenIndex, depth = position299, tokenIndex299, depth299
			return false
		},
		/* 77 Mapping <- <('m' 'a' 'p' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position301, tokenIndex301, depth301 := position, tokenIndex, depth
			{
				position302 := position
				depth++
				if buffer[position] != rune('m') {
					goto l301
				}
				position++
				if buffer[position] != rune('a') {
					goto l301
				}
				position++
				if buffer[position] != rune('p') {
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
				if !_rules[ruleLambdaOrExpr]() {
					goto l301
				}
				if buffer[position] != rune(']') {
					goto l301
				}
				position++
				depth--
				add(ruleMapping, position302)
			}
			return true
		l301:
			position, tokenIndex, depth = position301, tokenIndex301, depth301
			return false
		},
		/* 78 MapSelection <- <('s' 'e' 'l' 'e' 'c' 't' '{' Level7 LambdaOrExpr '}')> */
		func() bool {
			position303, tokenIndex303, depth303 := position, tokenIndex, depth
			{
				position304 := position
				depth++
				if buffer[position] != rune('s') {
					goto l303
				}
				position++
				if buffer[position] != rune('e') {
					goto l303
				}
				position++
				if buffer[position] != rune('l') {
					goto l303
				}
				position++
				if buffer[position] != rune('e') {
					goto l303
				}
				position++
				if buffer[position] != rune('c') {
					goto l303
				}
				position++
				if buffer[position] != rune('t') {
					goto l303
				}
				position++
				if buffer[position] != rune('{') {
					goto l303
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l303
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l303
				}
				if buffer[position] != rune('}') {
					goto l303
				}
				position++
				depth--
				add(ruleMapSelection, position304)
			}
			return true
		l303:
			position, tokenIndex, depth = position303, tokenIndex303, depth303
			return false
		},
		/* 79 Selection <- <('s' 'e' 'l' 'e' 'c' 't' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position305, tokenIndex305, depth305 := position, tokenIndex, depth
			{
				position306 := position
				depth++
				if buffer[position] != rune('s') {
					goto l305
				}
				position++
				if buffer[position] != rune('e') {
					goto l305
				}
				position++
				if buffer[position] != rune('l') {
					goto l305
				}
				position++
				if buffer[position] != rune('e') {
					goto l305
				}
				position++
				if buffer[position] != rune('c') {
					goto l305
				}
				position++
				if buffer[position] != rune('t') {
					goto l305
				}
				position++
				if buffer[position] != rune('[') {
					goto l305
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l305
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l305
				}
				if buffer[position] != rune(']') {
					goto l305
				}
				position++
				depth--
				add(ruleSelection, position306)
			}
			return true
		l305:
			position, tokenIndex, depth = position305, tokenIndex305, depth305
			return false
		},
		/* 80 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 LambdaOrExpr ']')> */
		func() bool {
			position307, tokenIndex307, depth307 := position, tokenIndex, depth
			{
				position308 := position
				depth++
				if buffer[position] != rune('s') {
					goto l307
				}
				position++
				if buffer[position] != rune('u') {
					goto l307
				}
				position++
				if buffer[position] != rune('m') {
					goto l307
				}
				position++
				if buffer[position] != rune('[') {
					goto l307
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l307
				}
				if buffer[position] != rune('|') {
					goto l307
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l307
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l307
				}
				if buffer[position] != rune(']') {
					goto l307
				}
				position++
				depth--
				add(ruleSum, position308)
			}
			return true
		l307:
			position, tokenIndex, depth = position307, tokenIndex307, depth307
			return false
		},
		/* 81 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position309, tokenIndex309, depth309 := position, tokenIndex, depth
			{
				position310 := position
				depth++
				if buffer[position] != rune('l') {
					goto l309
				}
				position++
				if buffer[position] != rune('a') {
					goto l309
				}
				position++
				if buffer[position] != rune('m') {
					goto l309
				}
				position++
				if buffer[position] != rune('b') {
					goto l309
				}
				position++
				if buffer[position] != rune('d') {
					goto l309
				}
				position++
				if buffer[position] != rune('a') {
					goto l309
				}
				position++
				{
					position311, tokenIndex311, depth311 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l312
					}
					goto l311
				l312:
					position, tokenIndex, depth = position311, tokenIndex311, depth311
					if !_rules[ruleLambdaExpr]() {
						goto l309
					}
				}
			l311:
				depth--
				add(ruleLambda, position310)
			}
			return true
		l309:
			position, tokenIndex, depth = position309, tokenIndex309, depth309
			return false
		},
		/* 82 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position313, tokenIndex313, depth313 := position, tokenIndex, depth
			{
				position314 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l313
				}
				if !_rules[ruleExpression]() {
					goto l313
				}
				depth--
				add(ruleLambdaRef, position314)
			}
			return true
		l313:
			position, tokenIndex, depth = position313, tokenIndex313, depth313
			return false
		},
		/* 83 LambdaExpr <- <(ws Params ws ('-' '>') Expression)> */
		func() bool {
			position315, tokenIndex315, depth315 := position, tokenIndex, depth
			{
				position316 := position
				depth++
				if !_rules[rulews]() {
					goto l315
				}
				if !_rules[ruleParams]() {
					goto l315
				}
				if !_rules[rulews]() {
					goto l315
				}
				if buffer[position] != rune('-') {
					goto l315
				}
				position++
				if buffer[position] != rune('>') {
					goto l315
				}
				position++
				if !_rules[ruleExpression]() {
					goto l315
				}
				depth--
				add(ruleLambdaExpr, position316)
			}
			return true
		l315:
			position, tokenIndex, depth = position315, tokenIndex315, depth315
			return false
		},
		/* 84 Params <- <('|' StartParams ws Names? '|')> */
		func() bool {
			position317, tokenIndex317, depth317 := position, tokenIndex, depth
			{
				position318 := position
				depth++
				if buffer[position] != rune('|') {
					goto l317
				}
				position++
				if !_rules[ruleStartParams]() {
					goto l317
				}
				if !_rules[rulews]() {
					goto l317
				}
				{
					position319, tokenIndex319, depth319 := position, tokenIndex, depth
					if !_rules[ruleNames]() {
						goto l319
					}
					goto l320
				l319:
					position, tokenIndex, depth = position319, tokenIndex319, depth319
				}
			l320:
				if buffer[position] != rune('|') {
					goto l317
				}
				position++
				depth--
				add(ruleParams, position318)
			}
			return true
		l317:
			position, tokenIndex, depth = position317, tokenIndex317, depth317
			return false
		},
		/* 85 StartParams <- <Action2> */
		func() bool {
			position321, tokenIndex321, depth321 := position, tokenIndex, depth
			{
				position322 := position
				depth++
				if !_rules[ruleAction2]() {
					goto l321
				}
				depth--
				add(ruleStartParams, position322)
			}
			return true
		l321:
			position, tokenIndex, depth = position321, tokenIndex321, depth321
			return false
		},
		/* 86 Names <- <(NextName (',' NextName)* DefaultValue? (',' NextName DefaultValue)* VarParams?)> */
		func() bool {
			position323, tokenIndex323, depth323 := position, tokenIndex, depth
			{
				position324 := position
				depth++
				if !_rules[ruleNextName]() {
					goto l323
				}
			l325:
				{
					position326, tokenIndex326, depth326 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l326
					}
					position++
					if !_rules[ruleNextName]() {
						goto l326
					}
					goto l325
				l326:
					position, tokenIndex, depth = position326, tokenIndex326, depth326
				}
				{
					position327, tokenIndex327, depth327 := position, tokenIndex, depth
					if !_rules[ruleDefaultValue]() {
						goto l327
					}
					goto l328
				l327:
					position, tokenIndex, depth = position327, tokenIndex327, depth327
				}
			l328:
			l329:
				{
					position330, tokenIndex330, depth330 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l330
					}
					position++
					if !_rules[ruleNextName]() {
						goto l330
					}
					if !_rules[ruleDefaultValue]() {
						goto l330
					}
					goto l329
				l330:
					position, tokenIndex, depth = position330, tokenIndex330, depth330
				}
				{
					position331, tokenIndex331, depth331 := position, tokenIndex, depth
					if !_rules[ruleVarParams]() {
						goto l331
					}
					goto l332
				l331:
					position, tokenIndex, depth = position331, tokenIndex331, depth331
				}
			l332:
				depth--
				add(ruleNames, position324)
			}
			return true
		l323:
			position, tokenIndex, depth = position323, tokenIndex323, depth323
			return false
		},
		/* 87 NextName <- <(ws Name ws)> */
		func() bool {
			position333, tokenIndex333, depth333 := position, tokenIndex, depth
			{
				position334 := position
				depth++
				if !_rules[rulews]() {
					goto l333
				}
				if !_rules[ruleName]() {
					goto l333
				}
				if !_rules[rulews]() {
					goto l333
				}
				depth--
				add(ruleNextName, position334)
			}
			return true
		l333:
			position, tokenIndex, depth = position333, tokenIndex333, depth333
			return false
		},
		/* 88 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position335, tokenIndex335, depth335 := position, tokenIndex, depth
			{
				position336 := position
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
						goto l335
					}
					position++
				}
			l339:
			l337:
				{
					position338, tokenIndex338, depth338 := position, tokenIndex, depth
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
							goto l338
						}
						position++
					}
				l343:
					goto l337
				l338:
					position, tokenIndex, depth = position338, tokenIndex338, depth338
				}
				depth--
				add(ruleName, position336)
			}
			return true
		l335:
			position, tokenIndex, depth = position335, tokenIndex335, depth335
			return false
		},
		/* 89 DefaultValue <- <('=' Expression)> */
		func() bool {
			position347, tokenIndex347, depth347 := position, tokenIndex, depth
			{
				position348 := position
				depth++
				if buffer[position] != rune('=') {
					goto l347
				}
				position++
				if !_rules[ruleExpression]() {
					goto l347
				}
				depth--
				add(ruleDefaultValue, position348)
			}
			return true
		l347:
			position, tokenIndex, depth = position347, tokenIndex347, depth347
			return false
		},
		/* 90 VarParams <- <('.' '.' '.' ws)> */
		func() bool {
			position349, tokenIndex349, depth349 := position, tokenIndex, depth
			{
				position350 := position
				depth++
				if buffer[position] != rune('.') {
					goto l349
				}
				position++
				if buffer[position] != rune('.') {
					goto l349
				}
				position++
				if buffer[position] != rune('.') {
					goto l349
				}
				position++
				if !_rules[rulews]() {
					goto l349
				}
				depth--
				add(ruleVarParams, position350)
			}
			return true
		l349:
			position, tokenIndex, depth = position349, tokenIndex349, depth349
			return false
		},
		/* 91 Reference <- <('.'? Key FollowUpRef)> */
		func() bool {
			position351, tokenIndex351, depth351 := position, tokenIndex, depth
			{
				position352 := position
				depth++
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
				if !_rules[ruleKey]() {
					goto l351
				}
				if !_rules[ruleFollowUpRef]() {
					goto l351
				}
				depth--
				add(ruleReference, position352)
			}
			return true
		l351:
			position, tokenIndex, depth = position351, tokenIndex351, depth351
			return false
		},
		/* 92 FollowUpRef <- <PathComponent*> */
		func() bool {
			{
				position356 := position
				depth++
			l357:
				{
					position358, tokenIndex358, depth358 := position, tokenIndex, depth
					if !_rules[rulePathComponent]() {
						goto l358
					}
					goto l357
				l358:
					position, tokenIndex, depth = position358, tokenIndex358, depth358
				}
				depth--
				add(ruleFollowUpRef, position356)
			}
			return true
		},
		/* 93 PathComponent <- <(('.' Key) / ('.'? Index))> */
		func() bool {
			position359, tokenIndex359, depth359 := position, tokenIndex, depth
			{
				position360 := position
				depth++
				{
					position361, tokenIndex361, depth361 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l362
					}
					position++
					if !_rules[ruleKey]() {
						goto l362
					}
					goto l361
				l362:
					position, tokenIndex, depth = position361, tokenIndex361, depth361
					{
						position363, tokenIndex363, depth363 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l363
						}
						position++
						goto l364
					l363:
						position, tokenIndex, depth = position363, tokenIndex363, depth363
					}
				l364:
					if !_rules[ruleIndex]() {
						goto l359
					}
				}
			l361:
				depth--
				add(rulePathComponent, position360)
			}
			return true
		l359:
			position, tokenIndex, depth = position359, tokenIndex359, depth359
			return false
		},
		/* 94 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position365, tokenIndex365, depth365 := position, tokenIndex, depth
			{
				position366 := position
				depth++
				{
					position367, tokenIndex367, depth367 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l368
					}
					position++
					goto l367
				l368:
					position, tokenIndex, depth = position367, tokenIndex367, depth367
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l369
					}
					position++
					goto l367
				l369:
					position, tokenIndex, depth = position367, tokenIndex367, depth367
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l370
					}
					position++
					goto l367
				l370:
					position, tokenIndex, depth = position367, tokenIndex367, depth367
					if buffer[position] != rune('_') {
						goto l365
					}
					position++
				}
			l367:
			l371:
				{
					position372, tokenIndex372, depth372 := position, tokenIndex, depth
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
							goto l377
						}
						position++
						goto l373
					l377:
						position, tokenIndex, depth = position373, tokenIndex373, depth373
						if buffer[position] != rune('-') {
							goto l372
						}
						position++
					}
				l373:
					goto l371
				l372:
					position, tokenIndex, depth = position372, tokenIndex372, depth372
				}
				{
					position378, tokenIndex378, depth378 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l378
					}
					position++
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
							goto l378
						}
						position++
					}
				l380:
				l384:
					{
						position385, tokenIndex385, depth385 := position, tokenIndex, depth
						{
							position386, tokenIndex386, depth386 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l387
							}
							position++
							goto l386
						l387:
							position, tokenIndex, depth = position386, tokenIndex386, depth386
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l388
							}
							position++
							goto l386
						l388:
							position, tokenIndex, depth = position386, tokenIndex386, depth386
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l389
							}
							position++
							goto l386
						l389:
							position, tokenIndex, depth = position386, tokenIndex386, depth386
							if buffer[position] != rune('_') {
								goto l390
							}
							position++
							goto l386
						l390:
							position, tokenIndex, depth = position386, tokenIndex386, depth386
							if buffer[position] != rune('-') {
								goto l385
							}
							position++
						}
					l386:
						goto l384
					l385:
						position, tokenIndex, depth = position385, tokenIndex385, depth385
					}
					goto l379
				l378:
					position, tokenIndex, depth = position378, tokenIndex378, depth378
				}
			l379:
				depth--
				add(ruleKey, position366)
			}
			return true
		l365:
			position, tokenIndex, depth = position365, tokenIndex365, depth365
			return false
		},
		/* 95 Index <- <('[' '-'? [0-9]+ ']')> */
		func() bool {
			position391, tokenIndex391, depth391 := position, tokenIndex, depth
			{
				position392 := position
				depth++
				if buffer[position] != rune('[') {
					goto l391
				}
				position++
				{
					position393, tokenIndex393, depth393 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l393
					}
					position++
					goto l394
				l393:
					position, tokenIndex, depth = position393, tokenIndex393, depth393
				}
			l394:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l391
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
				if buffer[position] != rune(']') {
					goto l391
				}
				position++
				depth--
				add(ruleIndex, position392)
			}
			return true
		l391:
			position, tokenIndex, depth = position391, tokenIndex391, depth391
			return false
		},
		/* 96 IP <- <([0-9]+ '.' [0-9]+ '.' [0-9]+ '.' [0-9]+)> */
		func() bool {
			position397, tokenIndex397, depth397 := position, tokenIndex, depth
			{
				position398 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l397
				}
				position++
			l399:
				{
					position400, tokenIndex400, depth400 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l400
					}
					position++
					goto l399
				l400:
					position, tokenIndex, depth = position400, tokenIndex400, depth400
				}
				if buffer[position] != rune('.') {
					goto l397
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l397
				}
				position++
			l401:
				{
					position402, tokenIndex402, depth402 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l402
					}
					position++
					goto l401
				l402:
					position, tokenIndex, depth = position402, tokenIndex402, depth402
				}
				if buffer[position] != rune('.') {
					goto l397
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l397
				}
				position++
			l403:
				{
					position404, tokenIndex404, depth404 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l404
					}
					position++
					goto l403
				l404:
					position, tokenIndex, depth = position404, tokenIndex404, depth404
				}
				if buffer[position] != rune('.') {
					goto l397
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l397
				}
				position++
			l405:
				{
					position406, tokenIndex406, depth406 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l406
					}
					position++
					goto l405
				l406:
					position, tokenIndex, depth = position406, tokenIndex406, depth406
				}
				depth--
				add(ruleIP, position398)
			}
			return true
		l397:
			position, tokenIndex, depth = position397, tokenIndex397, depth397
			return false
		},
		/* 97 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position408 := position
				depth++
			l409:
				{
					position410, tokenIndex410, depth410 := position, tokenIndex, depth
					{
						position411, tokenIndex411, depth411 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l412
						}
						position++
						goto l411
					l412:
						position, tokenIndex, depth = position411, tokenIndex411, depth411
						if buffer[position] != rune('\t') {
							goto l413
						}
						position++
						goto l411
					l413:
						position, tokenIndex, depth = position411, tokenIndex411, depth411
						if buffer[position] != rune('\n') {
							goto l414
						}
						position++
						goto l411
					l414:
						position, tokenIndex, depth = position411, tokenIndex411, depth411
						if buffer[position] != rune('\r') {
							goto l410
						}
						position++
					}
				l411:
					goto l409
				l410:
					position, tokenIndex, depth = position410, tokenIndex410, depth410
				}
				depth--
				add(rulews, position408)
			}
			return true
		},
		/* 98 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position415, tokenIndex415, depth415 := position, tokenIndex, depth
			{
				position416 := position
				depth++
				{
					position419, tokenIndex419, depth419 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l420
					}
					position++
					goto l419
				l420:
					position, tokenIndex, depth = position419, tokenIndex419, depth419
					if buffer[position] != rune('\t') {
						goto l421
					}
					position++
					goto l419
				l421:
					position, tokenIndex, depth = position419, tokenIndex419, depth419
					if buffer[position] != rune('\n') {
						goto l422
					}
					position++
					goto l419
				l422:
					position, tokenIndex, depth = position419, tokenIndex419, depth419
					if buffer[position] != rune('\r') {
						goto l415
					}
					position++
				}
			l419:
			l417:
				{
					position418, tokenIndex418, depth418 := position, tokenIndex, depth
					{
						position423, tokenIndex423, depth423 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l424
						}
						position++
						goto l423
					l424:
						position, tokenIndex, depth = position423, tokenIndex423, depth423
						if buffer[position] != rune('\t') {
							goto l425
						}
						position++
						goto l423
					l425:
						position, tokenIndex, depth = position423, tokenIndex423, depth423
						if buffer[position] != rune('\n') {
							goto l426
						}
						position++
						goto l423
					l426:
						position, tokenIndex, depth = position423, tokenIndex423, depth423
						if buffer[position] != rune('\r') {
							goto l418
						}
						position++
					}
				l423:
					goto l417
				l418:
					position, tokenIndex, depth = position418, tokenIndex418, depth418
				}
				depth--
				add(rulereq_ws, position416)
			}
			return true
		l415:
			position, tokenIndex, depth = position415, tokenIndex415, depth415
			return false
		},
		/* 100 Action0 <- <{}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 101 Action1 <- <{}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 102 Action2 <- <{}> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
	}
	p.rules = _rules
}
