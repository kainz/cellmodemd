# simple and ugly interface to ModemManager to get an IP Bearer from the first available modem, connect it, and wait for something to happen to the connection.

(Also write a systemd-networkd link/network file and trigger that)

Usage:

 cellmodemd -a NAME-OF-APN

32-bit raspbian build:

    GOARCH=arm GOARM=7 GOOS=linux go build -o rpi-cellmodemd.buster gopkg.in/kainz/cellmodemd.v0/modemd

64-bit raspbian build:

    # pi4s are armv8 pi3s are 64bit capable armv7, but goloang doesnt do anything
    # armv8 specific yet
    GOARCH=arm64 GOARM=7 GOOS=linux go build -o rpi-cellmodemd.buster gopkg.in/kainz/cellmodemd.v0/modemd
