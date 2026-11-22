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

Die Signaturfrage (Prüfsumme vom selben Server wie das Paket) löst das
nicht; sie steht unter „Bekannt und offen“ in `../HARDENING.md`.

## Tests

`server/service/application/update_server_security_test.go`.
