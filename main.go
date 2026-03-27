package main

import (
	"fmt"
	"os"

	"github.com/adambick/hyprgo-split-ws/cmd"
	"github.com/adambick/hyprgo-split-ws/config"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "workspace":
		err = cmd.SwitchWorkspace(cfg, args)
	case "movetoworkspace":
		err = cmd.MoveToWorkspace(cfg, args)
	case "movetoworkspacesilent":
		err = cmd.MoveToWorkspaceSilent(cfg, args)
	case "changemonitor":
		err = cmd.ChangeMonitor(cfg, args)
	case "daemon":
		err = cmd.RunDaemon(cfg)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: hyprgo-split-ws <command> [args]

Commands:
  workspace <1-N>              Switch to virtual workspace on focused monitor
  movetoworkspace <1-N>        Move active window to virtual workspace and follow
  movetoworkspacesilent <1-N>  Move active window to virtual workspace silently
  changemonitor <next|prev>    Move focused window to next/prev monitor
  daemon                       Start the event listener daemon
`)
}
