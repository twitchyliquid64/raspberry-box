package sysd

import (
	"os"
	"path/filepath"

	"github.com/twitchyliquid64/raspberry-box/conf/sysd"
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

// Install installs the specified unit using the given name.
func Install(fs FS, unitName string, conf *sysd.Unit, overwrite bool) error {
	exists, err := Exists(fs, unitName)
	if err != nil {
		return err
	}
	if exists && !overwrite {
		return os.ErrExist
	}

	b := []byte(conf.String())

	return fs.Write(filepath.Join("/lib/systemd/system", unitName), b, 0644)
}
