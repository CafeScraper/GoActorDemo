[English](https://docs.cafescraper.com/go-actor) | [中文](https://docs.cafescraper.com/cn/actor/actor/go-actor)


### Required Files (Located in Project Root)

```
├── main.go
├── main
├── input_schema.json
├── README.md
├── go.mod
├── go.sum
├── GoSdk
├────sdk.go
├────sdk.pd.go
├────sdk_grpc.pd.go

```

| File Name | Description |
| --- | --- |
| **main.go** | Script source code file |
| **main** | Script entry file (execution entry point), uniformly named `main` |
| **input_schema.json** | UI input form configuration file |
| **README.md** | Project documentation |
| **sdk.go** | SDK basic functionality, located in GoSdk directory |
| **sdk_pd.go** | Data processing enhancement module, located in GoSdk directory |
| **sdk_grpc.pd.go** | Network communication module, located in GoSdk directory |

### Go scripts need to be built into an executable file before uploading to the script marketplace
```shell
	set CGO_ENABLED=0 
	set GOOS=linux 
	set GOARCH=amd64 
	go build -o main ./main.go
```

# ⭐Core SDK Files

### 📁 File Description

The following three SDK files are essential. Place them in the **root directory** of your script:

| **File Name** | **Main Function** |
| --- | --- |
| `sdk.go` | Basic functionality module |
| `sdk_pd.go` | Data processing enhancement module |
| `sdk_grpc.pd.py` | Network communication module |

These three files form the script’s “toolbox,” providing all the core functions needed to interact with the backend system and run the scraper.
## 🔧 Core Function Usage

### 1. Environment Parameter Retrieval – Get Script Startup Configuration

When the script starts, you can pass configuration parameters from outside (e.g., target website URL, search keywords, etc.). Use the following method to retrieve them:
```go
// Get all input parameters as a JSON string
ctx := context.Background()
inputJSON, _ := cafesdk.Parameter.GetInputJSONString(ctx)

// Example: assuming website URL and keyword are provided
// Possible return: {"website": "example.com", "keyword": "Tech News"}

```

**Use Case:** If you need to scrape different websites for different tasks, you can pass different parameters without modifying the code.

---

### 2. Execution Logs – Record Script Process

During execution, you can record logs at different levels. These logs appear in the backend interface for monitoring and troubleshooting:

```go
ctx := context.Background()
// Debug (most detailed, for troubleshooting)
SDK.Log.Debug(ctx, "Connecting to target website...")

// Info (normal process logs)
SDK.Log.Info(ctx, "Successfully retrieved 10 news articles")

// Warning (attention needed but not an error)
SDK.Log.Warn(ctx, "Network connection is slow, may affect scraping speed")

// Error (used when execution fails)
SDK.Log.Error(ctx, "Cannot access target website, please check network connection")

```

**Log Level Explanation:**：

- **debug**：Most detailed, suitable for development
- **info**：Normal process logs, recommended for key steps
- **warn**：Warning message, indicates potential issues
- **error**：Error message, indicates an issue that requires attention

---

### 3. Result Submission – Send Scraped Data Back to Backend

After scraping, you need to return the data to the backend in two steps:

### Step 1: Set Table Header (Must be executed first)

Before pushing data, define the table structure like Excel column headers:

```go

// Set table header
headers := []*cafesdk.TableHeaderItem{
    {
        Label:  "Title",
        Key:    "title",
        Format: "text",
    },
    {
        Label:  "Content",
        Key:    "content",
        Format: "text",
    },
}
ctx := context.Background()
res, err := cafesdk.Result.SetTableHeader(ctx, headers)

```

**Field Explanation:**：

- **label**：Column title visible to users (recommended in English for global users)
- **key**：Unique identifier used in code (recommend lowercase with underscores)
- **format**：Data type, supports:
    - `text`
    - `integer`
    - `boolean`
    - `array`
    - `object`

### Step 2: Push Data Row by Row

After setting headers, push the scraped data:

```go
type result struct {
    Title   string `json:"title"`
    Content string `json:"content"`
}

resultData := []result{
    {Title: "Example Title 1", Content: "Example Content 1"},
    {Title: "Example Title 2", Content: "Example Content 2"},
}

ctx := context.Background()

for _, datum := range resultData {
    jsonBytes, _ := json.Marshal(datum)

    res, err := cafesdk.Result.PushData(ctx, string(jsonBytes))
    if err != nil {
        cafesdk.Log.Error(ctx, fmt.Sprintf("Push data failed: %v", err))
        return
    }
    fmt.Printf("PushData Response: %+v\n", res)
}


```

**Important Notes:**

1. Setting headers and pushing data can be done in any order
2. Keys in the data must match the table header keys exactly
3. Data must be pushed row by row, not all at once
4. Logging after each push is recommended for monitoring progress

---

### ⚠️ Common Issues and Precautions

1. File Location: Ensure the three SDK files are in the same folder as the main script
2. Import Method: Call functions directly via SDK or CafeSDK
3. Key Consistency: Pushed data keys must match header keys exactly
4. Error Handling: Check return results for each SDK call, especially when pushing data

These functions allow your script to integrate seamlessly with the backend system, providing flexible input parameters, transparent execution logs, and structured data submission.

---

# ⭐ Actor Entry File（main.go）

### 💡 Example Code

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    cafesdk "test/GoSdk"
    "time"
)

