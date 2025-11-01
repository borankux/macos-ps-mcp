package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/borankux/gops/internal/port"
	"github.com/borankux/gops/internal/process"
	"github.com/borankux/gops/internal/resource"
	"github.com/borankux/gops/internal/service"
	"github.com/borankux/gops/internal/window"
	"github.com/borankux/gops/pkg/types"
	"github.com/jedib0t/go-pretty/v6/table"
)

// DisplayProcesses displays processes in a formatted table
func DisplayProcesses(ctx context.Context) error {
	procs, err := process.GetUserApplications(ctx)
	if err != nil {
		return err
	}

	fmt.Println("📱 User Applications")
	fmt.Println()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"🔢 PID", "📛 Name", "👤 User", "📍 Path"})
	t.Style().Options.SeparateRows = true

	for _, p := range procs {
		t.AppendRow(table.Row{
			fmt.Sprintf("%d", p.PID),
			p.Name,
			p.User,
			truncateString(p.Path, 50),
		})
	}

	t.AppendFooter(table.Row{"Total", len(procs), "", ""})
	t.Render()

	return nil
}

// DisplayWindows displays open windows in a formatted table
func DisplayWindows(ctx context.Context) error {
	windows, err := window.GetOpenWindows(ctx)
	if err != nil {
		return err
	}

	fmt.Println("🪟 Open Windows")
	fmt.Println()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"🪟 Title", "🔢 PID", "📛 Process"})
	t.Style().Options.SeparateRows = true

	for _, w := range windows {
		t.AppendRow(table.Row{
			truncateString(w.Title, 60),
			fmt.Sprintf("%d", w.PID),
			w.Process,
		})
	}

	t.AppendFooter(table.Row{"Total", len(windows), ""})
	t.Render()

	return nil
}

// DisplayPorts displays open ports in a formatted table
func DisplayPorts(ctx context.Context, portFilter string, pidFilter string) error {
	var ports []types.PortInfo
	var err error

	if portFilter != "" {
		portNum, parseErr := strconv.ParseUint(portFilter, 10, 32)
		if parseErr != nil {
			return fmt.Errorf("invalid port number: %w", parseErr)
		}
		ports, err = port.GetPortInfoByPort(ctx, uint32(portNum))
	} else if pidFilter != "" {
		pid, parseErr := strconv.ParseInt(pidFilter, 10, 32)
		if parseErr != nil {
			return fmt.Errorf("invalid PID: %w", parseErr)
		}
		ports, err = port.GetPortsByPID(ctx, int32(pid))
	} else {
		ports, err = port.GetOpenPorts(ctx)
	}

	if err != nil {
		return err
	}

	fmt.Println("🌐 Open Ports")
	fmt.Println()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"🔌 Port", "📡 Protocol", "🔢 PID", "📛 Process", "📍 Path"})
	t.Style().Options.SeparateRows = true

	for _, p := range ports {
		t.AppendRow(table.Row{
			fmt.Sprintf("%d", p.Port),
			p.Protocol,
			fmt.Sprintf("%d", p.PID),
			p.Name,
			truncateString(p.Path, 50),
		})
	}

	t.AppendFooter(table.Row{"Total", "", "", "", len(ports)})
	t.Render()

	return nil
}

// DisplayResourceUsage displays resource usage for a process
func DisplayResourceUsage(ctx context.Context, pid int32) error {
	usage, err := resource.GetProcessResourceUsage(ctx, pid)
	if err != nil {
		return err
	}

	fmt.Printf("📊 Resource Usage for Process %d (%s)\n", usage.PID, usage.Name)
	fmt.Println()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Metric", "Value"})
	t.Style().Options.SeparateRows = true

	t.AppendRow(table.Row{"🔢 PID", fmt.Sprintf("%d", usage.PID)})
	t.AppendRow(table.Row{"📛 Name", usage.Name})
	t.AppendRow(table.Row{"💻 CPU Usage", usage.CPUHuman})
	t.AppendRow(table.Row{"🧠 Memory Usage", usage.MemoryHuman})
	t.AppendRow(table.Row{"📈 Memory %", fmt.Sprintf("%.2f%%", usage.MemoryPercent)})
	t.AppendRow(table.Row{"🧵 Threads", fmt.Sprintf("%d", usage.Threads)})
	t.AppendRow(table.Row{"📂 Open Files", fmt.Sprintf("%d", usage.OpenFiles)})

	t.Render()

	return nil
}

// DisplayServices displays services in a formatted table
func DisplayServices(ctx context.Context) error {
	services, err := service.GetServices(ctx)
	if err != nil {
		return err
	}

	fmt.Println("⚙️  System Services")
	fmt.Println()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"📛 Name", "🟢 Status", "🔢 PID", "💻 CPU", "🧠 Memory"})
	t.Style().Options.SeparateRows = true

	for _, s := range services {
		pidStr := "-"
		if s.PID > 0 {
			pidStr = fmt.Sprintf("%d", s.PID)
		}

		cpuStr := "-"
		memStr := "-"
		if s.PID > 0 && s.CPUPercent > 0 {
			cpuStr = s.CPUHuman
			memStr = s.MemoryHuman
		}

		statusEmoji := "🟢"
		if s.Status != "running" && s.Status != "active" {
			statusEmoji = "🔴"
		}

		t.AppendRow(table.Row{
			s.Name,
			fmt.Sprintf("%s %s", statusEmoji, s.Status),
			pidStr,
			cpuStr,
			memStr,
		})
	}

	t.AppendFooter(table.Row{"Total", "", "", "", len(services)})
	t.Render()

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
