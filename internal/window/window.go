package window

import (
	"context"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/borankux/gops/pkg/types"
)

// GetOpenWindows returns a list of open windows
func GetOpenWindows(ctx context.Context) ([]types.WindowInfo, error) {
	switch runtime.GOOS {
	case "darwin":
		return getMacOSWindows(ctx)
	case "linux":
		return getLinuxWindows(ctx)
	case "windows":
		return getWindowsWindows(ctx)
	default:
		return nil, nil
	}
}

// getMacOSWindows gets windows on macOS using osascript
func getMacOSWindows(ctx context.Context) ([]types.WindowInfo, error) {
	script := `
		tell application "System Events"
			set windowList to {}
			repeat with proc in every process
				if background only of proc is false then
					try
						set procName to name of proc
						set procPID to unix id of proc
						set winList to windows of proc
						repeat with win in winList
							try
								set winTitle to title of win
								if winTitle is not "" then
									set end of windowList to {procName, winTitle, procPID}
								end if
							end try
						end repeat
					end try
				end if
			end repeat
		end tell
		return windowList
	`

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), ", ")
	var windows []types.WindowInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "missing value") {
			continue
		}

		// Check if this is a new window entry (contains process name)
		// Format: processName, windowTitle, PID
		if strings.Contains(line, "text") {
			// This is AppleScript text item, skip parsing details
			continue
		}

		// Try to parse as: "processName", "windowTitle", PID
		// Simple parsing: split by comma and extract
		parts := strings.Split(line, ",")
		if len(parts) >= 2 {
			appName := strings.TrimSpace(strings.Trim(parts[0], "\""))
			title := strings.TrimSpace(strings.Trim(strings.Join(parts[1:len(parts)-1], ","), "\""))

			var pid int32
			if len(parts) >= 3 {
				pidStr := strings.TrimSpace(parts[len(parts)-1])
				if p, err := strconv.ParseInt(pidStr, 10, 32); err == nil {
					pid = int32(p)
				} else {
					pid = getPIDForApp(ctx, appName)
				}
			} else {
				pid = getPIDForApp(ctx, appName)
			}

			if appName != "" && title != "" {
				windows = append(windows, types.WindowInfo{
					Title:   title,
					PID:     pid,
					Process: appName,
					AppName: appName,
				})
			}
		}
	}

	// If parsing failed, try alternative approach
	if len(windows) == 0 {
		return getMacOSWindowsAlt(ctx)
	}

	return windows, nil
}

// getMacOSWindowsAlt is an alternative method to get windows on macOS
func getMacOSWindowsAlt(ctx context.Context) ([]types.WindowInfo, error) {
	// Use AppleScript with better formatting
	script := `tell application "System Events"
		set windowList to {}
		repeat with proc in every process
			if background only of proc is false then
				try
					set procName to name of proc
					set procPID to unix id of proc
					set winList to windows of proc
					repeat with win in winList
						try
							set winTitle to title of win
							if winTitle is not "" then
								set end of windowList to procName & "|" & winTitle & "|" & procPID
							end if
						end try
					end repeat
				end try
			end if
		end repeat
	end tell
	return windowList`

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var windows []types.WindowInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			appName := strings.TrimSpace(parts[0])
			title := strings.TrimSpace(parts[1])
			pidStr := strings.TrimSpace(parts[2])

			pid, err := strconv.ParseInt(pidStr, 10, 32)
			if err != nil {
				pid = int64(getPIDForApp(ctx, appName))
			}

			if appName != "" && title != "" {
				windows = append(windows, types.WindowInfo{
					Title:   title,
					PID:     int32(pid),
					Process: appName,
					AppName: appName,
				})
			}
		}
	}

	return windows, nil
}

// getLinuxWindows gets windows on Linux using wmctrl
func getLinuxWindows(ctx context.Context) ([]types.WindowInfo, error) {
	cmd := exec.CommandContext(ctx, "wmctrl", "-lp")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var windows []types.WindowInfo

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 5 {
			pidStr := parts[2]
			pid, _ := strconv.ParseInt(pidStr, 10, 32)
			title := strings.Join(parts[4:], " ")

			// Get process name
			procName := getProcessName(ctx, int32(pid))

			windows = append(windows, types.WindowInfo{
				Title:   title,
				PID:     int32(pid),
				Process: procName,
				AppName: procName,
			})
		}
	}

	return windows, nil
}

// getWindowsWindows gets windows on Windows using PowerShell
func getWindowsWindows(ctx context.Context) ([]types.WindowInfo, error) {
	psScript := `
		Get-Process | Where-Object {$_.MainWindowTitle -ne ""} | ForEach-Object {
			$_.Id.ToString() + "|" + $_.ProcessName + "|" + $_.MainWindowTitle
		}
	`

	cmd := exec.CommandContext(ctx, "powershell", "-Command", psScript)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var windows []types.WindowInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			pidStr := strings.TrimSpace(parts[0])
			processName := strings.TrimSpace(parts[1])
			title := strings.TrimSpace(parts[2])

			pid, err := strconv.ParseInt(pidStr, 10, 32)
			if err != nil {
				continue
			}

			windows = append(windows, types.WindowInfo{
				Title:   title,
				PID:     int32(pid),
				Process: processName,
				AppName: processName,
			})
		}
	}

	return windows, nil
}

func getPIDForApp(ctx context.Context, appName string) int32 {
	cmd := exec.CommandContext(ctx, "pgrep", "-f", appName)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	pidStr := strings.TrimSpace(string(output))
	pid, _ := strconv.ParseInt(pidStr, 10, 32)
	return int32(pid)
}

func getProcessName(ctx context.Context, pid int32) string {
	// Use ps or read from /proc
	if runtime.GOOS == "linux" {
		cmd := exec.CommandContext(ctx, "ps", "-p", strconv.FormatInt(int64(pid), 10), "-o", "comm=")
		output, err := cmd.Output()
		if err == nil {
			return strings.TrimSpace(string(output))
		}
	}
	return ""
}
