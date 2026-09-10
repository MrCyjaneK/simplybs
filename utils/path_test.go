package utils

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestToShellPathWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}
	got := ToShellPath(`C:\Users\user\work\simplybs\.buildlib\env-native`)
	want := `/cygdrive/c/Users/user/work/simplybs/.buildlib/env-native`
	if got != want {
		t.Fatalf("ToShellPath: got %q, want %q", got, want)
	}
}

func TestDestDirJoinWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows only")
	}
	staging := `C:\Users\user\work\simplybs\.buildlib\windows_amd64\staging\pkg`
	prefix := `C:\Users\user\work\simplybs\.buildlib\env-native`
	got := DestDirJoin(staging, prefix)
	shellAsWin := filepath.Join(staging, filepath.FromSlash(strings.TrimPrefix(ToShellPath(prefix), "/")))
	if got != shellAsWin {
		t.Fatalf("DestDirJoin vs shell concat:\n join %q\n mapped %q\n shellPrefix %q", got, shellAsWin, ToShellPath(prefix))
	}
}
