//go:build linux
// +build linux

package advanced

import (
	"bufio"
	"io/ioutil"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
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
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := t.fn()
			mu.Lock()
			out[t.key] = v
			mu.Unlock()
		}()
	}
	wg.Wait()
	return out
}

func detectHardwareHealth(deepHW bool) interface{} {
	result := map[string]interface{}{}

	// 电池信息 (笔记本电脑)
	if data, err := ioutil.ReadFile("/sys/class/power_supply/BAT0/capacity"); err == nil {
		capacity := strings.TrimSpace(string(data))
		result["battery_capacity"] = capacity
	}
	if data, err := ioutil.ReadFile("/sys/class/power_supply/BAT0/status"); err == nil {
		status := strings.TrimSpace(string(data))
		result["battery_status"] = status
	}

	// CPU 温度
	tempFiles, _ := filepath.Glob("/sys/class/thermal/thermal_zone*/temp")
	for _, f := range tempFiles {
		if data, err := ioutil.ReadFile(f); err == nil {
			tempStr := strings.TrimSpace(string(data))
			if temp, err := strconv.Atoi(tempStr); err == nil {
				zone := filepath.Base(filepath.Dir(f))
				result["cpu_temp_"+zone] = temp / 1000 // 转换为摄氏度
			}
		}
	}

	// GPU 信息
	if out, err := exec.Command("lspci").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		gpus := []string{}
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "vga") ||
				strings.Contains(strings.ToLower(line), "3d") ||
				strings.Contains(strings.ToLower(line), "display") {
				gpus = append(gpus, strings.TrimSpace(line))
			}
		}
		if len(gpus) > 0 {
			result["gpu"] = gpus
		}
	}

	// 主板信息
	if data, err := ioutil.ReadFile("/sys/class/dmi/id/board_vendor"); err == nil {
		result["board_vendor"] = strings.TrimSpace(string(data))
	}
	if data, err := ioutil.ReadFile("/sys/class/dmi/id/board_name"); err == nil {
		result["board_name"] = strings.TrimSpace(string(data))
	}
	if data, err := ioutil.ReadFile("/sys/class/dmi/id/bios_vendor"); err == nil {
		result["bios_vendor"] = strings.TrimSpace(string(data))
	}
	if data, err := ioutil.ReadFile("/sys/class/dmi/id/bios_version"); err == nil {
		result["bios_version"] = strings.TrimSpace(string(data))
	}

	if deepHW {
		result["disk_smart"] = detectDiskSMART()
		result["disk_health"] = detectDiskHealth()
	}

	return result
}

func detectDiskSMART() interface{} {
	// 检查 smartctl 是否可用
	if _, err := exec.LookPath("smartctl"); err != nil {
		return map[string]interface{}{
			"ok":    false,
			"error": "smartctl not found (install smartmontools)",
		}
	}

	// 获取所有磁盘
	disks := []string{}
	if entries, err := ioutil.ReadDir("/sys/block"); err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasPrefix(name, "sd") || strings.HasPrefix(name, "nvme") || strings.HasPrefix(name, "vd") {
				disks = append(disks, "/dev/"+name)
			}
		}
	}

	if len(disks) == 0 {
		return map[string]interface{}{"ok": true, "disks": []string{}}
	}

	results := make(map[string]interface{})
	for _, disk := range disks {
		// smartctl 需要 root 权限
		out, err := exec.Command("smartctl", "-H", disk).CombinedOutput()
		if err != nil {
			results[disk] = map[string]interface{}{
				"ok":    false,
				"error": "requires root or smartctl error",
			}
			continue
		}
		results[disk] = map[string]interface{}{
			"ok":     true,
			"output": string(out),
		}
	}

	return map[string]interface{}{
		"ok":   true,
		"disks": results,
	}
}

func detectDiskHealth() interface{} {
	result := map[string]interface{}{}

	// 读取磁盘统计信息
	if entries, err := ioutil.ReadDir("/sys/block"); err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if !strings.HasPrefix(name, "sd") && !strings.HasPrefix(name, "nvme") && !strings.HasPrefix(name, "vd") {
				continue
			}

			diskInfo := map[string]interface{}{}

			// 读取磁盘大小
			if data, err := ioutil.ReadFile(filepath.Join("/sys/block", name, "size")); err == nil {
				if sectors, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
					diskInfo["size_bytes"] = sectors * 512
				}
			}

			// 设备模型
			if data, err := ioutil.ReadFile(filepath.Join("/sys/block", name, "device/model")); err == nil {
				diskInfo["model"] = strings.TrimSpace(string(data))
			}

			// 旋转设备 (HDD vs SSD)
			if data, err := ioutil.ReadFile(filepath.Join("/sys/block", name, "queue/rotational")); err == nil {
				rot := strings.TrimSpace(string(data))
				if rot == "0" {
					diskInfo["type"] = "SSD"
				} else {
					diskInfo["type"] = "HDD"
				}
			}

			result[name] = diskInfo
		}
	}

	return result
}

