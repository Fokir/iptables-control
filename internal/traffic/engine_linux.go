//go:build linux

package traffic

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// CollectInterfaces reads /proc/net/dev and returns per-interface byte counters.
func (e *Engine) CollectInterfaces(names []string) ([]InterfaceSnapshot, error) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, fmt.Errorf("open /proc/net/dev: %w", err)
	}
	defer f.Close()

	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}

	var results []InterfaceSnapshot
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum <= 2 {
			continue // skip header lines
		}
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		ifName := strings.TrimSpace(parts[0])
		if len(nameSet) > 0 && !nameSet[ifName] {
			continue
		}

		fields := strings.Fields(parts[1])
		if len(fields) < 10 {
			continue
		}

		var bytesIn, bytesOut uint64
		fmt.Sscanf(fields[0], "%d", &bytesIn)
		fmt.Sscanf(fields[8], "%d", &bytesOut)

		results = append(results, InterfaceSnapshot{
			Name:     ifName,
			BytesIn:  bytesIn,
			BytesOut: bytesOut,
		})
	}

	return results, scanner.Err()
}
