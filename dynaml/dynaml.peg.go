package dynaml

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleDynaml
	rulePrefer
	ruleTemplate
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
	ruleEmptyHash
	ruleList
	ruleContents
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

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"Dynaml",
	"Prefer",
	"Template",
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
	"EmptyHash",
	"List",
	"Contents",
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

func (ast *node32) Print(buffer string) {
	ast.print(0, buffer)
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
		for i, _ := range states {
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
							write(token32{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
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
					write(token32{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
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
	for i, _ := range tokens {
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
	rules  [61]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range []rune(buffer) {
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
	p *DynamlGrammar
}

func (e *parseError) Error() string {
	tokens, error := e.p.tokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
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
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != end_symbol {
		p.buffer = append(p.buffer, end_symbol)
	}

	var tree tokenTree = &tokens32{tree: make([]token32, math.MaxInt16)}
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
		return &parseError{p}
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
	}

	matchDot := func() bool {
		if buffer[position] != end_symbol {
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
		/* 0 Dynaml <- <((Prefer / Template / Expression) !.)> */
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
					if !_rules[ruleTemplate]() {
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
		/* 2 Template <- <(ws ('&' 't' 'e' 'm' 'p' 'l' 'a' 't' 'e') ws)> */
		func() bool {
			position8, tokenIndex8, depth8 := position, tokenIndex, depth
			{
				position9 := position
				depth++
				if !_rules[rulews]() {
					goto l8
				}
				if buffer[position] != rune('&') {
					goto l8
				}
				position++
				if buffer[position] != rune('t') {
					goto l8
				}
				position++
				if buffer[position] != rune('e') {
					goto l8
				}
				position++
				if buffer[position] != rune('m') {
					goto l8
				}
				position++
				if buffer[position] != rune('p') {
					goto l8
				}
				position++
				if buffer[position] != rune('l') {
					goto l8
				}
				position++
				if buffer[position] != rune('a') {
					goto l8
				}
				position++
				if buffer[position] != rune('t') {
					goto l8
				}
				position++
				if buffer[position] != rune('e') {
					goto l8
				}
				position++
				if !_rules[rulews]() {
					goto l8
				}
				depth--
				add(ruleTemplate, position9)
			}
			return true
		l8:
			position, tokenIndex, depth = position8, tokenIndex8, depth8
			return false
		},
		/* 3 Expression <- <(ws (LambdaExpr / Level7) ws)> */
		func() bool {
			position10, tokenIndex10, depth10 := position, tokenIndex, depth
			{
				position11 := position
				depth++
				if !_rules[rulews]() {
					goto l10
				}
				{
					position12, tokenIndex12, depth12 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l13
					}
					goto l12
				l13:
					position, tokenIndex, depth = position12, tokenIndex12, depth12
					if !_rules[ruleLevel7]() {
						goto l10
					}
				}
			l12:
				if !_rules[rulews]() {
					goto l10
				}
				depth--
				add(ruleExpression, position11)
			}
			return true
		l10:
			position, tokenIndex, depth = position10, tokenIndex10, depth10
			return false
		},
		/* 4 Level7 <- <(Level6 (req_ws Or)*)> */
		func() bool {
			position14, tokenIndex14, depth14 := position, tokenIndex, depth
			{
				position15 := position
				depth++
				if !_rules[ruleLevel6]() {
					goto l14
				}
			l16:
				{
					position17, tokenIndex17, depth17 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l17
					}
					if !_rules[ruleOr]() {
						goto l17
					}
					goto l16
				l17:
					position, tokenIndex, depth = position17, tokenIndex17, depth17
				}
				depth--
				add(ruleLevel7, position15)
			}
			return true
		l14:
			position, tokenIndex, depth = position14, tokenIndex14, depth14
			return false
		},
		/* 5 Or <- <('|' '|' req_ws Level6)> */
		func() bool {
			position18, tokenIndex18, depth18 := position, tokenIndex, depth
			{
				position19 := position
				depth++
				if buffer[position] != rune('|') {
					goto l18
				}
				position++
				if buffer[position] != rune('|') {
					goto l18
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l18
				}
				if !_rules[ruleLevel6]() {
					goto l18
				}
				depth--
				add(ruleOr, position19)
			}
			return true
		l18:
			position, tokenIndex, depth = position18, tokenIndex18, depth18
			return false
		},
		/* 6 Level6 <- <(Conditional / Level5)> */
		func() bool {
			position20, tokenIndex20, depth20 := position, tokenIndex, depth
			{
				position21 := position
				depth++
				{
					position22, tokenIndex22, depth22 := position, tokenIndex, depth
					if !_rules[ruleConditional]() {
						goto l23
					}
					goto l22
				l23:
					position, tokenIndex, depth = position22, tokenIndex22, depth22
					if !_rules[ruleLevel5]() {
						goto l20
					}
				}
			l22:
				depth--
				add(ruleLevel6, position21)
			}
			return true
		l20:
			position, tokenIndex, depth = position20, tokenIndex20, depth20
			return false
		},
		/* 7 Conditional <- <(Level5 ws '?' Expression ':' Expression)> */
		func() bool {
			position24, tokenIndex24, depth24 := position, tokenIndex, depth
			{
				position25 := position
				depth++
				if !_rules[ruleLevel5]() {
					goto l24
				}
				if !_rules[rulews]() {
					goto l24
				}
				if buffer[position] != rune('?') {
					goto l24
				}
				position++
				if !_rules[ruleExpression]() {
					goto l24
				}
				if buffer[position] != rune(':') {
					goto l24
				}
				position++
				if !_rules[ruleExpression]() {
					goto l24
				}
				depth--
				add(ruleConditional, position25)
			}
			return true
		l24:
			position, tokenIndex, depth = position24, tokenIndex24, depth24
			return false
		},
		/* 8 Level5 <- <(Level4 Concatenation*)> */
		func() bool {
			position26, tokenIndex26, depth26 := position, tokenIndex, depth
			{
				position27 := position
				depth++
				if !_rules[ruleLevel4]() {
					goto l26
				}
			l28:
				{
					position29, tokenIndex29, depth29 := position, tokenIndex, depth
					if !_rules[ruleConcatenation]() {
						goto l29
					}
					goto l28
				l29:
					position, tokenIndex, depth = position29, tokenIndex29, depth29
				}
				depth--
				add(ruleLevel5, position27)
			}
			return true
		l26:
			position, tokenIndex, depth = position26, tokenIndex26, depth26
			return false
		},
		/* 9 Concatenation <- <(req_ws Level4)> */
		func() bool {
			position30, tokenIndex30, depth30 := position, tokenIndex, depth
			{
				position31 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l30
				}
				if !_rules[ruleLevel4]() {
					goto l30
				}
				depth--
				add(ruleConcatenation, position31)
			}
			return true
		l30:
			position, tokenIndex, depth = position30, tokenIndex30, depth30
			return false
		},
		/* 10 Level4 <- <(Level3 (req_ws (LogOr / LogAnd))*)> */
		func() bool {
			position32, tokenIndex32, depth32 := position, tokenIndex, depth
			{
				position33 := position
				depth++
				if !_rules[ruleLevel3]() {
					goto l32
				}
			l34:
				{
					position35, tokenIndex35, depth35 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l35
					}
					{
						position36, tokenIndex36, depth36 := position, tokenIndex, depth
						if !_rules[ruleLogOr]() {
							goto l37
						}
						goto l36
					l37:
						position, tokenIndex, depth = position36, tokenIndex36, depth36
						if !_rules[ruleLogAnd]() {
							goto l35
						}
					}
				l36:
					goto l34
				l35:
					position, tokenIndex, depth = position35, tokenIndex35, depth35
				}
				depth--
				add(ruleLevel4, position33)
			}
			return true
		l32:
			position, tokenIndex, depth = position32, tokenIndex32, depth32
			return false
		},
		/* 11 LogOr <- <('-' 'o' 'r' req_ws Level3)> */
		func() bool {
			position38, tokenIndex38, depth38 := position, tokenIndex, depth
			{
				position39 := position
				depth++
				if buffer[position] != rune('-') {
					goto l38
				}
				position++
				if buffer[position] != rune('o') {
					goto l38
				}
				position++
				if buffer[position] != rune('r') {
					goto l38
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l38
				}
				if !_rules[ruleLevel3]() {
					goto l38
				}
				depth--
				add(ruleLogOr, position39)
			}
			return true
		l38:
			position, tokenIndex, depth = position38, tokenIndex38, depth38
			return false
		},
		/* 12 LogAnd <- <('-' 'a' 'n' 'd' req_ws Level3)> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				if buffer[position] != rune('-') {
					goto l40
				}
				position++
				if buffer[position] != rune('a') {
					goto l40
				}
				position++
				if buffer[position] != rune('n') {
					goto l40
				}
				position++
				if buffer[position] != rune('d') {
					goto l40
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l40
				}
				if !_rules[ruleLevel3]() {
					goto l40
				}
				depth--
				add(ruleLogAnd, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 13 Level3 <- <(Level2 (req_ws Comparison)*)> */
		func() bool {
			position42, tokenIndex42, depth42 := position, tokenIndex, depth
			{
				position43 := position
				depth++
				if !_rules[ruleLevel2]() {
					goto l42
				}
			l44:
				{
					position45, tokenIndex45, depth45 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l45
					}
					if !_rules[ruleComparison]() {
						goto l45
					}
					goto l44
				l45:
					position, tokenIndex, depth = position45, tokenIndex45, depth45
				}
				depth--
				add(ruleLevel3, position43)
			}
			return true
		l42:
			position, tokenIndex, depth = position42, tokenIndex42, depth42
			return false
		},
		/* 14 Comparison <- <(CompareOp req_ws Level2)> */
		func() bool {
			position46, tokenIndex46, depth46 := position, tokenIndex, depth
			{
				position47 := position
				depth++
				if !_rules[ruleCompareOp]() {
					goto l46
				}
				if !_rules[rulereq_ws]() {
					goto l46
				}
				if !_rules[ruleLevel2]() {
					goto l46
				}
				depth--
				add(ruleComparison, position47)
			}
			return true
		l46:
			position, tokenIndex, depth = position46, tokenIndex46, depth46
			return false
		},
		/* 15 CompareOp <- <(('=' '=') / ('!' '=') / ('<' '=') / ('>' '=') / '>' / '<' / '>')> */
		func() bool {
			position48, tokenIndex48, depth48 := position, tokenIndex, depth
			{
				position49 := position
				depth++
				{
					position50, tokenIndex50, depth50 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l51
					}
					position++
					if buffer[position] != rune('=') {
						goto l51
					}
					position++
					goto l50
				l51:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if buffer[position] != rune('!') {
						goto l52
					}
					position++
					if buffer[position] != rune('=') {
						goto l52
					}
					position++
					goto l50
				l52:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if buffer[position] != rune('<') {
						goto l53
					}
					position++
					if buffer[position] != rune('=') {
						goto l53
					}
					position++
					goto l50
				l53:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if buffer[position] != rune('>') {
						goto l54
					}
					position++
					if buffer[position] != rune('=') {
						goto l54
					}
					position++
					goto l50
				l54:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if buffer[position] != rune('>') {
						goto l55
					}
					position++
					goto l50
				l55:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if buffer[position] != rune('<') {
						goto l56
					}
					position++
					goto l50
				l56:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if buffer[position] != rune('>') {
						goto l48
					}
					position++
				}
			l50:
				depth--
				add(ruleCompareOp, position49)
			}
			return true
		l48:
			position, tokenIndex, depth = position48, tokenIndex48, depth48
			return false
		},
		/* 16 Level2 <- <(Level1 (req_ws (Addition / Subtraction))*)> */
		func() bool {
			position57, tokenIndex57, depth57 := position, tokenIndex, depth
			{
				position58 := position
				depth++
				if !_rules[ruleLevel1]() {
					goto l57
				}
			l59:
				{
					position60, tokenIndex60, depth60 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l60
					}
					{
						position61, tokenIndex61, depth61 := position, tokenIndex, depth
						if !_rules[ruleAddition]() {
							goto l62
						}
						goto l61
					l62:
						position, tokenIndex, depth = position61, tokenIndex61, depth61
						if !_rules[ruleSubtraction]() {
							goto l60
						}
					}
				l61:
					goto l59
				l60:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
				}
				depth--
				add(ruleLevel2, position58)
			}
			return true
		l57:
			position, tokenIndex, depth = position57, tokenIndex57, depth57
			return false
		},
		/* 17 Addition <- <('+' req_ws Level1)> */
		func() bool {
			position63, tokenIndex63, depth63 := position, tokenIndex, depth
			{
				position64 := position
				depth++
				if buffer[position] != rune('+') {
					goto l63
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l63
				}
				if !_rules[ruleLevel1]() {
					goto l63
				}
				depth--
				add(ruleAddition, position64)
			}
			return true
		l63:
			position, tokenIndex, depth = position63, tokenIndex63, depth63
			return false
		},
		/* 18 Subtraction <- <('-' req_ws Level1)> */
		func() bool {
			position65, tokenIndex65, depth65 := position, tokenIndex, depth
			{
				position66 := position
				depth++
				if buffer[position] != rune('-') {
					goto l65
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l65
				}
				if !_rules[ruleLevel1]() {
					goto l65
				}
				depth--
				add(ruleSubtraction, position66)
			}
			return true
		l65:
			position, tokenIndex, depth = position65, tokenIndex65, depth65
			return false
		},
		/* 19 Level1 <- <(Level0 (req_ws (Multiplication / Division / Modulo))*)> */
		func() bool {
			position67, tokenIndex67, depth67 := position, tokenIndex, depth
			{
				position68 := position
				depth++
				if !_rules[ruleLevel0]() {
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
						if !_rules[ruleMultiplication]() {
							goto l72
						}
						goto l71
					l72:
						position, tokenIndex, depth = position71, tokenIndex71, depth71
						if !_rules[ruleDivision]() {
							goto l73
						}
						goto l71
					l73:
						position, tokenIndex, depth = position71, tokenIndex71, depth71
						if !_rules[ruleModulo]() {
							goto l70
						}
					}
				l71:
					goto l69
				l70:
					position, tokenIndex, depth = position70, tokenIndex70, depth70
				}
				depth--
				add(ruleLevel1, position68)
			}
			return true
		l67:
			position, tokenIndex, depth = position67, tokenIndex67, depth67
			return false
		},
		/* 20 Multiplication <- <('*' req_ws Level0)> */
		func() bool {
			position74, tokenIndex74, depth74 := position, tokenIndex, depth
			{
				position75 := position
				depth++
				if buffer[position] != rune('*') {
					goto l74
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l74
				}
				if !_rules[ruleLevel0]() {
					goto l74
				}
				depth--
				add(ruleMultiplication, position75)
			}
			return true
		l74:
			position, tokenIndex, depth = position74, tokenIndex74, depth74
			return false
		},
		/* 21 Division <- <('/' req_ws Level0)> */
		func() bool {
			position76, tokenIndex76, depth76 := position, tokenIndex, depth
			{
				position77 := position
				depth++
				if buffer[position] != rune('/') {
					goto l76
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l76
				}
				if !_rules[ruleLevel0]() {
					goto l76
				}
				depth--
				add(ruleDivision, position77)
			}
			return true
		l76:
			position, tokenIndex, depth = position76, tokenIndex76, depth76
			return false
		},
		/* 22 Modulo <- <('%' req_ws Level0)> */
		func() bool {
			position78, tokenIndex78, depth78 := position, tokenIndex, depth
			{
				position79 := position
				depth++
				if buffer[position] != rune('%') {
					goto l78
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l78
				}
				if !_rules[ruleLevel0]() {
					goto l78
				}
				depth--
				add(ruleModulo, position79)
			}
			return true
		l78:
			position, tokenIndex, depth = position78, tokenIndex78, depth78
			return false
		},
		/* 23 Level0 <- <(String / Integer / Boolean / EmptyHash / Nil / Not / Substitution / Merge / Auto / Lambda / Chained)> */
		func() bool {
			position80, tokenIndex80, depth80 := position, tokenIndex, depth
			{
				position81 := position
				depth++
				{
					position82, tokenIndex82, depth82 := position, tokenIndex, depth
					if !_rules[ruleString]() {
						goto l83
					}
					goto l82
				l83:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleInteger]() {
						goto l84
					}
					goto l82
				l84:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleBoolean]() {
						goto l85
					}
					goto l82
				l85:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleEmptyHash]() {
						goto l86
					}
					goto l82
				l86:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleNil]() {
						goto l87
					}
					goto l82
				l87:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleNot]() {
						goto l88
					}
					goto l82
				l88:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleSubstitution]() {
						goto l89
					}
					goto l82
				l89:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleMerge]() {
						goto l90
					}
					goto l82
				l90:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleAuto]() {
						goto l91
					}
					goto l82
				l91:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleLambda]() {
						goto l92
					}
					goto l82
				l92:
					position, tokenIndex, depth = position82, tokenIndex82, depth82
					if !_rules[ruleChained]() {
						goto l80
					}
				}
			l82:
				depth--
				add(ruleLevel0, position81)
			}
			return true
		l80:
			position, tokenIndex, depth = position80, tokenIndex80, depth80
			return false
		},
		/* 24 Chained <- <((Mapping / Sum / List / Range / ((Grouped / Reference) ChainedCall*)) (ChainedQualifiedExpression ChainedCall+)* ChainedQualifiedExpression?)> */
		func() bool {
			position93, tokenIndex93, depth93 := position, tokenIndex, depth
			{
				position94 := position
				depth++
				{
					position95, tokenIndex95, depth95 := position, tokenIndex, depth
					if !_rules[ruleMapping]() {
						goto l96
					}
					goto l95
				l96:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if !_rules[ruleSum]() {
						goto l97
					}
					goto l95
				l97:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if !_rules[ruleList]() {
						goto l98
					}
					goto l95
				l98:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if !_rules[ruleRange]() {
						goto l99
					}
					goto l95
				l99:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					{
						position100, tokenIndex100, depth100 := position, tokenIndex, depth
						if !_rules[ruleGrouped]() {
							goto l101
						}
						goto l100
					l101:
						position, tokenIndex, depth = position100, tokenIndex100, depth100
						if !_rules[ruleReference]() {
							goto l93
						}
					}
				l100:
				l102:
					{
						position103, tokenIndex103, depth103 := position, tokenIndex, depth
						if !_rules[ruleChainedCall]() {
							goto l103
						}
						goto l102
					l103:
						position, tokenIndex, depth = position103, tokenIndex103, depth103
					}
				}
			l95:
			l104:
				{
					position105, tokenIndex105, depth105 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l105
					}
					if !_rules[ruleChainedCall]() {
						goto l105
					}
				l106:
					{
						position107, tokenIndex107, depth107 := position, tokenIndex, depth
						if !_rules[ruleChainedCall]() {
							goto l107
						}
						goto l106
					l107:
						position, tokenIndex, depth = position107, tokenIndex107, depth107
					}
					goto l104
				l105:
					position, tokenIndex, depth = position105, tokenIndex105, depth105
				}
				{
					position108, tokenIndex108, depth108 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l108
					}
					goto l109
				l108:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
				}
			l109:
				depth--
				add(ruleChained, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 25 ChainedQualifiedExpression <- <('.' FollowUpRef)> */
		func() bool {
			position110, tokenIndex110, depth110 := position, tokenIndex, depth
			{
				position111 := position
				depth++
				if buffer[position] != rune('.') {
					goto l110
				}
				position++
				if !_rules[ruleFollowUpRef]() {
					goto l110
				}
				depth--
				add(ruleChainedQualifiedExpression, position111)
			}
			return true
		l110:
			position, tokenIndex, depth = position110, tokenIndex110, depth110
			return false
		},
		/* 26 ChainedCall <- <('(' Arguments ')')> */
		func() bool {
			position112, tokenIndex112, depth112 := position, tokenIndex, depth
			{
				position113 := position
				depth++
				if buffer[position] != rune('(') {
					goto l112
				}
				position++
				if !_rules[ruleArguments]() {
					goto l112
				}
				if buffer[position] != rune(')') {
					goto l112
				}
				position++
				depth--
				add(ruleChainedCall, position113)
			}
			return true
		l112:
			position, tokenIndex, depth = position112, tokenIndex112, depth112
			return false
		},
		/* 27 Arguments <- <(Expression NextExpression*)> */
		func() bool {
			position114, tokenIndex114, depth114 := position, tokenIndex, depth
			{
				position115 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l114
				}
			l116:
				{
					position117, tokenIndex117, depth117 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l117
					}
					goto l116
				l117:
					position, tokenIndex, depth = position117, tokenIndex117, depth117
				}
				depth--
				add(ruleArguments, position115)
			}
			return true
		l114:
			position, tokenIndex, depth = position114, tokenIndex114, depth114
			return false
		},
		/* 28 NextExpression <- <(',' Expression)> */
		func() bool {
			position118, tokenIndex118, depth118 := position, tokenIndex, depth
			{
				position119 := position
				depth++
				if buffer[position] != rune(',') {
					goto l118
				}
				position++
				if !_rules[ruleExpression]() {
					goto l118
				}
				depth--
				add(ruleNextExpression, position119)
			}
			return true
		l118:
			position, tokenIndex, depth = position118, tokenIndex118, depth118
			return false
		},
		/* 29 Substitution <- <('*' Level0)> */
		func() bool {
			position120, tokenIndex120, depth120 := position, tokenIndex, depth
			{
				position121 := position
				depth++
				if buffer[position] != rune('*') {
					goto l120
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l120
				}
				depth--
				add(ruleSubstitution, position121)
			}
			return true
		l120:
			position, tokenIndex, depth = position120, tokenIndex120, depth120
			return false
		},
		/* 30 Not <- <('!' ws Level0)> */
		func() bool {
			position122, tokenIndex122, depth122 := position, tokenIndex, depth
			{
				position123 := position
				depth++
				if buffer[position] != rune('!') {
					goto l122
				}
				position++
				if !_rules[rulews]() {
					goto l122
				}
				if !_rules[ruleLevel0]() {
					goto l122
				}
				depth--
				add(ruleNot, position123)
			}
			return true
		l122:
			position, tokenIndex, depth = position122, tokenIndex122, depth122
			return false
		},
		/* 31 Grouped <- <('(' Expression ')')> */
		func() bool {
			position124, tokenIndex124, depth124 := position, tokenIndex, depth
			{
				position125 := position
				depth++
				if buffer[position] != rune('(') {
					goto l124
				}
				position++
				if !_rules[ruleExpression]() {
					goto l124
				}
				if buffer[position] != rune(')') {
					goto l124
				}
				position++
				depth--
				add(ruleGrouped, position125)
			}
			return true
		l124:
			position, tokenIndex, depth = position124, tokenIndex124, depth124
			return false
		},
		/* 32 Range <- <('[' Expression ('.' '.') Expression ']')> */
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
				if buffer[position] != rune('.') {
					goto l126
				}
				position++
				if buffer[position] != rune('.') {
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
				add(ruleRange, position127)
			}
			return true
		l126:
			position, tokenIndex, depth = position126, tokenIndex126, depth126
			return false
		},
		/* 33 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position128, tokenIndex128, depth128 := position, tokenIndex, depth
			{
				position129 := position
				depth++
				{
					position130, tokenIndex130, depth130 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l130
					}
					position++
					goto l131
				l130:
					position, tokenIndex, depth = position130, tokenIndex130, depth130
				}
			l131:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l128
				}
				position++
			l132:
				{
					position133, tokenIndex133, depth133 := position, tokenIndex, depth
					{
						position134, tokenIndex134, depth134 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l135
						}
						position++
						goto l134
					l135:
						position, tokenIndex, depth = position134, tokenIndex134, depth134
						if buffer[position] != rune('_') {
							goto l133
						}
						position++
					}
				l134:
					goto l132
				l133:
					position, tokenIndex, depth = position133, tokenIndex133, depth133
				}
				depth--
				add(ruleInteger, position129)
			}
			return true
		l128:
			position, tokenIndex, depth = position128, tokenIndex128, depth128
			return false
		},
		/* 34 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position136, tokenIndex136, depth136 := position, tokenIndex, depth
			{
				position137 := position
				depth++
				if buffer[position] != rune('"') {
					goto l136
				}
				position++
			l138:
				{
					position139, tokenIndex139, depth139 := position, tokenIndex, depth
					{
						position140, tokenIndex140, depth140 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l141
						}
						position++
						if buffer[position] != rune('"') {
							goto l141
						}
						position++
						goto l140
					l141:
						position, tokenIndex, depth = position140, tokenIndex140, depth140
						{
							position142, tokenIndex142, depth142 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l142
							}
							position++
							goto l139
						l142:
							position, tokenIndex, depth = position142, tokenIndex142, depth142
						}
						if !matchDot() {
							goto l139
						}
					}
				l140:
					goto l138
				l139:
					position, tokenIndex, depth = position139, tokenIndex139, depth139
				}
				if buffer[position] != rune('"') {
					goto l136
				}
				position++
				depth--
				add(ruleString, position137)
			}
			return true
		l136:
			position, tokenIndex, depth = position136, tokenIndex136, depth136
			return false
		},
		/* 35 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position143, tokenIndex143, depth143 := position, tokenIndex, depth
			{
				position144 := position
				depth++
				{
					position145, tokenIndex145, depth145 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l146
					}
					position++
					if buffer[position] != rune('r') {
						goto l146
					}
					position++
					if buffer[position] != rune('u') {
						goto l146
					}
					position++
					if buffer[position] != rune('e') {
						goto l146
					}
					position++
					goto l145
				l146:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if buffer[position] != rune('f') {
						goto l143
					}
					position++
					if buffer[position] != rune('a') {
						goto l143
					}
					position++
					if buffer[position] != rune('l') {
						goto l143
					}
					position++
					if buffer[position] != rune('s') {
						goto l143
					}
					position++
					if buffer[position] != rune('e') {
						goto l143
					}
					position++
				}
			l145:
				depth--
				add(ruleBoolean, position144)
			}
			return true
		l143:
			position, tokenIndex, depth = position143, tokenIndex143, depth143
			return false
		},
		/* 36 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				{
					position149, tokenIndex149, depth149 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l150
					}
					position++
					if buffer[position] != rune('i') {
						goto l150
					}
					position++
					if buffer[position] != rune('l') {
						goto l150
					}
					position++
					goto l149
				l150:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
					if buffer[position] != rune('~') {
						goto l147
					}
					position++
				}
			l149:
				depth--
				add(ruleNil, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 37 EmptyHash <- <('{' '}')> */
		func() bool {
			position151, tokenIndex151, depth151 := position, tokenIndex, depth
			{
				position152 := position
				depth++
				if buffer[position] != rune('{') {
					goto l151
				}
				position++
				if buffer[position] != rune('}') {
					goto l151
				}
				position++
				depth--
				add(ruleEmptyHash, position152)
			}
			return true
		l151:
			position, tokenIndex, depth = position151, tokenIndex151, depth151
			return false
		},
		/* 38 List <- <('[' Contents? ']')> */
		func() bool {
			position153, tokenIndex153, depth153 := position, tokenIndex, depth
			{
				position154 := position
				depth++
				if buffer[position] != rune('[') {
					goto l153
				}
				position++
				{
					position155, tokenIndex155, depth155 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l155
					}
					goto l156
				l155:
					position, tokenIndex, depth = position155, tokenIndex155, depth155
				}
			l156:
				if buffer[position] != rune(']') {
					goto l153
				}
				position++
				depth--
				add(ruleList, position154)
			}
			return true
		l153:
			position, tokenIndex, depth = position153, tokenIndex153, depth153
			return false
		},
		/* 39 Contents <- <(Expression NextExpression*)> */
		func() bool {
			position157, tokenIndex157, depth157 := position, tokenIndex, depth
			{
				position158 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l157
				}
			l159:
				{
					position160, tokenIndex160, depth160 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l160
					}
					goto l159
				l160:
					position, tokenIndex, depth = position160, tokenIndex160, depth160
				}
				depth--
				add(ruleContents, position158)
			}
			return true
		l157:
			position, tokenIndex, depth = position157, tokenIndex157, depth157
			return false
		},
		/* 40 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position161, tokenIndex161, depth161 := position, tokenIndex, depth
			{
				position162 := position
				depth++
				{
					position163, tokenIndex163, depth163 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l164
					}
					goto l163
				l164:
					position, tokenIndex, depth = position163, tokenIndex163, depth163
					if !_rules[ruleSimpleMerge]() {
						goto l161
					}
				}
			l163:
				depth--
				add(ruleMerge, position162)
			}
			return true
		l161:
			position, tokenIndex, depth = position161, tokenIndex161, depth161
			return false
		},
		/* 41 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position165, tokenIndex165, depth165 := position, tokenIndex, depth
			{
				position166 := position
				depth++
				if buffer[position] != rune('m') {
					goto l165
				}
				position++
				if buffer[position] != rune('e') {
					goto l165
				}
				position++
				if buffer[position] != rune('r') {
					goto l165
				}
				position++
				if buffer[position] != rune('g') {
					goto l165
				}
				position++
				if buffer[position] != rune('e') {
					goto l165
				}
				position++
				{
					position167, tokenIndex167, depth167 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l167
					}
					if !_rules[ruleRequired]() {
						goto l167
					}
					goto l165
				l167:
					position, tokenIndex, depth = position167, tokenIndex167, depth167
				}
				{
					position168, tokenIndex168, depth168 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l168
					}
					{
						position170, tokenIndex170, depth170 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l171
						}
						goto l170
					l171:
						position, tokenIndex, depth = position170, tokenIndex170, depth170
						if !_rules[ruleOn]() {
							goto l168
						}
					}
				l170:
					goto l169
				l168:
					position, tokenIndex, depth = position168, tokenIndex168, depth168
				}
			l169:
				if !_rules[rulereq_ws]() {
					goto l165
				}
				if !_rules[ruleReference]() {
					goto l165
				}
				depth--
				add(ruleRefMerge, position166)
			}
			return true
		l165:
			position, tokenIndex, depth = position165, tokenIndex165, depth165
			return false
		},
		/* 42 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position172, tokenIndex172, depth172 := position, tokenIndex, depth
			{
				position173 := position
				depth++
				if buffer[position] != rune('m') {
					goto l172
				}
				position++
				if buffer[position] != rune('e') {
					goto l172
				}
				position++
				if buffer[position] != rune('r') {
					goto l172
				}
				position++
				if buffer[position] != rune('g') {
					goto l172
				}
				position++
				if buffer[position] != rune('e') {
					goto l172
				}
				position++
				{
					position174, tokenIndex174, depth174 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l174
					}
					{
						position176, tokenIndex176, depth176 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l177
						}
						goto l176
					l177:
						position, tokenIndex, depth = position176, tokenIndex176, depth176
						if !_rules[ruleRequired]() {
							goto l178
						}
						goto l176
					l178:
						position, tokenIndex, depth = position176, tokenIndex176, depth176
						if !_rules[ruleOn]() {
							goto l174
						}
					}
				l176:
					goto l175
				l174:
					position, tokenIndex, depth = position174, tokenIndex174, depth174
				}
			l175:
				depth--
				add(ruleSimpleMerge, position173)
			}
			return true
		l172:
			position, tokenIndex, depth = position172, tokenIndex172, depth172
			return false
		},
		/* 43 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position179, tokenIndex179, depth179 := position, tokenIndex, depth
			{
				position180 := position
				depth++
				if buffer[position] != rune('r') {
					goto l179
				}
				position++
				if buffer[position] != rune('e') {
					goto l179
				}
				position++
				if buffer[position] != rune('p') {
					goto l179
				}
				position++
				if buffer[position] != rune('l') {
					goto l179
				}
				position++
				if buffer[position] != rune('a') {
					goto l179
				}
				position++
				if buffer[position] != rune('c') {
					goto l179
				}
				position++
				if buffer[position] != rune('e') {
					goto l179
				}
				position++
				depth--
				add(ruleReplace, position180)
			}
			return true
		l179:
			position, tokenIndex, depth = position179, tokenIndex179, depth179
			return false
		},
		/* 44 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position181, tokenIndex181, depth181 := position, tokenIndex, depth
			{
				position182 := position
				depth++
				if buffer[position] != rune('r') {
					goto l181
				}
				position++
				if buffer[position] != rune('e') {
					goto l181
				}
				position++
				if buffer[position] != rune('q') {
					goto l181
				}
				position++
				if buffer[position] != rune('u') {
					goto l181
				}
				position++
				if buffer[position] != rune('i') {
					goto l181
				}
				position++
				if buffer[position] != rune('r') {
					goto l181
				}
				position++
				if buffer[position] != rune('e') {
					goto l181
				}
				position++
				if buffer[position] != rune('d') {
					goto l181
				}
				position++
				depth--
				add(ruleRequired, position182)
			}
			return true
		l181:
			position, tokenIndex, depth = position181, tokenIndex181, depth181
			return false
		},
		/* 45 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position183, tokenIndex183, depth183 := position, tokenIndex, depth
			{
				position184 := position
				depth++
				if buffer[position] != rune('o') {
					goto l183
				}
				position++
				if buffer[position] != rune('n') {
					goto l183
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l183
				}
				if !_rules[ruleName]() {
					goto l183
				}
				depth--
				add(ruleOn, position184)
			}
			return true
		l183:
			position, tokenIndex, depth = position183, tokenIndex183, depth183
			return false
		},
		/* 46 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position185, tokenIndex185, depth185 := position, tokenIndex, depth
			{
				position186 := position
				depth++
				if buffer[position] != rune('a') {
					goto l185
				}
				position++
				if buffer[position] != rune('u') {
					goto l185
				}
				position++
				if buffer[position] != rune('t') {
					goto l185
				}
				position++
				if buffer[position] != rune('o') {
					goto l185
				}
				position++
				depth--
				add(ruleAuto, position186)
			}
			return true
		l185:
			position, tokenIndex, depth = position185, tokenIndex185, depth185
			return false
		},
		/* 47 Mapping <- <('m' 'a' 'p' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position187, tokenIndex187, depth187 := position, tokenIndex, depth
			{
				position188 := position
				depth++
				if buffer[position] != rune('m') {
					goto l187
				}
				position++
				if buffer[position] != rune('a') {
					goto l187
				}
				position++
				if buffer[position] != rune('p') {
					goto l187
				}
				position++
				if buffer[position] != rune('[') {
					goto l187
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l187
				}
				{
					position189, tokenIndex189, depth189 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l190
					}
					goto l189
				l190:
					position, tokenIndex, depth = position189, tokenIndex189, depth189
					if buffer[position] != rune('|') {
						goto l187
					}
					position++
					if !_rules[ruleExpression]() {
						goto l187
					}
				}
			l189:
				if buffer[position] != rune(']') {
					goto l187
				}
				position++
				depth--
				add(ruleMapping, position188)
			}
			return true
		l187:
			position, tokenIndex, depth = position187, tokenIndex187, depth187
			return false
		},
		/* 48 Sum <- <('s' 'u' 'm' '[' Level7 '|' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position191, tokenIndex191, depth191 := position, tokenIndex, depth
			{
				position192 := position
				depth++
				if buffer[position] != rune('s') {
					goto l191
				}
				position++
				if buffer[position] != rune('u') {
					goto l191
				}
				position++
				if buffer[position] != rune('m') {
					goto l191
				}
				position++
				if buffer[position] != rune('[') {
					goto l191
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l191
				}
				if buffer[position] != rune('|') {
					goto l191
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l191
				}
				{
					position193, tokenIndex193, depth193 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l194
					}
					goto l193
				l194:
					position, tokenIndex, depth = position193, tokenIndex193, depth193
					if buffer[position] != rune('|') {
						goto l191
					}
					position++
					if !_rules[ruleExpression]() {
						goto l191
					}
				}
			l193:
				if buffer[position] != rune(']') {
					goto l191
				}
				position++
				depth--
				add(ruleSum, position192)
			}
			return true
		l191:
			position, tokenIndex, depth = position191, tokenIndex191, depth191
			return false
		},
		/* 49 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position195, tokenIndex195, depth195 := position, tokenIndex, depth
			{
				position196 := position
				depth++
				if buffer[position] != rune('l') {
					goto l195
				}
				position++
				if buffer[position] != rune('a') {
					goto l195
				}
				position++
				if buffer[position] != rune('m') {
					goto l195
				}
				position++
				if buffer[position] != rune('b') {
					goto l195
				}
				position++
				if buffer[position] != rune('d') {
					goto l195
				}
				position++
				if buffer[position] != rune('a') {
					goto l195
				}
				position++
				{
					position197, tokenIndex197, depth197 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l198
					}
					goto l197
				l198:
					position, tokenIndex, depth = position197, tokenIndex197, depth197
					if !_rules[ruleLambdaExpr]() {
						goto l195
					}
				}
			l197:
				depth--
				add(ruleLambda, position196)
			}
			return true
		l195:
			position, tokenIndex, depth = position195, tokenIndex195, depth195
			return false
		},
		/* 50 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position199, tokenIndex199, depth199 := position, tokenIndex, depth
			{
				position200 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l199
				}
				if !_rules[ruleExpression]() {
					goto l199
				}
				depth--
				add(ruleLambdaRef, position200)
			}
			return true
		l199:
			position, tokenIndex, depth = position199, tokenIndex199, depth199
			return false
		},
		/* 51 LambdaExpr <- <(ws '|' ws Name NextName* ws '|' ws ('-' '>') Expression)> */
		func() bool {
			position201, tokenIndex201, depth201 := position, tokenIndex, depth
			{
				position202 := position
				depth++
				if !_rules[rulews]() {
					goto l201
				}
				if buffer[position] != rune('|') {
					goto l201
				}
				position++
				if !_rules[rulews]() {
					goto l201
				}
				if !_rules[ruleName]() {
					goto l201
				}
			l203:
				{
					position204, tokenIndex204, depth204 := position, tokenIndex, depth
					if !_rules[ruleNextName]() {
						goto l204
					}
					goto l203
				l204:
					position, tokenIndex, depth = position204, tokenIndex204, depth204
				}
				if !_rules[rulews]() {
					goto l201
				}
				if buffer[position] != rune('|') {
					goto l201
				}
				position++
				if !_rules[rulews]() {
					goto l201
				}
				if buffer[position] != rune('-') {
					goto l201
				}
				position++
				if buffer[position] != rune('>') {
					goto l201
				}
				position++
				if !_rules[ruleExpression]() {
					goto l201
				}
				depth--
				add(ruleLambdaExpr, position202)
			}
			return true
		l201:
			position, tokenIndex, depth = position201, tokenIndex201, depth201
			return false
		},
		/* 52 NextName <- <(ws ',' ws Name)> */
		func() bool {
			position205, tokenIndex205, depth205 := position, tokenIndex, depth
			{
				position206 := position
				depth++
				if !_rules[rulews]() {
					goto l205
				}
				if buffer[position] != rune(',') {
					goto l205
				}
				position++
				if !_rules[rulews]() {
					goto l205
				}
				if !_rules[ruleName]() {
					goto l205
				}
				depth--
				add(ruleNextName, position206)
			}
			return true
		l205:
			position, tokenIndex, depth = position205, tokenIndex205, depth205
			return false
		},
		/* 53 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position207, tokenIndex207, depth207 := position, tokenIndex, depth
			{
				position208 := position
				depth++
				{
					position211, tokenIndex211, depth211 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l212
					}
					position++
					goto l211
				l212:
					position, tokenIndex, depth = position211, tokenIndex211, depth211
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l213
					}
					position++
					goto l211
				l213:
					position, tokenIndex, depth = position211, tokenIndex211, depth211
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l214
					}
					position++
					goto l211
				l214:
					position, tokenIndex, depth = position211, tokenIndex211, depth211
					if buffer[position] != rune('_') {
						goto l207
					}
					position++
				}
			l211:
			l209:
				{
					position210, tokenIndex210, depth210 := position, tokenIndex, depth
					{
						position215, tokenIndex215, depth215 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l216
						}
						position++
						goto l215
					l216:
						position, tokenIndex, depth = position215, tokenIndex215, depth215
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l217
						}
						position++
						goto l215
					l217:
						position, tokenIndex, depth = position215, tokenIndex215, depth215
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l218
						}
						position++
						goto l215
					l218:
						position, tokenIndex, depth = position215, tokenIndex215, depth215
						if buffer[position] != rune('_') {
							goto l210
						}
						position++
					}
				l215:
					goto l209
				l210:
					position, tokenIndex, depth = position210, tokenIndex210, depth210
				}
				depth--
				add(ruleName, position208)
			}
			return true
		l207:
			position, tokenIndex, depth = position207, tokenIndex207, depth207
			return false
		},
		/* 54 Reference <- <('.'? Key ('.' (Key / Index))*)> */
		func() bool {
			position219, tokenIndex219, depth219 := position, tokenIndex, depth
			{
				position220 := position
				depth++
				{
					position221, tokenIndex221, depth221 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l221
					}
					position++
					goto l222
				l221:
					position, tokenIndex, depth = position221, tokenIndex221, depth221
				}
			l222:
				if !_rules[ruleKey]() {
					goto l219
				}
			l223:
				{
					position224, tokenIndex224, depth224 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l224
					}
					position++
					{
						position225, tokenIndex225, depth225 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l226
						}
						goto l225
					l226:
						position, tokenIndex, depth = position225, tokenIndex225, depth225
						if !_rules[ruleIndex]() {
							goto l224
						}
					}
				l225:
					goto l223
				l224:
					position, tokenIndex, depth = position224, tokenIndex224, depth224
				}
				depth--
				add(ruleReference, position220)
			}
			return true
		l219:
			position, tokenIndex, depth = position219, tokenIndex219, depth219
			return false
		},
		/* 55 FollowUpRef <- <((Key / Index) ('.' (Key / Index))*)> */
		func() bool {
			position227, tokenIndex227, depth227 := position, tokenIndex, depth
			{
				position228 := position
				depth++
				{
					position229, tokenIndex229, depth229 := position, tokenIndex, depth
					if !_rules[ruleKey]() {
						goto l230
					}
					goto l229
				l230:
					position, tokenIndex, depth = position229, tokenIndex229, depth229
					if !_rules[ruleIndex]() {
						goto l227
					}
				}
			l229:
			l231:
				{
					position232, tokenIndex232, depth232 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l232
					}
					position++
					{
						position233, tokenIndex233, depth233 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l234
						}
						goto l233
					l234:
						position, tokenIndex, depth = position233, tokenIndex233, depth233
						if !_rules[ruleIndex]() {
							goto l232
						}
					}
				l233:
					goto l231
				l232:
					position, tokenIndex, depth = position232, tokenIndex232, depth232
				}
				depth--
				add(ruleFollowUpRef, position228)
			}
			return true
		l227:
			position, tokenIndex, depth = position227, tokenIndex227, depth227
			return false
		},
		/* 56 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position235, tokenIndex235, depth235 := position, tokenIndex, depth
			{
				position236 := position
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
						goto l235
					}
					position++
				}
			l237:
			l241:
				{
					position242, tokenIndex242, depth242 := position, tokenIndex, depth
					{
						position243, tokenIndex243, depth243 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l244
						}
						position++
						goto l243
					l244:
						position, tokenIndex, depth = position243, tokenIndex243, depth243
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l245
						}
						position++
						goto l243
					l245:
						position, tokenIndex, depth = position243, tokenIndex243, depth243
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l246
						}
						position++
						goto l243
					l246:
						position, tokenIndex, depth = position243, tokenIndex243, depth243
						if buffer[position] != rune('_') {
							goto l247
						}
						position++
						goto l243
					l247:
						position, tokenIndex, depth = position243, tokenIndex243, depth243
						if buffer[position] != rune('-') {
							goto l242
						}
						position++
					}
				l243:
					goto l241
				l242:
					position, tokenIndex, depth = position242, tokenIndex242, depth242
				}
				{
					position248, tokenIndex248, depth248 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l248
					}
					position++
					{
						position250, tokenIndex250, depth250 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l251
						}
						position++
						goto l250
					l251:
						position, tokenIndex, depth = position250, tokenIndex250, depth250
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l252
						}
						position++
						goto l250
					l252:
						position, tokenIndex, depth = position250, tokenIndex250, depth250
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l253
						}
						position++
						goto l250
					l253:
						position, tokenIndex, depth = position250, tokenIndex250, depth250
						if buffer[position] != rune('_') {
							goto l248
						}
						position++
					}
				l250:
				l254:
					{
						position255, tokenIndex255, depth255 := position, tokenIndex, depth
						{
							position256, tokenIndex256, depth256 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l257
							}
							position++
							goto l256
						l257:
							position, tokenIndex, depth = position256, tokenIndex256, depth256
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l258
							}
							position++
							goto l256
						l258:
							position, tokenIndex, depth = position256, tokenIndex256, depth256
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l259
							}
							position++
							goto l256
						l259:
							position, tokenIndex, depth = position256, tokenIndex256, depth256
							if buffer[position] != rune('_') {
								goto l260
							}
							position++
							goto l256
						l260:
							position, tokenIndex, depth = position256, tokenIndex256, depth256
							if buffer[position] != rune('-') {
								goto l255
							}
							position++
						}
					l256:
						goto l254
					l255:
						position, tokenIndex, depth = position255, tokenIndex255, depth255
					}
					goto l249
				l248:
					position, tokenIndex, depth = position248, tokenIndex248, depth248
				}
			l249:
				depth--
				add(ruleKey, position236)
			}
			return true
		l235:
			position, tokenIndex, depth = position235, tokenIndex235, depth235
			return false
		},
		/* 57 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position261, tokenIndex261, depth261 := position, tokenIndex, depth
			{
				position262 := position
				depth++
				if buffer[position] != rune('[') {
					goto l261
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l261
				}
				position++
			l263:
				{
					position264, tokenIndex264, depth264 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l264
					}
					position++
					goto l263
				l264:
					position, tokenIndex, depth = position264, tokenIndex264, depth264
				}
				if buffer[position] != rune(']') {
					goto l261
				}
				position++
				depth--
				add(ruleIndex, position262)
			}
			return true
		l261:
			position, tokenIndex, depth = position261, tokenIndex261, depth261
			return false
		},
		/* 58 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position266 := position
				depth++
			l267:
				{
					position268, tokenIndex268, depth268 := position, tokenIndex, depth
					{
						position269, tokenIndex269, depth269 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l270
						}
						position++
						goto l269
					l270:
						position, tokenIndex, depth = position269, tokenIndex269, depth269
						if buffer[position] != rune('\t') {
							goto l271
						}
						position++
						goto l269
					l271:
						position, tokenIndex, depth = position269, tokenIndex269, depth269
						if buffer[position] != rune('\n') {
							goto l272
						}
						position++
						goto l269
					l272:
						position, tokenIndex, depth = position269, tokenIndex269, depth269
						if buffer[position] != rune('\r') {
							goto l268
						}
						position++
					}
				l269:
					goto l267
				l268:
					position, tokenIndex, depth = position268, tokenIndex268, depth268
				}
				depth--
				add(rulews, position266)
			}
			return true
		},
		/* 59 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position273, tokenIndex273, depth273 := position, tokenIndex, depth
			{
				position274 := position
				depth++
				{
					position277, tokenIndex277, depth277 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l278
					}
					position++
					goto l277
				l278:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
					if buffer[position] != rune('\t') {
						goto l279
					}
					position++
					goto l277
				l279:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
					if buffer[position] != rune('\n') {
						goto l280
					}
					position++
					goto l277
				l280:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
					if buffer[position] != rune('\r') {
						goto l273
					}
					position++
				}
			l277:
			l275:
				{
					position276, tokenIndex276, depth276 := position, tokenIndex, depth
					{
						position281, tokenIndex281, depth281 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l282
						}
						position++
						goto l281
					l282:
						position, tokenIndex, depth = position281, tokenIndex281, depth281
						if buffer[position] != rune('\t') {
							goto l283
						}
						position++
						goto l281
					l283:
						position, tokenIndex, depth = position281, tokenIndex281, depth281
						if buffer[position] != rune('\n') {
							goto l284
						}
						position++
						goto l281
					l284:
						position, tokenIndex, depth = position281, tokenIndex281, depth281
						if buffer[position] != rune('\r') {
							goto l276
						}
						position++
					}
				l281:
					goto l275
				l276:
					position, tokenIndex, depth = position276, tokenIndex276, depth276
				}
				depth--
				add(rulereq_ws, position274)
			}
			return true
		l273:
			position, tokenIndex, depth = position273, tokenIndex273, depth273
			return false
		},
	}
	p.rules = _rules
}
