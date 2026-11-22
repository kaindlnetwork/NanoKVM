# Werkspasswort serverseitig erzwingen

## Angriff

Ein neues oder zurückgesetztes Gerät meldet sich mit `admin`/`admin` an.
Upstream zeigte danach nur einen Hinweis in der Web-UI, der sich wegklicken
ließ (`setSkipModifyPassword`). Die API, das Web-Terminal (root-Shell) und
die Fernsteuerung des Zielrechners waren mit dem Werkspasswort voll nutzbar.
Wer ein frisch angeschlossenes Gerät im Netz findet, hat damit sofort root.

## Fix

Solange ein Konto `mustChangePassword` trägt, beantwortet der Server jede
Anfrage außer den folgenden mit HTTP 428 (Precondition Required):

- `GET /api/auth/password`, `POST /api/auth/password`
- `GET /api/auth/account`
- `POST /api/auth/logout`
- `GET /api/vm/web-title`

Die Prüfung sitzt in `CheckToken` und `CheckTokenOrLoopbackInternalToken`,
also vor jeder Rechteprüfung und jedem Handler. Die Web-UI leitet bei 428
auf die Passwortseite um und geht nach einem Login mit Werkspasswort direkt
dorthin.

428 statt 403, weil die UI auf 403 die Rechte neu lädt und auf 401 abmeldet;
428 ist in dieser API sonst unbenutzt.

## Tests

`server/middleware/password_change_test.go`. Die Umleitung in der Web-UI
hat keinen automatischen Test (das Web-Projekt hat noch kein
Test-Framework) und steht deshalb in der Geräte-Prüfliste des PRs.
