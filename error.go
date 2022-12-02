package errx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"unsafe"

	"github.com/pkg/errors"
)

const (
	JSON = iota
	Text
)

var (
	Marshal = json.Marshal
)

type (
	Option           func(StackTraceError)
	ErrorFactoryFunc func(err error, tips string, withStack bool, opt ...Option) StackTraceError
)

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

// WithField append a field
func WithField(key string, value any) Option {
	return func(st StackTraceError) {
		if ctx, ok := st.(FieldContext); ok {
			ctx.Append(Field(key, value))
		}
	}
}

// WithFields append fields
func WithFields(fields ...LogField) Option {
	return func(st StackTraceError) {
		if ctx, ok := st.(FieldContext); ok {
			ctx.Append(fields...)
		}
	}
}

// WithTips create an error by GlobalErrorFactory and no stack log.
// could be used WithCode WithHttpCode to define a biz code
func WithTips(tips string, opt ...Option) error {
	return GlobalErrorFactory(nil, tips, opt...)
}

// Wrap create an error by GlobalErrorFactory
// could be used WithCode WithHttpCode to define a biz code
// if err is StackTraceError will append tips and return
func Wrap(err error, opt ...Option) error {
	return GlobalErrorFactory(err, Config.DefaultTips, opt...)
}

// WrapMessage create an error by GlobalErrorFactory
// could be used WithCode WithHttpCode to define a biz code
// if err is StackTraceError will append the tips and return
func WrapMessage(err error, tips string, opt ...Option) error {
	return GlobalErrorFactory(err, tips, opt...)
}

// New create a StackTraceError
// when err is nil, it will not stack trace
func New(err error, tips string, opt ...Option) error {
	withStack := false
	if err != nil {
		if stackTracerError, ok := err.(StackTraceError); ok {
			// 注意：stackTracerError最好不要嵌套 stackTracerError
			// 不然在调用ErrorMsg()方法时，会返回上一个的stackTracerError.Error()的堆栈信息
			// 导致错误消息过多，不好识别
			stackTracerError.AppendMsg(func(originMsg string) string {
				return originMsg + ": " + tips
			})

			return stackTracerError
		} else if _, ok := err.(stackTracer); !ok {
			err = errors.WithStack(err)
			withStack = true
		}
	}

	var st StackTraceError
	if Config.ErrorFactory != nil {
		st = Config.ErrorFactory(err, tips, withStack, opt...)
	} else {
		st = DefaultFactory(err, tips, withStack, opt...)
	}

	for i := range opt {
		opt[i](st)
	}

	return st
}

// Config error global ErrorConfig
var (
	_                  StackTraceError = (*DefaultStackTraceError)(nil)
	Config             ErrorConfig
	GlobalErrorFactory = New
)

// DefaultFactory default StackTraceError factory
func DefaultFactory(err error, tips string, withStack bool, o ...Option) StackTraceError {
	return &DefaultStackTraceError{
		tips:      tips,
		code:      Config.DefaultCode,
		httpCode:  Config.DefaultHttpCode,
		err:       err,
		withStack: withStack,
	}
}

func init() {
	Config = ErrorConfig{
		DefaultTips:     "服务器繁忙",
		DefaultCode:     "InternalError",
		DefaultHttpCode: http.StatusBadRequest,
		Skip:            2,
		ErrorFormat:     Text,
	}
}

type (
	// ErrorConfig error global ErrorConfig
	// Note: These are the code for error conditions is not normal
	ErrorConfig struct {
		DefaultTips     string
		DefaultCode     string
		DefaultHttpCode int
		DefaultGRPCCode uint32
		Skip            int
		ErrorFormat     int
		// filter stack func
		FilterStackTrace func(*StackTrace)
		ErrorFactory     ErrorFactoryFunc
	}

	// HttpError http error 自定义code
	HttpError interface {
		StackTraceError
		HttpCode() int
		SetHttpCode(int)
	}
	// StackTraceError 有堆栈错误的error，包括自定义biz code
	StackTraceError interface {
		FieldContext
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
	// withStack err是否为内部创建的with stack
	withStack bool
	fields    map[string]any
}

// Error implements the error interface
// tips: %v \n
// error: %v
// stack trace...
func (d *DefaultStackTraceError) Error() string {
	trace := BuildStackTrace(func(trace StackTrace) StackTrace {
		if d.withStack {
			return trace[Config.Skip:]
		}
		return trace
	}, d.err)

	if Config.ErrorFormat == Text {
		var out strings.Builder
		out.WriteString("tips: ")
		out.WriteString(d.tips)
		errorMsg := d.ErrorMsg()
		if errorMsg != "" {
			out.WriteString("\n")
			out.WriteString("error: ")
			out.WriteString(errorMsg)
		}
		if len(d.fields) > 0 {
			for k, v := range d.fields {
				out.WriteString(fmt.Sprintf("\t%s=%v", k, v))
			}
		}

		out.WriteString(trace)

		return out.String()
	}

	jsonMap := make(map[string]any, 4)
	jsonMap["tips"] = d.tips
	jsonMap["error"] = d.ErrorMsg()
	jsonMap["fields"] = d.fields
	jsonMap["calls"] = trace
	data, err := Marshal(jsonMap)
	if err != nil {
		return fmt.Sprintf("%+v", errors.Wrap(err, "marshal StackTraceError error"))
	}

	return *(*string)(unsafe.Pointer(&data))
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

func (d *DefaultStackTraceError) Append(fields ...LogField) {
	if len(fields) == 0 {
		return
	}

	if d.fields == nil {
		d.fields = make(map[string]any, len(d.fields))
	}
	for i := range fields {
		d.fields[fields[i].key] = fields[i].value
	}
}

// BuildStackTrace build stack trace
func BuildStackTrace(skip func(StackTrace) StackTrace, err error) string {
	if skip == nil {
		skip = func(trace StackTrace) StackTrace { return trace }
	}
	if err != nil {
		if tracerErr, ok := err.(stackTracer); ok {
			var all StackTrace
			causerErr := err

			// pkg/errors 嵌套多层处理
			for causerErr != nil {
				cause, ok := causerErr.(causer)
				if !ok {
					break
				}

				causerErr = cause.Cause()
				pkgErr, ok := causerErr.(stackTracer)
				if ok {
					// 提取最深的堆栈
					all = StackTrace(pkgErr.StackTrace())
				}
			}

			if all == nil {
				all = skip(StackTrace(tracerErr.StackTrace()))
			}

			if Config.FilterStackTrace != nil {
				Config.FilterStackTrace(&all)
			}

			return fmt.Sprintf("%+v", all)
		}
	}

	return ""
}

// PanicTrace 用于panic recover返回的堆栈信息，比如常见的使用goroutine时的recover
func PanicTrace(err any) string {
	buf := new(bytes.Buffer)
	_, _ = fmt.Fprintf(buf, "%v\n", err)
	for i := 1; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		_, _ = fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
	}
	return buf.String()
}
