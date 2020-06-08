package main

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	err := perform("test.bash", strings.NewReader(testFile))
	if err != nil {
		t.Fatal(err)
	}
}

const testFile = `
#!/usr/bin/env bash

# Prevent this file from being sourced by child shells.
export __NIX_DARWIN_SET_ENVIRONMENT_DONE=1
A=2
C=3 echo 23
export A

export PATH=$HOME/.nix-profile/bin:/run/current-system/sw/bin:/nix/var/nix/profiles/default/bin:/usr/local/bin:/usr/bin:/usr/sbin:/bin:/sbin
export EDITOR="nano"
export NIX_PATH="darwin-config=$HOME/dotfiles/darwin.nix:/nix/var/nix/profiles/per-user/root/channels:$HOME/.nix-defexpr/channels"
export NIX_SSL_CERT_FILE="/etc/ssl/certs/ca-certificates.crt"
export PAGER="less -R"
echo 123 | source
cat <(echo 123)
cat < test.bash
cool() {
	cat | cat
}
echo $(cat test.bash | cool | (cool | cool | ( echo 'cool' | cool)))
test -e /var/file.sh && source /var/file.sh
if [ -z "$SSH_AUTH_SOCK" ]; then
  export SSH_AUTH_SOCK=$(/bin/gpgconf --list-dirs agent-ssh-socket)
fi
if [ -d "/share/gsettings-schemas/name" ]; then
	export whatevs=$whatevs${whatevs:+:}/share/gsettings-schemas/name
elif false; then
	true
else
	true
fi
echo ${cool+a}
echo ${cool:+a}
echo ${cool-a}
echo ${cool:-a}
unset ASPELL_CONF
for i in a b c ; do
  if [ -d "$i/lib/aspell" ]; then
    export ASPELL_CONF="dict-dir $i/lib/aspell"
  fi
	echo yes
done
time sleep 1
while true; do
	echo 1
	echo 2
done
until true; do
	echo 1
	echo 2
done
call $me
echo ${#@}
echo ${#cool[@]}
echo ${#cool}
a=$(ok)
a="$(ok)"
`
