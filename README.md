# mdnsmap

`mdnsmap` 是一个使用 Go 编写的 mDNS 网站测绘命令行工具。程序输入 IPv4 网段或单个 IPv4 地址，并按端口范围过滤 mDNS 服务结果，输出服务、主机、IP、TTL 和 TXT banner 等资产信息。

## 功能说明

- 对目标网段内每个 IPv4 地址发起 UDP/5353 单播 mDNS 查询。
- 查询 `_services._dns-sd._udp.local` 发现服务类型。
- 内置补充查询常见服务类型：
  - `_workstation._tcp.local`
  - `_http._tcp.local`
  - `_smb._tcp.local`
  - `_qdiscover._tcp.local`
  - `_device-info._tcp.local`
  - `_afpovertcp._tcp.local`
- 解析 PTR、SRV、TXT、A、AAAA 记录。
- 按输入端口范围过滤 SRV 服务端口。
- 默认输出接近题目示例的文本格式，保留 TXT 深度 banner。

## 使用方式

```powershell
go run ./cmd/mdnsmap --cidr 192.168.1.0/24 --ports 1-65535 --timeout 2s --workers 128 --retries 1
```

## 参数说明

- `--cidr`：必填，IPv4 CIDR 或单个 IPv4 地址，例如 `192.168.1.0/24` 或 `192.168.1.10`。
- `--ports`：端口过滤范围，默认 `1-65535`，支持 `80,443,5000-6000`。
- `--timeout`：单次 mDNS 查询超时时间，默认 `2s`。
- `--workers`：并发扫描 worker 数量，默认 `128`。
- `--retries`：每个查询的重试次数，默认 `1`。
- `--service`：追加自定义 mDNS 服务类型，可重复指定，例如 `_http._tcp.local`。

## 输出说明

有结果时输出 `services`、`device-info` 和 `answers` 分组：

```text
services:
5000/tcp qdiscover:
Name=slw-nas
IPv4=192.168.1.20
IPv6=fe80::265e:beff:fe69:a313
Hostname=slw-nas.local
TTL=10
accessType=https,accessPort=86,model=TS-X64,displayModel=TS-464C,fwVer=5.2.9,fwBuildNum=20260214
answers:
PTR:
_qdiscover._tcp.local
```

未发现资产时输出：

```text
未发现 mDNS 资产
```

## 开发验证

```powershell
go test ./...
go vet ./...
go run ./cmd/mdnsmap --cidr 127.0.0.1/32 --ports 1-65535 --timeout 100ms --workers 1 --retries 0
```

## 代码结构

- `cmd/mdnsmap`：CLI 程序入口。
- `internal/cli`：命令行参数定义、解析、校验和配置构建。
- `internal/model`：扫描配置、资产、服务、结果等数据结构。
- `internal/scanner`：mDNS 单播查询、DNS 报文构造、响应解析。
- `internal/service`：扫描编排、单目标扫描、服务记录合并和结果收集。
- `internal/output`：文本格式输出。
- `internal/util`：IP 网段展开和端口范围解析。

说明：mDNS 本身是本地链路协议。本程序按题目要求采用“仅单播扫描”，部分真实设备可能不会响应单播 mDNS 查询。
