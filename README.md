# kit

这是一个用Go语言编写的工具包，提供了多个实用的功能模块。

This is a toolkit written in Go language, providing multiple practical functional modules.

## 功能模块 | Modules

### crosstime - 时间跨越检查 | Time Crossing Check
跨越时间检查工具，用于检测和处理小时、天、周、月的时间跨越事件。支持自定义周起始日。

A time crossing check tool for detecting and handling hour, day, week, and month crossing events. Supports custom week start day.

### excel - Excel文件处理 | Excel File Processing
提供Excel文件的读写功能，支持将Excel数据导出为JSON格式，可自定义导出字段和输出路径。同时支持将Excel数据结构导出为Go语言的结构体（struct）和TypeScript的接口（interface）定义。

Provides Excel file read and write functionality, supports exporting Excel data to JSON format with customizable export fields and output paths. Also supports exporting Excel data structure to Go struct and TypeScript interface definitions.

### listener - 网络监听器 | Network Listener
网络监听工具，支持TCP和Unix域套接字（Unix Domain Socket）两种监听方式。

Network listening tool that supports both TCP and Unix Domain Socket listening methods.

### lottery - 抽奖系统 | Lottery System
通用的权重抽奖系统，支持添加带权重的物品，并提供随机抽取功能。

A general-purpose weighted lottery system that supports adding weighted items and provides random drawing functionality.

### random - 随机数生成器 | Random Number Generator
提供基于种子的随机数生成器，可用于生成确定性的随机序列。

Provides a seed-based random number generator for generating deterministic random sequences.

### structparse - 结构体解析 | Struct Parser
用于解析Go结构体的工具，提供泛型支持。

A tool for parsing Go structs with generic support.

### timetunnel - 时间跨越检查 | Time Crossing Check
时间跨越检查工具，用于检测和处理小时、天、周、月的时间跨越事件。支持自定义周起始日。

A time crossing check tool for detecting and handling hour, day, week, and month crossing events. Supports custom week start day.

### IP工具 | IP Tools
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

## 许可证 | License
MIT License
