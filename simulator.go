package main

import (
	"fmt"
	"io"
	"strings"
)

// RunMachine executes the machine on the given input and writes a trace to w.
// As a bonus, it also reports the time complexity as the number of executed steps.
func RunMachine(m *Machine, input string, w io.Writer) error {
	if m.transIndex == nil {
		m.buildIndexes()
	}

	m.Describe(w)

	// Tape represented as a map of index -> symbol, plus min/max visited
	tape := make(map[int]string)
	for i, r := range input {
		tape[i] = string(r)
	}

	head := 0
	state := m.Initial

	const maxSteps = 1_000_000
	steps := 0

	for step := 0; step < maxSteps; step++ {
		symbol, ok := tape[head]
		if !ok {
			symbol = m.Blank
		}

		// Check for final state
		if _, isFinal := m.finalsSet[state]; isFinal {
			printTapeState(w, tape, head, state, symbol, nil, m.Blank)
			fmt.Fprintf(w, "Time complexity (steps): %d\n", steps)
			return nil
		}

		// Lookup transition
		stateTrans, ok := m.transIndex[state]
		if !ok {
			printTapeState(w, tape, head, state, symbol, nil, m.Blank)
			fmt.Fprintf(w, "Time complexity (steps): %d\n", steps)
			return fmt.Errorf("blocked: no transitions defined for state %q", state)
		}
		t, ok := stateTrans[symbol]
		if !ok {
			printTapeState(w, tape, head, state, symbol, nil, m.Blank)
			fmt.Fprintf(w, "Time complexity (steps): %d\n", steps)
			return fmt.Errorf("blocked: no transition for state %q and symbol %q", state, symbol)
		}

		// Print step trace
		printTapeState(w, tape, head, state, symbol, &t, m.Blank)

		// Count this transition for time complexity
		steps++

		// Apply transition: write symbol, move head, change state
		tape[head] = t.Write

		switch t.Action {
		case ActionLeft:
			head--
		case ActionRight:
			head++
		}
		state = t.ToState
	}

	fmt.Fprintf(w, "Time complexity (steps): %d\n", steps)
	return fmt.Errorf("blocked: maximum number of steps (%d) exceeded", maxSteps)
}

func tapeBounds(tape map[int]string, blank string) (int, int) {
	if len(tape) == 0 {
		return 0, 0
	}
	min := 0
	max := 0
	first := true
	for i, s := range tape {
		if s == blank {
			continue
		}
		if first || i < min {
			min = i
		}
		if first || i > max {
			max = i
		}
		first = false
	}
	if first {
		// all blanks
		return 0, 0
	}
	return min, max
}

func printTapeState(w io.Writer, tape map[int]string, head int, state string, read string, t *Transition, blank string) {
	// Find bounds including head
	min, max := tapeBounds(tape, blank)
	if head < min {
		min = head
	}
	if head > max {
		max = head
	}

	var b strings.Builder
	b.WriteString("[")
	for i := min; i <= max+3; i++ {
		sym, ok := tape[i]
		if !ok {
			sym = blank
		}
		if i == head {
			b.WriteString("<")
			b.WriteString(sym)
			b.WriteString(">")
		} else {
			b.WriteString(sym)
		}
	}
	b.WriteString("...] ")

	if t != nil {
		fmt.Fprintf(w, "%s(%s, %s) -> (%s, %s, %s)\n",
			b.String(), state, read, t.ToState, t.Write, t.Action)
	} else {
		fmt.Fprintf(w, "%s(%s, %s)\n", b.String(), state, read)
	}
}


