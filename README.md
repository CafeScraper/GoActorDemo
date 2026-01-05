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

---

# ⭐ Script Input Configuration (input_schema.json)

The input_schema.json file serves as the "face" of your script. By modifying this file, you control which parameters users need to fill in before starting the script (e.g., URL, keywords, dates) and how these input fields are displayed (dropdowns, checkboxes, text boxes, etc.).

### 1. Overall Structure Analysis

A standard configuration file consists of the following three parts:：

1. **description**：Introduce the script's functionality and usage to users.
2. **b (Concurrency Key Field)**：Determines how the script splits tasks.
3. **properties (Parameter List)**：Specific functional settings.
### 💡  Code Example

```json
{
  "description": "With our Instagram Reel information scraper tool, after a successful scrape, you can extract the Reel author's username, Reel caption, hashtags used in the post, number of comments on the Reel, Reel publish date, likes count, views count, play count, popular comments, unique post identifier, URL of the Reel's display image or video thumbnail, product type, Reel duration, video URL, post audio link, number of posts on the profile, number of followers on the profile, profile URL, whether the account is a paid partner, and other relevant information. Currently, the tool can scrape via Instagram username, URL, and other methods, and the scrape results can be downloaded in various structured formats.",
  "b": "startUrl",
  "properties": [
    {
      "title": "URL",
      "name": "startUrl",
      "type": "array",
      "editor": "requestList",
      "description": "This parameter is used to specify the Instagram access URL to be fetched.",
      "default": [
        {
          "url": "<https://www.instagram.com/reel/C5Rdyj_q7YN/>"
        },
        {
          "url": "<https://www.instagram.com/reel/C85BZjeSHuO>"
        }
      ],
      "required": true
    }
  ]
}

```

### Output Example: Input

