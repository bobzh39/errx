package errx

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type Option func(StackTraceError)

// WithCode set a custom biz code
func WithCode(code string) Option {
	return func(st StackTraceError) {
		st.SetCode(code)
	}
}

// WithHttpCode set a custom http code if the st is a HttpError
func WithHttpCode(code int) Option {
	return func(st StackTraceError) {
		if httpError, ok := st.(HttpError); ok {
			httpError.SetHttpCode(code)
		}
	}
}

// WithAppendMsg append tips content
func WithAppendMsg(fn func(string) string) Option {
	return func(st StackTraceError) {
		st.AppendMsg(fn)
	}
}

// WithGRPCCode set a custom grpc code if the st is a GRPCError
func WithGRPCCode(code uint32) Option {
	return func(st StackTraceError) {
		if grpcError, ok := st.(GRPCError); ok {
			grpcError.SetGRPCCode(code)
		}
	}
}

// WithTips create an error by ErrorFactory and no stack log.
// could be used WithCode WithHttpCode to define a biz code
func WithTips(tips string, opt ...Option) error {
	return ErrorFactory(nil, tips, opt...)
}

// Wrap create an error by ErrorFactory
// could be used WithCode WithHttpCode to define a biz code
// if err is StackTraceError will append tips and return
func Wrap(err error, opt ...Option) error {
	return ErrorFactory(err, Config.DefaultTips, opt...)

}

// WrapMessage create an error by ErrorFactory
// could be used WithCode WithHttpCode to define a biz code
// if err is StackTraceError will append the tips and return
func WrapMessage(err error, tips string, opt ...Option) error {
	return ErrorFactory(err, tips, opt...)
}

// New create a StackTraceError
// when err is nil, it will not stack trace
func New(err error, tips string, opt ...Option) error {
	err, ok := BuildStack(err, tips)
	if ok {
		return err
	}

	st := &DefaultStackTraceError{
		tips:     tips,
		code:     Config.DefaultCode,
		httpCode: Config.DefaultHttpCode,
		err:      err,
	}

	for i := range opt {
		opt[i](st)
	}

	return st
}

// Config error global ErrorConfig
var (
	Config = ErrorConfig{
		DefaultTips:     "服务器繁忙",
		DefaultCode:     "InternalError",
		DefaultHttpCode: http.StatusOK,
	}

	ErrorFactory = New
)

type (
	// ErrorConfig error global ErrorConfig
	// Note: These are the code for error conditions is not normal
	ErrorConfig struct {
		DefaultTips     string
		DefaultCode     string
		DefaultHttpCode int
		DefaultGRPCCode uint32
		// filter stack func
		FilterStackTrace func(*StackTrace)
	}

	// GRPCError grpc code interface
	GRPCError interface {
		StackTraceError
		// GRPCCode return a GRPCCode
		GRPCCode() uint32
		SetGRPCCode(uint32)
	}
	// HttpError http error 自定义code
	HttpError interface {
		StackTraceError
		HttpCode() int
		SetHttpCode(int)
	}
	// StackTraceError 有堆栈错误的error，包括自定义biz code
	StackTraceError interface {
		error
		// Msg 给到客户端的提示消息
		Msg() string
		// ErrorMsg 系统内部的错误信息
		ErrorMsg() string
		// AppendMsg 追加客户端提示消息
		AppendMsg(func(string) string)
		// Code 返回一个自定义的业务code
		Code() string
		// SetCode set a biz code
		SetCode(string)
		// Cause 返回当前实例中的err
		Cause() error
	}

	stackTracer interface {
		StackTrace() errors.StackTrace
	}

	causer interface {
		Cause() error
	}
)

type StackTrace errors.StackTrace

// Reverse slice of all
func (st *StackTrace) Reverse() {
	if len(*st) <= 1 {
		return
	}

	for i, j := 0, len(*st)-1; i < j; i, j = i+1, j-1 {
		(*st)[i], (*st)[j] = (*st)[j], (*st)[i]
	}
}

// Remove a frame
func (st *StackTrace) Remove(frame errors.Frame) {
	flag := 0
	for i := range *st {
		if (*st)[i] != frame {
			(*st)[flag] = (*st)[i]
			flag++
		}
	}

	*st = (*st)[:flag]
}

func (st StackTrace) Format(f fmt.State, verb rune) {
	errors.StackTrace(st).Format(f, verb)
}

// DefaultStackTraceError default implement of StackTraceError
type DefaultStackTraceError struct {
	tips     string
	code     string
	errMsg   string
	httpCode int
	err      error
}

// Error implements the error interface
// tips: %v \n
// error: %v
// stack trace...
func (d *DefaultStackTraceError) Error() string {
	var out strings.Builder
	out.WriteString("tips: ")
	out.WriteString(d.tips)
	out.WriteString("\n")
	out.WriteString("error: ")
	out.WriteString(d.ErrorMsg())
	out.WriteString(BuildStackTrace(1, 3, d.err))

	return out.String()
}

func (d *DefaultStackTraceError) Msg() string {
	return d.tips
}

func (d *DefaultStackTraceError) ErrorMsg() string {
	err := d.Cause()
	if err != nil {
		cause, ok := err.(causer)
		if ok {
			// 获取 pkg/errors中嵌套的错误消息
			return cause.Cause().Error()
		}

		return err.Error()
	}

	return ""

}

func (d *DefaultStackTraceError) AppendMsg(fn func(string) string) {
	d.tips = fn(d.tips)
}

func (d *DefaultStackTraceError) Code() string {
	return d.code
}

func (d *DefaultStackTraceError) SetCode(code string) {
	d.code = code
}

func (d *DefaultStackTraceError) Cause() error {
	return d.err
}

func (d *DefaultStackTraceError) HttpCode() int {
	return d.httpCode
}

func (d *DefaultStackTraceError) SetHttpCode(httpCode int) {
	d.httpCode = httpCode
}

// BuildStackTrace build stack trace
func BuildStackTrace(end, start int, err error) string {
	if err != nil {
		if tracerErr, ok := err.(stackTracer); ok {
			var all StackTrace
			causerErr := err
			for causerErr != nil {
				cause, ok := causerErr.(causer)
				if !ok {
					break
				}

				causerErr = cause.Cause()
				pkgErr, ok := causerErr.(stackTracer)
				if ok {
					// 提取每个error的栈的第一帧
					//all = append(all, pkgErr.StackTrace()[:end]...)
					//all = append(all, pkgErr.StackTrace()...)
					all = StackTrace(pkgErr.StackTrace())
				}
			}
			//all.Reverse()
			// 补充当前的堆栈
			//all = append(all, tracerErr.StackTrace()[start:]...)
			//all = append(all, tracerErr.StackTrace()...)
			if all == nil {
				all = StackTrace(tracerErr.StackTrace()[3:])
			}
			if Config.FilterStackTrace != nil {
				Config.FilterStackTrace(&all)
			}

			return fmt.Sprintf("%+v", all)
		}
	}

	return ""
}

// BuildStack build a stack error
func BuildStack(err error, msg string) (error, bool) {
	if err != nil {
		if stackTracerError, ok := err.(StackTraceError); ok {
			// 注意：stackTracerError最好不要嵌套 stackTracerError
			// 不然在调用ErrorMsg()方法时，会返回上一个的stackTracerError.Error()的堆栈信息
			// 导致错误消息过多，不好识别
			stackTracerError.AppendMsg(func(originMsg string) string {
				return originMsg + ": " + msg
			})
			return stackTracerError, true
		} else if _, ok := err.(stackTracer); !ok {
			err = errors.WithStack(err)
		}
	}
	return err, false
}
