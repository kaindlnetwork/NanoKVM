## Beschreibung der Änderung
<!-- Was wurde geändert und warum? Bei einem Sicherheitsfix: der Angriff in einem Satz. -->


## Art der Änderung
- [ ] Sicherheitsfix
- [ ] Sichere Voreinstellung
- [ ] Bugfix
- [ ] CI/CD
- [ ] Upstream-Sync
- [ ] Dokumentation

## Betroffene Teile
- [ ] Backend (`server/`)
- [ ] Web-UI (`web/`)
- [ ] Gerät (`kvmapp/`, Init-Skripte, Kernel-Module)
- [ ] CI (`ci/`, Workflows)

## TDD-Nachweis
<!-- Ausgabe des Tests OHNE den Fix (rot) und mit dem Fix (grün). -->

```text
RED:

GREEN:
```

## Risikobewertung (Selbsteinschätzung)
- [ ] **Hoch**: ändert Login, Rechte, Updates oder was als root läuft
- [ ] **Mittel**: ändert Verhalten, das Nutzer bemerken
- [ ] **Niedrig**: Tests, Doku, CI

## Rollback-Plan
<!-- Revert dieses PRs genügt? Oder muss auf dem Gerät etwas zurückgesetzt werden? -->


## Getestet?
- [ ] `./ci/gate.sh all` grün
- [ ] Auf einem NanoKVM geprüft (Modell und Version angeben)
- [ ] Nicht auf Hardware nötig (nur Backend-Logik)

---
*Template nach dem Muster von kaindlnetwork/scaling-adventure: Change Management und Risikobewertung (ISO 27001), dokumentierter Änderungsprozess (ISO 9001).*
