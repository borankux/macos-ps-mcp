package service

import (
	"context"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/allintech/gops/internal/resource"
	"github.com/allintech/gops/internal/utils"
	"github.com/allintech/gops/pkg/types"
	"github.com/shirou/gopsutil/v3/process"
)

// GetServices returns a list of system services with resource usage
func GetServices(ctx context.Context) ([]types.ServiceInfo, error) {
	switch runtime.GOOS {
	case "darwin":
		return getMacOSServices(ctx)
	case "linux":
		return getLinuxServices(ctx)
	case "windows":
		return getWindowsServices(ctx)
	default:
		return nil, nil
	}
}

// getMacOSServices gets services on macOS using launchctl
func getMacOSServices(ctx context.Context) ([]types.ServiceInfo, error) {
	cmd := exec.CommandContext(ctx, "launchctl", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var services []types.ServiceInfo

	// Skip header line
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		pidStr := fields[0]
		status := fields[1]
		name := strings.Join(fields[2:], " ")

		// Skip if pid is 0 or -1 (not running)
		pid, err := strconv.ParseInt(pidStr, 10, 32)
		if err != nil || pid <= 0 {
			services = append(services, types.ServiceInfo{
				Name:   name,
				Status: status,
			})
			continue
		}

		// Get resource usage
		usage, err := resource.GetProcessResourceUsage(ctx, int32(pid))
		if err != nil {
			services = append(services, types.ServiceInfo{
				Name:   name,
				Status: status,
				PID:    int32(pid),
			})
			continue
		}

		services = append(services, types.ServiceInfo{
			Name:          name,
			Status:        status,
			PID:           int32(pid),
			CPUPercent:    usage.CPUPercent,
			MemoryPercent: usage.MemoryPercent,
			MemoryHuman:   usage.MemoryHuman,
			CPUHuman:      usage.CPUHuman,
		})
	}

	return services, nil
}

// getLinuxServices gets services on Linux using systemctl
func getLinuxServices(ctx context.Context) ([]types.ServiceInfo, error) {
	cmd := exec.CommandContext(ctx, "systemctl", "list-units", "--type=service", "--no-pager", "--no-legend")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var services []types.ServiceInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		name := strings.ReplaceAll(fields[0], ".service", "")
		status := fields[2] // loaded, active, etc.

		// Try to get PID from systemctl show
		pidCmd := exec.CommandContext(ctx, "systemctl", "show", "--property=MainPID", "--value", fields[0])
		pidOutput, err := pidCmd.Output()
		var pid int32
		if err == nil {
			pidStr := strings.TrimSpace(string(pidOutput))
			if pidStr != "0" && pidStr != "" {
				if p, err := strconv.ParseInt(pidStr, 10, 32); err == nil && p > 0 {
					pid = int32(p)

					// Get resource usage
					usage, err := resource.GetProcessResourceUsage(ctx, pid)
					if err == nil {
						services = append(services, types.ServiceInfo{
							Name:          name,
							Status:        status,
							PID:           pid,
							CPUPercent:    usage.CPUPercent,
							MemoryPercent: usage.MemoryPercent,
							MemoryHuman:   usage.MemoryHuman,
							CPUHuman:      usage.CPUHuman,
						})
						continue
					}
				}
			}
		}

		services = append(services, types.ServiceInfo{
			Name:   name,
			Status: status,
			PID:    pid,
		})
	}

	return services, nil
}

// getWindowsServices gets services on Windows
func getWindowsServices(ctx context.Context) ([]types.ServiceInfo, error) {
	psScript := `
		Get-Service | Select-Object Name, Status, @{Name='PID';Expression={(Get-WmiObject Win32_Service -Filter "Name='$($_.Name)'").ProcessId}} | ConvertTo-Json
	`

	cmd := exec.CommandContext(ctx, "powershell", "-Command", psScript)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var services []types.ServiceInfo
	// Parse JSON output - simplified for now
	// In production, use proper JSON parsing
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	_ = lines

	return services, nil
}

