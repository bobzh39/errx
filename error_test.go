package errx

import (
	serr "errors"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func pkgErrors1() error {
	return errors.Wrap(errors2(), "err of one")
}
func errors2() error {
	return errors.Wrap(errors3(), "err of two")
}
func errors3() error {
	err := errors4()
	if err != nil {
		return err
	}
	return nil
	//return errors.Wrap(err, "err of three")
}
func errors4() error {
	return errors.Wrap(errors5(), "err of four")
}
func errors5() error {
	return serr.New("this is go errors")
}

func do() error {
	_, err := os.Open("nil")
	if err != nil {
		return Wrap(err)
	}

	return nil
}

func panic1() {
	panic2()
}

func panic2() {
	panic3()
}

func panic3() {
	panic("go routines panic")
}

func read() error {
	err := do()
	if err != nil {
		return WrapMessage(err, "read fail")
	}

	return nil
}

func TestWrap(t *testing.T) {
	TestErrConfig(t)
	err := Wrap(pkgErrors1())
	t.Log(err)

}

func TestWrapMessage(t *testing.T) {
	err := WrapMessage(pkgErrors1(), "server is busy")
	t.Log(err)
}

func TestDoSomething(t *testing.T) {
	err := do()
	t.Log(err)
}

func TestSameErr(t *testing.T) {
	err := read()
	t.Log(err)
}

func TestFilter(t *testing.T) {
	Config.FilterStackTrace = func(trace *StackTrace) {
		var removes []errors.Frame
		trace.Reverse()
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
	err := read()
	t.Log(err)
}

func TestPanicTrace(t *testing.T) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Log(PanicTrace(r))
			}
		}()
		panic1()
	}()

	time.Sleep(5 * time.Second)
}

func TestErrConfig(t *testing.T) {
	Config = ErrorConfig{
		DefaultTips:     "busy",
		DefaultCode:     "500",
		DefaultHttpCode: 400,
		DefaultGRPCCode: 16,
	}
	assert.Equal(t, Config.DefaultTips, Config.DefaultTips)
	assert.Equal(t, Config.DefaultCode, Config.DefaultCode)
	assert.Equal(t, Config.DefaultHttpCode, Config.DefaultHttpCode)
	assert.Equal(t, Config.DefaultGRPCCode, Config.DefaultGRPCCode)
}
