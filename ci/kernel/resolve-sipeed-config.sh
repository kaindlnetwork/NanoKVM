#!/usr/bin/env bash
#
# Loest die Kernel-Konfiguration des offiziellen NanoKVM-Images auf und
# prueft sie gegen ci/kernel/*.config.
#
# Das NanoKVM-Image ist das LicheeRV-Nano-Image (Board sg2002_licheervnano_sd)
# aus sipeed/LicheeRV-Nano-Build, Branch NanoKVM, plus App-Paket. Dessen
# defconfig nennt nur Abweichungen; was dann tatsaechlich gebaut wird, ergibt
# erst "make <defconfig>" gegen den Kconfig-Baum von linux_5.10.
#
#   ci/kernel/resolve-sipeed-config.sh [arbeitsverzeichnis]
#
# Braucht git, make, gcc, flex, bison. Holt nur Kconfig-Dateien und
# scripts/ (Sparse-Checkout ohne Blobs), kein vollstaendiger Kernel.
#
# Einschraenkung: cc-option-Pruefungen in Kconfig laufen gegen den
# Host-Compiler, nicht gegen die riscv-Toolchain. Massgeblich bleibt
# /proc/config.gz auf dem Geraet (CONFIG_IKCONFIG_PROC ist an):
#   zcat /proc/config.gz > nanokvm.config
#   ci/kconfig_check.py nanokvm.config --required ... (siehe unten)

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORK="${1:-${TMPDIR:-/tmp}/nanokvm-kernel}"
SDK_URL="${SDK_URL:-https://github.com/sipeed/LicheeRV-Nano-Build}"
SDK_BRANCH="${SDK_BRANCH:-NanoKVM}"
BOARD=sg2002_licheervnano_sd
DEFCONFIG="build/boards/sg200x/$BOARD/linux/${BOARD}_defconfig"

mkdir -p "$WORK"
cd "$WORK"
if [ ! -d sdk/.git ]; then
    git clone -q --filter=blob:none --no-checkout --depth 1 -b "$SDK_BRANCH" "$SDK_URL" sdk
fi
cd sdk
git config core.sparseCheckout true
cat > .git/info/sparse-checkout <<EOF
/$DEFCONFIG
/linux_5.10/Makefile
/linux_5.10/Kconfig
/linux_5.10/**/Kconfig*
/linux_5.10/scripts/
/linux_5.10/arch/riscv/Makefile
/linux_5.10/arch/riscv/configs/
/linux_5.10/include/
/linux_5.10/usr/
EOF
git read-tree -mu HEAD
echo "SDK-Stand: $(git log -1 --format='%h %ci %s')"

cp "$DEFCONFIG" linux_5.10/arch/riscv/configs/nanokvm_resolve_defconfig
make -s -C linux_5.10 ARCH=riscv CROSS_COMPILE= nanokvm_resolve_defconfig 2>/dev/null
cp linux_5.10/.config "$WORK/sipeed-resolved.config"
echo "Aufgeloest: $WORK/sipeed-resolved.config"

python3 "$REPO_ROOT/ci/kconfig_check.py" "$WORK/sipeed-resolved.config" \
    --required "$REPO_ROOT/ci/kernel/required.config" \
    --remove "$REPO_ROOT/ci/kernel/remove.config" \
    --harden "$REPO_ROOT/ci/kernel/harden.config"
