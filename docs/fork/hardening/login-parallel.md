# Parallele Login-Versuche umgehen den Brute-Force-Schutz

## Angriff

Der Login prüfte zuerst, ob die IP gesperrt ist, dann das Passwort (bcrypt
und 2 Sekunden Wartezeit) und erst danach zählte er den Fehlversuch.
Schickt ein Angreifer viele Anfragen gleichzeitig, passieren alle die
Sperrprüfung, bevor der erste Fehlversuch gezählt ist. Bei
`loginMaxFailures: 5` kamen so beliebig viele Versuche statt fünf durch,
und die 2 Sekunden Wartezeit bremsten nur die einzelne Verbindung, nicht
den Angreifer.

Dazu zählte die Sperre je einzelner IPv6-Adresse. Ein Anschluss bekommt
meist ein ganzes /64-Netz und kann nach jeder Sperre die Adresse wechseln.

## Fix

- Login-Versuche derselben Quelle laufen nacheinander
  (`LockLoginAttempt`). Die Sperre greift damit nach genau
  `loginMaxFailures` Versuchen, und die Wartezeit summiert sich.
- IPv6-Quellen werden je /64 gezählt, IPv4 je Adresse.

Andere Quellen warten nicht aufeinander; ein Angreifer bremst nur sich
selbst (und Nutzer hinter derselben IPv4-Adresse oder demselben /64).

## Tests

`server/service/auth/login_parallel_test.go`.
