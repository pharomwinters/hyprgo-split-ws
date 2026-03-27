package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/adambick/hyprgo-split-ws/config"
	"github.com/adambick/hyprgo-split-ws/ipc"
	"github.com/adambick/hyprgo-split-ws/workspace"
)

// RunDaemon starts the event listener daemon.
// It renames all existing workspaces on startup, then listens for events
// and renames new workspaces as they are created or moved.
func RunDaemon(cfg *config.Config) error {
	fmt.Println(":: hyprgo-split-ws daemon starting")

	// Build initial monitor mapping.
	mapping, err := workspace.BuildMonitorMapping(cfg.MonitorOrder)
	if err != nil {
		return fmt.Errorf("failed to build monitor mapping: %w", err)
	}
	log.Printf("Monitor mapping: %v", mapping.NameToIndex)

	// Rename all existing workspaces on startup.
	if err := workspace.RenameAllWorkspaces(cfg.WorkspacesPerMonitor); err != nil {
		log.Printf("warning: failed to rename existing workspaces: %v", err)
	}
	log.Println("Renamed existing workspaces")

	// Connect to event socket.
	listener, err := ipc.NewEventListener()
	if err != nil {
		return fmt.Errorf("failed to start event listener: %w", err)
	}
	defer listener.Close()

	fmt.Println(":: Listening for Hyprland events...")

	for {
		event, err := listener.Next()
		if err != nil {
			return fmt.Errorf("event listener error: %w", err)
		}

		switch event.Name {
		case "createworkspacev2":
			handleCreateWorkspace(event.Data, cfg.WorkspacesPerMonitor)

		case "moveworkspacev2":
			handleMoveWorkspace(event.Data, cfg.WorkspacesPerMonitor)

		case "monitoraddedv2":
			log.Printf("[monitoradded] %s", event.Data)
			mapping, err = workspace.BuildMonitorMapping(cfg.MonitorOrder)
			if err != nil {
				log.Printf("failed to rebuild monitor mapping: %v", err)
			} else {
				log.Printf("Monitor mapping updated: %v", mapping.NameToIndex)
				workspace.RenameAllWorkspaces(cfg.WorkspacesPerMonitor)
			}

		case "monitorremovedv2":
			log.Printf("[monitorremoved] %s", event.Data)
			removedMonitor := parseMonitorRemoved(event.Data)
			// Rebuild mapping without the removed monitor.
			mapping, err = workspace.BuildMonitorMapping(cfg.MonitorOrder)
			if err != nil {
				log.Printf("failed to rebuild monitor mapping: %v", err)
			} else {
				log.Printf("Monitor mapping updated: %v", mapping.NameToIndex)
			}
			// Recover windows that were on the removed monitor.
			if removedMonitor != "" {
				recoverRogueWindows(mapping, cfg)
			}
		}
	}
}

// handleCreateWorkspace renames a newly created workspace.
// Data format: WORKSPACEID,WORKSPACENAME
func handleCreateWorkspace(data string, perMonitor int) {
	parts := strings.SplitN(data, ",", 2)
	if len(parts) < 2 {
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		return
	}
	if err := workspace.RenameWorkspace(id, perMonitor); err != nil {
		log.Printf("failed to rename workspace %d: %v", id, err)
	} else {
		log.Printf("Renamed workspace %d -> %s", id, workspace.VirtualName(id, perMonitor))
	}
}

// handleMoveWorkspace renames a workspace that moved to a different monitor.
// Data format: WORKSPACEID,WORKSPACENAME,MONNAME
func handleMoveWorkspace(data string, perMonitor int) {
	parts := strings.SplitN(data, ",", 3)
	if len(parts) < 3 {
		return
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil || id <= 0 {
		return
	}
	if err := workspace.RenameWorkspace(id, perMonitor); err != nil {
		log.Printf("failed to rename moved workspace %d: %v", id, err)
	}
}

// parseMonitorRemoved extracts the monitor name from a monitorremovedv2 event.
// Data format: MONID,MONNAME,MONDESC
func parseMonitorRemoved(data string) string {
	parts := strings.SplitN(data, ",", 3)
	if len(parts) < 2 {
		return ""
	}
	return parts[1]
}

// recoverRogueWindows finds windows on workspaces that no longer belong to
// any active monitor and moves them to the first active monitor's workspace 1.
func recoverRogueWindows(mapping *workspace.MonitorMapping, cfg *config.Config) {
	clients, err := ipc.GetClients()
	if err != nil {
		log.Printf("failed to get clients for rogue recovery: %v", err)
		return
	}

	if len(mapping.IndexToName) == 0 {
		return
	}

	// Find the fallback workspace (monitor 0, virtual workspace 1).
	fallbackReal := workspace.ToReal(0, 1, cfg.WorkspacesPerMonitor)
	totalSlots := len(mapping.IndexToName) * cfg.WorkspacesPerMonitor

	recovered := 0
	for _, client := range clients {
		wsID := client.Workspace.ID
		if wsID <= 0 {
			continue // special workspace
		}
		// A workspace is "rogue" if its ID is beyond the range of active monitors.
		if wsID > totalSlots {
			log.Printf("Recovering window %s (%s) from workspace %d", client.Class, client.Address, wsID)
			_, err := ipc.Dispatch("movetoworkspacesilent", fmt.Sprintf("%d,address:%s", fallbackReal, client.Address))
			if err != nil {
				log.Printf("failed to recover window %s: %v", client.Address, err)
			} else {
				recovered++
			}
		}
	}

	if recovered > 0 {
		log.Printf("Recovered %d rogue window(s)", recovered)
		workspace.RenameAllWorkspaces(cfg.WorkspacesPerMonitor)
	}
}
