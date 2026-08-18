package deadband

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrDeadband = errors.New("deadband denied")

func CopyEUWindow(samples []float64, n int) []float64 {
	if n < 0 {
		n = 0
	}
	if n > len(samples) {
		n = len(samples)
	}
	out := make([]float64, n)
	copy(out, samples[:n])
	return out
}

type NodeBag struct {
	last map[string]float64
}

func NewNodeBag() *NodeBag {
	bag := &NodeBag{}
	bag.last = make(map[string]float64)
	return bag
}

func (b *NodeBag) Remember(node string, eu float64) {
	b.last[node] = eu
}

func (b *NodeBag) Last(node string) (float64, bool) {
	v, ok := b.last[node]
	return v, ok
}

func WrapDeadbandDenied(op, node string) error {
	if strings.TrimSpace(op) == "" {
		op = "subscribe"
	}
	if strings.TrimSpace(node) == "" {
		node = "unnamed"
	}
	return fmt.Errorf("%s: node %s: %w", op, node, ErrDeadband)
}

func WaitSettle(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func DumpItemsCSV(path, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriterSize(f, 4096)
	if _, err := w.WriteString(body); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return f.Sync()
}

func GrowSamples(dst []float64, extra float64) []float64 {
	out := make([]float64, len(dst)+1)
	copy(out, dst)
	out[len(dst)] = extra
	return out
}

type Node struct {
	Browse string
	EU     float64
}

func (n *Node) BrowseName() string {
	if n == nil {
		return ""
	}
	return n.Browse
}

func ParseEUSnapshot(b []byte) (map[string]float64, error) {
	var m map[string]float64
	if len(b) == 0 {
		return nil, errors.New("empty EU snapshot")
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func ExportNodeDump(root, rel string) (string, error) {
	if strings.TrimSpace(rel) == "" {
		return "", errors.New("empty node dump path")
	}
	if filepath.IsAbs(rel) {
		return "", errors.New("absolute node dump path")
	}
	clean := filepath.Clean(rel)
	full := filepath.Join(root, clean)
	relOut, err := filepath.Rel(filepath.Clean(root), full)
	if err != nil {
		return "", err
	}
	if relOut == ".." || strings.HasPrefix(relOut, ".."+string(filepath.Separator)) {
		return "", errors.New("node dump escapes root")
	}
	return full, nil
}
