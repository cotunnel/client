package utils

import "github.com/shirou/gopsutil/mem"

func GetMemoryUsage() (int, int, float64) {
	virtualMemoryStat, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, 0
	}

	return int(virtualMemoryStat.Used / 1024 / 1024), int(virtualMemoryStat.Total / 1024 / 1024), virtualMemoryStat.UsedPercent
}
