#!/usr/bin/env python3
"""Filtert Analyse-Befunde auf geaenderte Dateien (siehe ci/gate.sh).

Aufruf:
    findings.py gosec <server-root> --exit <rc> <changed-file>...       < gosec.json
    findings.py staticcheck <server-root> --exit <rc> <changed-file>... < staticcheck.json

<rc> ist der Exit-Code des Werkzeugs.

Exit 0: keine Befunde in geaenderten Dateien.
Exit 1: Befunde, sie stehen auf stdout.
Exit 2: Analyse unvollstaendig oder Eingabe nicht lesbar. Ein Gate darf an
        einer Ausgabe, die es nicht versteht, oder an einem Lauf, der nicht
        alles analysiert hat, nie gruen werden.
"""

import json
import sys


class Incomplete(Exception):
    """Das Werkzeug hat nicht alles analysiert."""


def rel(path, root):
    path = path.removeprefix(root.rstrip("/") + "/")
    return path.removeprefix("./")


def gosec(stream, root, rc):
    data = json.load(stream)
    if not isinstance(data, dict) or "Issues" not in data:
        raise ValueError("kein gosec-JSON")
    # Pakete, die gosec nicht laden konnte (Kompilierfehler, fehlende
    # Abhaengigkeit), stehen nur hier. Ihre Dateien wurden nicht geprueft.
    errors = data.get("Golang errors") or {}
    if errors:
        files = ", ".join(sorted(rel(f, root) or "<paket>" for f in errors))
        raise Incomplete(f"Analysefehler in: {files}")
    for issue in data["Issues"] or []:
        yield rel(issue["file"], root), issue["line"], f'{issue["rule_id"]} {issue["details"]}'


def staticcheck(stream, root, rc):
    # staticcheck -f json schreibt ein Objekt je Zeile und endet mit 1,
    # sobald es etwas meldet. Alles andere ist ein Abbruch.
    if rc not in (0, 1):
        raise Incomplete(f"staticcheck endete mit {rc}")
    items = []
    for line in stream:
        line = line.strip()
        if not line:
            continue
        item = json.loads(line)
        if item.get("code") == "compile" or not item["location"]["file"]:
            raise Incomplete(f'{item.get("code")}: {item.get("message", "").strip()}')
        items.append(item)
    if rc == 1 and not items:
        raise Incomplete("staticcheck endete mit 1 ohne Befund")
    for item in items:
        loc = item["location"]
        yield rel(loc["file"], root), loc["line"], f'{item["code"]} {item["message"]}'


def main(argv):
    if len(argv) < 5 or argv[1] not in ("gosec", "staticcheck") or argv[3] != "--exit":
        print(__doc__, file=sys.stderr)
        return 2
    tool, root, changed = argv[1], argv[2], set(argv[5:])
    try:
        rc = int(argv[4])
    except ValueError:
        print(__doc__, file=sys.stderr)
        return 2
    reader = gosec if tool == "gosec" else staticcheck
    try:
        found = list(reader(sys.stdin, root, rc))
    except Incomplete as err:
        print(f"{tool}: Analyse unvollstaendig: {err}", file=sys.stderr)
        return 2
    except (ValueError, KeyError, TypeError) as err:
        print(f"{tool}: Ausgabe nicht lesbar: {err}", file=sys.stderr)
        return 2
    hits = [f for f in found if f[0] in changed]
    for path, line, text in sorted(hits, key=lambda h: (h[0], int(h[1]))):
        print(f"{path}:{line}: {text}")
    print(f"{tool}: {len(found)} Befunde gesamt, {len(hits)} in geaenderten Dateien",
          file=sys.stderr)
    return 1 if hits else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
