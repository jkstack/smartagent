package hostinfo

import "github.com/shirou/gopsutil/v3/host"

// Info host info
func Info() (*InfoStat, error) {
	info, err := host.Info()
	if err != nil {
		return nil, err
	}
	return &InfoStat{
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		KernelArch:      info.KernelArch,
	}, nil
}
