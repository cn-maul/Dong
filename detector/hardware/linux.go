//go:build linux
// +build linux

package hardware

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

func Detect(cpu, memory, disk, network, osFlag bool) map[string]interface{} {
	result := make(map[string]interface{}, 5)
	all := !cpu && !memory && !disk && !network && !osFlag

	var mu sync.Mutex
	var wg sync.WaitGroup

	if all || cpu {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := detectCPU()
			mu.Lock()
			result["cpu"] = v
			mu.Unlock()
		}()
	}
	if all || memory {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := detectMemory()
			mu.Lock()
			result["memory"] = v
			mu.Unlock()
		}()
	}
	if all || disk {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := detectDisk()
			mu.Lock()
			result["disk"] = v
			mu.Unlock()
		}()
	}
	if all || network {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := detectNetwork()
			mu.Lock()
			result["network"] = v
			mu.Unlock()
		}()
	}
	if all || osFlag {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v := detectOS()
			mu.Lock()
			result["os"] = v
			mu.Unlock()
		}()
	}

	wg.Wait()
	return result
}

func detectCPU() map[string]interface{} {
	m := make(map[string]interface{})
	m["cores_logical"] = runtime.NumCPU()

	// 从 /proc/cpuinfo 读取
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return m
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	physicalCorePairs := make(map[string]bool)
	physicalIDs := make(map[string]bool)
	modelName := ""
	freqMHz := 0.0
	cpuCoresPerSocket := 0
	currentPhysicalID := ""
	currentCoreID := ""

	flushCorePair := func() {
		if currentPhysicalID != "" && currentCoreID != "" {
			physicalCorePairs[currentPhysicalID+"#"+currentCoreID] = true
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			flushCorePair()
			currentPhysicalID = ""
			currentCoreID = ""
			continue
		}
		fields := strings.SplitN(line, ":", 2)
		if len(fields) != 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		val := strings.TrimSpace(fields[1])

		switch key {
		case "model name":
			if modelName == "" {
				modelName = val
			}
		case "cpu MHz":
			if freqMHz == 0 {
				if f, err := strconv.ParseFloat(val, 64); err == nil {
					freqMHz = f
				}
			}
		case "physical id":
			physicalIDs[val] = true
			currentPhysicalID = val
		case "core id":
			currentCoreID = val
		case "cpu cores":
			if cpuCoresPerSocket == 0 {
				if n, err := strconv.Atoi(val); err == nil && n > 0 {
					cpuCoresPerSocket = n
				}
			}
		}
	}
	flushCorePair()

	if modelName != "" {
		m["model"] = modelName
	}
	if freqMHz > 0 {
		m["frequency_mhz"] = int(freqMHz)
	}
	physicalCores := len(physicalCorePairs)

	// 尝试获取更详细的 CPU 信息
	if out, err := exec.Command("lscpu").Output(); err == nil {
		coresPerSocket := 0
		sockets := 0
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Architecture:") {
				m["architecture"] = strings.TrimSpace(strings.TrimPrefix(line, "Architecture:"))
			}
			if strings.HasPrefix(line, "Vendor ID:") {
				m["vendor"] = strings.TrimSpace(strings.TrimPrefix(line, "Vendor ID:"))
			}
			if strings.HasPrefix(line, "Core(s) per socket:") {
				v := strings.TrimSpace(strings.TrimPrefix(line, "Core(s) per socket:"))
				if n, err := strconv.Atoi(v); err == nil && n > 0 {
					coresPerSocket = n
				}
			}
			if strings.HasPrefix(line, "Socket(s):") {
				v := strings.TrimSpace(strings.TrimPrefix(line, "Socket(s):"))
				if n, err := strconv.Atoi(v); err == nil && n > 0 {
					sockets = n
				}
			}
		}
		if physicalCores == 0 && coresPerSocket > 0 && sockets > 0 {
			physicalCores = coresPerSocket * sockets
		}
	}
	if physicalCores == 0 && cpuCoresPerSocket > 0 && len(physicalIDs) > 0 {
		physicalCores = cpuCoresPerSocket * len(physicalIDs)
	}
	if physicalCores == 0 {
		physicalCores = runtime.NumCPU()
	}
	m["cores_physical"] = physicalCores

	return m
}

func detectMemory() map[string]interface{} {
	m := make(map[string]interface{})

	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return m
	}
	defer file.Close()

	var memTotal, memAvailable, swapTotal, swapFree uint64

	scanner := bufio.NewScanner(file)
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
			memTotal = val
		case "MemAvailable":
			memAvailable = val
		case "SwapTotal":
			swapTotal = val
		case "SwapFree":
			swapFree = val
		}
	}

	// 转换为 GB (原始单位是 kB)
	m["total_gb"] = float64(memTotal) / 1024 / 1024
	m["available_gb"] = float64(memAvailable) / 1024 / 1024
	m["used_gb"] = float64(memTotal-memAvailable) / 1024 / 1024
	m["virtual_total_gb"] = float64(swapTotal) / 1024 / 1024
	m["virtual_available_gb"] = float64(swapFree) / 1024 / 1024
	m["virtual_used_gb"] = float64(swapTotal-swapFree) / 1024 / 1024

	if memTotal > 0 {
		m["memory_load_percent"] = int((memTotal - memAvailable) * 100 / memTotal)
	}

	// 尝试获取内存条信息 (需要 dmidecode 权限)
	m["modules"] = detectMemoryModules()

	return m
}

