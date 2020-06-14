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
