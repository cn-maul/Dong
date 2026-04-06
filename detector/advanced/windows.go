//go:build windows
// +build windows

package advanced

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const psTimeout = 30 * time.Second

func Detect(deepHW bool) map[string]interface{} {
	type task struct {
		key string
		fn  func() interface{}
	}

	tasks := []task{
		{key: "hardware_health", fn: func() interface{} { return detectHardwareHealth(deepHW) }},
		{key: "system_diagnostics", fn: detectSystemDiagnostics},
		{key: "network_diagnostics", fn: detectNetworkDiagnostics},
		{key: "driver_diagnostics", fn: detectDriverDiagnostics},
		{key: "performance_diagnostics", fn: detectPerformanceDiagnostics},
		{key: "software_inventory", fn: detectSoftwareInventory},
	}

	out := make(map[string]interface{}, len(tasks))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, t := range tasks {
		t := t
		wg.Go(func() {
			v := t.fn()
			mu.Lock()
			out[t.key] = v
			mu.Unlock()
		})
	}
	wg.Wait()
	return out
}

func runPowerShell(script string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), psTimeout)
	defer cancel()
	utf8Prefix := `[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false); $OutputEncoding = [Console]::OutputEncoding; `
	return exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", utf8Prefix+script).CombinedOutput()
}

func runPSJSON(script string) interface{} {
	b, err := runPowerShell(script)
	if err != nil {
		return map[string]interface{}{
			"ok":    false,
			"error": strings.TrimSpace(string(b)),
		}
	}

	raw := strings.TrimSpace(string(b))
	if raw == "" {
		return map[string]interface{}{"ok": true}
	}

	var v interface{}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return map[string]interface{}{
			"ok":    false,
			"error": "json_parse_failed",
			"raw":   raw,
		}
	}
	return v
}

func detectHardwareHealth(deepHW bool) interface{} {
	// 电池/主板/BIOS/GPU/外设识别概况
	script := `
$battery = Get-CimInstance Win32_Battery -ErrorAction SilentlyContinue | Select-Object DesignCapacity,FullChargeCapacity,CycleCount,EstimatedChargeRemaining,BatteryStatus
$bios = Get-CimInstance Win32_BIOS -ErrorAction SilentlyContinue | Select-Object SMBIOSBIOSVersion,ReleaseDate,Manufacturer,SerialNumber
$board = Get-CimInstance Win32_BaseBoard -ErrorAction SilentlyContinue | Select-Object Product,Manufacturer,SerialNumber
$gpu = Get-CimInstance Win32_VideoController -ErrorAction SilentlyContinue | Select-Object Name,DriverVersion,AdapterRAM,Status
[pscustomobject]@{
  battery=$battery
  bios=$bios
  baseboard=$board
  gpu=$gpu
} | ConvertTo-Json -Depth 6 -Compress
`
	base := runPSJSON(script)
	if !deepHW {
		return base
	}
	return map[string]interface{}{
		"base":        base,
		"disk_smart":  detectDiskSMART(),
		"disk_health": detectDiskHealth(),
	}
}

func detectDiskSMART() interface{} {
	script := `
$smart = Get-CimInstance -Namespace root\wmi -ClassName MSStorageDriver_FailurePredictStatus -ErrorAction SilentlyContinue | Select-Object InstanceName,PredictFailure,Reason
$vendor = Get-CimInstance -Namespace root\wmi -ClassName MSStorageDriver_ATAPISmartData -ErrorAction SilentlyContinue | Select-Object InstanceName,VendorSpecific
[pscustomobject]@{
  predict_status=$smart
  vendor_data=$vendor
} | ConvertTo-Json -Depth 6 -Compress
`
	return runPSJSON(script)
}

func detectDiskHealth() interface{} {
	script := `
$pd = Get-PhysicalDisk -ErrorAction SilentlyContinue | Select-Object FriendlyName,HealthStatus,OperationalStatus,MediaType,Size
$vol = Get-Volume -ErrorAction SilentlyContinue | Select-Object DriveLetter,FileSystemLabel,HealthStatus,Size,SizeRemaining
[pscustomobject]@{
  physical_disks=$pd
  volumes=$vol
} | ConvertTo-Json -Depth 6 -Compress
`
	return runPSJSON(script)
}

func detectSystemDiagnostics() interface{} {
	// 启动时间、蓝屏、启动项、激活、UEFI/Legacy、更新状态
	script := `
$os = Get-CimInstance Win32_OperatingSystem -ErrorAction SilentlyContinue | Select-Object LastBootUpTime
$uptime = if($os.LastBootUpTime){ [int]((Get-Date) - $os.LastBootUpTime).TotalSeconds } else { $null }
$startup = Get-CimInstance Win32_StartupCommand -ErrorAction SilentlyContinue | Select-Object Name,Command,Location,User | Select-Object -First 20
$activation = Get-CimInstance SoftwareLicensingProduct -ErrorAction SilentlyContinue | Where-Object { $_.PartialProductKey -and $_.Name -match 'Windows' } | Select-Object -First 1 Name,LicenseStatus,Description
$fw = (Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Control' -Name PEFirmwareType -ErrorAction SilentlyContinue).PEFirmwareType
$mode = switch($fw){ 1 {'Legacy'} 2 {'UEFI'} default {'Unknown'} }
[pscustomobject]@{
  last_boot_time=$os.LastBootUpTime
  uptime_seconds=$uptime
  startup_items=$startup
  activation=$activation
  boot_mode=$mode
} | ConvertTo-Json -Depth 6 -Compress
`
	return runPSJSON(script)
}

