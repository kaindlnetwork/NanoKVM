#!/usr/bin/env bash
#
# Merge-Gates des gehaerteten NanoKVM-Forks (kaindlnetwork/nanokvm).
#
# Diese Datei ist die einzige Stelle, die beschreibt, WAS ein Gate tut.
# .github/workflows/gates.yaml ruft nur auf. Ein Gate, das hier gruen ist,
# ist in der Pipeline gruen, auf einem selbst gehosteten Runner ebenso und
# lokal auch:
#
#   ./ci/gate.sh all              # alle Gates ohne Netz-Abhaengigkeit
#   ./ci/gate.sh go-test gosec    # nur einzelne
#
# "all" laeuft bewusst weiter, wenn ein Gate faellt, und meldet am Ende alle
# Fehlschlaege. Sonst kostet jede Runde nur den jeweils ersten Fehler.
#
# Upstream-Code (sipeed/NanoKVM) erfuellt nicht jede Regel dieses Forks.
# Die Analyse-Gates (staticcheck, gosec) melden deshalb nur Befunde in
# Dateien, die gegenueber GATE_BASE geaendert sind: neuer und angefasster
# Code muss sauber sein, Altbestand blockiert keinen Fix.

set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT" || exit 1

# Reihenfolge = Laufzeit aufsteigend.
ALL_GATES=(workflow-timeouts actionlint shellcheck prose gofmt go-vet
           go-test staticcheck gosec web-lint web-build)

# Nicht in "all": brauchen Netz zu einer Schwachstellen-Datenbank. Die
# Pipeline ruft sie trotzdem in jedem Lauf auf.
EXTRA_GATES=(govulncheck web-audit)

# Vergleichsbasis fuer "geaenderte Dateien". In der Pipeline setzt der
# Workflow die Basis des Pull Requests.
GATE_BASE="${GATE_BASE:-origin/main}"

# Obergrenze fuer timeout-minutes in den Workflows dieses Forks.
WORKFLOW_TIMEOUT_MAX=60

# Workflows, die von upstream stammen und deren Werte dieser Fork nicht
# bestimmt. Sie laufen im Fork nicht (siehe docs/fork/CI.md) und werden
# deshalb nicht geprueft; jede Aenderung daran waere ein Merge-Konflikt bei
# jedem Upstream-Sync.
UPSTREAM_WORKFLOWS=(
    .github/workflows/builder-image.yml
    .github/workflows/create-tag.yml
    .github/workflows/issue-tracker.yml
    .github/workflows/package.yml
    .github/workflows/release.yml
)

# Dokumente, die dieser Fork schreibt. Nur sie laufen durch "prose"; die
# Upstream-Doku ist nicht unsere.
PROSE_FILES=(CLAUDE.md docs/fork .claude .github/PULL_REQUEST_TEMPLATE.md)
PROSE_DASH_PER_1000=0.5

# Gegen libkvm.so wird gelinkt. Die echte Bibliothek ist riscv64; fuer Tests
# auf dem CI-Host baut ci/stub/kvm_stub.c einen Ersatz mit denselben Symbolen.
STUB_DIR="${STUB_DIR:-${TMPDIR:-/tmp}/nanokvm-kvmstub}"

usage() {
    cat <<EOF
Aufruf: ci/gate.sh <gate> [<gate> ...]
        ci/gate.sh all

Gates: ${ALL_GATES[*]}
Nur einzeln (nicht in "all"): ${EXTRA_GATES[*]}
EOF
}

if command -v git >/dev/null 2>&1; then
    git config --global --add safe.directory "$REPO_ROOT" 2>/dev/null || true
fi

need() {
    local tool
    for tool in "$@"; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            echo "Werkzeug fehlt: $tool (im CI-Image ci/Dockerfile enthalten)" >&2
            return 1
        fi
    done
}

# Geaenderte, noch vorhandene Dateien gegenueber GATE_BASE, inklusive noch
# nicht eingecheckter. Faellt die Basis weg (flacher Klon), zaehlt alles.
changed_files() {
    local base
    if base="$(git merge-base HEAD "$GATE_BASE" 2>/dev/null)"; then
        {
            git diff --name-only --diff-filter=d "$base"
            git ls-files --others --exclude-standard
        } | sort -u
    else
        echo "GATE_BASE $GATE_BASE nicht gefunden, pruefe alle Dateien" >&2
        git ls-files
    fi
}

