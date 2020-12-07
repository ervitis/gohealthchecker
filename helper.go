package gohealthchecker

import (
	"fmt"
	"github.com/ervitis/goprocinfo/linux"
	"net"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	Running  string = "R"
	Sleeping        = "S"
	Waiting         = "D"
	Zombie          = "Z"
	Stopped         = "T"
	Traced          = "t"
	Dead            = "X"
)

type (
	ipAddressInfo struct {
		ipAddress string
	}

	mappedProcessesStatus map[string]bool

	memoryData struct {
		total     uint64
		free      uint64
		available uint64
	}

	SystemInformation struct {
		mappedProcessStatus mappedProcessesStatus
		processStatus       string
		processActive       bool
		pid                 uint64
		startTime           time.Time
		memory              *memoryData
		ipAddress           *ipAddressInfo
		runtimeVersion      string
		canAcceptWork       bool

		mtx sync.Mutex
	}
)

// GetIpAddress returns the ip address of the machine
// it returns an IpAddressInfo object which stores the IpAddress
// we are using the net.Dial function to get the origin of the connection to the host we want to connect
func (info *SystemInformation) getIpAddress() *ipAddressInfo {
	conn, _ := net.Dial("udp", "1.1.1.1:80")
	defer conn.Close()

	return &ipAddressInfo{ipAddress: conn.LocalAddr().(*net.UDPAddr).IP.String()}
}

// GetRuntimeVersion returns the Golang version which was compiled
func (info *SystemInformation) getRuntimeVersion() string {
	return runtime.Version()
}

func (info *SystemInformation) initMappedStatus() {
	if info.mappedProcessStatus != nil {
		return
	}

	info.mtx.Lock()
	defer info.mtx.Unlock()
	info.mappedProcessStatus = make(mappedProcessesStatus)
	info.mappedProcessStatus[Running] = true
	info.mappedProcessStatus[Sleeping] = true
	info.mappedProcessStatus[Traced] = false
	info.mappedProcessStatus[Waiting] = false
	info.mappedProcessStatus[Zombie] = false
	info.mappedProcessStatus[Stopped] = false
	info.mappedProcessStatus[Dead] = false
}

// GetSystemInfo initialize the data of the current process running
// it returns an error if it cannot locate the *nix files to get the data
func (info *SystemInformation) GetSystemInfo() error {
	info.initMappedStatus()

	info.mtx.Lock()
	defer info.mtx.Unlock()

	processStatus, err := linux.ReadProcessStat(fmt.Sprintf("/proc/%d/stat", os.Getpid()))
	if err != nil {
		return err
	}

	memInfo, err := linux.ReadMemInfo("/proc/meminfo")
	if err != nil {
		return err
	}

	info.pid = processStatus.Pid
	info.processActive = info.mappedProcessStatus[processStatus.State]
	info.processStatus = processStatus.State
	info.memory = &memoryData{
		total:     memInfo.MemTotal,
		free:      memInfo.MemFree,
		available: memInfo.MemAvailable,
	}
	info.runtimeVersion = info.getRuntimeVersion()
	info.ipAddress = info.getIpAddress()
	info.canAcceptWork = info.processActive && (info.memory.available > 0)
	return nil
}
