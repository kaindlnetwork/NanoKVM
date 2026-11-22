# Abhängigkeiten mit bekannten Schwachstellen

## Befund

Der erste Lauf der Pipeline (`govulncheck`, `pnpm audit`) auf upstream
`4f6f13d`:

- **Go:** 45 Schwachstellen, die der Code tatsächlich aufruft. Die meisten
  in der Standardbibliothek, weil mit Go 1.25.0 gebaut wird (Fixes ab
  1.25.2 bis 1.26.11, u. a. `net/http`, `crypto/tls`, `crypto/x509`,
  `net/url`). Dazu `golang.org/x/net` 0.47.0, `pion/dtls` 3.1.2 und
  `pion/stun` 3.0.0 (WebRTC).
- **Web:** 38 Funde, 16 davon „high“, fast alle in `axios`, `form-data`
  und `follow-redirects`.

Upstream baut die Firmware ebenfalls mit Go 1.25.0 (`docker/Dockerfile`).
Die ausgelieferten Binärdateien enthalten diese Lücken also auch.

## Änderung

| Was | Vorher | Nachher |
|---|---|---|
| Go-Toolchain (`toolchain` in `go.mod`, CI-Image, `docker/Dockerfile`) | 1.25.0 | 1.27.4 |
| `go`-Mindestversion in `go.mod` | 1.25.0 | 1.26.0 (verlangt von `x/net` 0.60) |
| `golang.org/x/net` | 0.47.0 | 0.60.0 |
| `github.com/pion/dtls/v3` | 3.1.2 | 3.1.4 |
| `github.com/pion/stun/v3` | 3.0.0 | 3.1.5 |
| `axios` | 1.15.2 | 1.21.1 |
| `react-router-dom` | 6.30.3 | 6.30.6 |
| staticcheck, gosec, govulncheck, actionlint im CI-Image | 2025.1.1, 2.22.8, 1.1.4, 1.7.7 | 2026.2.2, 2.30.0, 1.10.0, 1.7.12 |

Mitgezogen von `go get`: `x/crypto` 0.57.0, `x/sys` 0.48.0, `x/text`
0.42.0, `pion/transport` 4.0.2.

## Was bleibt

- `react-router` (moderat): der Fix existiert nur in 7.x. Die Lücke betrifft
  serverseitiges Rendering; die Web-UI nutzt nur `createHashRouter` im
  Browser. Ein Wechsel auf 7.x ist ein eigener PR.
- `@babel/core` (niedrig): nur zur Build-Zeit über `jotai`.

`web-audit` prüft ab „high“ und ist damit grün.

## Prüfung auf dem Gerät

Die riscv64-Firmware wird mit dem Builder-Image aus `docker/Dockerfile`
gebaut. Ob Go 1.27.4 dort mit der T-Head-Toolchain und cgo sauber baut und
auf dem SG2002 läuft, zeigt erst ein Build mit neu gebautem Builder-Image
und ein Test auf dem Gerät (Video, HID, WebRTC wegen der `pion`-Updates).
