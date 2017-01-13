" Only run if $FUCHSIA_DIR has been set by env.sh
if $FUCHSIA_DIR != ""
  let g:ycm_global_ycm_extra_conf = $FUCHSIA_DIR . '/scripts/vim/ycm_extra_conf.py'

  function FuchsiaBuffer()
    " Set up path so that 'gf' and :find do what we want.
    let &l:path = $PWD. "/**" . "," . $FUCHSIA_DIR . "," .
          \ $FUCHSIA_BUILD_DIR . "," .
          \ $FUCHSIA_BUILD_DIR . "/gen"
    if g:loaded_youcompleteme
      " Replace the normal go to tag key with YCM.
      nnoremap <C-]> :YcmCompleter GoTo<cr>
    endif
  endfunction

  augroup fuchsia
    " Configure buffers that are in the fuchsia tree.
    autocmd BufRead,BufNewFile $FUCHSIA_DIR/** call FuchsiaBuffer()
  augroup END

endif
