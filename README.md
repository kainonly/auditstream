# Weplanx Transfer

[![Github Actions](https://img.shields.io/github/workflow/status/weplanx/transfer/单元测试?style=flat-square)](https://github.com/weplanx/transfer/actions)
[![Coveralls github](https://img.shields.io/coveralls/github/weplanx/transfer.svg?style=flat-square)](https://coveralls.io/github/weplanx/transfer)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/weplanx/transfer?style=flat-square)](https://github.com/weplanx/transfer)
[![Go Report Card](https://goreportcard.com/badge/github.com/weplanx/transfer?style=flat-square)](https://goreportcard.com/report/github.com/weplanx/transfer)
[![Release](https://img.shields.io/github/v/release/weplanx/transfer.svg?style=flat-square)](https://github.com/weplanx/transfer)
[![GitHub license](https://img.shields.io/github/license/weplanx/transfer?style=flat-square)](https://raw.githubusercontent.com/weplanx/transfer/main/LICENSE)

日志传输器是以 Push 为主的服务，作用是对日志流队列进行统一管理，将高频写入的日志数据进行削峰缓冲，同时填补非高可用日志系统的可靠性，配合相同命名空间的日志采集器写入日志系统
> 版本 `*.*.*` 为 [elastic-transfer](https://github.com/weplanx/transfer/tree/elastic-transfer) 已归档的分支项目 ，请使用 `v*.*.*` 发布的版本（预发布用于构建测试）

技术文档：[语雀](https://www.yuque.com/kainonly/weplanx/xpaakq)

## License

[BSD-3-Clause License](https://github.com/weplanx/transfer/blob/main/LICENSE)