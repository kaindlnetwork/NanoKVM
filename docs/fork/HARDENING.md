# Härtung: Übersicht

Jeder Fix hat eine eigene Datei unter `docs/fork/hardening/` und einen
eigenen Pull Request. Die Datei kommt mit dem PR; diese Übersicht ändert
sich nur, wenn ein Thema dazukommt. So bekommen parallele Fix-PRs keinen
Merge-Konflikt in der Doku.

Grundlage ist die Sicherheitsprüfung vom 22.11.2026 auf upstream
`4f6f13d` (Version 2.6.x).

## Backend und Web-UI (TDD, ohne Hardware prüfbar)

| Thema | Datei | Risiko vorher |
|---|---|---|
| Command-Injection und Path-Traversal bei Skripten | [script-injection](hardening/script-injection.md) | Hoch, für Konten mit Terminal-Recht |
| Werkspasswort serverseitig erzwingen | [factory-password](hardening/factory-password.md) | Hoch |
| Parallele Login-Versuche umgehen den Brute-Force-Schutz | [login-parallel](hardening/login-parallel.md) | Mittel |
| Sichere Standardwerte: Sperre an, Session 1 Tag | [secure-defaults](hardening/secure-defaults.md) | Mittel |
| `X-Forwarded-Proto` nur von konfigurierten Proxys | [trusted-proxy](hardening/trusted-proxy.md) | Niedrig |
| Kein `Access-Control-Allow-Origin: *` bei abgeschalteter Anmeldung | [no-wildcard-cors](hardening/no-wildcard-cors.md) | Hoch, wenn Anmeldung aus |
| Sicherheits-Header (Framing, Sniffing, Referrer) | [security-headers](hardening/security-headers.md) | Niedrig |
| Eigener Update-Server nur über HTTPS | [update-https](hardening/update-https.md) | Hoch, wenn eigener Server gesetzt |

## Gerät (Draft, Prüfung auf Hardware nötig)

| Thema | Datei | Stand |
|---|---|---|
| Nicht benötigte Kernel-Module nicht laden | [kernel-modules](hardening/kernel-modules.md) | Draft |
| Kernel-Bedarf, Streich- und Härtungsliste | [kernel/README](kernel/README.md) | Analyse, Grundlage für eigenen Kernel |

## Bekannt und offen

Nicht ohne größeren Umbau oder nicht im Repository lösbar:

- **Alles läuft als root.** Das Backend bräuchte einen eigenen Benutzer und
  gezielte Rechte für GPIO, HID-Gadget und `/etc/kvm`.
- **Web-Admin-Passwort ist das root-Passwort.** `ChangePassword` ruft
  `passwd root` auf. Eine Trennung braucht eine Migration für bestehende
  Geräte.
- **SSH ist standardmäßig an.** Umstellen ändert das Verhalten bestehender
  Geräte beim Update.
- **Updates sind nicht signiert.** Die SHA-512 kommt vom selben Server wie
  das Paket. Abhilfe wäre eine Signatur mit einem Schlüssel, der nicht auf
  dem CDN liegt.
- **Betriebssystem-Image und Kernel** liegen nicht in diesem Repository
  (Buildroot von sipeed) und bekommen über diesen Fork keine Updates.
- **Passwort-„Verschlüsselung“ im Browser** mit festem Schlüssel
  (`nanokvm-sipeed-2024`) ist nur Verschleierung. Schutz bietet allein HTTPS.
