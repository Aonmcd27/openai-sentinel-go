# Render API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 `openai-sentinel-go` 补一个可部署到 Render 的最小 HTTP 服务入口。

**Architecture:** 在现有 `sentinel` 库之上新增一个很薄的 `cmd/render-api` 入口，暴露 `GET /healthz` 和 `POST /build`。服务通过环境变量读取监听端口、Sentinel 配置和可选 Bearer 鉴权，不改动核心 token 生成逻辑。

**Tech Stack:** Go 标准库 `net/http`、现有 `openai-sentinel-go` 模块

---

### Task 1: 写失败测试

**Files:**
- Create: `D:\py\openai-sentinel-go\cmd\render-api\server_test.go`

- [ ] **Step 1: Write the failing test**
- [ ] **Step 2: Run test to verify it fails**
- [ ] **Step 3: Write minimal implementation**
- [ ] **Step 4: Run test to verify it passes**

### Task 2: 实现 handler 与启动入口

**Files:**
- Create: `D:\py\openai-sentinel-go\cmd\render-api\server.go`
- Create: `D:\py\openai-sentinel-go\cmd\render-api\main.go`

- [ ] **Step 1: 实现请求/响应结构和路由**
- [ ] **Step 2: 实现可选 Bearer 鉴权**
- [ ] **Step 3: 实现环境变量配置与启动**

### Task 3: Render 配置与说明

**Files:**
- Create: `D:\py\openai-sentinel-go\render.yaml`
- Modify: `D:\py\openai-sentinel-go\README.md`

- [ ] **Step 1: 添加 Render Blueprint**
- [ ] **Step 2: 补充部署说明**
- [ ] **Step 3: 运行 `go test ./...` 验证**
