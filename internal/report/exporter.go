package report

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

//这个模块负责将 Go 语言的结构体数据转化为 JSON 格式并写入文件

//ReportData

type ReportData struct {
	Target     string      `json:"target"`
	ScanTime   time.Time   `json:"scan_time"`
	Duration   string      `json:"duration"`
	TotalPorts int         `json:"total_ports"`
	Results    interface{} `json:"results"` //结果列表

}

// ExportJSON 负责将扫描结果导出为 JSON 文件
func ExportJSON(filename string, data ReportData) error {

	//1, 将 Go 结构体编码为 JSON 字节
	//json.MarshalIndent 可以实现带缩进的 JSON，更美观
	jsonData, err := json.MarshalIndent(data, "", "	")
	if err != nil {
		return fmt.Errorf("Error marshalling JSON: %w", err)
	}

	//2，将JSON字节写入文件
	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("Error writing file: %w", err)
	}

	return nil

}