func detectSystemDiagnostics() interface{} {
	result := map[string]interface{}{}

	// 启动时间
	if data, err := ioutil.ReadFile("/proc/uptime"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) > 0 {
			if uptime, err := strconv.ParseFloat(fields[0], 64); err == nil {
				result["uptime_seconds"] = int(uptime)
			}
		}
	}

	// 启动时间戳
	if out, err := exec.Command("uptime", "-s").Output(); err == nil {
		result["boot_time"] = strings.TrimSpace(string(out))
	}

	// 当前用户
	if currentUser, err := user.Current(); err == nil {
		result["current_user"] = currentUser.Username
	}

	// 登录用户列表
	if out, err := exec.Command("who").Output(); err == nil {
		users := []string{}
		for _, line := range strings.Split(string(out), "\n") {
			if line = strings.TrimSpace(line); line != "" {
				users = append(users, line)
			}
		}
		result["logged_in_users"] = users
	}

	// systemd 服务状态 (如果存在)
	if _, err := exec.LookPath("systemctl"); err == nil {
		if out, err := exec.Command("systemctl", "list-units", "--state=failed", "--no-pager").Output(); err == nil {
			failed := []string{}
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.Contains(line, "failed") {
					failed = append(failed, strings.Fields(line)[0])
				}
			}
			result["failed_services"] = failed
		}
	}

	return result
}

func detectNetworkDiagnostics() interface{} {
	result := map[string]interface{}{}

	// 默认网关
	if out, err := exec.Command("ip", "route", "show", "default").Output(); err == nil {
		line := strings.TrimSpace(string(out))
		if line != "" {
			fields := strings.Fields(line)
			for i, f := range fields {
				if f == "via" && i+1 < len(fields) {
					result["default_gateway"] = fields[i+1]
				}
			}
		}
	}

	// DNS 服务器
	if data, err := ioutil.ReadFile("/etc/resolv.conf"); err == nil {
		dns := []string{}
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "nameserver ") {
				dns = append(dns, strings.TrimPrefix(line, "nameserver "))
			}
		}
		result["dns_servers"] = dns
	}

	// 测试网络连接
	gateway, _ := result["default_gateway"].(string)
	if gateway != "" {
		if out, err := exec.Command("ping", "-c", "1", "-W", "2", gateway).CombinedOutput(); err == nil {
			result["ping_gateway_ok"] = true
		} else {
			result["ping_gateway_ok"] = false
			_ = out
		}
	}

	// 测试外网连接
	if out, err := exec.Command("ping", "-c", "1", "-W", "2", "223.5.5.5").CombinedOutput(); err == nil {
		result["ping_external_ok"] = true
	} else {
		result["ping_external_ok"] = false
		_ = out
	}

	// DNS 解析测试
	if _, err := exec.Command("nslookup", "www.baidu.com").CombinedOutput(); err == nil {
		result["dns_ok"] = true
	} else {
		result["dns_ok"] = false
	}

	return result
}

func detectDriverDiagnostics() interface{} {
	result := map[string]interface{}{}

	// 加载的内核模块
	if out, err := exec.Command("lsmod").Output(); err == nil {
		modules := []map[string]string{}
		lines := strings.Split(string(out), "\n")
		for i, line := range lines {
			if i == 0 || line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				modules = append(modules, map[string]string{
					"name":  fields[0],
					"size":  fields[1],
					"count": fields[2],
				})
			}
		}
		result["loaded_modules"] = modules
	}

	// 检查 dmesg 中的错误 (需要权限)
	if out, err := exec.Command("dmesg", "-l", "err,crit,alert,emerg").Output(); err == nil {
		errors := []string{}
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if line = strings.TrimSpace(line); line != "" {
				errors = append(errors, line)
			}
		}
		result["kernel_errors"] = len(errors)
	}

	return result
}

