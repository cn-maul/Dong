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
	"sort"
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

type driveDiskInfo struct {
	Drive           string `json:"drive"`
	DiskNumber      int    `json:"disk_number"`
	PartitionNumber int    `json:"partition_number"`
	DiskModel       string `json:"disk_model"`
	DiskSizeBytes   uint64 `json:"disk_size_bytes"`
	Media           string `json:"media"`
	BusType         string `json:"bus_type"`
}

func Detect(cpu, memory, disk, network, osFlag bool) map[string]interface{} {
	result := make(map[string]interface{}, 5)
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
			pn := detectPhysicalNetworkAdapters()
			mu.Lock()
			result["network"] = v
			result["network_physical_adapters"] = pn
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
		name, _ := windows.UTF16PtrFromString("ProcessorNameString")
		var buf [256]uint16
		var bufSize uint32 = uint32(len(buf) * 2)
		var regType uint32
		procRegQueryValueEx.Call(uintptr(hKey), uintptr(unsafe.Pointer(name)), 0, uintptr(unsafe.Pointer(&regType)), uintptr(unsafe.Pointer(&buf)), uintptr(unsafe.Pointer(&bufSize)))

		if bufSize > 0 {
			m["model"] = windows.UTF16ToString(buf[:bufSize/2])
		}

		// CPU current frequency (MHz), value name "~MHz"
		mhzName, _ := windows.UTF16PtrFromString("~MHz")
		var mhz uint32
		var mhzSize uint32 = 4
		var mhzType uint32
		procRegQueryValueEx.Call(uintptr(hKey), uintptr(unsafe.Pointer(mhzName)), 0, uintptr(unsafe.Pointer(&mhzType)), uintptr(unsafe.Pointer(&mhz)), uintptr(unsafe.Pointer(&mhzSize)))
		if mhz > 0 {
			m["frequency_mhz"] = mhz
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

	m["total_gb"] = bytesToGB(mem.ullTotalPhys)
	m["available_gb"] = bytesToGB(mem.ullAvailPhys)
	m["used_gb"] = bytesToGB(mem.ullTotalPhys - mem.ullAvailPhys)
	m["virtual_total_gb"] = bytesToGB(mem.ullTotalPageFile)
	m["virtual_available_gb"] = bytesToGB(mem.ullAvailPageFile)
	m["virtual_used_gb"] = bytesToGB(mem.ullTotalPageFile - mem.ullAvailPageFile)
	m["memory_load_percent"] = mem.dwMemoryLoad
	m["modules"] = detectMemoryModules()
	return m
}

func detectDisk() map[string]interface{} {
	partitions := make([]map[string]interface{}, 0, 8)
	driveInfo := detectDriveDiskInfo()

	ret, _, _ := procGetLogicalDrives.Call()
	logicalDrives := uint32(ret)

	for i := 0; i < 26; i++ {
		if (logicalDrives & (1 << i)) == 0 {
			continue
		}

		drive := strings.ToUpper(fmt.Sprintf("%c:", 'A'+i))
		rootPath, _ := windows.UTF16PtrFromString(drive + "\\")

		ret, _, _ := procGetDriveTypeW.Call(uintptr(unsafe.Pointer(rootPath)))
		if ret != windows.DRIVE_FIXED && ret != windows.DRIVE_REMOVABLE {
			continue
		}

		var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64
		procGetDiskFreeSpaceExW.Call(uintptr(unsafe.Pointer(rootPath)), uintptr(unsafe.Pointer(&freeBytesAvailable)), uintptr(unsafe.Pointer(&totalNumberOfBytes)), uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)))

		if totalNumberOfBytes == 0 {
			continue
		}

		d := make(map[string]interface{})
		d["drive"] = drive
		d["total_gb"] = bytesToGB(totalNumberOfBytes)
		d["free_gb"] = bytesToGB(totalNumberOfFreeBytes)
		d["used_gb"] = bytesToGB(totalNumberOfBytes - totalNumberOfFreeBytes)

		var volumeName [256]uint16
		var serialNumber, maxComponentLength, fileSystemFlags uint32
		var fileSystemName [256]uint16
		procGetVolumeInformation.Call(uintptr(unsafe.Pointer(rootPath)), uintptr(unsafe.Pointer(&volumeName)), 256, uintptr(unsafe.Pointer(&serialNumber)), uintptr(unsafe.Pointer(&maxComponentLength)), uintptr(unsafe.Pointer(&fileSystemFlags)), uintptr(unsafe.Pointer(&fileSystemName)), 256)
		d["filesystem"] = strings.TrimSpace(windows.UTF16ToString(fileSystemName[:]))

		if info, ok := driveInfo[drive]; ok {
			d["disk_number"] = info.DiskNumber
			d["partition_number"] = info.PartitionNumber
			d["disk_type"] = normalizeDiskType(info.Media)
		} else {
			d["disk_type"] = "Unknown"
		}

		partitions = append(partitions, d)
	}

	physicalMap := make(map[int]map[string]interface{})
	for _, p := range partitions {
		drive, _ := p["drive"].(string)
		info, ok := driveInfo[drive]
		if !ok {
			continue
		}
		if _, exists := physicalMap[info.DiskNumber]; !exists {
			physicalMap[info.DiskNumber] = map[string]interface{}{
				"disk_number":        info.DiskNumber,
				"model":              info.DiskModel,
				"disk_type":          normalizeDiskType(info.Media),
				"bus_type":           info.BusType,
				"total_gb":           bytesToGB(info.DiskSizeBytes),
				"logical_partitions": make([]map[string]interface{}, 0, 4),
			}
		}
		arr := physicalMap[info.DiskNumber]["logical_partitions"].([]map[string]interface{})
		arr = append(arr, p)
		physicalMap[info.DiskNumber]["logical_partitions"] = arr
	}

	numbers := make([]int, 0, len(physicalMap))
	for n := range physicalMap {
		numbers = append(numbers, n)
	}
	sort.Ints(numbers)
	physical := make([]map[string]interface{}, 0, len(numbers))
	for _, n := range numbers {
		physical = append(physical, physicalMap[n])
	}

	return map[string]interface{}{
		"physical_disks":     physical,
		"logical_partitions": partitions,
	}
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
	case strings.Contains(r, "HDD"), strings.Contains(r, "SATA"):
		return "HDD"
	default:
		return "Unknown"
	}
}

