# kit

这是一个用Go语言编写的工具包，提供了多个实用的功能模块。

This is a toolkit written in Go language, providing multiple practical functional modules.

## 功能模块 | Modules

### event - 事件封装 | Event Wrapper
轻量级的事件封装库，提供基本的泛型事件结构，方便在系统各处传递事件类型与负载。

A lightweight event wrapper library providing basic generic event structures for passing event types and payloads throughout the system.

### excel - Excel文件处理 | Excel File Processing
提供Excel文件的读写功能，支持将Excel数据导出为JSON格式，可自定义导出字段和输出路径。同时支持将Excel数据结构导出为Go语言的结构体（struct）和TypeScript的接口（interface）定义。

Provides Excel file read and write functionality, supports exporting Excel data to JSON format with customizable export fields and output paths. Also supports exporting Excel data structure to Go struct and TypeScript interface definitions.

### listener - 网络监听器 | Network Listener
网络监听工具，支持TCP和Unix域套接字（Unix Domain Socket）两种监听方式。

Network listening tool that supports both TCP and Unix Domain Socket listening methods.

### weighted - 权重抽取系统 | Weighted Selector System
通用的权重抽取系统，支持添加带权重的物品，并提供灵活配置的随机抽取功能。

A general-purpose weighted selector system that supports adding weighted items and provides random drawing functionality with flexible options.

### onlinestore - 在线状态管理 | Online Status Management
基于 Redis Sorted Set 的在线状态存储系统，支持用户心跳记录、在线统计、用户列表查询、分页查询、批量筛选和自动清理离线用户等功能。使用 Option 模式提供灵活的配置方式。

A Redis Sorted Set based online status storage system that supports user heartbeat recording, online statistics, user list queries, pagination, batch filtering, and automatic cleanup of offline users. Uses Option pattern for flexible configuration.

### protobuf - Protobuf 编译工具 | Protobuf Compilation Tool
用于简化 Protocol Buffers 编译流程的工具包，支持自动发现 `.proto` 文件，并通过 `protoc` 及 `protoc-go-inject-tag` 生成 Go 代码并注入 Tag。

A toolkit to simplify the Protocol Buffers compilation process. It supports auto-discovering `.proto` files and generating Go code with injected tags via `protoc` and `protoc-go-inject-tag`.

### random - 随机数生成器 | Random Number Generator
提供基于种子的随机数生成器，可用于生成确定性的随机序列。

Provides a seed-based random number generator for generating deterministic random sequences.

### structparse - 结构体解析 | Struct Parser
用于解析Go结构体的工具，提供泛型支持。

A tool for parsing Go structs with generic support.

### timepass - 时间流逝检查 | Time Passing Check
时间流逝检查工具，通过比对上一次访问时间与当前时间，依次触发分钟、小时、天、周、月的时间流逝事件。支持自定义周起始日。

A time-lapse check tool that compares the last access time with the current time to sequentially trigger minute, hour, day, week, and month passing events. Supports custom week start day.

### viperparser - 配置解析器 | Configuration Parser
基于 Viper 的配置解析工具，支持从本地文件、`.env` 文件、HTTP/HTTPS、以及 etcd3 等多种远程数据源加载和解析配置文件，使用更加简便。

A Viper-based configuration parsing tool that supports loading and parsing config files from local files, `.env` files, HTTP/HTTPS, and remote data sources like etcd3 with ease.

### IP及系统工具 | IP and System Tools
提供获取本机非环回IP地址的功能。

Provides functionality to obtain local non-loopback IP addresses.

## 安装 | Installation

```bash
go get github.com/8liang/kit
```

## 依赖 | Dependencies
- github.com/golang-module/carbon/v2
- github.com/pkg/errors
- github.com/xuri/excelize/v2
- github.com/redis/go-redis/v9

## 许可证 | License
MIT License
