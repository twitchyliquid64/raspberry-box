package fs

import (
	"fmt"
	"os"

	ext4 "github.com/dsoprea/go-ext4"
)

const sectorSize = 512

// Ext4FS represents access to an ext4 fileystem.
type Ext4FS struct {
	f             *os.File
	start, length uint64

	sb   *ext4.Superblock
	bgdl *ext4.BlockGroupDescriptorList
	root *ext4.BlockGroupDescriptor
}

// LoadExt4 loads an ext4 filesystem.
func LoadExt4(f *os.File, start, length uint64) (*Ext4FS, error) {
	if _, err := f.Seek(int64(start)+ext4.Superblock0Offset, 0); err != nil {
		return nil, fmt.Errorf("seek failed: %v", err)
	}

	sb, err := ext4.NewSuperblockWithReader(f)
	if err != nil {
		return nil, fmt.Errorf("loading superblock: %v", err)
	}
	bgdl, err := ext4.NewBlockGroupDescriptorListWithReadSeeker(f, sb)
	if err != nil {
		return nil, fmt.Errorf("loading block group descriptor: %v", err)
	}
	root, err := bgdl.GetWithAbsoluteInode(ext4.InodeRootDirectory)
	if err != nil {
		return nil, fmt.Errorf("loading root directory: %v", err)
	}

	return &Ext4FS{
		f:      f,
		start:  start,
		length: length,
		sb:     sb,
		bgdl:   bgdl,
		root:   root,
	}, nil
}
