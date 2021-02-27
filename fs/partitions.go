// Package fs exposes access to filesystems.
package fs

import (
	"fmt"
	"os/exec"

	"github.com/rekby/mbr"
)

// Partition types
const (
	PartitionTypeFAT32LBA             = 0xc
	PartitionTypeLinuxNativePartition = 0x83
)

// CheckPiPartitionTable returns an error if the pi image has a bad partition table.
func CheckPiPartitionTable(t *mbr.MBR) error {
	if err := t.Check(); err != nil {
		return err
	}
	partitions := t.GetAllPartitions()

	partitionSpec := []struct {
		Type mbr.PartitionType
	}{
		{Type: PartitionTypeFAT32LBA},
		{Type: PartitionTypeLinuxNativePartition},
		{Type: mbr.PART_EMPTY},
		{Type: mbr.PART_EMPTY},
	}

	if len(partitions) != 4 {
		return fmt.Errorf("4 partitions expected, got %d", len(partitions))
	}

	for i, spec := range partitionSpec {
		if spec.Type != partitions[i].GetType() {
			return fmt.Errorf("partition at index %d has type %v", i, partitions[i].GetType())
		}
	}

	return nil
}

func ExpandImage(path string, partition int) error {
	if err := exec.Command("parted", "-s", path, "resizepart", fmt.Sprint(partition), "100%").Run(); err != nil {
		return fmt.Errorf("parted: %v", err)
	}
	return nil
}
