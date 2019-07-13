// Package sysd analyzes and affects systemd configuration.
package sysd

import (
	"os"
)

// FS describes an interface which must be provided, so the package
// can interact with the filesystem.
type FS interface {
	Stat(path string) (os.FileInfo, error)
	LStat(path string) (os.FileInfo, error)
	Symlink(at, to string) error
	Mkdir(at string) error
	Write(path string, data []byte, perms os.FileMode) error
}
