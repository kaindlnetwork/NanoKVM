# kaindlnetwork/nanokvm: Projekt-Notizen für Claude

Gehärteter Fork von [sipeed/NanoKVM](https://github.com/sipeed/NanoKVM).
Ziel: dieselbe Funktion mit weniger Angriffsfläche. Jede Abweichung von
upstream ist ein Sicherheitsfix, eine sichere Voreinstellung oder die
Infrastruktur, die beides prüft.

> **Lese-Reihenfolge in einer neuen Session**
>
> 1. Diese Datei: Konventionen und Pflichten.
> 2. [`docs/fork/HARDENING.md`](docs/fork/HARDENING.md): welche Fixes es gibt,
>    welche offen sind und warum.
> 3. [`docs/fork/CI.md`](docs/fork/CI.md): Gates, Runner, Upstream-Sync.
> 4. Upstream-Doku unter `server/README.md`, `scripts/README.md`.

## Stack & Layout

- **Backend**: Go unter `server/` (gin, pion/webrtc). Läuft auf dem Gerät als
  root. Linkt per cgo gegen `server/dl_lib/libkvm.so` (riscv64).
- **Web-UI**: React/Vite unter `web/` (pnpm).
- **Gerät**: Init-Skripte und Payload unter `kvmapp/`.
- **Fork-Eigenes**: `ci/`, `docs/fork/`, `.claude/`, `.github/workflows/gates.yaml`,
  diese Datei.

## Test-Driven Development ist Pflicht

Für Sicherheitsfixes, Bugfixes, Refactorings und Verhaltensänderungen gilt
der Workflow aus
[`.claude/skills/test-driven-development/SKILL.md`](.claude/skills/test-driven-development/SKILL.md):

1. **RED**: failing Test zuerst (`./ci/gate.sh go-test`)
2. **GREEN**: minimaler Code, bis der Test grün ist
3. **REFACTOR**: aufräumen, Tests bleiben grün

Keine Produktivänderung ohne vorher gesehenen Fehlschlag. Der Fehlschlag muss
aus der fehlenden Absicherung kommen, nicht aus einem Tippfehler. Bei einem
Sicherheitsfix heißt das: der Test führt den Angriff aus und scheitert, weil
der Angriff gegen den alten Code gelingt.

Tests laufen ohne Gerät und ohne Netz. Gefaked wird die Hardware
(`ci/stub/kvm_stub.c` für `libkvm.so`, temporäre Verzeichnisse statt
`/etc/kvm`), nie die zu prüfende Logik.

## Ein Sicherheitsfix ist ein Pull Request

Jeder Fix bekommt einen eigenen Branch `fix/<thema>` und einen eigenen PR,
damit er einzeln geprüft, getestet und notfalls zurückgenommen werden kann.
Ein PR enthält:

- den Test, der den Angriff nachstellt,
- den Nachweis, dass dieser Test ohne den Fix rot ist (Ausgabe im PR-Text),
- den Fix,
- seine Beschreibung als eigene Datei `docs/fork/hardening/<thema>.md`
  (eine Datei je Fix, damit parallele PRs keinen Doku-Konflikt bekommen).

Fixes, die nur auf echter Hardware prüfbar sind (Init-Skripte,
Kernel-Module), sind Draft-PRs mit einer Prüfliste für das Gerät, bis jemand
sie auf einem NanoKVM durchgespielt hat.

## Upstream-Verträglichkeit

Der Fork soll upstream folgen können. Deshalb:

- **Upstream-Dateien minimal ändern.** Neue Logik in neue Dateien
  (`middleware/proxy.go`, `utils/safename.go`), im Bestand nur der Aufruf.
- **Upstream-Workflows nicht anfassen.** Sie sind im Fork deaktiviert, siehe
  `docs/fork/CI.md`.
- **Konfiguration rückwärtskompatibel.** Ein bestehendes
  `/etc/kvm/server.yaml` muss weiter laden. Wo ein alter Wert eine unsichere
  Bedeutung hatte (`loginLockoutDuration: 0` = aus), bekommt er die sichere
  und das Abschalten einen neuen, expliziten Wert (`-1`).
- **Fixes, die upstream auch helfen, gehen als PR zurück** an sipeed, sobald
  sie im Fork auf Hardware gelaufen sind.

## Gates vor dem Merge

```bash
./ci/gate.sh all                       # ohne Netz
./ci/gate.sh all govulncheck web-audit # wie die Pipeline
```

`ci/gate.sh` ist die einzige Stelle, an der steht, was ein Gate tut.
`.github/workflows/gates.yaml` ruft nur auf. Ein Gate gehört deshalb in das
Skript, nicht in den Workflow. Ausgeführt wird in einem vorgebauten Image
(`ci/Dockerfile`), dessen Tag aus seinem Inhalt entsteht.

`gofmt`, `staticcheck` und `gosec` melden nur Befunde in Dateien, die
gegenüber der Basis geändert sind. Upstream hat gut 150 gosec-Befunde; die blockieren
keinen Fix, aber jede Datei, die der Fork anfasst, muss danach sauber sein.

## Wie die Dokumentation geschrieben ist

Dieselben Regeln wie in kaindlnetwork/scaling-adventure, gehalten vom Gate
`prose` für die Dateien dieses Forks:

- Kein Gedankenstrich `—`. Komma, Punkt, Doppelpunkt oder Klammer.
- Keine Wörter, die eine Aussage durch eine Behauptung ersetzen
  (`nahtlos`, `leistungsstark`, ...). Statt „deutlich sicherer" der Angriff,
  der vorher ging und jetzt nicht mehr.
- Kein Emoji in einer Überschrift.
- Fettung ist ein Label am Absatzanfang, keine Betonung im Satz.

## Secrets gehören nicht ins Repository

Keine Schlüssel, Passwörter oder Tokens in Code, Tests oder Doku. Tests
erzeugen ihre Geheimnisse selbst (`test-secret`, temporäre Stores).
