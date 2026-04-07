//go:build linux
// +build linux

package software

import (
	"context"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const commandTimeout = 6 * time.Second

type SoftwareReport struct {
	Go      ToolSimple  `json:"go"`
	Node    ToolSimple  `json:"node"`
	Python  ToolSimple  `json:"python"`
	Java    ToolJava    `json:"java"`
	Git     ToolSimple  `json:"git"`
	Docker  ToolSimple  `json:"docker"`
	Kubectl ToolKubectl `json:"kubectl"`
	Dotnet  ToolSimple  `json:"dotnet"`
}

type ToolSimple struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version,omitempty"`
}

type ToolJava struct {
	Installed  bool   `json:"installed"`
	VersionRaw string `json:"version_raw,omitempty"`
}

type ToolKubectl struct {
	Installed bool   `json:"installed"`
	Raw       string `json:"raw,omitempty"`
}

func Detect(deep bool) SoftwareReport {
	var out SoftwareReport
	var wg sync.WaitGroup

	wg.Add(8)
	go func() { defer wg.Done(); out.Go = detectGo() }()
	go func() { defer wg.Done(); out.Node = detectNode() }()
	go func() { defer wg.Done(); out.Python = detectPython() }()
	go func() { defer wg.Done(); out.Java = detectJava() }()
	go func() { defer wg.Done(); out.Git = detectGit() }()
	go func() { defer wg.Done(); out.Docker = detectDocker(deep) }()
	go func() { defer wg.Done(); out.Kubectl = detectKubectl() }()
	go func() { defer wg.Done(); out.Dotnet = detectDotnet() }()

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

func detectGo() ToolSimple {
	var t ToolSimple
	out, err := runCommand("go", "version")
	if err != nil {
		return t
	}
	t.Installed = true
	t.Version = strings.TrimSpace(string(out))
	return t
}

func detectNode() ToolSimple {
	var t ToolSimple
	out, err := runCommand("node", "--version")
	if err != nil {
		return t
	}
	t.Installed = true
	t.Version = strings.TrimSpace(string(out))
	return t
}

func detectPython() ToolSimple {
	var t ToolSimple
	// 先尝试 python3
	out, err := runCombinedCommand("python3", "--version")
	if err == nil {
		t.Installed = true
		t.Version = strings.TrimSpace(string(out))
		return t
	}

	// 再尝试 python
	out, err = runCombinedCommand("python", "--version")
	if err == nil {
		t.Installed = true
		t.Version = strings.TrimSpace(string(out))
	}
	return t
}

func detectJava() ToolJava {
	var t ToolJava
	out, err := runCombinedCommand("java", "-version")
	if err != nil {
		return t
	}
	t.Installed = true
	lines := strings.Split(string(out), "\n")
	if len(lines) > 0 {
		t.VersionRaw = strings.TrimSpace(lines[0])
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

func detectDocker(deep bool) ToolSimple {
	var t ToolSimple
	out, err := runCommand("docker", "--version")
	if err != nil {
		return t
	}
	t.Installed = true
	t.Version = strings.TrimSpace(string(out))
	return t
}

func detectKubectl() ToolKubectl {
	var t ToolKubectl
	out, err := runCommand("kubectl", "version", "--client", "--output=json")
	if err != nil {
		return t
	}
	t.Installed = true
	t.Raw = string(out)
	return t
}

func detectDotnet() ToolSimple {
	var t ToolSimple
	out, err := runCommand("dotnet", "--version")
	if err != nil {
		return t
	}
	t.Installed = true
	t.Version = strings.TrimSpace(string(out))
	return t
}
