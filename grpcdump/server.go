package grpcdump

import (
	"context"
	"errors"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

const addr = "bufnet"

// Server is a struct that represents a gRPC server.
// It contains all the necessary fields and methods for managing a gRPC server.
type Server struct {
	listener *bufconn.Listener
	BufSize  int
	Server   *grpc.Server
}

// NewServer is a function that creates a new Server instance.
// It takes a variadic parameter of type grpc.ServerOption, which are options to configure the Server.
// It returns a pointer to the newly created Server.
func NewServer(opts ...grpc.ServerOption) *Server {
	return &Server{
		BufSize: bufSize,
		Server:  grpc.NewServer(opts...),
	}
}

// DialContext is a method on the Server struct that creates a new gRPC client connection.
// It takes a context and a variadic parameter of type grpc.DialOption, which are options to configure the connection.
// It returns a pointer to the newly created grpc.ClientConn and an error.
// If the connection is successful, the error is nil. Otherwise, the error contains the details of the failure.
func (s *Server) DialContext(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(opts, grpc.WithContextDialer(s.bufDialer))
	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// ListenAndServe is a method on the Server struct.
// It starts the server and listens for incoming connections.
// It returns a function that, when called, stops the server.
func (s *Server) ListenAndServe() func() {
	srv := s.Server
	s.listener = bufconn.Listen(s.BufSize)

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
