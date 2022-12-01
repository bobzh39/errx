package gerrx

import (
	"fmt"

	"github.com/bobzh39/errx"
	epb "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoadGRPCError 使用GRPC error
func LoadGRPCError() {
	errx.ErrorFactory = func(err error, msg string, opt ...errx.Option) error {
		stackTrace := errx.New(err, msg, opt...)
		grpcErr := &GRPCStackTraceError{
			DefaultStackTraceError: stackTrace.(*errx.DefaultStackTraceError),
			grpcCode:               codes.Code(errx.Config.DefaultGRPCCode),
		}

		for i := range opt {
			opt[i](grpcErr)
		}

		return grpcErr
	}
}

// GRPCStackTraceError grpc error
type GRPCStackTraceError struct {
	*errx.DefaultStackTraceError
	grpcCode codes.Code
}

func (g *GRPCStackTraceError) GRPCCode() uint32 {
	return uint32(g.grpcCode)
}

func (g *GRPCStackTraceError) SetGRPCCode(code uint32) {
	g.grpcCode = codes.Code(code)
}

// GRPCStatus grpc status code返回
func (g *GRPCStackTraceError) GRPCStatus() *status.Status {
	s := status.New(g.grpcCode, g.Msg())
	//proto.
	//proto.m
	res, err := s.WithDetails(&epb.ResourceInfo{
		ResourceName: g.Code(),
		Description:  g.Error(),
	})

	if err != nil {
		fmt.Println("GRPCStackTraceError.GRPCStatus error:", err)
		return s
	}

	return res
}