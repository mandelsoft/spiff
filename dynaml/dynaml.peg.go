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
	rules  [60]func() bool
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
		/* 24 Chained <- <((Mapping / List / Range / ((Grouped / Reference) ChainedCall*)) (ChainedQualifiedExpression ChainedCall+)* ChainedQualifiedExpression?)> */
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
					if !_rules[ruleList]() {
						goto l97
					}
					goto l95
				l97:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					if !_rules[ruleRange]() {
						goto l98
					}
					goto l95
				l98:
					position, tokenIndex, depth = position95, tokenIndex95, depth95
					{
						position99, tokenIndex99, depth99 := position, tokenIndex, depth
						if !_rules[ruleGrouped]() {
							goto l100
						}
						goto l99
					l100:
						position, tokenIndex, depth = position99, tokenIndex99, depth99
						if !_rules[ruleReference]() {
							goto l93
						}
					}
				l99:
				l101:
					{
						position102, tokenIndex102, depth102 := position, tokenIndex, depth
						if !_rules[ruleChainedCall]() {
							goto l102
						}
						goto l101
					l102:
						position, tokenIndex, depth = position102, tokenIndex102, depth102
					}
				}
			l95:
			l103:
				{
					position104, tokenIndex104, depth104 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l104
					}
					if !_rules[ruleChainedCall]() {
						goto l104
					}
				l105:
					{
						position106, tokenIndex106, depth106 := position, tokenIndex, depth
						if !_rules[ruleChainedCall]() {
							goto l106
						}
						goto l105
					l106:
						position, tokenIndex, depth = position106, tokenIndex106, depth106
					}
					goto l103
				l104:
					position, tokenIndex, depth = position104, tokenIndex104, depth104
				}
				{
					position107, tokenIndex107, depth107 := position, tokenIndex, depth
					if !_rules[ruleChainedQualifiedExpression]() {
						goto l107
					}
					goto l108
				l107:
					position, tokenIndex, depth = position107, tokenIndex107, depth107
				}
			l108:
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
			position109, tokenIndex109, depth109 := position, tokenIndex, depth
			{
				position110 := position
				depth++
				if buffer[position] != rune('.') {
					goto l109
				}
				position++
				if !_rules[ruleFollowUpRef]() {
					goto l109
				}
				depth--
				add(ruleChainedQualifiedExpression, position110)
			}
			return true
		l109:
			position, tokenIndex, depth = position109, tokenIndex109, depth109
			return false
		},
		/* 26 ChainedCall <- <('(' Arguments ')')> */
		func() bool {
			position111, tokenIndex111, depth111 := position, tokenIndex, depth
			{
				position112 := position
				depth++
				if buffer[position] != rune('(') {
					goto l111
				}
				position++
				if !_rules[ruleArguments]() {
					goto l111
				}
				if buffer[position] != rune(')') {
					goto l111
				}
				position++
				depth--
				add(ruleChainedCall, position112)
			}
			return true
		l111:
			position, tokenIndex, depth = position111, tokenIndex111, depth111
			return false
		},
		/* 27 Arguments <- <(Expression NextExpression*)> */
		func() bool {
			position113, tokenIndex113, depth113 := position, tokenIndex, depth
			{
				position114 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l113
				}
			l115:
				{
					position116, tokenIndex116, depth116 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l116
					}
					goto l115
				l116:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
				}
				depth--
				add(ruleArguments, position114)
			}
			return true
		l113:
			position, tokenIndex, depth = position113, tokenIndex113, depth113
			return false
		},
		/* 28 NextExpression <- <(',' Expression)> */
		func() bool {
			position117, tokenIndex117, depth117 := position, tokenIndex, depth
			{
				position118 := position
				depth++
				if buffer[position] != rune(',') {
					goto l117
				}
				position++
				if !_rules[ruleExpression]() {
					goto l117
				}
				depth--
				add(ruleNextExpression, position118)
			}
			return true
		l117:
			position, tokenIndex, depth = position117, tokenIndex117, depth117
			return false
		},
		/* 29 Substitution <- <('*' Level0)> */
		func() bool {
			position119, tokenIndex119, depth119 := position, tokenIndex, depth
			{
				position120 := position
				depth++
				if buffer[position] != rune('*') {
					goto l119
				}
				position++
				if !_rules[ruleLevel0]() {
					goto l119
				}
				depth--
				add(ruleSubstitution, position120)
			}
			return true
		l119:
			position, tokenIndex, depth = position119, tokenIndex119, depth119
			return false
		},
		/* 30 Not <- <('!' ws Level0)> */
		func() bool {
			position121, tokenIndex121, depth121 := position, tokenIndex, depth
			{
				position122 := position
				depth++
				if buffer[position] != rune('!') {
					goto l121
				}
				position++
				if !_rules[rulews]() {
					goto l121
				}
				if !_rules[ruleLevel0]() {
					goto l121
				}
				depth--
				add(ruleNot, position122)
			}
			return true
		l121:
			position, tokenIndex, depth = position121, tokenIndex121, depth121
			return false
		},
		/* 31 Grouped <- <('(' Expression ')')> */
		func() bool {
			position123, tokenIndex123, depth123 := position, tokenIndex, depth
			{
				position124 := position
				depth++
				if buffer[position] != rune('(') {
					goto l123
				}
				position++
				if !_rules[ruleExpression]() {
					goto l123
				}
				if buffer[position] != rune(')') {
					goto l123
				}
				position++
				depth--
				add(ruleGrouped, position124)
			}
			return true
		l123:
			position, tokenIndex, depth = position123, tokenIndex123, depth123
			return false
		},
		/* 32 Range <- <('[' Expression ('.' '.') Expression ']')> */
		func() bool {
			position125, tokenIndex125, depth125 := position, tokenIndex, depth
			{
				position126 := position
				depth++
				if buffer[position] != rune('[') {
					goto l125
				}
				position++
				if !_rules[ruleExpression]() {
					goto l125
				}
				if buffer[position] != rune('.') {
					goto l125
				}
				position++
				if buffer[position] != rune('.') {
					goto l125
				}
				position++
				if !_rules[ruleExpression]() {
					goto l125
				}
				if buffer[position] != rune(']') {
					goto l125
				}
				position++
				depth--
				add(ruleRange, position126)
			}
			return true
		l125:
			position, tokenIndex, depth = position125, tokenIndex125, depth125
			return false
		},
		/* 33 Integer <- <('-'? [0-9] ([0-9] / '_')*)> */
		func() bool {
			position127, tokenIndex127, depth127 := position, tokenIndex, depth
			{
				position128 := position
				depth++
				{
					position129, tokenIndex129, depth129 := position, tokenIndex, depth
					if buffer[position] != rune('-') {
						goto l129
					}
					position++
					goto l130
				l129:
					position, tokenIndex, depth = position129, tokenIndex129, depth129
				}
			l130:
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l127
				}
				position++
			l131:
				{
					position132, tokenIndex132, depth132 := position, tokenIndex, depth
					{
						position133, tokenIndex133, depth133 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l134
						}
						position++
						goto l133
					l134:
						position, tokenIndex, depth = position133, tokenIndex133, depth133
						if buffer[position] != rune('_') {
							goto l132
						}
						position++
					}
				l133:
					goto l131
				l132:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
				}
				depth--
				add(ruleInteger, position128)
			}
			return true
		l127:
			position, tokenIndex, depth = position127, tokenIndex127, depth127
			return false
		},
		/* 34 String <- <('"' (('\\' '"') / (!'"' .))* '"')> */
		func() bool {
			position135, tokenIndex135, depth135 := position, tokenIndex, depth
			{
				position136 := position
				depth++
				if buffer[position] != rune('"') {
					goto l135
				}
				position++
			l137:
				{
					position138, tokenIndex138, depth138 := position, tokenIndex, depth
					{
						position139, tokenIndex139, depth139 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l140
						}
						position++
						if buffer[position] != rune('"') {
							goto l140
						}
						position++
						goto l139
					l140:
						position, tokenIndex, depth = position139, tokenIndex139, depth139
						{
							position141, tokenIndex141, depth141 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l141
							}
							position++
							goto l138
						l141:
							position, tokenIndex, depth = position141, tokenIndex141, depth141
						}
						if !matchDot() {
							goto l138
						}
					}
				l139:
					goto l137
				l138:
					position, tokenIndex, depth = position138, tokenIndex138, depth138
				}
				if buffer[position] != rune('"') {
					goto l135
				}
				position++
				depth--
				add(ruleString, position136)
			}
			return true
		l135:
			position, tokenIndex, depth = position135, tokenIndex135, depth135
			return false
		},
		/* 35 Boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position142, tokenIndex142, depth142 := position, tokenIndex, depth
			{
				position143 := position
				depth++
				{
					position144, tokenIndex144, depth144 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l145
					}
					position++
					if buffer[position] != rune('r') {
						goto l145
					}
					position++
					if buffer[position] != rune('u') {
						goto l145
					}
					position++
					if buffer[position] != rune('e') {
						goto l145
					}
					position++
					goto l144
				l145:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('f') {
						goto l142
					}
					position++
					if buffer[position] != rune('a') {
						goto l142
					}
					position++
					if buffer[position] != rune('l') {
						goto l142
					}
					position++
					if buffer[position] != rune('s') {
						goto l142
					}
					position++
					if buffer[position] != rune('e') {
						goto l142
					}
					position++
				}
			l144:
				depth--
				add(ruleBoolean, position143)
			}
			return true
		l142:
			position, tokenIndex, depth = position142, tokenIndex142, depth142
			return false
		},
		/* 36 Nil <- <(('n' 'i' 'l') / '~')> */
		func() bool {
			position146, tokenIndex146, depth146 := position, tokenIndex, depth
			{
				position147 := position
				depth++
				{
					position148, tokenIndex148, depth148 := position, tokenIndex, depth
					if buffer[position] != rune('n') {
						goto l149
					}
					position++
					if buffer[position] != rune('i') {
						goto l149
					}
					position++
					if buffer[position] != rune('l') {
						goto l149
					}
					position++
					goto l148
				l149:
					position, tokenIndex, depth = position148, tokenIndex148, depth148
					if buffer[position] != rune('~') {
						goto l146
					}
					position++
				}
			l148:
				depth--
				add(ruleNil, position147)
			}
			return true
		l146:
			position, tokenIndex, depth = position146, tokenIndex146, depth146
			return false
		},
		/* 37 EmptyHash <- <('{' '}')> */
		func() bool {
			position150, tokenIndex150, depth150 := position, tokenIndex, depth
			{
				position151 := position
				depth++
				if buffer[position] != rune('{') {
					goto l150
				}
				position++
				if buffer[position] != rune('}') {
					goto l150
				}
				position++
				depth--
				add(ruleEmptyHash, position151)
			}
			return true
		l150:
			position, tokenIndex, depth = position150, tokenIndex150, depth150
			return false
		},
		/* 38 List <- <('[' Contents? ']')> */
		func() bool {
			position152, tokenIndex152, depth152 := position, tokenIndex, depth
			{
				position153 := position
				depth++
				if buffer[position] != rune('[') {
					goto l152
				}
				position++
				{
					position154, tokenIndex154, depth154 := position, tokenIndex, depth
					if !_rules[ruleContents]() {
						goto l154
					}
					goto l155
				l154:
					position, tokenIndex, depth = position154, tokenIndex154, depth154
				}
			l155:
				if buffer[position] != rune(']') {
					goto l152
				}
				position++
				depth--
				add(ruleList, position153)
			}
			return true
		l152:
			position, tokenIndex, depth = position152, tokenIndex152, depth152
			return false
		},
		/* 39 Contents <- <(Expression NextExpression*)> */
		func() bool {
			position156, tokenIndex156, depth156 := position, tokenIndex, depth
			{
				position157 := position
				depth++
				if !_rules[ruleExpression]() {
					goto l156
				}
			l158:
				{
					position159, tokenIndex159, depth159 := position, tokenIndex, depth
					if !_rules[ruleNextExpression]() {
						goto l159
					}
					goto l158
				l159:
					position, tokenIndex, depth = position159, tokenIndex159, depth159
				}
				depth--
				add(ruleContents, position157)
			}
			return true
		l156:
			position, tokenIndex, depth = position156, tokenIndex156, depth156
			return false
		},
		/* 40 Merge <- <(RefMerge / SimpleMerge)> */
		func() bool {
			position160, tokenIndex160, depth160 := position, tokenIndex, depth
			{
				position161 := position
				depth++
				{
					position162, tokenIndex162, depth162 := position, tokenIndex, depth
					if !_rules[ruleRefMerge]() {
						goto l163
					}
					goto l162
				l163:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
					if !_rules[ruleSimpleMerge]() {
						goto l160
					}
				}
			l162:
				depth--
				add(ruleMerge, position161)
			}
			return true
		l160:
			position, tokenIndex, depth = position160, tokenIndex160, depth160
			return false
		},
		/* 41 RefMerge <- <('m' 'e' 'r' 'g' 'e' !(req_ws Required) (req_ws (Replace / On))? req_ws Reference)> */
		func() bool {
			position164, tokenIndex164, depth164 := position, tokenIndex, depth
			{
				position165 := position
				depth++
				if buffer[position] != rune('m') {
					goto l164
				}
				position++
				if buffer[position] != rune('e') {
					goto l164
				}
				position++
				if buffer[position] != rune('r') {
					goto l164
				}
				position++
				if buffer[position] != rune('g') {
					goto l164
				}
				position++
				if buffer[position] != rune('e') {
					goto l164
				}
				position++
				{
					position166, tokenIndex166, depth166 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l166
					}
					if !_rules[ruleRequired]() {
						goto l166
					}
					goto l164
				l166:
					position, tokenIndex, depth = position166, tokenIndex166, depth166
				}
				{
					position167, tokenIndex167, depth167 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l167
					}
					{
						position169, tokenIndex169, depth169 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l170
						}
						goto l169
					l170:
						position, tokenIndex, depth = position169, tokenIndex169, depth169
						if !_rules[ruleOn]() {
							goto l167
						}
					}
				l169:
					goto l168
				l167:
					position, tokenIndex, depth = position167, tokenIndex167, depth167
				}
			l168:
				if !_rules[rulereq_ws]() {
					goto l164
				}
				if !_rules[ruleReference]() {
					goto l164
				}
				depth--
				add(ruleRefMerge, position165)
			}
			return true
		l164:
			position, tokenIndex, depth = position164, tokenIndex164, depth164
			return false
		},
		/* 42 SimpleMerge <- <('m' 'e' 'r' 'g' 'e' (req_ws (Replace / Required / On))?)> */
		func() bool {
			position171, tokenIndex171, depth171 := position, tokenIndex, depth
			{
				position172 := position
				depth++
				if buffer[position] != rune('m') {
					goto l171
				}
				position++
				if buffer[position] != rune('e') {
					goto l171
				}
				position++
				if buffer[position] != rune('r') {
					goto l171
				}
				position++
				if buffer[position] != rune('g') {
					goto l171
				}
				position++
				if buffer[position] != rune('e') {
					goto l171
				}
				position++
				{
					position173, tokenIndex173, depth173 := position, tokenIndex, depth
					if !_rules[rulereq_ws]() {
						goto l173
					}
					{
						position175, tokenIndex175, depth175 := position, tokenIndex, depth
						if !_rules[ruleReplace]() {
							goto l176
						}
						goto l175
					l176:
						position, tokenIndex, depth = position175, tokenIndex175, depth175
						if !_rules[ruleRequired]() {
							goto l177
						}
						goto l175
					l177:
						position, tokenIndex, depth = position175, tokenIndex175, depth175
						if !_rules[ruleOn]() {
							goto l173
						}
					}
				l175:
					goto l174
				l173:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
				}
			l174:
				depth--
				add(ruleSimpleMerge, position172)
			}
			return true
		l171:
			position, tokenIndex, depth = position171, tokenIndex171, depth171
			return false
		},
		/* 43 Replace <- <('r' 'e' 'p' 'l' 'a' 'c' 'e')> */
		func() bool {
			position178, tokenIndex178, depth178 := position, tokenIndex, depth
			{
				position179 := position
				depth++
				if buffer[position] != rune('r') {
					goto l178
				}
				position++
				if buffer[position] != rune('e') {
					goto l178
				}
				position++
				if buffer[position] != rune('p') {
					goto l178
				}
				position++
				if buffer[position] != rune('l') {
					goto l178
				}
				position++
				if buffer[position] != rune('a') {
					goto l178
				}
				position++
				if buffer[position] != rune('c') {
					goto l178
				}
				position++
				if buffer[position] != rune('e') {
					goto l178
				}
				position++
				depth--
				add(ruleReplace, position179)
			}
			return true
		l178:
			position, tokenIndex, depth = position178, tokenIndex178, depth178
			return false
		},
		/* 44 Required <- <('r' 'e' 'q' 'u' 'i' 'r' 'e' 'd')> */
		func() bool {
			position180, tokenIndex180, depth180 := position, tokenIndex, depth
			{
				position181 := position
				depth++
				if buffer[position] != rune('r') {
					goto l180
				}
				position++
				if buffer[position] != rune('e') {
					goto l180
				}
				position++
				if buffer[position] != rune('q') {
					goto l180
				}
				position++
				if buffer[position] != rune('u') {
					goto l180
				}
				position++
				if buffer[position] != rune('i') {
					goto l180
				}
				position++
				if buffer[position] != rune('r') {
					goto l180
				}
				position++
				if buffer[position] != rune('e') {
					goto l180
				}
				position++
				if buffer[position] != rune('d') {
					goto l180
				}
				position++
				depth--
				add(ruleRequired, position181)
			}
			return true
		l180:
			position, tokenIndex, depth = position180, tokenIndex180, depth180
			return false
		},
		/* 45 On <- <('o' 'n' req_ws Name)> */
		func() bool {
			position182, tokenIndex182, depth182 := position, tokenIndex, depth
			{
				position183 := position
				depth++
				if buffer[position] != rune('o') {
					goto l182
				}
				position++
				if buffer[position] != rune('n') {
					goto l182
				}
				position++
				if !_rules[rulereq_ws]() {
					goto l182
				}
				if !_rules[ruleName]() {
					goto l182
				}
				depth--
				add(ruleOn, position183)
			}
			return true
		l182:
			position, tokenIndex, depth = position182, tokenIndex182, depth182
			return false
		},
		/* 46 Auto <- <('a' 'u' 't' 'o')> */
		func() bool {
			position184, tokenIndex184, depth184 := position, tokenIndex, depth
			{
				position185 := position
				depth++
				if buffer[position] != rune('a') {
					goto l184
				}
				position++
				if buffer[position] != rune('u') {
					goto l184
				}
				position++
				if buffer[position] != rune('t') {
					goto l184
				}
				position++
				if buffer[position] != rune('o') {
					goto l184
				}
				position++
				depth--
				add(ruleAuto, position185)
			}
			return true
		l184:
			position, tokenIndex, depth = position184, tokenIndex184, depth184
			return false
		},
		/* 47 Mapping <- <('m' 'a' 'p' '[' Level7 (LambdaExpr / ('|' Expression)) ']')> */
		func() bool {
			position186, tokenIndex186, depth186 := position, tokenIndex, depth
			{
				position187 := position
				depth++
				if buffer[position] != rune('m') {
					goto l186
				}
				position++
				if buffer[position] != rune('a') {
					goto l186
				}
				position++
				if buffer[position] != rune('p') {
					goto l186
				}
				position++
				if buffer[position] != rune('[') {
					goto l186
				}
				position++
				if !_rules[ruleLevel7]() {
					goto l186
				}
				{
					position188, tokenIndex188, depth188 := position, tokenIndex, depth
					if !_rules[ruleLambdaExpr]() {
						goto l189
					}
					goto l188
				l189:
					position, tokenIndex, depth = position188, tokenIndex188, depth188
					if buffer[position] != rune('|') {
						goto l186
					}
					position++
					if !_rules[ruleExpression]() {
						goto l186
					}
				}
			l188:
				if buffer[position] != rune(']') {
					goto l186
				}
				position++
				depth--
				add(ruleMapping, position187)
			}
			return true
		l186:
			position, tokenIndex, depth = position186, tokenIndex186, depth186
			return false
		},
		/* 48 Lambda <- <('l' 'a' 'm' 'b' 'd' 'a' (LambdaRef / LambdaExpr))> */
		func() bool {
			position190, tokenIndex190, depth190 := position, tokenIndex, depth
			{
				position191 := position
				depth++
				if buffer[position] != rune('l') {
					goto l190
				}
				position++
				if buffer[position] != rune('a') {
					goto l190
				}
				position++
				if buffer[position] != rune('m') {
					goto l190
				}
				position++
				if buffer[position] != rune('b') {
					goto l190
				}
				position++
				if buffer[position] != rune('d') {
					goto l190
				}
				position++
				if buffer[position] != rune('a') {
					goto l190
				}
				position++
				{
					position192, tokenIndex192, depth192 := position, tokenIndex, depth
					if !_rules[ruleLambdaRef]() {
						goto l193
					}
					goto l192
				l193:
					position, tokenIndex, depth = position192, tokenIndex192, depth192
					if !_rules[ruleLambdaExpr]() {
						goto l190
					}
				}
			l192:
				depth--
				add(ruleLambda, position191)
			}
			return true
		l190:
			position, tokenIndex, depth = position190, tokenIndex190, depth190
			return false
		},
		/* 49 LambdaRef <- <(req_ws Expression)> */
		func() bool {
			position194, tokenIndex194, depth194 := position, tokenIndex, depth
			{
				position195 := position
				depth++
				if !_rules[rulereq_ws]() {
					goto l194
				}
				if !_rules[ruleExpression]() {
					goto l194
				}
				depth--
				add(ruleLambdaRef, position195)
			}
			return true
		l194:
			position, tokenIndex, depth = position194, tokenIndex194, depth194
			return false
		},
		/* 50 LambdaExpr <- <(ws '|' ws Name NextName* ws '|' ws ('-' '>') Expression)> */
		func() bool {
			position196, tokenIndex196, depth196 := position, tokenIndex, depth
			{
				position197 := position
				depth++
				if !_rules[rulews]() {
					goto l196
				}
				if buffer[position] != rune('|') {
					goto l196
				}
				position++
				if !_rules[rulews]() {
					goto l196
				}
				if !_rules[ruleName]() {
					goto l196
				}
			l198:
				{
					position199, tokenIndex199, depth199 := position, tokenIndex, depth
					if !_rules[ruleNextName]() {
						goto l199
					}
					goto l198
				l199:
					position, tokenIndex, depth = position199, tokenIndex199, depth199
				}
				if !_rules[rulews]() {
					goto l196
				}
				if buffer[position] != rune('|') {
					goto l196
				}
				position++
				if !_rules[rulews]() {
					goto l196
				}
				if buffer[position] != rune('-') {
					goto l196
				}
				position++
				if buffer[position] != rune('>') {
					goto l196
				}
				position++
				if !_rules[ruleExpression]() {
					goto l196
				}
				depth--
				add(ruleLambdaExpr, position197)
			}
			return true
		l196:
			position, tokenIndex, depth = position196, tokenIndex196, depth196
			return false
		},
		/* 51 NextName <- <(ws ',' ws Name)> */
		func() bool {
			position200, tokenIndex200, depth200 := position, tokenIndex, depth
			{
				position201 := position
				depth++
				if !_rules[rulews]() {
					goto l200
				}
				if buffer[position] != rune(',') {
					goto l200
				}
				position++
				if !_rules[rulews]() {
					goto l200
				}
				if !_rules[ruleName]() {
					goto l200
				}
				depth--
				add(ruleNextName, position201)
			}
			return true
		l200:
			position, tokenIndex, depth = position200, tokenIndex200, depth200
			return false
		},
		/* 52 Name <- <([a-z] / [A-Z] / [0-9] / '_')+> */
		func() bool {
			position202, tokenIndex202, depth202 := position, tokenIndex, depth
			{
				position203 := position
				depth++
				{
					position206, tokenIndex206, depth206 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l207
					}
					position++
					goto l206
				l207:
					position, tokenIndex, depth = position206, tokenIndex206, depth206
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l208
					}
					position++
					goto l206
				l208:
					position, tokenIndex, depth = position206, tokenIndex206, depth206
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l209
					}
					position++
					goto l206
				l209:
					position, tokenIndex, depth = position206, tokenIndex206, depth206
					if buffer[position] != rune('_') {
						goto l202
					}
					position++
				}
			l206:
			l204:
				{
					position205, tokenIndex205, depth205 := position, tokenIndex, depth
					{
						position210, tokenIndex210, depth210 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l211
						}
						position++
						goto l210
					l211:
						position, tokenIndex, depth = position210, tokenIndex210, depth210
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l212
						}
						position++
						goto l210
					l212:
						position, tokenIndex, depth = position210, tokenIndex210, depth210
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l213
						}
						position++
						goto l210
					l213:
						position, tokenIndex, depth = position210, tokenIndex210, depth210
						if buffer[position] != rune('_') {
							goto l205
						}
						position++
					}
				l210:
					goto l204
				l205:
					position, tokenIndex, depth = position205, tokenIndex205, depth205
				}
				depth--
				add(ruleName, position203)
			}
			return true
		l202:
			position, tokenIndex, depth = position202, tokenIndex202, depth202
			return false
		},
		/* 53 Reference <- <('.'? Key ('.' (Key / Index))*)> */
		func() bool {
			position214, tokenIndex214, depth214 := position, tokenIndex, depth
			{
				position215 := position
				depth++
				{
					position216, tokenIndex216, depth216 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l216
					}
					position++
					goto l217
				l216:
					position, tokenIndex, depth = position216, tokenIndex216, depth216
				}
			l217:
				if !_rules[ruleKey]() {
					goto l214
				}
			l218:
				{
					position219, tokenIndex219, depth219 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l219
					}
					position++
					{
						position220, tokenIndex220, depth220 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l221
						}
						goto l220
					l221:
						position, tokenIndex, depth = position220, tokenIndex220, depth220
						if !_rules[ruleIndex]() {
							goto l219
						}
					}
				l220:
					goto l218
				l219:
					position, tokenIndex, depth = position219, tokenIndex219, depth219
				}
				depth--
				add(ruleReference, position215)
			}
			return true
		l214:
			position, tokenIndex, depth = position214, tokenIndex214, depth214
			return false
		},
		/* 54 FollowUpRef <- <((Key / Index) ('.' (Key / Index))*)> */
		func() bool {
			position222, tokenIndex222, depth222 := position, tokenIndex, depth
			{
				position223 := position
				depth++
				{
					position224, tokenIndex224, depth224 := position, tokenIndex, depth
					if !_rules[ruleKey]() {
						goto l225
					}
					goto l224
				l225:
					position, tokenIndex, depth = position224, tokenIndex224, depth224
					if !_rules[ruleIndex]() {
						goto l222
					}
				}
			l224:
			l226:
				{
					position227, tokenIndex227, depth227 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l227
					}
					position++
					{
						position228, tokenIndex228, depth228 := position, tokenIndex, depth
						if !_rules[ruleKey]() {
							goto l229
						}
						goto l228
					l229:
						position, tokenIndex, depth = position228, tokenIndex228, depth228
						if !_rules[ruleIndex]() {
							goto l227
						}
					}
				l228:
					goto l226
				l227:
					position, tokenIndex, depth = position227, tokenIndex227, depth227
				}
				depth--
				add(ruleFollowUpRef, position223)
			}
			return true
		l222:
			position, tokenIndex, depth = position222, tokenIndex222, depth222
			return false
		},
		/* 55 Key <- <(([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')* (':' ([a-z] / [A-Z] / [0-9] / '_') ([a-z] / [A-Z] / [0-9] / '_' / '-')*)?)> */
		func() bool {
			position230, tokenIndex230, depth230 := position, tokenIndex, depth
			{
				position231 := position
				depth++
				{
					position232, tokenIndex232, depth232 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l233
					}
					position++
					goto l232
				l233:
					position, tokenIndex, depth = position232, tokenIndex232, depth232
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l234
					}
					position++
					goto l232
				l234:
					position, tokenIndex, depth = position232, tokenIndex232, depth232
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l235
					}
					position++
					goto l232
				l235:
					position, tokenIndex, depth = position232, tokenIndex232, depth232
					if buffer[position] != rune('_') {
						goto l230
					}
					position++
				}
			l232:
			l236:
				{
					position237, tokenIndex237, depth237 := position, tokenIndex, depth
					{
						position238, tokenIndex238, depth238 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l239
						}
						position++
						goto l238
					l239:
						position, tokenIndex, depth = position238, tokenIndex238, depth238
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l240
						}
						position++
						goto l238
					l240:
						position, tokenIndex, depth = position238, tokenIndex238, depth238
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l241
						}
						position++
						goto l238
					l241:
						position, tokenIndex, depth = position238, tokenIndex238, depth238
						if buffer[position] != rune('_') {
							goto l242
						}
						position++
						goto l238
					l242:
						position, tokenIndex, depth = position238, tokenIndex238, depth238
						if buffer[position] != rune('-') {
							goto l237
						}
						position++
					}
				l238:
					goto l236
				l237:
					position, tokenIndex, depth = position237, tokenIndex237, depth237
				}
				{
					position243, tokenIndex243, depth243 := position, tokenIndex, depth
					if buffer[position] != rune(':') {
						goto l243
					}
					position++
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
							goto l243
						}
						position++
					}
				l245:
				l249:
					{
						position250, tokenIndex250, depth250 := position, tokenIndex, depth
						{
							position251, tokenIndex251, depth251 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l252
							}
							position++
							goto l251
						l252:
							position, tokenIndex, depth = position251, tokenIndex251, depth251
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l253
							}
							position++
							goto l251
						l253:
							position, tokenIndex, depth = position251, tokenIndex251, depth251
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l254
							}
							position++
							goto l251
						l254:
							position, tokenIndex, depth = position251, tokenIndex251, depth251
							if buffer[position] != rune('_') {
								goto l255
							}
							position++
							goto l251
						l255:
							position, tokenIndex, depth = position251, tokenIndex251, depth251
							if buffer[position] != rune('-') {
								goto l250
							}
							position++
						}
					l251:
						goto l249
					l250:
						position, tokenIndex, depth = position250, tokenIndex250, depth250
					}
					goto l244
				l243:
					position, tokenIndex, depth = position243, tokenIndex243, depth243
				}
			l244:
				depth--
				add(ruleKey, position231)
			}
			return true
		l230:
			position, tokenIndex, depth = position230, tokenIndex230, depth230
			return false
		},
		/* 56 Index <- <('[' [0-9]+ ']')> */
		func() bool {
			position256, tokenIndex256, depth256 := position, tokenIndex, depth
			{
				position257 := position
				depth++
				if buffer[position] != rune('[') {
					goto l256
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l256
				}
				position++
			l258:
				{
					position259, tokenIndex259, depth259 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l259
					}
					position++
					goto l258
				l259:
					position, tokenIndex, depth = position259, tokenIndex259, depth259
				}
				if buffer[position] != rune(']') {
					goto l256
				}
				position++
				depth--
				add(ruleIndex, position257)
			}
			return true
		l256:
			position, tokenIndex, depth = position256, tokenIndex256, depth256
			return false
		},
		/* 57 ws <- <(' ' / '\t' / '\n' / '\r')*> */
		func() bool {
			{
				position261 := position
				depth++
			l262:
				{
					position263, tokenIndex263, depth263 := position, tokenIndex, depth
					{
						position264, tokenIndex264, depth264 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l265
						}
						position++
						goto l264
					l265:
						position, tokenIndex, depth = position264, tokenIndex264, depth264
						if buffer[position] != rune('\t') {
							goto l266
						}
						position++
						goto l264
					l266:
						position, tokenIndex, depth = position264, tokenIndex264, depth264
						if buffer[position] != rune('\n') {
							goto l267
						}
						position++
						goto l264
					l267:
						position, tokenIndex, depth = position264, tokenIndex264, depth264
						if buffer[position] != rune('\r') {
							goto l263
						}
						position++
					}
				l264:
					goto l262
				l263:
					position, tokenIndex, depth = position263, tokenIndex263, depth263
				}
				depth--
				add(rulews, position261)
			}
			return true
		},
		/* 58 req_ws <- <(' ' / '\t' / '\n' / '\r')+> */
		func() bool {
			position268, tokenIndex268, depth268 := position, tokenIndex, depth
			{
				position269 := position
				depth++
				{
					position272, tokenIndex272, depth272 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l273
					}
					position++
					goto l272
				l273:
					position, tokenIndex, depth = position272, tokenIndex272, depth272
					if buffer[position] != rune('\t') {
						goto l274
					}
					position++
					goto l272
				l274:
					position, tokenIndex, depth = position272, tokenIndex272, depth272
					if buffer[position] != rune('\n') {
						goto l275
					}
					position++
					goto l272
				l275:
					position, tokenIndex, depth = position272, tokenIndex272, depth272
					if buffer[position] != rune('\r') {
						goto l268
					}
					position++
				}
			l272:
			l270:
				{
					position271, tokenIndex271, depth271 := position, tokenIndex, depth
					{
						position276, tokenIndex276, depth276 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l277
						}
						position++
						goto l276
					l277:
						position, tokenIndex, depth = position276, tokenIndex276, depth276
						if buffer[position] != rune('\t') {
							goto l278
						}
						position++
						goto l276
					l278:
						position, tokenIndex, depth = position276, tokenIndex276, depth276
						if buffer[position] != rune('\n') {
							goto l279
						}
						position++
						goto l276
					l279:
						position, tokenIndex, depth = position276, tokenIndex276, depth276
						if buffer[position] != rune('\r') {
							goto l271
						}
						position++
					}
				l276:
					goto l270
				l271:
					position, tokenIndex, depth = position271, tokenIndex271, depth271
				}
				depth--
				add(rulereq_ws, position269)
			}
			return true
		l268:
			position, tokenIndex, depth = position268, tokenIndex268, depth268
			return false
		},
	}
	p.rules = _rules
}
