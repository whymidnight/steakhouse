let SessionLoad = 1
let s:so_save = &g:so | let s:siso_save = &g:siso | setg so=0 siso=0 | setl so=-1 siso=-1
let v:this_session=expand("<sfile>:p")
silent only
silent tabonly
cd ~/solana_staking/anchor-escrow/go
if expand('%') == '' && !&modified && line('$') <= 1 && getline(1) == ''
  let s:wipebuf = bufnr('%')
endif
set shortmess=aoO
badd +73 src/staking/init.go
badd +88 src/staking/events/events.go
badd +65 src/staking/events/onWalletCreate.go
badd +85 src/staking/events/eventsHandler.go
badd +77 src/utils/utils.go
badd +4 src/staking/typestructs/stake.go
argglobal
%argdel
edit src/staking/typestructs/stake.go
set splitbelow splitright
wincmd _ | wincmd |
vsplit
wincmd _ | wincmd |
vsplit
wincmd _ | wincmd |
vsplit
3wincmd h
wincmd w
wincmd w
wincmd w
wincmd _ | wincmd |
split
1wincmd k
wincmd w
set nosplitbelow
set nosplitright
wincmd t
set winminheight=0
set winheight=1
set winminwidth=0
set winwidth=1
exe 'vert 1resize ' . ((&columns * 59 + 119) / 238)
exe 'vert 2resize ' . ((&columns * 59 + 119) / 238)
exe 'vert 3resize ' . ((&columns * 59 + 119) / 238)
exe '4resize ' . ((&lines * 25 + 20) / 41)
exe 'vert 4resize ' . ((&columns * 58 + 119) / 238)
exe '5resize ' . ((&lines * 13 + 20) / 41)
exe 'vert 5resize ' . ((&columns * 58 + 119) / 238)
argglobal
balt src/staking/events/eventsHandler.go
setlocal fdm=manual
setlocal fde=0
setlocal fmr={{{,}}}
setlocal fdi=#
setlocal fdl=0
setlocal fml=1
setlocal fdn=20
setlocal fen
silent! normal! zE
let &fdl = &fdl
let s:l = 4 - ((2 * winheight(0) + 19) / 39)
if s:l < 1 | let s:l = 1 | endif
keepjumps exe s:l
normal! zt
keepjumps 4
normal! 020|
wincmd w
argglobal
if bufexists("src/staking/events/eventsHandler.go") | buffer src/staking/events/eventsHandler.go | else | edit src/staking/events/eventsHandler.go | endif
if &buftype ==# 'terminal'
  silent file src/staking/events/eventsHandler.go
endif
balt src/staking/events/onWalletCreate.go
setlocal fdm=manual
setlocal fde=0
setlocal fmr={{{,}}}
setlocal fdi=#
setlocal fdl=0
setlocal fml=1
setlocal fdn=20
setlocal fen
silent! normal! zE
let &fdl = &fdl
let s:l = 85 - ((16 * winheight(0) + 19) / 39)
if s:l < 1 | let s:l = 1 | endif
keepjumps exe s:l
normal! zt
keepjumps 85
normal! 015|
wincmd w
argglobal
if bufexists("src/staking/events/events.go") | buffer src/staking/events/events.go | else | edit src/staking/events/events.go | endif
if &buftype ==# 'terminal'
  silent file src/staking/events/events.go
endif
balt src/staking/init.go
setlocal fdm=manual
setlocal fde=0
setlocal fmr={{{,}}}
setlocal fdi=#
setlocal fdl=0
setlocal fml=1
setlocal fdn=20
setlocal fen
silent! normal! zE
let &fdl = &fdl
let s:l = 93 - ((16 * winheight(0) + 19) / 39)
if s:l < 1 | let s:l = 1 | endif
keepjumps exe s:l
normal! zt
keepjumps 93
normal! 055|
wincmd w
argglobal
if bufexists("src/utils/utils.go") | buffer src/utils/utils.go | else | edit src/utils/utils.go | endif
if &buftype ==# 'terminal'
  silent file src/utils/utils.go
endif
balt src/staking/init.go
setlocal fdm=manual
setlocal fde=0
setlocal fmr={{{,}}}
setlocal fdi=#
setlocal fdl=0
setlocal fml=1
setlocal fdn=20
setlocal fen
silent! normal! zE
let &fdl = &fdl
let s:l = 77 - ((18 * winheight(0) + 12) / 25)
if s:l < 1 | let s:l = 1 | endif
keepjumps exe s:l
normal! zt
keepjumps 77
normal! 05|
wincmd w
argglobal
if bufexists("src/staking/init.go") | buffer src/staking/init.go | else | edit src/staking/init.go | endif
if &buftype ==# 'terminal'
  silent file src/staking/init.go
endif
balt src/staking/init.go
setlocal fdm=manual
setlocal fde=0
setlocal fmr={{{,}}}
setlocal fdi=#
setlocal fdl=0
setlocal fml=1
setlocal fdn=20
setlocal fen
silent! normal! zE
let &fdl = &fdl
let s:l = 73 - ((6 * winheight(0) + 6) / 13)
if s:l < 1 | let s:l = 1 | endif
keepjumps exe s:l
normal! zt
keepjumps 73
normal! 0
wincmd w
exe 'vert 1resize ' . ((&columns * 59 + 119) / 238)
exe 'vert 2resize ' . ((&columns * 59 + 119) / 238)
exe 'vert 3resize ' . ((&columns * 59 + 119) / 238)
exe '4resize ' . ((&lines * 25 + 20) / 41)
exe 'vert 4resize ' . ((&columns * 58 + 119) / 238)
exe '5resize ' . ((&lines * 13 + 20) / 41)
exe 'vert 5resize ' . ((&columns * 58 + 119) / 238)
tabnext 1
if exists('s:wipebuf') && len(win_findbuf(s:wipebuf)) == 0&& getbufvar(s:wipebuf, '&buftype') isnot# 'terminal'
  silent exe 'bwipe ' . s:wipebuf
endif
unlet! s:wipebuf
set winheight=1 winwidth=20 winminheight=1 winminwidth=1 shortmess=filnxtToOF
let s:sx = expand("<sfile>:p:r")."x.vim"
if filereadable(s:sx)
  exe "source " . fnameescape(s:sx)
endif
let &g:so = s:so_save | let &g:siso = s:siso_save
nohlsearch
doautoall SessionLoadPost
unlet SessionLoad
" vim: set ft=vim :
