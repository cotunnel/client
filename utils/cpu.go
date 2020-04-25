package utils

import (
	"github.com/shirou/gopsutil/cpu"
	"time"
)

func GetCpuUsage() float64 {
	cpuStat, err := cpu.Percent(1*time.Second, false)
	if err != nil {
		return 0
	}

	return cpuStat[0]
}
