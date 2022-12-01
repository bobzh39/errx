# GRPC错误堆栈打印

## 快速开始
```go
go get github.com/bobzh39/errx/gerrx

// 定义默认code为 codes.Internal
errx.Config.DefaultGRPCCode = uint32(codes.Internal)
// 加载使用GRPC error工厂
gerrx.LoadGRPCError()
```
更多使用方式参考[errx](../README.md)