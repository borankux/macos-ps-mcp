package port

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/allintech/gops/pkg/types"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// GetOpenPorts returns a list of open ports with associated processes
func GetOpenPorts(ctx context.Context) ([]types.PortInfo, error) {
	connections, err := net.ConnectionsWithContext(ctx, "inet")
	if err != nil {
		return nil, err
	}

	portMap := make(map[string]*types.PortInfo)

	for _, conn := range connections {
		// Only show listening connections (ports that are open and listening)
		if conn.Status != "LISTEN" {
			continue
		}

		port := conn.Laddr.Port
		if port == 0 {
			continue
		}

		key := fmt.Sprintf("%s:%d", conn.Laddr.IP, port)

		// Get process info
		var procName string
		var exePath string
		if conn.Pid > 0 {
			p, err := process.NewProcessWithContext(ctx, conn.Pid)
			if err == nil {
				if name, err := p.NameWithContext(ctx); err == nil {
					procName = name
				}
				if exe, err := p.ExeWithContext(ctx); err == nil {
					exePath = exe
				}
			}
		}

		protocol := getProtocol(conn)
		portInfo := &types.PortInfo{
			Port:     uint32(port),
			Protocol: protocol,
			PID:      conn.Pid,
			Name:     procName,
			Path:     exePath,
			State:    conn.Status,
			LocalIP:  conn.Laddr.IP,
		}

		// Store port info (we only get LISTEN connections now)
		if existing, exists := portMap[key]; exists {
			// If port exists, keep the existing one unless this has better info
			if existing.Name == "" && procName != "" {
				portMap[key] = portInfo
			}
		} else {
			portMap[key] = portInfo
		}
	}

	var ports []types.PortInfo
	for _, portInfo := range portMap {
		ports = append(ports, *portInfo)
	}

	// Sort by port number
	for i := 0; i < len(ports)-1; i++ {
		for j := i + 1; j < len(ports); j++ {
			if ports[i].Port > ports[j].Port {
				ports[i], ports[j] = ports[j], ports[i]
			}
		}
	}

	return ports, nil
}

// getProtocol determines protocol from connection
func getProtocol(conn net.ConnectionStat) string {
	// Try to determine from local port or system call
	// For now, default to TCP, can be enhanced
	if strings.Contains(strings.ToLower(conn.Status), "udp") {
		return "UDP"
	}
	return "TCP"
}

// GetPortInfoByPort returns information about a specific port
func GetPortInfoByPort(ctx context.Context, port uint32) ([]types.PortInfo, error) {
	allPorts, err := GetOpenPorts(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []types.PortInfo
	for _, p := range allPorts {
		if p.Port == port {
			filtered = append(filtered, p)
		}
	}

	return filtered, nil
}

// GetPortsByPID returns ports used by a specific process
func GetPortsByPID(ctx context.Context, pid int32) ([]types.PortInfo, error) {
	allPorts, err := GetOpenPorts(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []types.PortInfo
	for _, p := range allPorts {
		if p.PID == pid {
			filtered = append(filtered, p)
		}
	}

	return filtered, nil
}

