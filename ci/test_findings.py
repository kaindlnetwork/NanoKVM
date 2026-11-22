"""Tests fuer ci/findings.py. Laufen im Gate ci-selftest."""

import io
import json
import os
import sys
import unittest

sys.path.insert(0, os.path.dirname(__file__))
import findings  # noqa: E402

ROOT = "/repo/server"


def run(tool, stdin, changed, exit_code="0"):
    out, err = io.StringIO(), io.StringIO()
    old = sys.stdin, sys.stdout, sys.stderr
    sys.stdin, sys.stdout, sys.stderr = io.StringIO(stdin), out, err
    try:
        rc = findings.main(["findings.py", tool, ROOT, "--exit", exit_code, *changed])
    finally:
        sys.stdin, sys.stdout, sys.stderr = old
    return rc, out.getvalue()


def gosec_report(issues=(), errors=None):
    return json.dumps({
        "Golang errors": errors or {},
        "Issues": [
            {"file": f"{ROOT}/{f}", "line": "3", "rule_id": "G204", "details": "x"}
            for f in issues
        ],
    })


def sc_item(path, code="SA1000"):
    return json.dumps({"code": code, "severity": "error",
                       "location": {"file": path, "line": 1, "column": 1},
                       "message": "m"})


class Gosec(unittest.TestCase):
    def test_finding_in_changed_file_fails(self):
        rc, out = run("gosec", gosec_report(["a.go"]), ["a.go"])
        self.assertEqual(rc, 1)
        self.assertIn("a.go:3", out)

    def test_finding_elsewhere_passes(self):
        rc, _ = run("gosec", gosec_report(["b.go"]), ["a.go"])
        self.assertEqual(rc, 0)

    def test_analysis_errors_fail_closed(self):
        report = gosec_report(["b.go"], errors={f"{ROOT}/a.go": [{"error": "syntax"}]})
        rc, _ = run("gosec", report, ["a.go"])
        self.assertEqual(rc, 2)

    def test_unreadable_output_fails_closed(self):
        rc, _ = run("gosec", "\x1b[31mcolored text", ["a.go"])
        self.assertEqual(rc, 2)


class Staticcheck(unittest.TestCase):
    def test_finding_in_changed_file_fails(self):
        rc, _ = run("staticcheck", sc_item(f"{ROOT}/a.go"), ["a.go"], exit_code="1")
        self.assertEqual(rc, 1)

    def test_compile_error_fails_closed(self):
        rc, _ = run("staticcheck", sc_item("", code="compile"), ["a.go"], exit_code="1")
        self.assertEqual(rc, 2)

    def test_abort_without_output_fails_closed(self):
        rc, _ = run("staticcheck", "", ["a.go"], exit_code="2")
        self.assertEqual(rc, 2)

    def test_nonzero_exit_without_findings_fails_closed(self):
        rc, _ = run("staticcheck", "", ["a.go"], exit_code="1")
        self.assertEqual(rc, 2)

    def test_clean_run_passes(self):
        rc, _ = run("staticcheck", "", ["a.go"], exit_code="0")
        self.assertEqual(rc, 0)


if __name__ == "__main__":
    unittest.main()
