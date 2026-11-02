//go:build windows

package volumedriver

import "os"

const stateReadFlags = os.O_RDONLY

func ownedBy(info os.FileInfo, uid int) bool {
	return true
}

func singleLink(info os.FileInfo) bool {
	return true
}
