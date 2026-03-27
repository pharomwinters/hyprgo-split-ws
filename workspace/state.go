package workspace

import (
	"fmt"
	"sort"

	"github.com/adambick/hyprgo-split-ws/ipc"
)

// MonitorMapping holds the association between monitor names and their indices.
type MonitorMapping struct {
	// NameToIndex maps monitor names (e.g., "DP-3") to their assigned index.
	NameToIndex map[string]int
	// IndexToName maps indices back to monitor names.
	IndexToName map[int]string
}

// BuildMonitorMapping queries Hyprland for active monitors and assigns
// indices. If customOrder is provided, monitors are ordered accordingly;
// any monitors not in the custom order are appended alphabetically.
// If customOrder is nil/empty, all monitors are sorted alphabetically.
func BuildMonitorMapping(customOrder []string) (*MonitorMapping, error) {
	monitors, err := ipc.GetMonitors()
	if err != nil {
		return nil, fmt.Errorf("failed to get monitors: %w", err)
	}

	if len(monitors) == 0 {
		return nil, fmt.Errorf("no monitors found")
	}

	mapping := &MonitorMapping{
		NameToIndex: make(map[string]int, len(monitors)),
		IndexToName: make(map[int]string, len(monitors)),
	}

	if len(customOrder) > 0 {
		// Build a set of active monitor names for lookup.
		active := make(map[string]bool, len(monitors))
		for _, m := range monitors {
			active[m.Name] = true
		}

		// Assign indices from custom order first (only if monitor is active).
		idx := 0
		for _, name := range customOrder {
			if active[name] {
				mapping.NameToIndex[name] = idx
				mapping.IndexToName[idx] = name
				idx++
			}
		}

		// Append any active monitors not in the custom order, sorted alphabetically.
		var remaining []string
		for _, m := range monitors {
			if _, ok := mapping.NameToIndex[m.Name]; !ok {
				remaining = append(remaining, m.Name)
			}
		}
		sort.Strings(remaining)
		for _, name := range remaining {
			mapping.NameToIndex[name] = idx
			mapping.IndexToName[idx] = name
			idx++
		}
	} else {
		// Default: sort alphabetically for stable ordering.
		sort.Slice(monitors, func(i, j int) bool {
			return monitors[i].Name < monitors[j].Name
		})
		for i, m := range monitors {
			mapping.NameToIndex[m.Name] = i
			mapping.IndexToName[i] = m.Name
		}
	}

	return mapping, nil
}

// MonitorIndex returns the index for a given monitor name.
func (m *MonitorMapping) MonitorIndex(name string) (int, error) {
	idx, ok := m.NameToIndex[name]
	if !ok {
		return 0, fmt.Errorf("unknown monitor: %s", name)
	}
	return idx, nil
}
