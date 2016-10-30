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
	ruleChainedCall
	ruleArguments
	ruleNextExpression
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
	rulews
	rulereq_ws

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
	"ChainedCall",
	"Arguments",
	"NextExpression",
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
	"ws",
	"req_ws",

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
	rules  [69]func() bool
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
		/* 2 MarkedExpression <- <(ws Marker (req_ws SubsequentMarker)* ws Grouped? ws)> */
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
					if !_rules[ruleGrouped]() {
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
		/* 4 Marker <- <('&' (('t' 'e' 'm' 'p' 'l' 'a' 't' 'e') / ('t' 'e' 'm' 'p' 'o' 'r' 'a' 'r' 'y')))> */
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
						goto l16
					}
					position++
					if buffer[position] != rune('e') {
						goto l16
					}
					position++
					if buffer[position] != rune('m') {
						goto l16
					}
					position++
					if buffer[position] != rune('p') {
						goto l16
					}
					position++
					if buffer[position] != rune('o') {
						goto l16
					}
					position++
					if buffer[position] != rune('r') {
						goto l16
					}
					position++
					if buffer[position] != rune('a') {
						goto l16
					}
					position++
					if buffer[position] != rune('r') {
						goto l16
					}
					position++
					if buffer[position] != rune('y') {
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
		/* 5 Expression <- <(ws (LambdaExpr / Level7) ws)> */
		func() bool {
			position20, tokenIndex20, depth20 := position, tokenIndex, depth
			{
				position21 := position
				depth++
				if !_rules[rulews]() {
					goto l20
				}
				{
					position22, tokenIndex22, depth22 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l23
					}
					goto l22
				l23:
					position, tokenIndex, depth = position22, tokenIndex22, depth22
					if !_rules[ruleLevel7]() {
						goto l20
					}
				}
			l22:
				if !_rules[rulews]() {
					goto l20
				}
				depth--
				add(ruleExpression, position21)
			}
			return true
		l20:
			position, tokenIndex, depth = position20, tokenIndex20, depth20
			return false
		},
		/* 6 Level7 <- <(Level6 (req_ws Or)*)> */
		func() bool {
			position24, tokenIndex24, depth24 := position, tokenIndex, depth
			{
				position25 := position
				depth++
				if !_rules[ruleLevel6]() {
					goto l24
				}
			l26:
				{
					position27, tokenIndex27, depth27 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l27
					}
					if !_rules[ruleOr]() {
						goto l27
					}
					goto l26
				l27:
					position, tokenIndex, depth = position27, tokenIndex27, depth27
				}
				depth--
				add(ruleLevel7, position25)
			}
			return true
		l24:
			position, tokenIndex, depth = position24, tokenIndex24, depth24
			return false
		},
		/* 7 Or <- <('|' '|' req_ws Level6)> */
		func() bool {
			position28, tokenIndex28, depth28 := position, tokenIndex, depth
			{
				position29 := position
				depth++
				if buffer[position] != rune('|') {
					goto l28
				}
				position++
				if buffer[position] != rune('|') {
					goto l28
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l28
				}
				if !_rules[ruleLevel6]() {
					goto l28
				}
				depth--
				add(ruleOr, position29)
			}
			return true
		l28:
			position, tokenIndex, depth = position28, tokenIndex28, depth28
			return false
		},
		/* 8 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position30, tokenIndex30, depth30 := position, tokenIndex, depth
			{
				position31 := position
				depth++
				{
					position32, tokenIndex32, depth32 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l33
					}
					goto l32
				l33:
					position, tokenIndex, depth = position32, tokenIndex32, depth32
					if !_rules[ruleLevel5]() {
						goto l30
					}
				}
			l32:
				depth--
				add(ruleLevel6, position31)
			}
			return true
		l30:
			position, tokenIndex, depth = position30, tokenIndex30, depth30
			return false
		},
		/* 9 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position34, tokenIndex34, depth34 := position, tokenIndex, depth
			{
				position35 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l34
				}
				if !_rules[rulews]() {
					goto l34
				}
				if buffer[position] != rune('?') {
					goto l34
				}
				position++
				if !_rules[ruleExpression]() {
					goto l34
				}
				if buffer[position] != rune(':') {
					goto l34
				}
				position++
				if !_rules[ruleExpression]() {
					goto l34
				}
				depth--
				add(ruleConditional, position35)
			}
			return true
		l34:
			position, tokenIndex, depth = position34, tokenIndex34, depth34
			return false
		},
		/* 10 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position36, tokenIndex36, depth36 := position, tokenIndex, depth
			{
				position37 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l36
				}
			l38:
				{
					position39, tokenIndex39, depth39 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l39
					}
					goto l38
				l39:
					position, tokenIndex, depth = position39, tokenIndex39, depth39
				}
				depth--
				add(ruleLevel5, position37)
			}
			return true
		l36:
			position, tokenIndex, depth = position36, tokenIndex36, depth36
			return false
		},
		/* 11 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l40
				}
				if !_rules[ruleLevel4]() {
					goto l40
				}
				depth--
				add(ruleConcatenation, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 12 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position42, tokenIndex42, depth42 := position, tokenIndex, depth
			{
				position43 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l42
				}
			l44:
				{
					position45, tokenIndex45, depth45 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l45
					}
					{
						position46, tokenIndex46, depth46 := position, tokenIndex, depth
						if !_rules[ruleLogOr]() {
							goto l47
						}
						goto l46
					l47:
						position, tokenIndex, depth = position46, tokenIndex46, depth46
						if !_rules[ruleLogAnd]() {
							goto l45
						}
					}
				l46:
					goto l44
				l45:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
				}
				depth--
				add(ruleLevel4, position43)
			}
			return true
		l42:
			position, tokenIndex, depth = position42, tokenIndex42, depth42
			return false
		},
		/* 13 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position48, tokenIndex48, depth48 := position, tokenIndex, depth
			{
				position49 := position
				depth++
				if buffer[position] != rune('-') {
					goto l48
				}
				position++
				if buffer[position] != rune('o') {
					goto l48
				}
				position++
				if buffer[position] != rune('r') {
					goto l48
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l48
				}
				if !_rules[ruleLevel3]() {
					goto l48
				}
				depth--
				add(ruleLogOr, position49)
			}
			return true
		l48:
			position, tokenIndex, depth = position48, tokenIndex48, depth48
			return false
		},
		/* 14 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position50, tokenIndex50, depth50 := position, tokenIndex, depth
			{
				position51 := position
				depth++
				if buffer[position] != rune('-') {
					goto l50
				}
				position++
				if buffer[position] != rune('a') {
					goto l50
				}
				position++
				if buffer[position] != rune('n') {
					goto l50
				}
				position++
				if buffer[position] != rune('d') {
					goto l50
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l50
				}
				if !_rules[ruleLevel3]() {
					goto l50
				}
				depth--
				add(ruleLogAnd, position51)
			}
			return true
		l50:
			position, tokenIndex, depth = position50, tokenIndex50, depth50
			return false
		},
		/* 15 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position52, tokenIndex52, depth52 := position, tokenIndex, depth
			{
				position53 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l52
				}
			l54:
				{
					position55, tokenIndex55, depth55 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l55
					}
					if !_rules[ruleComparison]() {
						goto l55
					}
					goto l54
				l55:
					position, tokenIndex, depth = position55, tokenIndex55, depth55
				}
				depth--
				add(ruleLevel3, position53)
			}
			return true
		l52:
			position, tokenIndex, depth = position52, tokenIndex52, depth52
			return false
		},
		/* 16 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position56, tokenIndex56, depth56 := position, tokenIndex, depth
			{
				position57 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l56
				}
				if !_rules[rulereq_ws]() {
					goto l56
				}
				if !_rules[ruleLevel2]() {
					goto l56
				}
				depth--
				add(ruleComparison, position57)
			}
			return true
		l56:
			position, tokenIndex, depth = position56, tokenIndex56, depth56
			return false
		},
		/* 17 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position58, tokenIndex58, depth58 := position, tokenIndex, depth
			{
				position59 := position
				depth++
				{
					position60, tokenIndex60, depth60 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l61
					}
					position++
					if buffer[position] != rune('=') {
						goto l61
					}
					position++
					goto l60
				l61:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
					if buffer[position] != rune('!') {
						goto l62
					}
					position++
					if buffer[position] != rune('=') {
						goto l62
					}
					position++
					goto l60
				l62:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
					if buffer[position] != rune('<') {
						goto l63
					}
					position++
					if buffer[position] != rune('=') {
						goto l63
					}
					position++
					goto l60
				l63:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
					if buffer[position] != rune('>') {
						goto l64
					}
					position++
					if buffer[position] != rune('=') {
						goto l64
					}
					position++
					goto l60
				l64:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
					if buffer[position] != rune('>') {
						goto l65
					}
					position++
					goto l60
				l65:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
					if buffer[position] != rune('<') {
						goto l66
					}
					position++
					goto l60
				l66:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
					if buffer[position] != rune('>') {
						goto l58
					}
					position++
				}
			l60:
				depth--
				add(ruleCompareOp, position59)
			}
			return true
		l58:
			position, tokenIndex, depth = position58, tokenIndex58, depth58
			return false
		},
		/* 18 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position67, tokenIndex67, depth67 := position, tokenIndex, depth
			{
				position68 := position
				depth++
				if !_rules[ruleLevel1]() {
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
						if !_rules[ruleAddition]() {
							goto l72
						}
						goto l71
					l72:
						position, tokenIndex, depth = position71, tokenIndex71, depth71
						if !_rules[ruleSubtraction]() {
							goto l70
						}
					}
				l71:
					goto l69
				l70:
					position, tokenIndex, depth = position70, tokenIndex70, depth70
				}
				depth--
				add(ruleLevel2, position68)
			}
			return true
		l67:
			position, tokenIndex, depth = position67, tokenIndex67, depth67
			return false
		},
		/* 19 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position73, tokenIndex73, depth73 := position, tokenIndex, depth
			{
				position74 := position
				depth++
				if buffer[position] != rune('+') {
					goto l73
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l73
				}
				if !_rules[ruleLevel1]() {
					goto l73
				}
				depth--
				add(ruleAddition, position74)
			}
			return true
		l73:
			position, tokenIndex, depth = position73, tokenIndex73, depth73
			return false
		},
		/* 20 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position75, tokenIndex75, depth75 := position, tokenIndex, depth
			{
				position76 := position
				depth++
				if buffer[position] != rune('-') {
					goto l75
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l75
				}
				if !_rules[ruleLevel1]() {
					goto l75
				}
				depth--
				add(ruleSubtraction, position76)
			}
			return true
		l75:
			position, tokenIndex, depth = position75, tokenIndex75, depth75
			return false
		},
		/* 21 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position77, tokenIndex77, depth77 := position, tokenIndex, depth
			{
				position78 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l77
				}
			l79:
				{
					position80, tokenIndex80, depth80 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l80
					}
					{
						position81, tokenIndex81, depth81 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l82
						}
						goto l81
					l82:
						position, tokenIndex, depth = position81, tokenIndex81, depth81
						if !_rules[ruleDivision]() {
							goto l83
						}
						goto l81
					l83:
						position, tokenIndex, depth = position81, tokenIndex81, depth81
						if !_rules[ruleModulo]() {
							goto l80
						}
					}
				l81:
					goto l79
				l80:
					position, tokenIndex, depth = position80, tokenIndex80, depth80
				}
				depth--
				add(ruleLevel1, position78)
			}
			return true
		l77:
			position, tokenIndex, depth = position77, tokenIndex77, depth77
			return false
		},
		/* 22 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position84, tokenIndex84, depth84 := position, tokenIndex, depth
			{
				position85 := position
				depth++
				if buffer[position] != rune('*') {
					goto l84
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l84
				}
				if !_rules[ruleLevel0]() {
					goto l84
				}
				depth--
				add(ruleMultiplication, position85)
			}
			return true
		l84:
			position, tokenIndex, depth = position84, tokenIndex84, depth84
			return false
		},
		/* 23 Division <- <('/' req_ws Level0)> */
		func() bool {
			position86, tokenIndex86, depth86 := position, tokenIndex, depth
			{
				position87 := position
				depth++
				if buffer[position] != rune('/') {
					goto l86
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l86
				}
				if !_rules[ruleLevel0]() {
					goto l86
				}
				depth--
				add(ruleDivision, position87)
			}
			return true
		l86:
			position, tokenIndex, depth = position86, tokenIndex86, depth86
			return false
		},
		/* 24 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position88, tokenIndex88, depth88 := position, tokenIndex, depth
			{
				position89 := position
				depth++
				if buffer[position] != rune('%') {
					goto l88
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l88
				}
				if !_rules[ruleLevel0]() {
					goto l88
				}
				depth--
				add(ruleModulo, position89)
			}
			return true
		l88:
			position, tokenIndex, depth = position88, tokenIndex88, depth88
			return false
		},
		/* 25 Level0 <- <(String / Integer / Boolean / Undefined / Nil / Not / Substitution / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position90, tokenIndex90, depth90 := position, tokenIndex, depth
			{
				position91 := position
				depth++
				{
					position92, tokenIndex92, depth92 := position, tokenIndex, depth
					if !_rules[ruleString]() {
						goto l93
					}
					goto l92
				l93:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleInteger]() {
						goto l94
					}
					goto l92
				l94:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleBoolean]() {
						goto l95
					}
					goto l92
				l95:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleUndefined]() {
						goto l96
					}
					goto l92
				l96:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleNil]() {
						goto l97
					}
					goto l92
				l97:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleNot]() {
						goto l98
					}
					goto l92
				l98:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleSubstitution]() {
						goto l99
					}
					goto l92
				l99:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleMerge]() {
						goto l100
					}
					goto l92
				l100:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleAuto]() {
						goto l101
					}
					goto l92
				l101:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleLambda]() {
						goto l102
					}
					goto l92
				l102:
					position, tokenIndex, depth = position92, tokenIndex92, depth92
					if !_rules[ruleChained]() {
						goto l90
					}
				}
			l92:
				depth--
				add(ruleLevel0, position91)
			}
			return true
		l90:
			position, tokenIndex, depth = position90, tokenIndex90, depth90
			return false
		},
		/* 26 Chained <- <((Mapping / Sum / List / Map / Range / Grouped / Reference) ChainedQualifiedExpression*)> */
		func() bool {
			position103, tokenIndex103, depth103 := position, tokenIndex, depth
			{
				position104 := position
				depth++
				{
					position105, tokenIndex105, depth105 := position, tokenIndex, depth
					if !_rules[ruleMapping]() {
						goto l106
					}
					goto l105
				l106:
					position, tokenIndex, depth = position105, tokenIndex105, depth105
					if !_rules[ruleSum]() {
						goto l107
					}
					goto l105
				l107:
					position, tokenIndex, depth = position105, tokenIndex105, depth105
					if !_rules[ruleList]() {
						goto l108
					}
					goto l105
				l108:
					position, tokenIndex, depth = position105, tokenIndex105, depth105
					if !_rules[ruleMap]() {
						goto l109
					}
					goto l105
				l109:
					position, tokenIndex, depth = position105, tokenIndex105, depth105
					if !_rules[ruleRange]() {
						goto l110
					}
					goto l105
				l110:
					position, tokenIndex, depth = position105, tokenIndex105, depth105
					if !_rules[ruleGrouped]() {
						goto l111
					}
					goto l105
				l111:
					position, tokenIndex, depth = position105, tokenIndex105, depth105
					if !_rules[ruleReference]() {
						goto l103
					}
				}
			l105:
			l112:
				{
					position113, tokenIndex113, depth113 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l113
					}
					goto l112
				l113:
					position, tokenIndex, depth = position113, tokenIndex113, depth113
				}
				depth--
				add(ruleChained, position104)
			}
			return true
		l103:
			position, tokenIndex, depth = position103, tokenIndex103, depth103
			return false
		},
		/* 27 ChainedQualifiedExpression <- <(ChainedCall / ('.' (ChainedRef / ChainedDynRef)))> */
		func() bool {
			position114, tokenIndex114, depth114 := position, tokenIndex, depth
			{
				position115 := position
				depth++
				{
					position116, tokenIndex116, depth116 := position, tokenIndex, depth
					if !_rules[ruleChainedCall]() {
						goto l117
					}
					goto l116
				l117:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
					if buffer[position] != rune('.') {
						goto l114
					}
					position++
					{
						position118, tokenIndex118, depth118 := position, tokenIndex, depth
						if !_rules[ruleChainedRef]() {
							goto l119
						}
						goto l118
					l119:
						position, tokenIndex, depth = position118, tokenIndex118, depth118
						if !_rules[ruleChainedDynRef]() {
							goto l114
						}
					}
				l118:
				}
			l116:
				depth--
				add(ruleChainedQualifiedExpression, position115)
			}
			return true
		l114:
			position, tokenIndex, depth = position114, tokenIndex114, depth114
			return false
		},
		/* 28 ChainedRef <- <((Key / Index) FollowUpRef)> */
		func() bool {
			position120, tokenIndex120, depth120 := position, tokenIndex, depth
			{
				position121 := position
				depth++
				{
					position122, tokenIndex122, depth122 := position, tokenIndex, depth
					if !_rules[ruleKey]() {
						goto l123
					}
					goto l122
				l123:
					position, tokenIndex, depth = position122, tokenIndex122, depth122
					if !_rules[ruleIndex]() {
						goto l120
					}
				}
			l122:
				if !_rules[ruleFollowUpRef]() {
					goto l120
				}
				depth--
				add(ruleChainedRef, position121)
			}
			return true
		l120:
			position, tokenIndex, depth = position120, tokenIndex120, depth120
			return false
		},
		/* 29 ChainedDynRef <- <('[' Expression ']')> */
		func() bool {
			position124, tokenIndex124, depth124 := position, tokenIndex, depth
			{
				position125 := position
				depth++
				if buffer[position] != rune('[') {
					goto l124
				}
				position++
				if !_rules[ruleExpression]() {
					goto l124
				}
				if buffer[position] != rune(']') {
					goto l124
				}
				position++
				depth--
				add(ruleChainedDynRef, position125)
			}
			return true
		l124:
			position, tokenIndex, depth = position124, tokenIndex124, depth124
			return false
		},
		/* 30 ChainedCall <- <('(' Arguments ')')> */
		func() bool {
			position126, tokenIndex126, depth126 := position, tokenIndex, depth
			{
				position127 := position
				depth++
				if buffer[position] != rune('(') {
					goto l126
				}
				position++
				if !_rules[ruleArguments]() {
					goto l126
				}
				if buffer[position] != rune(')') {
					goto l126
				}
				position++
				depth--
				add(ruleChainedCall, position127)
			}
			return true
		l126:
			position, tokenIndex, depth = position126, tokenIndex126, depth126
			return false
		},
		/* 31 Arguments <- <(Expression NextExpression*)> */
		func() bool {
			position128, tokenIndex128, depth128 := position, tokenIndex, depth
			{
				position129 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l128
				}
			l130:
				{
					position131, tokenIndex131, depth131 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l131
					}
					goto l130
				l131:
					position, tokenIndex, depth = position131, tokenIndex131, depth131
				}
				depth--
				add(ruleArguments, position129)
			}
			return true
		l128:
			position, tokenIndex, depth = position128, tokenIndex128, depth128
			return false
		},
		/* 32 NextExpression <- <(',' Expression)> */
		func() bool {
			position132, tokenIndex132, depth132 := position, tokenIndex, depth
			{
				position133 := position
				depth++
				if buffer[position] != rune(',') {
					goto l132
				}
				position++
				if !_rules[ruleExpression]() {
					goto l132
				}
				depth--
				add(ruleNextExpression, position133)
			}
			return true
		l132:
			position, tokenIndex, depth = position132, tokenIndex132, depth132
			return false
		},
		/* 33 Substitution <- <('*' Level0)> */
		func() bool {
			position134, tokenIndex134, depth134 := position, tokenIndex, depth
			{
				position135 := position
				depth++
				if buffer[position] != rune('*') {
					goto l134
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l134
				}
				depth--
				add(ruleSubstitution, position135)
			}
			return true
		l134:
			position, tokenIndex, depth = position134, tokenIndex134, depth134
			return false
		},
		/* 34 Not <- <('!' ws Level0)> */
		func() bool {
			position136, tokenIndex136, depth136 := position, tokenIndex, depth
			{
				position137 := position
				depth++
				if buffer[position] != rune('!') {
					goto l136
				}
				position++
				if !_rules[rulews]() {
					goto l136
				}
				if !_rules[ruleLevel0]() {
					goto l136
				}
				depth--
				add(ruleNot, position137)
			}
			return true
		l136:
			position, tokenIndex, depth = position136, tokenIndex136, depth136
			return false
		},
		/* 35 Grouped <- <('(' Expression ')')> */
		func() bool {
			position138, tokenIndex138, depth138 := position, tokenIndex, depth
			{
				position139 := position
				depth++
				if buffer[position] != rune('(') {
					goto l138
				}
				position++
				if !_rules[ruleExpression]() {
					goto l138
				}
				if buffer[position] != rune(')') {
					goto l138
				}
				position++
				depth--
				add(ruleGrouped, position139)
			}
			return true
		l138:
			position, tokenIndex, depth = position138, tokenIndex138, depth138
			return false
		},
		/* 36 Range <- <('[' Expression ('.' '.') Expression ']')> */
		func() bool {
			position140, tokenIndex140, depth140 := position, tokenIndex, depth
			{
				position141 := position
				depth++
				if buffer[position] != rune('[') {
					goto l140
				}
				position++
				if !_rules[ruleExpression]() {
					goto l140
				}
				if buffer[position] != rune('.') {
					goto l140
				}
				position++
				if buffer[position] != rune('.') {
					goto l140
				}
				position++
				if !_rules[ruleExpression]() {
					goto l140
				}
				if buffer[position] != rune(']') {
					goto l140
				}
				position++
				depth--
				add(ruleRange, position141)
			}
			return true
		l140:
			position, tokenIndex, depth = position140, tokenIndex140, depth140
			return false
		},
		/* 37 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position142, tokenIndex142, depth142 := position, tokenIndex, depth
			{
				position143 := position
				depth++
				{
					position144, tokenIndex144, depth144 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l144
					}
					position++
					goto l145
				l144:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
				}
			l145:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l142
				}
				position++
			l146:
				{
					position147, tokenIndex147, depth147 := position, tokenIndex, depth
					{
						position148, tokenIndex148, depth148 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l149
						}
						position++
						goto l148
					l149:
						position, tokenIndex, depth = position148, tokenIndex148, depth148
						if buffer[position] != rune('_') {
							goto l147
						}
						position++
					}
				l148:
					goto l146
				l147:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
				}
				depth--
				add(ruleInteger, position143)
			}
			return true
		l142:
			position, tokenIndex, depth = position142, tokenIndex142, depth142
			return false
		},
		/* 38 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position150, tokenIndex150, depth150 := position, tokenIndex, depth
			{
				position151 := position
				depth++
				if buffer[position] != rune('"') {
					goto l150
				}
				position++
			l152:
				{
					position153, tokenIndex153, depth153 := position, tokenIndex, depth
					{
						position154, tokenIndex154, depth154 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l155
						}
						position++
						if buffer[position] != rune('"') {
							goto l155
						}
						position++
						goto l154
					l155:
						position, tokenIndex, depth = position154, tokenIndex154, depth154
						{
							position156, tokenIndex156, depth156 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l156
							}
							position++
							goto l153
						l156:
							position, tokenIndex, depth = position156, tokenIndex156, depth156
						}
						if !matchDot() {
							goto l153
						}
					}
				l154:
					goto l152
				l153:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
				}
				if buffer[position] != rune('"') {
					goto l150
				}
				position++
				depth--
				add(ruleString, position151)
			}
			return true
		l150:
			position, tokenIndex, depth = position150, tokenIndex150, depth150
			return false
		},
		/* 39 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position157, tokenIndex157, depth157 := position, tokenIndex, depth
			{
				position158 := position
				depth++
				{
					position159, tokenIndex159, depth159 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l160
					}
					position++
					if buffer[position] != rune('r') {
						goto l160
					}
					position++
					if buffer[position] != rune('u') {
						goto l160
					}
					position++
					if buffer[position] != rune('e') {
						goto l160
					}
					position++
					goto l159
				l160:
					position, tokenIndex, depth = position159, tokenIndex159, depth159
					if buffer[position] != rune('f') {
						goto l157
					}
					position++
					if buffer[position] != rune('a') {
						goto l157
					}
					position++
					if buffer[position] != rune('l') {
						goto l157
					}
					position++
					if buffer[position] != rune('s') {
						goto l157
					}
					position++
					if buffer[position] != rune('e') {
						goto l157
					}
					position++
				}
			l159:
				depth--
				add(ruleBoolean, position158)
			}
			return true
		l157:
			position, tokenIndex, depth = position157, tokenIndex157, depth157
			return false
		},
		/* 40 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position161, tokenIndex161, depth161 := position, tokenIndex, depth
			{
				position162 := position
				depth++
				{
					position163, tokenIndex163, depth163 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l164
					}
					position++
					if buffer[position] != rune('i') {
						goto l164
					}
					position++
					if buffer[position] != rune('l') {
						goto l164
					}
					position++
					goto l163
				l164:
					position, tokenIndex, depth = position163, tokenIndex163, depth163
					if buffer[position] != rune('~') {
						goto l161
					}
					position++
				}
			l163:
				depth--
				add(ruleNil, position162)
			}
			return true
		l161:
			position, tokenIndex, depth = position161, tokenIndex161, depth161
			return false
		},
		/* 41 Undefined <- <('~' '~')> */
		func() bool {
			position165, tokenIndex165, depth165 := position, tokenIndex, depth
			{
				position166 := position
				depth++
				if buffer[position] != rune('~') {
					goto l165
				}
				position++
				if buffer[position] != rune('~') {
					goto l165
				}
				position++
				depth--
				add(ruleUndefined, position166)
			}
			return true
		l165:
			position, tokenIndex, depth = position165, tokenIndex165, depth165
			return false
		},
		/* 42 List <- <('[' Contents? ']')> */
		func() bool {
			position167, tokenIndex167, depth167 := position, tokenIndex, depth
			{
				position168 := position
				depth++
				if buffer[position] != rune('[') {
					goto l167
				}
				position++
				{
					position169, tokenIndex169, depth169 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l169
					}
					goto l170
				l169:
					position, tokenIndex, depth = position169, tokenIndex169, depth169
				}
			l170:
				if buffer[position] != rune(']') {
					goto l167
				}
				position++
				depth--
				add(ruleList, position168)
			}
			return true
		l167:
			position, tokenIndex, depth = position167, tokenIndex167, depth167
			return false
		},
		/* 43 Contents <- <(Expression NextExpression*)> */
		func() bool {
			position171, tokenIndex171, depth171 := position, tokenIndex, depth
			{
				position172 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l171
				}
			l173:
				{
					position174, tokenIndex174, depth174 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l174
					}
					goto l173
				l174:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
				}
				depth--
				add(ruleContents, position172)
			}
			return true
		l171:
			position, tokenIndex, depth = position171, tokenIndex171, depth171
			return false
		},
		/* 44 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position175, tokenIndex175, depth175 := position, tokenIndex, depth
			{
				position176 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l175
				}
				if !_rules[rulews]() {
					goto l175
				}
				{
					position177, tokenIndex177, depth177 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l177
					}
					goto l178
				l177:
					position, tokenIndex, depth = position177, tokenIndex177, depth177
				}
			l178:
				if buffer[position] != rune('}') {
					goto l175
				}
				position++
				depth--
				add(ruleMap, position176)
			}
			return true
		l175:
			position, tokenIndex, depth = position175, tokenIndex175, depth175
			return false
		},
		/* 45 CreateMap <- <'{'> */
		func() bool {
			position179, tokenIndex179, depth179 := position, tokenIndex, depth
			{
				position180 := position
				depth++
				if buffer[position] != rune('{') {
					goto l179
				}
				position++
				depth--
				add(ruleCreateMap, position180)
			}
			return true
		l179:
			position, tokenIndex, depth = position179, tokenIndex179, depth179
			return false
		},
		/* 46 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position181, tokenIndex181, depth181 := position, tokenIndex, depth
			{
				position182 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l181
				}
			l183:
				{
					position184, tokenIndex184, depth184 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l184
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l184
					}
					goto l183
				l184:
					position, tokenIndex, depth = position184, tokenIndex184, depth184
				}
				depth--
				add(ruleAssignments, position182)
			}
			return true
		l181:
			position, tokenIndex, depth = position181, tokenIndex181, depth181
			return false
		},
		/* 47 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position185, tokenIndex185, depth185 := position, tokenIndex, depth
			{
				position186 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l185
				}
				if buffer[position] != rune('=') {
					goto l185
				}
				position++
				if !_rules[ruleExpression]() {
					goto l185
				}
				depth--
				add(ruleAssignment, position186)
			}
			return true
		l185:
			position, tokenIndex, depth = position185, tokenIndex185, depth185
			return false
		},
		/* 48 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position187, tokenIndex187, depth187 := position, tokenIndex, depth
			{
				position188 := position
				depth++
				{
					position189, tokenIndex189, depth189 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l190
					}
					goto l189
				l190:
					position, tokenIndex, depth = position189, tokenIndex189, depth189
					if !_rules[ruleSimpleMerge]() {
						goto l187
					}
				}
			l189:
				depth--
				add(ruleMerge, position188)
			}
			return true
		l187:
			position, tokenIndex, depth = position187, tokenIndex187, depth187
			return false
		},
		/* 49 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position191, tokenIndex191, depth191 := position, tokenIndex, depth
			{
				position192 := position
				depth++
				if buffer[position] != rune('m') {
					goto l191
				}
				position++
				if buffer[position] != rune('e') {
					goto l191
				}
				position++
				if buffer[position] != rune('r') {
					goto l191
				}
				position++
				if buffer[position] != rune('g') {
					goto l191
				}
				position++
				if buffer[position] != rune('e') {
					goto l191
				}
				position++
				{
					position193, tokenIndex193, depth193 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l193
					}
					if !_rules[ruleRequired]() {
						goto l193
					}
					goto l191
				l193:
					position, tokenIndex, depth = position193, tokenIndex193, depth193
				}
				{
					position194, tokenIndex194, depth194 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l194
					}
					{
						position196, tokenIndex196, depth196 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l197
						}
						goto l196
					l197:
						position, tokenIndex, depth = position196, tokenIndex196, depth196
						if !_rules[ruleOn]() {
							goto l194
						}
					}
				l196:
					goto l195
				l194:
					position, tokenIndex, depth = position194, tokenIndex194, depth194
				}
			l195:
				if !_rules[rulereq_ws]() {
					goto l191
				}
				if !_rules[ruleReference]() {
					goto l191
				}
				depth--
				add(ruleRefMerge, position192)
			}
			return true
		l191:
			position, tokenIndex, depth = position191, tokenIndex191, depth191
			return false
		},
		/* 50 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position198, tokenIndex198, depth198 := position, tokenIndex, depth
			{
				position199 := position
				depth++
				if buffer[position] != rune('m') {
					goto l198
				}
				position++
				if buffer[position] != rune('e') {
					goto l198
				}
				position++
				if buffer[position] != rune('r') {
					goto l198
				}
				position++
				if buffer[position] != rune('g') {
					goto l198
				}
				position++
				if buffer[position] != rune('e') {
					goto l198
				}
				position++
				{
					position200, tokenIndex200, depth200 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l200
					}
					{
						position202, tokenIndex202, depth202 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l203
						}
						goto l202
					l203:
						position, tokenIndex, depth = position202, tokenIndex202, depth202
						if !_rules[ruleRequired]() {
							goto l204
						}
						goto l202
					l204:
						position, tokenIndex, depth = position202, tokenIndex202, depth202
						if !_rules[ruleOn]() {
							goto l200
						}
					}
				l202:
					goto l201
				l200:
					position, tokenIndex, depth = position200, tokenIndex200, depth200
				}
			l201:
				depth--
				add(ruleSimpleMerge, position199)
			}
			return true
		l198:
			position, tokenIndex, depth = position198, tokenIndex198, depth198
			return false
		},
		/* 51 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position205, tokenIndex205, depth205 := position, tokenIndex, depth
			{
				position206 := position
				depth++
				if buffer[position] != rune('r') {
					goto l205
				}
				position++
				if buffer[position] != rune('e') {
					goto l205
				}
				position++
				if buffer[position] != rune('p') {
					goto l205
				}
				position++
				if buffer[position] != rune('l') {
					goto l205
				}
				position++
				if buffer[position] != rune('a') {
					goto l205
				}
				position++
				if buffer[position] != rune('c') {
					goto l205
				}
				position++
				if buffer[position] != rune('e') {
					goto l205
				}
				position++
				depth--
				add(ruleReplace, position206)
			}
			return true
		l205:
			position, tokenIndex, depth = position205, tokenIndex205, depth205
			return false
		},
		/* 52 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position207, tokenIndex207, depth207 := position, tokenIndex, depth
			{
				position208 := position
				depth++
				if buffer[position] != rune('r') {
					goto l207
				}
				position++
				if buffer[position] != rune('e') {
					goto l207
				}
				position++
				if buffer[position] != rune('q') {
					goto l207
				}
				position++
				if buffer[position] != rune('u') {
					goto l207
				}
				position++
				if buffer[position] != rune('i') {
					goto l207
				}
				position++
				if buffer[position] != rune('r') {
					goto l207
				}
				position++
				if buffer[position] != rune('e') {
					goto l207
				}
				position++
				if buffer[position] != rune('d') {
					goto l207
				}
				position++
				depth--
				add(ruleRequired, position208)
			}
			return true
		l207:
			position, tokenIndex, depth = position207, tokenIndex207, depth207
			return false
		},
		/* 53 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position209, tokenIndex209, depth209 := position, tokenIndex, depth
			{
				position210 := position
				depth++
				if buffer[position] != rune('o') {
					goto l209
				}
				position++
				if buffer[position] != rune('n') {
					goto l209
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l209
				}
				if !_rules[ruleName]() {
					goto l209
				}
				depth--
				add(ruleOn, position210)
			}
			return true
		l209:
			position, tokenIndex, depth = position209, tokenIndex209, depth209
			return false
		},
		/* 54 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position211, tokenIndex211, depth211 := position, tokenIndex, depth
			{
				position212 := position
				depth++
				if buffer[position] != rune('a') {
					goto l211
				}
				position++
				if buffer[position] != rune('u') {
					goto l211
				}
				position++
				if buffer[position] != rune('t') {
					goto l211
				}
				position++
				if buffer[position] != rune('o') {
					goto l211
				}
				position++
				depth--
				add(ruleAuto, position212)
			}
			return true
		l211:
			position, tokenIndex, depth = position211, tokenIndex211, depth211
			return false
		},
		/* 55 Mapping <- <('m' 'a' 'p' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position213, tokenIndex213, depth213 := position, tokenIndex, depth
			{
				position214 := position
				depth++
				if buffer[position] != rune('m') {
					goto l213
				}
				position++
				if buffer[position] != rune('a') {
					goto l213
				}
				position++
				if buffer[position] != rune('p') {
					goto l213
				}
				position++
				if buffer[position] != rune('[') {
					goto l213
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l213
				}
				{
					position215, tokenIndex215, depth215 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l216
					}
					goto l215
				l216:
					position, tokenIndex, depth = position215, tokenIndex215, depth215
					if buffer[position] != rune('|') {
						goto l213
					}
					position++
					if !_rules[ruleExpression]() {
						goto l213
					}
				}
			l215:
				if buffer[position] != rune(']') {
					goto l213
				}
				position++
				depth--
				add(ruleMapping, position214)
			}
			return true
		l213:
			position, tokenIndex, depth = position213, tokenIndex213, depth213
			return false
		},
		/* 56 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position217, tokenIndex217, depth217 := position, tokenIndex, depth
			{
				position218 := position
				depth++
				if buffer[position] != rune('s') {
					goto l217
				}
				position++
				if buffer[position] != rune('u') {
					goto l217
				}
				position++
				if buffer[position] != rune('m') {
					goto l217
				}
				position++
				if buffer[position] != rune('[') {
					goto l217
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l217
				}
				if buffer[position] != rune('|') {
					goto l217
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l217
				}
				{
					position219, tokenIndex219, depth219 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l220
					}
					goto l219
				l220:
					position, tokenIndex, depth = position219, tokenIndex219, depth219
					if buffer[position] != rune('|') {
						goto l217
					}
					position++
					if !_rules[ruleExpression]() {
						goto l217
					}
				}
			l219:
				if buffer[position] != rune(']') {
					goto l217
				}
				position++
				depth--
				add(ruleSum, position218)
			}
			return true
		l217:
			position, tokenIndex, depth = position217, tokenIndex217, depth217
			return false
		},
		/* 57 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position221, tokenIndex221, depth221 := position, tokenIndex, depth
			{
				position222 := position
				depth++
				if buffer[position] != rune('l') {
					goto l221
				}
				position++
				if buffer[position] != rune('a') {
					goto l221
				}
				position++
				if buffer[position] != rune('m') {
					goto l221
				}
				position++
				if buffer[position] != rune('b') {
					goto l221
				}
				position++
				if buffer[position] != rune('d') {
					goto l221
				}
				position++
				if buffer[position] != rune('a') {
					goto l221
				}
				position++
				{
					position223, tokenIndex223, depth223 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l224
					}
					goto l223
				l224:
					position, tokenIndex, depth = position223, tokenIndex223, depth223
					if !_rules[ruleLambdaExpr]() {
						goto l221
					}
				}
			l223:
				depth--
				add(ruleLambda, position222)
			}
			return true
		l221:
			position, tokenIndex, depth = position221, tokenIndex221, depth221
			return false
		},
		/* 58 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position225, tokenIndex225, depth225 := position, tokenIndex, depth
			{
				position226 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l225
				}
				if !_rules[ruleExpression]() {
					goto l225
				}
				depth--
				add(ruleLambdaRef, position226)
			}
			return true
		l225:
			position, tokenIndex, depth = position225, tokenIndex225, depth225
			return false
		},
		/* 59 LambdaExpr <- <(ws '|' ws Name NextName* ws '|' ws ('-' '>') Expression)> */
		func() bool {
			position227, tokenIndex227, depth227 := position, tokenIndex, depth
			{
				position228 := position
				depth++
				if !_rules[rulews]() {
					goto l227
				}
				if buffer[position] != rune('|') {
					goto l227
				}
				position++
				if !_rules[rulews]() {
					goto l227
				}
				if !_rules[ruleName]() {
					goto l227
				}
			l229:
				{
					position230, tokenIndex230, depth230 := position, tokenIndex, depth
					if !_rules[ruleNextName]() {
						goto l230
					}
					goto l229
				l230:
					position, tokenIndex, depth = position230, tokenIndex230, depth230
				}
				if !_rules[rulews]() {
					goto l227
				}
				if buffer[position] != rune('|') {
					goto l227
				}
				position++
				if !_rules[rulews]() {
					goto l227
				}
				if buffer[position] != rune('-') {
					goto l227
				}
				position++
				if buffer[position] != rune('>') {
					goto l227
				}
				position++
				if !_rules[ruleExpression]() {
					goto l227
				}
				depth--
				add(ruleLambdaExpr, position228)
			}
			return true
		l227:
			position, tokenIndex, depth = position227, tokenIndex227, depth227
			return false
		},
		/* 60 NextName <- <(ws ',' ws Name)> */
		func() bool {
			position231, tokenIndex231, depth231 := position, tokenIndex, depth
			{
				position232 := position
				depth++
				if !_rules[rulews]() {
					goto l231
				}
				if buffer[position] != rune(',') {
					goto l231
				}
				position++
				if !_rules[rulews]() {
					goto l231
				}
				if !_rules[ruleName]() {
					goto l231
				}
				depth--
				add(ruleNextName, position232)
			}
			return true
		l231:
			position, tokenIndex, depth = position231, tokenIndex231, depth231
			return false
		},
		/* 61 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position233, tokenIndex233, depth233 := position, tokenIndex, depth
			{
				position234 := position
				depth++
				{
					position237, tokenIndex237, depth237 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l238
					}
					position++
					goto l237
				l238:
					position, tokenIndex, depth = position237, tokenIndex237, depth237
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l239
					}
					position++
					goto l237
				l239:
					position, tokenIndex, depth = position237, tokenIndex237, depth237
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l240
					}
					position++
					goto l237
				l240:
					position, tokenIndex, depth = position237, tokenIndex237, depth237
					if buffer[position] != rune('_') {
						goto l233
					}
					position++
				}
			l237:
			l235:
				{
					position236, tokenIndex236, depth236 := position, tokenIndex, depth
					{
						position241, tokenIndex241, depth241 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l242
						}
						position++
						goto l241
					l242:
						position, tokenIndex, depth = position241, tokenIndex241, depth241
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l243
						}
						position++
						goto l241
					l243:
						position, tokenIndex, depth = position241, tokenIndex241, depth241
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l244
						}
						position++
						goto l241
					l244:
						position, tokenIndex, depth = position241, tokenIndex241, depth241
						if buffer[position] != rune('_') {
							goto l236
						}
						position++
					}
				l241:
					goto l235
				l236:
					position, tokenIndex, depth = position236, tokenIndex236, depth236
				}
				depth--
				add(ruleName, position234)
			}
			return true
		l233:
			position, tokenIndex, depth = position233, tokenIndex233, depth233
			return false
		},
		/* 62 Reference <- <('.'? Key FollowUpRef)> */
		func() bool {
			position245, tokenIndex245, depth245 := position, tokenIndex, depth
			{
				position246 := position
				depth++
				{
					position247, tokenIndex247, depth247 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l247
					}
					position++
					goto l248
				l247:
					position, tokenIndex, depth = position247, tokenIndex247, depth247
				}
			l248:
				if !_rules[ruleKey]() {
					goto l245
				}
				if !_rules[ruleFollowUpRef]() {
					goto l245
				}
				depth--
				add(ruleReference, position246)
			}
			return true
		l245:
			position, tokenIndex, depth = position245, tokenIndex245, depth245
			return false
		},
		/* 63 FollowUpRef <- <('.' (Key / Index))*> */
		func() bool {
			{
				position250 := position
				depth++
			l251:
				{
					position252, tokenIndex252, depth252 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l252
					}
					position++
					{
						position253, tokenIndex253, depth253 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l254
						}
						goto l253
					l254:
						position, tokenIndex, depth = position253, tokenIndex253, depth253
						if !_rules[ruleIndex]() {
							goto l252
						}
					}
				l253:
					goto l251
				l252:
					position, tokenIndex, depth = position252, tokenIndex252, depth252
				}
				depth--
				add(ruleFollowUpRef, position250)
			}
			return true
		},
		/* 64 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position255, tokenIndex255, depth255 := position, tokenIndex, depth
			{
				position256 := position
				depth++
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
						goto l255
					}
					position++
				}
			l257:
			l261:
				{
					position262, tokenIndex262, depth262 := position, tokenIndex, depth
					{
						position263, tokenIndex263, depth263 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l264
						}
						position++
						goto l263
					l264:
						position, tokenIndex, depth = position263, tokenIndex263, depth263
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l265
						}
						position++
						goto l263
					l265:
						position, tokenIndex, depth = position263, tokenIndex263, depth263
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l266
						}
						position++
						goto l263
					l266:
						position, tokenIndex, depth = position263, tokenIndex263, depth263
						if buffer[position] != rune('_') {
							goto l267
						}
						position++
						goto l263
					l267:
						position, tokenIndex, depth = position263, tokenIndex263, depth263
						if buffer[position] != rune('-') {
							goto l262
						}
						position++
					}
				l263:
					goto l261
				l262:
					position, tokenIndex, depth = position262, tokenIndex262, depth262
				}
				{
					position268, tokenIndex268, depth268 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l268
					}
					position++
					{
						position270, tokenIndex270, depth270 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l271
						}
						position++
						goto l270
					l271:
						position, tokenIndex, depth = position270, tokenIndex270, depth270
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l272
						}
						position++
						goto l270
					l272:
						position, tokenIndex, depth = position270, tokenIndex270, depth270
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l273
						}
						position++
						goto l270
					l273:
						position, tokenIndex, depth = position270, tokenIndex270, depth270
						if buffer[position] != rune('_') {
							goto l268
						}
						position++
					}
				l270:
				l274:
					{
						position275, tokenIndex275, depth275 := position, tokenIndex, depth
						{
							position276, tokenIndex276, depth276 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l277
							}
							position++
							goto l276
						l277:
							position, tokenIndex, depth = position276, tokenIndex276, depth276
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l278
							}
							position++
							goto l276
						l278:
							position, tokenIndex, depth = position276, tokenIndex276, depth276
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l279
							}
							position++
							goto l276
						l279:
							position, tokenIndex, depth = position276, tokenIndex276, depth276
							if buffer[position] != rune('_') {
								goto l280
							}
							position++
							goto l276
						l280:
							position, tokenIndex, depth = position276, tokenIndex276, depth276
							if buffer[position] != rune('-') {
								goto l275
							}
							position++
						}
					l276:
						goto l274
					l275:
						position, tokenIndex, depth = position275, tokenIndex275, depth275
					}
					goto l269
				l268:
					position, tokenIndex, depth = position268, tokenIndex268, depth268
				}
			l269:
				depth--
				add(ruleKey, position256)
			}
			return true
		l255:
			position, tokenIndex, depth = position255, tokenIndex255, depth255
			return false
		},
		/* 65 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position281, tokenIndex281, depth281 := position, tokenIndex, depth
			{
				position282 := position
				depth++
				if buffer[position] != rune('[') {
					goto l281
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l281
				}
				position++
			l283:
				{
					position284, tokenIndex284, depth284 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l284
					}
					position++
					goto l283
				l284:
					position, tokenIndex, depth = position284, tokenIndex284, depth284
				}
				if buffer[position] != rune(']') {
					goto l281
				}
				position++
				depth--
				add(ruleIndex, position282)
			}
			return true
		l281:
			position, tokenIndex, depth = position281, tokenIndex281, depth281
			return false
		},
		/* 66 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position286 := position
				depth++
			l287:
				{
					position288, tokenIndex288, depth288 := position, tokenIndex, depth
					{
						position289, tokenIndex289, depth289 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l290
						}
						position++
						goto l289
					l290:
						position, tokenIndex, depth = position289, tokenIndex289, depth289
						if buffer[position] != rune('\t') {
							goto l291
						}
						position++
						goto l289
					l291:
						position, tokenIndex, depth = position289, tokenIndex289, depth289
						if buffer[position] != rune('\n') {
							goto l292
						}
						position++
						goto l289
					l292:
						position, tokenIndex, depth = position289, tokenIndex289, depth289
						if buffer[position] != rune('\r') {
							goto l288
						}
						position++
					}
				l289:
					goto l287
				l288:
					position, tokenIndex, depth = position288, tokenIndex288, depth288
				}
				depth--
				add(rulews, position286)
			}
			return true
		},
		/* 67 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position293, tokenIndex293, depth293 := position, tokenIndex, depth
			{
				position294 := position
				depth++
				{
					position297, tokenIndex297, depth297 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l298
					}
					position++
					goto l297
				l298:
					position, tokenIndex, depth = position297, tokenIndex297, depth297
					if buffer[position] != rune('\t') {
						goto l299
					}
					position++
					goto l297
				l299:
					position, tokenIndex, depth = position297, tokenIndex297, depth297
					if buffer[position] != rune('\n') {
						goto l300
					}
					position++
					goto l297
				l300:
					position, tokenIndex, depth = position297, tokenIndex297, depth297
					if buffer[position] != rune('\r') {
						goto l293
					}
					position++
				}
			l297:
			l295:
				{
					position296, tokenIndex296, depth296 := position, tokenIndex, depth
					{
						position301, tokenIndex301, depth301 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l302
						}
						position++
						goto l301
					l302:
						position, tokenIndex, depth = position301, tokenIndex301, depth301
						if buffer[position] != rune('\t') {
							goto l303
						}
						position++
						goto l301
					l303:
						position, tokenIndex, depth = position301, tokenIndex301, depth301
						if buffer[position] != rune('\n') {
							goto l304
						}
						position++
						goto l301
					l304:
						position, tokenIndex, depth = position301, tokenIndex301, depth301
						if buffer[position] != rune('\r') {
							goto l296
						}
						position++
					}
				l301:
					goto l295
				l296:
					position, tokenIndex, depth = position296, tokenIndex296, depth296
				}
				depth--
				add(rulereq_ws, position294)
			}
			return true
		l293:
			position, tokenIndex, depth = position293, tokenIndex293, depth293
			return false
		},
	}
	p.rules = _rules
}
