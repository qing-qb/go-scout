package main

import (
	"context"
	"flag"
	"fmt"
	"go-scout/internal/aiagent"
	"go-scout/internal/report"
	"go-scout/internal/scanner"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var targetIP *string //测试的IP地址
var portRange *string
var concurrency *int
var outputFile *string  //新增输出文件名参数 day12
var analyzeFile *string //新增分析文件参数
var aiKey *string       //新增AL Key参数

func init() {
	targetIP = flag.String("t", "", "target ip")
	portRange = flag.String("p", "1-1024", "target port range")
	concurrency = flag.Int("c", 1000, "concurrency number")
	outputFile = flag.String("o", "", "output file")
	analyzeFile = flag.String("a", "", "analyze file")
	aiKey = flag.String("key", "", "AI key")
}

//parsePorts 解析端口范围字符串

func parsePorts(portsStr string) ([]int, error) {
	ports := make([]int, 0)
	parts := strings.Split(portsStr, ",") //以逗号分隔

	for _, part := range parts {
		if strings.Contains(part, "-") {
			//处理范围扫描，如1-2024 ， 如果其中一个value包含—
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) == 2 {
				start, err1 := strconv.Atoi(rangeParts[0]) //字符串转整数
				end, err2 := strconv.Atoi(rangeParts[1])
				if err1 == nil && err2 == nil && start <= end {
					for i := start; i <= end; i++ {
						ports = append(ports, i)
					}
				}
			}
		} else {
			//处理单个窗口
			p, err := strconv.Atoi(part)
			if err == nil {
				ports = append(ports, p)
			}
		}
	}
	return ports, nil
}

func main() {
	// 解析命令行参数，必须在所有 flag 定义之后调用
	flag.Parse()

	tFlagProvided := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "t" {
			tFlagProvided = true
		}
	})
	//day15 ai纯分析模式
	if *analyzeFile != "" && !tFlagProvided {
		if *aiKey == "" {
			fmt.Println("[!] 错误: 必须提供 API Key 才能使用 AI 分析功能。")
			return
		}
		fmt.Printf("\n--- 🧠 AI 分析报告: %s ---\n", *analyzeFile)
		aiResult, err := aiagent.AnalyzeReport(*analyzeFile, *aiKey)
		if err != nil {
			fmt.Println("[!]AI 分析失败", err)
		} else {
			fmt.Println(aiResult)
		}
		return

	}
	//2, 扫描模式
	if !tFlagProvided {
		flag.Usage()
		return
	}
	//1,解析端口范围
	portsToScan, err := parsePorts(*portRange)
	if err != nil || len(portsToScan) == 0 {
		fmt.Println("端口范围解析错误，请使用 -h 查看用法。")
		return
	}
	mainCtx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel() // 确保在 main 函数结束时调用 cancel()  回收系统资源 避免多个 Goroutine 无休止地运行，确保程序的稳定性和高效性

	//2, day12 信号监听 在新的 Goroutine 中监听 Ctrl+C (SIGINT)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM) //监听中断信号
	go func() {
		select {
		case <-sigCh:
			fmt.Println("\n[!] 接收到中断信号 (Ctrl+C)。正在优雅停止扫描...")
			cancel()
		case <-mainCtx.Done():
			// 如果是超时导致 Context 结束，这里不会被触发，但这是 Go 惯用写法
		}
	}()

	timeout := 500 * time.Millisecond //设置500毫秒超时

	startTime := time.Now()
	fmt.Printf("开始对 %s 扫描 %d 个端口，并发度：%d...\n", *targetIP, len(portsToScan), *concurrency)

	// 2,调用新的 StartScan 函数(使用命令行解析后的变量)  扫描器核心
	results := scanner.StartScan(mainCtx, *targetIP, portsToScan, *concurrency, timeout)

	duration := time.Since(startTime)
	fmt.Printf("\n扫描完成，耗时: %s\n", duration)

	//day12 判断扫描是否真正完成

	if mainCtx.Err() == context.DeadlineExceeded {
		fmt.Printf("\n[!] 扫描超时（超过 5 分钟）！提前中止，已耗时: %s\n", duration)
	} else if mainCtx.Err() == context.Canceled {
		fmt.Printf("\n[!] 扫描被用户取消！已耗时: %s\n", duration)
	} else {
		fmt.Printf("\n扫描完成，总计耗时: %s\n", duration)
	}
	//3, 打印开放端口
	fmt.Println("\n--- 开放端口列表 ---")
	openCount := 0
	for _, result := range results {
		if result.State == "open" {
			if result.Banner != "" {
				fmt.Printf("[+] Port %d is %s banner is %s\n", result.Port, result.State, result.Banner)
			} else {
				fmt.Printf("[+] Port %d is %s\n", result.Port, result.State)
			}
			openCount++
		}
	}
	fmt.Printf("总计发现 %d 个开放端口。\n", openCount)

	//day 19 自动化升级
	var finalReportPath string
	if *outputFile != "" {
		finalReportPath = *outputFile
	} else if *aiKey != "" {
		timestamp := time.Now().Format("20060102_150405")
		finalReportPath = fmt.Sprintf("scan_%s_%s.json", *targetIP, timestamp)
		fmt.Printf("[💡] 自动生成报告文件: %s\n", finalReportPath)
	}

	if finalReportPath != "" {
		reportData := report.ReportData{
			Target:     *targetIP,
			ScanTime:   startTime,
			Duration:   duration.String(),
			TotalPorts: len(portsToScan),
			Results:    results,
		}

		// 先执行导出
		err := report.ExportJSON(finalReportPath, reportData)
		if err != nil {
			fmt.Printf("[!] 导出失败: %s\n", err)
		} else {
			fmt.Printf("[+] 报告成功导出：%s\n", finalReportPath)

			// 只有导出成功了，且有 Key，才紧接着执行 AI 分析
			if *aiKey != "" {
				fmt.Printf("\n--- 🧠 自动 AI 分析: %s ---\n", finalReportPath)
				aiResult, err := aiagent.AnalyzeReport(finalReportPath, *aiKey)
				if err != nil {
					fmt.Println("[!] AI 分析失败:", err)
				} else {
					fmt.Println(aiResult)
				}
			}
		}
	}
}
