package sysd

import (
	"os"
	"path/filepath"
)

// Exists returns true if the unit exists.
func Exists(fs FS, unit string) (bool, error) {
	_, err := fs.Stat(filepath.Join("/lib/systemd/system", unit))
	switch {
	case err != nil && !os.IsNotExist(err):
		return false, err
	case err != nil && os.IsNotExist(err):
		return false, nil
	default:
		return true, nil
	}
}
