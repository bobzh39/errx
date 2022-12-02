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
	errx.Config.DefaultGRPCCode = uint32(codes.Internal)
	errx.Config.ErrorFactory = func(err error, tips string, withStack bool, opt ...errx.Option) errx.StackTraceError {
		st := errx.DefaultFactory(err, tips, withStack, opt...)
		return &GRPCStackTraceError{
			DefaultStackTraceError: st.(*errx.DefaultStackTraceError),
			grpcCode:               codes.Code(errx.Config.DefaultGRPCCode),
		}
	}
}

// GRPCError grpc code interface
type GRPCError interface {
	errx.StackTraceError
	// GRPCCode return a GRPCCode
	GRPCCode() codes.Code
	SetGRPCCode(code codes.Code)
}

// WithGRPCCode set a custom grpc code if the st is a GRPCError
func WithGRPCCode(code codes.Code) errx.Option {
	return func(st errx.StackTraceError) {
		if grpcError, ok := st.(GRPCError); ok {
			grpcError.SetGRPCCode(code)
		}
	}
}

// GRPCStackTraceError grpc error
type GRPCStackTraceError struct {
	*errx.DefaultStackTraceError
	grpcCode codes.Code
}

func (g *GRPCStackTraceError) GRPCCode() codes.Code {
	return g.grpcCode
}

func (g *GRPCStackTraceError) SetGRPCCode(code codes.Code) {
	g.grpcCode = code
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
