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

func TestParseChr(t *testing.T) {
	err := perform("test.bash", strings.NewReader(chruby))
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseArithm(t *testing.T) {
	err := perform("test.bash", strings.NewReader(arithm))
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
. /etc/bashrc
`

const chruby = `
CHRUBY_VERSION="0.3.9"
RUBIES=()

for dir in "$PREFIX/opt/rubies" "$HOME/.rubies"; do
	[[ -d "$dir" && -n "$(ls -A "$dir")" ]] && RUBIES+=("$dir"/*)
done
unset dir

function chruby_reset()
{
	[[ -z "$RUBY_ROOT" ]] && return

	PATH=":$PATH:"; PATH="${PATH//:$RUBY_ROOT\/bin:/:}"
	[[ -n "$GEM_ROOT" ]] && PATH="${PATH//:$GEM_ROOT\/bin:/:}"

	if (( UID != 0 )); then
		[[ -n "$GEM_HOME" ]] && PATH="${PATH//:$GEM_HOME\/bin:/:}"

		GEM_PATH=":$GEM_PATH:"
		[[ -n "$GEM_HOME" ]] && GEM_PATH="${GEM_PATH//:$GEM_HOME:/:}"
		[[ -n "$GEM_ROOT" ]] && GEM_PATH="${GEM_PATH//:$GEM_ROOT:/:}"
		GEM_PATH="${GEM_PATH#:}"; GEM_PATH="${GEM_PATH%:}"

		unset GEM_HOME
		[[ -z "$GEM_PATH" ]] && unset GEM_PATH
	fi

	PATH="${PATH#:}"; PATH="${PATH%:}"
	unset RUBY_ROOT RUBY_ENGINE RUBY_VERSION RUBYOPT GEM_ROOT
	hash -r
}

function chruby_use()
{
	if [[ ! -x "$1/bin/ruby" ]]; then
		echo "chruby: $1/bin/ruby not executable" >&2
		return 1
	fi

	[[ -n "$RUBY_ROOT" ]] && chruby_reset

	export RUBY_ROOT="$1"
	export RUBYOPT="$2"
	export PATH="$RUBY_ROOT/bin:$PATH"

	eval "$(RUBYGEMS_GEMDEPS="" "$RUBY_ROOT/bin/ruby" - <<EOF
puts "export RUBY_ENGINE=#{Object.const_defined?(:RUBY_ENGINE) ? RUBY_ENGINE : 'ruby'};"
puts "export RUBY_VERSION=#{RUBY_VERSION};"
begin; require 'rubygems'; puts "export GEM_ROOT=#{Gem.default_dir.inspect};"; rescue LoadError; end
EOF
)"
	export PATH="${GEM_ROOT:+$GEM_ROOT/bin:}$PATH"

	if (( UID != 0 )); then
		export GEM_HOME="$HOME/.gem/$RUBY_ENGINE/$RUBY_VERSION"
		export GEM_PATH="$GEM_HOME${GEM_ROOT:+:$GEM_ROOT}${GEM_PATH:+:$GEM_PATH}"
		export PATH="$GEM_HOME/bin:$PATH"
	fi

	hash -r
}

function chruby()
{
	case "$1" in
		-h|--help)
			echo "usage: chruby [RUBY|VERSION|system] [RUBYOPT...]"
			;;
		-V|--version)
			echo "chruby: $CHRUBY_VERSION"
			;;
		"")
			local dir ruby
			for dir in "${RUBIES[@]}"; do
				dir="${dir%%/}"; ruby="${dir##*/}"
				if [[ "$dir" == "$RUBY_ROOT" ]]; then
					echo " * ${ruby} ${RUBYOPT}"
				else
					echo "   ${ruby}"
				fi

			done
			;;
		system) chruby_reset ;;
		*)
			local dir ruby match
			for dir in "${RUBIES[@]}"; do
				dir="${dir%%/}"; ruby="${dir##*/}"
				case "$ruby" in
					"$1")	match="$dir" && break ;;
					*"$1"*)	match="$dir" ;;
				esac
			done

			if [[ -z "$match" ]]; then
				echo "chruby: unknown Ruby: $1" >&2
				return 1
			fi

			shift
			chruby_use "$match" "$*"
			;;
	esac
}
`

const arithm = `
(( 123 ))
`
