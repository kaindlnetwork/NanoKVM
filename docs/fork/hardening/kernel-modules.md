# Nicht benötigte Kernel-Module nicht laden

Die vollständige Analyse des Kernels (Konfiguration, Pflicht- und
Streichlisten, Härtung, `/dev/mem`) steht in
[`../kernel/README.md`](../kernel/README.md). Diese Datei behandelt nur die
Module, die `S00kmod` zur Laufzeit lädt.

Stand: Draft. Braucht einen Test auf echter Hardware, bevor es gemerged wird.

## Ausgangslage

Der Kernel und alle Module kommen aus dem Betriebssystem-Image von sipeed
(Buildroot), nicht aus diesem Repository. Beeinflussbar ist nur, welche
Module `S00kmod` beim Start mit `insmod` lädt. Upstream lädt 22
SoC-Module des SG2002, darunter Treiber für Hardware, die ein NanoKVM nicht
hat oder nicht nutzt.

Jedes geladene Modul ist Kernel-Code mit Gerätedatei und ioctl-Schnittstelle.
Was nicht geladen ist, kann keine Lücke haben, die jemand mit root-Zugang
zum Kernel-Ausbruch oder für Persistenz nutzt.

## Was der KVM-Stack braucht

Die Videokette (`server/dl_lib/libkvm_mmf.so`) ruft nur diese Teile des
Media-Frameworks auf (aus den undefinierten Symbolen der Bibliothek):
`SYS`, `VB`, `MIPI`, `VI`, `VPSS`, `VENC`, `ISP`, `AE`, `AWB`, `BIN`. Dazu
gehören `soph_sys`, `soph_base`, `soph_mipi_rx`, `soph_snsr_i2c`, `soph_vi`,
`soph_vpss`, `soph_vcodec`, `soph_jpeg`, `soph_vc_driver`.

## Stufe 1: in diesem PR nicht mehr geladen

| Modul | Zweck | Beleg, dass es nicht gebraucht wird |
|---|---|---|
| `soph_wiegand` | Wiegand-Bus für Zutrittskartenleser | keine solche Hardware |
| `soph_tpu` | NPU | kein `TPU`/`CVI_NN`-Aufruf in `libkvm*`, `kvm_system`, Backend |
| `soph_ive` | Bildanalyse-Engine | kein `IVE`-Aufruf |
| `soph_dwa` | Entzerrung (Dewarp) | kein `DWA`/`GDC`-Aufruf |
| `soph_mon` | Performance-Monitor | nur Diagnose |

## Stufe 2: geprüft, noch geladen

| Modul | Offene Frage |
|---|---|
| `soph_vo` | Videoausgang. Kein `VO`-Aufruf in `libkvm*`, aber vielleicht von `vpss` oder vom HDMI-Loop-Out einzelner Modelle genutzt. |
| `soph_rgn` | Overlays. Kein `RGN`-Aufruf, `vpss` kann aber intern darauf zugreifen. |
| `soph_pwm` | PWM. Möglicherweise Lüfter oder Hintergrundbeleuchtung je Modell. |

Diese kommen erst nach einem Test je Modell weg.

## Wichtig: ein Update ändert das allein nicht

`/kvmapp/system/init.d/S00kmod` ist die Kopie im App-Paket. Beim Booten
läuft `/etc/init.d/S00kmod` aus dem Image. Das OTA-Update kopiert keine
Init-Skripte nach `/etc/init.d` (nur `S95nanokvm` wird aus `/kvmapp`
gestartet). Wirksam wird die Änderung erst, wenn die Datei auf dem Gerät
nach `/etc/init.d/S00kmod` kopiert wird. Ein automatischer Weg dafür ist ein
eigener PR.

## Prüfliste auf dem Gerät

1. `lsmod` vor der Änderung sichern.
2. `cp /kvmapp/system/init.d/S00kmod /etc/init.d/S00kmod`, Neustart.
3. `lsmod`: die fünf Module fehlen, der Rest ist da.
4. Video: MJPEG und H.264 im Browser, Auflösungswechsel am Zielrechner,
   1080p60 falls unterstützt.
5. HID: Tastatur, Maus absolut und relativ, Einfügen.
6. Virtueller Datenträger einhängen, OLED-Anzeige, Power-/Reset-Taster.
7. `dmesg` ohne neue Fehler zu `soph_*`.
8. Modell und Image-Version im PR notieren.

Zurück: Sicherung von `/etc/init.d/S00kmod` einspielen oder die Zeilen
wieder eintragen.
