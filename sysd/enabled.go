package sysd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var (
	// ErrNotInstalled is returned if an action is invoked
	// on a unit which does not exist.
	ErrNotInstalled = errors.New("cannot perform action on uninstalled unit")
)

// IsEnabled returns true if the given unit is enabled on the target.
func IsEnabledOnTarget(fs FS, unit, target string) (bool, error) {
	s, err := fs.LStat(filepath.Join("/etc/systemd/system", target+".wants", unit))
	switch {
	case err != nil && !os.IsNotExist(err):
		return false, err
	case err == nil && s.Mode()&os.ModeSymlink == 0:
		return false, fmt.Errorf("expected symlink on %s", filepath.Join("/etc/systemd/system", target+".wants"))
	case err == nil && s.Mode()&os.ModeSymlink != 0:
		return true, nil
	}

	s, err = fs.LStat(filepath.Join("/lib/systemd/system", target+".wants", unit))
	switch {
	case err != nil && !os.IsNotExist(err):
		return false, err
	case err == nil && s.Mode()&os.ModeSymlink == 0:
		return false, fmt.Errorf("expected symlink on %s", filepath.Join("/lib/systemd/system", target+".wants"))
	case err == nil && s.Mode()&os.ModeSymlink != 0:
		return true, nil
	}

	return false, nil
}

// Enable enables the given unit as a WantedBy target.
func Enable(fs FS, unit, target string) error {
	enabled, err := IsEnabledOnTarget(fs, unit, target)
	if err != nil {
		return err
	}
	if enabled {
		return nil
	}
	exists, err := Exists(fs, unit)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotInstalled
	}

	if _, err := fs.Stat(filepath.Join("/lib/systemd/system", target+".wants")); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := fs.Mkdir(filepath.Join("/lib/systemd/system", target+".wants")); err != nil {
			return err
		}
	}

	return fs.Symlink(filepath.Join("/lib/systemd/system", target+".wants", unit), "../"+unit)
}