func detectNetworkDiagnostics() interface{} {
	script := `
$cfg = Get-CimInstance Win32_NetworkAdapterConfiguration -ErrorAction SilentlyContinue | Where-Object { $_.IPEnabled -eq $true } | Select-Object -First 1 Description,DefaultIPGateway,DNSServerSearchOrder,IPAddress
$gw = $null
if($cfg -and $cfg.DefaultIPGateway){ $gw = $cfg.DefaultIPGateway[0] }
$pingGwOk = $false
if($gw){
  try {
    $gwRes = Test-Connection -ComputerName $gw -Count 1 -Quiet -ErrorAction Stop
    $pingGwOk = [bool]$gwRes
  } catch {}
}
$pingOutOk = $false
try {
  $outRes = Test-Connection -ComputerName 223.5.5.5 -Count 1 -Quiet -ErrorAction Stop
  $pingOutOk = [bool]$outRes
} catch {}
$dnsOk = $false
try { Resolve-DnsName www.microsoft.com -ErrorAction Stop | Out-Null; $dnsOk = $true } catch {}
$proxy = [string]::Join(' ', (netsh winhttp show proxy))
$proxyEnabled = -not ($proxy -match 'Direct access')
[pscustomobject]@{
  adapter=$cfg
  ping_gateway_ok=$pingGwOk
  ping_external_ok=$pingOutOk
  dns_ok=$dnsOk
  winhttp_proxy_enabled=$proxyEnabled
} | ConvertTo-Json -Depth 6 -Compress
`
	return runPSJSON(script)
}

func detectDriverDiagnostics() interface{} {
	script := `
$bad = Get-CimInstance Win32_PnPEntity -ErrorAction SilentlyContinue | Where-Object { $_.ConfigManagerErrorCode -ne 0 } | Select-Object Name,PNPClass,ConfigManagerErrorCode,Status
[pscustomobject]@{
  problematic_devices=$bad
} | ConvertTo-Json -Depth 6 -Compress
`
	return runPSJSON(script)
}

func detectPerformanceDiagnostics() interface{} {
	script := `
$cpuTop = Get-Process -ErrorAction SilentlyContinue | Sort-Object CPU -Descending | Select-Object -First 5 -ExpandProperty ProcessName
$memTop = Get-Process -ErrorAction SilentlyContinue | Sort-Object WS -Descending | Select-Object -First 5 -ExpandProperty ProcessName
$cpuNow = Get-CimInstance Win32_PerfFormattedData_PerfOS_Processor -ErrorAction SilentlyContinue | Where-Object { $_.Name -eq '_Total' } | Select-Object -First 1 PercentProcessorTime
$diskNow = Get-CimInstance Win32_PerfFormattedData_PerfDisk_PhysicalDisk -ErrorAction SilentlyContinue | Where-Object { $_.Name -eq '_Total' } | Select-Object -First 1 PercentDiskTime
$osNow = Get-CimInstance Win32_OperatingSystem -ErrorAction SilentlyContinue | Select-Object -First 1 FreePhysicalMemory
$gpuTotalBytes = [int64]((Get-CimInstance Win32_VideoController -ErrorAction SilentlyContinue | Measure-Object -Property AdapterRAM -Sum).Sum)
$gpuUsedBytes = $null
try {
  $c = Get-Counter '\\GPU Adapter Memory(*)\\Dedicated Usage' -ErrorAction Stop
  if($c -and $c.CounterSamples){
    $gpuUsedBytes = [int64](($c.CounterSamples | Measure-Object -Property CookedValue -Sum).Sum)
  }
} catch {}
$gpuTotalMB = if($gpuTotalBytes -gt 0){ [int]($gpuTotalBytes / 1MB) } else { $null }
$gpuUsedMB = if($null -ne $gpuUsedBytes){ [int]($gpuUsedBytes / 1MB) } else { $null }
[pscustomobject]@{
  top_cpu_process_names=$cpuTop
  top_memory_process_names=$memTop
  cpu_percent=([int]$cpuNow.PercentProcessorTime)
  disk_busy_percent=([int]$diskNow.PercentDiskTime)
  memory_available_mb=([int]([int64]$osNow.FreePhysicalMemory / 1024))
  gpu_memory_total_mb=$gpuTotalMB
  gpu_memory_used_mb=$gpuUsedMB
} | ConvertTo-Json -Depth 6 -Compress
`
	return runPSJSON(script)
}

func detectSoftwareInventory() interface{} {
	script := `
function Get-Soft($path){
  if(Test-Path $path){
    Get-ItemProperty $path -ErrorAction SilentlyContinue |
      Where-Object { $_.DisplayName } |
      Select-Object DisplayName,DisplayVersion,Publisher,InstallDate
  }
}
$soft = @()
$soft += Get-Soft 'HKLM:\Software\Microsoft\Windows\CurrentVersion\Uninstall\*'
$soft += Get-Soft 'HKLM:\Software\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*'
$browsers = @()
[pscustomobject]@{
  installed_software=($soft | Sort-Object DisplayName | Select-Object -First 300)
} | ConvertTo-Json -Depth 6 -Compress
`
	return runPSJSON(script)
}
