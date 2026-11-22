# Kernel: was der NanoKVM wirklich braucht

Ziel dieser Datei ist ein eigener, reduzierter Kernel für den NanoKVM. Sie
hält fest, welche Kernel-Funktionen die NanoKVM-Software nachweislich nutzt,
welche das offizielle Image zusätzlich mitbringt und welche Härtung fehlt.
Die maschinenlesbare Fassung steht in `ci/kernel/`:

| Datei | Inhalt |
|---|---|
| `ci/kernel/required.config` | Pflicht. Fehlt eine Option, fällt eine Kernfunktion aus. |
| `ci/kernel/remove.config` | Entbehrlich. Jede Option mit Beleg, warum nichts sie nutzt. |
| `ci/kernel/harden.config` | Härtung, die der Sipeed-Kernel nicht setzt. |
| `ci/kconfig_check.py` | Prüft eine `.config` gegen die drei Listen. |
| `ci/kernel/resolve-sipeed-config.sh` | Löst Sipeeds defconfig auf und prüft sie. |

Die drei Listen sind gewöhnliche Kconfig-Fragmente. Beim Kernel-Build lassen
sie sich direkt einmischen:

```bash
scripts/kconfig/merge_config.sh -m arch/riscv/configs/<board>_defconfig \
    ci/kernel/required.config ci/kernel/remove.config ci/kernel/harden.config
make olddefconfig
ci/kconfig_check.py .config --required ci/kernel/required.config \
    --remove ci/kernel/remove.config --harden ci/kernel/harden.config --strict
```

`--strict` macht übrig gebliebene oder fehlende Härtung zum Fehler. Ohne
`--strict` scheitert die Prüfung nur, wenn eine Pflicht-Option fehlt.

## Quelle der Kernelkonfiguration