func detectDriveDiskInfo() map[string]driveDiskInfo {
	out := make(map[string]driveDiskInfo)
	script := `$items = Get-Partition -ErrorAction SilentlyContinue |
  Where-Object { $_.DriveLetter } |
  ForEach-Object {
    $p = $_
    $d = Get-Disk -Number $p.DiskNumber -ErrorAction SilentlyContinue
    $media = if($d -and $d.MediaType){ $d.MediaType.ToString() } else { "Unknown" }
    if($media -eq "Unspecified" -or $media -eq "Unknown" -or [string]::IsNullOrWhiteSpace($media)){
      $name = if($d -and $d.FriendlyName){ [string]$d.FriendlyName } else { "" }
      $bus = if($d -and $d.BusType){ [string]$d.BusType } else { "" }
      if($bus -match "NVMe" -or $bus -eq "17"){ $media = "SSD" }
      elseif($name -match "SSD|NVME|M\.2"){ $media = "SSD" }
      elseif($name -match "HDD|SATA"){ $media = "HDD" }
      else { $media = "Unknown" }
    }
    [pscustomobject]@{
      drive = "$($p.DriveLetter):"
      disk_number = [int]$p.DiskNumber
      partition_number = [int]$p.PartitionNumber
      disk_model = if($d){ [string]$d.FriendlyName } else { "" }
      disk_size_bytes = if($d){ [uint64]$d.Size } else { [uint64]0 }
      media = $media
      bus_type = if($d -and $d.BusType){ [string]$d.BusType } else { "" }
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

	if strings.HasPrefix(raw, "{") {
		var one driveDiskInfo
		if err := json.Unmarshal([]byte(raw), &one); err == nil && one.Drive != "" {
			one.Drive = strings.ToUpper(strings.TrimSpace(one.Drive))
			out[one.Drive] = one
		}
		return out
	}

	var arr []driveDiskInfo
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return out
	}
	for _, it := range arr {
		drive := strings.ToUpper(strings.TrimSpace(it.Drive))
		if drive == "" {
			continue
		}
		it.Drive = drive
		out[drive] = it
	}
	return out
}

func detectNetwork() []map[string]interface{} {
	nets := make([]map[string]interface{}, 0, 8)
	phyMap := make(map[string]string)
	for _, p := range detectPhysicalNetworkAdapters() {
		if n, ok := p["name"].(string); ok {
			if model, ok := p["model"].(string); ok && model != "" {
				phyMap[strings.ToLower(strings.TrimSpace(n))] = model
			}
		}
	}
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
		if model, ok := phyMap[strings.ToLower(strings.TrimSpace(iface.Name))]; ok {
			n["model"] = model
		}
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

func detectMemoryModules() []map[string]interface{} {
	out := make([]map[string]interface{}, 0, 4)
	script := `$m = Get-CimInstance Win32_PhysicalMemory -ErrorAction SilentlyContinue |
  Select-Object Manufacturer,PartNumber,ConfiguredClockSpeed,Capacity
$m | ConvertTo-Json -Compress`
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

	if strings.HasPrefix(raw, "{") {
		var one map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &one); err == nil {
			out = append(out, normalizeMemoryModule(one))
		}
		return out
	}
	var arr []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return out
	}
	for _, m := range arr {
		out = append(out, normalizeMemoryModule(m))
	}
	return out
}

func normalizeMemoryModule(m map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	manufacturer, _ := m["Manufacturer"].(string)
	partNumber, _ := m["PartNumber"].(string)
	out["manufacturer"] = strings.TrimSpace(manufacturer)
	out["part_number"] = strings.TrimSpace(partNumber)

	if capRaw, ok := m["Capacity"]; ok {
		switch v := capRaw.(type) {
		case float64:
			out["capacity_gb"] = bytesToGB(uint64(v))
		case string:
			var n uint64
			fmt.Sscanf(v, "%d", &n)
			out["capacity_gb"] = bytesToGB(n)
		}
	}
	out["configured_clock_mhz"] = m["ConfiguredClockSpeed"]
	return out
}

func detectPhysicalNetworkAdapters() []map[string]interface{} {
	out := make([]map[string]interface{}, 0, 4)
	script := `$n = Get-NetAdapter -Physical -ErrorAction SilentlyContinue |
  Select-Object Name,InterfaceDescription,Status,MacAddress,LinkSpeed
$n | ConvertTo-Json -Compress`
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

	if strings.HasPrefix(raw, "{") {
		var one map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &one); err == nil {
			out = append(out, map[string]interface{}{
				"name":       safePSString(one["Name"]),
				"model":      safePSString(one["InterfaceDescription"]),
				"status":     safePSString(one["Status"]),
				"mac":        safePSString(one["MacAddress"]),
				"link_speed": safePSString(one["LinkSpeed"]),
			})
		}
		return out
	}
	var arr []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return out
	}
	for _, n := range arr {
		out = append(out, map[string]interface{}{
			"name":       safePSString(n["Name"]),
			"model":      safePSString(n["InterfaceDescription"]),
			"status":     safePSString(n["Status"]),
			"mac":        safePSString(n["MacAddress"]),
			"link_speed": safePSString(n["LinkSpeed"]),
		})
	}
	return out
}

func safePSString(v interface{}) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func detectOS() map[string]interface{} {
	m := make(map[string]interface{})

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
		m["build"] = osVersion.dwBuildNumber

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

	var sysDir [260]uint16
	procGetSystemDirectoryW.Call(uintptr(unsafe.Pointer(&sysDir)), 260)
	m["system_dir"] = windows.UTF16ToString(sysDir[:])

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
