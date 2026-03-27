package workspace

import (
	"fmt"

	"github.com/adambick/hyprgo-split-ws/ipc"
)

// ToReal converts a virtual workspace number (1-N) and a monitor index
// to the real Hyprland workspace ID.
func ToReal(monitorIndex int, virtual int, perMonitor int) int {
	return (monitorIndex * perMonitor) + virtual
}

// ToVirtual converts a real Hyprland workspace ID to its virtual number
// and the monitor index it belongs to.
func ToVirtual(real int, perMonitor int) (monitorIndex int, virtual int) {
	if real <= 0 {
		return 0, 0
	}
	monitorIndex = (real - 1) / perMonitor
	virtual = real - (monitorIndex * perMonitor)
	return monitorIndex, virtual
}

// VirtualName returns the display name for a real workspace ID.
func VirtualName(real int, perMonitor int) string {
	_, virtual := ToVirtual(real, perMonitor)
	return fmt.Sprintf("%d", virtual)
}

// RenameAllWorkspaces renames all existing workspaces to their virtual numbers.
func RenameAllWorkspaces(perMonitor int) error {
	workspaces, err := ipc.GetWorkspaces()
	if err != nil {
		return fmt.Errorf("failed to get workspaces: %w", err)
	}

	for _, ws := range workspaces {
		if ws.ID <= 0 {
			continue // skip special workspaces
		}
		name := VirtualName(ws.ID, perMonitor)
		if err := ipc.RenameWorkspace(ws.ID, name); err != nil {
			return fmt.Errorf("failed to rename workspace %d: %w", ws.ID, err)
		}
	}
	return nil
}

// RenameWorkspace renames a single workspace to its virtual number.
func RenameWorkspace(realID int, perMonitor int) error {
	if realID <= 0 {
		return nil
	}
	return ipc.RenameWorkspace(realID, VirtualName(realID, perMonitor))
}
