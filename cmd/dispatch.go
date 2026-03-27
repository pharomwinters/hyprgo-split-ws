package cmd

import (
	"fmt"
	"strconv"

	"github.com/adambick/hyprgo-split-ws/config"
	"github.com/adambick/hyprgo-split-ws/ipc"
	"github.com/adambick/hyprgo-split-ws/workspace"
)

// resolveRealWorkspace gets the focused monitor, builds the mapping,
// and converts a virtual workspace number to a real one.
func resolveRealWorkspace(cfg *config.Config, virtual int) (int, error) {
	if virtual < 1 || virtual > cfg.WorkspacesPerMonitor {
		return 0, fmt.Errorf("workspace must be between 1 and %d", cfg.WorkspacesPerMonitor)
	}

	monitor, err := ipc.GetFocusedMonitor()
	if err != nil {
		return 0, err
	}

	mapping, err := workspace.BuildMonitorMapping(cfg.MonitorOrder)
	if err != nil {
		return 0, err
	}

	idx, err := mapping.MonitorIndex(monitor.Name)
	if err != nil {
		return 0, err
	}

	return workspace.ToReal(idx, virtual, cfg.WorkspacesPerMonitor), nil
}

// SwitchWorkspace switches to a virtual workspace on the focused monitor.
func SwitchWorkspace(cfg *config.Config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: hyprgo-split-ws workspace <1-%d>", cfg.WorkspacesPerMonitor)
	}

	virtual, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid workspace number: %s", args[0])
	}

	real, err := resolveRealWorkspace(cfg, virtual)
	if err != nil {
		return err
	}

	if _, err = ipc.Dispatch("workspace", strconv.Itoa(real)); err != nil {
		return err
	}
	workspace.RenameWorkspace(real, cfg.WorkspacesPerMonitor)
	return nil
}

// MoveToWorkspace moves the active window to a virtual workspace on the focused monitor.
func MoveToWorkspace(cfg *config.Config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: hyprgo-split-ws movetoworkspace <1-%d>", cfg.WorkspacesPerMonitor)
	}

	virtual, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid workspace number: %s", args[0])
	}

	real, err := resolveRealWorkspace(cfg, virtual)
	if err != nil {
		return err
	}

	if _, err = ipc.Dispatch("movetoworkspace", strconv.Itoa(real)); err != nil {
		return err
	}
	workspace.RenameWorkspace(real, cfg.WorkspacesPerMonitor)
	return nil
}

// MoveToWorkspaceSilent moves the active window without following it.
func MoveToWorkspaceSilent(cfg *config.Config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: hyprgo-split-ws movetoworkspacesilent <1-%d>", cfg.WorkspacesPerMonitor)
	}

	virtual, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid workspace number: %s", args[0])
	}

	real, err := resolveRealWorkspace(cfg, virtual)
	if err != nil {
		return err
	}

	if _, err = ipc.Dispatch("movetoworkspacesilent", strconv.Itoa(real)); err != nil {
		return err
	}
	workspace.RenameWorkspace(real, cfg.WorkspacesPerMonitor)
	return nil
}

// ChangeMonitor moves the focused window to the next or previous monitor.
// direction should be "next" or "prev".
func ChangeMonitor(cfg *config.Config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: hyprgo-split-ws changemonitor <next|prev>")
	}

	direction := args[0]
	if direction != "next" && direction != "prev" {
		return fmt.Errorf("direction must be 'next' or 'prev', got %q", direction)
	}

	monitors, err := ipc.GetMonitors()
	if err != nil {
		return err
	}

	if len(monitors) < 2 {
		return nil // nothing to do with one monitor
	}

	mapping, err := workspace.BuildMonitorMapping(cfg.MonitorOrder)
	if err != nil {
		return err
	}

	// Find focused monitor's index.
	var focusedName string
	for _, m := range monitors {
		if m.Focused {
			focusedName = m.Name
			break
		}
	}
	if focusedName == "" {
		return fmt.Errorf("no focused monitor found")
	}

	currentIdx, err := mapping.MonitorIndex(focusedName)
	if err != nil {
		return err
	}

	// Calculate target monitor index with wrapping.
	total := len(monitors)
	var targetIdx int
	if direction == "next" {
		targetIdx = (currentIdx + 1) % total
	} else {
		targetIdx = (currentIdx - 1 + total) % total
	}

	targetName := mapping.IndexToName[targetIdx]

	// Move the window to the target monitor.
	_, err = ipc.Dispatch("movewindow", fmt.Sprintf("mon:%s", targetName))
	return err
}
