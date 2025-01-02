# We are using -S to ensure the scope is correct
function _babelfish_source -S
  if test "$argv[1]" = '-' || string match -q '*.fish' "$argv[1]" || test -z "$argv[1]"
    builtin source $argv
  else
    babelfish < $argv[1] | builtin source
  end
end

function source -S
  _babelfish_source $argv
end

function . -S
  _babelfish_source $argv
end

function _babelfish_translate_bash
    set -l data
    if type -q pbpaste
        set data (pbpaste 2>/dev/null | string collect -N)
    else if set -q WAYLAND_DISPLAY; and type -q wl-paste
        set data (wl-paste -n 2>/dev/null | string collect -N)
    else if set -q DISPLAY; and type -q xsel
        set data (xsel --clipboard | string collect -N)
    else if set -q DISPLAY; and type -q xclip
        set data (xclip -selection clipboard -o 2>/dev/null | string collect -N)
    else if type -q powershell.exe
        set data (powershell.exe Get-Clipboard | string trim -r -c \r | string collect -N)
    end

    # Issue 6254: Handle zero-length clipboard content
    if not string length -q -- "$data"
        return 1
    end

    if not isatty stdout
        # If we're redirected, just write the data *as-is*.
        printf %s $data
        return
    end

    set data  (echo "$data" | babelfish)
    commandline -i -- $data
end

bind \cb _fish_translate_bash
