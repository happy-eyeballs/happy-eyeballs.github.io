# Local testbed setup

## Setup for dedicated macOS client device (e.g. Macbook)

1. Configure 10.0.1.1/16 and 3000::1:1/64 as IP addresses for the interface (note: when I used private IPv6 addresses on
   macOS such as fc00::1:1, it would always prefer IPv4 over IPv6, therefore I switched to a public IPv6 prefix)
2. Add SSH key for remote user on client device
3. Check that selenium works for the browser (e.g. enable remote automation in Safari Developer Settings)
4. Configure DNS nameserver for happyeyeballs.local by creating the file /etc/resolver/happyeyeballs.local with the
   content:
   ```
   domain happyeyeballs.local
   nameserver 10.0.2.1
   ```
   Then restart DNS resolver using `killall -HUP mDNSResponder`.
   Verify using `scutil --dns`
5. Allow `tcpdump` command to be executed without sudo password interaction by
   adding `kirstein ALL = NOPASSWD: /usr/sbin/tcpdump` at the end of `sudo visudo`


## Setup for local machine as server

1. Generate SSH key with name `id_ed25519` in this folder
2. Adjust username of remote SSH user in `runner-config.yml`
3. Adjust interface name in `runner-config.yml` as the `NETEM_INTERFACE` and update set it in `setup-interface.sh` (must be the interface which the client is connected to)
4. Adjust client's interface name (interface that this device is connected to) as the `TCPDUMP_INTERFACE` in `runner-config.yml`
5. Execute `setup-interface.sh` to configure the IP addresses for the interface

Execute `run.sh` to start the tests.
