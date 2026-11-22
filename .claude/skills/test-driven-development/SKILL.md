---
name: test-driven-development
description: Strikte Test-Driven-Development-Schleife (Red, Green, Refactor) für den gehärteten NanoKVM-Fork, Go-Backend unter server/ und Web-UI unter web/. Triggern bei jedem Sicherheitsfix, Bugfix, Refactoring oder jeder Verhaltensänderung. Angepasst aus kaindlnetwork/scaling-adventure, Quelle https://github.com/obra/superpowers/tree/main/skills/test-driven-development
---

# Test-Driven Development (TDD)

## Overview

Write the test first. Watch it fail. Write minimal code to pass.

**Core principle:** If you didn't watch the test fail, you don't know if it
tests the right thing.

## When to Use

**Always:**

- Sicherheitsfixes (jeder Fix in `docs/fork/HARDENING.md`)
- Bugfixes
- Refactorings
- Verhaltensänderungen am Backend oder an der Web-UI

**Exceptions (ask your human partner):**

- Reine Konfiguration (`ci/Dockerfile`, Workflows)
- Änderungen, die nur auf echter Hardware prüfbar sind (Init-Skripte,
  Kernel-Module). Die werden Draft-PRs mit Geräte-Prüfliste.

## The Iron Law

```text
NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST
```

## Red-Green-Refactor

### RED: Write Failing Test

Ein minimaler Test, der ein Verhalten zeigt. Ablage neben dem Code
(`server/<paket>/<thema>_test.go`). Für einen Sicherheitsfix stellt der Test
den Angriff nach: der Shell-Befehl im Dateinamen, das `../` im Pfad, die
parallelen Login-Versuche.

### Verify RED: Watch It Fail

**MANDATORY. Never skip.**

```bash
./ci/gate.sh go-test
# oder gezielt:
cd server && go test ./service/vm -run RunScript -count=1
```

Für `go test` direkt braucht es den Ersatz für `libkvm.so`; `ci/gate.sh`
baut ihn selbst. Von Hand:

```bash
cc -shared -fPIC -I server/include -o /tmp/kvmstub/libkvm.so ci/stub/kvm_stub.c
export CGO_LDFLAGS=-L/tmp/kvmstub LD_LIBRARY_PATH=/tmp/kvmstub
```

Prüfen:

- Der Test schlägt fehl, und zwar mit der Meldung, die den Angriff
  beschreibt („injected command was executed“), nicht mit einem
  Kompilierfehler, den du nicht erwartet hast.
- Er scheitert an der fehlenden Absicherung, nicht an einem Tippfehler.

**Test passes?** Du testest bestehendes Verhalten. Test korrigieren.

Die Ausgabe des roten Laufs gehört in den PR-Text. Sie ist der Beleg, dass
der Test den Angriff wirklich erkennt.

### GREEN: Minimal Code

Der einfachste Code, der den Test grün macht. Neue Logik in neue Dateien,
im Upstream-Bestand nur der Aufruf (siehe CLAUDE.md, Upstream-Verträglichkeit).

### Verify GREEN: Watch It Pass

```bash
./ci/gate.sh all
```

Alle Gates grün. **Test rot?** Code korrigieren, nicht den Test.

### REFACTOR: Clean Up

Erst wenn grün. Kein neues Verhalten.

## Pflicht-Tests je Sicherheitsfix

1. **Der Angriff scheitert** (der eigentliche RED-Test).
2. **Der legitime Weg funktioniert weiter.** Ein Fix, der Skripte gar nicht
   mehr ausführt, ist auch „sicher“. Der Test, dass `hello.sh` weiter läuft,
   verhindert das.
3. **Die Grenze stimmt.** Vertrauenswürdiger Proxy ja, beliebiger Client
   nein; Passwortseite ja, Terminal nein.
4. **Alte Konfiguration lädt weiter**, wenn der Fix eine Voreinstellung
   ändert.

## Good Tests

| Quality | Good | Bad |
|---|---|---|
| **Minimal** | Ein Verhalten. | `TestScriptAndAutostartAndUpload` |
| **Clear** | Name beschreibt das Verhalten | `TestScript1` |
| **Shows intent** | Der Name sagt, was nicht passieren darf | `TestRunScript` |

`TestRunScriptDoesNotPassNameToShell` statt `TestRunScriptWorks`.

## Common Rationalizations

| Excuse | Reality |
|---|---|
| "Das ist nur eine Validierung" | Genau die fehlte, sonst gäbe es den Fix nicht. |
| "Ohne Gerät kann ich nichts testen" | Der Stub für `libkvm.so` und temporäre Verzeichnisse decken Handler, Auth und Konfiguration ab. |
| "Upstream hat auch keine Tests" | Du änderst es gerade. |
| "Ich teste danach" | Tests, die sofort grün sind, beweisen nichts. |

## Red Flags: STOP and Start Over

- Code vor Test
- Test ist sofort grün
- Du kannst nicht erklären, warum der Test rot war
- Der rote Lauf fehlt im PR-Text

## Verification Checklist

- [ ] Jeder Fix hat einen Test, der den Angriff nachstellt
- [ ] Jeden Test rot gesehen, Ausgabe im PR-Text
- [ ] Legitimer Weg weiter getestet
- [ ] `./ci/gate.sh all` vollständig grün
- [ ] Beschreibung in `docs/fork/hardening/<thema>.md`

## Final Rule

```text
Production code → test exists and failed first
Otherwise → not TDD
```
