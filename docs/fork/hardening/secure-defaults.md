# Sichere Standardwerte: Brute-Force-Sperre an, Session 1 Tag

## Problem

- `security.loginLockoutDuration` war standardmäßig `0`, und `0` bedeutete
  „Schutz aus“. Jede von upstream erzeugte `/etc/kvm/server.yaml` enthält
  diese `0` ausdrücklich. Passwörter ließen sich damit unbegrenzt raten,
  gebremst nur durch 2 Sekunden je Verbindung.
- Eine Anmeldung galt 31 Tage (`refreshTokenDuration: 2678400`). Ein
  gestohlenes Cookie war einen Monat lang eine Fernsteuerung mit root.

## Fix

- `loginLockoutDuration: 0` (oder fehlend) heißt jetzt 300 Sekunden Sperre
  nach `loginMaxFailures` (Standard 5) Fehlversuchen. Abschalten geht nur
  noch ausdrücklich mit einem negativen Wert, z. B. `-1`.
- Neue Installationen schreiben `refreshTokenDuration: 86400` (1 Tag).

## Bestehende Geräte

Die Sperre greift nach dem Update sofort, weil die vorhandene `0` jetzt den
sicheren Standard bedeutet.

Die Session-Dauer ändert sich auf bestehenden Geräten nicht: deren
`server.yaml` enthält `2678400` ausdrücklich, und ein ausdrücklich
gesetzter Wert wird nicht überschrieben (der Fork kann ihn nicht von einer
bewussten Entscheidung unterscheiden). Empfehlung: von Hand auf `86400`
setzen.

Die Doku in `server/README.md` ist angepasst; die chinesische und
japanische Fassung nicht (Upstream-Übersetzungen, siehe CLAUDE.md).

## Tests

`server/config/default_security_test.go`.
