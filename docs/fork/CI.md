# CI: Gates, Runner und Upstream-Sync

Dieses Dokument beschreibt, wie die Merge-Gates aus [`CLAUDE.md`](../../CLAUDE.md)
ausgeführt werden. Aufbau nach dem Muster von kaindlnetwork/scaling-adventure.

## Drei Schichten

| Schicht | Datei | Zuständig für |
|---|---|---|
| Gates | `ci/gate.sh` | **Was** geprüft wird |
| Umgebung | `ci/Dockerfile`, `ci/deps-tag.sh` | **Womit** geprüft wird |
| Verkabelung | `.github/workflows/gates.yaml` | **Wo** geprüft wird |

Nur die unterste Schicht ist an GitHub gebunden, und sie enthält keine
Prüflogik. Ein Gate, das lokal grün ist, ist in der Pipeline grün.

## Gates

| Gate | Prüft | in `all` |
|---|---|---|
| `workflow-timeouts` | jeder Job der Fork-Workflows hat `timeout-minutes` <= 60 | ja |
| `actionlint` | Fork-Workflows | ja |
| `shellcheck` | `ci/*.sh` | ja |
| `prose` | Stilregeln der Fork-Doku | ja |
| `gofmt` | Formatierung, nur geänderte Dateien | ja |
| `go-vet` | `go vet ./...` | ja |
| `go-test` | `go test -race ./...` gegen den `libkvm.so`-Stub | ja |
| `staticcheck` | nur geänderte Dateien | ja |
| `gosec` | nur geänderte Dateien, ohne G104/G115 | ja |
| `web-lint` | ESLint, Warnungen dürfen nicht mehr werden | ja |
| `web-build` | `tsc` und `vite build` | ja |
| `govulncheck` | bekannte Schwachstellen in Go-Abhängigkeiten, die der Code erreicht | nein (Netz) |
| `web-audit` | `pnpm audit --prod`, ab „high“ | nein (Netz) |

Die Pipeline ruft `./ci/gate.sh all govulncheck web-audit` auf.

### Warum nur geänderte Dateien bei `gofmt`, `staticcheck` und `gosec`

Upstream liefert 16 staticcheck- und gut 150 gosec-Befunde, und nicht jede Datei ist gofmt-formatiert. Ein Gate, das
die alle meldet, ist ab dem ersten Tag rot und wird ignoriert. Stattdessen:
was der Fork anfasst, muss sauber sein. Ein begründeter Befund bekommt ein
`#nosec Gxxx` mit Kommentar, warum er hier kein Problem ist.

Der Filter steht in `ci/findings.py` und liest JSON. Er scheitert mit
Exit 2 an jeder Ausgabe, die er nicht versteht. Anlass: eine erste Fassung
filterte die eingefärbte Textausgabe von gosec, fand darin nie einen
Dateinamen und meldete deshalb immer grün.

### `libkvm.so`-Stub

Das Backend linkt per cgo gegen `server/dl_lib/libkvm.so`, eine
riscv64-Bibliothek. Auf dem CI-Host (x86_64/arm64) kann der Linker sie nicht
nutzen; die Pakete `service/ws` und `service/picoclaw` ließen sich nicht
einmal bauen. `ci/stub/kvm_stub.c` exportiert dieselben Symbole ohne
Funktion. Er wird nie ausgeliefert und testet nichts an der Videokette; er
macht nur den Rest testbar.

## Runner und Actions-Kontingent

Der Runner kommt aus der Repository-Variable `CI_RUNNER`:

| `CI_RUNNER` | Läuft auf |
|---|---|
| leer | `ubuntu-latest`, verbraucht Actions-Minuten |
| `self-hosted` | eigener Runner, z. B. die geplante Runner-VM auf dem Proxmox |

Der eigene Runner braucht Docker, weil die Gates im CI-Image laufen.

Ohne Runner und ohne Minuten lässt sich derselbe Lauf lokal oder in einer
Claude-Session machen:

```bash
./ci/gate.sh all
```

Das Ergebnis eines solchen Laufs kann als Commit-Status `gates/local` am PR
hängen; dann ist im PR sichtbar, auf welchem Stand er grün war.

## Upstream-Workflows

`package.yml`, `release.yml`, `builder-image.yml`, `create-tag.yml` und
`issue-tracker.yml` stammen von sipeed und bauen und veröffentlichen das
offizielle Release. Im Fork laufen sie ins Leere (kein Builder-Image unter
`ghcr.io/kaindlnetwork`, kein Zugriff auf `cdn.sipeed.com`) und kosten nur
Minuten. Sie bleiben unverändert im Repository, damit ein Upstream-Sync
keinen Konflikt erzeugt, und werden unter Actions deaktiviert
(„Disable workflow“).

## Upstream-Sync

```bash
./ci/upstream-sync.sh
```

holt `sipeed/NanoKVM` `main`, legt einen Branch `sync/upstream-<datum>` an
und merged upstream hinein. Konflikte entstehen fast nur in den wenigen
Upstream-Dateien, die die Fixes ändern; die Fix-Logik selbst steht in
eigenen Dateien. Danach `./ci/gate.sh all` und ein PR wie jeder andere.
