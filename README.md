# hyprgo-split-ws

Per-monitor workspace management for [Hyprland](https://hyprland.org/), written in Go. Gives each monitor its own independent set of workspaces (1-10), similar to KDE Plasma's virtual desktops per screen.

## Why?

Hyprland has global workspaces — workspace 1 exists on one monitor only, and switching to it jumps you there. This tool makes `Super+1` go to workspace 1 on **whichever monitor you're focused on**, independently per monitor.

## How it works

Each monitor is assigned a range of Hyprland workspace IDs via offsetting:

| Monitor | Virtual 1-10 | Real Hyprland IDs |
|---------|-------------|-------------------|
| Monitor 0 | 1-10 | 1-10 |
| Monitor 1 | 1-10 | 11-20 |
| Monitor 2 | 1-10 | 21-30 |

A background daemon renames workspaces so Waybar displays the virtual number (1-10), not the real ID.

```
Keybinds ──> hyprgo-split-ws CLI ──> Hyprland IPC
                                          ^
Hyprland events ──> hyprgo-split-ws daemon (renames, hotplug recovery)
```

## Installation

### From source

Requires Go 1.21+.

```bash
git clone https://github.com/pharomwinters/hyprgo-split-ws.git
cd hyprgo-split-ws
go build -o hyprgo-split-ws .
cp hyprgo-split-ws ~/.local/bin/
```

## Configuration

Create `~/.config/hypr/hyprgo-split-ws.conf`:

```conf
# Monitor ordering (overrides alphabetical default)
monitor_order = DP-3, DP-2, DP-1

# Workspaces per monitor (default: 10)
workspaces_per_monitor = 10
```

If no config file exists, monitors are ordered alphabetically with 10 workspaces each.

## Hyprland setup

### Autostart the daemon

Add to your Hyprland autostart config:

```conf
exec-once = hyprgo-split-ws daemon &
```

### Keybinds

Replace your workspace keybinds:

```conf
# Switch workspaces (per-monitor)
bind = $mainMod, 1, exec, hyprgo-split-ws workspace 1
bind = $mainMod, 2, exec, hyprgo-split-ws workspace 2
bind = $mainMod, 3, exec, hyprgo-split-ws workspace 3
bind = $mainMod, 4, exec, hyprgo-split-ws workspace 4
bind = $mainMod, 5, exec, hyprgo-split-ws workspace 5
bind = $mainMod, 6, exec, hyprgo-split-ws workspace 6
bind = $mainMod, 7, exec, hyprgo-split-ws workspace 7
bind = $mainMod, 8, exec, hyprgo-split-ws workspace 8
bind = $mainMod, 9, exec, hyprgo-split-ws workspace 9
bind = $mainMod, 0, exec, hyprgo-split-ws workspace 10

# Move window to workspace (per-monitor)
bind = $mainMod SHIFT, 1, exec, hyprgo-split-ws movetoworkspace 1
bind = $mainMod SHIFT, 2, exec, hyprgo-split-ws movetoworkspace 2
bind = $mainMod SHIFT, 3, exec, hyprgo-split-ws movetoworkspace 3
bind = $mainMod SHIFT, 4, exec, hyprgo-split-ws movetoworkspace 4
bind = $mainMod SHIFT, 5, exec, hyprgo-split-ws movetoworkspace 5
bind = $mainMod SHIFT, 6, exec, hyprgo-split-ws movetoworkspace 6
bind = $mainMod SHIFT, 7, exec, hyprgo-split-ws movetoworkspace 7
bind = $mainMod SHIFT, 8, exec, hyprgo-split-ws movetoworkspace 8
bind = $mainMod SHIFT, 9, exec, hyprgo-split-ws movetoworkspace 9
bind = $mainMod SHIFT, 0, exec, hyprgo-split-ws movetoworkspace 10

# Move window between monitors
bind = $mainMod SHIFT, left, exec, hyprgo-split-ws changemonitor prev
bind = $mainMod SHIFT, right, exec, hyprgo-split-ws changemonitor next
```

### Waybar

Set `all-outputs: false` and use `{name}` format so each monitor's bar shows only its own workspaces with virtual numbers:

```jsonc
"hyprland/workspaces": {
    "on-click": "activate",
    "format": "{name}",
    "all-outputs": false,
    "disable-scroll": false,
    "active-only": false
}
```

## CLI reference

| Command | Description |
|---------|-------------|
| `hyprgo-split-ws workspace <1-N>` | Switch to virtual workspace on focused monitor |
| `hyprgo-split-ws movetoworkspace <1-N>` | Move active window to virtual workspace and follow |
| `hyprgo-split-ws movetoworkspacesilent <1-N>` | Move active window without following |
| `hyprgo-split-ws changemonitor <next\|prev>` | Move focused window to next/prev monitor |
| `hyprgo-split-ws daemon` | Start the event listener daemon |

## Features

- **Per-monitor workspaces** — each monitor gets its own workspace 1-10
- **Waybar integration** — daemon renames workspaces so Waybar displays virtual numbers
- **Monitor hotplug** — rebuilds monitor mapping when monitors are added/removed
- **Rogue window recovery** — orphaned windows are moved to a fallback workspace on monitor disconnect
- **Configurable monitor ordering** — override the default alphabetical sort
- **Zero dependencies** — single static Go binary, communicates directly via Hyprland's Unix IPC sockets

## License

MIT
