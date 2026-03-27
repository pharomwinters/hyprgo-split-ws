# hyprgo-split-ws

A Go daemon that provides per-monitor workspace management for Hyprland, similar to KDE Plasma's virtual desktops per screen.

## Problem

Hyprland has global workspaces — workspace 1 exists on one monitor only, and switching to it jumps you there. We want each monitor to have its own independent set of workspaces (1-10), so pressing `Super+1` on DP-2 goes to DP-2's workspace 1, and `Super+1` on DP-3 goes to DP-3's workspace 1.

## Approach: Workspace Offsetting

Each monitor is assigned a range of Hyprland workspace IDs:

| Monitor | Virtual 1-10 | Real Hyprland IDs |
|---------|-------------|-------------------|
| DP-3 (id 0) | 1-10 | 1-10 |
| DP-2 (id 1) | 1-10 | 11-20 |
| DP-4 (id 2) | 1-10 | 21-30 |

Formula: `real_workspace = (monitor_index * 10) + virtual_workspace`

Waybar displays the virtual number (1-10), not the real ID.

## Architecture

```
┌──────────────┐     keybind exec      ┌─────────────────────┐
│   Hyprland   │ ──────────────────────>│  hyprgo-split-ws CLI  │
│  keybinds    │                        │  (dispatch mode)    │
└──────────────┘                        └────────┬────────────┘
                                                 │ command
                                                 v
┌──────────────┐     events (socket2)   ┌─────────────────────┐
│   Hyprland   │ ─────────────────────> │  hyprgo-split-ws      │
│  compositor  │ <───────────────────── │  daemon              │
│              │   dispatch (socket)    │                     │
└──────────────┘                        └─────────────────────┘
```

### Single Binary, Two Modes

- **`hyprgo-split-ws daemon`** — Long-running process started in Hyprland autostart. Connects to the IPC event socket, maintains monitor/workspace state, and handles rogue window recovery on monitor disconnect/reconnect.

- **`hyprgo-split-ws [command] [args]`** — Short-lived CLI calls from keybinds. Communicates with the daemon (or directly with Hyprland IPC if the daemon is not required for that command).

## Hyprland IPC

### Sockets

| Socket | Path | Purpose |
|--------|------|---------|
| Command | `$XDG_RUNTIME_DIR/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket.sock` | Send commands, get JSON responses |
| Event | `$XDG_RUNTIME_DIR/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket2.sock` | Stream real-time events |

**Critical:** Command socket connections are processed synchronously by Hyprland. Open, write, read, close — do not hold connections open or the compositor freezes (5s timeout).

### Relevant Events (socket2)

| Event | Data | Use |
|-------|------|-----|
| `focusedmonv2` | `MONNAME,WORKSPACEID` | Track which monitor is focused |
| `workspacev2` | `WORKSPACEID,WORKSPACENAME` | Track workspace changes |
| `monitoraddedv2` | `MONID,MONNAME,MONDESC` | Handle monitor hotplug |
| `monitorremovedv2` | `MONID,MONNAME,MONDESC` | Recover orphaned windows |
| `movewindowv2` | `WINADDR,WSID,WSNAME` | Track window movement |
| `openwindow` | `WINADDR,WSNAME,CLASS,TITLE` | Track new windows |

### Key Dispatchers

| Dispatcher | Syntax | Description |
|------------|--------|-------------|
| `workspace` | `workspace ID` | Switch to workspace |
| `movetoworkspace` | `workspace [,window]` | Move window and follow |
| `movetoworkspacesilent` | `workspace [,window]` | Move window, stay |
| `focusworkspaceoncurrentmonitor` | `workspace` | Focus workspace on current monitor (swaps if on another) |
| `moveworkspacetomonitor` | `workspace monitor` | Move workspace to monitor |

**Note:** `focusworkspaceoncurrentmonitor` is a built-in dispatcher that handles some of what we need, but our tool provides consistent numbering, Waybar integration, and edge case handling that it doesn't.

### Querying State

```
j/monitors       — array of monitor objects (id, name, activeWorkspace, focused)
j/workspaces     — array of workspace objects (id, name, monitor, monitorID, windows)
j/clients        — array of window objects (address, workspace, monitor, class, title)
j/activeworkspace — current workspace object
```

