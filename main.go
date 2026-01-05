package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	cafesdk "test/GoSdk"
	"time"
)

func run() {
	ctx := context.Background()

	time.Sleep(2 * time.Second)
	cafesdk.Log.Info(ctx, "golang gRPC SDK client started......")

	// 1. 获取输入参数
	inputJSON, err := cafesdk.Parameter.GetInputJSONString(ctx)
	if err != nil {
		cafesdk.Log.Error(ctx, fmt.Sprintf("获取输入参数失败: %v", err))
		return
	}
	cafesdk.Log.Debug(ctx, fmt.Sprintf("输入参数: %s", inputJSON))

	// 2. 获取代理配置
	proxyDomain := "proxy-inner.cafescraper.com:6000"

	var proxyAuth string
	proxyAuth = os.Getenv("PROXY_AUTH")
	cafesdk.Log.Info(ctx, fmt.Sprintf("代理认证信息: %s", proxyAuth))

	// 3. 拼接代理 URL
	var proxyURL string
	if proxyAuth != "" {
		proxyURL = fmt.Sprintf("socks5://%s@%s", proxyAuth, proxyDomain)
	}
	cafesdk.Log.Info(ctx, fmt.Sprintf("代理地址: %s", proxyURL))

	// 4. 业务逻辑处理（示例）
	cafesdk.Log.Info(ctx, "开始处理业务逻辑")

	// 创建自定义 HTTP 客户端，支持代理
	httpClient := &http.Client{
		Timeout: time.Second * 30, // 设置超时时间
	}

	// 如果配置了代理，设置代理传输层
	if proxyURL != "" {
		// 解析代理URL
		proxyParsed, err := url.Parse(proxyURL)
		if err != nil {
			cafesdk.Log.Error(ctx, fmt.Sprintf("解析代理URL失败: %v", err))
			return
		}

		// 创建带代理的传输层
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyParsed),
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 仅测试使用，生产环境应配置正确的证书
			},
		}

		cafesdk.Log.Info(ctx, "已配置代理客户端")
	}

	// 发送请求到 ipinfo.io
	targetURL := "https://ipinfo.io/ip"
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		cafesdk.Log.Error(ctx, fmt.Sprintf("创建请求失败: %v", err))
		return
	}

	cafesdk.Log.Info(ctx, fmt.Sprintf("开始请求: %s", targetURL))

	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		cafesdk.Log.Error(ctx, fmt.Sprintf("请求失败: %v", err))
		return
	}
	defer resp.Body.Close()

	cafesdk.Log.Info(ctx, fmt.Sprintf("响应状态码: %d", resp.StatusCode))

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		cafesdk.Log.Error(ctx, fmt.Sprintf("读取响应失败: %v", err))
		return
	}

	// 打印返回的IP地址
	ip := strings.TrimSpace(string(body))
	cafesdk.Log.Info(ctx, fmt.Sprintf("当前IP地址: %s", ip))

	// 如果需要JSON格式输出，可以使用更结构化的方式
	cafesdk.Log.Info(ctx, "业务逻辑处理完成")

	type result struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	resultData := []result{
		{Title: "实列标题1", Content: "实列内容1"},
		{Title: "实列标题2", Content: "实列内容2"},
	}

	// 5. 推送结果数据

	for _, datum := range resultData {
		jsonBytes, _ := json.Marshal(datum)

		res, err := cafesdk.Result.PushData(ctx, string(jsonBytes))
		if err != nil {
			cafesdk.Log.Error(ctx, fmt.Sprintf("推送数据失败: %v", err))
			return
		}
		fmt.Printf("PushData Response: %+v\n", res)
	}

	// 6. 设置表格表头
	headers := []*cafesdk.TableHeaderItem{
		{
			Label:  "标题",
			Key:    "title",
			Format: "text",
		},
		{
			Label:  "内容",
			Key:    "content",
			Format: "text",
		},
	}

	res, err := cafesdk.Result.SetTableHeader(ctx, headers)
	if err != nil {
		cafesdk.Log.Error(ctx, fmt.Sprintf("设置表头失败: %v", err))
		return
	}
	fmt.Printf("SetTableHeader Response: %+v\n", res)

	cafesdk.Log.Info(ctx, "脚本执行完成")
}

func main() {
	run()
}
