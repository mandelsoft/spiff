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
	ruleArguments
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
	ruleContents
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
	"Arguments",
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
	"Contents",
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
	rules  [75]func() bool
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
		/* 32 ChainedCall <- <('(' Arguments ')')> */
		func() bool {
			position133, tokenIndex133, depth133 := position, tokenIndex, depth
			{
				position134 := position
				depth++
				if buffer[position] != rune('(') {
					goto l133
				}
				position++
				if !_rules[ruleArguments]() {
					goto l133
				}
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
		/* 33 Arguments <- <(Expression NextExpression*)> */
		func() bool {
			position135, tokenIndex135, depth135 := position, tokenIndex, depth
			{
				position136 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l135
				}
			l137:
				{
					position138, tokenIndex138, depth138 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l138
					}
					goto l137
				l138:
					position, tokenIndex, depth = position138, tokenIndex138, depth138
				}
				depth--
				add(ruleArguments, position136)
			}
			return true
		l135:
			position, tokenIndex, depth = position135, tokenIndex135, depth135
			return false
		},
		/* 34 NextExpression <- <(',' Expression)> */
		func() bool {
			position139, tokenIndex139, depth139 := position, tokenIndex, depth
			{
				position140 := position
				depth++
				if buffer[position] != rune(',') {
					goto l139
				}
				position++
				if !_rules[ruleExpression]() {
					goto l139
				}
				depth--
				add(ruleNextExpression, position140)
			}
			return true
		l139:
			position, tokenIndex, depth = position139, tokenIndex139, depth139
			return false
		},
		/* 35 Projection <- <((('[' '*' ']') / Slice) ProjectionValue ChainedQualifiedExpression*)> */
		func() bool {
			position141, tokenIndex141, depth141 := position, tokenIndex, depth
			{
				position142 := position
				depth++
				{
					position143, tokenIndex143, depth143 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l144
					}
					position++
					if buffer[position] != rune('*') {
						goto l144
					}
					position++
					if buffer[position] != rune(']') {
						goto l144
					}
					position++
					goto l143
				l144:
					position, tokenIndex, depth = position143, tokenIndex143, depth143
					if !_rules[ruleSlice]() {
						goto l141
					}
				}
			l143:
				if !_rules[ruleProjectionValue]() {
					goto l141
				}
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
				add(ruleProjection, position142)
			}
			return true
		l141:
			position, tokenIndex, depth = position141, tokenIndex141, depth141
			return false
		},
		/* 36 ProjectionValue <- <Action0> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				if !_rules[ruleAction0]() {
					goto l147
				}
				depth--
				add(ruleProjectionValue, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 37 Substitution <- <('*' Level0)> */
		func() bool {
			position149, tokenIndex149, depth149 := position, tokenIndex, depth
			{
				position150 := position
				depth++
				if buffer[position] != rune('*') {
					goto l149
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l149
				}
				depth--
				add(ruleSubstitution, position150)
			}
			return true
		l149:
			position, tokenIndex, depth = position149, tokenIndex149, depth149
			return false
		},
		/* 38 Not <- <('!' ws Level0)> */
		func() bool {
			position151, tokenIndex151, depth151 := position, tokenIndex, depth
			{
				position152 := position
				depth++
				if buffer[position] != rune('!') {
					goto l151
				}
				position++
				if !_rules[rulews]() {
					goto l151
				}
				if !_rules[ruleLevel0]() {
					goto l151
				}
				depth--
				add(ruleNot, position152)
			}
			return true
		l151:
			position, tokenIndex, depth = position151, tokenIndex151, depth151
			return false
		},
		/* 39 Grouped <- <('(' Expression ')')> */
		func() bool {
			position153, tokenIndex153, depth153 := position, tokenIndex, depth
			{
				position154 := position
				depth++
				if buffer[position] != rune('(') {
					goto l153
				}
				position++
				if !_rules[ruleExpression]() {
					goto l153
				}
				if buffer[position] != rune(')') {
					goto l153
				}
				position++
				depth--
				add(ruleGrouped, position154)
			}
			return true
		l153:
			position, tokenIndex, depth = position153, tokenIndex153, depth153
			return false
		},
		/* 40 Range <- <('[' Expression ('.' '.') Expression ']')> */
		func() bool {
			position155, tokenIndex155, depth155 := position, tokenIndex, depth
			{
				position156 := position
				depth++
				if buffer[position] != rune('[') {
					goto l155
				}
				position++
				if !_rules[ruleExpression]() {
					goto l155
				}
				if buffer[position] != rune('.') {
					goto l155
				}
				position++
				if buffer[position] != rune('.') {
					goto l155
				}
				position++
				if !_rules[ruleExpression]() {
					goto l155
				}
				if buffer[position] != rune(']') {
					goto l155
				}
				position++
				depth--
				add(ruleRange, position156)
			}
			return true
		l155:
			position, tokenIndex, depth = position155, tokenIndex155, depth155
			return false
		},
		/* 41 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position157, tokenIndex157, depth157 := position, tokenIndex, depth
			{
				position158 := position
				depth++
				{
					position159, tokenIndex159, depth159 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l159
					}
					position++
					goto l160
				l159:
					position, tokenIndex, depth = position159, tokenIndex159, depth159
				}
			l160:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l157
				}
				position++
			l161:
				{
					position162, tokenIndex162, depth162 := position, tokenIndex, depth
					{
						position163, tokenIndex163, depth163 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l164
						}
						position++
						goto l163
					l164:
						position, tokenIndex, depth = position163, tokenIndex163, depth163
						if buffer[position] != rune('_') {
							goto l162
						}
						position++
					}
				l163:
					goto l161
				l162:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
				}
				depth--
				add(ruleInteger, position158)
			}
			return true
		l157:
			position, tokenIndex, depth = position157, tokenIndex157, depth157
			return false
		},
		/* 42 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position165, tokenIndex165, depth165 := position, tokenIndex, depth
			{
				position166 := position
				depth++
				if buffer[position] != rune('"') {
					goto l165
				}
				position++
			l167:
				{
					position168, tokenIndex168, depth168 := position, tokenIndex, depth
					{
						position169, tokenIndex169, depth169 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l170
						}
						position++
						if buffer[position] != rune('"') {
							goto l170
						}
						position++
						goto l169
					l170:
						position, tokenIndex, depth = position169, tokenIndex169, depth169
						{
							position171, tokenIndex171, depth171 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l171
							}
							position++
							goto l168
						l171:
							position, tokenIndex, depth = position171, tokenIndex171, depth171
						}
						if !matchDot() {
							goto l168
						}
					}
				l169:
					goto l167
				l168:
					position, tokenIndex, depth = position168, tokenIndex168, depth168
				}
				if buffer[position] != rune('"') {
					goto l165
				}
				position++
				depth--
				add(ruleString, position166)
			}
			return true
		l165:
			position, tokenIndex, depth = position165, tokenIndex165, depth165
			return false
		},
		/* 43 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position172, tokenIndex172, depth172 := position, tokenIndex, depth
			{
				position173 := position
				depth++
				{
					position174, tokenIndex174, depth174 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l175
					}
					position++
					if buffer[position] != rune('r') {
						goto l175
					}
					position++
					if buffer[position] != rune('u') {
						goto l175
					}
					position++
					if buffer[position] != rune('e') {
						goto l175
					}
					position++
					goto l174
				l175:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
					if buffer[position] != rune('f') {
						goto l172
					}
					position++
					if buffer[position] != rune('a') {
						goto l172
					}
					position++
					if buffer[position] != rune('l') {
						goto l172
					}
					position++
					if buffer[position] != rune('s') {
						goto l172
					}
					position++
					if buffer[position] != rune('e') {
						goto l172
					}
					position++
				}
			l174:
				depth--
				add(ruleBoolean, position173)
			}
			return true
		l172:
			position, tokenIndex, depth = position172, tokenIndex172, depth172
			return false
		},
		/* 44 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position176, tokenIndex176, depth176 := position, tokenIndex, depth
			{
				position177 := position
				depth++
				{
					position178, tokenIndex178, depth178 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l179
					}
					position++
					if buffer[position] != rune('i') {
						goto l179
					}
					position++
					if buffer[position] != rune('l') {
						goto l179
					}
					position++
					goto l178
				l179:
					position, tokenIndex, depth = position178, tokenIndex178, depth178
					if buffer[position] != rune('~') {
						goto l176
					}
					position++
				}
			l178:
				depth--
				add(ruleNil, position177)
			}
			return true
		l176:
			position, tokenIndex, depth = position176, tokenIndex176, depth176
			return false
		},
		/* 45 Undefined <- <('~' '~')> */
		func() bool {
			position180, tokenIndex180, depth180 := position, tokenIndex, depth
			{
				position181 := position
				depth++
				if buffer[position] != rune('~') {
					goto l180
				}
				position++
				if buffer[position] != rune('~') {
					goto l180
				}
				position++
				depth--
				add(ruleUndefined, position181)
			}
			return true
		l180:
			position, tokenIndex, depth = position180, tokenIndex180, depth180
			return false
		},
		/* 46 List <- <('[' Contents? ']')> */
		func() bool {
			position182, tokenIndex182, depth182 := position, tokenIndex, depth
			{
				position183 := position
				depth++
				if buffer[position] != rune('[') {
					goto l182
				}
				position++
				{
					position184, tokenIndex184, depth184 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l184
					}
					goto l185
				l184:
					position, tokenIndex, depth = position184, tokenIndex184, depth184
				}
			l185:
				if buffer[position] != rune(']') {
					goto l182
				}
				position++
				depth--
				add(ruleList, position183)
			}
			return true
		l182:
			position, tokenIndex, depth = position182, tokenIndex182, depth182
			return false
		},
		/* 47 Contents <- <(Expression NextExpression*)> */
		func() bool {
			position186, tokenIndex186, depth186 := position, tokenIndex, depth
			{
				position187 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l186
				}
			l188:
				{
					position189, tokenIndex189, depth189 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l189
					}
					goto l188
				l189:
					position, tokenIndex, depth = position189, tokenIndex189, depth189
				}
				depth--
				add(ruleContents, position187)
			}
			return true
		l186:
			position, tokenIndex, depth = position186, tokenIndex186, depth186
			return false
		},
		/* 48 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position190, tokenIndex190, depth190 := position, tokenIndex, depth
			{
				position191 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l190
				}
				if !_rules[rulews]() {
					goto l190
				}
				{
					position192, tokenIndex192, depth192 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l192
					}
					goto l193
				l192:
					position, tokenIndex, depth = position192, tokenIndex192, depth192
				}
			l193:
				if buffer[position] != rune('}') {
					goto l190
				}
				position++
				depth--
				add(ruleMap, position191)
			}
			return true
		l190:
			position, tokenIndex, depth = position190, tokenIndex190, depth190
			return false
		},
		/* 49 CreateMap <- <'{'> */
		func() bool {
			position194, tokenIndex194, depth194 := position, tokenIndex, depth
			{
				position195 := position
				depth++
				if buffer[position] != rune('{') {
					goto l194
				}
				position++
				depth--
				add(ruleCreateMap, position195)
			}
			return true
		l194:
			position, tokenIndex, depth = position194, tokenIndex194, depth194
			return false
		},
		/* 50 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position196, tokenIndex196, depth196 := position, tokenIndex, depth
			{
				position197 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l196
				}
			l198:
				{
					position199, tokenIndex199, depth199 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l199
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l199
					}
					goto l198
				l199:
					position, tokenIndex, depth = position199, tokenIndex199, depth199
				}
				depth--
				add(ruleAssignments, position197)
			}
			return true
		l196:
			position, tokenIndex, depth = position196, tokenIndex196, depth196
			return false
		},
		/* 51 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position200, tokenIndex200, depth200 := position, tokenIndex, depth
			{
				position201 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l200
				}
				if buffer[position] != rune('=') {
					goto l200
				}
				position++
				if !_rules[ruleExpression]() {
					goto l200
				}
				depth--
				add(ruleAssignment, position201)
			}
			return true
		l200:
			position, tokenIndex, depth = position200, tokenIndex200, depth200
			return false
		},
		/* 52 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position202, tokenIndex202, depth202 := position, tokenIndex, depth
			{
				position203 := position
				depth++
				{
					position204, tokenIndex204, depth204 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l205
					}
					goto l204
				l205:
					position, tokenIndex, depth = position204, tokenIndex204, depth204
					if !_rules[ruleSimpleMerge]() {
						goto l202
					}
				}
			l204:
				depth--
				add(ruleMerge, position203)
			}
			return true
		l202:
			position, tokenIndex, depth = position202, tokenIndex202, depth202
			return false
		},
		/* 53 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position206, tokenIndex206, depth206 := position, tokenIndex, depth
			{
				position207 := position
				depth++
				if buffer[position] != rune('m') {
					goto l206
				}
				position++
				if buffer[position] != rune('e') {
					goto l206
				}
				position++
				if buffer[position] != rune('r') {
					goto l206
				}
				position++
				if buffer[position] != rune('g') {
					goto l206
				}
				position++
				if buffer[position] != rune('e') {
					goto l206
				}
				position++
				{
					position208, tokenIndex208, depth208 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l208
					}
					if !_rules[ruleRequired]() {
						goto l208
					}
					goto l206
				l208:
					position, tokenIndex, depth = position208, tokenIndex208, depth208
				}
				{
					position209, tokenIndex209, depth209 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l209
					}
					{
						position211, tokenIndex211, depth211 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l212
						}
						goto l211
					l212:
						position, tokenIndex, depth = position211, tokenIndex211, depth211
						if !_rules[ruleOn]() {
							goto l209
						}
					}
				l211:
					goto l210
				l209:
					position, tokenIndex, depth = position209, tokenIndex209, depth209
				}
			l210:
				if !_rules[rulereq_ws]() {
					goto l206
				}
				if !_rules[ruleReference]() {
					goto l206
				}
				depth--
				add(ruleRefMerge, position207)
			}
			return true
		l206:
			position, tokenIndex, depth = position206, tokenIndex206, depth206
			return false
		},
		/* 54 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' !'(' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position213, tokenIndex213, depth213 := position, tokenIndex, depth
			{
				position214 := position
				depth++
				if buffer[position] != rune('m') {
					goto l213
				}
				position++
				if buffer[position] != rune('e') {
					goto l213
				}
				position++
				if buffer[position] != rune('r') {
					goto l213
				}
				position++
				if buffer[position] != rune('g') {
					goto l213
				}
				position++
				if buffer[position] != rune('e') {
					goto l213
				}
				position++
				{
					position215, tokenIndex215, depth215 := position, tokenIndex, depth
					if buffer[position] != rune('(') {
						goto l215
					}
					position++
					goto l213
				l215:
					position, tokenIndex, depth = position215, tokenIndex215, depth215
				}
				{
					position216, tokenIndex216, depth216 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l216
					}
					{
						position218, tokenIndex218, depth218 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l219
						}
						goto l218
					l219:
						position, tokenIndex, depth = position218, tokenIndex218, depth218
						if !_rules[ruleRequired]() {
							goto l220
						}
						goto l218
					l220:
						position, tokenIndex, depth = position218, tokenIndex218, depth218
						if !_rules[ruleOn]() {
							goto l216
						}
					}
				l218:
					goto l217
				l216:
					position, tokenIndex, depth = position216, tokenIndex216, depth216
				}
			l217:
				depth--
				add(ruleSimpleMerge, position214)
			}
			return true
		l213:
			position, tokenIndex, depth = position213, tokenIndex213, depth213
			return false
		},
		/* 55 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position221, tokenIndex221, depth221 := position, tokenIndex, depth
			{
				position222 := position
				depth++
				if buffer[position] != rune('r') {
					goto l221
				}
				position++
				if buffer[position] != rune('e') {
					goto l221
				}
				position++
				if buffer[position] != rune('p') {
					goto l221
				}
				position++
				if buffer[position] != rune('l') {
					goto l221
				}
				position++
				if buffer[position] != rune('a') {
					goto l221
				}
				position++
				if buffer[position] != rune('c') {
					goto l221
				}
				position++
				if buffer[position] != rune('e') {
					goto l221
				}
				position++
				depth--
				add(ruleReplace, position222)
			}
			return true
		l221:
			position, tokenIndex, depth = position221, tokenIndex221, depth221
			return false
		},
		/* 56 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
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
				if buffer[position] != rune('q') {
					goto l223
				}
				position++
				if buffer[position] != rune('u') {
					goto l223
				}
				position++
				if buffer[position] != rune('i') {
					goto l223
				}
				position++
				if buffer[position] != rune('r') {
					goto l223
				}
				position++
				if buffer[position] != rune('e') {
					goto l223
				}
				position++
				if buffer[position] != rune('d') {
					goto l223
				}
				position++
				depth--
				add(ruleRequired, position224)
			}
			return true
		l223:
			position, tokenIndex, depth = position223, tokenIndex223, depth223
			return false
		},
		/* 57 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position225, tokenIndex225, depth225 := position, tokenIndex, depth
			{
				position226 := position
				depth++
				if buffer[position] != rune('o') {
					goto l225
				}
				position++
				if buffer[position] != rune('n') {
					goto l225
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l225
				}
				if !_rules[ruleName]() {
					goto l225
				}
				depth--
				add(ruleOn, position226)
			}
			return true
		l225:
			position, tokenIndex, depth = position225, tokenIndex225, depth225
			return false
		},
		/* 58 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position227, tokenIndex227, depth227 := position, tokenIndex, depth
			{
				position228 := position
				depth++
				if buffer[position] != rune('a') {
					goto l227
				}
				position++
				if buffer[position] != rune('u') {
					goto l227
				}
				position++
				if buffer[position] != rune('t') {
					goto l227
				}
				position++
				if buffer[position] != rune('o') {
					goto l227
				}
				position++
				depth--
				add(ruleAuto, position228)
			}
			return true
		l227:
			position, tokenIndex, depth = position227, tokenIndex227, depth227
			return false
		},
		/* 59 Mapping <- <('m' 'a' 'p' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position229, tokenIndex229, depth229 := position, tokenIndex, depth
			{
				position230 := position
				depth++
				if buffer[position] != rune('m') {
					goto l229
				}
				position++
				if buffer[position] != rune('a') {
					goto l229
				}
				position++
				if buffer[position] != rune('p') {
					goto l229
				}
				position++
				if buffer[position] != rune('[') {
					goto l229
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l229
				}
				{
					position231, tokenIndex231, depth231 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l232
					}
					goto l231
				l232:
					position, tokenIndex, depth = position231, tokenIndex231, depth231
					if buffer[position] != rune('|') {
						goto l229
					}
					position++
					if !_rules[ruleExpression]() {
						goto l229
					}
				}
			l231:
				if buffer[position] != rune(']') {
					goto l229
				}
				position++
				depth--
				add(ruleMapping, position230)
			}
			return true
		l229:
			position, tokenIndex, depth = position229, tokenIndex229, depth229
			return false
		},
		/* 60 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position233, tokenIndex233, depth233 := position, tokenIndex, depth
			{
				position234 := position
				depth++
				if buffer[position] != rune('s') {
					goto l233
				}
				position++
				if buffer[position] != rune('u') {
					goto l233
				}
				position++
				if buffer[position] != rune('m') {
					goto l233
				}
				position++
				if buffer[position] != rune('[') {
					goto l233
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l233
				}
				if buffer[position] != rune('|') {
					goto l233
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l233
				}
				{
					position235, tokenIndex235, depth235 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l236
					}
					goto l235
				l236:
					position, tokenIndex, depth = position235, tokenIndex235, depth235
					if buffer[position] != rune('|') {
						goto l233
					}
					position++
					if !_rules[ruleExpression]() {
						goto l233
					}
				}
			l235:
				if buffer[position] != rune(']') {
					goto l233
				}
				position++
				depth--
				add(ruleSum, position234)
			}
			return true
		l233:
			position, tokenIndex, depth = position233, tokenIndex233, depth233
			return false
		},
		/* 61 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position237, tokenIndex237, depth237 := position, tokenIndex, depth
			{
				position238 := position
				depth++
				if buffer[position] != rune('l') {
					goto l237
				}
				position++
				if buffer[position] != rune('a') {
					goto l237
				}
				position++
				if buffer[position] != rune('m') {
					goto l237
				}
				position++
				if buffer[position] != rune('b') {
					goto l237
				}
				position++
				if buffer[position] != rune('d') {
					goto l237
				}
				position++
				if buffer[position] != rune('a') {
					goto l237
				}
				position++
				{
					position239, tokenIndex239, depth239 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l240
					}
					goto l239
				l240:
					position, tokenIndex, depth = position239, tokenIndex239, depth239
					if !_rules[ruleLambdaExpr]() {
						goto l237
					}
				}
			l239:
				depth--
				add(ruleLambda, position238)
			}
			return true
		l237:
			position, tokenIndex, depth = position237, tokenIndex237, depth237
			return false
		},
		/* 62 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position241, tokenIndex241, depth241 := position, tokenIndex, depth
			{
				position242 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l241
				}
				if !_rules[ruleExpression]() {
					goto l241
				}
				depth--
				add(ruleLambdaRef, position242)
			}
			return true
		l241:
			position, tokenIndex, depth = position241, tokenIndex241, depth241
			return false
		},
		/* 63 LambdaExpr <- <(ws '|' ws Name NextName* ws '|' ws ('-' '>') Expression)> */
		func() bool {
			position243, tokenIndex243, depth243 := position, tokenIndex, depth
			{
				position244 := position
				depth++
				if !_rules[rulews]() {
					goto l243
				}
				if buffer[position] != rune('|') {
					goto l243
				}
				position++
				if !_rules[rulews]() {
					goto l243
				}
				if !_rules[ruleName]() {
					goto l243
				}
			l245:
				{
					position246, tokenIndex246, depth246 := position, tokenIndex, depth
					if !_rules[ruleNextName]() {
						goto l246
					}
					goto l245
				l246:
					position, tokenIndex, depth = position246, tokenIndex246, depth246
				}
				if !_rules[rulews]() {
					goto l243
				}
				if buffer[position] != rune('|') {
					goto l243
				}
				position++
				if !_rules[rulews]() {
					goto l243
				}
				if buffer[position] != rune('-') {
					goto l243
				}
				position++
				if buffer[position] != rune('>') {
					goto l243
				}
				position++
				if !_rules[ruleExpression]() {
					goto l243
				}
				depth--
				add(ruleLambdaExpr, position244)
			}
			return true
		l243:
			position, tokenIndex, depth = position243, tokenIndex243, depth243
			return false
		},
		/* 64 NextName <- <(ws ',' ws Name)> */
		func() bool {
			position247, tokenIndex247, depth247 := position, tokenIndex, depth
			{
				position248 := position
				depth++
				if !_rules[rulews]() {
					goto l247
				}
				if buffer[position] != rune(',') {
					goto l247
				}
				position++
				if !_rules[rulews]() {
					goto l247
				}
				if !_rules[ruleName]() {
					goto l247
				}
				depth--
				add(ruleNextName, position248)
			}
			return true
		l247:
			position, tokenIndex, depth = position247, tokenIndex247, depth247
			return false
		},
		/* 65 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position249, tokenIndex249, depth249 := position, tokenIndex, depth
			{
				position250 := position
				depth++
				{
					position253, tokenIndex253, depth253 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l254
					}
					position++
					goto l253
				l254:
					position, tokenIndex, depth = position253, tokenIndex253, depth253
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l255
					}
					position++
					goto l253
				l255:
					position, tokenIndex, depth = position253, tokenIndex253, depth253
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l256
					}
					position++
					goto l253
				l256:
					position, tokenIndex, depth = position253, tokenIndex253, depth253
					if buffer[position] != rune('_') {
						goto l249
					}
					position++
				}
			l253:
			l251:
				{
					position252, tokenIndex252, depth252 := position, tokenIndex, depth
					{
						position257, tokenIndex257, depth257 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l258
						}
						position++
						goto l257
					l258:
						position, tokenIndex, depth = position257, tokenIndex257, depth257
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l259
						}
						position++
						goto l257
					l259:
						position, tokenIndex, depth = position257, tokenIndex257, depth257
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l260
						}
						position++
						goto l257
					l260:
						position, tokenIndex, depth = position257, tokenIndex257, depth257
						if buffer[position] != rune('_') {
							goto l252
						}
						position++
					}
				l257:
					goto l251
				l252:
					position, tokenIndex, depth = position252, tokenIndex252, depth252
				}
				depth--
				add(ruleName, position250)
			}
			return true
		l249:
			position, tokenIndex, depth = position249, tokenIndex249, depth249
			return false
		},
		/* 66 Reference <- <('.'? Key FollowUpRef)> */
		func() bool {
			position261, tokenIndex261, depth261 := position, tokenIndex, depth
			{
				position262 := position
				depth++
				{
					position263, tokenIndex263, depth263 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l263
					}
					position++
					goto l264
				l263:
					position, tokenIndex, depth = position263, tokenIndex263, depth263
				}
			l264:
				if !_rules[ruleKey]() {
					goto l261
				}
				if !_rules[ruleFollowUpRef]() {
					goto l261
				}
				depth--
				add(ruleReference, position262)
			}
			return true
		l261:
			position, tokenIndex, depth = position261, tokenIndex261, depth261
			return false
		},
		/* 67 FollowUpRef <- <('.' (Key / Index))*> */
		func() bool {
			{
				position266 := position
				depth++
			l267:
				{
					position268, tokenIndex268, depth268 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l268
					}
					position++
					{
						position269, tokenIndex269, depth269 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l270
						}
						goto l269
					l270:
						position, tokenIndex, depth = position269, tokenIndex269, depth269
						if !_rules[ruleIndex]() {
							goto l268
						}
					}
				l269:
					goto l267
				l268:
					position, tokenIndex, depth = position268, tokenIndex268, depth268
				}
				depth--
				add(ruleFollowUpRef, position266)
			}
			return true
		},
		/* 68 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position271, tokenIndex271, depth271 := position, tokenIndex, depth
			{
				position272 := position
				depth++
				{
					position273, tokenIndex273, depth273 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l274
					}
					position++
					goto l273
				l274:
					position, tokenIndex, depth = position273, tokenIndex273, depth273
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l275
					}
					position++
					goto l273
				l275:
					position, tokenIndex, depth = position273, tokenIndex273, depth273
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l276
					}
					position++
					goto l273
				l276:
					position, tokenIndex, depth = position273, tokenIndex273, depth273
					if buffer[position] != rune('_') {
						goto l271
					}
					position++
				}
			l273:
			l277:
				{
					position278, tokenIndex278, depth278 := position, tokenIndex, depth
					{
						position279, tokenIndex279, depth279 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l280
						}
						position++
						goto l279
					l280:
						position, tokenIndex, depth = position279, tokenIndex279, depth279
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l281
						}
						position++
						goto l279
					l281:
						position, tokenIndex, depth = position279, tokenIndex279, depth279
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l282
						}
						position++
						goto l279
					l282:
						position, tokenIndex, depth = position279, tokenIndex279, depth279
						if buffer[position] != rune('_') {
							goto l283
						}
						position++
						goto l279
					l283:
						position, tokenIndex, depth = position279, tokenIndex279, depth279
						if buffer[position] != rune('-') {
							goto l278
						}
						position++
					}
				l279:
					goto l277
				l278:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
				}
				{
					position284, tokenIndex284, depth284 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l284
					}
					position++
					{
						position286, tokenIndex286, depth286 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l287
						}
						position++
						goto l286
					l287:
						position, tokenIndex, depth = position286, tokenIndex286, depth286
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l288
						}
						position++
						goto l286
					l288:
						position, tokenIndex, depth = position286, tokenIndex286, depth286
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l289
						}
						position++
						goto l286
					l289:
						position, tokenIndex, depth = position286, tokenIndex286, depth286
						if buffer[position] != rune('_') {
							goto l284
						}
						position++
					}
				l286:
				l290:
					{
						position291, tokenIndex291, depth291 := position, tokenIndex, depth
						{
							position292, tokenIndex292, depth292 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l293
							}
							position++
							goto l292
						l293:
							position, tokenIndex, depth = position292, tokenIndex292, depth292
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l294
							}
							position++
							goto l292
						l294:
							position, tokenIndex, depth = position292, tokenIndex292, depth292
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l295
							}
							position++
							goto l292
						l295:
							position, tokenIndex, depth = position292, tokenIndex292, depth292
							if buffer[position] != rune('_') {
								goto l296
							}
							position++
							goto l292
						l296:
							position, tokenIndex, depth = position292, tokenIndex292, depth292
							if buffer[position] != rune('-') {
								goto l291
							}
							position++
						}
					l292:
						goto l290
					l291:
						position, tokenIndex, depth = position291, tokenIndex291, depth291
					}
					goto l285
				l284:
					position, tokenIndex, depth = position284, tokenIndex284, depth284
				}
			l285:
				depth--
				add(ruleKey, position272)
			}
			return true
		l271:
			position, tokenIndex, depth = position271, tokenIndex271, depth271
			return false
		},
		/* 69 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position297, tokenIndex297, depth297 := position, tokenIndex, depth
			{
				position298 := position
				depth++
				if buffer[position] != rune('[') {
					goto l297
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l297
				}
				position++
			l299:
				{
					position300, tokenIndex300, depth300 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l300
					}
					position++
					goto l299
				l300:
					position, tokenIndex, depth = position300, tokenIndex300, depth300
				}
				if buffer[position] != rune(']') {
					goto l297
				}
				position++
				depth--
				add(ruleIndex, position298)
			}
			return true
		l297:
			position, tokenIndex, depth = position297, tokenIndex297, depth297
			return false
		},
		/* 70 IP <- <([0-9]+ '.' [0-9]+ '.' [0-9]+ '.' [0-9]+)> */
		func() bool {
			position301, tokenIndex301, depth301 := position, tokenIndex, depth
			{
				position302 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l301
				}
				position++
			l303:
				{
					position304, tokenIndex304, depth304 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l304
					}
					position++
					goto l303
				l304:
					position, tokenIndex, depth = position304, tokenIndex304, depth304
				}
				if buffer[position] != rune('.') {
					goto l301
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l301
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
					goto l301
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l301
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
					goto l301
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l301
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
				depth--
				add(ruleIP, position302)
			}
			return true
		l301:
			position, tokenIndex, depth = position301, tokenIndex301, depth301
			return false
		},
		/* 71 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position312 := position
				depth++
			l313:
				{
					position314, tokenIndex314, depth314 := position, tokenIndex, depth
					{
						position315, tokenIndex315, depth315 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l316
						}
						position++
						goto l315
					l316:
						position, tokenIndex, depth = position315, tokenIndex315, depth315
						if buffer[position] != rune('\t') {
							goto l317
						}
						position++
						goto l315
					l317:
						position, tokenIndex, depth = position315, tokenIndex315, depth315
						if buffer[position] != rune('\n') {
							goto l318
						}
						position++
						goto l315
					l318:
						position, tokenIndex, depth = position315, tokenIndex315, depth315
						if buffer[position] != rune('\r') {
							goto l314
						}
						position++
					}
				l315:
					goto l313
				l314:
					position, tokenIndex, depth = position314, tokenIndex314, depth314
				}
				depth--
				add(rulews, position312)
			}
			return true
		},
		/* 72 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position319, tokenIndex319, depth319 := position, tokenIndex, depth
			{
				position320 := position
				depth++
				{
					position323, tokenIndex323, depth323 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l324
					}
					position++
					goto l323
				l324:
					position, tokenIndex, depth = position323, tokenIndex323, depth323
					if buffer[position] != rune('\t') {
						goto l325
					}
					position++
					goto l323
				l325:
					position, tokenIndex, depth = position323, tokenIndex323, depth323
					if buffer[position] != rune('\n') {
						goto l326
					}
					position++
					goto l323
				l326:
					position, tokenIndex, depth = position323, tokenIndex323, depth323
					if buffer[position] != rune('\r') {
						goto l319
					}
					position++
				}
			l323:
			l321:
				{
					position322, tokenIndex322, depth322 := position, tokenIndex, depth
					{
						position327, tokenIndex327, depth327 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l328
						}
						position++
						goto l327
					l328:
						position, tokenIndex, depth = position327, tokenIndex327, depth327
						if buffer[position] != rune('\t') {
							goto l329
						}
						position++
						goto l327
					l329:
						position, tokenIndex, depth = position327, tokenIndex327, depth327
						if buffer[position] != rune('\n') {
							goto l330
						}
						position++
						goto l327
					l330:
						position, tokenIndex, depth = position327, tokenIndex327, depth327
						if buffer[position] != rune('\r') {
							goto l322
						}
						position++
					}
				l327:
					goto l321
				l322:
					position, tokenIndex, depth = position322, tokenIndex322, depth322
				}
				depth--
				add(rulereq_ws, position320)
			}
			return true
		l319:
			position, tokenIndex, depth = position319, tokenIndex319, depth319
			return false
		},
		/* 74 Action0 <- <{}> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
	}
	p.rules = _rules
}
