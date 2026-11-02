//go:build !windows

package volumedriver

import (
	"os"
	"syscall"
)

// stateReadFlags refuses to follow a symlink in the final path component and
// does not block on a FIFO.
const stateReadFlags = os.O_RDONLY | syscall.O_NOFOLLOW | syscall.O_NONBLOCK

func ownedBy(info os.FileInfo, uid int) bool {
	st, ok := info.Sys().(*syscall.Stat_t)
	return ok && int(st.Uid) == uid
}

// singleLink reports whether the file has exactly one directory entry; O_NOFOLLOW
// does not stop a hardlink to some other file the driver user owns.
func singleLink(info os.FileInfo) bool {
	st, ok := info.Sys().(*syscall.Stat_t)
	return ok && st.Nlink == 1
}
