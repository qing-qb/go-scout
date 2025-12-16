package scanner

import (
	"context"
	"fmt"
	"go-scout/internal/httpx"
	"net"
	"time"
)

// 扫描结果结构体

type ScanResult struct {
	Port   int
	State  string //open /close
	Banner string //day 11 新增：服务指纹
}

// // 端口扫描器的核心函数，现在用 Channel 来接收结果   worker pool;

func StartScan(ctx context.Context, target string, ports []int, concurrency int, timeout time.Duration) []ScanResult {

	//想象一下：你正在扫描一个网络，突然网络连接中断，或者用户发现扫错目标了，按下 Ctrl+C。你的 1000 个 Goroutine 还在后台疯跑，
	//直到操作系统强制终止，这会浪费资源，甚至造成数据丢失。
	//📅 Day 13 任务：健壮性与上下文控制 (Context)
	//在 Go 语言中，解决并发中的超时、取消和生命周期管理，唯一的答案是 context 包

	// 1. 创建任务通道 (jobs) 和结果通道 (results)
	jobs := make(chan int, concurrency)
	results := make(chan ScanResult, len(ports))

	//2,启动协程池 (Worker Pool)
	for i := 1; i < concurrency; i++ {
		go worker(ctx, target, jobs, results, timeout)
	}

	//3,分发任务
	go func() {
		for _, port := range ports {
			select {
			case jobs <- port:
			case <-ctx.Done():
				fmt.Println("\n[!] 任务通道关闭：扫描被中断或超时。")
				return // 退出分发 Goroutine
			}
		}
		close(jobs)
	}()
	//4,收集结果
	var finalResults []ScanResult
	for i := 1; i <= len(ports); i++ {
		select {
		case result := <-results: //从结果管道接收结果
			finalResults = append(finalResults, result)
		case <-ctx.Done():
			fmt.Println("\n[!] 结果收集器关闭：扫描被中断或超时。")
			return finalResults
			// 退出收集循环，返回目前已收集到的结果
		}
	}
	//可以在这里对其进行排序过滤
	return finalResults
}

// worker是协程池中的一个工作单元
func worker(ctx context.Context, target string, jobs <-chan int, results chan<- ScanResult, timeout time.Duration) {
	//从jobs管道接受任务
	for port := range jobs {
		//day12  核心：在处理每个任务前，检查 context 是否已经取消
		select {
		case <-ctx.Done():
			return
		default:
			state := "closed"
			banner := ""

			//调用Day8核心扫描逻辑CheckPort
			isOpen := CheckPort(target, port, timeout)

			if isOpen {
				state = "open"
				banner = httpx.GetWebBanner(target, port)
				// 🎯 新增逻辑：只有端口开放时，才去探测是不是 Web 服务
				// 简单的优化：通常只对常见 Web 端口或所有开放端口做这一步
			}
			results <- ScanResult{
				port,
				state,
				banner}
		}

	}
}

// ScanPort 尝试连接目标IP的指定端口，并判断是否开放
// target: 目标IP地址，例如 "127.0.0.1"
// port: 目标端口号，例如 80
// timeout: 连接超时时间
// 返回 true 表示开放，false 表示关闭或超时

func CheckPort(target string, port int, timeout time.Duration) bool {
	//拼接地址格式为 IP：Port
	addr := fmt.Sprintf("%s:%d", target, port)
	//使用net.DialTimeout尝试建立Tcp连接
	//"tcp" 是协议类型，address 是目标地址，timeout 是超时时间
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		//连接失败
		return false
	}
	defer conn.Close()
	return true

}
