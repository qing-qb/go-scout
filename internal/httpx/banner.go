package httpx

//GetWebBanner 尝试获取目标端口的Web指纹(server头和Title
import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

//HTTP探测向开放端口，尝试发送请求
//指纹提取： 提取响应头中的Server字段和HTML中的<title>标签

// banner.go建立HTTP连接并提取信息

func GetWebBanner(ip string, port int) string {

	//1,构造URL（简易版：443默认HTTPS，其他默认HTTP）
	protocol := "http"
	if port == 8443 || port == 443 {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://%s:%d", protocol, ip, port)

	//2,创建自定义Client,必须设置超时，防止卡死 Worker(portScan.go)
	client := &http.Client{
		Timeout: time.Second * 2,
		// HTTP 请求超时通常比 TCP 连接长一点
		// 忽略 HTTPS 证书错误 (安全扫描通常需要忽略证书)
		// Transport: &http.Transport{ TLSClientConfig: &tls.Config{InsecureSkipVerify: true} },
		// 注意：为了保持 Day 11 代码简洁，暂不引入 crypto/tls，如果扫 HTTPS 报错先忽略，后续再加,
	}

	//3,发送请求
	resp, err := client.Get(url)
	if err != nil {
		return "lose" // // 不是 Web 服务或连接失败
	}
	defer resp.Body.Close()

	//4, 提取Server头（例如Nginx,Apache)
	serverHeader := resp.Header.Get("Server")

	//5,提取Title(读取前2KB内容查找<title>)
	var title string
	bodyStart := make([]byte, 2048)
	n, _ := resp.Body.Read(bodyStart)
	if n > 0 {
		content := string(bodyStart[:n])
		// 使用正则提取 <title>...</title>
		re := regexp.MustCompile(`(?i)<title>(.*?)</title>`)
		//(?i)忽略大小写
		matches := re.FindStringSubmatch(content)
		if len(matches) > 0 {
			title = matches[1]
		}
	}

	//6,格式化返回
	var bannerParts []string
	if serverHeader != "" {
		bannerParts = append(bannerParts, "Server:"+serverHeader)
	} else {
		fmt.Println("lsoe")
	}
	if title != "" {
		bannerParts = append(bannerParts, "Title:"+strings.TrimSpace(title))
		//TrimSpace(s)	删掉两端所有空白字符
		//Trim(s, "abc")	删掉两端指定字符集
		//TrimLeft/TrimRight	删除左侧或右侧字符
	}
	if len(bannerParts) > 0 {
		return strings.Join(bannerParts, " | ")
	}
	return ""
}
