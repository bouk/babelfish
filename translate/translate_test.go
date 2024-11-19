package translate

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"mvdan.cc/sh/v3/syntax"
	"strings"
	"testing"
)

func TestEscapedString(t *testing.T) {
	tr := NewTranslator()
	tr.escapedString(`cool 'shit' yo`)
	s := tr.buf.String()
	equal(t, `'cool \'shit\' yo'`, s)
}

func equal(t testing.TB, wanted, actual interface{}) {
	if diff := cmp.Diff(wanted, actual); diff != "" {
		t.Errorf("%s", diff)
		fmt.Println(actual)
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected string
	}{
		{
			name:     "chruby.sh",
			in:       chruby,
			expected: chrubyExpected,
		},
		{
			name:     "test.sh",
			in:       testFile,
			expected: testExpected,
		},
		{
			name:     "command-not-found.sh",
			in:       nixIndexFile,
			expected: nixIndexExpected,
		},
		{
			name: "java home",
			in:   `if [ -z "${JAVA_HOME-}" ]; then export JAVA_HOME=/bla/lib/openjdk; fi`,
			expected: `if [ -z (set -q JAVA_HOME && echo "$JAVA_HOME" || echo '') ]
  set -gx JAVA_HOME '/bla/lib/openjdk'
end
`,
		},
		{
			name: "recursive translation",
			in:   `source /opt/source.sh`,
			expected: `/bin/babelfish < /opt/source.sh | source
`,
		},
		{
			name: "append to PATH",
			in: `
export NIX_PATH="nixpkgs=/nix/var/nix/profiles/per-user/root/channels/nixos:nixos-config=/etc/nixos/configuration.nix"
export NIX_PATH="$HOME/.nix-defexpr/channels${NIX_PATH:+:$NIX_PATH}"`,
			expected: `set -gx NIX_PATH 'nixpkgs=/nix/var/nix/profiles/per-user/root/channels/nixos:nixos-config=/etc/nixos/configuration.nix'
set -gx NIX_PATH "$HOME"'/.nix-defexpr/channels'(test -n "$NIX_PATH" && echo ':'"$NIX_PATH" || echo)
`,
		},
		{
			name: "unset function and variable",
			in:   "unset -f foo -v bar",
			expected: `functions -e foo; set -e bar
`,
		},
		{name: "hash in name",
			in: `a=nixpkgs
nix run $a#hello
`, expected: `set a 'nixpkgs'
nix run $a#hello
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tr := NewTranslator()
			tr.babelFishLocation = "/bin/babelfish"
			p := syntax.NewParser(syntax.KeepComments(true), syntax.Variant(syntax.LangBash))
			f, err := p.Parse(strings.NewReader(test.in), test.name)
			if err != nil {
				t.Error(err)
				return
			}
			err = tr.File(f)
			if err != nil {
				t.Error(err)
				return
			}
			s := tr.buf.String()
			equal(t, test.expected, s)
		})
	}
}

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
          "$1")  match="$dir" && break ;;
          *"$1"*)  match="$dir" ;;
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

const chrubyExpected = `set CHRUBY_VERSION '0.3.9'
set RUBIES
for dir in "$PREFIX"'/opt/rubies' "$HOME"'/.rubies'
  test -d "$dir" && test -n (ls -A "$dir" | string collect; or echo) && set -a RUBIES "$dir"/*
end
set -e dir
function chruby_reset
  test -z "$RUBY_ROOT" && return
  set PATH ':'"$PATH"':'
  set PATH (string replace --all ':'"$RUBY_ROOT"'/bin:' ':' "$PATH")
  test -n "$GEM_ROOT" && set PATH (string replace --all ':'"$GEM_ROOT"'/bin:' ':' "$PATH")
  if test (id -ru) -ne 0
    test -n "$GEM_HOME" && set PATH (string replace --all ':'"$GEM_HOME"'/bin:' ':' "$PATH")
    set GEM_PATH ':'"$GEM_PATH"':'
    test -n "$GEM_HOME" && set GEM_PATH (string replace --all ':'"$GEM_HOME"':' ':' "$GEM_PATH")
    test -n "$GEM_ROOT" && set GEM_PATH (string replace --all ':'"$GEM_ROOT"':' ':' "$GEM_PATH")
    set GEM_PATH (string replace -r '^(\\.?:)' '' "$GEM_PATH")
    set GEM_PATH (string replace -r '(:\\.?)$' '' "$GEM_PATH")
    set -e GEM_HOME
    test -z "$GEM_PATH" && set -e GEM_PATH
  end
  set PATH (string replace -r '^(\\.?:)' '' "$PATH")
  set PATH (string replace -r '(:\\.?)$' '' "$PATH")
  set -e RUBY_ROOT; set -e RUBY_ENGINE; set -e RUBY_VERSION; set -e RUBYOPT; set -e GEM_ROOT
  true
end

function chruby_use
  if test ! -x $argv[1]'/bin/ruby'
    echo 'chruby: '$argv[1]'/bin/ruby not executable' >&2
    return 1
  end
  test -n "$RUBY_ROOT" && chruby_reset
  set -gx RUBY_ROOT $argv[1]
  set -gx RUBYOPT $argv[2]
  set -gx PATH "$RUBY_ROOT"'/bin:'"$PATH"
  eval (RUBYGEMS_GEMDEPS='' "$RUBY_ROOT"'/bin/ruby' - <(echo 'puts "export RUBY_ENGINE=#{Object.const_defined?(:RUBY_ENGINE) ? RUBY_ENGINE : \'ruby\'};"
puts "export RUBY_VERSION=#{RUBY_VERSION};"
begin; require \'rubygems\'; puts "export GEM_ROOT=#{Gem.default_dir.inspect};"; rescue LoadError; end
'| psub) | string collect; or echo)
  set -gx PATH (test -n "$GEM_ROOT" && echo "$GEM_ROOT"'/bin:' || echo)"$PATH"
  if test (id -ru) -ne 0
    set -gx GEM_HOME "$HOME"'/.gem/'"$RUBY_ENGINE"'/'"$RUBY_VERSION"
    set -gx GEM_PATH "$GEM_HOME"(test -n "$GEM_ROOT" && echo ':'"$GEM_ROOT" || echo)(test -n "$GEM_PATH" && echo ':'"$GEM_PATH" || echo)
    set -gx PATH "$GEM_HOME"'/bin:'"$PATH"
  end
  true
end

function chruby
  switch $argv[1]
  case '-h' '--help'
    echo 'usage: chruby [RUBY|VERSION|system] [RUBYOPT...]'
  case '-V' '--version'
    echo 'chruby: '"$CHRUBY_VERSION"
  case ''
    set -l dir $dir; set -l ruby $ruby
    for dir in $RUBIES
      set dir (string replace -r '(/)$' '' "$dir")
      set ruby (string replace -r '^(.*/)' '' "$dir")
      if test "$dir" = "$RUBY_ROOT"
        echo ' * '"$ruby"' '"$RUBYOPT"
      else
        echo '   '"$ruby"
      end
    end
  case 'system'
    chruby_reset
  case '*'
    set -l dir $dir; set -l ruby $ruby; set -l match $match
    for dir in $RUBIES
      set dir (string replace -r '(/)$' '' "$dir")
      set ruby (string replace -r '^(.*/)' '' "$dir")
      switch "$ruby"
      case $argv[1]
        set match "$dir" && break
      case '*'$argv[1]'*'
        set match "$dir"
      end
    end
    if test -z "$match"
      echo 'chruby: unknown Ruby: '$argv[1] >&2
      return 1
    end
    set -e argv[1]
    chruby_use "$match" "$argv"
  end
end
`

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
for cmd
do
  echo "$cmd"
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
(( 123 ))
`

const testExpected = `#!/usr/bin/env bash
# Prevent this file from being sourced by child shells.
set -gx __NIX_DARWIN_SET_ENVIRONMENT_DONE '1'
set A '2'
C='3' echo 23
set -gx A $A
set -gx PATH "$HOME"'/.nix-profile/bin:/run/current-system/sw/bin:/nix/var/nix/profiles/default/bin:/usr/local/bin:/usr/bin:/usr/sbin:/bin:/sbin'
set -gx EDITOR 'nano'
set -gx NIX_PATH 'darwin-config='"$HOME"'/dotfiles/darwin.nix:/nix/var/nix/profiles/per-user/root/channels:'"$HOME"'/.nix-defexpr/channels'
set -gx NIX_SSL_CERT_FILE '/etc/ssl/certs/ca-certificates.crt'
set -gx PAGER 'less -R'
echo 123 | source
cat (echo 123 | psub)
cat <test.bash
function cool
  cat | cat
end
echo (cat test.bash | cool | fish -c 'cool | cool | fish -c \'echo \\\'cool\\\' | cool\'')
test -e /var/file.sh && /bin/babelfish < /var/file.sh | source
if [ -z "$SSH_AUTH_SOCK" ]
  set -gx SSH_AUTH_SOCK (/bin/gpgconf --list-dirs agent-ssh-socket | string collect; or echo)
end
if [ -d '/share/gsettings-schemas/name' ]
  set -gx whatevs "$whatevs"(test -n "$whatevs" && echo ':' || echo)'/share/gsettings-schemas/name'
else if false
  true
else
  true
end
echo (set -q cool && echo 'a' || echo)
echo (test -n "$cool" && echo 'a' || echo)
echo (set -q cool && echo "$cool" || echo 'a')
echo (test -n "$cool" && echo "$cool" || echo 'a')
set -e ASPELL_CONF
for i in a b c
  if [ -d "$i"'/lib/aspell' ]
    set -gx ASPELL_CONF 'dict-dir '"$i"'/lib/aspell'
  end
  echo yes
end
for cmd in $argv
  echo "$cmd"
end
time sleep 1
while true
  echo 1
  echo 2
end
while not true
  echo 1
  echo 2
end
call $me
echo (count $argv)
echo (count $cool)
echo (string length "$cool")
set a (ok | string collect; or echo)
set a (ok | string collect; or echo)
/bin/babelfish < /etc/bashrc | source
test 123 != 0
`

const nixIndexFile = `#!/bin/sh

# for bash 4
# this will be called when a command is entered
# but not found in the user’s path + environment
command_not_found_handle () {

    # TODO: use "command not found" gettext translations

    # taken from http://www.linuxjournal.com/content/bash-command-not-found
    # - do not run when inside Midnight Commander or within a Pipe
    if [ -n "${MC_SID-}" ] || ! [ -t 1 ]; then
        >&2 echo "$1: command not found"
        return 127
    fi

    toplevel=nixpkgs # nixpkgs should always be available even in NixOS
    cmd=$1
    attrs=$(@out@/bin/nix-locate --minimal --no-group --type x --type s --top-level --whole-name --at-root "/bin/$cmd")
    len=$(echo -n "$attrs" | grep -c "^")

    case $len in
        0)
            >&2 echo "$cmd: command not found"
            ;;
        1)
            # if only 1 package provides this, then we can invoke it
            # without asking the users if they have opted in with one
            # of 2 environment variables

            # they are based on the ones found in
            # command-not-found.sh:

            #   NIX_AUTO_INSTALL : install the missing command into the
            #                      user’s environment
            #   NIX_AUTO_RUN     : run the command transparently inside of
            #                      nix shell

            # these will not return 127 if they worked correctly

            if ! [ -z "${NIX_AUTO_INSTALL-}" ]; then
                >&2 cat <<EOF
The program '$cmd' is currently not installed. It is provided by
the package '$toplevel.$attrs', which I will now install for you.
EOF
                nix-env -iA $toplevel.$attrs
                if [ "$?" -eq 0 ]; then
                    $@ # TODO: handle pipes correctly if AUTO_RUN/INSTALL is possible
                    return $?
                else
                    >&2 cat <<EOF
Failed to install $toplevel.attrs.
$cmd: command not found
EOF
                fi
            elif ! [ -z "${NIX_AUTO_RUN-}" ]; then
                nix-build --no-out-link -A $attrs "<$toplevel>"
                if [ "$?" -eq 0 ]; then
                    # how nix-shell handles commands is weird
                    # $(echo $@) is need to handle this
                    nix-shell -p $attrs --run "$(echo $@)"
                    return $?
                else
                    >&2 cat <<EOF
Failed to install $toplevel.attrs.
$cmd: command not found
EOF
                fi
            else
                >&2 cat <<EOF
The program '$cmd' is currently not installed. You can install it
by typing:
  nix-env -iA $toplevel.$attrs
EOF
            fi
            ;;
        *)
            >&2 cat <<EOF
The program '$cmd' is currently not installed. It is provided by
several packages. You can install it by typing one of the following:
EOF

            # ensure we get each element of attrs
            # in a cross platform way
            while read attr; do
                >&2 echo "  nix-env -iA $toplevel.$attr"
            done <<< "$attrs"
            ;;
    esac

    return 127 # command not found should always exit with 127
}

# for zsh...
# we just pass it to the bash handler above
# apparently they work identically
command_not_found_handler () {
    command_not_found_handle $@
    return $?
}`

const nixIndexExpected = `#!/bin/sh
# for bash 4
# this will be called when a command is entered
# but not found in the user’s path + environment
function command_not_found_handle
  # TODO: use "command not found" gettext translations
  # taken from http://www.linuxjournal.com/content/bash-command-not-found
  # - do not run when inside Midnight Commander or within a Pipe
  if [ -n (set -q MC_SID && echo "$MC_SID" || echo '') ] || ! [ -t 1 ]
    echo $argv[1]': command not found' >&2
    return 127
  end
  # nixpkgs should always be available even in NixOS
  set toplevel 'nixpkgs'
  set cmd $argv[1]
  set attrs (@out@/bin/nix-locate --minimal --no-group --type x --type s --top-level --whole-name --at-root '/bin/'"$cmd" | string collect; or echo)
  set len (echo -n "$attrs" | grep -c '^' | string collect; or echo)
  switch "$len"
  case '0'
    echo "$cmd"': command not found' >&2
  case '1'
    # if only 1 package provides this, then we can invoke it
    # without asking the users if they have opted in with one
    # of 2 environment variables
    # they are based on the ones found in
    # command-not-found.sh:
    #   NIX_AUTO_INSTALL : install the missing command into the
    #                      user’s environment
    #   NIX_AUTO_RUN     : run the command transparently inside of
    #                      nix shell
    # these will not return 127 if they worked correctly
    if ! [ -z (set -q NIX_AUTO_INSTALL && echo "$NIX_AUTO_INSTALL" || echo '') ]
      cat >&2 <(echo 'The program \''"$cmd"'\' is currently not installed. It is provided by
the package \''"$toplevel"'.'"$attrs"'\', which I will now install for you.
'| psub)
      nix-env -iA $toplevel.$attrs
      if [ "$status" -eq 0 ]
        # TODO: handle pipes correctly if AUTO_RUN/INSTALL is possible
        $argv
        return $status
      else
        cat >&2 <(echo 'Failed to install '"$toplevel"'.attrs.
'"$cmd"': command not found
'| psub)
      end
    else if ! [ -z (set -q NIX_AUTO_RUN && echo "$NIX_AUTO_RUN" || echo '') ]
      nix-build --no-out-link -A $attrs '<'"$toplevel"'>'
      if [ "$status" -eq 0 ]
        # how nix-shell handles commands is weird
        # $(echo $@) is need to handle this
        nix-shell -p $attrs --run (echo $argv | string collect; or echo)
        return $status
      else
        cat >&2 <(echo 'Failed to install '"$toplevel"'.attrs.
'"$cmd"': command not found
'| psub)
      end
    else
      cat >&2 <(echo 'The program \''"$cmd"'\' is currently not installed. You can install it
by typing:
  nix-env -iA '"$toplevel"'.'"$attrs"'
'| psub)
    end
  case '*'
    cat >&2 <(echo 'The program \''"$cmd"'\' is currently not installed. It is provided by
several packages. You can install it by typing one of the following:
'| psub)
    # ensure we get each element of attrs
    # in a cross platform way
    while read attr
      echo '  nix-env -iA '"$toplevel"'.'"$attr" >&2
    end <(echo "$attrs"| psub)
  end
  # command not found should always exit with 127
  return 127
end

# for zsh...
# we just pass it to the bash handler above
# apparently they work identically
function command_not_found_handler
  command_not_found_handle $argv
  return $status
end
`
