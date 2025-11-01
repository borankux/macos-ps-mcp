package resource

import (
	"context"

	"github.com/borankux/gops/internal/utils"
	"github.com/borankux/gops/pkg/types"
	"github.com/shirou/gopsutil/v3/process"
)

// GetProcessResourceUsage returns resource usage for a specific process
func GetProcessResourceUsage(ctx context.Context, pid int32) (*types.ResourceUsage, error) {
	p, err := process.NewProcessWithContext(ctx, pid)
	if err != nil {
		return nil, err
	}

	name, _ := p.NameWithContext(ctx)
	cpuPercent, _ := p.CPUPercentWithContext(ctx)
	memPercent, _ := p.MemoryPercentWithContext(ctx)

	memInfo, err := p.MemoryInfoWithContext(ctx)
	if err != nil {
		return nil, err
	}

	var memoryRSS uint64
	var memoryVMS uint64
	if memInfo != nil {
		memoryRSS = memInfo.RSS
		memoryVMS = memInfo.VMS
	}

	memoryHuman := utils.FormatBytes(memoryRSS)
	cpuHuman := utils.FormatCPU(cpuPercent)

	threads, _ := p.NumThreadsWithContext(ctx)
	openFiles, _ := p.NumFDsWithContext(ctx)

	return &types.ResourceUsage{
		PID:           pid,
		Name:          name,
		CPUPercent:    cpuPercent,
		MemoryPercent: memPercent,
		MemoryRSS:     memoryRSS,
		MemoryVMS:     memoryVMS,
		MemoryHuman:   memoryHuman,
		CPUHuman:      cpuHuman,
		Threads:       threads,
		OpenFiles:     openFiles,
	}, nil
}

// GetTopProcesses returns top N processes by CPU or memory
func GetTopProcesses(ctx context.Context, limit int, sortBy string) ([]types.ResourceUsage, error) {
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, err
	}

	var usages []types.ResourceUsage
	for _, p := range procs {
		pid := p.Pid
		usage, err := GetProcessResourceUsage(ctx, pid)
		if err != nil {
			continue
		}
		usages = append(usages, *usage)
	}

	// Sort by CPU or Memory
	if sortBy == "cpu" {
		for i := 0; i < len(usages)-1; i++ {
			for j := i + 1; j < len(usages); j++ {
				if usages[i].CPUPercent < usages[j].CPUPercent {
					usages[i], usages[j] = usages[j], usages[i]
				}
			}
		}
	} else {
		// Sort by memory
		for i := 0; i < len(usages)-1; i++ {
			for j := i + 1; j < len(usages); j++ {
				if usages[i].MemoryRSS < usages[j].MemoryRSS {
					usages[i], usages[j] = usages[j], usages[i]
				}
			}
		}
	}

	if limit > 0 && limit < len(usages) {
		usages = usages[:limit]
	}

	return usages, nil
}
