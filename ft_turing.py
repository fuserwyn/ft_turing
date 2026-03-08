#!/usr/bin/env python3
from __future__ import annotations

import json
import sys
from dataclasses import dataclass
from typing import Dict, List, Mapping, Optional, Tuple


USAGE_TEXT = """usage: ft_turing [-h] jsonfile input
positional arguments:
  jsonfile   json description of the machine
  input      input of the machine
optional arguments:
  -h, --help show this help message and exit
"""

MAX_STEPS = 1_000_000


@dataclass(frozen=True)
class Transition:
    read: str
    to_state: str
    write: str
    action: str  # LEFT | RIGHT


@dataclass(frozen=True)
class Machine:
    name: str
    alphabet: Tuple[str, ...]
    blank: str
    states: Tuple[str, ...]
    initial: str
    finals: Tuple[str, ...]
    transitions: Mapping[str, Tuple[Transition, ...]]
    alpha_set: frozenset[str]
    state_set: frozenset[str]
    finals_set: frozenset[str]
    trans_index: Mapping[str, Mapping[str, Transition]]


@dataclass(frozen=True)
class Configuration:
    tape: Mapping[int, str]
    head: int
    state: str
    steps: int


@dataclass(frozen=True)
class StepEvent:
    before: Configuration
    read: str
    transition: Optional[Transition]


@dataclass(frozen=True)
class SimulationResult:
    events: Tuple[StepEvent, ...]
    final: Configuration
    error: Optional[str]


def fail(message: str) -> None:
    raise ValueError(message)


def _require_obj(value: object, what: str) -> Dict[str, object]:
    if not isinstance(value, dict):
        fail(f"invalid json description: expected object for {what}")
    return value


def _require_str(value: object, what: str) -> str:
    if not isinstance(value, str):
        fail(f"invalid json description: expected string for {what}")
    return value


def _require_str_list(value: object, what: str) -> List[str]:
    if not isinstance(value, list):
        fail(f"invalid json description: expected array for {what}")
    out: List[str] = []
    for item in value:
        if not isinstance(item, str):
            fail(f"invalid json description: expected string array for {what}")
        out.append(item)
    return out


def _parse_transition(value: object) -> Transition:
    obj = _require_obj(value, "transition")
    read = _require_str(obj.get("read"), "transition.read")
    to_state = _require_str(obj.get("to_state"), "transition.to_state")
    write = _require_str(obj.get("write"), "transition.write")
    action = _require_str(obj.get("action"), "transition.action")
    if action not in {"LEFT", "RIGHT"}:
        fail(f"transition has invalid action {action!r}")
    return Transition(read=read, to_state=to_state, write=write, action=action)


def _as_unique_single_char_symbols(symbols: List[str]) -> frozenset[str]:
    if not symbols:
        fail("alphabet must not be empty")
    out = set()
    for s in symbols:
        if len(s) != 1:
            fail(f"alphabet symbol {s!r} must be a single character")
        if s in out:
            fail(f"duplicate alphabet symbol {s!r}")
        out.add(s)
    return frozenset(out)


def _as_unique_states(states: List[str]) -> frozenset[str]:
    if not states:
        fail("states list must not be empty")
    out = set()
    for s in states:
        if s == "":
            fail("state names must not be empty")
        if s in out:
            fail(f"duplicate state {s!r}")
        out.add(s)
    return frozenset(out)


def _build_trans_index(
    transitions: Mapping[str, Tuple[Transition, ...]]
) -> Mapping[str, Mapping[str, Transition]]:
    index: Dict[str, Dict[str, Transition]] = {}
    for state, transition_list in transitions.items():
        by_read: Dict[str, Transition] = {}
        for t in transition_list:
            if t.read in by_read:
                fail(
                    f"duplicate transition in state {state!r} for read symbol {t.read!r}"
                )
            by_read[t.read] = t
        index[state] = by_read
    return index


