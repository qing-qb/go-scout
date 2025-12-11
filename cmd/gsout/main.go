package main

import (
	"flag"
	"fmt"
	"go-scout/internal/scanner"
	"strconv"
	"strings"
	"time"
)

var targetIP *string //测试的IP地址
var portRange *string
var concurrency *int

func init() {
	targetIP = flag.String("t", "127.0.0.1", "target ip")
	portRange = flag.String("p", "1-1024", "target port range")
	concurrency = flag.Int("c", 1000, "concurrency number")
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
	//1,解析端口范围
	portsToScan, err := parsePorts(*portRange)
	if err != nil || len(portsToScan) == 0 {
		fmt.Println("端口范围解析错误，请使用 -h 查看用法。")
		return
	}

	//portsToScan := make([]int, 0)
	//for i := 1; i <= 1000; i++ {
	//	portsToScan = append(portsToScan, i)
	//}
	//核心 设置并发度
	//concurrency := 1000
	timeout := 500 * time.Millisecond //设置500毫秒超时

	startTime := time.Now()
	fmt.Printf("开始对 %s 扫描 %d 个端口，并发度：%d...\n", *targetIP, len(portsToScan), *concurrency)

	// 2,调用新的 StartScan 函数(使用命令行解析后的变量)
	results := scanner.StartScan(*targetIP, portsToScan, *concurrency, timeout)

	duration := time.Since(startTime)
	fmt.Printf("\n扫描完成，耗时: %s\n", duration)

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
}
