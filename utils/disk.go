package utils

import (
	"github.com/ricochet2200/go-disk-usage/du"
	"runtime"
)

type DiskStatus struct {
	All   int
	Used  int
	Free  int
	Avail int
}

func GetDiskUsage() (disk DiskStatus) {

	path := "/"

	if runtime.GOOS == "windows" {
		path = "C:\\"
	}

	usage := du.NewDiskUsage(path)

	// Convert to mb
	disk.All = int(usage.Size() / 1024 / 1014)
	disk.Avail = int(usage.Available() / 1024 / 1024)
	disk.Free = int(usage.Free() / 1024 / 1024)
	disk.Used = int(usage.Used() / 1024 / 1024)
	return
}