changed_go_files() {
    changed_files | grep '^server/.*\.go$' | sed 's#^server/##'
}

build_stub() {
    need cc || return 1
    mkdir -p "$STUB_DIR"
    cc -shared -fPIC -I server/include -o "$STUB_DIR/libkvm.so" ci/stub/kvm_stub.c
}

go_env() {
    export CGO_ENABLED=1
    export CGO_LDFLAGS="-L$STUB_DIR"
    export LD_LIBRARY_PATH="$STUB_DIR${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
}

gate_workflow_timeouts() {
    local file fail=0 skip upstream
    for file in .github/workflows/*.yaml .github/workflows/*.yml; do
        [ -e "$file" ] || continue
        skip=0
        for upstream in "${UPSTREAM_WORKFLOWS[@]}"; do
            [ "$file" = "$upstream" ] && skip=1
        done
        [ "$skip" = 1 ] && continue
        python3 - "$file" "$WORKFLOW_TIMEOUT_MAX" <<'EOF' || fail=1
import sys, yaml
path, limit = sys.argv[1], int(sys.argv[2])
doc = yaml.safe_load(open(path))
bad = []
for name, job in (doc.get("jobs") or {}).items():
    if "uses" in job:
        continue
    t = job.get("timeout-minutes")
    if not isinstance(t, int) or t > limit:
        bad.append(f"{path}: job {name}: timeout-minutes={t!r} (Pflicht, <= {limit})")
for line in bad:
    print(line)
sys.exit(1 if bad else 0)
EOF
    done
    return "$fail"
}

gate_actionlint() {
    need actionlint || return 1
    local files=() file skip upstream
    for file in .github/workflows/*.yaml .github/workflows/*.yml; do
        [ -e "$file" ] || continue
        skip=0
        for upstream in "${UPSTREAM_WORKFLOWS[@]}"; do
            [ "$file" = "$upstream" ] && skip=1
        done
        [ "$skip" = 0 ] && files+=("$file")
    done
    [ "${#files[@]}" -eq 0 ] && return 0
    actionlint -shellcheck= "${files[@]}"
}

gate_shellcheck() {
    need shellcheck || return 1
    local files=()
    mapfile -t files < <(git ls-files --cached --others --exclude-standard 'ci/*.sh')
    [ "${#files[@]}" -eq 0 ] && return 0
    shellcheck "${files[@]}"
}

# Mechanisch pruefbare Stilregeln fuer die Doku dieses Forks, uebernommen aus
# kaindlnetwork/scaling-adventure (siehe CLAUDE.md, "Wie die Dokumentation
# geschrieben ist").
gate_prose() {
    local files=()
    mapfile -t files < <(git ls-files --cached --others --exclude-standard -- "${PROSE_FILES[@]}" | grep -E '\.md$')
    [ "${#files[@]}" -eq 0 ] && return 0
    python3 - "$PROSE_DASH_PER_1000" "${files[@]}" <<'EOF'
import re, sys
limit = float(sys.argv[1])
words = ["nahtlos", "ganzheitlich", "leistungsstark", "revolutionär",
         "bahnbrechend", "state-of-the-art", "robust und skalierbar",
         "best-in-class", "game-changer", "seamless", "cutting-edge"]
emoji = re.compile("[\U0001F300-\U0001FAFF☀-➿]")
fail = False
for path in sys.argv[2:]:
    text = open(path, encoding="utf-8").read()
    text = re.sub(r"```.*?```", "", text, flags=re.S)
    text = re.sub(r"`[^`]*`", "", text)
    n_words = max(len(text.split()), 1)
    dashes = text.count("—")
    if dashes / n_words * 1000 > limit:
        print(f"{path}: {dashes} Gedankenstriche (—) bei {n_words} Woertern")
        fail = True
    low = text.lower()
    for w in words:
        if w in low:
            print(f"{path}: Behauptungswort '{w}'")
            fail = True
    for line in text.splitlines():
        if line.startswith("#") and emoji.search(line):
            print(f"{path}: Emoji in Ueberschrift: {line.strip()}")
            fail = True
sys.exit(1 if fail else 0)
EOF
}

gate_gofmt() {
    need gofmt || return 1
    local files=() out
    mapfile -t files < <(changed_go_files)
    [ "${#files[@]}" -eq 0 ] && return 0
    out="$(cd server && gofmt -l "${files[@]}")"
    if [ -n "$out" ]; then
        echo "Nicht gofmt-formatiert:"
        echo "$out"
        return 1
    fi
}

gate_go_vet() {
    need go || return 1
    build_stub || return 1
    go_env
    (cd server && go vet ./...)
}

gate_go_test() {
    need go || return 1
    build_stub || return 1
    go_env
    (cd server && go test -count=1 -race ./...)
}

# Meldet nur Befunde in geaenderten Dateien, siehe Kopf der Datei. Der
# Filter (ci/findings.py) liest JSON und scheitert hart an allem, was er
# nicht versteht: eine leere oder kaputte Werkzeugausgabe ist kein Gruen.
gate_staticcheck() {
    need staticcheck go python3 || return 1
    build_stub || return 1
    go_env
    local json changed=()
    mapfile -t changed < <(changed_go_files)
    # staticcheck endet mit 1, sobald es etwas findet; das wertet der Filter.
    json="$(cd server && staticcheck -f json ./...)"
    printf '%s\n' "$json" | python3 ci/findings.py staticcheck "$REPO_ROOT/server" "${changed[@]}"
}

gate_gosec() {
    need gosec go python3 || return 1
    build_stub || return 1
    go_env
    local json changed=()
    mapfile -t changed < <(changed_go_files)
    # G104 (ungeprueftes err) deckt go vet/staticcheck ab; G115
    # (int-Konversion) ist im Altbestand allgegenwaertig.
    json="$(cd server && gosec -quiet -fmt json -exclude=G104,G115 -exclude-dir=dl_lib ./... 2>/dev/null)"
    printf '%s\n' "$json" | python3 ci/findings.py gosec "$REPO_ROOT/server" "${changed[@]}"
}

gate_govulncheck() {
    need govulncheck go || return 1
    build_stub || return 1
    go_env
    (cd server && govulncheck ./...)
}

web_install() {
    need pnpm || return 1
    (cd web && pnpm install --frozen-lockfile --reporter=silent)
}

gate_web_lint() {
    web_install || return 1
    (cd web && pnpm exec eslint . --max-warnings=12)
}

gate_web_build() {
    web_install || return 1
    (cd web && pnpm run build)
}

gate_web_audit() {
    web_install || return 1
    (cd web && pnpm audit --prod --audit-level=high)
}

run_gate() {
    local gate="$1" fn
    fn="gate_${gate//-/_}"
    if ! declare -F "$fn" >/dev/null; then
        echo "Unbekanntes Gate: $gate" >&2
        usage >&2
        return 2
    fi
    echo "::group::gate $gate"
    local start rc
    start=$(date +%s)
    "$fn"
    rc=$?
    echo "::endgroup::"
    if [ "$rc" -eq 0 ]; then
        echo "PASS $gate ($(( $(date +%s) - start ))s)"
    else
        echo "FAIL $gate ($(( $(date +%s) - start ))s)"
    fi
    return "$rc"
}

main() {
    [ "$#" -eq 0 ] && { usage; exit 2; }
    local gates=() failed=()
    for arg in "$@"; do
        if [ "$arg" = all ]; then
            gates+=("${ALL_GATES[@]}")
        else
            gates+=("$arg")
        fi
    done
    for gate in "${gates[@]}"; do
        run_gate "$gate" || failed+=("$gate")
    done
    echo
    if [ "${#failed[@]}" -gt 0 ]; then
        echo "Gefallene Gates: ${failed[*]}"
        exit 1
    fi
    echo "Alle Gates gruen: ${gates[*]}"
}

main "$@"
