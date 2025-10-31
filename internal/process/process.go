package process

import (
	"context"
	"runtime"
	"sort"
	"strings"

	"github.com/allintech/gops/pkg/types"
	"github.com/shirou/gopsutil/v3/process"
)

// GetUserApplications returns a list of non-system user applications
func GetUserApplications(ctx context.Context) ([]types.ProcessInfo, error) {
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	var userProcs []types.ProcessInfo
	systemPrefixes := getSystemPrefixes()

	for _, p := range procs {
		name, err := p.NameWithContext(ctx)
		if err != nil {
			continue
		}

		// Skip system processes
		if isSystemProcess(name, systemPrefixes) {
			continue
		}

		// Skip kernel processes
		exe, err := p.ExeWithContext(ctx)
		if err != nil {
			// No executable path might indicate kernel process
			continue
		}

		pid := p.Pid
		status := ""
		if st, err := p.StatusWithContext(ctx); err == nil {
			status = strings.Join(st, ",")
		}

		username := ""
		if u, err := p.UsernameWithContext(ctx); err == nil {
			username = u
		}

		startTime := ""
		if st, err := p.CreateTimeWithContext(ctx); err == nil {
			startTime = formatTime(st)
		}

		userProcs = append(userProcs, types.ProcessInfo{
			PID:       pid,
			Name:      name,
			Path:      exe,
			Status:    status,
			User:      username,
			StartTime: startTime,
		})
	}

	// Sort by PID
	sort.Slice(userProcs, func(i, j int) bool {
		return userProcs[i].PID < userProcs[j].PID
	})

	return userProcs, nil
}

// getSystemPrefixes returns OS-specific system process prefixes
func getSystemPrefixes() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"com.apple", "kernel", "WindowServer", "launchd", "syspolicyd", "trustd"}
	case "linux":
		return []string{"[", "kthreadd", "ksoftirqd", "migration", "rcu_", "systemd", "init"}
	case "windows":
		return []string{"System", "smss", "csrss", "winlogon", "services", "lsass", "svchost", "spoolsv", "SearchIndexer"}
	default:
		return []string{"kernel", "init", "system"}
	}
}

// isSystemProcess checks if a process is a system process
func isSystemProcess(name string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func formatTime(timestamp int64) string {
	return ""
	// Can be expanded to format timestamp to readable date
}

