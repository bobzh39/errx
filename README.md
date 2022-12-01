# errx

`go``web`服务使用必备，能够自定错误码和客户端消息，打印错误堆栈信息，过滤堆栈信息，支持全局配置错误码和客户端消息。
觉得好用的，希望大佬们点点`star`，感谢。

## 特色

- 使用fmt.Println(err)直接打印错误堆栈信息
- 支持[pkg/errors](https://github.com/pkg/errors)类型的堆栈信息打印
- 错误码、客户端提示消息、http状态码单个定义和全局定义
- 支持堆栈信息栈帧过滤
- goroutine panic 堆栈信息打印

## 快速开始
```go
go get -u github.com/bobzh39/errx

if err !=nil {
    //  包装一个err，并保存堆栈信息
    return errx.Wrap(err)
}

if err !=nil {
    //  包装一个err和提示消息，并保存堆栈信息
    return errx.WrapMessage(err, "tips")
}

if code == "" {
    //  包装一个err和提示消息，没有堆栈信息
	return errx.WithTips("code不能为空")
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
        // 打印err堆栈信息
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
    // 打印err堆栈信息
    log.Error(err)
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

### 全局参数使用
```go
// 单个默认code覆盖
errx.Config.DefaultCode = "500"

// 覆盖全局默认信息配置
errx.Config = errx.ErroConfig{...}

// 堆栈信息过滤
errx.Config.FilterStackTrace = func(trace *errx.StackTrace){
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
```
#### 堆栈信息过滤
>  使用trace.Reverse()从栈的倒序过滤，这样更快哦，记得在调用一次Reverse()把堆栈顺序改回来，或者直接倒序循环过滤。

## 自定义错误工厂
自定义结构体创建错误，至少要实现StackTraceError接口，一般实现HttpError接口处理错误
```go

// customErrorStruct 至少要实现StackTraceError接口
type customErrorStruct struct {}
func (c customErrorStruct) Error() string {}
func (c customErrorStruct) Msg() string {}
func (c customErrorStruct) ErrorMsg() string {}
func (c customErrorStruct) AppendMsg(f func(string) string) {}
func (c customErrorStruct) Code() string {}
func (c customErrorStruct) SetCode(s string) {}
func (c customErrorStruct) Cause() error {}

// 自定义错误工厂创建
errx.ErrorFactory = func(err error, msg string, opt ...errx.Option) error {
   return &customErrorStruct{}
}
```

### Option介绍
用于每个错误创建时的自定义Code、httpCode, 中间件则可以拿到对应的code设置http的状态码
```go
if err != nil {
    return errx.WrapMessage(err, "无权限", errx.WithCode("401"),errx.WithHttpCode(401))
}
```

## Goroutine Recover堆栈信息
>每个自己开启的goroutine一定要defer recover，否则发生panic会导致整个程序挂掉。
```go
go func() {
    defer func() {
        if r := recover(); r != nil {
            t.Log(PanicTrace(r))
        }
    }()

    panic1()
}()
```


## GRPC错误处理
使用[gerrx](./gerrx)返回GRPC中的错误，打印堆栈

