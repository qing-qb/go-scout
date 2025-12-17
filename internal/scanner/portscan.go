package scanner

import (
	"context"
	"fmt"
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
			results <- ScanPort(
				target,
				port,
				timeout)
		}

	}
}

// ScanPort 尝试连接目标IP的指定端口，并判断是否开放
// target: 目标IP地址，例如 "127.0.0.1"
// port: 目标端口号，例如 80
// timeout: 连接超时时间
// 返回 true 表示开放，false 表示关闭或超时

func ScanPort(target string, port int, timeout time.Duration) ScanResult {
	//拼接地址格式为 IP：Port
	address := fmt.Sprintf("%s:%d", target, port)
	//使用net.DialTimeout尝试建立Tcp连接
	//"tcp" 是协议类型，address 是目标地址，timeout 是超时时间

	//  第一步 基础检测 （TCP握手）
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		//连接失败
		return ScanResult{Port: port, State: "closed"}
	}
	conn.Close()

	//第二步 ： 获取指纹（day19 的核心功能
	banner := grabBanner(target, port, timeout)
	return ScanResult{Port: port, State: "open", Banner: banner}

}

// grabBanner 尝试获取端口指纹（Banner）
// 策略：先尝试读取（针对 SSH/FTP 等主动服务），如果超时，发送 HTTP 探测包再读取
//Banner 包含

func grabBanner(ip string, port int, timeout time.Duration) string {
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return ""
	}
	defer conn.Close()
	//设置超时
	readTimeout := 2 * time.Second
	conn.SetReadDeadline(time.Now().Add(readTimeout))

	//缓冲区
	buffer := make([]byte, 1024)

	//1,被动模式：先尝试直接读取（适用于SSH, FTP,SMTP)
	n, err := conn.Read(buffer)
	if err == nil && n > 0 {
		return cleanBanner(string(buffer[:n]))
	}

	//2,主动模式 如果没有读取到数据
	//发送HTTP HEAD请求
	httpRequest := "HEAD/HTTP/1.0\r\n\r\n" //\r\n\r\n代表只返回请求头
	conn.Write([]byte(httpRequest))

	//再次尝试
	conn.SetReadDeadline(time.Now().Add(readTimeout))
	n, err = conn.Read(buffer)
	if err == nil && n > 0 {
		return cleanBanner(string(buffer[:n]))
	}

	return "unknown"
}

// cleanBanner 清理 Banner 字符串中的换行和特殊字符

func cleanBanner(s string) string {
	// 这里可以加更多过滤逻辑，这里简单处理一下换行
	// 实际项目中可能需要正则表达式提取 Server: 字段
	if len(s) > 50 {
		s = s[:50] + "..."
	}
	return fmt.Sprintf("%q", s)
}
