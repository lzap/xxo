function! CommandE(tab, cmd)
  call system('videx "' . a:tab . '" ' . getcwd() . ' "' . a:cmd . '"')
  redraw
  echo 'Command Exec:' a:cmd
endfunction

map <F5> :wa\|echo unused<CR>
map <F6> :wa\|echo unused<CR>
map <F7> :wa\|echo unused<CR>
map <F8> :wa\|echo unused<CR>
map <F9> :wa\|silent! make\|cw<CR>
map <F10> :wa\|!go build && ./xxo<CR>

echo 'CommandE: Project loaded...'