func detectMemoryModules() []map[string]interface{} {
	out := make([]map[string]interface{}, 0)

	// 尝试使用 dmidecode (需要 root)
	if _, err := exec.LookPath("dmidecode"); err == nil {
		if output, err := exec.Command("dmidecode", "-t", "memory").Output(); err == nil {
			// 解析 dmidecode 输出比较复杂，这里简化处理
			_ = output
		}
	}

	// 从 /sys/class/dmi/id/ 读取一些信息
	dmiPath := "/sys/class/dmi/id"
	if info, err := ioutil.ReadFile(filepath.Join(dmiPath, "board_vendor")); err == nil {
		if len(out) == 0 {
			out = append(out, map[string]interface{}{
				"manufacturer": strings.TrimSpace(string(info)),
			})
		}
	}

	return out
}

func detectDisk() map[string]interface{} {
	result := map[string]interface{}{
		"physical_disks":     []map[string]interface{}{},
		"logical_partitions": []map[string]interface{}{},
	}

	// 使用 lsblk 获取磁盘信息
	usageByDevice := make(map[string]map[string]interface{})
	if out, err := exec.Command("lsblk", "-b", "-o", "NAME,SIZE,TYPE,MOUNTPOINT,FSTYPE,MODEL,ROTA", "-J").Output(); err == nil {
		if dfOut, err := exec.Command("df", "-B1", "-T", "-P").Output(); err == nil {
			logical := parseDF(string(dfOut))
			result["logical_partitions"] = logical
			for _, p := range logical {
				device, _ := p["device"].(string)
				if strings.TrimSpace(device) != "" {
					usageByDevice[device] = p
				}
			}
		}
		result["physical_disks"] = parseLSBLK(string(out), usageByDevice)
	}
	if len(result["logical_partitions"].([]map[string]interface{})) == 0 {
		if out, err := exec.Command("df", "-B1", "-T", "-P").Output(); err == nil {
			result["logical_partitions"] = parseDF(string(out))
		}
	}

	return result
}

func parseLSBLK(output string, usageByDevice map[string]map[string]interface{}) []map[string]interface{} {
	type blockDevice struct {
		Name     string        `json:"name"`
		Size     uint64        `json:"size"`
		Type     string        `json:"type"`
		Mount    interface{}   `json:"mountpoint"`
		FSType   interface{}   `json:"fstype"`
		Model    interface{}   `json:"model"`
		Rota     interface{}   `json:"rota"`
		Children []blockDevice `json:"children"`
	}
	type lsblkResp struct {
		Blockdevices []blockDevice `json:"blockdevices"`
	}

	var parsed lsblkResp
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		return []map[string]interface{}{}
	}

	disks := make([]map[string]interface{}, 0)

	for i, d := range parsed.Blockdevices {
		if d.Type != "disk" {
			continue
		}
		disk := map[string]interface{}{
			"disk_number":        i,
			"model":              asString(d.Model),
			"disk_type":          diskTypeFromRota(d.Rota),
			"total_gb":           bytesToGB(d.Size),
			"logical_partitions": []map[string]interface{}{},
		}
		parts := make([]map[string]interface{}, 0, len(d.Children))
		for _, c := range d.Children {
			if c.Type != "part" {
				continue
			}
			device := "/dev/" + c.Name
			p := map[string]interface{}{
				"device":           device,
				"drive":            device,
				"partition_number": partitionNumberFromName(c.Name),
				"mountpoint":       asString(c.Mount),
				"filesystem":       asString(c.FSType),
				"total_gb":         bytesToGB(c.Size),
			}
			if usage, ok := usageByDevice[device]; ok {
				if v, ok := usage["used_gb"]; ok {
					p["used_gb"] = v
				}
				if v, ok := usage["free_gb"]; ok {
					p["free_gb"] = v
				}
				if v, ok := usage["usage_percent"]; ok {
					p["usage_percent"] = v
				}
				if v, ok := usage["mountpoint"]; ok && asString(v) != "" {
					p["mountpoint"] = v
				}
				if v, ok := usage["filesystem"]; ok && asString(v) != "" {
					p["filesystem"] = v
				}
			}
			parts = append(parts, p)
		}
		disk["logical_partitions"] = parts
		disks = append(disks, disk)
	}

	return disks
}