Das offizielle NanoKVM-Image ist kein eigenes Kernel-Projekt. Es ist das
allgemeine LicheeRV-Nano-Image aus
[sipeed/LicheeRV-Nano-Build](https://github.com/sipeed/LicheeRV-Nano-Build),
Branch `NanoKVM`, Board `sg2002_licheervnano_sd`, in das
`kvm/merge_nanokvm_app.sh` das App-Paket kopiert. Der Kernel ist Linux 5.10
(`linux_5.10/`), konfiguriert über
`build/boards/sg200x/sg2002_licheervnano_sd/linux/sg2002_licheervnano_sd_defconfig`.

Analysiert wurde der Stand `064abf7` vom 18.11.2026. Das defconfig nennt
nur Abweichungen vom Kconfig-Standard; aufgelöst gegen den Kconfig-Baum von
`linux_5.10` ergeben sich 1022 fest eingebaute Optionen und 27 Module.
Darunter sind Bluetooth, Audio, Display, Touch, USB-Kameras, NAND- und
NFS-Unterstützung und fast der komplette Netfilter-Baukasten. Der Kernel ist
für ein Entwicklerboard konfiguriert, nicht für ein KVM. `kvm_system` räumt
das zum Teil zur Laufzeit wieder ab (es löscht `S40bluetoothd`,
`S04backlight`, `S05tp` und entlädt `soph_saradc`); die Treiber bleiben
trotzdem im Kernel.

## Methode: was nutzt die Software?

Gesucht wurde in `server/`, `kvmapp/` und `support/sg2002/kvm_system/` nach
allem, was den Kernel anspricht:

- **USB-Gadget**: Funktionen in configfs (`functions/<typ>.<name>`): nur
  `hid` (Tastatur, Maus), `mass_storage` (virtueller Datenträger), `rndis`
  und `ncm` (virtuelles Netz zum Zielrechner, abschaltbar).
- **Gerätedateien und sysfs**: `/dev/hidg0..2`, `/dev/loop-control`,
  `/dev/mmcblk0p3`, `/dev/input/event0`, `/sys/class/gpio/*`,
  `/sys/class/udc`, `/sys/kernel/config/usb_gadget`.
- **Dateisysteme**: ext4 (Root), vfat (`/boot`), exfat (Datenpartition),
  Loop-Geräte für Images.
- **Busse**: I2C für OLED und den HDMI-Wandler LT6911 (`kvm_system`,
  I2C-Bus 4), `i2c-gpio` als Modul (`S15kvmhwd`).
- **Netz**: Ethernet (stmmac), WLAN je nach Modell, `udhcpc` (AF_PACKET),
  iptables mit `-m state`, Tailscale über TUN.
- **Videokette**: `server/dl_lib/libkvm_mmf.so` ruft laut ihren
  undefinierten Symbolen nur `SYS`, `VB`, `MIPI`, `VI`, `VPSS`, `VENC`,
  `ISP`, `AE`, `AWB`, `BIN` auf. Die Treiber dafür sind die
  Herstellermodule `soph_*` (nicht Teil der Kernelkonfiguration, sondern
  aus `osdrv/`); sie brauchen ION, CMA und dma-buf im Kernel.
- **Audio**: nirgends. Kein Audio-Code in Server, Web-UI, `libkvm*` oder
  `kvm_system`.

## Ergebnis gegen den Sipeed-Kernel

```text
FEHLT (Pflicht): 0
NOCH AN (entbehrlich): 59
NOCH AUS (Haertung): 7
```

Alle Pflicht-Optionen sind gesetzt; die Liste stimmt also mit dem, was heute
läuft, überein. Die 59 entbehrlichen Optionen gliedern sich so:

| Gruppe | Beispiele | Warum entbehrlich |
|---|---|---|
| Bluetooth | `BT`, `BT_RFCOMM`, `BT_HIDP` | keine Nutzung, Dienst wird gelöscht |
| Audio | `SND`, `SND_SOC`, `USB_CONFIGFS_F_UAC1` | kein Audio-Code; ohne Treiber kein Mikrofonzugriff (Upstream-Issue #693) |
| Display, Touch | `FB`, `BACKLIGHT_CLASS_DEVICE`, `INPUT_TOUCHSCREEN` | kein Panel |
| Kameras | `USB_VIDEO_CLASS`, `USB_CONFIGFS_F_UVC` | Video kommt über MIPI-CSI vom LT6911 |
| USB-Host-Klassen | `USB_STORAGE`, `USB_SERIAL`, `USB_ACM` | der Port läuft als Gadget |
| Gadget-Funktionen | `ACM`, `ECM`, `EEM`, `SERIAL`, `F_FS` | nur hid, mass_storage, rndis, ncm in Gebrauch |
| Flash, NFS | `MTD`, `UBIFS_FS`, `JFFS2_FS`, `NFS_FS` | Boot von SD/eMMC |
| Tunnel | `WIREGUARD`, `IPV6_TUNNEL`, `VETH` | Tailscale nutzt TUN |
| Netfilter-Helfer | `NF_CONNTRACK_FTP`, `_SIP`, `_H323`, `XT_MATCH_BPF` | parsen fremde Protokolle im Kernel, ein KVM routet nichts |
| Busse für Userspace | `SPI_SPIDEV`, `I2C_SLAVE`, `INPUT_UINPUT` | nichts nutzt sie; uinput erlaubt Eingaben zu injizieren |
| Kernel-Oberfläche | `MAGIC_SYSRQ`, `USER_NS`, `IO_URING`, `UEVENT_HELPER`, `MODULE_FORCE_LOAD`, `THERMAL_WRITABLE_TRIPS` | kein Nutzen auf dem Gerät; User-Namespaces und io_uring sind häufige Ausgangspunkte für Rechteausweitung, beschreibbare Temperaturgrenzen erlauben, den Überhitzungsschutz abzuschalten |

Die vollständige Liste mit Begründung je Block steht in
`ci/kernel/remove.config`. `TOUCHSCREEN_HYNITRON_TS` und
`VIDEO_CVITEK_MIPI_TX` sind Herstellersymbole: im defconfig gesetzt, aber
nicht in den Kconfig-Dateien von `linux_5.10` definiert.

Fehlende Härtung (`ci/kernel/harden.config`): `STRICT_KERNEL_RWX`,
`HARDENED_USERCOPY`, `SLAB_FREELIST_HARDENED`, `SLAB_FREELIST_RANDOM`,
`SHUFFLE_PAGE_ALLOCATOR`, `INIT_ON_ALLOC_DEFAULT_ON`,
`SECURITY_DMESG_RESTRICT`. Vorhanden sind bereits `SECCOMP_FILTER` und
`STACKPROTECTOR_STRONG`; seccomp ist damit für die Prozess-Härtung des
Servers nutzbar. Auch `NF_TABLES` ist eingebaut, ein nftables-Regelsatz
braucht also keinen anderen Kernel.

## /dev/mem: der größte Einzelbefund

`S15kvmhwd` setzt 39-mal Pinmux-Register per `devmem` (z. B.
`devmem 0x030010D0 32 0x2  # I2C1_SCL`), `kvm_system` dreimal. Dafür
braucht der Kernel `CONFIG_DEVMEM`, und auf riscv gibt es in 5.10 weder
`STRICT_DEVMEM` noch `IO_STRICT_DEVMEM`. `/dev/mem` erlaubt root damit
Lesen und Schreiben des gesamten physischen Speichers, auch des Kernels.
Wer einmal root hat, kann so den Kernel zur Laufzeit verändern, ohne ein
Modul zu laden.

`DEVMEM` steht deshalb vorerst in `required.config`. Der Weg heraus:
Pinmux im Device Tree beschreiben (der Treiber `CVITEK_PINCTRL_CV1835` ist
vorhanden), die `devmem`-Aufrufe entfernen, dann `DEVMEM` abschalten.

## Kernel-Module zur Laufzeit

Unabhängig von der Kernelkonfiguration lädt `S00kmod` die Herstellermodule
`soph_*`. Davon nutzt der KVM-Stack nachweislich nicht: `soph_wiegand`
(Kartenleser-Bus), `soph_tpu` (NPU), `soph_ive` (Bildanalyse), `soph_dwa`
(Dewarp), `soph_mon` (Perf-Monitor). Ungeklärt sind `soph_vo`
(Videoausgang), `soph_rgn` (Overlays) und `soph_pwm`; sie brauchen einen
Test je Modell. Details in `docs/fork/hardening/kernel-modules.md`.

Nach dem letzten Laden lässt sich das Nachladen ganz sperren:
`echo 1 > /proc/sys/kernel/modules_disabled`. `S15kvmhwd` lädt `i2c-gpio`
aber je nach erkannter Hardware nach; die Sperre gehört deshalb ans Ende des
Bootvorgangs, nach `S15kvmhwd`, und braucht einen Hardware-Test.

## Reihenfolge für einen eigenen Kernel

1. **Laufzeit, ohne neuen Kernel**: ungenutzte `soph_*`-Module nicht laden
   (PR #10), `modules_disabled` nach dem Boot, `debugfs` nicht mounten
   (`S01fs`), sofern kein Herstellermodul es braucht.
2. **Kernel neu bauen mit den Fragmenten**: erst `harden.config` und die
   eindeutigen Gruppen aus `remove.config` (Bluetooth, Audio, Display,
   Kameras, Flash, NFS, Netfilter-Helfer).
3. **Pinmux in den Device Tree**, dann `DEVMEM` aus.
4. **Optional**: Modul-Signaturen (`MODULE_SIG_FORCE`) mit eigenem
   Schlüssel für die `soph_*`-Module, sobald der Build sie signieren kann.

Jeder Schritt braucht einen Test auf dem Gerät (Video, HID, virtueller
Datenträger, Netz, WLAN-Modell, OLED). Maßgeblich ist danach
`/proc/config.gz` auf dem Gerät:

```bash
zcat /proc/config.gz > nanokvm.config
ci/kconfig_check.py nanokvm.config --required ci/kernel/required.config \
    --remove ci/kernel/remove.config --harden ci/kernel/harden.config
```

## Grenzen dieser Analyse

- Die Auflösung lief gegen den Host-Compiler. `cc-option`-Prüfungen im
  Kconfig (z. B. für den Stack-Protector) können mit der riscv-Toolchain
  anders ausfallen. `/proc/config.gz` vom Gerät ist maßgeblich.
- Welche Kernel-Symbole die Herstellermodule `soph_*` brauchen, steht nicht
  im Repository; nur ein Build plus `modprobe --dry-run` oder ein
  Gerätetest zeigt, ob eine gestrichene Option eines davon bricht. ION,
  CMA, dma-buf und Pinctrl stehen deshalb in den Pflicht-Optionen.
- WLAN-Treiber (`AIC_WLAN_SUPPORT`) und `CFG80211` sind modellabhängig und
  stehen in keiner Liste. Auf Modellen ohne WLAN sind sie entbehrlich.
