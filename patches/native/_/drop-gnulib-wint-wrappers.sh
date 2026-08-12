#!/bin/sh
# Gnulib's stddef wrapper breaks Cygwin's __need_wint_t protocol (sys/_types.h
# sees an unknown wint_t). Replace stddef.h with #include_next, then restore
# gnulib's unreachable() helper that newer packages expect from stddef.h.
# Leave *.in.h present and older than the stub so make does not regenerate it.
set -e
stub_stddef() {
	dir=$1
	f="$dir/stddef.h"
	inh="$dir/stddef.in.h"
	[ -f "$f" ] || [ -f "$inh" ] || return 0
	rm -f "$f"
	cat >"$f" <<'EOF'
#include_next <stddef.h>
#ifndef unreachable
# if defined __has_builtin
#  if __has_builtin (__builtin_unreachable)
#   define unreachable() __builtin_unreachable ()
#  endif
# endif
#endif
#ifndef unreachable
# if defined __GNUC__ || defined __clang__
#  define unreachable() __builtin_unreachable ()
# else
#  define unreachable() ((void) 0)
# endif
#endif
EOF
	cp -f "$f" "$inh"
	touch -t 197001010000.01 "$inh" 2>/dev/null || true
	touch "$f"
}
for dir in lib libgrep libgzip gnulib-lib; do
	[ -d "$dir" ] || continue
	stub_stddef "$dir"
done
find . -type d \( -name lib -o -name gnulib-lib \) 2>/dev/null |
while read -r dir; do
	stub_stddef "$dir"
done
