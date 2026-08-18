package deadband

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWaitSettleHonorsCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	err := WaitSettle(ctx, 600*time.Millisecond)
	if err == nil {
		t.Fatal("expected cancel error")
	}
	if time.Since(start) > 250*time.Millisecond {
		t.Fatalf("WaitSettle ignored cancel, elapsed=%s", time.Since(start))
	}
}

func TestWrapDeadbandDeniedIs(t *testing.T) {
	err := WrapDeadbandDenied("subscribe", "Kiln.Temp")
	if !errors.Is(err, ErrDeadband) {
		t.Fatalf("lost ErrDeadband: %v", err)
	}
}

func TestCopyEUWindowIndependent(t *testing.T) {
	src := []float64{842, 843, 844, 845}
	got := CopyEUWindow(src, 2)
	got[0] = 1
	if src[0] != 842 {
		t.Fatal("CopyEUWindow aliased the sample window")
	}
}

func TestAfterWriteRejectsEUDrop(t *testing.T) {
	min := ""
	get := func() (string, error) { return min, nil }
	set := func(v string) error { min = v; return nil }
	if err := AfterWrite(get, set, "eu=842.0 deadband=2.5"); err != nil {
		t.Fatal(err)
	}
	if err := AfterWrite(get, set, "eu=800 deadband=2.5"); err == nil {
		t.Fatal("expected EU drop below last committed value to be rejected")
	}
}

func TestNilNodeBrowseName(t *testing.T) {
	var n *Node
	if n.BrowseName() != "" {
		t.Fatalf("got %q", n.BrowseName())
	}
}

func TestParseEUSnapshotRejectsInvalid(t *testing.T) {
	if _, err := ParseEUSnapshot([]byte("eu=not-json")); err == nil {
		t.Fatal("expected JSON error")
	}
}

func TestDumpItemsCSVPersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "items.csv")
	body := "node,eu\nKiln.Temp,842\n"
	if err := DumpItemsCSV(path, body); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != body {
		t.Fatalf("got %q", b)
	}
}

func TestExportNodeDumpRejectsEscape(t *testing.T) {
	if _, err := ExportNodeDump(t.TempDir(), filepath.Join("..", "secret")); err == nil {
		t.Fatal("expected path escape to be rejected")
	}
}

func TestNodeBagRemember(t *testing.T) {
	bag := NewNodeBag()
	bag.Remember("Kiln.Temp", 842)
	v, ok := bag.Last("Kiln.Temp")
	if !ok || v != 842 {
		t.Fatalf("got %v %v", v, ok)
	}
}

func TestGrowSamplesNoWriteThrough(t *testing.T) {
	dst := make([]float64, 2, 8)
	dst[0], dst[1] = 1, 2
	got := GrowSamples(dst, 3)
	got[0] = 99
	if dst[0] != 1 {
		t.Fatal("GrowSamples wrote through into the sample buffer")
	}
}
