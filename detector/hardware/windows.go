//go:build windows
// +build windows

package hardware

import (
	"fmt"
	"net"
	"runtime"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32 = windows.MustLoadDLL("kernel32.dll")
	advapi32 = windows.MustLoadDLL("advapi32.dll")
	ntdll    = windows.MustLoadDLL("ntdll.dll")

	procGetSystemInfo        = kernel32.MustFindProc("GetSystemInfo")
	procGlobalMemoryStatusEx = kernel32.MustFindProc("GlobalMemoryStatusEx")
	procGetLogicalDrives     = kernel32.MustFindProc("GetLogicalDrives")
	procGetDriveTypeW        = kernel32.MustFindProc("GetDriveTypeW")
	procGetDiskFreeSpaceExW  = kernel32.MustFindProc("GetDiskFreeSpaceExW")
	procGetVolumeInformation = kernel32.MustFindProc("GetVolumeInformationW")
	procGetSystemDirectoryW  = kernel32.MustFindProc("GetSystemDirectoryW")
	procGetComputerNameW     = kernel32.MustFindProc("GetComputerNameW")

	procRegOpenKeyExW   = advapi32.MustFindProc("RegOpenKeyExW")
	procRegQueryValueEx = advapi32.MustFindProc("RegQueryValueExW")
	procRegCloseKey     = advapi32.MustFindProc("RegCloseKey")
	procGetUserNameW    = advapi32.MustFindProc("GetUserNameW")
	procRtlGetVersion   = ntdll.MustFindProc("RtlGetVersion")
)

func Detect(cpu, memory, disk, network, osFlag bool) map[string]interface{} {
	result := make(map[string]interface{}, 5)

	// 全量检测或按需
	all := !cpu && !memory && !disk && !network && !osFlag

	var mu sync.Mutex
	var wg sync.WaitGroup

	if all || cpu {
		wg.Go(func() {
			v := detectCPU()
			mu.Lock()
			result["cpu"] = v
			mu.Unlock()
		})
	}
	if all || memory {
		wg.Go(func() {
			v := detectMemory()
			mu.Lock()
			result["memory"] = v
			mu.Unlock()
		})
	}
	if all || disk {
		wg.Go(func() {
			v := detectDisk()
			mu.Lock()
			result["disk"] = v
			mu.Unlock()
		})
	}
	if all || network {
		wg.Go(func() {
			v := detectNetwork()
			mu.Lock()
			result["network"] = v
			mu.Unlock()
		})
	}
	if all || osFlag {
		wg.Go(func() {
			v := detectOS()
			mu.Lock()
			result["os"] = v
			mu.Unlock()
		})
	}

	wg.Wait()

	return result
}

func detectCPU() map[string]interface{} {
	m := make(map[string]interface{})

	// 逻辑核心数
	m["cores_logical"] = runtime.NumCPU()

	type systemInfo struct {
		wProcessorArchitecture      uint16
		wReserved                   uint16
		dwPageSize                  uint32
		lpMinimumApplicationAddress uintptr
		lpMaximumApplicationAddress uintptr
		dwActiveProcessorMask       uintptr
		dwNumberOfProcessors        uint32
		dwProcessorType             uint32
		dwAllocationGranularity     uint32
		wProcessorLevel             uint16
		wProcessorRevision          uint16
	}

	var si systemInfo
	procGetSystemInfo.Call(uintptr(unsafe.Pointer(&si)))
	m["cores_physical"] = si.dwNumberOfProcessors

	var hKey windows.Handle
	subkey, _ := windows.UTF16PtrFromString(`HARDWARE\DESCRIPTION\System\CentralProcessor\0`)
	procRegOpenKeyExW.Call(uintptr(windows.HKEY_LOCAL_MACHINE), uintptr(unsafe.Pointer(subkey)), 0, windows.KEY_READ, uintptr(unsafe.Pointer(&hKey)))

	if hKey != 0 {
		defer procRegCloseKey.Call(uintptr(hKey))

		// 查询 ProcessorNameString
		name, _ := windows.UTF16PtrFromString("ProcessorNameString")
		var buf [256]uint16
		var bufSize uint32 = uint32(len(buf) * 2)
		var regType uint32
		procRegQueryValueEx.Call(uintptr(hKey), uintptr(unsafe.Pointer(name)), 0, uintptr(unsafe.Pointer(&regType)), uintptr(unsafe.Pointer(&buf)), uintptr(unsafe.Pointer(&bufSize)))

		if bufSize > 0 {
			m["model"] = windows.UTF16ToString(buf[:bufSize/2])
		}
	}

	return m
}

func detectMemory() map[string]interface{} {
	m := make(map[string]interface{})

	type memStatus struct {
		dwLength                uint32
		dwMemoryLoad            uint32
		ullTotalPhys            uint64
		ullAvailPhys            uint64
		ullTotalPageFile        uint64
		ullAvailPageFile        uint64
		ullTotalVirtual         uint64
		ullAvailVirtual         uint64
		ullAvailExtendedVirtual uint64
	}

	var mem memStatus
	mem.dwLength = uint32(unsafe.Sizeof(mem))
	procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&mem)))

	m["total_bytes"] = mem.ullTotalPhys
	m["available_bytes"] = mem.ullAvailPhys
	m["used_bytes"] = mem.ullTotalPhys - mem.ullAvailPhys
	m["memory_load_percent"] = mem.dwMemoryLoad

	return m
}