func detectPerformanceDiagnostics() interface{} {
	result := map[string]interface{}{}

	// CPU 使用率
	if stat, err := ioutil.ReadFile("/proc/stat"); err == nil {
		lines := strings.Split(string(stat), "\n")
		if len(lines) > 0 {
			fields := strings.Fields(lines[0])
			if len(fields) >= 5 && fields[0] == "cpu" {
				user, _ := strconv.ParseInt(fields[1], 10, 64)
				nice, _ := strconv.ParseInt(fields[2], 10, 64)
				system, _ := strconv.ParseInt(fields[3], 10, 64)
				idle, _ := strconv.ParseInt(fields[4], 10, 64)
				total := user + nice + system + idle
				if total > 0 {
					result["cpu_usage_percent"] = int((total - idle) * 100 / total)
				}
			}
		}
	}

	// 负载
	if data, err := ioutil.ReadFile("/proc/loadavg"); err == nil {
		fields := strings.Fields(string(data))
		if len(fields) >= 3 {
			result["load_1min"], _ = strconv.ParseFloat(fields[0], 64)
			result["load_5min"], _ = strconv.ParseFloat(fields[1], 64)
			result["load_15min"], _ = strconv.ParseFloat(fields[2], 64)
		}
	}

	// 内存使用
	if data, err := ioutil.ReadFile("/proc/meminfo"); err == nil {
		var total, available uint64
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			key := strings.TrimSuffix(fields[0], ":")
			val, _ := strconv.ParseUint(fields[1], 10, 64)

			switch key {
			case "MemTotal":
				total = val
			case "MemAvailable":
				available = val
			}
		}
		if total > 0 {
			result["memory_usage_percent"] = int((total - available) * 100 / total)
		}
	}

	// 占用资源最多的进程
	if out, err := exec.Command("ps", "aux", "--sort=-%cpu").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		topCPU := []map[string]string{}
		for i, line := range lines {
			if i == 0 || i > 5 || line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) >= 11 {
				topCPU = append(topCPU, map[string]string{
					"user":    fields[0],
					"pid":     fields[1],
					"cpu":     fields[2],
					"mem":     fields[3],
					"command": fields[10],
				})
			}
		}
		result["top_cpu_processes"] = topCPU
	}

	if out, err := exec.Command("ps", "aux", "--sort=-%mem").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		topMem := []map[string]string{}
		for i, line := range lines {
			if i == 0 || i > 5 || line == "" {
				continue
			}
			fields := strings.Fields(line)
			if len(fields) >= 11 {
				topMem = append(topMem, map[string]string{
					"user":    fields[0],
					"pid":     fields[1],
					"cpu":     fields[2],
					"mem":     fields[3],
					"command": fields[10],
				})
			}
		}
		result["top_memory_processes"] = topMem
	}

	return result
}

func detectSoftwareInventory() interface{} {
	result := map[string]interface{}{}

	// 检测包管理器
	installed := []map[string]string{}

	// dpkg (Debian/Ubuntu/Deepin)
	if _, err := exec.LookPath("dpkg"); err == nil {
		if out, err := exec.Command("dpkg", "-l").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				fields := strings.Fields(line)
				if len(fields) >= 3 && fields[0] == "ii" {
					installed = append(installed, map[string]string{
						"name":    fields[1],
						"version": fields[2],
						"manager": "dpkg",
					})
				}
				if len(installed) >= 300 {
					break
				}
			}
		}
	}

	// rpm (Fedora/RHEL)
	if len(installed) == 0 {
		if _, err := exec.LookPath("rpm"); err == nil {
			if out, err := exec.Command("rpm", "-qa", "--queryformat", "%{NAME} %{VERSION}-%{RELEASE}\n").Output(); err == nil {
				lines := strings.Split(string(out), "\n")
				for _, line := range lines {
					if line = strings.TrimSpace(line); line != "" {
						parts := strings.SplitN(line, " ", 2)
						if len(parts) == 2 {
							installed = append(installed, map[string]string{
								"name":    parts[0],
								"version": parts[1],
								"manager": "rpm",
							})
						}
					}
					if len(installed) >= 300 {
						break
					}
				}
			}
		}
	}

	// pacman (Arch)
	if len(installed) == 0 {
		if _, err := exec.LookPath("pacman"); err == nil {
			if out, err := exec.Command("pacman", "-Q").Output(); err == nil {
				lines := strings.Split(string(out), "\n")
				for _, line := range lines {
					fields := strings.Fields(line)
					if len(fields) >= 2 {
						installed = append(installed, map[string]string{
							"name":    fields[0],
							"version": fields[1],
							"manager": "pacman",
						})
					}
					if len(installed) >= 300 {
						break
					}
				}
			}
		}
	}

	result["installed_packages"] = installed

	return result
}