def load_machine(path: str) -> Machine:
    try:
        with open(path, "r", encoding="utf-8") as f:
            data = json.load(f)
    except FileNotFoundError as exc:
        fail(f"cannot read json file: {exc}")
    except json.JSONDecodeError as exc:
        fail(f"invalid json description: {exc}")
    except OSError as exc:
        fail(f"cannot read json file: {exc}")

    obj = _require_obj(data, "top-level")
    name = _require_str(obj.get("name"), "name")
    if name == "":
        fail("missing machine name")

    alphabet = _require_str_list(obj.get("alphabet"), "alphabet")
    blank = _require_str(obj.get("blank"), "blank")
    states = _require_str_list(obj.get("states"), "states")
    initial = _require_str(obj.get("initial"), "initial")
    finals = _require_str_list(obj.get("finals"), "finals")

    trans_raw = _require_obj(obj.get("transitions"), "transitions")
    if len(trans_raw) == 0:
        fail("transitions must not be empty")

    transitions: Dict[str, Tuple[Transition, ...]] = {}
    for state, entries in trans_raw.items():
        if not isinstance(entries, list):
            fail(f"transitions[{state}] must be an array")
        transitions[state] = tuple(_parse_transition(x) for x in entries)

    alpha_set = _as_unique_single_char_symbols(alphabet)
    if blank not in alpha_set:
        fail(f"blank symbol {blank!r} must be part of alphabet")

    state_set = _as_unique_states(states)
    if initial not in state_set:
        fail(f"initial state {initial!r} must be in states list")

    if not finals:
        fail("finals list must not be empty")
    finals_set = set()
    for final_state in finals:
        if final_state not in state_set:
            fail(f"final state {final_state!r} must be in states list")
        finals_set.add(final_state)

    for state, transition_list in transitions.items():
        if state not in state_set:
            fail(f"transition defined for unknown state {state!r}")
        for t in transition_list:
            if t.read not in alpha_set:
                fail(
                    f"transition in state {state!r} reads unknown symbol {t.read!r}"
                )
            if t.write not in alpha_set:
                fail(
                    f"transition in state {state!r} writes unknown symbol {t.write!r}"
                )
            if t.to_state not in state_set:
                fail(
                    f"transition in state {state!r} goes to unknown state {t.to_state!r}"
                )

    trans_index = _build_trans_index(transitions)
    return Machine(
        name=name,
        alphabet=tuple(alphabet),
        blank=blank,
        states=tuple(states),
        initial=initial,
        finals=tuple(finals),
        transitions=transitions,
        alpha_set=frozenset(alpha_set),
        state_set=frozenset(state_set),
        finals_set=frozenset(finals_set),
        trans_index=trans_index,
    )


def validate_input(machine: Machine, input_value: str) -> None:
    if machine.blank in input_value:
        fail(f"input must not contain blank symbol {machine.blank!r}")
    for ch in input_value:
        symbol = ch
        if symbol not in machine.alpha_set:
            fail(f"input contains symbol {symbol!r} not in alphabet")


def describe_machine(machine: Machine) -> str:
    lines = ["***", f"* {machine.name}", "*", "***"]
    lines.append(f"alphabet: [{', '.join(sorted(machine.alphabet))}]")
    lines.append(f"states  : [{', '.join(sorted(machine.states))}]")
    lines.append(f"initial : {machine.initial}")
    lines.append(f"finals  : [{', '.join(sorted(machine.finals))}]")
    for state in sorted(machine.transitions.keys()):
        for t in machine.transitions[state]:
            lines.append(
                f"[{state}, {t.read}) -> ({t.to_state}, {t.write}, {t.action})"
            )
    lines.append("***")
    return "\n".join(lines)


def build_tape(input_value: str) -> Mapping[int, str]:
    return {idx: ch for idx, ch in enumerate(input_value)}


def initial_configuration(machine: Machine, input_value: str) -> Configuration:
    return Configuration(tape=build_tape(input_value), head=0, state=machine.initial, steps=0)


def read_symbol(tape: Mapping[int, str], head: int, blank: str) -> str:
    return tape.get(head, blank)


def write_tape(
    tape: Mapping[int, str], position: int, symbol: str, blank: str
) -> Mapping[int, str]:
    # Persistent-style update: produce a new mapping.
    next_tape = dict(tape)
    if symbol == blank:
        next_tape.pop(position, None)
    else:
        next_tape[position] = symbol
    return next_tape


def apply_transition(conf: Configuration, t: Transition, blank: str) -> Configuration:
    next_head = conf.head - 1 if t.action == "LEFT" else conf.head + 1
    return Configuration(
        tape=write_tape(conf.tape, conf.head, t.write, blank),
        head=next_head,
        state=t.to_state,
        steps=conf.steps + 1,
    )


