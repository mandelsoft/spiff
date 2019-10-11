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
	rules  [99]func() bool
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
		/* 32 ChainedQualifiedExpression <- <(ChainedCall / ChainedRef / ChainedDynRef / Projection)> */
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
					if !_rules[ruleChainedRef]() {
						goto l144
					}
					goto l142
				l144:
					position, tokenIndex, depth = position142, tokenIndex142, depth142
					if !_rules[ruleChainedDynRef]() {
						goto l145
					}
					goto l142
				l145:
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
			position146, tokenIndex146, depth146 := position, tokenIndex, depth
			{
				position147 := position
				depth++
				if !_rules[rulePathComponent]() {
					goto l146
				}
				if !_rules[ruleFollowUpRef]() {
					goto l146
				}
				depth--
				add(ruleChainedRef, position147)
			}
			return true
		l146:
			position, tokenIndex, depth = position146, tokenIndex146, depth146
			return false
		},
		/* 34 ChainedDynRef <- <('.'? '[' Expression ']')> */
		func() bool {
			position148, tokenIndex148, depth148 := position, tokenIndex, depth
			{
				position149 := position
				depth++
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l150
					}
					position++
					goto l151
				l150:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
				}
			l151:
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
			position152, tokenIndex152, depth152 := position, tokenIndex, depth
			{
				position153 := position
				depth++
				if !_rules[ruleRange]() {
					goto l152
				}
				depth--
				add(ruleSlice, position153)
			}
			return true
		l152:
			position, tokenIndex, depth = position152, tokenIndex152, depth152
			return false
		},
		/* 36 ChainedCall <- <(StartArguments ExpressionList? ')')> */
		func() bool {
			position154, tokenIndex154, depth154 := position, tokenIndex, depth
			{
				position155 := position
				depth++
				if !_rules[ruleStartArguments]() {
					goto l154
				}
				{
					position156, tokenIndex156, depth156 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l156
					}
					goto l157
				l156:
					position, tokenIndex, depth = position156, tokenIndex156, depth156
				}
			l157:
				if buffer[position] != rune(')') {
					goto l154
				}
				position++
				depth--
				add(ruleChainedCall, position155)
			}
			return true
		l154:
			position, tokenIndex, depth = position154, tokenIndex154, depth154
			return false
		},
		/* 37 StartArguments <- <('(' ws)> */
		func() bool {
			position158, tokenIndex158, depth158 := position, tokenIndex, depth
			{
				position159 := position
				depth++
				if buffer[position] != rune('(') {
					goto l158
				}
				position++
				if !_rules[rulews]() {
					goto l158
				}
				depth--
				add(ruleStartArguments, position159)
			}
			return true
		l158:
			position, tokenIndex, depth = position158, tokenIndex158, depth158
			return false
		},
		/* 38 ExpressionList <- <(NextExpression (',' NextExpression)*)> */
		func() bool {
			position160, tokenIndex160, depth160 := position, tokenIndex, depth
			{
				position161 := position
				depth++
				if !_rules[ruleNextExpression]() {
					goto l160
				}
			l162:
				{
					position163, tokenIndex163, depth163 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l163
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l163
					}
					goto l162
				l163:
					position, tokenIndex, depth = position163, tokenIndex163, depth163
				}
				depth--
				add(ruleExpressionList, position161)
			}
			return true
		l160:
			position, tokenIndex, depth = position160, tokenIndex160, depth160
			return false
		},
		/* 39 NextExpression <- <(Expression VarArgs?)> */
		func() bool {
			position164, tokenIndex164, depth164 := position, tokenIndex, depth
			{
				position165 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l164
				}
				{
					position166, tokenIndex166, depth166 := position, tokenIndex, depth
					if !_rules[ruleVarArgs]() {
						goto l166
					}
					goto l167
				l166:
					position, tokenIndex, depth = position166, tokenIndex166, depth166
				}
			l167:
				depth--
				add(ruleNextExpression, position165)
			}
			return true
		l164:
			position, tokenIndex, depth = position164, tokenIndex164, depth164
			return false
		},
		/* 40 VarArgs <- <('.' '.' '.' ws)> */
		func() bool {
			position168, tokenIndex168, depth168 := position, tokenIndex, depth
			{
				position169 := position
				depth++
				if buffer[position] != rune('.') {
					goto l168
				}
				position++
				if buffer[position] != rune('.') {
					goto l168
				}
				position++
				if buffer[position] != rune('.') {
					goto l168
				}
				position++
				if !_rules[rulews]() {
					goto l168
				}
				depth--
				add(ruleVarArgs, position169)
			}
			return true
		l168:
			position, tokenIndex, depth = position168, tokenIndex168, depth168
			return false
		},
		/* 41 Projection <- <('.'? (('[' '*' ']') / Slice) ProjectionValue ChainedQualifiedExpression*)> */
		func() bool {
			position170, tokenIndex170, depth170 := position, tokenIndex, depth
			{
				position171 := position
				depth++
				{
					position172, tokenIndex172, depth172 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l172
					}
					position++
					goto l173
				l172:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
				}
			l173:
				{
					position174, tokenIndex174, depth174 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l175
					}
					position++
					if buffer[position] != rune('*') {
						goto l175
					}
					position++
					if buffer[position] != rune(']') {
						goto l175
					}
					position++
					goto l174
				l175:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
					if !_rules[ruleSlice]() {
						goto l170
					}
				}
			l174:
				if !_rules[ruleProjectionValue]() {
					goto l170
				}
			l176:
				{
					position177, tokenIndex177, depth177 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l177
					}
					goto l176
				l177:
					position, tokenIndex, depth = position177, tokenIndex177, depth177
				}
				depth--
				add(ruleProjection, position171)
			}
			return true
		l170:
			position, tokenIndex, depth = position170, tokenIndex170, depth170
			return false
		},
		/* 42 ProjectionValue <- <Action0> */
		func() bool {
			position178, tokenIndex178, depth178 := position, tokenIndex, depth
			{
				position179 := position
				depth++
				if !_rules[ruleAction0]() {
					goto l178
				}
				depth--
				add(ruleProjectionValue, position179)
			}
			return true
		l178:
			position, tokenIndex, depth = position178, tokenIndex178, depth178
			return false
		},
		/* 43 Substitution <- <('*' Level0)> */
		func() bool {
			position180, tokenIndex180, depth180 := position, tokenIndex, depth
			{
				position181 := position
				depth++
				if buffer[position] != rune('*') {
					goto l180
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l180
				}
				depth--
				add(ruleSubstitution, position181)
			}
			return true
		l180:
			position, tokenIndex, depth = position180, tokenIndex180, depth180
			return false
		},
		/* 44 Not <- <('!' ws Level0)> */
		func() bool {
			position182, tokenIndex182, depth182 := position, tokenIndex, depth
			{
				position183 := position
				depth++
				if buffer[position] != rune('!') {
					goto l182
				}
				position++
				if !_rules[rulews]() {
					goto l182
				}
				if !_rules[ruleLevel0]() {
					goto l182
				}
				depth--
				add(ruleNot, position183)
			}
			return true
		l182:
			position, tokenIndex, depth = position182, tokenIndex182, depth182
			return false
		},
		/* 45 Grouped <- <('(' Expression ')')> */
		func() bool {
			position184, tokenIndex184, depth184 := position, tokenIndex, depth
			{
				position185 := position
				depth++
				if buffer[position] != rune('(') {
					goto l184
				}
				position++
				if !_rules[ruleExpression]() {
					goto l184
				}
				if buffer[position] != rune(')') {
					goto l184
				}
				position++
				depth--
				add(ruleGrouped, position185)
			}
			return true
		l184:
			position, tokenIndex, depth = position184, tokenIndex184, depth184
			return false
		},
		/* 46 Range <- <(StartRange Expression? RangeOp Expression? ']')> */
		func() bool {
			position186, tokenIndex186, depth186 := position, tokenIndex, depth
			{
				position187 := position
				depth++
				if !_rules[ruleStartRange]() {
					goto l186
				}
				{
					position188, tokenIndex188, depth188 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l188
					}
					goto l189
				l188:
					position, tokenIndex, depth = position188, tokenIndex188, depth188
				}
			l189:
				if !_rules[ruleRangeOp]() {
					goto l186
				}
				{
					position190, tokenIndex190, depth190 := position, tokenIndex, depth
					if !_rules[ruleExpression]() {
						goto l190
					}
					goto l191
				l190:
					position, tokenIndex, depth = position190, tokenIndex190, depth190
				}
			l191:
				if buffer[position] != rune(']') {
					goto l186
				}
				position++
				depth--
				add(ruleRange, position187)
			}
			return true
		l186:
			position, tokenIndex, depth = position186, tokenIndex186, depth186
			return false
		},
		/* 47 StartRange <- <'['> */
		func() bool {
			position192, tokenIndex192, depth192 := position, tokenIndex, depth
			{
				position193 := position
				depth++
				if buffer[position] != rune('[') {
					goto l192
				}
				position++
				depth--
				add(ruleStartRange, position193)
			}
			return true
		l192:
			position, tokenIndex, depth = position192, tokenIndex192, depth192
			return false
		},
		/* 48 RangeOp <- <('.' '.')> */
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
				depth--
				add(ruleRangeOp, position195)
			}
			return true
		l194:
			position, tokenIndex, depth = position194, tokenIndex194, depth194
			return false
		},
		/* 49 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position196, tokenIndex196, depth196 := position, tokenIndex, depth
			{
				position197 := position
				depth++
				{
					position198, tokenIndex198, depth198 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l198
					}
					position++
					goto l199
				l198:
					position, tokenIndex, depth = position198, tokenIndex198, depth198
				}
			l199:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l196
				}
				position++
			l200:
				{
					position201, tokenIndex201, depth201 := position, tokenIndex, depth
					{
						position202, tokenIndex202, depth202 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l203
						}
						position++
						goto l202
					l203:
						position, tokenIndex, depth = position202, tokenIndex202, depth202
						if buffer[position] != rune('_') {
							goto l201
						}
						position++
					}
				l202:
					goto l200
				l201:
					position, tokenIndex, depth = position201, tokenIndex201, depth201
				}
				depth--
				add(ruleInteger, position197)
			}
			return true
		l196:
			position, tokenIndex, depth = position196, tokenIndex196, depth196
			return false
		},
		/* 50 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position204, tokenIndex204, depth204 := position, tokenIndex, depth
			{
				position205 := position
				depth++
				if buffer[position] != rune('"') {
					goto l204
				}
				position++
			l206:
				{
					position207, tokenIndex207, depth207 := position, tokenIndex, depth
					{
						position208, tokenIndex208, depth208 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l209
						}
						position++
						if buffer[position] != rune('"') {
							goto l209
						}
						position++
						goto l208
					l209:
						position, tokenIndex, depth = position208, tokenIndex208, depth208
						{
							position210, tokenIndex210, depth210 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l210
							}
							position++
							goto l207
						l210:
							position, tokenIndex, depth = position210, tokenIndex210, depth210
						}
						if !matchDot() {
							goto l207
						}
					}
				l208:
					goto l206
				l207:
					position, tokenIndex, depth = position207, tokenIndex207, depth207
				}
				if buffer[position] != rune('"') {
					goto l204
				}
				position++
				depth--
				add(ruleString, position205)
			}
			return true
		l204:
			position, tokenIndex, depth = position204, tokenIndex204, depth204
			return false
		},
		/* 51 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position211, tokenIndex211, depth211 := position, tokenIndex, depth
			{
				position212 := position
				depth++
				{
					position213, tokenIndex213, depth213 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l214
					}
					position++
					if buffer[position] != rune('r') {
						goto l214
					}
					position++
					if buffer[position] != rune('u') {
						goto l214
					}
					position++
					if buffer[position] != rune('e') {
						goto l214
					}
					position++
					goto l213
				l214:
					position, tokenIndex, depth = position213, tokenIndex213, depth213
					if buffer[position] != rune('f') {
						goto l211
					}
					position++
					if buffer[position] != rune('a') {
						goto l211
					}
					position++
					if buffer[position] != rune('l') {
						goto l211
					}
					position++
					if buffer[position] != rune('s') {
						goto l211
					}
					position++
					if buffer[position] != rune('e') {
						goto l211
					}
					position++
				}
			l213:
				depth--
				add(ruleBoolean, position212)
			}
			return true
		l211:
			position, tokenIndex, depth = position211, tokenIndex211, depth211
			return false
		},
		/* 52 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position215, tokenIndex215, depth215 := position, tokenIndex, depth
			{
				position216 := position
				depth++
				{
					position217, tokenIndex217, depth217 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l218
					}
					position++
					if buffer[position] != rune('i') {
						goto l218
					}
					position++
					if buffer[position] != rune('l') {
						goto l218
					}
					position++
					goto l217
				l218:
					position, tokenIndex, depth = position217, tokenIndex217, depth217
					if buffer[position] != rune('~') {
						goto l215
					}
					position++
				}
			l217:
				depth--
				add(ruleNil, position216)
			}
			return true
		l215:
			position, tokenIndex, depth = position215, tokenIndex215, depth215
			return false
		},
		/* 53 Undefined <- <('~' '~')> */
		func() bool {
			position219, tokenIndex219, depth219 := position, tokenIndex, depth
			{
				position220 := position
				depth++
				if buffer[position] != rune('~') {
					goto l219
				}
				position++
				if buffer[position] != rune('~') {
					goto l219
				}
				position++
				depth--
				add(ruleUndefined, position220)
			}
			return true
		l219:
			position, tokenIndex, depth = position219, tokenIndex219, depth219
			return false
		},
		/* 54 Symbol <- <('$' Name)> */
		func() bool {
			position221, tokenIndex221, depth221 := position, tokenIndex, depth
			{
				position222 := position
				depth++
				if buffer[position] != rune('$') {
					goto l221
				}
				position++
				if !_rules[ruleName]() {
					goto l221
				}
				depth--
				add(ruleSymbol, position222)
			}
			return true
		l221:
			position, tokenIndex, depth = position221, tokenIndex221, depth221
			return false
		},
		/* 55 List <- <(StartList ExpressionList? ']')> */
		func() bool {
			position223, tokenIndex223, depth223 := position, tokenIndex, depth
			{
				position224 := position
				depth++
				if !_rules[ruleStartList]() {
					goto l223
				}
				{
					position225, tokenIndex225, depth225 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l225
					}
					goto l226
				l225:
					position, tokenIndex, depth = position225, tokenIndex225, depth225
				}
			l226:
				if buffer[position] != rune(']') {
					goto l223
				}
				position++
				depth--
				add(ruleList, position224)
			}
			return true
		l223:
			position, tokenIndex, depth = position223, tokenIndex223, depth223
			return false
		},
		/* 56 StartList <- <('[' ws)> */
		func() bool {
			position227, tokenIndex227, depth227 := position, tokenIndex, depth
			{
				position228 := position
				depth++
				if buffer[position] != rune('[') {
					goto l227
				}
				position++
				if !_rules[rulews]() {
					goto l227
				}
				depth--
				add(ruleStartList, position228)
			}
			return true
		l227:
			position, tokenIndex, depth = position227, tokenIndex227, depth227
			return false
		},
		/* 57 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position229, tokenIndex229, depth229 := position, tokenIndex, depth
			{
				position230 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l229
				}
				if !_rules[rulews]() {
					goto l229
				}
				{
					position231, tokenIndex231, depth231 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l231
					}
					goto l232
				l231:
					position, tokenIndex, depth = position231, tokenIndex231, depth231
				}
			l232:
				if buffer[position] != rune('}') {
					goto l229
				}
				position++
				depth--
				add(ruleMap, position230)
			}
			return true
		l229:
			position, tokenIndex, depth = position229, tokenIndex229, depth229
			return false
		},
		/* 58 CreateMap <- <'{'> */
		func() bool {
			position233, tokenIndex233, depth233 := position, tokenIndex, depth
			{
				position234 := position
				depth++
				if buffer[position] != rune('{') {
					goto l233
				}
				position++
				depth--
				add(ruleCreateMap, position234)
			}
			return true
		l233:
			position, tokenIndex, depth = position233, tokenIndex233, depth233
			return false
		},
		/* 59 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position235, tokenIndex235, depth235 := position, tokenIndex, depth
			{
				position236 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l235
				}
			l237:
				{
					position238, tokenIndex238, depth238 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l238
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l238
					}
					goto l237
				l238:
					position, tokenIndex, depth = position238, tokenIndex238, depth238
				}
				depth--
				add(ruleAssignments, position236)
			}
			return true
		l235:
			position, tokenIndex, depth = position235, tokenIndex235, depth235
			return false
		},
		/* 60 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position239, tokenIndex239, depth239 := position, tokenIndex, depth
			{
				position240 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l239
				}
				if buffer[position] != rune('=') {
					goto l239
				}
				position++
				if !_rules[ruleExpression]() {
					goto l239
				}
				depth--
				add(ruleAssignment, position240)
			}
			return true
		l239:
			position, tokenIndex, depth = position239, tokenIndex239, depth239
			return false
		},
		/* 61 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position241, tokenIndex241, depth241 := position, tokenIndex, depth
			{
				position242 := position
				depth++
				{
					position243, tokenIndex243, depth243 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l244
					}
					goto l243
				l244:
					position, tokenIndex, depth = position243, tokenIndex243, depth243
					if !_rules[ruleSimpleMerge]() {
						goto l241
					}
				}
			l243:
				depth--
				add(ruleMerge, position242)
			}
			return true
		l241:
			position, tokenIndex, depth = position241, tokenIndex241, depth241
			return false
		},
		/* 62 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position245, tokenIndex245, depth245 := position, tokenIndex, depth
			{
				position246 := position
				depth++
				if buffer[position] != rune('m') {
					goto l245
				}
				position++
				if buffer[position] != rune('e') {
					goto l245
				}
				position++
				if buffer[position] != rune('r') {
					goto l245
				}
				position++
				if buffer[position] != rune('g') {
					goto l245
				}
				position++
				if buffer[position] != rune('e') {
					goto l245
				}
				position++
				{
					position247, tokenIndex247, depth247 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l247
					}
					if !_rules[ruleRequired]() {
						goto l247
					}
					goto l245
				l247:
					position, tokenIndex, depth = position247, tokenIndex247, depth247
				}
				{
					position248, tokenIndex248, depth248 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l248
					}
					{
						position250, tokenIndex250, depth250 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l251
						}
						goto l250
					l251:
						position, tokenIndex, depth = position250, tokenIndex250, depth250
						if !_rules[ruleOn]() {
							goto l248
						}
					}
				l250:
					goto l249
				l248:
					position, tokenIndex, depth = position248, tokenIndex248, depth248
				}
			l249:
				if !_rules[rulereq_ws]() {
					goto l245
				}
				if !_rules[ruleReference]() {
					goto l245
				}
				depth--
				add(ruleRefMerge, position246)
			}
			return true
		l245:
			position, tokenIndex, depth = position245, tokenIndex245, depth245
			return false
		},
		/* 63 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' !'(' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position252, tokenIndex252, depth252 := position, tokenIndex, depth
			{
				position253 := position
				depth++
				if buffer[position] != rune('m') {
					goto l252
				}
				position++
				if buffer[position] != rune('e') {
					goto l252
				}
				position++
				if buffer[position] != rune('r') {
					goto l252
				}
				position++
				if buffer[position] != rune('g') {
					goto l252
				}
				position++
				if buffer[position] != rune('e') {
					goto l252
				}
				position++
				{
					position254, tokenIndex254, depth254 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l254
					}
					position++
					goto l252
				l254:
					position, tokenIndex, depth = position254, tokenIndex254, depth254
				}
				{
					position255, tokenIndex255, depth255 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l255
					}
					{
						position257, tokenIndex257, depth257 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l258
						}
						goto l257
					l258:
						position, tokenIndex, depth = position257, tokenIndex257, depth257
						if !_rules[ruleRequired]() {
							goto l259
						}
						goto l257
					l259:
						position, tokenIndex, depth = position257, tokenIndex257, depth257
						if !_rules[ruleOn]() {
							goto l255
						}
					}
				l257:
					goto l256
				l255:
					position, tokenIndex, depth = position255, tokenIndex255, depth255
				}
			l256:
				depth--
				add(ruleSimpleMerge, position253)
			}
			return true
		l252:
			position, tokenIndex, depth = position252, tokenIndex252, depth252
			return false
		},
		/* 64 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position260, tokenIndex260, depth260 := position, tokenIndex, depth
			{
				position261 := position
				depth++
				if buffer[position] != rune('r') {
					goto l260
				}
				position++
				if buffer[position] != rune('e') {
					goto l260
				}
				position++
				if buffer[position] != rune('p') {
					goto l260
				}
				position++
				if buffer[position] != rune('l') {
					goto l260
				}
				position++
				if buffer[position] != rune('a') {
					goto l260
				}
				position++
				if buffer[position] != rune('c') {
					goto l260
				}
				position++
				if buffer[position] != rune('e') {
					goto l260
				}
				position++
				depth--
				add(ruleReplace, position261)
			}
			return true
		l260:
			position, tokenIndex, depth = position260, tokenIndex260, depth260
			return false
		},
		/* 65 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position262, tokenIndex262, depth262 := position, tokenIndex, depth
			{
				position263 := position
				depth++
				if buffer[position] != rune('r') {
					goto l262
				}
				position++
				if buffer[position] != rune('e') {
					goto l262
				}
				position++
				if buffer[position] != rune('q') {
					goto l262
				}
				position++
				if buffer[position] != rune('u') {
					goto l262
				}
				position++
				if buffer[position] != rune('i') {
					goto l262
				}
				position++
				if buffer[position] != rune('r') {
					goto l262
				}
				position++
				if buffer[position] != rune('e') {
					goto l262
				}
				position++
				if buffer[position] != rune('d') {
					goto l262
				}
				position++
				depth--
				add(ruleRequired, position263)
			}
			return true
		l262:
			position, tokenIndex, depth = position262, tokenIndex262, depth262
			return false
		},
		/* 66 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position264, tokenIndex264, depth264 := position, tokenIndex, depth
			{
				position265 := position
				depth++
				if buffer[position] != rune('o') {
					goto l264
				}
				position++
				if buffer[position] != rune('n') {
					goto l264
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l264
				}
				if !_rules[ruleName]() {
					goto l264
				}
				depth--
				add(ruleOn, position265)
			}
			return true
		l264:
			position, tokenIndex, depth = position264, tokenIndex264, depth264
			return false
		},
		/* 67 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position266, tokenIndex266, depth266 := position, tokenIndex, depth
			{
				position267 := position
				depth++
				if buffer[position] != rune('a') {
					goto l266
				}
				position++
				if buffer[position] != rune('u') {
					goto l266
				}
				position++
				if buffer[position] != rune('t') {
					goto l266
				}
				position++
				if buffer[position] != rune('o') {
					goto l266
				}
				position++
				depth--
				add(ruleAuto, position267)
			}
			return true
		l266:
			position, tokenIndex, depth = position266, tokenIndex266, depth266
			return false
		},
		/* 68 Default <- <Action1> */
		func() bool {
			position268, tokenIndex268, depth268 := position, tokenIndex, depth
			{
				position269 := position
				depth++
				if !_rules[ruleAction1]() {
					goto l268
				}
				depth--
				add(ruleDefault, position269)
			}
			return true
		l268:
			position, tokenIndex, depth = position268, tokenIndex268, depth268
			return false
		},
		/* 69 Sync <- <('s' 'y' 'n' 'c' '[' Level7 ((((LambdaExpr LambdaExt) / (LambdaOrExpr LambdaOrExpr)) (('|' Expression) / Default)) / (LambdaOrExpr Default Default)) ']')> */
		func() bool {
			position270, tokenIndex270, depth270 := position, tokenIndex, depth
			{
				position271 := position
				depth++
				if buffer[position] != rune('s') {
					goto l270
				}
				position++
				if buffer[position] != rune('y') {
					goto l270
				}
				position++
				if buffer[position] != rune('n') {
					goto l270
				}
				position++
				if buffer[position] != rune('c') {
					goto l270
				}
				position++
				if buffer[position] != rune('[') {
					goto l270
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l270
				}
				{
					position272, tokenIndex272, depth272 := position, tokenIndex, depth
					{
						position274, tokenIndex274, depth274 := position, tokenIndex, depth
						if !_rules[ruleLambdaExpr]() {
							goto l275
						}
						if !_rules[ruleLambdaExt]() {
							goto l275
						}
						goto l274
					l275:
						position, tokenIndex, depth = position274, tokenIndex274, depth274
						if !_rules[ruleLambdaOrExpr]() {
							goto l273
						}
						if !_rules[ruleLambdaOrExpr]() {
							goto l273
						}
					}
				l274:
					{
						position276, tokenIndex276, depth276 := position, tokenIndex, depth
						if buffer[position] != rune('|') {
							goto l277
						}
						position++
						if !_rules[ruleExpression]() {
							goto l277
						}
						goto l276
					l277:
						position, tokenIndex, depth = position276, tokenIndex276, depth276
						if !_rules[ruleDefault]() {
							goto l273
						}
					}
				l276:
					goto l272
				l273:
					position, tokenIndex, depth = position272, tokenIndex272, depth272
					if !_rules[ruleLambdaOrExpr]() {
						goto l270
					}
					if !_rules[ruleDefault]() {
						goto l270
					}
					if !_rules[ruleDefault]() {
						goto l270
					}
				}
			l272:
				if buffer[position] != rune(']') {
					goto l270
				}
				position++
				depth--
				add(ruleSync, position271)
			}
			return true
		l270:
			position, tokenIndex, depth = position270, tokenIndex270, depth270
			return false
		},
		/* 70 LambdaExt <- <(',' Expression)> */
		func() bool {
			position278, tokenIndex278, depth278 := position, tokenIndex, depth
			{
				position279 := position
				depth++
				if buffer[position] != rune(',') {
					goto l278
				}
				position++
				if !_rules[ruleExpression]() {
					goto l278
				}
				depth--
				add(ruleLambdaExt, position279)
			}
			return true
		l278:
			position, tokenIndex, depth = position278, tokenIndex278, depth278
			return false
		},
		/* 71 LambdaOrExpr <- <(LambdaExpr / ('|' Expression))> */
		func() bool {
			position280, tokenIndex280, depth280 := position, tokenIndex, depth
			{
				position281 := position
				depth++
				{
					position282, tokenIndex282, depth282 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l283
					}
					goto l282
				l283:
					position, tokenIndex, depth = position282, tokenIndex282, depth282
					if buffer[position] != rune('|') {
						goto l280
					}
					position++
					if !_rules[ruleExpression]() {
						goto l280
					}
				}
			l282:
				depth--
				add(ruleLambdaOrExpr, position281)
			}
			return true
		l280:
			position, tokenIndex, depth = position280, tokenIndex280, depth280
			return false
		},
		/* 72 Catch <- <('c' 'a' 't' 'c' 'h' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position284, tokenIndex284, depth284 := position, tokenIndex, depth
			{
				position285 := position
				depth++
				if buffer[position] != rune('c') {
					goto l284
				}
				position++
				if buffer[position] != rune('a') {
					goto l284
				}
				position++
				if buffer[position] != rune('t') {
					goto l284
				}
				position++
				if buffer[position] != rune('c') {
					goto l284
				}
				position++
				if buffer[position] != rune('h') {
					goto l284
				}
				position++
				if buffer[position] != rune('[') {
					goto l284
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l284
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l284
				}
				if buffer[position] != rune(']') {
					goto l284
				}
				position++
				depth--
				add(ruleCatch, position285)
			}
			return true
		l284:
			position, tokenIndex, depth = position284, tokenIndex284, depth284
			return false
		},
		/* 73 MapMapping <- <('m' 'a' 'p' '{' Level7 LambdaOrExpr '}')> */
		func() bool {
			position286, tokenIndex286, depth286 := position, tokenIndex, depth
			{
				position287 := position
				depth++
				if buffer[position] != rune('m') {
					goto l286
				}
				position++
				if buffer[position] != rune('a') {
					goto l286
				}
				position++
				if buffer[position] != rune('p') {
					goto l286
				}
				position++
				if buffer[position] != rune('{') {
					goto l286
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l286
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l286
				}
				if buffer[position] != rune('}') {
					goto l286
				}
				position++
				depth--
				add(ruleMapMapping, position287)
			}
			return true
		l286:
			position, tokenIndex, depth = position286, tokenIndex286, depth286
			return false
		},
		/* 74 Mapping <- <('m' 'a' 'p' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position288, tokenIndex288, depth288 := position, tokenIndex, depth
			{
				position289 := position
				depth++
				if buffer[position] != rune('m') {
					goto l288
				}
				position++
				if buffer[position] != rune('a') {
					goto l288
				}
				position++
				if buffer[position] != rune('p') {
					goto l288
				}
				position++
				if buffer[position] != rune('[') {
					goto l288
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l288
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l288
				}
				if buffer[position] != rune(']') {
					goto l288
				}
				position++
				depth--
				add(ruleMapping, position289)
			}
			return true
		l288:
			position, tokenIndex, depth = position288, tokenIndex288, depth288
			return false
		},
		/* 75 MapSelection <- <('s' 'e' 'l' 'e' 'c' 't' '{' Level7 LambdaOrExpr '}')> */
		func() bool {
			position290, tokenIndex290, depth290 := position, tokenIndex, depth
			{
				position291 := position
				depth++
				if buffer[position] != rune('s') {
					goto l290
				}
				position++
				if buffer[position] != rune('e') {
					goto l290
				}
				position++
				if buffer[position] != rune('l') {
					goto l290
				}
				position++
				if buffer[position] != rune('e') {
					goto l290
				}
				position++
				if buffer[position] != rune('c') {
					goto l290
				}
				position++
				if buffer[position] != rune('t') {
					goto l290
				}
				position++
				if buffer[position] != rune('{') {
					goto l290
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l290
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l290
				}
				if buffer[position] != rune('}') {
					goto l290
				}
				position++
				depth--
				add(ruleMapSelection, position291)
			}
			return true
		l290:
			position, tokenIndex, depth = position290, tokenIndex290, depth290
			return false
		},
		/* 76 Selection <- <('s' 'e' 'l' 'e' 'c' 't' '[' Level7 LambdaOrExpr ']')> */
		func() bool {
			position292, tokenIndex292, depth292 := position, tokenIndex, depth
			{
				position293 := position
				depth++
				if buffer[position] != rune('s') {
					goto l292
				}
				position++
				if buffer[position] != rune('e') {
					goto l292
				}
				position++
				if buffer[position] != rune('l') {
					goto l292
				}
				position++
				if buffer[position] != rune('e') {
					goto l292
				}
				position++
				if buffer[position] != rune('c') {
					goto l292
				}
				position++
				if buffer[position] != rune('t') {
					goto l292
				}
				position++
				if buffer[position] != rune('[') {
					goto l292
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l292
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l292
				}
				if buffer[position] != rune(']') {
					goto l292
				}
				position++
				depth--
				add(ruleSelection, position293)
			}
			return true
		l292:
			position, tokenIndex, depth = position292, tokenIndex292, depth292
			return false
		},
		/* 77 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 LambdaOrExpr ']')> */
		func() bool {
			position294, tokenIndex294, depth294 := position, tokenIndex, depth
			{
				position295 := position
				depth++
				if buffer[position] != rune('s') {
					goto l294
				}
				position++
				if buffer[position] != rune('u') {
					goto l294
				}
				position++
				if buffer[position] != rune('m') {
					goto l294
				}
				position++
				if buffer[position] != rune('[') {
					goto l294
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l294
				}
				if buffer[position] != rune('|') {
					goto l294
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l294
				}
				if !_rules[ruleLambdaOrExpr]() {
					goto l294
				}
				if buffer[position] != rune(']') {
					goto l294
				}
				position++
				depth--
				add(ruleSum, position295)
			}
			return true
		l294:
			position, tokenIndex, depth = position294, tokenIndex294, depth294
			return false
		},
		/* 78 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position296, tokenIndex296, depth296 := position, tokenIndex, depth
			{
				position297 := position
				depth++
				if buffer[position] != rune('l') {
					goto l296
				}
				position++
				if buffer[position] != rune('a') {
					goto l296
				}
				position++
				if buffer[position] != rune('m') {
					goto l296
				}
				position++
				if buffer[position] != rune('b') {
					goto l296
				}
				position++
				if buffer[position] != rune('d') {
					goto l296
				}
				position++
				if buffer[position] != rune('a') {
					goto l296
				}
				position++
				{
					position298, tokenIndex298, depth298 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l299
					}
					goto l298
				l299:
					position, tokenIndex, depth = position298, tokenIndex298, depth298
					if !_rules[ruleLambdaExpr]() {
						goto l296
					}
				}
			l298:
				depth--
				add(ruleLambda, position297)
			}
			return true
		l296:
			position, tokenIndex, depth = position296, tokenIndex296, depth296
			return false
		},
		/* 79 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position300, tokenIndex300, depth300 := position, tokenIndex, depth
			{
				position301 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l300
				}
				if !_rules[ruleExpression]() {
					goto l300
				}
				depth--
				add(ruleLambdaRef, position301)
			}
			return true
		l300:
			position, tokenIndex, depth = position300, tokenIndex300, depth300
			return false
		},
		/* 80 LambdaExpr <- <(ws Params ws ('-' '>') Expression)> */
		func() bool {
			position302, tokenIndex302, depth302 := position, tokenIndex, depth
			{
				position303 := position
				depth++
				if !_rules[rulews]() {
					goto l302
				}
				if !_rules[ruleParams]() {
					goto l302
				}
				if !_rules[rulews]() {
					goto l302
				}
				if buffer[position] != rune('-') {
					goto l302
				}
				position++
				if buffer[position] != rune('>') {
					goto l302
				}
				position++
				if !_rules[ruleExpression]() {
					goto l302
				}
				depth--
				add(ruleLambdaExpr, position303)
			}
			return true
		l302:
			position, tokenIndex, depth = position302, tokenIndex302, depth302
			return false
		},
		/* 81 Params <- <('|' StartParams ws Names? '|')> */
		func() bool {
			position304, tokenIndex304, depth304 := position, tokenIndex, depth
			{
				position305 := position
				depth++
				if buffer[position] != rune('|') {
					goto l304
				}
				position++
				if !_rules[ruleStartParams]() {
					goto l304
				}
				if !_rules[rulews]() {
					goto l304
				}
				{
					position306, tokenIndex306, depth306 := position, tokenIndex, depth
					if !_rules[ruleNames]() {
						goto l306
					}
					goto l307
				l306:
					position, tokenIndex, depth = position306, tokenIndex306, depth306
				}
			l307:
				if buffer[position] != rune('|') {
					goto l304
				}
				position++
				depth--
				add(ruleParams, position305)
			}
			return true
		l304:
			position, tokenIndex, depth = position304, tokenIndex304, depth304
			return false
		},
		/* 82 StartParams <- <Action2> */
		func() bool {
			position308, tokenIndex308, depth308 := position, tokenIndex, depth
			{
				position309 := position
				depth++
				if !_rules[ruleAction2]() {
					goto l308
				}
				depth--
				add(ruleStartParams, position309)
			}
			return true
		l308:
			position, tokenIndex, depth = position308, tokenIndex308, depth308
			return false
		},
		/* 83 Names <- <(NextName (',' NextName)* VarParams?)> */
		func() bool {
			position310, tokenIndex310, depth310 := position, tokenIndex, depth
			{
				position311 := position
				depth++
				if !_rules[ruleNextName]() {
					goto l310
				}
			l312:
				{
					position313, tokenIndex313, depth313 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l313
					}
					position++
					if !_rules[ruleNextName]() {
						goto l313
					}
					goto l312
				l313:
					position, tokenIndex, depth = position313, tokenIndex313, depth313
				}
				{
					position314, tokenIndex314, depth314 := position, tokenIndex, depth
					if !_rules[ruleVarParams]() {
						goto l314
					}
					goto l315
				l314:
					position, tokenIndex, depth = position314, tokenIndex314, depth314
				}
			l315:
				depth--
				add(ruleNames, position311)
			}
			return true
		l310:
			position, tokenIndex, depth = position310, tokenIndex310, depth310
			return false
		},
		/* 84 NextName <- <(ws Name ws)> */
		func() bool {
			position316, tokenIndex316, depth316 := position, tokenIndex, depth
			{
				position317 := position
				depth++
				if !_rules[rulews]() {
					goto l316
				}
				if !_rules[ruleName]() {
					goto l316
				}
				if !_rules[rulews]() {
					goto l316
				}
				depth--
				add(ruleNextName, position317)
			}
			return true
		l316:
			position, tokenIndex, depth = position316, tokenIndex316, depth316
			return false
		},
		/* 85 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position318, tokenIndex318, depth318 := position, tokenIndex, depth
			{
				position319 := position
				depth++
				{
					position322, tokenIndex322, depth322 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l323
					}
					position++
					goto l322
				l323:
					position, tokenIndex, depth = position322, tokenIndex322, depth322
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l324
					}
					position++
					goto l322
				l324:
					position, tokenIndex, depth = position322, tokenIndex322, depth322
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l325
					}
					position++
					goto l322
				l325:
					position, tokenIndex, depth = position322, tokenIndex322, depth322
					if buffer[position] != rune('_') {
						goto l318
					}
					position++
				}
			l322:
			l320:
				{
					position321, tokenIndex321, depth321 := position, tokenIndex, depth
					{
						position326, tokenIndex326, depth326 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l327
						}
						position++
						goto l326
					l327:
						position, tokenIndex, depth = position326, tokenIndex326, depth326
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l328
						}
						position++
						goto l326
					l328:
						position, tokenIndex, depth = position326, tokenIndex326, depth326
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l329
						}
						position++
						goto l326
					l329:
						position, tokenIndex, depth = position326, tokenIndex326, depth326
						if buffer[position] != rune('_') {
							goto l321
						}
						position++
					}
				l326:
					goto l320
				l321:
					position, tokenIndex, depth = position321, tokenIndex321, depth321
				}
				depth--
				add(ruleName, position319)
			}
			return true
		l318:
			position, tokenIndex, depth = position318, tokenIndex318, depth318
			return false
		},
		/* 86 VarParams <- <('.' '.' '.' ws)> */
		func() bool {
			position330, tokenIndex330, depth330 := position, tokenIndex, depth
			{
				position331 := position
				depth++
				if buffer[position] != rune('.') {
					goto l330
				}
				position++
				if buffer[position] != rune('.') {
					goto l330
				}
				position++
				if buffer[position] != rune('.') {
					goto l330
				}
				position++
				if !_rules[rulews]() {
					goto l330
				}
				depth--
				add(ruleVarParams, position331)
			}
			return true
		l330:
			position, tokenIndex, depth = position330, tokenIndex330, depth330
			return false
		},
		/* 87 Reference <- <('.'? Key FollowUpRef)> */
		func() bool {
			position332, tokenIndex332, depth332 := position, tokenIndex, depth
			{
				position333 := position
				depth++
				{
					position334, tokenIndex334, depth334 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l334
					}
					position++
					goto l335
				l334:
					position, tokenIndex, depth = position334, tokenIndex334, depth334
				}
			l335:
				if !_rules[ruleKey]() {
					goto l332
				}
				if !_rules[ruleFollowUpRef]() {
					goto l332
				}
				depth--
				add(ruleReference, position333)
			}
			return true
		l332:
			position, tokenIndex, depth = position332, tokenIndex332, depth332
			return false
		},
		/* 88 FollowUpRef <- <PathComponent*> */
		func() bool {
			{
				position337 := position
				depth++
			l338:
				{
					position339, tokenIndex339, depth339 := position, tokenIndex, depth
					if !_rules[rulePathComponent]() {
						goto l339
					}
					goto l338
				l339:
					position, tokenIndex, depth = position339, tokenIndex339, depth339
				}
				depth--
				add(ruleFollowUpRef, position337)
			}
			return true
		},
		/* 89 PathComponent <- <(('.' Key) / ('.'? Index))> */
		func() bool {
			position340, tokenIndex340, depth340 := position, tokenIndex, depth
			{
				position341 := position
				depth++
				{
					position342, tokenIndex342, depth342 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l343
					}
					position++
					if !_rules[ruleKey]() {
						goto l343
					}
					goto l342
				l343:
					position, tokenIndex, depth = position342, tokenIndex342, depth342
					{
						position344, tokenIndex344, depth344 := position, tokenIndex, depth
						if buffer[position] != rune('.') {
							goto l344
						}
						position++
						goto l345
					l344:
						position, tokenIndex, depth = position344, tokenIndex344, depth344
					}
				l345:
					if !_rules[ruleIndex]() {
						goto l340
					}
				}
			l342:
				depth--
				add(rulePathComponent, position341)
			}
			return true
		l340:
			position, tokenIndex, depth = position340, tokenIndex340, depth340
			return false
		},
		/* 90 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position346, tokenIndex346, depth346 := position, tokenIndex, depth
			{
				position347 := position
				depth++
				{
					position348, tokenIndex348, depth348 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l349
					}
					position++
					goto l348
				l349:
					position, tokenIndex, depth = position348, tokenIndex348, depth348
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l350
					}
					position++
					goto l348
				l350:
					position, tokenIndex, depth = position348, tokenIndex348, depth348
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l351
					}
					position++
					goto l348
				l351:
					position, tokenIndex, depth = position348, tokenIndex348, depth348
					if buffer[position] != rune('_') {
						goto l346
					}
					position++
				}
			l348:
			l352:
				{
					position353, tokenIndex353, depth353 := position, tokenIndex, depth
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
							goto l358
						}
						position++
						goto l354
					l358:
						position, tokenIndex, depth = position354, tokenIndex354, depth354
						if buffer[position] != rune('-') {
							goto l353
						}
						position++
					}
				l354:
					goto l352
				l353:
					position, tokenIndex, depth = position353, tokenIndex353, depth353
				}
				{
					position359, tokenIndex359, depth359 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l359
					}
					position++
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
							goto l359
						}
						position++
					}
				l361:
				l365:
					{
						position366, tokenIndex366, depth366 := position, tokenIndex, depth
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
								goto l371
							}
							position++
							goto l367
						l371:
							position, tokenIndex, depth = position367, tokenIndex367, depth367
							if buffer[position] != rune('-') {
								goto l366
							}
							position++
						}
					l367:
						goto l365
					l366:
						position, tokenIndex, depth = position366, tokenIndex366, depth366
					}
					goto l360
				l359:
					position, tokenIndex, depth = position359, tokenIndex359, depth359
				}
			l360:
				depth--
				add(ruleKey, position347)
			}
			return true
		l346:
			position, tokenIndex, depth = position346, tokenIndex346, depth346
			return false
		},
		/* 91 Index <- <('[' '-'? [0-9]+ ']')> */
		func() bool {
			position372, tokenIndex372, depth372 := position, tokenIndex, depth
			{
				position373 := position
				depth++
				if buffer[position] != rune('[') {
					goto l372
				}
				position++
				{
					position374, tokenIndex374, depth374 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l374
					}
					position++
					goto l375
				l374:
					position, tokenIndex, depth = position374, tokenIndex374, depth374
				}
			l375:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l372
				}
				position++
			l376:
				{
					position377, tokenIndex377, depth377 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l377
					}
					position++
					goto l376
				l377:
					position, tokenIndex, depth = position377, tokenIndex377, depth377
				}
				if buffer[position] != rune(']') {
					goto l372
				}
				position++
				depth--
				add(ruleIndex, position373)
			}
			return true
		l372:
			position, tokenIndex, depth = position372, tokenIndex372, depth372
			return false
		},
		/* 92 IP <- <([0-9]+ '.' [0-9]+ '.' [0-9]+ '.' [0-9]+)> */
		func() bool {
			position378, tokenIndex378, depth378 := position, tokenIndex, depth
			{
				position379 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l378
				}
				position++
			l380:
				{
					position381, tokenIndex381, depth381 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l381
					}
					position++
					goto l380
				l381:
					position, tokenIndex, depth = position381, tokenIndex381, depth381
				}
				if buffer[position] != rune('.') {
					goto l378
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l378
				}
				position++
			l382:
				{
					position383, tokenIndex383, depth383 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l383
					}
					position++
					goto l382
				l383:
					position, tokenIndex, depth = position383, tokenIndex383, depth383
				}
				if buffer[position] != rune('.') {
					goto l378
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l378
				}
				position++
			l384:
				{
					position385, tokenIndex385, depth385 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l385
					}
					position++
					goto l384
				l385:
					position, tokenIndex, depth = position385, tokenIndex385, depth385
				}
				if buffer[position] != rune('.') {
					goto l378
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l378
				}
				position++
			l386:
				{
					position387, tokenIndex387, depth387 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l387
					}
					position++
					goto l386
				l387:
					position, tokenIndex, depth = position387, tokenIndex387, depth387
				}
				depth--
				add(ruleIP, position379)
			}
			return true
		l378:
			position, tokenIndex, depth = position378, tokenIndex378, depth378
			return false
		},
		/* 93 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position389 := position
				depth++
			l390:
				{
					position391, tokenIndex391, depth391 := position, tokenIndex, depth
					{
						position392, tokenIndex392, depth392 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l393
						}
						position++
						goto l392
					l393:
						position, tokenIndex, depth = position392, tokenIndex392, depth392
						if buffer[position] != rune('\t') {
							goto l394
						}
						position++
						goto l392
					l394:
						position, tokenIndex, depth = position392, tokenIndex392, depth392
						if buffer[position] != rune('\n') {
							goto l395
						}
						position++
						goto l392
					l395:
						position, tokenIndex, depth = position392, tokenIndex392, depth392
						if buffer[position] != rune('\r') {
							goto l391
						}
						position++
					}
				l392:
					goto l390
				l391:
					position, tokenIndex, depth = position391, tokenIndex391, depth391
				}
				depth--
				add(rulews, position389)
			}
			return true
		},
		/* 94 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position396, tokenIndex396, depth396 := position, tokenIndex, depth
			{
				position397 := position
				depth++
				{
					position400, tokenIndex400, depth400 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l401
					}
					position++
					goto l400
				l401:
					position, tokenIndex, depth = position400, tokenIndex400, depth400
					if buffer[position] != rune('\t') {
						goto l402
					}
					position++
					goto l400
				l402:
					position, tokenIndex, depth = position400, tokenIndex400, depth400
					if buffer[position] != rune('\n') {
						goto l403
					}
					position++
					goto l400
				l403:
					position, tokenIndex, depth = position400, tokenIndex400, depth400
					if buffer[position] != rune('\r') {
						goto l396
					}
					position++
				}
			l400:
			l398:
				{
					position399, tokenIndex399, depth399 := position, tokenIndex, depth
					{
						position404, tokenIndex404, depth404 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l405
						}
						position++
						goto l404
					l405:
						position, tokenIndex, depth = position404, tokenIndex404, depth404
						if buffer[position] != rune('\t') {
							goto l406
						}
						position++
						goto l404
					l406:
						position, tokenIndex, depth = position404, tokenIndex404, depth404
						if buffer[position] != rune('\n') {
							goto l407
						}
						position++
						goto l404
					l407:
						position, tokenIndex, depth = position404, tokenIndex404, depth404
						if buffer[position] != rune('\r') {
							goto l399
						}
						position++
					}
				l404:
					goto l398
				l399:
					position, tokenIndex, depth = position399, tokenIndex399, depth399
				}
				depth--
				add(rulereq_ws, position397)
			}
			return true
		l396:
			position, tokenIndex, depth = position396, tokenIndex396, depth396
			return false
		},
		/* 96 Action0 <- <{}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 97 Action1 <- <{}> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 98 Action2 <- <{}> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
	}
	p.rules = _rules
}