func run() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Printf("panic: %v\n", r)
        }
    }()

    time.Sleep(2 * time.Second)
    fmt.Println("golang gRPC SDK client started......")

    ctx := context.Background()

    // 1. Get input parameters
    inputJSON, err := cafesdk.Parameter.GetInputJSONString(ctx)
    if err != nil {
        cafesdk.Log.Error(ctx, fmt.Sprintf("Failed to get input parameters: %v", err))
        return
    }
    cafesdk.Log.Debug(ctx, fmt.Sprintf("Input parameters: %s", inputJSON))

    // 2. Get proxy configuration
    proxyDomain := "proxy-inner.cafescraper.com:6000"

    var proxyAuth string
    proxyAuth = os.Getenv("PROXY_AUTH")
    cafesdk.Log.Info(ctx, fmt.Sprintf("Proxy authentication: %s", proxyAuth))

    // 3. Build proxy URL
    var proxyURL string
    if proxyAuth != "" {
        proxyURL = fmt.Sprintf("socks5://%s@%s", proxyAuth, proxyDomain)
    }
    cafesdk.Log.Info(ctx, fmt.Sprintf("Proxy URL: %s", proxyURL))

    // Create custom HTTP client with proxy support
    httpClient := &http.Client{
        Timeout: time.Second * 30,
    }

    if proxyURL != "" {
        proxyParsed, err := url.Parse(proxyURL)
        if err != nil {
            cafesdk.Log.Error(ctx, fmt.Sprintf("Failed to parse proxy URL: %v", err))
            return
        }

        httpClient.Transport = &http.Transport{
            Proxy: http.ProxyURL(proxyParsed),
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true,
            },
        }

        cafesdk.Log.Info(ctx, "Proxy client configured")
    }

    // 4. Business logic (example)
    cafesdk.Log.Info(ctx, "Start processing business logic")
    targetURL := "https://ipinfo.io/ip"
    req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
    if err != nil {
        cafesdk.Log.Error(ctx, fmt.Sprintf("Failed to create request: %v", err))
        return
    }

    cafesdk.Log.Info(ctx, fmt.Sprintf("Requesting: %s", targetURL))
    resp, err := httpClient.Do(req)
    if err != nil {
        cafesdk.Log.Error(ctx, fmt.Sprintf("Request failed: %v", err))
        return
    }
    defer resp.Body.Close()

    cafesdk.Log.Info(ctx, fmt.Sprintf("Response status code: %d", resp.StatusCode))

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        cafesdk.Log.Error(ctx, fmt.Sprintf("Failed to read response: %v", err))
        return
    }

    ip := strings.TrimSpace(string(body))
    cafesdk.Log.Info(ctx, fmt.Sprintf("Current IP: %s", ip))
    cafesdk.Log.Info(ctx, "Business logic completed")

    // 5. Push result data
    type result struct {
        Title   string `json:"title"`
        Content string `json:"content"`
    }

    resultData := []result{
        {Title: "Example Title 1", Content: "Example Content 1"},
        {Title: "Example Title 2", Content: "Example Content 2"},
    }

    for _, datum := range resultData {
        jsonBytes, _ := json.Marshal(datum)

        res, err := cafesdk.Result.PushData(ctx, string(jsonBytes))
        if err != nil {
            cafesdk.Log.Error(ctx, fmt.Sprintf("Push data failed: %v", err))
            return
        }
        fmt.Printf("PushData Response: %+v\n", res)
    }

    // 6. Set table header
    headers := []*cafesdk.TableHeaderItem{
        {
            Label:  "Title",
            Key:    "title",
            Format: "text",
        },
        {
            Label:  "Content",
            Key:    "content",
            Format: "text",
        },
    }

    res, err := cafesdk.Result.SetTableHeader(ctx, headers)
    if err != nil {
        cafesdk.Log.Error(ctx, fmt.Sprintf("Set table header failed: %v", err))
        return
    }
    fmt.Printf("SetTableHeader Response: %+v\n", res)

    cafesdk.Log.Info(ctx, "Script execution completed")
}

func main() {
    run()
}

```

# Automated Data Scraper: Operation & Principles Guide

### 1. Script Overview

This is an automation tool template. It works like a “digital worker,” automatically opening specified web pages (e.g., social media pages), extracting required information, and organizing it into structured tables.

### 2. How It Works

The process can be simplified into four main stages:

### Step 1: Receive Instructions (Get Input Parameters)

Before starting, you provide instructions (e.g., which webpage to scrape, how many entries to retrieve).

### Step 2: Stealth Preparation (Proxy Network Configuration)

To access restricted or overseas websites smoothly, the script automatically configures a secure tunnel (proxy server).

### Step 3: Automated Job (Business Logic)

This is the core. The script visits target pages, reading titles, content, images, etc., according to the input.

### Step 4: Report Results (Data Push & Table Creation)

After scraping, the script converts raw data into a standardized format and submits it. It also sets up the table headers automatically (e.g., first column “URL,” second column “Content”).