# mdnsmap

`mdnsmap` 是一个使用 Go 编写的 mDNS 局域网资产发现命令行工具。它可以扫描 IPv4 网段或单个 IPv4 地址，查询 mDNS 服务信息，并按端口范围过滤输出服务、主机名、IP、TTL 和 TXT banner 等资产信息。

## 功能

- 默认先发送 mDNS 组播查询，发现只响应组播的局域网设备。
- 继续对目标网段内每个 IPv4 地址发送 UDP/5353 单播查询，补充单播可见结果。
- 查询 `_services._dns-sd._udp.local.` 发现服务类型。
- 内置补充查询常见服务类型：
  - `_workstation._tcp.local.`
  - `_http._tcp.local.`
  - `_smb._tcp.local.`
  - `_qdiscover._tcp.local.`
  - `_device-info._tcp.local.`
  - `_afpovertcp._tcp.local.`
- 支持通过 `--service` 追加自定义 mDNS 服务类型。
- 解析 PTR、SRV、TXT、A、AAAA 记录。
- 按输入端口范围过滤 SRV 服务端口。
- 合并同一 IP 的组播和单播结果，避免重复输出。

## 使用方式

扫描一个局域网网段：

```powershell
go run ./cmd/mdnsmap --cidr 192.168.1.0/24 --ports 1-65535 --timeout 2s --workers 128 --retries 1
```

扫描单个 IPv4 地址：

```powershell
go run ./cmd/mdnsmap --cidr 192.168.1.20 --ports 80,443,5000-6000
```

只使用单播扫描：

```powershell
go run ./cmd/mdnsmap --cidr 192.168.1.0/24 --multicast=false
```

追加自定义服务类型：

```powershell
go run ./cmd/mdnsmap --cidr 192.168.1.0/24 --service _airplay._tcp.local. --service _raop._tcp.local.
```

## 参数

- `--cidr`：必填，IPv4 CIDR 或单个 IPv4 地址，例如 `192.168.1.0/24` 或 `192.168.1.10`。
- `--ports`：端口过滤范围，默认 `1-65535`，支持 `80,443,5000-6000`。
- `--timeout`：单次 mDNS 查询的响应等待时间，默认 `2s`。
- `--workers`：单播扫描 worker 数量，默认 `128`。
- `--retries`：每个 mDNS 查询的重试次数，默认 `1`。
- `--service`：追加自定义 mDNS 服务类型，可重复指定。
- `--multicast`：是否启用组播查询，默认 `true`；设置为 `false` 时只执行单播扫描。

## 输出示例

发现资产时会输出 `services`、`device-info` 和 `answers` 分组：

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

## 构建与验证

```powershell
go test ./...
go vet ./...
go build -o mdnsmap.exe ./cmd/mdnsmap
```

本地快速验证：

```powershell
go run ./cmd/mdnsmap --cidr 127.0.0.1/32 --ports 1-65535 --timeout 100ms --workers 1 --retries 0 --multicast=false
```

## 代码结构

- `cmd/mdnsmap`：CLI 程序入口。
- `internal/cli`：命令行参数定义、解析、校验和配置构建。
- `internal/model`：扫描配置、资产、服务、错误和结果等数据结构。
- `internal/scanner`：mDNS 查询、DNS 报文构造、组播/单播响应读取和响应解析。
- `internal/service`：扫描编排、目标扫描、组播扫描、服务记录合并和结果收集。
- `internal/output`：文本格式输出。
- `internal/util`：IP 网段展开和端口范围解析。

## 注意事项

mDNS 是本地链路协议，组播地址为 `224.0.0.251:5353`，通常只能发现当前二层网络内的设备。防火墙、网卡、VPN、虚拟网卡和路由器组播设置都可能影响扫描结果。

组播响应会按 `--cidr` 指定的目标范围过滤，避免把同一局域网中不属于本次扫描范围的设备混入结果。
