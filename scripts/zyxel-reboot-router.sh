#!/bin/sh
sshpass -p 'mypassw0rd' ssh \
  -oHostKeyAlgorithms=+ssh-rsa \
  -oPubkeyAcceptedAlgorithms=+ssh-rsa \
  -oStrictHostKeyChecking=no \
  -oUserKnownHostsFile=/dev/null \
  admin@192.168.1.1 <<'EOF'
zycli reboot
exit
EOF