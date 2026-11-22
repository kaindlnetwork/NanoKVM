# Eigener Update-Server nur über HTTPS

## Angriff

Unter „Update-Server“ ließ sich eine eigene Quelle für Updates eintragen,
auch mit `http://`. Das Update-Paket läuft als root, und seine SHA-512
steht in der `latest.json` vom selben Server. Wer den Weg zwischen Gerät
und Server mitlesen und verändern kann (WLAN, Router, DNS), tauscht beides
aus und bekommt beim nächsten Update root auf dem Gerät.

## Fix

Ein eigener Update-Server muss `https://` sein, beim Speichern und beim
Laden. Eine vor dem Update gespeicherte `http://`-Adresse wird nicht mehr
benutzt: der Update-Versuch bricht mit Fehler ab, bis eine
`https://`-Adresse gespeichert oder der eigene Server abgeschaltet ist.
Bewusst kein stiller Rückfall auf `cdn.sipeed.com`, denn wer einen eigenen
Server einträgt, will dessen Pakete und nicht die von upstream.

Ist der eigene Server deaktiviert, wird eine alte `http://`-Adresse beim
Laden still durch die Standardquelle ersetzt; sie wird ohnehin nie
angefragt und soll Updates von dort nicht blockieren.

Weiterleitungen von `https://` auf `http://` lehnt der Update-Client immer
ab (`utils/http.go`), auch für `cdn.sipeed.com`. Upstream tat das nur, wenn
Zugangsdaten in der URL standen; ohne sie folgte er dem Downgrade, und die
HTTPS-Pflicht ließ sich über einen Redirect umgehen. Gefunden im
Codex-Review zu PR #9. Weiterleitungen innerhalb von HTTPS (CDNs) gehen
weiter.

Die Signaturfrage (Prüfsumme vom selben Server wie das Paket) löst das
nicht; sie steht unter „Bekannt und offen“ in `../HARDENING.md`.

## Tests

`server/service/application/update_server_security_test.go`,
`server/utils/http_redirect_security_test.go`.
