# babelfish

Translate bash scripts to [fish](https://fishshell.com).

## Why?

Because I got annoyed by having to use [fish-foreign-env](https://github.com/oh-my-fish/plugin-foreign-env) or [bass](https://github.com/edc/bass), which are slow, since they create multiple bash processes. With this program I can translate bash scripts to fish, and run them directly in fish.

## But how?

`babelfish` parses the script using [mvdan.cc/sh](https://github.com/mvdan/sh), and then translates bash expressions to the equivalent fish code. That's it! You can find the code that walks the AST and emits fish code [here](https://github.com/bouk/babelfish/blob/master/translate/translate.go).

## Install

`GO111MODULE=on go get bou.ke/babelfish`

## Example

```sh
# Pass some code on stdin to translate it
$ echo 'f() { export SSH_AUTH_SOCK=$(gpgconf --list-dirs agent-ssh-socket); local cool=yep; }' | babelfish
function f
  set -gx SSH_AUTH_SOCK (gpgconf --list-dirs agent-ssh-socket | string collect; or echo)
  set -l cool 'yep'
end
# Pass the result to source to load it into fish
$ echo 'echo Nice to meet you user $UID' | babelfish | source
Nice to meet you user 502
# Or install the shell hook!
$ source babel.fish
$ source chruby.sh
$ chruby
   ruby-2.5
   ruby-2.6
   ruby-2.7
```

## To do

Probably still a lot. There's a couple variables like `$BASH_SOURCE` that aren't translated, and not all arithmetic expressions are implemented either. Pull requests and issues welcome!
