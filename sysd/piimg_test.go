package sysd

import (
	"flag"
	"os"
	"testing"

	"github.com/rekby/mbr"
	"github.com/twitchyliquid64/raspberry-box/fs"
)

var (
	imgPath    = flag.String("pi-img", "", "Path to a mint raspbian image.")
	sectorSize = uint32(512)
)

func TestKMountPiImgUnitExists(t *testing.T) {
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

	m, err := fs.KMountExt4(*imgPath, uint64(tab.GetPartition(2).GetLBAStart()*sectorSize), uint64(tab.GetPartition(2).GetLBALen()*sectorSize))
	if err != nil {
		t.Fatalf("KMountExt4() failed: %v", err)
	}
	defer m.Close()

	enabled, err := Exists(m, "ssh.service")
	if err != nil {
		t.Errorf("Exists('ssh.service') failed: %v", err)
	}
	if !enabled {
		t.Errorf("Exists('ssh.service') = %v, want true", enabled)
	}
}

func TestKMountPiImgUnitEnable(t *testing.T) {
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

	m, err := fs.KMountExt4(*imgPath, uint64(tab.GetPartition(2).GetLBAStart()*sectorSize), uint64(tab.GetPartition(2).GetLBALen()*sectorSize))
	if err != nil {
		t.Fatalf("KMountExt4() failed: %v", err)
	}
	defer m.Close()

	enabled, err := IsEnabledOnTarget(m, "ssh.service", "default.target")
	if err != nil {
		t.Errorf("IsEnabledOnTarget('ssh.service') failed: %v", err)
	}
	if enabled {
		t.SkipNow()
		return
	}

	if err := Enable(m, "ssh.service", "default.target"); err != nil {
		t.Errorf("Enable() failed: %v", err)
	}
}

func TestKMountPiImgEnabled(t *testing.T) {
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

	m, err := fs.KMountExt4(*imgPath, uint64(tab.GetPartition(2).GetLBAStart()*sectorSize), uint64(tab.GetPartition(2).GetLBALen()*sectorSize))
	if err != nil {
		t.Fatalf("KMountExt4() failed: %v", err)
	}
	defer m.Close()

	enabled, err := IsEnabledOnTarget(m, "cron.service", "multi-user.target")
	if err != nil {
		t.Errorf("IsEnabledOnTarget('cron.service', 'multi-user.target') failed: %v", err)
	}
	if !enabled {
		t.Errorf("IsEnabledOnTarget('cron.service', 'multi-user.target') = %v, want true", enabled)
	}
}
