#!/bin/sh
# Clang rejects __sync_* on _Atomic* in gnulib's Cygwin pthread-once workaround.
# Safe no-op if the file or pattern is absent.
set -e
f=lib/pthread-once.c
if [ -f "$f" ] && grep -q '__sync_bool_compare_and_swap (&state_p->done, 1, 2)' "$f"; then
	sed -i.bak 's/__sync_bool_compare_and_swap (&state_p->done, 1, 2)/__sync_bool_compare_and_swap ((unsigned short *) \&state_p->done, 1, 2)/' "$f"
fi
