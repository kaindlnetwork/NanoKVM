# Command-Injection und Path-Traversal bei Skripten

## Angriff

`POST /api/vm/script/run` setzte den Skriptnamen ungeprüft in
`sh -c "/etc/kvm/scripts/<name>"`. Ein Name wie `ok.sh; reboot #.sh` führte
damit beliebige Befehle als root aus. `../` im Namen erreichte Skripte
außerhalb des Verzeichnisses; `DELETE /api/vm/script` löschte mit `../`
beliebige Dateien. Die Autostart-Endpunkte bauten Pfade auf dieselbe Weise.

Betroffen sind Konten mit dem Recht `terminal`. Die haben über das
Web-Terminal ohnehin eine root-Shell; der Fix schließt also keine
Rechteausweitung, aber einen Weg, der über Skript-Automatisierung,
gestohlene Sessions oder künftige feinere Rechte ausnutzbar wäre.

## Fix

- Namen sind genau ein Pfadelement (`utils.SafeFileName`): kein `/`, kein
  `..`, keine Steuerzeichen, kein führendes `-`.
- Lesen, Schreiben und Löschen laufen über `os.Root`
  (`utils/confined.go`). Damit führen auch Symlinks nicht aus dem
  Verzeichnis heraus.
- Ausgeführt wird nur eine reguläre Datei direkt im Skriptverzeichnis
  (`utils.ConfinedRegularFile`). Ein Symlink würde sonst außerhalb von
  `os.Root` von `exec` bzw. `python` erneut aufgelöst und könnte überallhin
  zeigen. Uploads schreiben über `os.Root`, also nie durch einen
  vorhandenen Symlink hindurch; sie sind auf 1 MiB begrenzt, weil sie dafür
  in den Speicher gelesen werden. Beides aus dem Codex-Review zu PR #2.
- Skripte werden ohne Shell gestartet: `python <pfad>` für `.py`, direkt bei
  `#!`, sonst `sh <pfad>` (dasselbe Verhalten wie vorher die Shell).
- Skript- und Autostart-Verzeichnis werden mit `0700` angelegt,
  Autostart-Einträge mit `0700` statt `0755`.

## Tests

`server/service/vm/script_security_test.go`, `server/utils/safename_test.go`,
`server/utils/confined_test.go`.
