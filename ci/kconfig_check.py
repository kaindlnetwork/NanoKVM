#!/usr/bin/env python3
"""Prueft eine aufgeloeste Kernel-.config gegen die Fragmente des Forks.

Aufruf:
    kconfig_check.py <.config> --required ci/kernel/required.config \\
        --remove ci/kernel/remove.config --harden ci/kernel/harden.config [--strict]

Die Fragmente sind normale Kconfig-Fragmente (scripts/kconfig/merge_config.sh
kann sie direkt verwenden). Jeder Eintrag braucht darueber eine Zeile
"# Grund: ...", damit die Liste begruendet bleibt.

- required: muss gesetzt sein (=y oder =m). Fehlt etwas, funktioniert der
  NanoKVM nicht: Exit 1, immer.
- remove:   soll nicht gesetzt sein. Ist es noch gesetzt, wird es gemeldet.
- harden:   soll gesetzt sein. Fehlt es, wird es gemeldet.
  remove/harden machen das Ergebnis nur mit --strict rot.

Exit 2: Fragment kaputt (Eintrag ohne Grund, Symbol doppelt, unlesbar).
"""

import argparse
import re
import sys

ENTRY = re.compile(r"^(?:CONFIG_(\w+)=(\S+)|# CONFIG_(\w+) is not set)\s*$")


class FragmentError(Exception):
    pass


def parse_fragment(path):
    """Liefert {symbol: (wert, grund)}; wert ist "y", "m" oder "n"."""
    entries, reason = {}, None
    with open(path, encoding="utf-8") as handle:
        for number, raw in enumerate(handle, 1):
            line = raw.rstrip("\n")
            match = ENTRY.match(line)
            if match:
                symbol = match.group(1) or match.group(3)
                value = "n" if match.group(3) else match.group(2)
                if not reason:
                    raise FragmentError(f"{path}:{number}: CONFIG_{symbol} ohne '# Grund:'-Zeile")
                if symbol in entries:
                    raise FragmentError(f"{path}:{number}: CONFIG_{symbol} doppelt")
                entries[symbol] = (value, reason)
                continue
            stripped = line.strip()
            if stripped.startswith("# Grund:"):
                reason = stripped[len("# Grund:"):].strip()
            elif not stripped:
                reason = None
    return entries


def load_fragments(paths):
    merged = {}
    for path in paths:
        for symbol, entry in parse_fragment(path).items():
            if symbol in merged:
                raise FragmentError(f"CONFIG_{symbol} steht in mehreren Fragmenten")
            merged[symbol] = entry
    return merged


def parse_config(path):
    values = {}
    with open(path, encoding="utf-8") as handle:
        for line in handle:
            match = ENTRY.match(line.rstrip("\n"))
            if match:
                symbol = match.group(1) or match.group(3)
                values[symbol] = "n" if match.group(3) else match.group(2)
    return values


def enabled(value):
    return value not in (None, "n", '""')


def main(argv):
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("config")
    parser.add_argument("--required", required=True)
    parser.add_argument("--remove", required=True)
    parser.add_argument("--harden", required=True)
    parser.add_argument("--strict", action="store_true")
    args = parser.parse_args(argv[1:])

    try:
        required = parse_fragment(args.required)
        remove = parse_fragment(args.remove)
        harden = parse_fragment(args.harden)
        load_fragments([args.required, args.remove, args.harden])
        config = parse_config(args.config)
    except (OSError, FragmentError) as err:
        print(f"Fragment ungueltig: {err}")
        return 2

    missing = [(s, r) for s, (_, r) in sorted(required.items()) if not enabled(config.get(s))]
    leftover = [(s, r) for s, (_, r) in sorted(remove.items()) if enabled(config.get(s))]
    unhardened = [(s, r) for s, (_, r) in sorted(harden.items()) if not enabled(config.get(s))]

    for title, items in (("FEHLT (Pflicht)", missing),
                         ("NOCH AN (entbehrlich)", leftover),
                         ("NOCH AUS (Haertung)", unhardened)):
        print(f"{title}: {len(items)}")
        for symbol, reason in items:
            print(f"  CONFIG_{symbol}  ({reason})")

    if missing:
        return 1
    if args.strict and (leftover or unhardened):
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
