package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
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
	webMode := flag.Bool("web", false, "start web UI server")
	cliMode := flag.Bool("cli", false, "force CLI mode (disable default web mode for web build)")
	webAddr := flag.String("web-addr", "127.0.0.1:18080", "web server listen address")
	showVersion := flag.Bool("v", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Dong v%s (Go %s)\n", version, goVersion)
		os.Exit(0)
	}

	runAll := *all || (!*hardwareFlag && !*softwareFlag && !*cpu && !*memory && !*disk && !*network && !*osFlag)
	result := buildReport(runAll, *hardwareFlag, *softwareFlag, *cpu, *memory, *disk, *network, *osFlag, *fast, *advancedFlag, *deepHW)

	if _, hasWebAssets := webUIFS(); hasWebAssets && !*cliMode {
		*webMode = true
	}

	if *webMode {
		startWebServer(*webAddr, result, func() Report {
			return buildReport(runAll, *hardwareFlag, *softwareFlag, *cpu, *memory, *disk, *network, *osFlag, *fast, *advancedFlag, *deepHW)
		})
		return
	}

	if *output != "" {
		dir := filepath.Join(filepath.Dir(os.Args[0]), "reports")
		if exeDir := os.Getenv("DONG_REPORTS_DIR"); exeDir != "" {
			dir = exeDir
		}
		_ = os.MkdirAll(dir, 0o755)
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

func buildReport(runAll, hardwareFlag, softwareFlag, cpu, memory, disk, network, osFlag, fast, advancedFlag, deepHW bool) Report {
	start := time.Now()
	hostname, _ := os.Hostname()
	result := Report{
		Timestamp: time.Now().Unix(),
		Hostname:  hostname,
		GoVersion: goVersion,
		Runtime:   runtime.GOOS + "/" + runtime.GOARCH,
	}

	fmt.Printf("[扫描] 开始：主机=%s 平台=%s 快速模式=%v 进阶诊断=%v 深度硬件=%v\n", hostname, runtime.GOOS+"/"+runtime.GOARCH, fast, advancedFlag || (runAll && !fast), deepHW)

	if runAll || hardwareFlag || cpu || memory || disk || network || osFlag {
		fmt.Println("[扫描] 正在检测硬件信息...")
		result.Hardware = hardware.Detect(cpu, memory, disk, network, osFlag)
		fmt.Println("[扫描] 硬件检测完成")
	}

	if runAll || softwareFlag {
		fmt.Printf("[扫描] 正在检测软件信息... 深度检测=%v\n", !fast)
		s := software.Detect(!fast)
		result.Software = &s
		fmt.Println("[扫描] 软件检测完成")
	}

	if advancedFlag || (runAll && !fast) {
		fmt.Printf("[扫描] 正在执行进阶诊断... 深度硬件=%v\n", deepHW)
		result.Advanced = advanced.Detect(deepHW)
		fmt.Println("[扫描] 进阶诊断完成")
	}

	fmt.Printf("[扫描] 全部完成，用时 %s\n", time.Since(start).Round(time.Millisecond))
	return result
}

func startWebServer(addr string, initial Report, refreshFn func() Report) {
	var mu sync.RWMutex
	report := initial

	mux := http.NewServeMux()

	mux.HandleFunc("/api/report", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		mu.RLock()
		defer mu.RUnlock()
		_ = json.NewEncoder(w).Encode(report)
	})

	mux.HandleFunc("/api/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		fmt.Println("[扫描] 收到前端刷新请求，开始重新扫描...")
		next := refreshFn()

		mu.Lock()
		report = next
		mu.Unlock()

		fmt.Println("[扫描] 前端刷新请求处理完成")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(next)
	})

	if root, ok := webUIFS(); ok {
		if sub, err := fs.Sub(root, "web"); err == nil {
			mux.Handle("/", http.FileServer(http.FS(sub)))
		}
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			fmt.Fprintln(w, "当前二进制未嵌入前端页面。请使用 build-with-frontend.bat 重新编译。")
		})
	}

	fmt.Printf("Web UI listening: http://%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "web server error: %v\n", err)
		os.Exit(1)
	}
}
