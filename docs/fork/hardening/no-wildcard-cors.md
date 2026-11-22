# Kein `Access-Control-Allow-Origin: *` bei abgeschalteter Anmeldung

## Angriff

Mit `authentication: disable` schaltete upstream zusätzlich
`cors.AllowAll()` ein. Jede Antwort erlaubte damit jeder fremden Website
den Zugriff. Öffnet jemand im selben Netz eine präparierte Seite, kann deren
JavaScript über seinen Browser die API des NanoKVM aufrufen und die
Antworten lesen: Tastatureingaben an den Zielrechner, Power-Taster,
Skripte, Einstellungen.

Ohne Anmeldung ist das Gerät für jeden im Netz offen; das ist bekannt und
gewollt. Neu durch CORS `*` war, dass die Angreiferin gar nicht im Netz
sein muss, sondern nur eine Website, die jemand im Netz besucht.

## Fix

Zwei Teile, weil CORS allein nur das Lesen der Antwort regelt:

**Formular-POST von fremden Seiten.** Ein normales HTML-Formular braucht
keinen CORS-Preflight. Eine fremde Seite konnte deshalb auch ohne
CORS-Header `POST /api/vm/gpio` mit `application/x-www-form-urlencoded`
abschicken; bei abgeschalteter Anmeldung lief die Aktion (Power, Reset),
nur die Antwort blieb unsichtbar. Gefunden im Codex-Review zu PR #7.
Jetzt lehnt `middleware.CrossOriginProtection` (Go
`http.CrossOriginProtection`) jede zustandsändernde Anfrage (alles außer
GET, HEAD, OPTIONS) ab, die laut `Sec-Fetch-Site` oder `Origin` von einer
anderen Origin kommt. Das gilt mit und ohne Anmeldung. Anfragen ohne diese
Header (curl, Skripte) sind keine Browser-Anfragen und laufen weiter.

**CORS.** Kein CORS-Middleware mehr, auch nicht ohne Anmeldung. Die eigene Web-UI
läuft vom selben Ursprung und braucht keine CORS-Header. Beim Start ohne
Anmeldung schreibt der Server eine Warnung ins Log.

Die Abhängigkeit `github.com/rs/cors` entfällt.

## Tests

`server/main_security_test.go`: CORS-Header, Formular-POST mit
`Sec-Fetch-Site: cross-site` und mit fremdem `Origin`, Same-Origin-POST der
eigenen UI. `newEngine` ist aus `run()` herausgelöst,
damit die globale Middleware testbar ist; der Umbau allein ändert kein
Verhalten (der rote Lauf im PR ist gegen genau diesen Stand).
