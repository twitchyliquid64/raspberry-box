package fs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"time"

	losetup "github.com/freddierice/go-losetup"
)

// KMount represents access to a filesystem contained within a image file,
// mediated by the kernel.
type KMount struct {
	mntPoint     string
	loop         losetup.Device
	needsUnmount bool
}

// Cat implements interpreter.FS.
func (m *KMount) Cat(path string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(m.mntPoint, path))
}

// Remove implements interpreter.FS.
func (m *KMount) Remove(path string) error {
	return os.Remove(filepath.Join(m.mntPoint, path))
}

// RemoveAll implements interpreter.FS.
func (m *KMount) RemoveAll(path string) error {
	return os.RemoveAll(filepath.Join(m.mntPoint, path))
}

// Chmod implements interpreter.FS.
func (m *KMount) Chmod(path string, mode os.FileMode) error {
	return os.Chmod(filepath.Join(m.mntPoint, path), mode)
}

// Chown implements interpreter.FS.
func (m *KMount) Chown(path string, uid, gid int) error {
	return os.Chown(filepath.Join(m.mntPoint, path), uid, gid)
}

// LStat implements sysd.FS.
func (m *KMount) LStat(path string) (os.FileInfo, error) {
	return os.Lstat(filepath.Join(m.mntPoint, path))
}

// Stat implements sysd.FS.
func (m *KMount) Stat(path string) (os.FileInfo, error) {
	return os.Stat(filepath.Join(m.mntPoint, path))
}

// Symlink implements sysd.FS.
func (m *KMount) Symlink(at, to string) error {
	return os.Symlink(to, filepath.Join(m.mntPoint, at))
}

// Mkdir implements sysd.FS.
func (m *KMount) Mkdir(at string) error {
	return os.Mkdir(filepath.Join(m.mntPoint, at), 0755)
}

// Write implements sysd.FS.
func (m *KMount) Write(path string, data []byte, perms os.FileMode) error {
	return ioutil.WriteFile(filepath.Join(m.mntPoint, path), data, perms)
}

// KMountExt4 invokes mount() to mount the ext4 filesystem in the given image,
// at the provided mount point.
func KMountExt4(img string, start, length uint64) (*KMount, error) {
	mntPoint, err := ioutil.TempDir("", "raspberry-box")
	if err != nil {
		return nil, fmt.Errorf("TempDir() failed: %v", err)
	}

	l, err := losetup.Attach(img, start, false)
	if err != nil {
		return nil, fmt.Errorf("loop failed: %v", err)
	}

	if err := syscall.Mount(l.Path(), mntPoint, "ext4", syscall.MS_NOATIME, ""); err != nil {
		l.Detach()
		return nil, fmt.Errorf("mount failed: %v", err)
	}
	return &KMount{
		mntPoint:     mntPoint,
		loop:         l,
		needsUnmount: true,
	}, nil
}

// KMountVFat invokes mount() to mount the vfat filesystem in the given image,
// at the provided mount point.
func KMountVFat(img string, start, length uint64) (*KMount, error) {
	mntPoint, err := ioutil.TempDir("", "raspberry-box")
	if err != nil {
		return nil, fmt.Errorf("TempDir() failed: %v", err)
	}

	l, err := losetup.Attach(img, start, false)
	if err != nil {
		return nil, fmt.Errorf("loop failed: %v", err)
	}

	if err := syscall.Mount(l.Path(), mntPoint, "vfat", syscall.MS_NOATIME, ""); err != nil {
		l.Detach()
		return nil, fmt.Errorf("mount failed: %v", err)
	}
	return &KMount{
		mntPoint:     mntPoint,
		loop:         l,
		needsUnmount: true,
	}, nil
}

// Close gracefully shuts down the mount, removing any loopbacks
// or mount points created in the process of mounting.
func (m *KMount) Close() error {
	var umntErr error
	for i := 0; i < 4; i++ {
		if umntErr = syscall.Unmount(m.mntPoint, syscall.MNT_DETACH); umntErr != nil {
			time.Sleep(250 * time.Millisecond)
			continue
		}
		break
	}
	if umntErr != nil {
		return fmt.Errorf("unmount failed: %v", umntErr)
	}

	if err := m.loop.Detach(); err != nil {
		return fmt.Errorf("loopback detach failed: %v", err)
	}
	if m.mntPoint != "" {
		return os.Remove(m.mntPoint)
	}
	return nil
}
