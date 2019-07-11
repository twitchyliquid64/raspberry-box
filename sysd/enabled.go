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

// FS describes an interface which must be provided, so the package
// can interact with the filesystem.
type FS interface {
	Stat(path string) (os.FileInfo, error)
	LStat(path string) (os.FileInfo, error)
}

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

	return nil
}