def next_step(
    machine: Machine, conf: Configuration
) -> Tuple[StepEvent, Configuration, bool, Optional[str]]:
    read = read_symbol(conf.tape, conf.head, machine.blank)
    event = StepEvent(before=conf, read=read, transition=None)

    if conf.state in machine.finals_set:
        return event, conf, True, None

    by_read = machine.trans_index.get(conf.state)
    if by_read is None:
        return (
            event,
            conf,
            False,
            f"blocked: no transitions defined for state {conf.state!r}",
        )

    t = by_read.get(read)
    if t is None:
        return (
            event,
            conf,
            False,
            f"blocked: no transition for state {conf.state!r} and symbol {read!r}",
        )

    event_with_transition = StepEvent(before=conf, read=read, transition=t)
    return (
        event_with_transition,
        apply_transition(conf, t, machine.blank),
        False,
        None,
    )


def simulate(machine: Machine, input_value: str, max_steps: int = MAX_STEPS) -> SimulationResult:
    conf = initial_configuration(machine, input_value)
    events: List[StepEvent] = []

    while conf.steps < max_steps:
        event, next_conf, done, error = next_step(machine, conf)
        events.append(event)
        if done:
            return SimulationResult(events=tuple(events), final=conf, error=None)
        if error is not None:
            return SimulationResult(events=tuple(events), final=conf, error=error)
        conf = next_conf

    return SimulationResult(
        events=tuple(events),
        final=conf,
        error=f"blocked: maximum number of steps ({max_steps}) exceeded",
    )


def tape_bounds(tape: Mapping[int, str], blank: str) -> Tuple[int, int]:
    non_blank_positions = [idx for idx, symbol in tape.items() if symbol != blank]
    if not non_blank_positions:
        return (0, 0)
    return (min(non_blank_positions), max(non_blank_positions))


def render_tape(tape: Mapping[int, str], head: int, blank: str) -> str:
    mn, mx = tape_bounds(tape, blank)
    left = min(mn, head)
    right = max(mx, head)
    chunks: List[str] = ["["]
    for idx in range(left, right + 4):
        sym = read_symbol(tape, idx, blank)
        if idx == head:
            chunks.append(f"<{sym}>")
        else:
            chunks.append(sym)
    chunks.append("...] ")
    return "".join(chunks)


def format_step_event(event: StepEvent, blank: str) -> str:
    tape_view = render_tape(event.before.tape, event.before.head, blank)
    if event.transition is None:
        return f"{tape_view}({event.before.state}, {event.read})"
    t = event.transition
    return (
        f"{tape_view}({event.before.state}, {event.read}) -> "
        f"({t.to_state}, {t.write}, {t.action})"
    )


def run_machine(machine: Machine, input_value: str) -> Tuple[str, Optional[str]]:
    result = simulate(machine, input_value, MAX_STEPS)
    lines = [describe_machine(machine)]
    lines.extend(format_step_event(e, machine.blank) for e in result.events)
    lines.append(f"Time complexity (steps): {result.final.steps}")
    return ("\n".join(lines), result.error)


def parse_args(argv: List[str]) -> Tuple[Optional[Tuple[str, str]], bool]:
    args = argv[1:]
    if args in (["-h"], ["--help"]):
        return (None, True)
    if len(args) != 2:
        return (None, False)
    return ((args[0], args[1]), False)


def main(argv: List[str]) -> int:
    parsed, is_help = parse_args(argv)
    if is_help:
        sys.stderr.write(USAGE_TEXT)
        return 0
    if parsed is None:
        sys.stderr.write(USAGE_TEXT)
        return 1

    json_path, input_value = parsed

    try:
        machine = load_machine(json_path)
    except ValueError as exc:
        sys.stderr.write(f"error: {exc}\n")
        return 1

    try:
        validate_input(machine, input_value)
    except ValueError as exc:
        sys.stderr.write(f"invalid input: {exc}\n")
        return 1

    output, runtime_error = run_machine(machine, input_value)
    sys.stdout.write(output + "\n")
    if runtime_error is not None:
        sys.stderr.write(f"runtime error: {runtime_error}\n")
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
