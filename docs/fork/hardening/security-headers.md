# Sicherheits-Header

## Problem

Weder UI noch API setzten Sicherheits-Header. Eine fremde Seite konnte die
NanoKVM-Oberfläche in einem unsichtbaren Rahmen einbetten und Klicks
darauf umlenken (Clickjacking). Das `SameSite=Strict`-Cookie verhindert in
aktuellen Browsern zwar, dass die eingebettete Seite angemeldet ist; bei
abgeschalteter Anmeldung oder älteren Browsern hilft das nicht.

## Fix

`middleware.SecurityHeaders()` in `router.web()`, vor jeder Route, also für
UI und API:

| Header | Wert | Wirkung |
|---|---|---|
| `X-Frame-Options` | `DENY` | kein Einbetten, ältere Browser |
| `Content-Security-Policy` | `frame-ancestors 'none'; base-uri 'self'; form-action 'self'; object-src 'none'` | kein Einbetten, keine fremde Basis-URL, keine Formulare nach außen, keine Plugins |
| `X-Content-Type-Options` | `nosniff` | Browser rät keinen Dateityp |
| `Referrer-Policy` | `no-referrer` | Adresse des Geräts geht nicht an verlinkte Seiten |

Bewusst nicht gesetzt: `script-src`, `connect-src` und Verwandte. Die
Konsole nutzt WebRTC, WebSockets und Inline-Styles; eine strengere Policy
braucht einen Test auf dem Gerät und kommt als eigener PR. Ebenso nicht
`Strict-Transport-Security`, weil viele Geräte mit selbstsigniertem
Zertifikat oder über HTTP laufen und HSTS sie aussperren könnte.

## Tests

`server/router/security_headers_test.go`.
