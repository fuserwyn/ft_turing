import subprocess
import sys
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
ENTRYPOINT = ROOT / "ft_turing"
MACHINES = ROOT / "machines"

sys.path.insert(0, str(ROOT))
import ft_turing as core  # noqa: E402


def decision_symbol(conf: core.Configuration) -> str:
    if not conf.tape:
        return ""
    idx = max(conf.tape.keys())
    return conf.tape[idx]


class CliTests(unittest.TestCase):
    def test_help(self) -> None:
        proc = subprocess.run(
            [str(ENTRYPOINT), "--help"],
            cwd=ROOT,
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertEqual(proc.returncode, 0)
        self.assertIn("usage: ft_turing [-h] jsonfile input", proc.stderr)

    def test_invalid_input_rejected(self) -> None:
        proc = subprocess.run(
            [str(ENTRYPOINT), "machines/unary_sub.json", "11.1"],
            cwd=ROOT,
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertEqual(proc.returncode, 1)
        self.assertIn("invalid input:", proc.stderr)


class MachineBehaviorTests(unittest.TestCase):
    def run_sim(self, machine_name: str, input_value: str) -> core.SimulationResult:
        machine = core.load_machine(str(MACHINES / machine_name))
        core.validate_input(machine, input_value)
        return core.simulate(machine, input_value)

    def test_unary_add_runs_to_completion(self) -> None:
        result = self.run_sim("unary_add.json", "111+11=")
        self.assertIsNone(result.error)
        self.assertGreater(result.final.steps, 0)

    def test_palindrome_accepts(self) -> None:
        result = self.run_sim("palindrome.json", "0110")
        self.assertIsNone(result.error)
        self.assertEqual(decision_symbol(result.final), "y")

    def test_palindrome_rejects(self) -> None:
        result = self.run_sim("palindrome.json", "011")
        self.assertIsNone(result.error)
        self.assertEqual(decision_symbol(result.final), "n")

    def test_zero_n_one_n_accepts(self) -> None:
        result = self.run_sim("zero_n_one_n.json", "000111")
        self.assertIsNone(result.error)
        self.assertEqual(decision_symbol(result.final), "y")

    def test_zero_n_one_n_rejects(self) -> None:
        result = self.run_sim("zero_n_one_n.json", "00111")
        self.assertIsNone(result.error)
        self.assertEqual(decision_symbol(result.final), "n")

    def test_zero_2n_accepts(self) -> None:
        result = self.run_sim("zero_2n.json", "0000")
        self.assertIsNone(result.error)
        self.assertEqual(decision_symbol(result.final), "y")

    def test_zero_2n_rejects(self) -> None:
        result = self.run_sim("zero_2n.json", "000")
        self.assertIsNone(result.error)
        self.assertEqual(decision_symbol(result.final), "n")

    def test_meta_unary_add_runs(self) -> None:
        result = self.run_sim("meta_unary_add.json", "mmm|111+11=")
        self.assertIsNone(result.error)
        self.assertGreater(result.final.steps, 0)


if __name__ == "__main__":
    unittest.main()
