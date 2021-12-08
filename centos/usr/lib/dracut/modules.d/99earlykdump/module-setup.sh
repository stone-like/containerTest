#!/bin/bash

. /etc/sysconfig/kdump
. /lib/kdump/kdump-lib.sh

KDUMP_KERNEL=""
KDUMP_INITRD=""

check() {
    if [ ! -f /etc/sysconfig/kdump ] || [ ! -f /lib/kdump/kdump-lib.sh ]\
        || [ -n "${IN_KDUMP}" ]
    then
        return 1
    fi
    return 255
}

depends() {
    echo "base shutdown"
    return 0
}

prepare_kernel_initrd() {
    KDUMP_BOOTDIR=$(check_boot_dir "${KDUMP_BOOTDIR}")
    if [ -z "$KDUMP_KERNELVER" ]; then
        kdump_kver=`uname -r`
    else
        kdump_kver=$KDUMP_KERNELVER
    fi
    KDUMP_KERNEL="${KDUMP_BOOTDIR}/${KDUMP_IMG}-${kdump_kver}${KDUMP_IMG_EXT}"
    KDUMP_INITRD="${KDUMP_BOOTDIR}/initramfs-${kdump_kver}kdump.img"
}

install() {
    inst_multiple tail find cut dirname hexdump
    inst_simple "/etc/sysconfig/kdump"
    inst_binary "/usr/sbin/kexec"
    inst_binary "/usr/bin/gawk" "/usr/bin/awk"
    inst_script "/lib/kdump/kdump-lib.sh" "/lib/kdump-lib.sh"
    inst_hook cmdline 00 "$moddir/early-kdump.sh"
    prepare_kernel_initrd
    inst_binary "$KDUMP_KERNEL"
    inst_binary "$KDUMP_INITRD"
    chmod -x "${initdir}/$KDUMP_KERNEL"
}
