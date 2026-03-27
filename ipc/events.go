package ipc

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// Event represents a parsed Hyprland IPC event.
type Event struct {
	Name string
	Data string
}

// EventListener connects to the Hyprland event socket and streams parsed events.
type EventListener struct {
	conn    net.Conn
	scanner *bufio.Scanner
}

// NewEventListener creates a new listener connected to the Hyprland event socket.
func NewEventListener() (*EventListener, error) {
	sockPath, err := EventSocketPath()
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to event socket: %w", err)
	}

	return &EventListener{
		conn:    conn,
		scanner: bufio.NewScanner(conn),
	}, nil
}

// Next blocks until the next event is available and returns it.
// Returns an error if the connection is closed or fails.
func (l *EventListener) Next() (Event, error) {
	if !l.scanner.Scan() {
		err := l.scanner.Err()
		if err == nil {
			return Event{}, fmt.Errorf("event socket closed")
		}
		return Event{}, fmt.Errorf("event socket error: %w", err)
	}

	line := l.scanner.Text()
	name, data, found := strings.Cut(line, ">>")
	if !found {
		return Event{}, fmt.Errorf("malformed event: %s", line)
	}

	return Event{Name: name, Data: data}, nil
}

// Close closes the event socket connection.
func (l *EventListener) Close() error {
	return l.conn.Close()
}
