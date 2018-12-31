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
	ruleInteger
	ruleString
	ruleBoolean
	ruleNil
	ruleUndefined
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
	ruleMapping
	ruleSum
	ruleLambda
	ruleLambdaRef
	ruleLambdaExpr
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
	"Integer",
	"String",
	"Boolean",
	"Nil",
	"Undefined",
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
	"Mapping",
	"Sum",
	"Lambda",
	"LambdaRef",
	"LambdaExpr",
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
	rules  [76]func() bool
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
		/* 4 Marker <- <('&' (('t' 'e' 'm' 'p' 'l' 'a' 't' 'e') / ('t' 'e' 'm' 'p' 'o' 'r' 'a' 'r' 'y') / ('l' 'o' 'c' 'a' 'l')))> */
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
						goto l16
					}
					position++
					if buffer[position] != rune('o') {
						goto l16
					}
					position++
					if buffer[position] != rune('c') {
						goto l16
					}
					position++
					if buffer[position] != rune('a') {
						goto l16
					}
					position++
					if buffer[position] != rune('l') {
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
			position21, tokenIndex21, depth21 := position, tokenIndex, depth
			{
				position22 := position
				depth++
				if !_rules[ruleGrouped]() {
					goto l21
				}
				depth--
				add(ruleMarkerExpression, position22)
			}
			return true
		l21:
			position, tokenIndex, depth = position21, tokenIndex21, depth21
			return false
		},
		/* 6 Expression <- <(ws (LambdaExpr / Level7) ws)> */
		func() bool {
			position23, tokenIndex23, depth23 := position, tokenIndex, depth
			{
				position24 := position
				depth++
				if !_rules[rulews]() {
					goto l23
				}
				{
					position25, tokenIndex25, depth25 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l26
					}
					goto l25
				l26:
					position, tokenIndex, depth = position25, tokenIndex25, depth25
					if !_rules[ruleLevel7]() {
						goto l23
					}
				}
			l25:
				if !_rules[rulews]() {
					goto l23
				}
				depth--
				add(ruleExpression, position24)
			}
			return true
		l23:
			position, tokenIndex, depth = position23, tokenIndex23, depth23
			return false
		},
		/* 7 Level7 <- <(Level6 (req_ws Or)*)> */
		func() bool {
			position27, tokenIndex27, depth27 := position, tokenIndex, depth
			{
				position28 := position
				depth++
				if !_rules[ruleLevel6]() {
					goto l27
				}
			l29:
				{
					position30, tokenIndex30, depth30 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l30
					}
					if !_rules[ruleOr]() {
						goto l30
					}
					goto l29
				l30:
					position, tokenIndex, depth = position30, tokenIndex30, depth30
				}
				depth--
				add(ruleLevel7, position28)
			}
			return true
		l27:
			position, tokenIndex, depth = position27, tokenIndex27, depth27
			return false
		},
		/* 8 Or <- <('|' '|' req_ws Level6)> */
		func() bool {
			position31, tokenIndex31, depth31 := position, tokenIndex, depth
			{
				position32 := position
				depth++
				if buffer[position] != rune('|') {
					goto l31
				}
				position++
				if buffer[position] != rune('|') {
					goto l31
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l31
				}
				if !_rules[ruleLevel6]() {
					goto l31
				}
				depth--
				add(ruleOr, position32)
			}
			return true
		l31:
			position, tokenIndex, depth = position31, tokenIndex31, depth31
			return false
		},
		/* 9 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position33, tokenIndex33, depth33 := position, tokenIndex, depth
			{
				position34 := position
				depth++
				{
					position35, tokenIndex35, depth35 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l36
					}
					goto l35
				l36:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
					if !_rules[ruleLevel5]() {
						goto l33
					}
				}
			l35:
				depth--
				add(ruleLevel6, position34)
			}
			return true
		l33:
			position, tokenIndex, depth = position33, tokenIndex33, depth33
			return false
		},
		/* 10 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position37, tokenIndex37, depth37 := position, tokenIndex, depth
			{
				position38 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l37
				}
				if !_rules[rulews]() {
					goto l37
				}
				if buffer[position] != rune('?') {
					goto l37
				}
				position++
				if !_rules[ruleExpression]() {
					goto l37
				}
				if buffer[position] != rune(':') {
					goto l37
				}
				position++
				if !_rules[ruleExpression]() {
					goto l37
				}
				depth--
				add(ruleConditional, position38)
			}
			return true
		l37:
			position, tokenIndex, depth = position37, tokenIndex37, depth37
			return false
		},
		/* 11 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position39, tokenIndex39, depth39 := position, tokenIndex, depth
			{
				position40 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l39
				}
			l41:
				{
					position42, tokenIndex42, depth42 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l42
					}
					goto l41
				l42:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
				}
				depth--
				add(ruleLevel5, position40)
			}
			return true
		l39:
			position, tokenIndex, depth = position39, tokenIndex39, depth39
			return false
		},
		/* 12 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position43, tokenIndex43, depth43 := position, tokenIndex, depth
			{
				position44 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l43
				}
				if !_rules[ruleLevel4]() {
					goto l43
				}
				depth--
				add(ruleConcatenation, position44)
			}
			return true
		l43:
			position, tokenIndex, depth = position43, tokenIndex43, depth43
			return false
		},
		/* 13 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position45, tokenIndex45, depth45 := position, tokenIndex, depth
			{
				position46 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l45
				}
			l47:
				{
					position48, tokenIndex48, depth48 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l48
					}
					{
						position49, tokenIndex49, depth49 := position, tokenIndex, depth
						if !_rules[ruleLogOr]() {
							goto l50
						}
						goto l49
					l50:
						position, tokenIndex, depth = position49, tokenIndex49, depth49
						if !_rules[ruleLogAnd]() {
							goto l48
						}
					}
				l49:
					goto l47
				l48:
					position, tokenIndex, depth = position48, tokenIndex48, depth48
				}
				depth--
				add(ruleLevel4, position46)
			}
			return true
		l45:
			position, tokenIndex, depth = position45, tokenIndex45, depth45
			return false
		},
		/* 14 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position51, tokenIndex51, depth51 := position, tokenIndex, depth
			{
				position52 := position
				depth++
				if buffer[position] != rune('-') {
					goto l51
				}
				position++
				if buffer[position] != rune('o') {
					goto l51
				}
				position++
				if buffer[position] != rune('r') {
					goto l51
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l51
				}
				if !_rules[ruleLevel3]() {
					goto l51
				}
				depth--
				add(ruleLogOr, position52)
			}
			return true
		l51:
			position, tokenIndex, depth = position51, tokenIndex51, depth51
			return false
		},
		/* 15 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position53, tokenIndex53, depth53 := position, tokenIndex, depth
			{
				position54 := position
				depth++
				if buffer[position] != rune('-') {
					goto l53
				}
				position++
				if buffer[position] != rune('a') {
					goto l53
				}
				position++
				if buffer[position] != rune('n') {
					goto l53
				}
				position++
				if buffer[position] != rune('d') {
					goto l53
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l53
				}
				if !_rules[ruleLevel3]() {
					goto l53
				}
				depth--
				add(ruleLogAnd, position54)
			}
			return true
		l53:
			position, tokenIndex, depth = position53, tokenIndex53, depth53
			return false
		},
		/* 16 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position55, tokenIndex55, depth55 := position, tokenIndex, depth
			{
				position56 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l55
				}
			l57:
				{
					position58, tokenIndex58, depth58 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l58
					}
					if !_rules[ruleComparison]() {
						goto l58
					}
					goto l57
				l58:
					position, tokenIndex, depth = position58, tokenIndex58, depth58
				}
				depth--
				add(ruleLevel3, position56)
			}
			return true
		l55:
			position, tokenIndex, depth = position55, tokenIndex55, depth55
			return false
		},
		/* 17 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position59, tokenIndex59, depth59 := position, tokenIndex, depth
			{
				position60 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l59
				}
				if !_rules[rulereq_ws]() {
					goto l59
				}
				if !_rules[ruleLevel2]() {
					goto l59
				}
				depth--
				add(ruleComparison, position60)
			}
			return true
		l59:
			position, tokenIndex, depth = position59, tokenIndex59, depth59
			return false
		},
		/* 18 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position61, tokenIndex61, depth61 := position, tokenIndex, depth
			{
				position62 := position
				depth++
				{
					position63, tokenIndex63, depth63 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l64
					}
					position++
					if buffer[position] != rune('=') {
						goto l64
					}
					position++
					goto l63
				l64:
					position, tokenIndex, depth = position63, tokenIndex63, depth63
					if buffer[position] != rune('!') {
						goto l65
					}
					position++
					if buffer[position] != rune('=') {
						goto l65
					}
					position++
					goto l63
				l65:
					position, tokenIndex, depth = position63, tokenIndex63, depth63
					if buffer[position] != rune('<') {
						goto l66
					}
					position++
					if buffer[position] != rune('=') {
						goto l66
					}
					position++
					goto l63
				l66:
					position, tokenIndex, depth = position63, tokenIndex63, depth63
					if buffer[position] != rune('>') {
						goto l67
					}
					position++
					if buffer[position] != rune('=') {
						goto l67
					}
					position++
					goto l63
				l67:
					position, tokenIndex, depth = position63, tokenIndex63, depth63
					if buffer[position] != rune('>') {
						goto l68
					}
					position++
					goto l63
				l68:
					position, tokenIndex, depth = position63, tokenIndex63, depth63
					if buffer[position] != rune('<') {
						goto l69
					}
					position++
					goto l63
				l69:
					position, tokenIndex, depth = position63, tokenIndex63, depth63
					if buffer[position] != rune('>') {
						goto l61
					}
					position++
				}
			l63:
				depth--
				add(ruleCompareOp, position62)
			}
			return true
		l61:
			position, tokenIndex, depth = position61, tokenIndex61, depth61
			return false
		},
		/* 19 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position70, tokenIndex70, depth70 := position, tokenIndex, depth
			{
				position71 := position
				depth++
				if !_rules[ruleLevel1]() {
					goto l70
				}
			l72:
				{
					position73, tokenIndex73, depth73 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l73
					}
					{
						position74, tokenIndex74, depth74 := position, tokenIndex, depth
						if !_rules[ruleAddition]() {
							goto l75
						}
						goto l74
					l75:
						position, tokenIndex, depth = position74, tokenIndex74, depth74
						if !_rules[ruleSubtraction]() {
							goto l73
						}
					}
				l74:
					goto l72
				l73:
					position, tokenIndex, depth = position73, tokenIndex73, depth73
				}
				depth--
				add(ruleLevel2, position71)
			}
			return true
		l70:
			position, tokenIndex, depth = position70, tokenIndex70, depth70
			return false
		},
		/* 20 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position76, tokenIndex76, depth76 := position, tokenIndex, depth
			{
				position77 := position
				depth++
				if buffer[position] != rune('+') {
					goto l76
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l76
				}
				if !_rules[ruleLevel1]() {
					goto l76
				}
				depth--
				add(ruleAddition, position77)
			}
			return true
		l76:
			position, tokenIndex, depth = position76, tokenIndex76, depth76
			return false
		},
		/* 21 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position78, tokenIndex78, depth78 := position, tokenIndex, depth
			{
				position79 := position
				depth++
				if buffer[position] != rune('-') {
					goto l78
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l78
				}
				if !_rules[ruleLevel1]() {
					goto l78
				}
				depth--
				add(ruleSubtraction, position79)
			}
			return true
		l78:
			position, tokenIndex, depth = position78, tokenIndex78, depth78
			return false
		},
		/* 22 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position80, tokenIndex80, depth80 := position, tokenIndex, depth
			{
				position81 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l80
				}
			l82:
				{
					position83, tokenIndex83, depth83 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l83
					}
					{
						position84, tokenIndex84, depth84 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l85
						}
						goto l84
					l85:
						position, tokenIndex, depth = position84, tokenIndex84, depth84
						if !_rules[ruleDivision]() {
							goto l86
						}
						goto l84
					l86:
						position, tokenIndex, depth = position84, tokenIndex84, depth84
						if !_rules[ruleModulo]() {
							goto l83
						}
					}
				l84:
					goto l82
				l83:
					position, tokenIndex, depth = position83, tokenIndex83, depth83
				}
				depth--
				add(ruleLevel1, position81)
			}
			return true
		l80:
			position, tokenIndex, depth = position80, tokenIndex80, depth80
			return false
		},
		/* 23 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position87, tokenIndex87, depth87 := position, tokenIndex, depth
			{
				position88 := position
				depth++
				if buffer[position] != rune('*') {
					goto l87
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l87
				}
				if !_rules[ruleLevel0]() {
					goto l87
				}
				depth--
				add(ruleMultiplication, position88)
			}
			return true
		l87:
			position, tokenIndex, depth = position87, tokenIndex87, depth87
			return false
		},
		/* 24 Division <- <('/' req_ws Level0)> */
		func() bool {
			position89, tokenIndex89, depth89 := position, tokenIndex, depth
			{
				position90 := position
				depth++
				if buffer[position] != rune('/') {
					goto l89
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l89
				}
				if !_rules[ruleLevel0]() {
					goto l89
				}
				depth--
				add(ruleDivision, position90)
			}
			return true
		l89:
			position, tokenIndex, depth = position89, tokenIndex89, depth89
			return false
		},
		/* 25 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position91, tokenIndex91, depth91 := position, tokenIndex, depth
			{
				position92 := position
				depth++
				if buffer[position] != rune('%') {
					goto l91
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l91
				}
				if !_rules[ruleLevel0]() {
					goto l91
				}
				depth--
				add(ruleModulo, position92)
			}
			return true
		l91:
			position, tokenIndex, depth = position91, tokenIndex91, depth91
			return false
		},
		/* 26 Level0 <- <(IP / String / Integer / Boolean / Undefined / Nil / Not / Substitution / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position93, tokenIndex93, depth93 := position, tokenIndex, depth
			{
				position94 := position
				depth++
				{
					position95, tokenIndex95, depth95 := position, tokenIndex, depth
					if !_rules[ruleIP]() {
						goto l96
					}
					goto l95
				l96:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if !_rules[ruleString]() {
						goto l97
					}
					goto l95
				l97:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if !_rules[ruleInteger]() {
						goto l98
					}
					goto l95
				l98:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if !_rules[ruleBoolean]() {
						goto l99
					}
					goto l95
				l99:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if !_rules[ruleUndefined]() {
						goto l100
					}
					goto l95
				l100:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if !_rules[ruleNil]() {
						goto l101
					}
					goto l95
				l101:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if !_rules[ruleNot]() {
						goto l102
					}
					goto l95
				l102:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if !_rules[ruleSubstitution]() {
						goto l103
					}
					goto l95
				l103:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if !_rules[ruleMerge]() {
						goto l104
					}
					goto l95
				l104:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if !_rules[ruleAuto]() {
						goto l105
					}
					goto l95
				l105:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if !_rules[ruleLambda]() {
						goto l106
					}
					goto l95
				l106:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if !_rules[ruleChained]() {
						goto l93
					}
				}
			l95:
				depth--
				add(ruleLevel0, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 27 Chained <- <((Mapping / Sum / List / Map / Range / Grouped / Reference) ChainedQualifiedExpression*)> */
		func() bool {
			position107, tokenIndex107, depth107 := position, tokenIndex, depth
			{
				position108 := position
				depth++
				{
					position109, tokenIndex109, depth109 := position, tokenIndex, depth
					if !_rules[ruleMapping]() {
						goto l110
					}
					goto l109
				l110:
					position, tokenIndex, depth = position109, tokenIndex109, depth109
					if !_rules[ruleSum]() {
						goto l111
					}
					goto l109
				l111:
					position, tokenIndex, depth = position109, tokenIndex109, depth109
					if !_rules[ruleList]() {
						goto l112
					}
					goto l109
				l112:
					position, tokenIndex, depth = position109, tokenIndex109, depth109
					if !_rules[ruleMap]() {
						goto l113
					}
					goto l109
				l113:
					position, tokenIndex, depth = position109, tokenIndex109, depth109
					if !_rules[ruleRange]() {
						goto l114
					}
					goto l109
				l114:
					position, tokenIndex, depth = position109, tokenIndex109, depth109
					if !_rules[ruleGrouped]() {
						goto l115
					}
					goto l109
				l115:
					position, tokenIndex, depth = position109, tokenIndex109, depth109
					if !_rules[ruleReference]() {
						goto l107
					}
				}
			l109:
			l116:
				{
					position117, tokenIndex117, depth117 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l117
					}
					goto l116
				l117:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
				}
				depth--
				add(ruleChained, position108)
			}
			return true
		l107:
			position, tokenIndex, depth = position107, tokenIndex107, depth107
			return false
		},
		/* 28 ChainedQualifiedExpression <- <(ChainedCall / ('.' (ChainedRef / ChainedDynRef / Projection)))> */
		func() bool {
			position118, tokenIndex118, depth118 := position, tokenIndex, depth
			{
				position119 := position
				depth++
				{
					position120, tokenIndex120, depth120 := position, tokenIndex, depth
					if !_rules[ruleChainedCall]() {
						goto l121
					}
					goto l120
				l121:
					position, tokenIndex, depth = position120, tokenIndex120, depth120
					if buffer[position] != rune('.') {
						goto l118
					}
					position++
					{
						position122, tokenIndex122, depth122 := position, tokenIndex, depth
						if !_rules[ruleChainedRef]() {
							goto l123
						}
						goto l122
					l123:
						position, tokenIndex, depth = position122, tokenIndex122, depth122
						if !_rules[ruleChainedDynRef]() {
							goto l124
						}
						goto l122
					l124:
						position, tokenIndex, depth = position122, tokenIndex122, depth122
						if !_rules[ruleProjection]() {
							goto l118
						}
					}
				l122:
				}
			l120:
				depth--
				add(ruleChainedQualifiedExpression, position119)
			}
			return true
		l118:
			position, tokenIndex, depth = position118, tokenIndex118, depth118
			return false
		},
		/* 29 ChainedRef <- <((Key / Index) FollowUpRef)> */
		func() bool {
			position125, tokenIndex125, depth125 := position, tokenIndex, depth
			{
				position126 := position
				depth++
				{
					position127, tokenIndex127, depth127 := position, tokenIndex, depth
					if !_rules[ruleKey]() {
						goto l128
					}
					goto l127
				l128:
					position, tokenIndex, depth = position127, tokenIndex127, depth127
					if !_rules[ruleIndex]() {
						goto l125
					}
				}
			l127:
				if !_rules[ruleFollowUpRef]() {
					goto l125
				}
				depth--
				add(ruleChainedRef, position126)
			}
			return true
		l125:
			position, tokenIndex, depth = position125, tokenIndex125, depth125
			return false
		},
		/* 30 ChainedDynRef <- <('[' Expression ']')> */
		func() bool {
			position129, tokenIndex129, depth129 := position, tokenIndex, depth
			{
				position130 := position
				depth++
				if buffer[position] != rune('[') {
					goto l129
				}
				position++
				if !_rules[ruleExpression]() {
					goto l129
				}
				if buffer[position] != rune(']') {
					goto l129
				}
				position++
				depth--
				add(ruleChainedDynRef, position130)
			}
			return true
		l129:
			position, tokenIndex, depth = position129, tokenIndex129, depth129
			return false
		},
		/* 31 Slice <- <Range> */
		func() bool {
			position131, tokenIndex131, depth131 := position, tokenIndex, depth
			{
				position132 := position
				depth++
				if !_rules[ruleRange]() {
					goto l131
				}
				depth--
				add(ruleSlice, position132)
			}
			return true
		l131:
			position, tokenIndex, depth = position131, tokenIndex131, depth131
			return false
		},
		/* 32 ChainedCall <- <(StartArguments ExpressionList? ')')> */
		func() bool {
			position133, tokenIndex133, depth133 := position, tokenIndex, depth
			{
				position134 := position
				depth++
				if !_rules[ruleStartArguments]() {
					goto l133
				}
				{
					position135, tokenIndex135, depth135 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l135
					}
					goto l136
				l135:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
				}
			l136:
				if buffer[position] != rune(')') {
					goto l133
				}
				position++
				depth--
				add(ruleChainedCall, position134)
			}
			return true
		l133:
			position, tokenIndex, depth = position133, tokenIndex133, depth133
			return false
		},
		/* 33 StartArguments <- <('(' ws)> */
		func() bool {
			position137, tokenIndex137, depth137 := position, tokenIndex, depth
			{
				position138 := position
				depth++
				if buffer[position] != rune('(') {
					goto l137
				}
				position++
				if !_rules[rulews]() {
					goto l137
				}
				depth--
				add(ruleStartArguments, position138)
			}
			return true
		l137:
			position, tokenIndex, depth = position137, tokenIndex137, depth137
			return false
		},
		/* 34 ExpressionList <- <(NextExpression (',' NextExpression)*)> */
		func() bool {
			position139, tokenIndex139, depth139 := position, tokenIndex, depth
			{
				position140 := position
				depth++
				if !_rules[ruleNextExpression]() {
					goto l139
				}
			l141:
				{
					position142, tokenIndex142, depth142 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l142
					}
					position++
					if !_rules[ruleNextExpression]() {
						goto l142
					}
					goto l141
				l142:
					position, tokenIndex, depth = position142, tokenIndex142, depth142
				}
				depth--
				add(ruleExpressionList, position140)
			}
			return true
		l139:
			position, tokenIndex, depth = position139, tokenIndex139, depth139
			return false
		},
		/* 35 NextExpression <- <Expression> */
		func() bool {
			position143, tokenIndex143, depth143 := position, tokenIndex, depth
			{
				position144 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l143
				}
				depth--
				add(ruleNextExpression, position144)
			}
			return true
		l143:
			position, tokenIndex, depth = position143, tokenIndex143, depth143
			return false
		},
		/* 36 Projection <- <((('[' '*' ']') / Slice) ProjectionValue ChainedQualifiedExpression*)> */
		func() bool {
			position145, tokenIndex145, depth145 := position, tokenIndex, depth
			{
				position146 := position
				depth++
				{
					position147, tokenIndex147, depth147 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l148
					}
					position++
					if buffer[position] != rune('*') {
						goto l148
					}
					position++
					if buffer[position] != rune(']') {
						goto l148
					}
					position++
					goto l147
				l148:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
					if !_rules[ruleSlice]() {
						goto l145
					}
				}
			l147:
				if !_rules[ruleProjectionValue]() {
					goto l145
				}
			l149:
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l150
					}
					goto l149
				l150:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
				}
				depth--
				add(ruleProjection, position146)
			}
			return true
		l145:
			position, tokenIndex, depth = position145, tokenIndex145, depth145
			return false
		},
		/* 37 ProjectionValue <- <Action0> */
		func() bool {
			position151, tokenIndex151, depth151 := position, tokenIndex, depth
			{
				position152 := position
				depth++
				if !_rules[ruleAction0]() {
					goto l151
				}
				depth--
				add(ruleProjectionValue, position152)
			}
			return true
		l151:
			position, tokenIndex, depth = position151, tokenIndex151, depth151
			return false
		},
		/* 38 Substitution <- <('*' Level0)> */
		func() bool {
			position153, tokenIndex153, depth153 := position, tokenIndex, depth
			{
				position154 := position
				depth++
				if buffer[position] != rune('*') {
					goto l153
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l153
				}
				depth--
				add(ruleSubstitution, position154)
			}
			return true
		l153:
			position, tokenIndex, depth = position153, tokenIndex153, depth153
			return false
		},
		/* 39 Not <- <('!' ws Level0)> */
		func() bool {
			position155, tokenIndex155, depth155 := position, tokenIndex, depth
			{
				position156 := position
				depth++
				if buffer[position] != rune('!') {
					goto l155
				}
				position++
				if !_rules[rulews]() {
					goto l155
				}
				if !_rules[ruleLevel0]() {
					goto l155
				}
				depth--
				add(ruleNot, position156)
			}
			return true
		l155:
			position, tokenIndex, depth = position155, tokenIndex155, depth155
			return false
		},
		/* 40 Grouped <- <('(' Expression ')')> */
		func() bool {
			position157, tokenIndex157, depth157 := position, tokenIndex, depth
			{
				position158 := position
				depth++
				if buffer[position] != rune('(') {
					goto l157
				}
				position++
				if !_rules[ruleExpression]() {
					goto l157
				}
				if buffer[position] != rune(')') {
					goto l157
				}
				position++
				depth--
				add(ruleGrouped, position158)
			}
			return true
		l157:
			position, tokenIndex, depth = position157, tokenIndex157, depth157
			return false
		},
		/* 41 Range <- <('[' Expression ('.' '.') Expression ']')> */
		func() bool {
			position159, tokenIndex159, depth159 := position, tokenIndex, depth
			{
				position160 := position
				depth++
				if buffer[position] != rune('[') {
					goto l159
				}
				position++
				if !_rules[ruleExpression]() {
					goto l159
				}
				if buffer[position] != rune('.') {
					goto l159
				}
				position++
				if buffer[position] != rune('.') {
					goto l159
				}
				position++
				if !_rules[ruleExpression]() {
					goto l159
				}
				if buffer[position] != rune(']') {
					goto l159
				}
				position++
				depth--
				add(ruleRange, position160)
			}
			return true
		l159:
			position, tokenIndex, depth = position159, tokenIndex159, depth159
			return false
		},
		/* 42 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position161, tokenIndex161, depth161 := position, tokenIndex, depth
			{
				position162 := position
				depth++
				{
					position163, tokenIndex163, depth163 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l163
					}
					position++
					goto l164
				l163:
					position, tokenIndex, depth = position163, tokenIndex163, depth163
				}
			l164:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l161
				}
				position++
			l165:
				{
					position166, tokenIndex166, depth166 := position, tokenIndex, depth
					{
						position167, tokenIndex167, depth167 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l168
						}
						position++
						goto l167
					l168:
						position, tokenIndex, depth = position167, tokenIndex167, depth167
						if buffer[position] != rune('_') {
							goto l166
						}
						position++
					}
				l167:
					goto l165
				l166:
					position, tokenIndex, depth = position166, tokenIndex166, depth166
				}
				depth--
				add(ruleInteger, position162)
			}
			return true
		l161:
			position, tokenIndex, depth = position161, tokenIndex161, depth161
			return false
		},
		/* 43 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position169, tokenIndex169, depth169 := position, tokenIndex, depth
			{
				position170 := position
				depth++
				if buffer[position] != rune('"') {
					goto l169
				}
				position++
			l171:
				{
					position172, tokenIndex172, depth172 := position, tokenIndex, depth
					{
						position173, tokenIndex173, depth173 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l174
						}
						position++
						if buffer[position] != rune('"') {
							goto l174
						}
						position++
						goto l173
					l174:
						position, tokenIndex, depth = position173, tokenIndex173, depth173
						{
							position175, tokenIndex175, depth175 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l175
							}
							position++
							goto l172
						l175:
							position, tokenIndex, depth = position175, tokenIndex175, depth175
						}
						if !matchDot() {
							goto l172
						}
					}
				l173:
					goto l171
				l172:
					position, tokenIndex, depth = position172, tokenIndex172, depth172
				}
				if buffer[position] != rune('"') {
					goto l169
				}
				position++
				depth--
				add(ruleString, position170)
			}
			return true
		l169:
			position, tokenIndex, depth = position169, tokenIndex169, depth169
			return false
		},
		/* 44 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position176, tokenIndex176, depth176 := position, tokenIndex, depth
			{
				position177 := position
				depth++
				{
					position178, tokenIndex178, depth178 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l179
					}
					position++
					if buffer[position] != rune('r') {
						goto l179
					}
					position++
					if buffer[position] != rune('u') {
						goto l179
					}
					position++
					if buffer[position] != rune('e') {
						goto l179
					}
					position++
					goto l178
				l179:
					position, tokenIndex, depth = position178, tokenIndex178, depth178
					if buffer[position] != rune('f') {
						goto l176
					}
					position++
					if buffer[position] != rune('a') {
						goto l176
					}
					position++
					if buffer[position] != rune('l') {
						goto l176
					}
					position++
					if buffer[position] != rune('s') {
						goto l176
					}
					position++
					if buffer[position] != rune('e') {
						goto l176
					}
					position++
				}
			l178:
				depth--
				add(ruleBoolean, position177)
			}
			return true
		l176:
			position, tokenIndex, depth = position176, tokenIndex176, depth176
			return false
		},
		/* 45 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position180, tokenIndex180, depth180 := position, tokenIndex, depth
			{
				position181 := position
				depth++
				{
					position182, tokenIndex182, depth182 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l183
					}
					position++
					if buffer[position] != rune('i') {
						goto l183
					}
					position++
					if buffer[position] != rune('l') {
						goto l183
					}
					position++
					goto l182
				l183:
					position, tokenIndex, depth = position182, tokenIndex182, depth182
					if buffer[position] != rune('~') {
						goto l180
					}
					position++
				}
			l182:
				depth--
				add(ruleNil, position181)
			}
			return true
		l180:
			position, tokenIndex, depth = position180, tokenIndex180, depth180
			return false
		},
		/* 46 Undefined <- <('~' '~')> */
		func() bool {
			position184, tokenIndex184, depth184 := position, tokenIndex, depth
			{
				position185 := position
				depth++
				if buffer[position] != rune('~') {
					goto l184
				}
				position++
				if buffer[position] != rune('~') {
					goto l184
				}
				position++
				depth--
				add(ruleUndefined, position185)
			}
			return true
		l184:
			position, tokenIndex, depth = position184, tokenIndex184, depth184
			return false
		},
		/* 47 List <- <(StartList ExpressionList? ']')> */
		func() bool {
			position186, tokenIndex186, depth186 := position, tokenIndex, depth
			{
				position187 := position
				depth++
				if !_rules[ruleStartList]() {
					goto l186
				}
				{
					position188, tokenIndex188, depth188 := position, tokenIndex, depth
					if !_rules[ruleExpressionList]() {
						goto l188
					}
					goto l189
				l188:
					position, tokenIndex, depth = position188, tokenIndex188, depth188
				}
			l189:
				if buffer[position] != rune(']') {
					goto l186
				}
				position++
				depth--
				add(ruleList, position187)
			}
			return true
		l186:
			position, tokenIndex, depth = position186, tokenIndex186, depth186
			return false
		},
		/* 48 StartList <- <'['> */
		func() bool {
			position190, tokenIndex190, depth190 := position, tokenIndex, depth
			{
				position191 := position
				depth++
				if buffer[position] != rune('[') {
					goto l190
				}
				position++
				depth--
				add(ruleStartList, position191)
			}
			return true
		l190:
			position, tokenIndex, depth = position190, tokenIndex190, depth190
			return false
		},
		/* 49 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position192, tokenIndex192, depth192 := position, tokenIndex, depth
			{
				position193 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l192
				}
				if !_rules[rulews]() {
					goto l192
				}
				{
					position194, tokenIndex194, depth194 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l194
					}
					goto l195
				l194:
					position, tokenIndex, depth = position194, tokenIndex194, depth194
				}
			l195:
				if buffer[position] != rune('}') {
					goto l192
				}
				position++
				depth--
				add(ruleMap, position193)
			}
			return true
		l192:
			position, tokenIndex, depth = position192, tokenIndex192, depth192
			return false
		},
		/* 50 CreateMap <- <'{'> */
		func() bool {
			position196, tokenIndex196, depth196 := position, tokenIndex, depth
			{
				position197 := position
				depth++
				if buffer[position] != rune('{') {
					goto l196
				}
				position++
				depth--
				add(ruleCreateMap, position197)
			}
			return true
		l196:
			position, tokenIndex, depth = position196, tokenIndex196, depth196
			return false
		},
		/* 51 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position198, tokenIndex198, depth198 := position, tokenIndex, depth
			{
				position199 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l198
				}
			l200:
				{
					position201, tokenIndex201, depth201 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l201
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l201
					}
					goto l200
				l201:
					position, tokenIndex, depth = position201, tokenIndex201, depth201
				}
				depth--
				add(ruleAssignments, position199)
			}
			return true
		l198:
			position, tokenIndex, depth = position198, tokenIndex198, depth198
			return false
		},
		/* 52 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position202, tokenIndex202, depth202 := position, tokenIndex, depth
			{
				position203 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l202
				}
				if buffer[position] != rune('=') {
					goto l202
				}
				position++
				if !_rules[ruleExpression]() {
					goto l202
				}
				depth--
				add(ruleAssignment, position203)
			}
			return true
		l202:
			position, tokenIndex, depth = position202, tokenIndex202, depth202
			return false
		},
		/* 53 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position204, tokenIndex204, depth204 := position, tokenIndex, depth
			{
				position205 := position
				depth++
				{
					position206, tokenIndex206, depth206 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l207
					}
					goto l206
				l207:
					position, tokenIndex, depth = position206, tokenIndex206, depth206
					if !_rules[ruleSimpleMerge]() {
						goto l204
					}
				}
			l206:
				depth--
				add(ruleMerge, position205)
			}
			return true
		l204:
			position, tokenIndex, depth = position204, tokenIndex204, depth204
			return false
		},
		/* 54 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position208, tokenIndex208, depth208 := position, tokenIndex, depth
			{
				position209 := position
				depth++
				if buffer[position] != rune('m') {
					goto l208
				}
				position++
				if buffer[position] != rune('e') {
					goto l208
				}
				position++
				if buffer[position] != rune('r') {
					goto l208
				}
				position++
				if buffer[position] != rune('g') {
					goto l208
				}
				position++
				if buffer[position] != rune('e') {
					goto l208
				}
				position++
				{
					position210, tokenIndex210, depth210 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l210
					}
					if !_rules[ruleRequired]() {
						goto l210
					}
					goto l208
				l210:
					position, tokenIndex, depth = position210, tokenIndex210, depth210
				}
				{
					position211, tokenIndex211, depth211 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l211
					}
					{
						position213, tokenIndex213, depth213 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l214
						}
						goto l213
					l214:
						position, tokenIndex, depth = position213, tokenIndex213, depth213
						if !_rules[ruleOn]() {
							goto l211
						}
					}
				l213:
					goto l212
				l211:
					position, tokenIndex, depth = position211, tokenIndex211, depth211
				}
			l212:
				if !_rules[rulereq_ws]() {
					goto l208
				}
				if !_rules[ruleReference]() {
					goto l208
				}
				depth--
				add(ruleRefMerge, position209)
			}
			return true
		l208:
			position, tokenIndex, depth = position208, tokenIndex208, depth208
			return false
		},
		/* 55 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' !'(' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position215, tokenIndex215, depth215 := position, tokenIndex, depth
			{
				position216 := position
				depth++
				if buffer[position] != rune('m') {
					goto l215
				}
				position++
				if buffer[position] != rune('e') {
					goto l215
				}
				position++
				if buffer[position] != rune('r') {
					goto l215
				}
				position++
				if buffer[position] != rune('g') {
					goto l215
				}
				position++
				if buffer[position] != rune('e') {
					goto l215
				}
				position++
				{
					position217, tokenIndex217, depth217 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l217
					}
					position++
					goto l215
				l217:
					position, tokenIndex, depth = position217, tokenIndex217, depth217
				}
				{
					position218, tokenIndex218, depth218 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l218
					}
					{
						position220, tokenIndex220, depth220 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l221
						}
						goto l220
					l221:
						position, tokenIndex, depth = position220, tokenIndex220, depth220
						if !_rules[ruleRequired]() {
							goto l222
						}
						goto l220
					l222:
						position, tokenIndex, depth = position220, tokenIndex220, depth220
						if !_rules[ruleOn]() {
							goto l218
						}
					}
				l220:
					goto l219
				l218:
					position, tokenIndex, depth = position218, tokenIndex218, depth218
				}
			l219:
				depth--
				add(ruleSimpleMerge, position216)
			}
			return true
		l215:
			position, tokenIndex, depth = position215, tokenIndex215, depth215
			return false
		},
		/* 56 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position223, tokenIndex223, depth223 := position, tokenIndex, depth
			{
				position224 := position
				depth++
				if buffer[position] != rune('r') {
					goto l223
				}
				position++
				if buffer[position] != rune('e') {
					goto l223
				}
				position++
				if buffer[position] != rune('p') {
					goto l223
				}
				position++
				if buffer[position] != rune('l') {
					goto l223
				}
				position++
				if buffer[position] != rune('a') {
					goto l223
				}
				position++
				if buffer[position] != rune('c') {
					goto l223
				}
				position++
				if buffer[position] != rune('e') {
					goto l223
				}
				position++
				depth--
				add(ruleReplace, position224)
			}
			return true
		l223:
			position, tokenIndex, depth = position223, tokenIndex223, depth223
			return false
		},
		/* 57 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position225, tokenIndex225, depth225 := position, tokenIndex, depth
			{
				position226 := position
				depth++
				if buffer[position] != rune('r') {
					goto l225
				}
				position++
				if buffer[position] != rune('e') {
					goto l225
				}
				position++
				if buffer[position] != rune('q') {
					goto l225
				}
				position++
				if buffer[position] != rune('u') {
					goto l225
				}
				position++
				if buffer[position] != rune('i') {
					goto l225
				}
				position++
				if buffer[position] != rune('r') {
					goto l225
				}
				position++
				if buffer[position] != rune('e') {
					goto l225
				}
				position++
				if buffer[position] != rune('d') {
					goto l225
				}
				position++
				depth--
				add(ruleRequired, position226)
			}
			return true
		l225:
			position, tokenIndex, depth = position225, tokenIndex225, depth225
			return false
		},
		/* 58 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position227, tokenIndex227, depth227 := position, tokenIndex, depth
			{
				position228 := position
				depth++
				if buffer[position] != rune('o') {
					goto l227
				}
				position++
				if buffer[position] != rune('n') {
					goto l227
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l227
				}
				if !_rules[ruleName]() {
					goto l227
				}
				depth--
				add(ruleOn, position228)
			}
			return true
		l227:
			position, tokenIndex, depth = position227, tokenIndex227, depth227
			return false
		},
		/* 59 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position229, tokenIndex229, depth229 := position, tokenIndex, depth
			{
				position230 := position
				depth++
				if buffer[position] != rune('a') {
					goto l229
				}
				position++
				if buffer[position] != rune('u') {
					goto l229
				}
				position++
				if buffer[position] != rune('t') {
					goto l229
				}
				position++
				if buffer[position] != rune('o') {
					goto l229
				}
				position++
				depth--
				add(ruleAuto, position230)
			}
			return true
		l229:
			position, tokenIndex, depth = position229, tokenIndex229, depth229
			return false
		},
		/* 60 Mapping <- <('m' 'a' 'p' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position231, tokenIndex231, depth231 := position, tokenIndex, depth
			{
				position232 := position
				depth++
				if buffer[position] != rune('m') {
					goto l231
				}
				position++
				if buffer[position] != rune('a') {
					goto l231
				}
				position++
				if buffer[position] != rune('p') {
					goto l231
				}
				position++
				if buffer[position] != rune('[') {
					goto l231
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l231
				}
				{
					position233, tokenIndex233, depth233 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l234
					}
					goto l233
				l234:
					position, tokenIndex, depth = position233, tokenIndex233, depth233
					if buffer[position] != rune('|') {
						goto l231
					}
					position++
					if !_rules[ruleExpression]() {
						goto l231
					}
				}
			l233:
				if buffer[position] != rune(']') {
					goto l231
				}
				position++
				depth--
				add(ruleMapping, position232)
			}
			return true
		l231:
			position, tokenIndex, depth = position231, tokenIndex231, depth231
			return false
		},
		/* 61 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position235, tokenIndex235, depth235 := position, tokenIndex, depth
			{
				position236 := position
				depth++
				if buffer[position] != rune('s') {
					goto l235
				}
				position++
				if buffer[position] != rune('u') {
					goto l235
				}
				position++
				if buffer[position] != rune('m') {
					goto l235
				}
				position++
				if buffer[position] != rune('[') {
					goto l235
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l235
				}
				if buffer[position] != rune('|') {
					goto l235
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l235
				}
				{
					position237, tokenIndex237, depth237 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l238
					}
					goto l237
				l238:
					position, tokenIndex, depth = position237, tokenIndex237, depth237
					if buffer[position] != rune('|') {
						goto l235
					}
					position++
					if !_rules[ruleExpression]() {
						goto l235
					}
				}
			l237:
				if buffer[position] != rune(']') {
					goto l235
				}
				position++
				depth--
				add(ruleSum, position236)
			}
			return true
		l235:
			position, tokenIndex, depth = position235, tokenIndex235, depth235
			return false
		},
		/* 62 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position239, tokenIndex239, depth239 := position, tokenIndex, depth
			{
				position240 := position
				depth++
				if buffer[position] != rune('l') {
					goto l239
				}
				position++
				if buffer[position] != rune('a') {
					goto l239
				}
				position++
				if buffer[position] != rune('m') {
					goto l239
				}
				position++
				if buffer[position] != rune('b') {
					goto l239
				}
				position++
				if buffer[position] != rune('d') {
					goto l239
				}
				position++
				if buffer[position] != rune('a') {
					goto l239
				}
				position++
				{
					position241, tokenIndex241, depth241 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l242
					}
					goto l241
				l242:
					position, tokenIndex, depth = position241, tokenIndex241, depth241
					if !_rules[ruleLambdaExpr]() {
						goto l239
					}
				}
			l241:
				depth--
				add(ruleLambda, position240)
			}
			return true
		l239:
			position, tokenIndex, depth = position239, tokenIndex239, depth239
			return false
		},
		/* 63 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position243, tokenIndex243, depth243 := position, tokenIndex, depth
			{
				position244 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l243
				}
				if !_rules[ruleExpression]() {
					goto l243
				}
				depth--
				add(ruleLambdaRef, position244)
			}
			return true
		l243:
			position, tokenIndex, depth = position243, tokenIndex243, depth243
			return false
		},
		/* 64 LambdaExpr <- <(ws '|' ws Name NextName* ws '|' ws ('-' '>') Expression)> */
		func() bool {
			position245, tokenIndex245, depth245 := position, tokenIndex, depth
			{
				position246 := position
				depth++
				if !_rules[rulews]() {
					goto l245
				}
				if buffer[position] != rune('|') {
					goto l245
				}
				position++
				if !_rules[rulews]() {
					goto l245
				}
				if !_rules[ruleName]() {
					goto l245
				}
			l247:
				{
					position248, tokenIndex248, depth248 := position, tokenIndex, depth
					if !_rules[ruleNextName]() {
						goto l248
					}
					goto l247
				l248:
					position, tokenIndex, depth = position248, tokenIndex248, depth248
				}
				if !_rules[rulews]() {
					goto l245
				}
				if buffer[position] != rune('|') {
					goto l245
				}
				position++
				if !_rules[rulews]() {
					goto l245
				}
				if buffer[position] != rune('-') {
					goto l245
				}
				position++
				if buffer[position] != rune('>') {
					goto l245
				}
				position++
				if !_rules[ruleExpression]() {
					goto l245
				}
				depth--
				add(ruleLambdaExpr, position246)
			}
			return true
		l245:
			position, tokenIndex, depth = position245, tokenIndex245, depth245
			return false
		},
		/* 65 NextName <- <(ws ',' ws Name)> */
		func() bool {
			position249, tokenIndex249, depth249 := position, tokenIndex, depth
			{
				position250 := position
				depth++
				if !_rules[rulews]() {
					goto l249
				}
				if buffer[position] != rune(',') {
					goto l249
				}
				position++
				if !_rules[rulews]() {
					goto l249
				}
				if !_rules[ruleName]() {
					goto l249
				}
				depth--
				add(ruleNextName, position250)
			}
			return true
		l249:
			position, tokenIndex, depth = position249, tokenIndex249, depth249
			return false
		},
		/* 66 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position251, tokenIndex251, depth251 := position, tokenIndex, depth
			{
				position252 := position
				depth++
				{
					position255, tokenIndex255, depth255 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l256
					}
					position++
					goto l255
				l256:
					position, tokenIndex, depth = position255, tokenIndex255, depth255
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l257
					}
					position++
					goto l255
				l257:
					position, tokenIndex, depth = position255, tokenIndex255, depth255
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l258
					}
					position++
					goto l255
				l258:
					position, tokenIndex, depth = position255, tokenIndex255, depth255
					if buffer[position] != rune('_') {
						goto l251
					}
					position++
				}
			l255:
			l253:
				{
					position254, tokenIndex254, depth254 := position, tokenIndex, depth
					{
						position259, tokenIndex259, depth259 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l260
						}
						position++
						goto l259
					l260:
						position, tokenIndex, depth = position259, tokenIndex259, depth259
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l261
						}
						position++
						goto l259
					l261:
						position, tokenIndex, depth = position259, tokenIndex259, depth259
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l262
						}
						position++
						goto l259
					l262:
						position, tokenIndex, depth = position259, tokenIndex259, depth259
						if buffer[position] != rune('_') {
							goto l254
						}
						position++
					}
				l259:
					goto l253
				l254:
					position, tokenIndex, depth = position254, tokenIndex254, depth254
				}
				depth--
				add(ruleName, position252)
			}
			return true
		l251:
			position, tokenIndex, depth = position251, tokenIndex251, depth251
			return false
		},
		/* 67 Reference <- <('.'? Key FollowUpRef)> */
		func() bool {
			position263, tokenIndex263, depth263 := position, tokenIndex, depth
			{
				position264 := position
				depth++
				{
					position265, tokenIndex265, depth265 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l265
					}
					position++
					goto l266
				l265:
					position, tokenIndex, depth = position265, tokenIndex265, depth265
				}
			l266:
				if !_rules[ruleKey]() {
					goto l263
				}
				if !_rules[ruleFollowUpRef]() {
					goto l263
				}
				depth--
				add(ruleReference, position264)
			}
			return true
		l263:
			position, tokenIndex, depth = position263, tokenIndex263, depth263
			return false
		},
		/* 68 FollowUpRef <- <('.' (Key / Index))*> */
		func() bool {
			{
				position268 := position
				depth++
			l269:
				{
					position270, tokenIndex270, depth270 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l270
					}
					position++
					{
						position271, tokenIndex271, depth271 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l272
						}
						goto l271
					l272:
						position, tokenIndex, depth = position271, tokenIndex271, depth271
						if !_rules[ruleIndex]() {
							goto l270
						}
					}
				l271:
					goto l269
				l270:
					position, tokenIndex, depth = position270, tokenIndex270, depth270
				}
				depth--
				add(ruleFollowUpRef, position268)
			}
			return true
		},
		/* 69 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position273, tokenIndex273, depth273 := position, tokenIndex, depth
			{
				position274 := position
				depth++
				{
					position275, tokenIndex275, depth275 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l276
					}
					position++
					goto l275
				l276:
					position, tokenIndex, depth = position275, tokenIndex275, depth275
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l277
					}
					position++
					goto l275
				l277:
					position, tokenIndex, depth = position275, tokenIndex275, depth275
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l278
					}
					position++
					goto l275
				l278:
					position, tokenIndex, depth = position275, tokenIndex275, depth275
					if buffer[position] != rune('_') {
						goto l273
					}
					position++
				}
			l275:
			l279:
				{
					position280, tokenIndex280, depth280 := position, tokenIndex, depth
					{
						position281, tokenIndex281, depth281 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l282
						}
						position++
						goto l281
					l282:
						position, tokenIndex, depth = position281, tokenIndex281, depth281
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l283
						}
						position++
						goto l281
					l283:
						position, tokenIndex, depth = position281, tokenIndex281, depth281
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l284
						}
						position++
						goto l281
					l284:
						position, tokenIndex, depth = position281, tokenIndex281, depth281
						if buffer[position] != rune('_') {
							goto l285
						}
						position++
						goto l281
					l285:
						position, tokenIndex, depth = position281, tokenIndex281, depth281
						if buffer[position] != rune('-') {
							goto l280
						}
						position++
					}
				l281:
					goto l279
				l280:
					position, tokenIndex, depth = position280, tokenIndex280, depth280
				}
				{
					position286, tokenIndex286, depth286 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l286
					}
					position++
					{
						position288, tokenIndex288, depth288 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l289
						}
						position++
						goto l288
					l289:
						position, tokenIndex, depth = position288, tokenIndex288, depth288
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l290
						}
						position++
						goto l288
					l290:
						position, tokenIndex, depth = position288, tokenIndex288, depth288
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l291
						}
						position++
						goto l288
					l291:
						position, tokenIndex, depth = position288, tokenIndex288, depth288
						if buffer[position] != rune('_') {
							goto l286
						}
						position++
					}
				l288:
				l292:
					{
						position293, tokenIndex293, depth293 := position, tokenIndex, depth
						{
							position294, tokenIndex294, depth294 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l295
							}
							position++
							goto l294
						l295:
							position, tokenIndex, depth = position294, tokenIndex294, depth294
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l296
							}
							position++
							goto l294
						l296:
							position, tokenIndex, depth = position294, tokenIndex294, depth294
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l297
							}
							position++
							goto l294
						l297:
							position, tokenIndex, depth = position294, tokenIndex294, depth294
							if buffer[position] != rune('_') {
								goto l298
							}
							position++
							goto l294
						l298:
							position, tokenIndex, depth = position294, tokenIndex294, depth294
							if buffer[position] != rune('-') {
								goto l293
							}
							position++
						}
					l294:
						goto l292
					l293:
						position, tokenIndex, depth = position293, tokenIndex293, depth293
					}
					goto l287
				l286:
					position, tokenIndex, depth = position286, tokenIndex286, depth286
				}
			l287:
				depth--
				add(ruleKey, position274)
			}
			return true
		l273:
			position, tokenIndex, depth = position273, tokenIndex273, depth273
			return false
		},
		/* 70 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position299, tokenIndex299, depth299 := position, tokenIndex, depth
			{
				position300 := position
				depth++
				if buffer[position] != rune('[') {
					goto l299
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l299
				}
				position++
			l301:
				{
					position302, tokenIndex302, depth302 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l302
					}
					position++
					goto l301
				l302:
					position, tokenIndex, depth = position302, tokenIndex302, depth302
				}
				if buffer[position] != rune(']') {
					goto l299
				}
				position++
				depth--
				add(ruleIndex, position300)
			}
			return true
		l299:
			position, tokenIndex, depth = position299, tokenIndex299, depth299
			return false
		},
		/* 71 IP <- <([0-9]+ '.' [0-9]+ '.' [0-9]+ '.' [0-9]+)> */
		func() bool {
			position303, tokenIndex303, depth303 := position, tokenIndex, depth
			{
				position304 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l303
				}
				position++
			l305:
				{
					position306, tokenIndex306, depth306 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l306
					}
					position++
					goto l305
				l306:
					position, tokenIndex, depth = position306, tokenIndex306, depth306
				}
				if buffer[position] != rune('.') {
					goto l303
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l303
				}
				position++
			l307:
				{
					position308, tokenIndex308, depth308 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l308
					}
					position++
					goto l307
				l308:
					position, tokenIndex, depth = position308, tokenIndex308, depth308
				}
				if buffer[position] != rune('.') {
					goto l303
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l303
				}
				position++
			l309:
				{
					position310, tokenIndex310, depth310 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l310
					}
					position++
					goto l309
				l310:
					position, tokenIndex, depth = position310, tokenIndex310, depth310
				}
				if buffer[position] != rune('.') {
					goto l303
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l303
				}
				position++
			l311:
				{
					position312, tokenIndex312, depth312 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l312
					}
					position++
					goto l311
				l312:
					position, tokenIndex, depth = position312, tokenIndex312, depth312
				}
				depth--
				add(ruleIP, position304)
			}
			return true
		l303:
			position, tokenIndex, depth = position303, tokenIndex303, depth303
			return false
		},
		/* 72 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position314 := position
				depth++
			l315:
				{
					position316, tokenIndex316, depth316 := position, tokenIndex, depth
					{
						position317, tokenIndex317, depth317 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l318
						}
						position++
						goto l317
					l318:
						position, tokenIndex, depth = position317, tokenIndex317, depth317
						if buffer[position] != rune('\t') {
							goto l319
						}
						position++
						goto l317
					l319:
						position, tokenIndex, depth = position317, tokenIndex317, depth317
						if buffer[position] != rune('\n') {
							goto l320
						}
						position++
						goto l317
					l320:
						position, tokenIndex, depth = position317, tokenIndex317, depth317
						if buffer[position] != rune('\r') {
							goto l316
						}
						position++
					}
				l317:
					goto l315
				l316:
					position, tokenIndex, depth = position316, tokenIndex316, depth316
				}
				depth--
				add(rulews, position314)
			}
			return true
		},
		/* 73 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position321, tokenIndex321, depth321 := position, tokenIndex, depth
			{
				position322 := position
				depth++
				{
					position325, tokenIndex325, depth325 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l326
					}
					position++
					goto l325
				l326:
					position, tokenIndex, depth = position325, tokenIndex325, depth325
					if buffer[position] != rune('\t') {
						goto l327
					}
					position++
					goto l325
				l327:
					position, tokenIndex, depth = position325, tokenIndex325, depth325
					if buffer[position] != rune('\n') {
						goto l328
					}
					position++
					goto l325
				l328:
					position, tokenIndex, depth = position325, tokenIndex325, depth325
					if buffer[position] != rune('\r') {
						goto l321
					}
					position++
				}
			l325:
			l323:
				{
					position324, tokenIndex324, depth324 := position, tokenIndex, depth
					{
						position329, tokenIndex329, depth329 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l330
						}
						position++
						goto l329
					l330:
						position, tokenIndex, depth = position329, tokenIndex329, depth329
						if buffer[position] != rune('\t') {
							goto l331
						}
						position++
						goto l329
					l331:
						position, tokenIndex, depth = position329, tokenIndex329, depth329
						if buffer[position] != rune('\n') {
							goto l332
						}
						position++
						goto l329
					l332:
						position, tokenIndex, depth = position329, tokenIndex329, depth329
						if buffer[position] != rune('\r') {
							goto l324
						}
						position++
					}
				l329:
					goto l323
				l324:
					position, tokenIndex, depth = position324, tokenIndex324, depth324
				}
				depth--
				add(rulereq_ws, position322)
			}
			return true
		l321:
			position, tokenIndex, depth = position321, tokenIndex321, depth321
			return false
		},
		/* 75 Action0 <- <{}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
	}
	p.rules = _rules
}
