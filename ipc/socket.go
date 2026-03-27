package ipc

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

func socketDir() (string, error) {
	sig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if sig == "" {
		return "", fmt.Errorf("HYPRLAND_INSTANCE_SIGNATURE not set — is Hyprland running?")
	}
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		return "", fmt.Errorf("XDG_RUNTIME_DIR not set")
	}
	return filepath.Join(runtimeDir, "hypr", sig), nil
}

func commandSocketPath() (string, error) {
	dir, err := socketDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".socket.sock"), nil
}

func EventSocketPath() (string, error) {
	dir, err := socketDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".socket2.sock"), nil
}

// Command sends a raw command to Hyprland and returns the response.
func Command(cmd string) ([]byte, error) {
	sockPath, err := commandSocketPath()
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Hyprland socket: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write([]byte(cmd))
	if err != nil {
		return nil, fmt.Errorf("failed to write command: %w", err)
	}

	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 4096)
	for {
		n, err := conn.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}

	return buf, nil
}

// Dispatch sends a dispatcher command to Hyprland.
func Dispatch(dispatcher string, args string) ([]byte, error) {
	cmd := fmt.Sprintf("/dispatch %s %s", dispatcher, args)
	return Command(cmd)
}

// QueryJSON sends a JSON query and unmarshals the response.
func QueryJSON(query string, v any) error {
	resp, err := Command("j/" + query)
	if err != nil {
		return err
	}
	return json.Unmarshal(resp, v)
}

// Monitor represents a Hyprland monitor.
type Monitor struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	ActiveWorkspace Workspace `json:"activeWorkspace"`
	Focused         bool      `json:"focused"`
}

// Workspace represents a Hyprland workspace (as nested in monitor/query responses).
type Workspace struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// WorkspaceFull represents a full workspace object from the workspaces query.
type WorkspaceFull struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Monitor   string `json:"monitor"`
	MonitorID int    `json:"monitorID"`
	Windows   int    `json:"windows"`
}

// GetMonitors returns all monitors.
func GetMonitors() ([]Monitor, error) {
	var monitors []Monitor
	err := QueryJSON("monitors", &monitors)
	return monitors, err
}

// GetFocusedMonitor returns the currently focused monitor.
func GetFocusedMonitor() (*Monitor, error) {
	monitors, err := GetMonitors()
	if err != nil {
		return nil, err
	}
	for _, m := range monitors {
		if m.Focused {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("no focused monitor found")
}

// GetWorkspaces returns all workspaces.
func GetWorkspaces() ([]WorkspaceFull, error) {
	var workspaces []WorkspaceFull
	err := QueryJSON("workspaces", &workspaces)
	return workspaces, err
}

// Client represents a Hyprland window/client.
type Client struct {
	Address   string    `json:"address"`
	Class     string    `json:"class"`
	Title     string    `json:"title"`
	Workspace Workspace `json:"workspace"`
	Monitor   int       `json:"monitor"`
	PID       int       `json:"pid"`
}

// GetClients returns all clients/windows.
func GetClients() ([]Client, error) {
	var clients []Client
	err := QueryJSON("clients", &clients)
	return clients, err
}

// RenameWorkspace renames a workspace by its ID.
func RenameWorkspace(id int, name string) error {
	_, err := Command(fmt.Sprintf("/dispatch renameworkspace %d %s", id, name))
	return err
}
