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

Kein CORS-Middleware mehr, auch nicht ohne Anmeldung. Die eigene Web-UI
läuft vom selben Ursprung und braucht keine CORS-Header. Beim Start ohne
Anmeldung schreibt der Server eine Warnung ins Log.

Die Abhängigkeit `github.com/rs/cors` entfällt.

## Tests

`server/main_security_test.go`. `newEngine` ist aus `run()` herausgelöst,
damit die globale Middleware testbar ist; der Umbau allein ändert kein
Verhalten (der rote Lauf im PR ist gegen genau diesen Stand).
