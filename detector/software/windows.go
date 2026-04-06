//go:build windows
// +build windows

package software

import (
	"context"
	"encoding/csv"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const commandTimeout = 6 * time.Second

type SoftwareReport struct {
	Go        ToolGo            `json:"go"`
	Node      ToolNode          `json:"node"`
	Python    ToolPython        `json:"python"`
	Java      ToolJava          `json:"java"`
	Git       ToolSimple        `json:"git"`
	Docker    ToolDocker        `json:"docker"`
	Kubectl   ToolKubectl       `json:"kubectl"`
	Dotnet    ToolDotnet        `json:"dotnet"`
	Env       map[string]string `json:"env"`
	Processes []ProcessInfo     `json:"processes"`
}

type ToolSimple struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version,omitempty"`
}

type ToolGo struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version,omitempty"`
	Goroot    string `json:"goroot,omitempty"`
	Gopath    string `json:"gopath,omitempty"`
}

type ToolNode struct {
	Installed  bool   `json:"installed"`
	Version    string `json:"version,omitempty"`
	NpmVersion string `json:"npm_version,omitempty"`
}

type ToolPython struct {
	Installed  bool   `json:"installed"`
	Version    string `json:"version,omitempty"`
	Executable string `json:"executable,omitempty"`
}

type ToolJava struct {
	Installed  bool   `json:"installed"`
	VersionRaw string `json:"version_raw,omitempty"`
	JavaHome   string `json:"java_home,omitempty"`
}

type ToolDocker struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version,omitempty"`
	Running   *bool  `json:"running,omitempty"`
}

type ToolKubectl struct {
	Installed bool   `json:"installed"`
	Raw       string `json:"raw,omitempty"`
}

type ToolDotnet struct {
	Installed bool     `json:"installed"`
	Version   string   `json:"version,omitempty"`
	Sdks      []string `json:"sdks,omitempty"`
}

type ProcessInfo struct {
	Name   string `json:"name"`
	Memory string `json:"memory"`
}

func Detect(deep bool) SoftwareReport {
	var out SoftwareReport
	var wg sync.WaitGroup

	wg.Go(func() { out.Go = detectGo() })
	wg.Go(func() { out.Node = detectNode() })
	wg.Go(func() { out.Python = detectPython() })
	wg.Go(func() { out.Java = detectJava() })
	wg.Go(func() { out.Git = detectGit() })
	wg.Go(func() { out.Docker = detectDocker(deep) })
	wg.Go(func() { out.Kubectl = detectKubectl() })
	wg.Go(func() { out.Dotnet = detectDotnet() })
	wg.Go(func() { out.Env = detectEnv() })
	wg.Go(func() { out.Processes = detectTopProcesses() })

	wg.Wait()
	return out
}

func runCommand(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	return exec.CommandContext(ctx, args[0], args[1:]...).Output()
}

func runCombinedCommand(args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	return exec.CommandContext(ctx, args[0], args[1:]...).CombinedOutput()
}

func detectGo() ToolGo {
	var t ToolGo
	// 一次命令拿到关键信息，减少进程启动开销
	out, err := runCommand("go", "env", "GOROOT", "GOPATH", "GOVERSION")
	if err != nil {
		return t
	}

	t.Installed = true
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) > 0 {
		t.Goroot = strings.TrimSpace(lines[0])
	}
	if len(lines) > 1 {
		t.Gopath = strings.TrimSpace(lines[1])
	}
	if len(lines) > 2 {
		t.Version = strings.TrimPrefix(strings.TrimSpace(lines[2]), "go")
	}

	return t
}

func detectNode() ToolNode {
	var t ToolNode
	out, err := runCommand("node", "--version")
	if err != nil {
		return t
	}

	t.Installed = true
	t.Version = strings.TrimSpace(string(out))

	out2, _ := runCommand("npm", "--version")
	t.NpmVersion = strings.TrimSpace(string(out2))

	return t
}

func detectPython() ToolPython {
	var t ToolPython
	// 尝试 python3
	out, err := runCombinedCommand("python3", "--version")
	if err == nil {
		t.Installed = true
		t.Version = strings.TrimSpace(string(out))
		t.Executable = "python3"
		return t
	}

	// 尝试 python
	out, err = runCombinedCommand("python", "--version")
	if err == nil {
		t.Installed = true
		t.Version = strings.TrimSpace(string(out))
		t.Executable = "python"
		return t
	}

	return t
}