![Description](https://oss.cafehook.com/cafehook/20251226/31517262663254017.png)

### 2. Key Root Field Explanation

| Field Name | Required | Description |
| --- | --- | --- |
| **description** | No | **Scraper Overview**。Displayed at the top of the page, supports describing script functionality, notes, etc. |
| **b** | **Yes** | **Task Splitting Key**。 Must contain the `name` of an element in `properties` . The script splits tasks for concurrent processing based on this field (e.g., by number of URLs).
| **properties** | **Yes** | **Parameter Configuration Array**。Stores all input items, with each element representing an input box or selector on the page. |

---

### 3. Detailed Parameter Item Properties (Properties Inside)

Each specific input item can include the following configurations:

- **title**:Label displayed on the page (e.g., "Search Keywords").
- **name(Unique Identifier)**: Internal ID for the program, **must be unique**. Cannot contain Chinese characters.
- **type**:
    - string
    - integer
    - boolean
    - array
    - object
- **editor (Editor Type)**: Determines the form style of the input item on the webpage (see table below).
- **description**: Supplementary hint text below the input box, guiding the user on how to fill it out.
- **default**: Supplementary prompt text below the input box to guide users on filling.
- **required**: Set to true to prevent script startup if the field is empty.

---

### 4.  Editor Type Selection Guide

Optimize user experience by selecting different editors based on your needs:
#### 4.1. Basic Text & Numeric

| Type | Use Case |
| --- | --- |
| **input**  | Short text, keywords, or account name |
| **textarea** | Remarks or detailed text description |
| **number** | Limit the number of items to collect, page number, and wait time in seconds |

#### 4.2. Selectors

| Type | Example |
| --- | --- |
| **select** | Select gender, language, region. |
| **radio** | Single selection from 2-3 options (button layout). |
| **checkbox** | Select multiple tags of interest. |
| **switch** | Enable/disable options. |

#### 4.3. Time & Special Lists

| Type |  Use Case  |
| --- | --- |
| **datepicker** | Filter posts by specific publish date. |
| **requestList** |Batch input webpage links to scrape. |
| **requestListSource** |Customize additional parameters. |
| **stringList** | Batch input multiple keywords.|

---

### 5.  Common Component Code Examples

#### 5.1. Single-line Text Box (input)

```json
{
    "title": "📍 Location (use only one location per run)",
    "name": "location",
    "type": "string",
    "editor": "input",
    "default": "New York, USA"
}
```
![Description](https://oss.cafehook.com/cafehook/20251226/31517269654503424.png)


#### 5.2. Multi-line Text Box (textarea)

```json
{
    "title": "Filter reviews by keywords",
    "name": "keywords",
    "type": "string",
    "editor": "textarea"
}
```

![Description](https://oss.cafehook.com/cafehook/20251226/31517272128487424.png)

#### 5.3.  Numeric Adjustment Box (number)

```json
{
    "title": "Number of places to extract (per each search term or URL)",
    "name": "maxPlacesPerSearch",
    "type": "integer",
    "editor": "number",
    "default": 4
}
```
![Description](https://oss.cafehook.com/cafehook/20251226/31517292686082049.png)

#### 5.4. Dropdown Menu (select)

```json
{
    "title": "🌍 Language",
    "name": "language",
    "type": "string",
    "editor": "select",
    "options": [
        {
            "label": "English",
            "value": "en"
        },
        {
            "label": "Afrikaans",
            "value": "af"
        },
        {
            "label": "azərbaycan",
            "value": "az"
        }
    ],
    "default": "en"
}
```

![Description](https://oss.cafehook.com/cafehook/20251226/31517279399968769.png)

#### 5.5.  Radio Buttons (radio)

```json
{
    "title": "🏢 Category",
    "name": "radio",
    "type": "integer",
    "editor": "radioGroup",
    "options": [
        {
            "label": "hotel",
            "value": 1
        },
        {
            "label": "restaurant",
            "value": 2
        }
    ],
    "default": 1
}
```
![Description](https://oss.cafehook.com/cafehook/20251226/31517296009150465.png)

#### 5.6. Checkboxes (checkbox)

```json
{
    "title": "Data Sections to Scrape",
    "name": "data_sections",
    "type": "array",
    "editor": "checkboxGroup",
    "options": [
        {
            "label": "Reviews",
            "value": "reviews"
        },
        {
            "label": "Address",
            "value": "address"
        },
        {
            "label": "Phone Number",
            "value": "phone_number"
        }
    ],
    "default": ["reviews", "address"]
}
```
![Description](https://oss.cafehook.com/cafehook/20251226/31517297526177792.png)


#### 5.7. Date Picker (datepicker)

```json
{
    "title": "📅 Extract posts that are newer than",
    "name": "date",
    "type": "string",
    "editor": "datepicker",
    "format": "DD/MM/YYYY", 
    "valueFormat": "DD/MM/YYYY" 
}
```
![Description](https://oss.cafehook.com/cafehook/20251226/31517299223560193.png)

```json
// editor:"datepicker"
// dateType: 'absoluteOrRelative'
{
    "title": "📅 Extract posts that are newer than",
    "name": "date",
    "type": "string",
    "editor": "datepicker",
    "dateType": "absoluteOrRelative"
}
```

![Description](https://oss.cafehook.com/cafehook/20251226/31517300750417920.png)

![Description](https://oss.cafehook.com/cafehook/20251226/31517301740404737.png)

#### 5.8. Toggle Switch (switch)

```json
{
    "title": "⏩ Skip closed places",
    "name": "skipClosed",
    "type": "boolean",
    "editor": "checkbox"
}
```

![Description](https://oss.cafehook.com/cafehook/20251226/31517302786752512.png)

#### 5.9. URL List (requestList)

```json
{
    "name": "startURLs",
    "type": "array",
    "title": "Start URLs",
    "editor": "requestList",
    "default": [
        {
            "url": "https://www.google.com/search?sca_esv=593729410&q=Software+Engineer+jobs&uds=AMIYvT8-5jbJIP1-CbwNj1OVjAm_ezkS5e9c6xL1Cc4ifVo4bFIMuuQemtnb3giV7cKava9luZMDXVTS5p4powtoyb0ACtDGDu9unNkXZkFxC0i7ZSwrZd_aHgim6pFgOWgs0dte0pnb&sa=X&ictx=0&biw=1621&bih=648&dpr=2&ibp=htl;jobs&ved=2ahUKEwjt-4-Y6KyDAxUog4kEHSJ8DjQQudcGKAF6BAgRECo"
        },
        {
            "url": "https://www.google.com.hk/search?q=software+engineer+salary&newwindow=1&sca_esv=593729410&biw=1588&bih=1273&ei=vEtOadCxI-3AkPIP952z0Qc&oq=Software+Engineer&gs_lp=Egxnd3Mtd2l6LXNlcnAiEVNvZnR3YXJlIEVuZ2luZWVyKgIIAjIKEAAYgAQYQxiKBTIFEAAYgAQyBRAAGIAEMgUQABiABDIFEAAYgAQyBRAAGIAEMgUQABiABDIFEAAYgAQyBRAAGIAEMgUQABiABEj2yQFQ0TFYlbEBcAR4AZABBJgBkwOgAbUXqgEHMi0zLjQuMrgBA8gBAPgBAZgCBqACiAWoAgPCAgoQABiwAxjWBBhHwgIgEAAYgAQYtAIY1AMY5QIY5wYYtwMYigUY6gIYigPYAQGYAwTxBcFiu8bFvIEOiAYBkAYKugYECAEYB5IHCTQuMC4xLjAuMaAHvy6yBwcyLTEuMC4xuAf1BMIHAzItNsgHGIAIAA&sclient=gws-wiz-serp"
        }
    ],
    "required": true,
    "description": "The URLs of the website to scrape"
}
```
![Description](https://oss.cafehook.com/cafehook/20251226/31517304617041920.png)

#### 5.10. URL Request List Source (requestListSource)

```json
{
    "title": "startURLs",
    "name": "url",
    "type": "array",
    "editor": "requestListSource",
    "default": [
        {
            "url": "https://www.instagram.com/espn",
            "end_date": "",
            "start_date": "",
            "num_of_posts": "10",
            "posts_to_not_include": ""
        }
    ],
    "param_list": [
        {
            "param": "url",
            "title": "URL",
            "editor": "input",
            "type": "string",
            "required": true,
            "description": "This parameter is used to specify the Instagram access URL to be fetched."
        },
        {
            "param": "num_of_posts",
            "title": "Maximum Number of Reels",
            "type": "integer",
            "editor": "number",
            "description": "This parameter is used to specify the maximum number of Reels to fetch."
        },
        {
            "param": "start_date",
            "title": "Start Date",
            "type": "string",
            "editor": "datepicker",
            "format": "MM-DD-YYYY",
            "valueFormat": "MM-DD-YYYY",
            "description": "This parameter is used to specify the start time of the post, format: mm-dd-yyyy, and should be earlier than the \"end_date\"."
        },
        {
            "param": "end_date",
            "title": "End Date",
            "type": "string",
            "editor": "datepicker",
            "format": "MM-DD-YYYY",
            "valueFormat": "MM-DD-YYYY",
            "description": "This parameter is used to specify the end time of the post, format: mm-dd-yyyy, and should be later than the \"start_date\"."
        }
    ],
    "description": "The URLs of the website to scrape"
}
```
![Description](https://oss.cafehook.com/cafehook/20251226/31517306406043648.png)

#### 5.11. String List

```json
{
    "title": "🔍 Search term(s)",
    "name": "searchTerms",
    "type": "array",
    "editor": "stringList",
    "default": [
        {
            "string": "restaurant"
        },
        {
            "string": "school"
        }
    ]
}
```
![Description](https://oss.cafehook.com/cafehook/20251226/31517307891482624.png)

### 💡 Configuration Tips

1. **Clear Prompts:**：Ensure `description` is clear and accurate to help your script be discovered by more target users.
2. **Set Default Values**：Reasonable`default`values allow users to run the script with one click, significantly lowering the usage barrier.
3. **Mandatory Field Validation**：For parameters required for script execution (e.g., login cookies, main URL), always set `required`: true.