## CLI Commands

| Command | Keybind Example | Description |
|---------|----------------|-------------|
| `hyprgo-split-ws workspace N` | `$mainMod, 1` | Switch to virtual workspace N on focused monitor |
| `hyprgo-split-ws movetoworkspace N` | `$mainMod SHIFT, 1` | Move active window to virtual workspace N on focused monitor |
| `hyprgo-split-ws movetoworkspacesilent N` | — | Move window without following |
| `hyprgo-split-ws changemonitor` | `$mainMod, comma/period` | Move focused window to next/prev monitor |
| `hyprgo-split-ws daemon` | autostart | Start the event listener daemon |

## Keybind Integration

```conf
# In 08-keybinds.conf
bind = $mainMod, 1, exec, hyprgo-split-ws workspace 1
bind = $mainMod, 2, exec, hyprgo-split-ws workspace 2
# ...
bind = $mainMod SHIFT, 1, exec, hyprgo-split-ws movetoworkspace 1
bind = $mainMod SHIFT, 2, exec, hyprgo-split-ws movetoworkspace 2
# ...
```

## Waybar Integration

Waybar will show virtual workspace numbers. Options:

1. **Rename workspaces** — The daemon renames Hyprland workspaces to their virtual numbers using `hyprctl dispatch renameworkspace`. Waybar shows the name instead of the ID.

2. **Custom Waybar module** — A `custom/workspaces` module that queries the daemon for display state.

Option 1 is simpler and works with the built-in `hyprland/workspaces` module.

## Project Structure

```
hyprgo-split-ws/
├── DESIGN.md
├── go.mod
├── main.go              # CLI entry point, command routing
├── cmd/
│   ├── daemon.go        # Event listener, state manager
│   └── dispatch.go      # Workspace/window dispatch commands
├── ipc/
│   ├── socket.go        # Hyprland command socket (send/receive)
│   └── events.go        # Hyprland event socket (listener, parser)
└── workspace/
    ├── mapper.go         # Virtual <-> real workspace ID mapping
    └── state.go          # Monitor/workspace state tracking
```

## Edge Cases

- **Monitor disconnect:** Windows on that monitor's workspaces need to be moved to remaining monitors
- **Monitor reconnect:** Restore workspace assignments, recover rogue windows
- **Workspace already exists on wrong monitor:** Use `moveworkspacetomonitor` to relocate it
- **Monitor ordering:** Assign monitor indices based on config order, not Hyprland IDs (which can change)

## Configuration

Phase 3 introduces a simple config file for user preferences. State is always read live via Hyprland IPC — the config only stores user overrides.

**Approach:** Simple key-value config file (`~/.config/hypr/hyprgo-split-ws.conf`), not hyprlang. IPC provides live monitor/workspace state, so the config only needs to cover user preferences like monitor ordering and workspace counts.

**Future consideration:** If the hyprgo tool suite grows (hyprgo-net, hyprgo-setup, etc.), a shared Go-native hyprlang parser could be built as a common package to read directly from a user's Hyprland config. Not worth building until multiple tools need it.

Example config:
```conf
# Monitor ordering (overrides alphabetical default)
monitor_order = DP-1, DP-3, DP-2

# Workspaces per monitor (default: 10)
workspaces_per_monitor = 10
```

## Phase 1 (MVP) — COMPLETE

1. Workspace offsetting — `workspace` and `movetoworkspace` commands
2. Direct Hyprland IPC (no daemon yet, just calculate and dispatch)
3. Keybind integration

## Phase 2 — COMPLETE

4. Daemon with event listener for state tracking
5. Workspace renaming for Waybar display
6. Monitor hotplug handling

## Phase 3

7. Config file for monitor ordering and workspace counts
8. `changemonitor` command for moving windows between monitors
9. Rogue window recovery

## Phase 4

10. Custom Waybar module — replace the built-in `hyprland/workspaces` with a custom IPC module that queries the daemon for per-monitor workspace state, active indicators, and window counts
11. GUI management panel — a graphical interface (GTK4 or similar) for:
    - Visualizing workspace layout across monitors
    - Drag-and-drop window movement between workspaces/monitors
    - Monitor configuration (ordering, workspace count per monitor)
    - Live preview of workspace contents
