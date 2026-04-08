# openai-sentinel-go

一个可单独分享的纯 Go 版 OpenAI Sentinel 实现，包含：

- requirements token 生成
- enforcement token / PoW 求解
- turnstile dx VM 求解
- 最小可用的会话与 persona 类型
- 针对当前实现的基础测试

另外，`Persona` 支持覆盖最新浏览器样本中的关键字段，适合做中英文环境精确对齐：

- `DateString`
- `RequirementsScriptURL`
- `NavigatorProbe`
- `DocumentProbe`
- `WindowProbe`
- `PerformanceNow`
- `RequirementsElapsed`

## 文件说明

- `service.go`：核心对外 API、token 生成、PoW
- `turnstile_vm.go`：dx VM 与浏览器环境重建
- `random.go`：最小随机辅助函数
- `*_test.go`：基础回归测试

## 最小用法

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"

    sentinel "openai-sentinel-go"
)

func main() {
    svc := sentinel.NewService(sentinel.Config{
        SentinelBaseURL:     "https://sentinel.openai.com",
        SentinelTimeout:     10 * time.Second,
        SentinelMaxAttempts: 2,
    })

    session := &sentinel.Session{
        Client:              &http.Client{Timeout: 10 * time.Second},
        DeviceID:            "device-123",
        UserAgent:           "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
        ScreenWidth:         1920,
        ScreenHeight:        1080,
        HeapLimit:           4294705152,
        HardwareConcurrency: 32,
        Language:            "zh-CN",
        LanguagesJoin:       "zh-CN,en-US",
        Persona: sentinel.Persona{
            Platform:   "Win32",
            Vendor:     "Google Inc.",
            SessionID:  "30ac1e73-e555-40f9-8ac4-76a1328458a3",
            TimeOrigin: 1775190798250,
        },
    }

    token, err := svc.Build(context.Background(), session, "username_password_create", "https://auth.openai.com/create-account/password", "")
    if err != nil {
        panic(err)
    }
    fmt.Printf("sentinel token: %+v\n", token)
}
```

## 验证

```bash
cd share/openai-sentinel-go
gofmt -w .
go test ./...
```

### 作者

LINUXDO：ius.

## Render deployment

This repository now includes a minimal Render-ready HTTP service at:

- `./cmd/render-api`

Endpoints:

- `GET /healthz`
- `POST /build`

### Build on Render

- Build Command:

```bash
go build -tags netgo -ldflags '-s -w' -o app ./cmd/render-api
```

- Start Command:

```bash
./app
```

### Environment variables

- `PORT` - Render injects this automatically
- `LISTEN_ADDR` - defaults to `0.0.0.0`
- `API_BEARER_TOKEN` - optional but strongly recommended
- `CLIENT_TIMEOUT_MS` - defaults to `10000`
- `SENTINEL_BASE_URL` - defaults to `https://sentinel.openai.com`
- `SENTINEL_TIMEOUT_MS` - defaults to `10000`
- `SENTINEL_MAX_ATTEMPTS` - defaults to `2`
- `SENTINEL_DIRECT_FALLBACK` - defaults to `false`
- `TURNSTILE_STATIC_TOKEN` - optional

### Example request

```bash
curl -X POST http://127.0.0.1:10000/build \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <API_BEARER_TOKEN>" \
  -d '{
    "flow": "username_password_create",
    "referer": "https://auth.openai.com/create-account/password",
    "turnstileToken": "",
    "session": {
      "deviceId": "device-123",
      "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
      "screenWidth": 1920,
      "screenHeight": 1080,
      "heapLimit": 4294705152,
      "hardwareConcurrency": 8,
      "language": "zh-CN",
      "languagesJoin": "zh-CN,en-US",
      "persona": {
        "platform": "Win32",
        "vendor": "Google Inc.",
        "sessionId": "30ac1e73-e555-40f9-8ac4-76a1328458a3",
        "timeOrigin": 1775190798250
      }
    }
  }'
```

## ClawCloud deployment

Recommended path:

1. Build and publish a Docker image to GHCR
2. Deploy that image in ClawCloud Run App Launchpad

### Local Docker build

```bash
docker build -t openai-sentinel-go:local .
docker run --rm -p 10000:10000 -e API_BEARER_TOKEN=your-token openai-sentinel-go:local
```

### GHCR publishing

This repository includes a GitHub Actions workflow:

- `.github/workflows/docker-ghcr.yml`

It publishes:

- `ghcr.io/aonmcd27/openai-sentinel-go:latest`
- `ghcr.io/aonmcd27/openai-sentinel-go:sha-...`

After pushing to `main`, go to the package page and make the package public if needed.

### ClawCloud Run settings

In App Launchpad, use:

- **Application Name**: `openai-sentinel-go`
- **Image Type**: `Public`
- **Image Name**: `ghcr.io/aonmcd27/openai-sentinel-go:latest`
- **Container Port**: `10000`
- **Public Access**: enabled

Environment variables:

- `API_BEARER_TOKEN` = your random secret
- `PORT` = `10000`
- `LISTEN_ADDR` = `0.0.0.0`
- `SENTINEL_BASE_URL` = `https://sentinel.openai.com`
- `SENTINEL_TIMEOUT_MS` = `10000`
- `SENTINEL_MAX_ATTEMPTS` = `2`
- `SENTINEL_DIRECT_FALLBACK` = `false`

### ClawCloud health check

After deployment, verify:

```bash
curl https://your-app-domain/healthz
```
