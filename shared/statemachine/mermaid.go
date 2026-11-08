package statemachine

import (
	"fmt"
	"strings"
)

// Mermaid renders the machine as a Mermaid stateDiagram-v2. The output is
// deterministic: states are numbered in order of first appearance in the
// declaration and edges follow the declaration order. States and events are
// rendered with fmt.Sprint and escaped, so a hostile label cannot add lines,
// edges or quotes to the diagram.
func (d *Definition[S, E, C]) Mermaid() string {
	var b strings.Builder
	b.WriteString("stateDiagram-v2\n")
	ids := make(map[S]int)
	var order []S
	id := func(s S) int {
		n, ok := ids[s]
		if !ok {
			n = len(order)
			ids[s] = n
			order = append(order, s)
		}
		return n
	}
	for _, t := range d.ts {
		id(t.From)
		id(t.To)
	}
	for n, s := range order {
		fmt.Fprintf(&b, "    state \"%s\" as s%d\n", escapeLabel(fmt.Sprint(s)), n)
	}
	for _, t := range d.ts {
		fmt.Fprintf(&b, "    s%d --> s%d : %s\n", ids[t.From], ids[t.To], escapeLabel(fmt.Sprint(t.Event)))
	}
	return b.String()
}

var labelEscapes = map[rune]string{
	'#': "#35;", '"': "#quot;", ':': "#58;", ';': "#59;",
	'<': "#lt;", '>': "#gt;", '{': "#123;", '}': "#125;",
}

func isLineBreakOrControl(r rune) bool {
	return r < 0x20 || r == 0x7f || r == '\u2028' || r == '\u2029' || r == '\u0085'
}

// escapeLabel makes a label safe inside a quoted Mermaid string or after the
// edge colon. Characters with meaning to the parser become Mermaid entity
// codes; line breaks and other control characters become a space.
func escapeLabel(s string) string {
	var b strings.Builder
	for _, r := range s {
		if esc, ok := labelEscapes[r]; ok {
			b.WriteString(esc)
		} else if isLineBreakOrControl(r) {
			b.WriteByte(' ')
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
