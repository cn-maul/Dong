//go:build windows
// +build windows

package hardware

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const mediaDetectTimeout = 6 * time.Second

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

type driveMediaInfo struct {
	Drive string `json:"drive"`
	Media string `json:"media"`
}

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
	m["total_gb"] = bytesToGB(mem.ullTotalPhys)
	m["available_gb"] = bytesToGB(mem.ullAvailPhys)
	m["used_gb"] = bytesToGB(mem.ullTotalPhys - mem.ullAvailPhys)
	m["memory_load_percent"] = mem.dwMemoryLoad

	return m
}

func detectDisk() []map[string]interface{} {
	disks := make([]map[string]interface{}, 0, 8)
	driveMedia := detectDriveMediaTypes()
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
			d["total_gb"] = bytesToGB(totalNumberOfBytes)
			d["free_gb"] = bytesToGB(totalNumberOfFreeBytes)
			d["used_gb"] = bytesToGB(totalNumberOfBytes - totalNumberOfFreeBytes)
			d["disk_type"] = normalizeDiskType(driveMedia[drive])

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

func bytesToGB(v uint64) float64 {
	gb := float64(v) / (1024 * 1024 * 1024)
	return float64(int(gb*10+0.5)) / 10
}

func normalizeDiskType(raw string) string {
	r := strings.ToUpper(strings.TrimSpace(raw))
	switch {
	case strings.Contains(r, "SSD"), strings.Contains(r, "NVME"), r == "17":
		return "SSD"
	case strings.Contains(r, "HDD"):
		return "HDD"
	default:
		return "Unknown"
	}
}

func detectDriveMediaTypes() map[string]string {
	out := make(map[string]string)
	script := `$items = Get-Partition -ErrorAction SilentlyContinue |
  Where-Object { $_.DriveLetter } |
  ForEach-Object {
    $p = $_
    $d = Get-Disk -Number $p.DiskNumber -ErrorAction SilentlyContinue
    $media = if($d -and $d.MediaType){ $d.MediaType.ToString() } else { "Unknown" }
    if($media -eq "Unspecified" -or $media -eq "Unknown" -or [string]::IsNullOrWhiteSpace($media)){
      $pd = Get-PhysicalDisk -ErrorAction SilentlyContinue | Where-Object { $_.DeviceId -eq $p.DiskNumber } | Select-Object -First 1
      if($pd -and $pd.MediaType){
        $media = $pd.MediaType.ToString()
      }
      if(($media -eq "Unspecified" -or $media -eq "Unknown" -or [string]::IsNullOrWhiteSpace($media)) -and $pd){
        if($null -ne $pd.SpindleSpeed){
          if([int64]$pd.SpindleSpeed -gt 0){ $media = "HDD" } else { $media = "SSD" }
        }
      }
    }
    if($media -eq "Unspecified" -or $media -eq "Unknown" -or [string]::IsNullOrWhiteSpace($media)){
      $name = ""
      if($d -and $d.FriendlyName){ $name = $d.FriendlyName }
      $bus = ""
      if($d -and $d.BusType){ $bus = $d.BusType.ToString() }
      if($bus -match "NVMe" -or $bus -eq "17"){ $media = "SSD" }
      elseif($name -match "SSD|NVME|M\\.2"){ $media = "SSD" }
      elseif($name -match "HDD|SATA"){ $media = "HDD" }
      else { $media = "Unknown" }
    }
    if(($media -eq "Unspecified" -or $media -eq "Unknown" -or [string]::IsNullOrWhiteSpace($media)) -and $d -and $d.BusType){
      $media = $d.BusType.ToString()
    }
    [pscustomobject]@{
      drive = "$($p.DriveLetter):"
      media = $media
    }
  }
$items | ConvertTo-Json -Compress`

	ctx, cancel := context.WithTimeout(context.Background(), mediaDetectTimeout)
	defer cancel()
	b, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", script).CombinedOutput()
	if err != nil {
		return out
	}

	raw := strings.TrimSpace(string(b))
	if raw == "" {
		return out
	}

	var arr []driveMediaInfo
	if strings.HasPrefix(raw, "{") {
		var one driveMediaInfo
		if err := json.Unmarshal([]byte(raw), &one); err == nil && one.Drive != "" {
			out[strings.ToUpper(strings.TrimSpace(one.Drive))] = one.Media
		}
		return out
	}
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return out
	}
	for _, it := range arr {
		drive := strings.ToUpper(strings.TrimSpace(it.Drive))
		if drive == "" {
			continue
		}
		out[drive] = it.Media
	}
	return out
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
