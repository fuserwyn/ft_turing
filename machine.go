package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type Action string

const (
	ActionLeft  Action = "LEFT"
	ActionRight Action = "RIGHT"
)

type Transition struct {
	Read    string `json:"read"`
	ToState string `json:"to_state"`
	Write   string `json:"write"`
	Action  Action `json:"action"`
}

type Machine struct {
	Name        string                         `json:"name"`
	Alphabet    []string                       `json:"alphabet"`
	Blank       string                         `json:"blank"`
	States      []string                       `json:"states"`
	Initial     string                         `json:"initial"`
	Finals      []string                       `json:"finals"`
	Transitions map[string][]Transition        `json:"transitions"`

	// derived fields for faster lookup
	alphaSet   map[string]struct{}            `json:"-"`
	stateSet   map[string]struct{}            `json:"-"`
	finalsSet  map[string]struct{}            `json:"-"`
	transIndex map[string]map[string]Transition `json:"-"`
}

func LoadMachineFromFile(path string) (*Machine, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read json file: %w", err)
	}
	var m Machine
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("invalid json description: %w", err)
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	indexed := m.withIndexes()
	return &indexed, nil
}

func (m Machine) withIndexes() Machine {
	m.alphaSet = toSet(m.Alphabet)
	m.stateSet = toSet(m.States)
	m.finalsSet = toSet(m.Finals)
	m.transIndex = toTransitionIndex(m.Transitions)
	return m
}

func toSet(items []string) map[string]struct{} {
	out := make(map[string]struct{}, len(items))
	for _, item := range items {
		out[item] = struct{}{}
	}
	return out
}

func toTransitionIndex(transitions map[string][]Transition) map[string]map[string]Transition {
	out := make(map[string]map[string]Transition, len(transitions))
	for state, list := range transitions {
		byRead := make(map[string]Transition, len(list))
		for _, t := range list {
			byRead[t.Read] = t
		}
		out[state] = byRead
	}
	return out
}

func (m *Machine) Validate() error {
	if m.Name == "" {
		return errors.New("missing machine name")
	}
	if len(m.Alphabet) == 0 {
		return errors.New("alphabet must not be empty")
	}
	alphaSet := make(map[string]struct{})
	for _, a := range m.Alphabet {
		if len([]rune(a)) != 1 {
			return fmt.Errorf("alphabet symbol %q must be a single character", a)
		}
		if _, exists := alphaSet[a]; exists {
			return fmt.Errorf("duplicate alphabet symbol %q", a)
		}
		alphaSet[a] = struct{}{}
	}
	if _, ok := alphaSet[m.Blank]; !ok {
		return fmt.Errorf("blank symbol %q must be part of alphabet", m.Blank)
	}
	if len(m.States) == 0 {
		return errors.New("states list must not be empty")
	}
	stateSet := make(map[string]struct{})
	for _, s := range m.States {
		if s == "" {
			return errors.New("state names must not be empty")
		}
		if _, exists := stateSet[s]; exists {
			return fmt.Errorf("duplicate state %q", s)
		}
		stateSet[s] = struct{}{}
	}
	if _, ok := stateSet[m.Initial]; !ok {
		return fmt.Errorf("initial state %q must be in states list", m.Initial)
	}
	if len(m.Finals) == 0 {
		return errors.New("finals list must not be empty")
	}
	for _, f := range m.Finals {
		if _, ok := stateSet[f]; !ok {
			return fmt.Errorf("final state %q must be in states list", f)
		}
	}
	if len(m.Transitions) == 0 {
		return errors.New("transitions must not be empty")
	}
	for state, list := range m.Transitions {
		if _, ok := stateSet[state]; !ok {
			return fmt.Errorf("transition defined for unknown state %q", state)
		}
		seenReads := make(map[string]struct{})
		for _, t := range list {
			if _, ok := alphaSet[t.Read]; !ok {
				return fmt.Errorf("transition in state %q reads unknown symbol %q", state, t.Read)
			}
			if _, ok := alphaSet[t.Write]; !ok {
				return fmt.Errorf("transition in state %q writes unknown symbol %q", state, t.Write)
			}
			if _, ok := stateSet[t.ToState]; !ok {
				return fmt.Errorf("transition in state %q goes to unknown state %q", state, t.ToState)
			}
			if t.Action != ActionLeft && t.Action != ActionRight {
				return fmt.Errorf("transition in state %q has invalid action %q", state, t.Action)
			}
			if _, exists := seenReads[t.Read]; exists {
				return fmt.Errorf("duplicate transition in state %q for read symbol %q", state, t.Read)
			}
			seenReads[t.Read] = struct{}{}
		}
	}
	return nil
}

func (m Machine) ValidateInput(input string) error {
	alphaSet := m.alphaSet
	if alphaSet == nil {
		alphaSet = toSet(m.Alphabet)
	}

	// Input must not contain blank symbol
	if strings.Contains(input, m.Blank) {
		return fmt.Errorf("input must not contain blank symbol %q", m.Blank)
	}

	// Every rune must be in alphabet
	for _, r := range input {
		s := string(r)
		if _, ok := alphaSet[s]; !ok {
			return fmt.Errorf("input contains symbol %q not in alphabet", s)
		}
	}
	return nil
}

func (m *Machine) Describe(w io.Writer) {
	fmt.Fprintln(w, "***")
	fmt.Fprintf(w, "* %s\n", m.Name)
	fmt.Fprintln(w, "*")
	fmt.Fprintln(w, "***")

	alpha := append([]string(nil), m.Alphabet...)
	states := append([]string(nil), m.States...)
	finals := append([]string(nil), m.Finals...)
	sort.Strings(alpha)
	sort.Strings(states)
	sort.Strings(finals)

	fmt.Fprintf(w, "alphabet: [%s]\n", strings.Join(alpha, ", "))
	fmt.Fprintf(w, "states  : [%s]\n", strings.Join(states, ", "))
	fmt.Fprintf(w, "initial : %s\n", m.Initial)
	fmt.Fprintf(w, "finals  : [%s]\n", strings.Join(finals, ", "))

	// Print transitions in a stable order
	stateNames := make([]string, 0, len(m.Transitions))
	for s := range m.Transitions {
		stateNames = append(stateNames, s)
	}
	sort.Strings(stateNames)
	for _, s := range stateNames {
		for _, t := range m.Transitions[s] {
			fmt.Fprintf(w, "[%s, %s) -> (%s, %s, %s)\n",
				s, t.Read, t.ToState, t.Write, t.Action)
		}
	}
	fmt.Fprintln(w, "***")
}