func detectDisk() []map[string]interface{} {
	disks := make([]map[string]interface{}, 0, 8)
	ret, _, _ := procGetLogicalDrives.Call()
	logicalDrives := uint32(ret)

	// 遍历 A-Z
	for i := 0; i < 26; i++ {
		if (logicalDrives & (1 << i)) == 0 {
			continue
		}

		drive := fmt.Sprintf("%c:", 'A'+i)
		rootPath, _ := windows.UTF16PtrFromString(drive + "\\")

		// 检查驱动器类型
		ret, _, _ := procGetDriveTypeW.Call(uintptr(unsafe.Pointer(rootPath)))
		if ret != windows.DRIVE_FIXED && ret != windows.DRIVE_REMOVABLE {
			continue
		}

		// 获取磁盘空间
		var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64
		procGetDiskFreeSpaceExW.Call(uintptr(unsafe.Pointer(rootPath)), uintptr(unsafe.Pointer(&freeBytesAvailable)), uintptr(unsafe.Pointer(&totalNumberOfBytes)), uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)))

		if totalNumberOfBytes > 0 {
			d := make(map[string]interface{})
			d["drive"] = drive
			d["total_bytes"] = totalNumberOfBytes
			d["free_bytes"] = totalNumberOfFreeBytes
			d["used_bytes"] = totalNumberOfBytes - totalNumberOfFreeBytes

			// 获取文件系统
			var volumeName [256]uint16
			var serialNumber, maxComponentLength, fileSystemFlags uint32
			var fileSystemName [256]uint16
			procGetVolumeInformation.Call(uintptr(unsafe.Pointer(rootPath)), uintptr(unsafe.Pointer(&volumeName)), 256, uintptr(unsafe.Pointer(&serialNumber)), uintptr(unsafe.Pointer(&maxComponentLength)), uintptr(unsafe.Pointer(&fileSystemFlags)), uintptr(unsafe.Pointer(&fileSystemName)), 256)
			d["filesystem"] = windows.UTF16ToString(fileSystemName[:])

			disks = append(disks, d)
		}
	}

	return disks
}

func detectNetwork() []map[string]interface{} {
	nets := make([]map[string]interface{}, 0, 8)
	ifaces, err := net.Interfaces()
	if err != nil {
		return nets
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		n := map[string]interface{}{"name": iface.Name}
		if mac := strings.TrimSpace(iface.HardwareAddr.String()); mac != "" {
			n["mac"] = mac
		}

		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				ipNet, ok := addr.(*net.IPNet)
				if !ok {
					continue
				}
				ipv4 := ipNet.IP.To4()
				if ipv4 == nil {
					continue
				}
				ip := ipv4.String()
				if ip != "0.0.0.0" && !strings.HasPrefix(ip, "169.254.") {
					n["ipv4"] = ip
					break
				}
			}
		}

		if n["name"] != "" && n["ipv4"] != nil {
			nets = append(nets, n)
		}
	}

	return nets
}

func detectOS() map[string]interface{} {
	m := make(map[string]interface{})
	m["os"] = "windows"

	type osVersionInfoEx struct {
		dwOSVersionInfoSize uint32
		dwMajorVersion      uint32
		dwMinorVersion      uint32
		dwBuildNumber       uint32
		dwPlatformId        uint32
		szCSDVersion        [128]uint16
		wServicePackMajor   uint16
		wServicePackMinor   uint16
		wSuiteMask          uint16
		wProductType        byte
		wReserved           byte
	}

	if procRtlGetVersion != nil {
		var osVersion osVersionInfoEx
		osVersion.dwOSVersionInfoSize = uint32(unsafe.Sizeof(osVersion))
		procRtlGetVersion.Call(uintptr(unsafe.Pointer(&osVersion)))

		m["version"] = fmt.Sprintf("%d.%d.%d", osVersion.dwMajorVersion, osVersion.dwMinorVersion, osVersion.dwBuildNumber)
		m["major"] = osVersion.dwMajorVersion
		m["minor"] = osVersion.dwMinorVersion
		m["build"] = osVersion.dwBuildNumber
		m["service_pack_major"] = osVersion.wServicePackMajor

		// 版本名称
		switch osVersion.dwMajorVersion {
		case 10:
			if osVersion.dwBuildNumber >= 22000 {
				m["display_name"] = "Windows 11"
			} else {
				m["display_name"] = "Windows 10"
			}
		case 6:
			switch osVersion.dwMinorVersion {
			case 1:
				m["display_name"] = "Windows 7"
			case 2:
				m["display_name"] = "Windows 8"
			case 3:
				m["display_name"] = "Windows 8.1"
			}
		}
	}

	// 获取系统目录
	var sysDir [260]uint16
	procGetSystemDirectoryW.Call(uintptr(unsafe.Pointer(&sysDir)), 260)
	m["system_dir"] = windows.UTF16ToString(sysDir[:])

	// 计算机名和用户名
	var cn [256]uint16
	var cnLen uint32 = 256
	if procGetComputerNameW != nil {
		procGetComputerNameW.Call(uintptr(unsafe.Pointer(&cn)), uintptr(unsafe.Pointer(&cnLen)))
		m["computer_name"] = windows.UTF16ToString(cn[:])
	}

	if procGetUserNameW != nil {
		var un [256]uint16
		var unLen uint32 = 256
		procGetUserNameW.Call(uintptr(unsafe.Pointer(&un)), uintptr(unsafe.Pointer(&unLen)))
		m["user_name"] = windows.UTF16ToString(un[:])
	}

	return m
}