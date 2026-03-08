## ft_turing (Python implementation)

This project is a Python implementation of a single-tape, single-head Turing machine simulator.

### Requirements

- Python 3.10+ (`python3`)

### Build

```bash
cd "/Users/lofitravel/Desktop/IT 42/ft_turing/ft_turing"
make
```

This builds `./ft_turing`.

### Usage

```bash
./ft_turing --help
```

```text
usage: ft_turing [-h] jsonfile input
positional arguments:
  jsonfile   json description of the machine
  input      input of the machine
optional arguments:
  -h, --help show this help message and exit
```

### Example machines

The directory `machines/` contains several example Turing machines:

- `unary_sub.json` – unary subtraction (from the subject PDF).
- `unary_add.json` – unary addition.
- `palindrome.json` – decides if input over `{0,1}` is a palindrome; writes `y` or `n` at the right end.
- `zero_n_one_n.json` – decides if input is in the language `0^n1^n`; writes `y` or `n`.
- `zero_2n.json` – decides if input is in the language `0^{2n}`; writes `y` or `n`.
- `meta_unary_add.json` – encoded unary-add runner.

### `meta_unary_add` encoding

`meta_unary_add` expects input in the form:

```text
m...m|<unary_add_payload>
```

Where:

- `m...m` is the encoded metadata area (chosen encoding).
- `|` separates metadata from the payload.
- `<unary_add_payload>` is the unary addition input (for example `111+11=`).

If the `|` separator is missing, the machine writes `n` and halts.

### Run examples

```bash
cd "/Users/lofitravel/Desktop/IT 42/ft_turing/ft_turing"

# Unary subtraction
./ft_turing machines/unary_sub.json "111-11="

# Unary addition
./ft_turing machines/unary_add.json "111+11="

# Palindrome check
./ft_turing machines/palindrome.json "0110"

# 0^n1^n
./ft_turing machines/zero_n_one_n.json "000111"

# 0^{2n}
./ft_turing machines/zero_2n.json "0000"

# Meta unary_add (encoded input)
./ft_turing machines/meta_unary_add.json "mmm|111+11="
```

Each run prints:

- A description of the machine (alphabet, states, transitions).
- The full step-by-step trace of the tape with the head marked by `< >`.
- The computed **time complexity** as the number of executed steps.



