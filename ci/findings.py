#!/usr/bin/env python3
"""Filtert Analyse-Befunde auf geaenderte Dateien (siehe ci/gate.sh).

Aufruf:
    findings.py gosec <server-root> <changed-file>...      < gosec.json
    findings.py staticcheck <server-root> <changed-file>... < staticcheck.json

Exit 0: keine Befunde in geaenderten Dateien.
Exit 1: Befunde, sie stehen auf stdout.
Exit 2: Eingabe nicht lesbar. Ein Gate darf an einer Ausgabe, die es nicht
        versteht, nie gruen werden.
"""

import json
import sys


def rel(path, root):
    path = path.removeprefix(root.rstrip("/") + "/")
    return path.removeprefix("./")


def gosec(stream, root):
    data = json.load(stream)
    if not isinstance(data, dict) or "Issues" not in data:
        raise ValueError("kein gosec-JSON")
    for issue in data["Issues"] or []:
        yield rel(issue["file"], root), issue["line"], f'{issue["rule_id"]} {issue["details"]}'


def staticcheck(stream, root):
    # staticcheck -f json schreibt ein Objekt je Zeile.
    for line in stream:
        line = line.strip()
        if not line:
            continue
        item = json.loads(line)
        loc = item["location"]
        yield rel(loc["file"], root), loc["line"], f'{item["code"]} {item["message"]}'


def main(argv):
    if len(argv) < 3 or argv[1] not in ("gosec", "staticcheck"):
        print(__doc__, file=sys.stderr)
        return 2
    tool, root, changed = argv[1], argv[2], set(argv[3:])
    reader = gosec if tool == "gosec" else staticcheck
    try:
        findings = list(reader(sys.stdin, root))
    except (ValueError, KeyError, TypeError) as err:
        print(f"{tool}: Ausgabe nicht lesbar: {err}", file=sys.stderr)
        return 2
    hits = [f for f in findings if f[0] in changed]
    for path, line, text in sorted(hits):
        print(f"{path}:{line}: {text}")
    print(f"{tool}: {len(findings)} Befunde gesamt, {len(hits)} in geaenderten Dateien",
          file=sys.stderr)
    return 1 if hits else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
