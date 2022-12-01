# errx

`go` `web`服务使用必备，能够自定错误码和客户端消息，打印错误堆栈信息，过滤堆栈信息，支持全局配置错误码和客户端消息。
堆栈信息基于[pkg/errors](https://github.com/pkg/errors)， 支持对[pkg/errors](https://github.com/pkg/errors)错误堆栈打印。

## 快速开始
```go
go get -u github.com/bobzh39/errx

// 全局默认信息配置
errx.Config = errx.ErroConfig{}
// 堆栈信息过滤
errx.Config.FilterStackTrace = func(*errx.StackTrace){
    var removes []errors.Frame
    for i := range *trace {
        pc := runtime.FuncForPC(uintptr((*trace)[i]))
        if pc == nil {
            continue
        } else {
            if !strings.Contains(pc.Name(), "bobzh39") {
                removes = append(removes, (*trace)[i])
            }
        }
    }
    
    for i := range removes {
        trace.Remove(removes[i])
    }
}

// 自定义错误创建
// 也可以参考`gerrx`目录下实现
errx.ErrorFactory = func(err error, msg string, opt ...errx.Option) error {
   return customErrorStruct
}

```

## 应用场景

### web服务使用

```go

// controller 日志打印
type UserContrller struct {}

func (u UserContrller) Create(ctx Context)  {
    user:=UserService{}
    data,err := user.Create(...)
    if err != nil {
        // 通过print会把堆栈信息打印处理
        log.Error(err)
    }
}

type UserService struct {}

func (user *UserService) Create(user User) (User,error){
	
    res,err := conn.Exec(...)
    if err != nil {
        return nil, errx.Wrap(err)	
    }
}



// 中间件错误处理
res ,err := handle(req,ctx)
if err != nil {
    log.Error(err)
    fmt.Println(err)
}
```
### 错误响应返回

```go
    // CustomResponse 由业务情况自己定义
    type CustomResponse struct {
	    Msg string
		Code string
    }


    if stackTrace,ok := errx.(HttpError);ok{
        //stackTrace.HttpCode() 用于返回http状态码
        //stackTrace.Code()  用于业务自定义code
        //stackTrace.Msg()  一般用于客户端提示消息
        //stackTrace.ErrorMsg()  系统内部的消息
        return CustomResponse{
            Msg: stackTrace.Msg()
            Code: stackTrace.Code()
        }
       
    }
	
```

## GRPC错误处理
使用[gerrx](./gerrx)返回GRPC中的错误，打印堆栈