func parseDF(output string) []map[string]interface{} {
	partitions := make([]map[string]interface{}, 0)
	lines := strings.Split(output, "\n")

	for i, line := range lines {
		if i == 0 { // skip header
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		device := fields[0]
		fsType := fields[1]
		total, _ := strconv.ParseUint(fields[2], 10, 64)
		used, _ := strconv.ParseUint(fields[3], 10, 64)
		available, _ := strconv.ParseUint(fields[4], 10, 64)
		usagePercent := fields[5]
		mountpoint := fields[6]

		// 只关注实际文件系统
		if !strings.HasPrefix(device, "/dev") {
			continue
		}

		partitions = append(partitions, map[string]interface{}{
			"device":           device,
			"drive":            device,
			"partition_number": partitionNumberFromName(strings.TrimPrefix(device, "/dev/")),
			"mountpoint":       mountpoint,
			"total_gb":         float64(total) / 1024 / 1024 / 1024,
			"used_gb":          float64(used) / 1024 / 1024 / 1024,
			"free_gb":          float64(available) / 1024 / 1024 / 1024,
			"filesystem":       fsType,
			"usage_percent":    usagePercent,
		})
	}

	return partitions
}

func detectNetwork() []map[string]interface{} {
	nets := make([]map[string]interface{}, 0)

	// 从 /sys/class/net 获取网络接口
	netPath := "/sys/class/net"
	entries, err := ioutil.ReadDir(netPath)
	if err != nil {
		return nets
	}

	for _, entry := range entries {
		name := entry.Name()
		if name == "lo" {
			continue
		}

		net := map[string]interface{}{"name": name}

		// 读取 MAC 地址
		if mac, err := ioutil.ReadFile(filepath.Join(netPath, name, "address")); err == nil {
			net["mac"] = strings.TrimSpace(string(mac))
		}

		// 读取状态
		if state, err := ioutil.ReadFile(filepath.Join(netPath, name, "operstate")); err == nil {
			net["status"] = strings.TrimSpace(string(state))
		}

		// 使用 ip 命令获取 IP 地址
		if out, err := exec.Command("ip", "-4", "addr", "show", name).Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.Contains(line, "inet ") {
					fields := strings.Fields(line)
					for i, f := range fields {
						if f == "inet" && i+1 < len(fields) {
							ip := strings.Split(fields[i+1], "/")[0]
							net["ipv4"] = ip
							break
						}
					}
				}
			}
		}

		if net["ipv4"] != nil {
			nets = append(nets, net)
		}
	}

	return nets
}

func detectOS() map[string]interface{} {
	m := make(map[string]interface{})

	// 读取 /etc/os-release
	if data, err := ioutil.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				m["display_name"] = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
			}
			if strings.HasPrefix(line, "VERSION=") {
				m["version"] = strings.Trim(strings.TrimPrefix(line, "VERSION="), `"`)
			}
			if strings.HasPrefix(line, "ID=") {
				m["id"] = strings.Trim(strings.TrimPrefix(line, "ID="), `"`)
			}
		}
	}

	// 读取内核版本
	if out, err := exec.Command("uname", "-r").Output(); err == nil {
		m["kernel"] = strings.TrimSpace(string(out))
	}

	// 读取架构
	if out, err := exec.Command("uname", "-m").Output(); err == nil {
		m["architecture"] = strings.TrimSpace(string(out))
	}

	// 主机名
	if hostname, err := os.Hostname(); err == nil {
		m["hostname"] = hostname
	}

	// 当前用户
	m["user_name"] = os.Getenv("USER")

	return m
}

func bytesToGB(v uint64) float64 {
	gb := float64(v) / (1024 * 1024 * 1024)
	return float64(int(gb*10+0.5)) / 10
}

func asString(v interface{}) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	default:
		return ""
	}
}

func diskTypeFromRota(v interface{}) string {
	switch x := v.(type) {
	case bool:
		if x {
			return "HDD"
		}
		return "SSD"
	case float64:
		if int(x) == 1 {
			return "HDD"
		}
		if int(x) == 0 {
			return "SSD"
		}
	case string:
		s := strings.TrimSpace(strings.ToLower(x))
		if s == "1" || s == "true" {
			return "HDD"
		}
		if s == "0" || s == "false" {
			return "SSD"
		}
	}
	return "Unknown"
}

func partitionNumberFromName(name string) int {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0
	}
	// nvme0n1p7 -> 7
	if i := strings.LastIndex(name, "p"); i >= 0 && i+1 < len(name) {
		if n, err := strconv.Atoi(name[i+1:]); err == nil {
			return n
		}
	}
	// sda3 -> 3
	j := len(name) - 1
	for j >= 0 && name[j] >= '0' && name[j] <= '9' {
		j--
	}
	if j < len(name)-1 {
		if n, err := strconv.Atoi(name[j+1:]); err == nil {
			return n
		}
	}
	return 0
}
