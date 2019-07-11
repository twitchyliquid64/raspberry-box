package fs

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/rekby/mbr"
)

var (
	imgPath = flag.String("pi-img", "", "Path to a mint raspbian image.")
)

func init() {
	flag.Parse()
}

func TestSoftLoadPiImg(t *testing.T) {
	if *imgPath == "" {
		t.SkipNow()
	}

	f, err := os.Open(*imgPath)
	if err != nil {
		t.Fatalf("Failed to open image: %v\n", err)
	}

	tab, err := mbr.Read(f)
	if err != nil {
		t.Fatalf("Failed to read table: %v\n", err)
	}
	defer f.Close()

	if err := CheckPiPartitionTable(tab); err != nil {
		t.Errorf("partition table check failed: %v", err)
	}

	ext4, err := LoadExt4(f, uint64(tab.GetPartition(2).GetLBAStart()*sectorSize), uint64(tab.GetPartition(2).GetLBALen()*sectorSize))
	if err != nil {
		t.Fatalf("failed to create fs: %v", err)
	}
	t.Log(ext4)
}

func TestKMountPiImg(t *testing.T) {
	if *imgPath == "" {
		t.SkipNow()
	}

	f, err := os.Open(*imgPath)
	if err != nil {
		t.Fatalf("Failed to open image: %v\n", err)
	}

	tab, err := mbr.Read(f)
	if err != nil {
		t.Fatalf("Failed to read table: %v\n", err)
	}
	defer f.Close()

	if err := CheckPiPartitionTable(tab); err != nil {
		t.Errorf("partition table check failed: %v", err)
	}

	tmpMnt, err := ioutil.TempDir("", "raspberry-box")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpMnt)

	m, err := KMountExt4(*imgPath, tmpMnt, uint64(tab.GetPartition(2).GetLBAStart()*sectorSize), uint64(tab.GetPartition(2).GetLBALen()*sectorSize))
	if err != nil {
		t.Fatalf("KMountExt4() failed: %v", err)
	}
	if err := m.Close(); err != nil {
		t.Errorf("Close() failed: %v", err)
	}
}
