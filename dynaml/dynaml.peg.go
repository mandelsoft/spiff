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
	ruleSlice
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
	"Slice",
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
	rules  [70]func() bool
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
		/* 5 Expression <- <(ws (LambdaExpr / Level7) ws)> */
		func() bool {
			position21, tokenIndex21, depth21 := position, tokenIndex, depth
			{
				position22 := position
				depth++
				if !_rules[rulews]() {
					goto l21
				}
				{
					position23, tokenIndex23, depth23 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l24
					}
					goto l23
				l24:
					position, tokenIndex, depth = position23, tokenIndex23, depth23
					if !_rules[ruleLevel7]() {
						goto l21
					}
				}
			l23:
				if !_rules[rulews]() {
					goto l21
				}
				depth--
				add(ruleExpression, position22)
			}
			return true
		l21:
			position, tokenIndex, depth = position21, tokenIndex21, depth21
			return false
		},
		/* 6 Level7 <- <(Level6 (req_ws Or)*)> */
		func() bool {
			position25, tokenIndex25, depth25 := position, tokenIndex, depth
			{
				position26 := position
				depth++
				if !_rules[ruleLevel6]() {
					goto l25
				}
			l27:
				{
					position28, tokenIndex28, depth28 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l28
					}
					if !_rules[ruleOr]() {
						goto l28
					}
					goto l27
				l28:
					position, tokenIndex, depth = position28, tokenIndex28, depth28
				}
				depth--
				add(ruleLevel7, position26)
			}
			return true
		l25:
			position, tokenIndex, depth = position25, tokenIndex25, depth25
			return false
		},
		/* 7 Or <- <('|' '|' req_ws Level6)> */
		func() bool {
			position29, tokenIndex29, depth29 := position, tokenIndex, depth
			{
				position30 := position
				depth++
				if buffer[position] != rune('|') {
					goto l29
				}
				position++
				if buffer[position] != rune('|') {
					goto l29
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l29
				}
				if !_rules[ruleLevel6]() {
					goto l29
				}
				depth--
				add(ruleOr, position30)
			}
			return true
		l29:
			position, tokenIndex, depth = position29, tokenIndex29, depth29
			return false
		},
		/* 8 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position31, tokenIndex31, depth31 := position, tokenIndex, depth
			{
				position32 := position
				depth++
				{
					position33, tokenIndex33, depth33 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l34
					}
					goto l33
				l34:
					position, tokenIndex, depth = position33, tokenIndex33, depth33
					if !_rules[ruleLevel5]() {
						goto l31
					}
				}
			l33:
				depth--
				add(ruleLevel6, position32)
			}
			return true
		l31:
			position, tokenIndex, depth = position31, tokenIndex31, depth31
			return false
		},
		/* 9 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position35, tokenIndex35, depth35 := position, tokenIndex, depth
			{
				position36 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l35
				}
				if !_rules[rulews]() {
					goto l35
				}
				if buffer[position] != rune('?') {
					goto l35
				}
				position++
				if !_rules[ruleExpression]() {
					goto l35
				}
				if buffer[position] != rune(':') {
					goto l35
				}
				position++
				if !_rules[ruleExpression]() {
					goto l35
				}
				depth--
				add(ruleConditional, position36)
			}
			return true
		l35:
			position, tokenIndex, depth = position35, tokenIndex35, depth35
			return false
		},
		/* 10 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position37, tokenIndex37, depth37 := position, tokenIndex, depth
			{
				position38 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l37
				}
			l39:
				{
					position40, tokenIndex40, depth40 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l40
					}
					goto l39
				l40:
					position, tokenIndex, depth = position40, tokenIndex40, depth40
				}
				depth--
				add(ruleLevel5, position38)
			}
			return true
		l37:
			position, tokenIndex, depth = position37, tokenIndex37, depth37
			return false
		},
		/* 11 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position41, tokenIndex41, depth41 := position, tokenIndex, depth
			{
				position42 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l41
				}
				if !_rules[ruleLevel4]() {
					goto l41
				}
				depth--
				add(ruleConcatenation, position42)
			}
			return true
		l41:
			position, tokenIndex, depth = position41, tokenIndex41, depth41
			return false
		},
		/* 12 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position43, tokenIndex43, depth43 := position, tokenIndex, depth
			{
				position44 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l43
				}
			l45:
				{
					position46, tokenIndex46, depth46 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l46
					}
					{
						position47, tokenIndex47, depth47 := position, tokenIndex, depth
						if !_rules[ruleLogOr]() {
							goto l48
						}
						goto l47
					l48:
						position, tokenIndex, depth = position47, tokenIndex47, depth47
						if !_rules[ruleLogAnd]() {
							goto l46
						}
					}
				l47:
					goto l45
				l46:
					position, tokenIndex, depth = position46, tokenIndex46, depth46
				}
				depth--
				add(ruleLevel4, position44)
			}
			return true
		l43:
			position, tokenIndex, depth = position43, tokenIndex43, depth43
			return false
		},
		/* 13 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position49, tokenIndex49, depth49 := position, tokenIndex, depth
			{
				position50 := position
				depth++
				if buffer[position] != rune('-') {
					goto l49
				}
				position++
				if buffer[position] != rune('o') {
					goto l49
				}
				position++
				if buffer[position] != rune('r') {
					goto l49
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l49
				}
				if !_rules[ruleLevel3]() {
					goto l49
				}
				depth--
				add(ruleLogOr, position50)
			}
			return true
		l49:
			position, tokenIndex, depth = position49, tokenIndex49, depth49
			return false
		},
		/* 14 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position51, tokenIndex51, depth51 := position, tokenIndex, depth
			{
				position52 := position
				depth++
				if buffer[position] != rune('-') {
					goto l51
				}
				position++
				if buffer[position] != rune('a') {
					goto l51
				}
				position++
				if buffer[position] != rune('n') {
					goto l51
				}
				position++
				if buffer[position] != rune('d') {
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
				add(ruleLogAnd, position52)
			}
			return true
		l51:
			position, tokenIndex, depth = position51, tokenIndex51, depth51
			return false
		},
		/* 15 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position53, tokenIndex53, depth53 := position, tokenIndex, depth
			{
				position54 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l53
				}
			l55:
				{
					position56, tokenIndex56, depth56 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l56
					}
					if !_rules[ruleComparison]() {
						goto l56
					}
					goto l55
				l56:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
				}
				depth--
				add(ruleLevel3, position54)
			}
			return true
		l53:
			position, tokenIndex, depth = position53, tokenIndex53, depth53
			return false
		},
		/* 16 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position57, tokenIndex57, depth57 := position, tokenIndex, depth
			{
				position58 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l57
				}
				if !_rules[rulereq_ws]() {
					goto l57
				}
				if !_rules[ruleLevel2]() {
					goto l57
				}
				depth--
				add(ruleComparison, position58)
			}
			return true
		l57:
			position, tokenIndex, depth = position57, tokenIndex57, depth57
			return false
		},
		/* 17 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position59, tokenIndex59, depth59 := position, tokenIndex, depth
			{
				position60 := position
				depth++
				{
					position61, tokenIndex61, depth61 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l62
					}
					position++
					if buffer[position] != rune('=') {
						goto l62
					}
					position++
					goto l61
				l62:
					position, tokenIndex, depth = position61, tokenIndex61, depth61
					if buffer[position] != rune('!') {
						goto l63
					}
					position++
					if buffer[position] != rune('=') {
						goto l63
					}
					position++
					goto l61
				l63:
					position, tokenIndex, depth = position61, tokenIndex61, depth61
					if buffer[position] != rune('<') {
						goto l64
					}
					position++
					if buffer[position] != rune('=') {
						goto l64
					}
					position++
					goto l61
				l64:
					position, tokenIndex, depth = position61, tokenIndex61, depth61
					if buffer[position] != rune('>') {
						goto l65
					}
					position++
					if buffer[position] != rune('=') {
						goto l65
					}
					position++
					goto l61
				l65:
					position, tokenIndex, depth = position61, tokenIndex61, depth61
					if buffer[position] != rune('>') {
						goto l66
					}
					position++
					goto l61
				l66:
					position, tokenIndex, depth = position61, tokenIndex61, depth61
					if buffer[position] != rune('<') {
						goto l67
					}
					position++
					goto l61
				l67:
					position, tokenIndex, depth = position61, tokenIndex61, depth61
					if buffer[position] != rune('>') {
						goto l59
					}
					position++
				}
			l61:
				depth--
				add(ruleCompareOp, position60)
			}
			return true
		l59:
			position, tokenIndex, depth = position59, tokenIndex59, depth59
			return false
		},
		/* 18 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position68, tokenIndex68, depth68 := position, tokenIndex, depth
			{
				position69 := position
				depth++
				if !_rules[ruleLevel1]() {
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
						if !_rules[ruleAddition]() {
							goto l73
						}
						goto l72
					l73:
						position, tokenIndex, depth = position72, tokenIndex72, depth72
						if !_rules[ruleSubtraction]() {
							goto l71
						}
					}
				l72:
					goto l70
				l71:
					position, tokenIndex, depth = position71, tokenIndex71, depth71
				}
				depth--
				add(ruleLevel2, position69)
			}
			return true
		l68:
			position, tokenIndex, depth = position68, tokenIndex68, depth68
			return false
		},
		/* 19 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position74, tokenIndex74, depth74 := position, tokenIndex, depth
			{
				position75 := position
				depth++
				if buffer[position] != rune('+') {
					goto l74
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l74
				}
				if !_rules[ruleLevel1]() {
					goto l74
				}
				depth--
				add(ruleAddition, position75)
			}
			return true
		l74:
			position, tokenIndex, depth = position74, tokenIndex74, depth74
			return false
		},
		/* 20 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position76, tokenIndex76, depth76 := position, tokenIndex, depth
			{
				position77 := position
				depth++
				if buffer[position] != rune('-') {
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
				add(ruleSubtraction, position77)
			}
			return true
		l76:
			position, tokenIndex, depth = position76, tokenIndex76, depth76
			return false
		},
		/* 21 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position78, tokenIndex78, depth78 := position, tokenIndex, depth
			{
				position79 := position
				depth++
				if !_rules[ruleLevel0]() {
					goto l78
				}
			l80:
				{
					position81, tokenIndex81, depth81 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l81
					}
					{
						position82, tokenIndex82, depth82 := position, tokenIndex, depth
						if !_rules[ruleMultiplication]() {
							goto l83
						}
						goto l82
					l83:
						position, tokenIndex, depth = position82, tokenIndex82, depth82
						if !_rules[ruleDivision]() {
							goto l84
						}
						goto l82
					l84:
						position, tokenIndex, depth = position82, tokenIndex82, depth82
						if !_rules[ruleModulo]() {
							goto l81
						}
					}
				l82:
					goto l80
				l81:
					position, tokenIndex, depth = position81, tokenIndex81, depth81
				}
				depth--
				add(ruleLevel1, position79)
			}
			return true
		l78:
			position, tokenIndex, depth = position78, tokenIndex78, depth78
			return false
		},
		/* 22 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position85, tokenIndex85, depth85 := position, tokenIndex, depth
			{
				position86 := position
				depth++
				if buffer[position] != rune('*') {
					goto l85
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l85
				}
				if !_rules[ruleLevel0]() {
					goto l85
				}
				depth--
				add(ruleMultiplication, position86)
			}
			return true
		l85:
			position, tokenIndex, depth = position85, tokenIndex85, depth85
			return false
		},
		/* 23 Division <- <('/' req_ws Level0)> */
		func() bool {
			position87, tokenIndex87, depth87 := position, tokenIndex, depth
			{
				position88 := position
				depth++
				if buffer[position] != rune('/') {
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
				add(ruleDivision, position88)
			}
			return true
		l87:
			position, tokenIndex, depth = position87, tokenIndex87, depth87
			return false
		},
		/* 24 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position89, tokenIndex89, depth89 := position, tokenIndex, depth
			{
				position90 := position
				depth++
				if buffer[position] != rune('%') {
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
				add(ruleModulo, position90)
			}
			return true
		l89:
			position, tokenIndex, depth = position89, tokenIndex89, depth89
			return false
		},
		/* 25 Level0 <- <(String / Integer / Boolean / Undefined / Nil / Not / Substitution / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position91, tokenIndex91, depth91 := position, tokenIndex, depth
			{
				position92 := position
				depth++
				{
					position93, tokenIndex93, depth93 := position, tokenIndex, depth
					if !_rules[ruleString]() {
						goto l94
					}
					goto l93
				l94:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
					if !_rules[ruleInteger]() {
						goto l95
					}
					goto l93
				l95:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
					if !_rules[ruleBoolean]() {
						goto l96
					}
					goto l93
				l96:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
					if !_rules[ruleUndefined]() {
						goto l97
					}
					goto l93
				l97:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
					if !_rules[ruleNil]() {
						goto l98
					}
					goto l93
				l98:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
					if !_rules[ruleNot]() {
						goto l99
					}
					goto l93
				l99:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
					if !_rules[ruleSubstitution]() {
						goto l100
					}
					goto l93
				l100:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
					if !_rules[ruleMerge]() {
						goto l101
					}
					goto l93
				l101:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
					if !_rules[ruleAuto]() {
						goto l102
					}
					goto l93
				l102:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
					if !_rules[ruleLambda]() {
						goto l103
					}
					goto l93
				l103:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
					if !_rules[ruleChained]() {
						goto l91
					}
				}
			l93:
				depth--
				add(ruleLevel0, position92)
			}
			return true
		l91:
			position, tokenIndex, depth = position91, tokenIndex91, depth91
			return false
		},
		/* 26 Chained <- <((Mapping / Sum / List / Map / Range / Grouped / Reference) ChainedQualifiedExpression*)> */
		func() bool {
			position104, tokenIndex104, depth104 := position, tokenIndex, depth
			{
				position105 := position
				depth++
				{
					position106, tokenIndex106, depth106 := position, tokenIndex, depth
					if !_rules[ruleMapping]() {
						goto l107
					}
					goto l106
				l107:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleSum]() {
						goto l108
					}
					goto l106
				l108:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleList]() {
						goto l109
					}
					goto l106
				l109:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleMap]() {
						goto l110
					}
					goto l106
				l110:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleRange]() {
						goto l111
					}
					goto l106
				l111:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleGrouped]() {
						goto l112
					}
					goto l106
				l112:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
					if !_rules[ruleReference]() {
						goto l104
					}
				}
			l106:
			l113:
				{
					position114, tokenIndex114, depth114 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l114
					}
					goto l113
				l114:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
				}
				depth--
				add(ruleChained, position105)
			}
			return true
		l104:
			position, tokenIndex, depth = position104, tokenIndex104, depth104
			return false
		},
		/* 27 ChainedQualifiedExpression <- <(ChainedCall / ('.' (ChainedRef / ChainedDynRef / Slice)))> */
		func() bool {
			position115, tokenIndex115, depth115 := position, tokenIndex, depth
			{
				position116 := position
				depth++
				{
					position117, tokenIndex117, depth117 := position, tokenIndex, depth
					if !_rules[ruleChainedCall]() {
						goto l118
					}
					goto l117
				l118:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
					if buffer[position] != rune('.') {
						goto l115
					}
					position++
					{
						position119, tokenIndex119, depth119 := position, tokenIndex, depth
						if !_rules[ruleChainedRef]() {
							goto l120
						}
						goto l119
					l120:
						position, tokenIndex, depth = position119, tokenIndex119, depth119
						if !_rules[ruleChainedDynRef]() {
							goto l121
						}
						goto l119
					l121:
						position, tokenIndex, depth = position119, tokenIndex119, depth119
						if !_rules[ruleSlice]() {
							goto l115
						}
					}
				l119:
				}
			l117:
				depth--
				add(ruleChainedQualifiedExpression, position116)
			}
			return true
		l115:
			position, tokenIndex, depth = position115, tokenIndex115, depth115
			return false
		},
		/* 28 ChainedRef <- <((Key / Index) FollowUpRef)> */
		func() bool {
			position122, tokenIndex122, depth122 := position, tokenIndex, depth
			{
				position123 := position
				depth++
				{
					position124, tokenIndex124, depth124 := position, tokenIndex, depth
					if !_rules[ruleKey]() {
						goto l125
					}
					goto l124
				l125:
					position, tokenIndex, depth = position124, tokenIndex124, depth124
					if !_rules[ruleIndex]() {
						goto l122
					}
				}
			l124:
				if !_rules[ruleFollowUpRef]() {
					goto l122
				}
				depth--
				add(ruleChainedRef, position123)
			}
			return true
		l122:
			position, tokenIndex, depth = position122, tokenIndex122, depth122
			return false
		},
		/* 29 ChainedDynRef <- <('[' Expression ']')> */
		func() bool {
			position126, tokenIndex126, depth126 := position, tokenIndex, depth
			{
				position127 := position
				depth++
				if buffer[position] != rune('[') {
					goto l126
				}
				position++
				if !_rules[ruleExpression]() {
					goto l126
				}
				if buffer[position] != rune(']') {
					goto l126
				}
				position++
				depth--
				add(ruleChainedDynRef, position127)
			}
			return true
		l126:
			position, tokenIndex, depth = position126, tokenIndex126, depth126
			return false
		},
		/* 30 Slice <- <Range> */
		func() bool {
			position128, tokenIndex128, depth128 := position, tokenIndex, depth
			{
				position129 := position
				depth++
				if !_rules[ruleRange]() {
					goto l128
				}
				depth--
				add(ruleSlice, position129)
			}
			return true
		l128:
			position, tokenIndex, depth = position128, tokenIndex128, depth128
			return false
		},
		/* 31 ChainedCall <- <('(' Arguments ')')> */
		func() bool {
			position130, tokenIndex130, depth130 := position, tokenIndex, depth
			{
				position131 := position
				depth++
				if buffer[position] != rune('(') {
					goto l130
				}
				position++
				if !_rules[ruleArguments]() {
					goto l130
				}
				if buffer[position] != rune(')') {
					goto l130
				}
				position++
				depth--
				add(ruleChainedCall, position131)
			}
			return true
		l130:
			position, tokenIndex, depth = position130, tokenIndex130, depth130
			return false
		},
		/* 32 Arguments <- <(Expression NextExpression*)> */
		func() bool {
			position132, tokenIndex132, depth132 := position, tokenIndex, depth
			{
				position133 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l132
				}
			l134:
				{
					position135, tokenIndex135, depth135 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l135
					}
					goto l134
				l135:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
				}
				depth--
				add(ruleArguments, position133)
			}
			return true
		l132:
			position, tokenIndex, depth = position132, tokenIndex132, depth132
			return false
		},
		/* 33 NextExpression <- <(',' Expression)> */
		func() bool {
			position136, tokenIndex136, depth136 := position, tokenIndex, depth
			{
				position137 := position
				depth++
				if buffer[position] != rune(',') {
					goto l136
				}
				position++
				if !_rules[ruleExpression]() {
					goto l136
				}
				depth--
				add(ruleNextExpression, position137)
			}
			return true
		l136:
			position, tokenIndex, depth = position136, tokenIndex136, depth136
			return false
		},
		/* 34 Substitution <- <('*' Level0)> */
		func() bool {
			position138, tokenIndex138, depth138 := position, tokenIndex, depth
			{
				position139 := position
				depth++
				if buffer[position] != rune('*') {
					goto l138
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l138
				}
				depth--
				add(ruleSubstitution, position139)
			}
			return true
		l138:
			position, tokenIndex, depth = position138, tokenIndex138, depth138
			return false
		},
		/* 35 Not <- <('!' ws Level0)> */
		func() bool {
			position140, tokenIndex140, depth140 := position, tokenIndex, depth
			{
				position141 := position
				depth++
				if buffer[position] != rune('!') {
					goto l140
				}
				position++
				if !_rules[rulews]() {
					goto l140
				}
				if !_rules[ruleLevel0]() {
					goto l140
				}
				depth--
				add(ruleNot, position141)
			}
			return true
		l140:
			position, tokenIndex, depth = position140, tokenIndex140, depth140
			return false
		},
		/* 36 Grouped <- <('(' Expression ')')> */
		func() bool {
			position142, tokenIndex142, depth142 := position, tokenIndex, depth
			{
				position143 := position
				depth++
				if buffer[position] != rune('(') {
					goto l142
				}
				position++
				if !_rules[ruleExpression]() {
					goto l142
				}
				if buffer[position] != rune(')') {
					goto l142
				}
				position++
				depth--
				add(ruleGrouped, position143)
			}
			return true
		l142:
			position, tokenIndex, depth = position142, tokenIndex142, depth142
			return false
		},
		/* 37 Range <- <('[' Expression ('.' '.') Expression ']')> */
		func() bool {
			position144, tokenIndex144, depth144 := position, tokenIndex, depth
			{
				position145 := position
				depth++
				if buffer[position] != rune('[') {
					goto l144
				}
				position++
				if !_rules[ruleExpression]() {
					goto l144
				}
				if buffer[position] != rune('.') {
					goto l144
				}
				position++
				if buffer[position] != rune('.') {
					goto l144
				}
				position++
				if !_rules[ruleExpression]() {
					goto l144
				}
				if buffer[position] != rune(']') {
					goto l144
				}
				position++
				depth--
				add(ruleRange, position145)
			}
			return true
		l144:
			position, tokenIndex, depth = position144, tokenIndex144, depth144
			return false
		},
		/* 38 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position146, tokenIndex146, depth146 := position, tokenIndex, depth
			{
				position147 := position
				depth++
				{
					position148, tokenIndex148, depth148 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l148
					}
					position++
					goto l149
				l148:
					position, tokenIndex, depth = position148, tokenIndex148, depth148
				}
			l149:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l146
				}
				position++
			l150:
				{
					position151, tokenIndex151, depth151 := position, tokenIndex, depth
					{
						position152, tokenIndex152, depth152 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l153
						}
						position++
						goto l152
					l153:
						position, tokenIndex, depth = position152, tokenIndex152, depth152
						if buffer[position] != rune('_') {
							goto l151
						}
						position++
					}
				l152:
					goto l150
				l151:
					position, tokenIndex, depth = position151, tokenIndex151, depth151
				}
				depth--
				add(ruleInteger, position147)
			}
			return true
		l146:
			position, tokenIndex, depth = position146, tokenIndex146, depth146
			return false
		},
		/* 39 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position154, tokenIndex154, depth154 := position, tokenIndex, depth
			{
				position155 := position
				depth++
				if buffer[position] != rune('"') {
					goto l154
				}
				position++
			l156:
				{
					position157, tokenIndex157, depth157 := position, tokenIndex, depth
					{
						position158, tokenIndex158, depth158 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l159
						}
						position++
						if buffer[position] != rune('"') {
							goto l159
						}
						position++
						goto l158
					l159:
						position, tokenIndex, depth = position158, tokenIndex158, depth158
						{
							position160, tokenIndex160, depth160 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l160
							}
							position++
							goto l157
						l160:
							position, tokenIndex, depth = position160, tokenIndex160, depth160
						}
						if !matchDot() {
							goto l157
						}
					}
				l158:
					goto l156
				l157:
					position, tokenIndex, depth = position157, tokenIndex157, depth157
				}
				if buffer[position] != rune('"') {
					goto l154
				}
				position++
				depth--
				add(ruleString, position155)
			}
			return true
		l154:
			position, tokenIndex, depth = position154, tokenIndex154, depth154
			return false
		},
		/* 40 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position161, tokenIndex161, depth161 := position, tokenIndex, depth
			{
				position162 := position
				depth++
				{
					position163, tokenIndex163, depth163 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l164
					}
					position++
					if buffer[position] != rune('r') {
						goto l164
					}
					position++
					if buffer[position] != rune('u') {
						goto l164
					}
					position++
					if buffer[position] != rune('e') {
						goto l164
					}
					position++
					goto l163
				l164:
					position, tokenIndex, depth = position163, tokenIndex163, depth163
					if buffer[position] != rune('f') {
						goto l161
					}
					position++
					if buffer[position] != rune('a') {
						goto l161
					}
					position++
					if buffer[position] != rune('l') {
						goto l161
					}
					position++
					if buffer[position] != rune('s') {
						goto l161
					}
					position++
					if buffer[position] != rune('e') {
						goto l161
					}
					position++
				}
			l163:
				depth--
				add(ruleBoolean, position162)
			}
			return true
		l161:
			position, tokenIndex, depth = position161, tokenIndex161, depth161
			return false
		},
		/* 41 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position165, tokenIndex165, depth165 := position, tokenIndex, depth
			{
				position166 := position
				depth++
				{
					position167, tokenIndex167, depth167 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l168
					}
					position++
					if buffer[position] != rune('i') {
						goto l168
					}
					position++
					if buffer[position] != rune('l') {
						goto l168
					}
					position++
					goto l167
				l168:
					position, tokenIndex, depth = position167, tokenIndex167, depth167
					if buffer[position] != rune('~') {
						goto l165
					}
					position++
				}
			l167:
				depth--
				add(ruleNil, position166)
			}
			return true
		l165:
			position, tokenIndex, depth = position165, tokenIndex165, depth165
			return false
		},
		/* 42 Undefined <- <('~' '~')> */
		func() bool {
			position169, tokenIndex169, depth169 := position, tokenIndex, depth
			{
				position170 := position
				depth++
				if buffer[position] != rune('~') {
					goto l169
				}
				position++
				if buffer[position] != rune('~') {
					goto l169
				}
				position++
				depth--
				add(ruleUndefined, position170)
			}
			return true
		l169:
			position, tokenIndex, depth = position169, tokenIndex169, depth169
			return false
		},
		/* 43 List <- <('[' Contents? ']')> */
		func() bool {
			position171, tokenIndex171, depth171 := position, tokenIndex, depth
			{
				position172 := position
				depth++
				if buffer[position] != rune('[') {
					goto l171
				}
				position++
				{
					position173, tokenIndex173, depth173 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l173
					}
					goto l174
				l173:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
				}
			l174:
				if buffer[position] != rune(']') {
					goto l171
				}
				position++
				depth--
				add(ruleList, position172)
			}
			return true
		l171:
			position, tokenIndex, depth = position171, tokenIndex171, depth171
			return false
		},
		/* 44 Contents <- <(Expression NextExpression*)> */
		func() bool {
			position175, tokenIndex175, depth175 := position, tokenIndex, depth
			{
				position176 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l175
				}
			l177:
				{
					position178, tokenIndex178, depth178 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l178
					}
					goto l177
				l178:
					position, tokenIndex, depth = position178, tokenIndex178, depth178
				}
				depth--
				add(ruleContents, position176)
			}
			return true
		l175:
			position, tokenIndex, depth = position175, tokenIndex175, depth175
			return false
		},
		/* 45 Map <- <(CreateMap ws Assignments? '}')> */
		func() bool {
			position179, tokenIndex179, depth179 := position, tokenIndex, depth
			{
				position180 := position
				depth++
				if !_rules[ruleCreateMap]() {
					goto l179
				}
				if !_rules[rulews]() {
					goto l179
				}
				{
					position181, tokenIndex181, depth181 := position, tokenIndex, depth
					if !_rules[ruleAssignments]() {
						goto l181
					}
					goto l182
				l181:
					position, tokenIndex, depth = position181, tokenIndex181, depth181
				}
			l182:
				if buffer[position] != rune('}') {
					goto l179
				}
				position++
				depth--
				add(ruleMap, position180)
			}
			return true
		l179:
			position, tokenIndex, depth = position179, tokenIndex179, depth179
			return false
		},
		/* 46 CreateMap <- <'{'> */
		func() bool {
			position183, tokenIndex183, depth183 := position, tokenIndex, depth
			{
				position184 := position
				depth++
				if buffer[position] != rune('{') {
					goto l183
				}
				position++
				depth--
				add(ruleCreateMap, position184)
			}
			return true
		l183:
			position, tokenIndex, depth = position183, tokenIndex183, depth183
			return false
		},
		/* 47 Assignments <- <(Assignment (',' Assignment)*)> */
		func() bool {
			position185, tokenIndex185, depth185 := position, tokenIndex, depth
			{
				position186 := position
				depth++
				if !_rules[ruleAssignment]() {
					goto l185
				}
			l187:
				{
					position188, tokenIndex188, depth188 := position, tokenIndex, depth
					if buffer[position] != rune(',') {
						goto l188
					}
					position++
					if !_rules[ruleAssignment]() {
						goto l188
					}
					goto l187
				l188:
					position, tokenIndex, depth = position188, tokenIndex188, depth188
				}
				depth--
				add(ruleAssignments, position186)
			}
			return true
		l185:
			position, tokenIndex, depth = position185, tokenIndex185, depth185
			return false
		},
		/* 48 Assignment <- <(Expression '=' Expression)> */
		func() bool {
			position189, tokenIndex189, depth189 := position, tokenIndex, depth
			{
				position190 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l189
				}
				if buffer[position] != rune('=') {
					goto l189
				}
				position++
				if !_rules[ruleExpression]() {
					goto l189
				}
				depth--
				add(ruleAssignment, position190)
			}
			return true
		l189:
			position, tokenIndex, depth = position189, tokenIndex189, depth189
			return false
		},
		/* 49 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position191, tokenIndex191, depth191 := position, tokenIndex, depth
			{
				position192 := position
				depth++
				{
					position193, tokenIndex193, depth193 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l194
					}
					goto l193
				l194:
					position, tokenIndex, depth = position193, tokenIndex193, depth193
					if !_rules[ruleSimpleMerge]() {
						goto l191
					}
				}
			l193:
				depth--
				add(ruleMerge, position192)
			}
			return true
		l191:
			position, tokenIndex, depth = position191, tokenIndex191, depth191
			return false
		},
		/* 50 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position195, tokenIndex195, depth195 := position, tokenIndex, depth
			{
				position196 := position
				depth++
				if buffer[position] != rune('m') {
					goto l195
				}
				position++
				if buffer[position] != rune('e') {
					goto l195
				}
				position++
				if buffer[position] != rune('r') {
					goto l195
				}
				position++
				if buffer[position] != rune('g') {
					goto l195
				}
				position++
				if buffer[position] != rune('e') {
					goto l195
				}
				position++
				{
					position197, tokenIndex197, depth197 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l197
					}
					if !_rules[ruleRequired]() {
						goto l197
					}
					goto l195
				l197:
					position, tokenIndex, depth = position197, tokenIndex197, depth197
				}
				{
					position198, tokenIndex198, depth198 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l198
					}
					{
						position200, tokenIndex200, depth200 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l201
						}
						goto l200
					l201:
						position, tokenIndex, depth = position200, tokenIndex200, depth200
						if !_rules[ruleOn]() {
							goto l198
						}
					}
				l200:
					goto l199
				l198:
					position, tokenIndex, depth = position198, tokenIndex198, depth198
				}
			l199:
				if !_rules[rulereq_ws]() {
					goto l195
				}
				if !_rules[ruleReference]() {
					goto l195
				}
				depth--
				add(ruleRefMerge, position196)
			}
			return true
		l195:
			position, tokenIndex, depth = position195, tokenIndex195, depth195
			return false
		},
		/* 51 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position202, tokenIndex202, depth202 := position, tokenIndex, depth
			{
				position203 := position
				depth++
				if buffer[position] != rune('m') {
					goto l202
				}
				position++
				if buffer[position] != rune('e') {
					goto l202
				}
				position++
				if buffer[position] != rune('r') {
					goto l202
				}
				position++
				if buffer[position] != rune('g') {
					goto l202
				}
				position++
				if buffer[position] != rune('e') {
					goto l202
				}
				position++
				{
					position204, tokenIndex204, depth204 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l204
					}
					{
						position206, tokenIndex206, depth206 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l207
						}
						goto l206
					l207:
						position, tokenIndex, depth = position206, tokenIndex206, depth206
						if !_rules[ruleRequired]() {
							goto l208
						}
						goto l206
					l208:
						position, tokenIndex, depth = position206, tokenIndex206, depth206
						if !_rules[ruleOn]() {
							goto l204
						}
					}
				l206:
					goto l205
				l204:
					position, tokenIndex, depth = position204, tokenIndex204, depth204
				}
			l205:
				depth--
				add(ruleSimpleMerge, position203)
			}
			return true
		l202:
			position, tokenIndex, depth = position202, tokenIndex202, depth202
			return false
		},
		/* 52 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position209, tokenIndex209, depth209 := position, tokenIndex, depth
			{
				position210 := position
				depth++
				if buffer[position] != rune('r') {
					goto l209
				}
				position++
				if buffer[position] != rune('e') {
					goto l209
				}
				position++
				if buffer[position] != rune('p') {
					goto l209
				}
				position++
				if buffer[position] != rune('l') {
					goto l209
				}
				position++
				if buffer[position] != rune('a') {
					goto l209
				}
				position++
				if buffer[position] != rune('c') {
					goto l209
				}
				position++
				if buffer[position] != rune('e') {
					goto l209
				}
				position++
				depth--
				add(ruleReplace, position210)
			}
			return true
		l209:
			position, tokenIndex, depth = position209, tokenIndex209, depth209
			return false
		},
		/* 53 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position211, tokenIndex211, depth211 := position, tokenIndex, depth
			{
				position212 := position
				depth++
				if buffer[position] != rune('r') {
					goto l211
				}
				position++
				if buffer[position] != rune('e') {
					goto l211
				}
				position++
				if buffer[position] != rune('q') {
					goto l211
				}
				position++
				if buffer[position] != rune('u') {
					goto l211
				}
				position++
				if buffer[position] != rune('i') {
					goto l211
				}
				position++
				if buffer[position] != rune('r') {
					goto l211
				}
				position++
				if buffer[position] != rune('e') {
					goto l211
				}
				position++
				if buffer[position] != rune('d') {
					goto l211
				}
				position++
				depth--
				add(ruleRequired, position212)
			}
			return true
		l211:
			position, tokenIndex, depth = position211, tokenIndex211, depth211
			return false
		},
		/* 54 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position213, tokenIndex213, depth213 := position, tokenIndex, depth
			{
				position214 := position
				depth++
				if buffer[position] != rune('o') {
					goto l213
				}
				position++
				if buffer[position] != rune('n') {
					goto l213
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l213
				}
				if !_rules[ruleName]() {
					goto l213
				}
				depth--
				add(ruleOn, position214)
			}
			return true
		l213:
			position, tokenIndex, depth = position213, tokenIndex213, depth213
			return false
		},
		/* 55 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position215, tokenIndex215, depth215 := position, tokenIndex, depth
			{
				position216 := position
				depth++
				if buffer[position] != rune('a') {
					goto l215
				}
				position++
				if buffer[position] != rune('u') {
					goto l215
				}
				position++
				if buffer[position] != rune('t') {
					goto l215
				}
				position++
				if buffer[position] != rune('o') {
					goto l215
				}
				position++
				depth--
				add(ruleAuto, position216)
			}
			return true
		l215:
			position, tokenIndex, depth = position215, tokenIndex215, depth215
			return false
		},
		/* 56 Mapping <- <('m' 'a' 'p' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position217, tokenIndex217, depth217 := position, tokenIndex, depth
			{
				position218 := position
				depth++
				if buffer[position] != rune('m') {
					goto l217
				}
				position++
				if buffer[position] != rune('a') {
					goto l217
				}
				position++
				if buffer[position] != rune('p') {
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
				add(ruleMapping, position218)
			}
			return true
		l217:
			position, tokenIndex, depth = position217, tokenIndex217, depth217
			return false
		},
		/* 57 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position221, tokenIndex221, depth221 := position, tokenIndex, depth
			{
				position222 := position
				depth++
				if buffer[position] != rune('s') {
					goto l221
				}
				position++
				if buffer[position] != rune('u') {
					goto l221
				}
				position++
				if buffer[position] != rune('m') {
					goto l221
				}
				position++
				if buffer[position] != rune('[') {
					goto l221
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l221
				}
				if buffer[position] != rune('|') {
					goto l221
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l221
				}
				{
					position223, tokenIndex223, depth223 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l224
					}
					goto l223
				l224:
					position, tokenIndex, depth = position223, tokenIndex223, depth223
					if buffer[position] != rune('|') {
						goto l221
					}
					position++
					if !_rules[ruleExpression]() {
						goto l221
					}
				}
			l223:
				if buffer[position] != rune(']') {
					goto l221
				}
				position++
				depth--
				add(ruleSum, position222)
			}
			return true
		l221:
			position, tokenIndex, depth = position221, tokenIndex221, depth221
			return false
		},
		/* 58 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position225, tokenIndex225, depth225 := position, tokenIndex, depth
			{
				position226 := position
				depth++
				if buffer[position] != rune('l') {
					goto l225
				}
				position++
				if buffer[position] != rune('a') {
					goto l225
				}
				position++
				if buffer[position] != rune('m') {
					goto l225
				}
				position++
				if buffer[position] != rune('b') {
					goto l225
				}
				position++
				if buffer[position] != rune('d') {
					goto l225
				}
				position++
				if buffer[position] != rune('a') {
					goto l225
				}
				position++
				{
					position227, tokenIndex227, depth227 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l228
					}
					goto l227
				l228:
					position, tokenIndex, depth = position227, tokenIndex227, depth227
					if !_rules[ruleLambdaExpr]() {
						goto l225
					}
				}
			l227:
				depth--
				add(ruleLambda, position226)
			}
			return true
		l225:
			position, tokenIndex, depth = position225, tokenIndex225, depth225
			return false
		},
		/* 59 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position229, tokenIndex229, depth229 := position, tokenIndex, depth
			{
				position230 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l229
				}
				if !_rules[ruleExpression]() {
					goto l229
				}
				depth--
				add(ruleLambdaRef, position230)
			}
			return true
		l229:
			position, tokenIndex, depth = position229, tokenIndex229, depth229
			return false
		},
		/* 60 LambdaExpr <- <(ws '|' ws Name NextName* ws '|' ws ('-' '>') Expression)> */
		func() bool {
			position231, tokenIndex231, depth231 := position, tokenIndex, depth
			{
				position232 := position
				depth++
				if !_rules[rulews]() {
					goto l231
				}
				if buffer[position] != rune('|') {
					goto l231
				}
				position++
				if !_rules[rulews]() {
					goto l231
				}
				if !_rules[ruleName]() {
					goto l231
				}
			l233:
				{
					position234, tokenIndex234, depth234 := position, tokenIndex, depth
					if !_rules[ruleNextName]() {
						goto l234
					}
					goto l233
				l234:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
				}
				if !_rules[rulews]() {
					goto l231
				}
				if buffer[position] != rune('|') {
					goto l231
				}
				position++
				if !_rules[rulews]() {
					goto l231
				}
				if buffer[position] != rune('-') {
					goto l231
				}
				position++
				if buffer[position] != rune('>') {
					goto l231
				}
				position++
				if !_rules[ruleExpression]() {
					goto l231
				}
				depth--
				add(ruleLambdaExpr, position232)
			}
			return true
		l231:
			position, tokenIndex, depth = position231, tokenIndex231, depth231
			return false
		},
		/* 61 NextName <- <(ws ',' ws Name)> */
		func() bool {
			position235, tokenIndex235, depth235 := position, tokenIndex, depth
			{
				position236 := position
				depth++
				if !_rules[rulews]() {
					goto l235
				}
				if buffer[position] != rune(',') {
					goto l235
				}
				position++
				if !_rules[rulews]() {
					goto l235
				}
				if !_rules[ruleName]() {
					goto l235
				}
				depth--
				add(ruleNextName, position236)
			}
			return true
		l235:
			position, tokenIndex, depth = position235, tokenIndex235, depth235
			return false
		},
		/* 62 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position237, tokenIndex237, depth237 := position, tokenIndex, depth
			{
				position238 := position
				depth++
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
						goto l237
					}
					position++
				}
			l241:
			l239:
				{
					position240, tokenIndex240, depth240 := position, tokenIndex, depth
					{
						position245, tokenIndex245, depth245 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l246
						}
						position++
						goto l245
					l246:
						position, tokenIndex, depth = position245, tokenIndex245, depth245
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l247
						}
						position++
						goto l245
					l247:
						position, tokenIndex, depth = position245, tokenIndex245, depth245
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l248
						}
						position++
						goto l245
					l248:
						position, tokenIndex, depth = position245, tokenIndex245, depth245
						if buffer[position] != rune('_') {
							goto l240
						}
						position++
					}
				l245:
					goto l239
				l240:
					position, tokenIndex, depth = position240, tokenIndex240, depth240
				}
				depth--
				add(ruleName, position238)
			}
			return true
		l237:
			position, tokenIndex, depth = position237, tokenIndex237, depth237
			return false
		},
		/* 63 Reference <- <('.'? Key FollowUpRef)> */
		func() bool {
			position249, tokenIndex249, depth249 := position, tokenIndex, depth
			{
				position250 := position
				depth++
				{
					position251, tokenIndex251, depth251 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l251
					}
					position++
					goto l252
				l251:
					position, tokenIndex, depth = position251, tokenIndex251, depth251
				}
			l252:
				if !_rules[ruleKey]() {
					goto l249
				}
				if !_rules[ruleFollowUpRef]() {
					goto l249
				}
				depth--
				add(ruleReference, position250)
			}
			return true
		l249:
			position, tokenIndex, depth = position249, tokenIndex249, depth249
			return false
		},
		/* 64 FollowUpRef <- <('.' (Key / Index))*> */
		func() bool {
			{
				position254 := position
				depth++
			l255:
				{
					position256, tokenIndex256, depth256 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l256
					}
					position++
					{
						position257, tokenIndex257, depth257 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l258
						}
						goto l257
					l258:
						position, tokenIndex, depth = position257, tokenIndex257, depth257
						if !_rules[ruleIndex]() {
							goto l256
						}
					}
				l257:
					goto l255
				l256:
					position, tokenIndex, depth = position256, tokenIndex256, depth256
				}
				depth--
				add(ruleFollowUpRef, position254)
			}
			return true
		},
		/* 65 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position259, tokenIndex259, depth259 := position, tokenIndex, depth
			{
				position260 := position
				depth++
				{
					position261, tokenIndex261, depth261 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l262
					}
					position++
					goto l261
				l262:
					position, tokenIndex, depth = position261, tokenIndex261, depth261
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l263
					}
					position++
					goto l261
				l263:
					position, tokenIndex, depth = position261, tokenIndex261, depth261
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l264
					}
					position++
					goto l261
				l264:
					position, tokenIndex, depth = position261, tokenIndex261, depth261
					if buffer[position] != rune('_') {
						goto l259
					}
					position++
				}
			l261:
			l265:
				{
					position266, tokenIndex266, depth266 := position, tokenIndex, depth
					{
						position267, tokenIndex267, depth267 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l268
						}
						position++
						goto l267
					l268:
						position, tokenIndex, depth = position267, tokenIndex267, depth267
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l269
						}
						position++
						goto l267
					l269:
						position, tokenIndex, depth = position267, tokenIndex267, depth267
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l270
						}
						position++
						goto l267
					l270:
						position, tokenIndex, depth = position267, tokenIndex267, depth267
						if buffer[position] != rune('_') {
							goto l271
						}
						position++
						goto l267
					l271:
						position, tokenIndex, depth = position267, tokenIndex267, depth267
						if buffer[position] != rune('-') {
							goto l266
						}
						position++
					}
				l267:
					goto l265
				l266:
					position, tokenIndex, depth = position266, tokenIndex266, depth266
				}
				{
					position272, tokenIndex272, depth272 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l272
					}
					position++
					{
						position274, tokenIndex274, depth274 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l275
						}
						position++
						goto l274
					l275:
						position, tokenIndex, depth = position274, tokenIndex274, depth274
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l276
						}
						position++
						goto l274
					l276:
						position, tokenIndex, depth = position274, tokenIndex274, depth274
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l277
						}
						position++
						goto l274
					l277:
						position, tokenIndex, depth = position274, tokenIndex274, depth274
						if buffer[position] != rune('_') {
							goto l272
						}
						position++
					}
				l274:
				l278:
					{
						position279, tokenIndex279, depth279 := position, tokenIndex, depth
						{
							position280, tokenIndex280, depth280 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l281
							}
							position++
							goto l280
						l281:
							position, tokenIndex, depth = position280, tokenIndex280, depth280
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l282
							}
							position++
							goto l280
						l282:
							position, tokenIndex, depth = position280, tokenIndex280, depth280
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l283
							}
							position++
							goto l280
						l283:
							position, tokenIndex, depth = position280, tokenIndex280, depth280
							if buffer[position] != rune('_') {
								goto l284
							}
							position++
							goto l280
						l284:
							position, tokenIndex, depth = position280, tokenIndex280, depth280
							if buffer[position] != rune('-') {
								goto l279
							}
							position++
						}
					l280:
						goto l278
					l279:
						position, tokenIndex, depth = position279, tokenIndex279, depth279
					}
					goto l273
				l272:
					position, tokenIndex, depth = position272, tokenIndex272, depth272
				}
			l273:
				depth--
				add(ruleKey, position260)
			}
			return true
		l259:
			position, tokenIndex, depth = position259, tokenIndex259, depth259
			return false
		},
		/* 66 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position285, tokenIndex285, depth285 := position, tokenIndex, depth
			{
				position286 := position
				depth++
				if buffer[position] != rune('[') {
					goto l285
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l285
				}
				position++
			l287:
				{
					position288, tokenIndex288, depth288 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l288
					}
					position++
					goto l287
				l288:
					position, tokenIndex, depth = position288, tokenIndex288, depth288
				}
				if buffer[position] != rune(']') {
					goto l285
				}
				position++
				depth--
				add(ruleIndex, position286)
			}
			return true
		l285:
			position, tokenIndex, depth = position285, tokenIndex285, depth285
			return false
		},
		/* 67 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position290 := position
				depth++
			l291:
				{
					position292, tokenIndex292, depth292 := position, tokenIndex, depth
					{
						position293, tokenIndex293, depth293 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l294
						}
						position++
						goto l293
					l294:
						position, tokenIndex, depth = position293, tokenIndex293, depth293
						if buffer[position] != rune('\t') {
							goto l295
						}
						position++
						goto l293
					l295:
						position, tokenIndex, depth = position293, tokenIndex293, depth293
						if buffer[position] != rune('\n') {
							goto l296
						}
						position++
						goto l293
					l296:
						position, tokenIndex, depth = position293, tokenIndex293, depth293
						if buffer[position] != rune('\r') {
							goto l292
						}
						position++
					}
				l293:
					goto l291
				l292:
					position, tokenIndex, depth = position292, tokenIndex292, depth292
				}
				depth--
				add(rulews, position290)
			}
			return true
		},
		/* 68 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position297, tokenIndex297, depth297 := position, tokenIndex, depth
			{
				position298 := position
				depth++
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
						goto l297
					}
					position++
				}
			l301:
			l299:
				{
					position300, tokenIndex300, depth300 := position, tokenIndex, depth
					{
						position305, tokenIndex305, depth305 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l306
						}
						position++
						goto l305
					l306:
						position, tokenIndex, depth = position305, tokenIndex305, depth305
						if buffer[position] != rune('\t') {
							goto l307
						}
						position++
						goto l305
					l307:
						position, tokenIndex, depth = position305, tokenIndex305, depth305
						if buffer[position] != rune('\n') {
							goto l308
						}
						position++
						goto l305
					l308:
						position, tokenIndex, depth = position305, tokenIndex305, depth305
						if buffer[position] != rune('\r') {
							goto l300
						}
						position++
					}
				l305:
					goto l299
				l300:
					position, tokenIndex, depth = position300, tokenIndex300, depth300
				}
				depth--
				add(rulereq_ws, position298)
			}
			return true
		l297:
			position, tokenIndex, depth = position297, tokenIndex297, depth297
			return false
		},
	}
	p.rules = _rules
}
