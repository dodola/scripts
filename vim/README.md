# Helpful VIM tools for Fuchsia development

## Features

* Configure YouCompleteMe to provide error checking, completion and source navigation within the Fuchsia tree.
* Set path so that `:find` and `gf` know how to find files.

## Installation

Make sure `env.sh` is being called in your login scripts. This code depends on variables set in `env.sh` and by the
`fset` command.

Add this to your `vimrc`:
```
if $FUCHSIA_DIR != ""
  source $FUCHSIA_DIR/scripts/vim/fuchsia.vim
endif
```

Optionally install [YouCompleteMe](https://github.com/Valloric/YouCompleteMe) for fancy completion, source navigation
and inline errors.

## TODO

In the future it would be nice to support:
* Syntax highlighting and indentation for the languages we use
* Build system integration
