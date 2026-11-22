# X-Forwarded-Proto nur von konfigurierten Proxys

## Problem

Zwei Stellen glaubten jedem Client, der `X-Forwarded-Proto: https`
mitschickte:

- Das Session-Cookie bekam dann das Attribut `Secure`, obwohl die
  Verbindung unverschlüsselt war.
- Die Origin-Prüfung für WebSockets rechnete mit dem behaupteten Schema.

Ein Browser setzt diesen Header bei fremden Seiten nicht, ausnutzbar war
das also vor allem durch Werkzeuge außerhalb des Browsers. Der Header ist
trotzdem eine Behauptung des Clients und darf nur von einem Proxy kommen,
den der Betreiber kennt.

## Fix

Neue Einstellung in `/etc/kvm/server.yaml`:

```yaml
security:
  trustedProxies:
    - 192.0.2.10      # einzelne Adresse
    - 10.0.0.0/24     # oder ein Netz
```

Nur Anfragen, deren direkter Absender in dieser Liste steht, dürfen
`X-Forwarded-Proto` setzen. Leer (Standard): niemand.

Wer den NanoKVM hinter einem Reverse-Proxy mit TLS betreibt (z. B. vor
authentik), trägt dessen Adresse ein. Ohne Eintrag funktioniert der Zugang
weiter, nur das Cookie verliert hinter dem Proxy das `Secure`-Attribut.

Die eingebaute Proxy-Erkennung von gin (`ClientIP`) bleibt unverändert: der
Brute-Force-Schutz nutzt die direkte Gegenstelle (`RemoteIP`), `ClientIP`
erscheint nur in Debug-Logs.

## Tests

`server/service/auth/trusted_proxy_test.go`,
`server/middleware/trusted_proxy_test.go`, `server/middleware/proxy_test.go`.
