package service

import (
	"context"
	"encoding/json"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/allintech/gops/internal/resource"
	"github.com/allintech/gops/pkg/types"
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
		Get-Service | ForEach-Object {
			$pid = (Get-WmiObject Win32_Service -Filter "Name='$($_.Name)'" -ErrorAction SilentlyContinue).ProcessId
			if ($pid -eq $null) { $pid = 0 }
			[PSCustomObject]@{
				Name = $_.Name
				Status = $_.Status.ToString()
				PID = $pid
			}
		} | ConvertTo-Json -Compress
	`

	cmd := exec.CommandContext(ctx, "powershell", "-Command", psScript)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var services []types.ServiceInfo

	// Parse JSON output
	var serviceObjs []struct {
		Name   string `json:"Name"`
		Status string `json:"Status"`
		PID    int    `json:"PID"`
	}

	if err := json.Unmarshal(output, &serviceObjs); err != nil {
		// If array parsing fails, try single object
		var serviceObj struct {
			Name   string `json:"Name"`
			Status string `json:"Status"`
			PID    int    `json:"PID"`
		}
		if err2 := json.Unmarshal(output, &serviceObj); err2 == nil {
			serviceObjs = []struct {
				Name   string `json:"Name"`
				Status string `json:"Status"`
				PID    int    `json:"PID"`
			}{serviceObj}
		} else {
			return nil, err
		}
	}

	for _, s := range serviceObjs {
		serviceInfo := types.ServiceInfo{
			Name:   s.Name,
			Status: strings.ToLower(s.Status),
			PID:    int32(s.PID),
		}

		// Get resource usage if PID is available
		if s.PID > 0 {
			usage, err := resource.GetProcessResourceUsage(ctx, int32(s.PID))
			if err == nil {
				serviceInfo.CPUPercent = usage.CPUPercent
				serviceInfo.MemoryPercent = usage.MemoryPercent
				serviceInfo.MemoryHuman = usage.MemoryHuman
				serviceInfo.CPUHuman = usage.CPUHuman
			}
		}

		services = append(services, serviceInfo)
	}

	return services, nil
}
