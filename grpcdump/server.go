package grpcdump

import (
	"context"
	"errors"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

//var lis *bufconn.Listener

const addr = "bufnet"

type Server struct {
	listener *bufconn.Listener
	BufSize  int
}

func NewServer() *Server {
	return &Server{
		BufSize: bufSize,
	}
}

func (s *Server) DialContext(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(opts, grpc.WithContextDialer(s.bufDialer))
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (s *Server) ListenAndServe(fn func(*grpc.Server), opts ...grpc.ServerOption) func() {
	s.listener = bufconn.Listen(s.BufSize)

	srv := grpc.NewServer(opts...)
	fn(srv)

	done := make(chan bool)

	go func() {
		defer close(done)
		if err := srv.Serve(s.listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			panic(err)
		}
	}()

	return func() {
		srv.Stop()
		s.listener.Close()
		<-done
	}
}

func (s *Server) bufDialer(context.Context, string) (net.Conn, error) {
	return s.listener.Dial()
}
