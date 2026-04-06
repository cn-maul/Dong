package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"dong/detector/advanced"
	"dong/detector/hardware"
	"dong/detector/software"
)

var (
	version   = "0.1.0"
	goVersion = runtime.Version()
)

type Report struct {
	Timestamp int64                    `json:"timestamp"`
	Hostname  string                   `json:"hostname"`
	GoVersion string                   `json:"go_version"`
	Runtime   string                   `json:"runtime"`
	Hardware  map[string]interface{}   `json:"hardware,omitempty"`
	Software  *software.SoftwareReport `json:"software,omitempty"`
	Advanced  map[string]interface{}   `json:"advanced,omitempty"`
}

func main() {
	// 命令行参数
	all := flag.Bool("all", false, "run all detection")
	hardwareFlag := flag.Bool("hardware", false, "detect hardware info")
	softwareFlag := flag.Bool("software", false, "detect software info")
	cpu := flag.Bool("cpu", false, "detect CPU only")
	memory := flag.Bool("memory", false, "detect memory only")
	disk := flag.Bool("disk", false, "detect disk only")
	network := flag.Bool("network", false, "detect network only")
	osFlag := flag.Bool("os", false, "detect OS info only")
	fast := flag.Bool("fast", false, "fast mode: skip expensive software checks")
	advancedFlag := flag.Bool("advanced", false, "run advanced diagnostics (auto enabled in full scan unless -fast)")
	deepHW := flag.Bool("deep-hw", false, "deep hardware health scan (SMART/physical disk health)")
	output := flag.String("o", "", "output to file (under reports/)")
	pretty := flag.Bool("pretty", false, "pretty print JSON")
	showVersion := flag.Bool("v", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Dong v%s (Go %s)\n", version, goVersion)
		os.Exit(0)
	}

	// 决定检测哪些项
	runAll := *all || (!*hardwareFlag && !*softwareFlag && !*cpu && !*memory && !*disk && !*network && !*osFlag)

	hostname, _ := os.Hostname()
	result := Report{
		Timestamp: time.Now().Unix(),
		Hostname:  hostname,
		GoVersion: goVersion,
		Runtime:   runtime.GOOS + "/" + runtime.GOARCH,
	}

	if runAll || *hardwareFlag || *cpu || *memory || *disk || *network || *osFlag {
		result.Hardware = hardware.Detect(*cpu, *memory, *disk, *network, *osFlag)
	}

	if runAll || *softwareFlag {
		s := software.Detect(!*fast)
		result.Software = &s
	}

	if *advancedFlag || (runAll && !*fast) {
		result.Advanced = advanced.Detect(*deepHW)
	}

	if *output != "" {
		// 输出到文件
		dir := filepath.Join(filepath.Dir(os.Args[0]), "reports")
		if exeDir := os.Getenv("DONG_REPORTS_DIR"); exeDir != "" {
			dir = exeDir
		}
		os.MkdirAll(dir, 0755)
		path := filepath.Join(dir, *output)
		if filepath.Ext(path) == "" {
			path += ".json"
		}
		f, err := os.Create(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create report file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		enc := json.NewEncoder(f)
		if *pretty {
			enc.SetIndent("", "  ")
		}
		if err := enc.Encode(result); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write report: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Report saved to: %s\n", path)
	} else {
		// stdout
		enc := json.NewEncoder(os.Stdout)
		if *pretty {
			enc.SetIndent("", "  ")
		}
		if err := enc.Encode(result); err != nil {
			fmt.Fprintf(os.Stderr, "failed to output report: %v\n", err)
			os.Exit(1)
		}
	}
}
