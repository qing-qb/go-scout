package scanner

import (
	"fmt"
	"net"
	"time"
)

// 扫描结果结构体
type ScanResult struct {
	Port  int
	State string //open /close
}

// // 端口扫描器的核心函数，现在用 Channel 来接收结果   worker pool;
func StartScan(target string, ports []int, concurrency int, timeout time.Duration) []ScanResult {
	// 1. 创建任务通道 (jobs) 和结果通道 (results)
	jobs := make(chan int, concurrency)
	results := make(chan ScanResult, len(ports))

	//2,启动协程池 (Worker Pool)
	for i := 1; i < concurrency; i++ {
		go worker(target, jobs, results, timeout)
	}

	//3,分发任务
	for _, port := range ports {
		jobs <- port
	}
	close(jobs)
	//4,收集结果
	var finalResults []ScanResult
	for i := 1; i <= len(ports); i++ {
		result := <-results //从结果管道接收结果
		finalResults = append(finalResults, result)
	}
	//可以在这里对其进行排序过滤
	return finalResults
}

// worker是协程池中的一个工作单元
func worker(target string, jobs <-chan int, results chan<- ScanResult, timeout time.Duration) {
	//从jobs管道接受任务
	for port := range jobs {
		//调用Day8核心扫描逻辑
		isOpen := CheckPort(target, port, timeout)
		State := "close"
		if isOpen {
			State = "open"
		}
		results <- ScanResult{port, State}
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
