package main

import (
	"fmt"
	"io"
	"strings"
)

type Tape map[int]string

type Configuration struct {
	Tape  Tape
	Head  int
	State string
	Steps int
}

type StepEvent struct {
	Before     Configuration
	Read       string
	Transition *Transition
}

type SimulationResult struct {
	Events []StepEvent
	Final  Configuration
	Err    error
}

const maxSteps = 1_000_000

// RunMachine executes the machine on the given input and writes a trace to w.
// As a bonus, it also reports the time complexity as the number of executed steps.
func RunMachine(m *Machine, input string, w io.Writer) error {
	machine := *m
	if machine.transIndex == nil || machine.finalsSet == nil || machine.alphaSet == nil {
		machine = machine.withIndexes()
	}

	machine.Describe(w)

	result := simulate(machine, input, maxSteps)
	for _, event := range result.Events {
		writeStepEvent(w, event, machine.Blank)
	}
	fmt.Fprintf(w, "Time complexity (steps): %d\n", result.Final.Steps)
	return result.Err
}

func initialConfiguration(m *Machine, input string) Configuration {
	return Configuration{
		Tape:  buildTape(input),
		Head:  0,
		State: m.Initial,
		Steps: 0,
	}
}

func buildTape(input string) Tape {
	tape := make(Tape, len(input))
	for i, r := range input {
		tape[i] = string(r)
	}
	return tape
}

func simulate(m Machine, input string, max int) SimulationResult {
	conf := initialConfiguration(&m, input)
	events := make([]StepEvent, 0, 128)

	for conf.Steps < max {
		event, nextConf, done, err := nextStep(m, conf)
		events = append(events, event)
		if done {
			return SimulationResult{
				Events: events,
				Final:  conf,
				Err:    nil,
			}
		}
		if err != nil {
			return SimulationResult{
				Events: events,
				Final:  conf,
				Err:    err,
			}
		}
		conf = nextConf
	}

	return SimulationResult{
		Events: events,
		Final:  conf,
		Err:    fmt.Errorf("blocked: maximum number of steps (%d) exceeded", max),
	}
}

func nextStep(m Machine, conf Configuration) (StepEvent, Configuration, bool, error) {
	read := readSymbol(conf.Tape, conf.Head, m.Blank)
	event := StepEvent{
		Before:     conf,
		Read:       read,
		Transition: nil,
	}

	if _, isFinal := m.finalsSet[conf.State]; isFinal {
		return event, conf, true, nil
	}

	stateTrans, ok := m.transIndex[conf.State]
	if !ok {
		return event, conf, false, fmt.Errorf("blocked: no transitions defined for state %q", conf.State)
	}

	t, ok := stateTrans[read]
	if !ok {
		return event, conf, false, fmt.Errorf("blocked: no transition for state %q and symbol %q", conf.State, read)
	}
	event.Transition = &t

	return StepEvent{
			Before:     conf,
			Read:       read,
			Transition: &t,
		},
		applyTransition(conf, t, m.Blank),
		false,
		nil
}

func applyTransition(conf Configuration, t Transition, blank string) Configuration {
	nextTape := tapeWithWrite(conf.Tape, conf.Head, t.Write, blank)
	nextHead := conf.Head
	switch t.Action {
	case ActionLeft:
		nextHead--
	case ActionRight:
		nextHead++
	}

	return Configuration{
		Tape:  nextTape,
		Head:  nextHead,
		State: t.ToState,
		Steps: conf.Steps + 1,
	}
}

func tapeWithWrite(original Tape, position int, symbol string, blank string) Tape {
	next := cloneTape(original)
	if symbol == blank {
		delete(next, position)
		return next
	}
	next[position] = symbol
	return next
}

func cloneTape(original Tape) Tape {
	next := make(Tape, len(original))
	for idx, symbol := range original {
		next[idx] = symbol
	}
	return next
}

func readSymbol(tape Tape, head int, blank string) string {
	symbol, ok := tape[head]
	if !ok {
		return blank
	}
	return symbol
}

func tapeBounds(tape Tape, blank string) (int, int) {
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

func writeStepEvent(w io.Writer, event StepEvent, blank string) {
	fmt.Fprintln(w, formatStepEvent(event, blank))
}

func formatStepEvent(event StepEvent, blank string) string {
	tapeView := renderTape(event.Before.Tape, event.Before.Head, blank)
	if event.Transition == nil {
		return fmt.Sprintf("%s(%s, %s)", tapeView, event.Before.State, event.Read)
	}
	t := event.Transition
	return fmt.Sprintf("%s(%s, %s) -> (%s, %s, %s)",
		tapeView, event.Before.State, event.Read, t.ToState, t.Write, t.Action)
}

func renderTape(tape Tape, head int, blank string) string {
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
	return b.String()
}


