# Arch Linux ARM on Raspberry Pi 4 + 5
This is a guide to install Arch Linux ARM on a Raspberry Pi 4 or 5.

## Prerequisites

`Hardware`
- Raspberry Pi 4 or 5
- MicroSD card
- USB-C power supply
- HDMI cable
- Monitor

`Software`
- [Arch Linux ARM image for Raspberry Pi 4 or 5](http://arch.jlake.co/aarch64/ArchLinuxARM-rpi-aarch64-2024-11-07.tar.gz)
- [fdisk (util-linux)](https://www.archlinux.org/packages/core/x86_64/util-linux/)
- [bsdtar (libarchive)](https://www.archlinux.org/packages/extra/x86_64/libarchive/)
- [nano (nano)](https://www.archlinux.org/packages/core/x86_64/nano/)

## Installation

1. Download the latest Arch Linux ARM image for Raspberry Pi 4 or 5 from the [my website](https://arch.jlake.co/aarch64/ArchLinuxARM-rpi-aarch64-2024-11-07.tar.gz).

```bash
wget http://arch.jlake.co/aarch64/ArchLinuxARM-rpi-aarch64-2024-11-07.tar.gz
```

or

```bash
curl -O http://arch.jlake.co/aarch64/ArchLinuxARM-rpi-aarch64-2024-11-07.tar.gz
```

2. Format SD Card.

```bash
sudo fdisk /dev/sdX
```

```
Command (m for help): o # Create a new empty DOS partition table
Command (m for help): n # Create a new partition
Command (m for help): p # Primary partition
Partition number (1-4): 1
First sector (2048-62333951, default 2048): # Press Enter
Last sector, +/-sectors or +/-size{K,M,G,T,P} (2048-62333951, default 62333951): +200M # Press Enter
Command (m for help): t # Change partition type
Partition number (1-4): 1
Hex code (type L to list all codes): c # W95 FAT32 (LBA)
Command (m for help): n # Create a new partition
Command (m for help): p # Primary partition
Partition number (1-4): 2
First sector (411648-62333951, default 411648): # Press Enter
Last sector, +/-sectors or +/-size{K,M,G,T,P} (411648-62333951, default 62333951): # Press Enter
Command (m for help): w # Write changes
```

```bash
sudo mkfs.vfat /dev/sdX1
sudo mkfs.ext4 /dev/sdX2
```

3. Mount the SD card.

```bash
sudo mkdir boot
sudo mkdir root
sudo mount /dev/sdX1 boot
sudo mount /dev/sdX2 root
```

4. Extract the Arch Linux ARM image.

```bash
bsdtar -xpf ArchLinuxARM-rpi-aarch64-2024-11-07.tar.gz -C root
sync
mv root/boot/* boot
```

5. Configure the system.

```bash
sudo nano root/etc/fstab
```

```bash
/dev/mmcblk0p1  /boot   vfat    defaults        0       0
/dev/mmcblk0p2  /       ext4    defaults,rw,errors=remount-ro 0       1
```

6. Unmount the SD card.

```bash
sudo umount boot root
```

7. Insert the SD card into the Raspberry Pi.

8. Connect the HDMI cable to the monitor.

9. Connect the USB-C power supply to the Raspberry Pi.

10. Boot the Raspberry Pi.

11. Login as the `alarm` user with the password `alarm`.

12. Change to the `root` user with the password `root`.
```bash
su -
```

14. Initialize the pacman keyring.

```bash
pacman-key --init
pacman-key --populate archlinuxarm
```

14. Update the system.

```bash
pacman -Syu
```

15. Reboot the Raspberry Pi.

```bash
reboot
```