func detectJava() ToolJava {
	var t ToolJava
	// java -version 主要输出到 stderr
	out, err := runCombinedCommand("java", "-version")
	if err != nil {
		return t
	}

	t.Installed = true
	lines := strings.Split(string(out), "\n")
	if len(lines) > 0 {
		t.VersionRaw = strings.TrimSpace(lines[0])
	}

	jh := os.Getenv("JAVA_HOME")
	if jh != "" {
		t.JavaHome = jh
	}

	return t
}

func detectGit() ToolSimple {
	var t ToolSimple
	out, err := runCommand("git", "--version")
	if err != nil {
		return t
	}

	t.Installed = true
	t.Version = strings.TrimSpace(string(out))
	return t
}

func detectDocker(deep bool) ToolDocker {
	var t ToolDocker
	out, err := runCommand("docker", "--version")
	if err != nil {
		return t
	}

	t.Installed = true
	t.Version = strings.TrimSpace(string(out))

	if deep {
		// 深度模式才执行相对更慢的命令
		_, err = runCommand("docker", "info")
		running := err == nil
		t.Running = &running
	}

	return t
}

func detectKubectl() ToolKubectl {
	var t ToolKubectl
	out, err := runCommand("kubectl", "version", "--client", "--output=json")
	if err != nil {
		return t
	}

	t.Installed = true
	// 简单处理，JSON 可能带颜色代码
	t.Raw = string(out)

	return t
}

func detectDotnet() ToolDotnet {
	var t ToolDotnet
	out, err := runCommand("dotnet", "--version")
	if err != nil {
		return t
	}

	t.Installed = true
	t.Version = strings.TrimSpace(string(out))

	// 获取 SDK 列表
	out2, _ := runCommand("dotnet", "--list-sdks")
	lines := strings.Split(string(out2), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			t.Sdks = append(t.Sdks, line)
		}
	}
	return t
}

func detectEnv() map[string]string {
	m := make(map[string]string, 8)

	// 缁翠慨鍦烘櫙鍙繚鐣欏叧閿幆澧冧俊鎭紝閬垮厤 PATH 鍜屼釜浜哄伐鍏疯矾寰勫櫔澹般€?
	importantVars := []string{
		"GOROOT", "GOPATH", "JAVA_HOME", "NODE_HOME", "PYTHON_HOME", "KUBECONFIG", "DOCKER_HOST",
	}
	for _, v := range importantVars {
		if val := os.Getenv(v); val != "" {
			m[v] = val
		}
	}

	if rawPath := os.Getenv("PATH"); rawPath != "" {
		parts := strings.Split(rawPath, string(os.PathListSeparator))
		core := make([]string, 0, len(parts))
		seen := make(map[string]struct{}, len(parts))
		for _, p := range parts {
			path := strings.TrimSpace(p)
			if path == "" {
				continue
			}
			lp := strings.ToLower(path)
			if strings.Contains(lp, `\windows\system32`) ||
				strings.Contains(lp, `\windowspowershell\`) ||
				strings.Contains(lp, `\openssh\`) ||
				strings.Contains(lp, `\program files\git\cmd`) ||
				strings.Contains(lp, `\program files\go\bin`) ||
				strings.Contains(lp, `\program files\nodejs`) ||
				strings.Contains(lp, `\program files\dotnet`) {
				if _, ok := seen[lp]; !ok {
					seen[lp] = struct{}{}
					core = append(core, path)
				}
			}
		}
		if len(core) > 0 {
			m["PATH_CORE"] = strings.Join(core, string(os.PathListSeparator))
		}
	}

	return m
}

func detectTopProcesses() []ProcessInfo {
	procs := make([]ProcessInfo, 0, 10)

	// 先切到 UTF-8 代码页，避免中文系统下 tasklist 输出乱码
	out, _ := runCommand("cmd", "/C", "chcp 65001>nul & tasklist /FO CSV /NH")
	r := csv.NewReader(strings.NewReader(string(out)))
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return procs
	}

	systemNoise := map[string]struct{}{
		"system idle process": {},
		"system":              {},
		"secure system":       {},
		"registry":            {},
		"smss.exe":            {},
		"csrss.exe":           {},
		"wininit.exe":         {},
		"winlogon.exe":        {},
		"services.exe":        {},
	}

	for _, row := range rows {
		if len(row) < 5 {
			continue
		}
		name := strings.TrimSpace(row[0])
		mem := strings.TrimSpace(row[4])
		if name == "" {
			continue
		}
		if _, skip := systemNoise[strings.ToLower(name)]; skip {
			continue
		}
		procs = append(procs, ProcessInfo{Name: name, Memory: mem})
		if len(procs) >= 10 {
			break
		}
	}

	return procs
}
