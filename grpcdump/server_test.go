package grpcdump_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alextanhongpin/testdump/grpcdump"
	pb "github.com/alextanhongpin/testdump/grpcdump/testdata/helloworld/v1"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/examples/data"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var s *grpcdump.Server

func TestMain(m *testing.M) {
	cert, err := tls.LoadX509KeyPair(data.Path("x509/server_cert.pem"), data.Path("x509/server_key.pem"))
	if err != nil {
		panic(err)
	}

	s = grpcdump.NewServer(
		// Setup grpcdump on the server side.
		// Also set on the client side, but only the `WithUnaryInterceptor`.
		grpcdump.StreamInterceptor(),
		grpc.ChainUnaryInterceptor(
			grpcdump.UnaryServerInterceptor,
			ensureValidToken,
		),
		grpc.Creds(credentials.NewServerTLSFromCert(&cert)),
	)

	pb.RegisterGreeterServiceServer(s.Server, &server{})

	stop := s.ListenAndServe()
	code := m.Run()
	stop()
	os.Exit(code)
}

func TestGRPCClientStreaming(t *testing.T) {
	t.Setenv("GODEBUG", "x509sha1=1")

	ctx := context.Background()
	conn := grpcDialContext(t, ctx)

	// Create a new client.
	client := pb.NewGreeterServiceClient(conn)

	// Create a new recorder.
	ctx = grpcdump.NewRecorder(t, ctx, grpcdump.IgnoreMetadata("user-agent"))
	ctx = metadata.AppendToOutgoingContext(ctx,
		"md-val", "md-val",
		"md-val-bin", "md-val-bin",
	)

	assert := assert.New(t)
	stream, err := client.RecordGreetings(ctx)
	assert.Nil(err, err)

	// Send 5 greetings.
	n := 5
	for i := 0; i < n; i++ {
		err := stream.Send(&pb.RecordGreetingsRequest{
			Message: "hi sir",
		})
		assert.Nil(err)
	}

	reply, err := stream.CloseAndRecv()
	assert.Nil(err)
	assert.Equal(reply.GetCount(), int64(n))
}

func TestGRPCBidirectionalStreaming(t *testing.T) {
	t.Setenv("GODEBUG", "x509sha1=1")

	ctx := context.Background()
	conn := grpcDialContext(t, ctx)

	// Create a new client.
	client := pb.NewGreeterServiceClient(conn)

	// Create a new recorder.
	ctx = grpcdump.NewRecorder(t, ctx, grpcdump.IgnoreMetadata("user-agent"))
	ctx = metadata.AppendToOutgoingContext(ctx,
		"md-val", "md-val",
		"md-val-bin", "md-val-bin",
	)

	assert := assert.New(t)
	stream, err := client.Chat(ctx)
	assert.Nil(err)

	done := make(chan bool)

	go func() {
		for {
			_, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			assert.Nil(err)
		}
	}()

	for _, msg := range []string{"foo", "bar"} {
		err := stream.Send(&pb.ChatRequest{
			Message: msg,
		})
		assert.Nil(err)
	}
	stream.CloseSend()

	<-done
}

func TestGRPCServerStreaming(t *testing.T) {
	t.Setenv("GODEBUG", "x509sha1=1")

	t.Run("success", func(t *testing.T) {
		assert := assert.New(t)
		err := testServerStreaming(t, &pb.ListGreetingsRequest{
			Count: 5,
		})
		assert.Nil(err)
	})

	t.Run("failed", func(t *testing.T) {
		assert := assert.New(t)
		err := testServerStreaming(t, &pb.ListGreetingsRequest{
			Count: -99,
		})
		assert.NotNil(err)

		// Convert the error to gRPC status.
		st := status.Convert(err)
		for _, detail := range st.Details() {
			switch v := detail.(type) {
			case *errdetails.BadRequest:
				for _, violation := range v.FieldViolations {
					assert.Equal("Count", violation.GetField())
					assert.Equal("Count cannot be negative", violation.GetDescription())
				}
			default:
				t.Fatalf("unhandled error: %v", v)
			}
		}
	})

	t.Run("zero", func(t *testing.T) {
		// NOTE: In testdata/TestGRPCServerStreaming/zero.http, you will see the
		// ListGreetingsRequest to be `{}`. This is expected, as zero values won't
		// be serialized.
		// https://protobuf.dev/programming-guides/proto3/#default:~:text=Note%20that%20for,on%20the%20wire.
		assert := assert.New(t)
		err := testServerStreaming(t, &pb.ListGreetingsRequest{
			Count: 0,
		})
		assert.NotNil(err)

		// Convert the error to gRPC status.
		st := status.Convert(err)
		for _, detail := range st.Details() {
			switch v := detail.(type) {
			case *errdetails.BadRequest:
				for _, violation := range v.FieldViolations {
					assert.Equal("Count", violation.GetField())
					assert.Equal("Count cannot be negative", violation.GetDescription())
				}
			default:
				t.Fatalf("unhandled error: %v", v)
			}
		}
	})
}

func TestGRPCUnary(t *testing.T) {
	t.Setenv("GODEBUG", "x509sha1=1")
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		conn := grpcDialContext(t, ctx)

		// Create a new client.
		client := pb.NewGreeterServiceClient(conn)

		// Create a new recorder.
		ctx = grpcdump.NewRecorder(t, ctx,
			grpcdump.MaskMetadata("[MASKED]", []string{"authorization"}),
			grpcdump.IgnoreMetadata("user-agent"),
		)

		ctx = metadata.AppendToOutgoingContext(ctx,
			"md-val", "md-val",
			"md-val-bin", "md-val-bin",
		)

		_, err := client.SayHello(ctx, &pb.SayHelloRequest{
			Name: "John Doe",
		})
		assert.Nil(t, err)
	})

	t.Run("unauthorized", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, tokenCtxKey, "abc")

		conn := grpcDialContext(t, ctx)

		// Create a new client.
		client := pb.NewGreeterServiceClient(conn)

		// Create a new recorder.
		ctx = grpcdump.NewRecorder(t, ctx,
			grpcdump.IgnoreMetadata("user-agent"),
		)

		ctx = metadata.AppendToOutgoingContext(ctx,
			"md-val", "md-val",
			"md-val-bin", "md-val-bin",
		)

		// Anything linked to this variable will fetch response headers.
		_, err := client.SayHello(ctx, &pb.SayHelloRequest{
			Name: "John Doe",
		})
		assert.NotNil(t, err)
	})
}

func TestMaskOptions(t *testing.T) {
	t.Setenv("GODEBUG", "x509sha1=1")

	tests := []struct {
		name string
		opts []grpcdump.Option
	}{
		{
			name: "mask metadata",
			opts: []grpcdump.Option{
				grpcdump.MaskMetadata("[MASKED]", []string{"authorization"}),
				grpcdump.IgnoreMetadata("user-agent"),
			},
		},
		{
			name: "mask trailer",
			opts: []grpcdump.Option{
				grpcdump.MaskTrailer("[MASKED]", []string{"trailer-key"}),
				grpcdump.IgnoreMetadata("user-agent"),
			},
		},
		{
			name: "mask header",
			opts: []grpcdump.Option{
				grpcdump.MaskHeader("[MASKED]", []string{"header-key"}),
				grpcdump.IgnoreMetadata("user-agent"),
			},
		},
		{
			name: "mask message fields",
			opts: []grpcdump.Option{
				grpcdump.MaskMessageFields("[MASKED]", []string{"name"}),
				grpcdump.IgnoreMetadata("user-agent"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			conn := grpcDialContext(t, ctx)
			client := pb.NewGreeterServiceClient(conn)
			ctx = grpcdump.NewRecorder(t, ctx, tt.opts...)
			_, err := client.SayHello(ctx, &pb.SayHelloRequest{
				Name: "John Doe",
			})
			assert.Nil(t, err)
		})
	}
}

func TestIgnoreOptions(t *testing.T) {
	ctx := context.Background()

	gs := grpcdump.NewServer(
		// Setup grpcdump on the server side.
		// Also set on the client side, but only the `WithUnaryInterceptor`.
		grpcdump.StreamInterceptor(),
		grpcdump.UnaryInterceptor(),
		/* If you need to chain multiple unary server interceptor, do this:
		grpc.ChainUnaryInterceptor(
			grpcdump.UnaryServerInterceptor,
			ensureValidToken,
		),
		*/
	)

	pb.RegisterGreeterServiceServer(gs.Server, &server{dynamic: true}) // Set dynamic payload
	stop := gs.ListenAndServe()
	defer stop()

	conn, err := gs.DialContext(ctx,
		// Setup grpcdump on the client side.
		grpcdump.WithUnaryInterceptor(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		conn.Close()
	})

	client := pb.NewGreeterServiceClient(conn)
	ctx = grpcdump.NewRecorder(t, ctx,
		grpcdump.IgnoreMessageFields("name", "message"),
		grpcdump.IgnoreHeader("header-date"),
		grpcdump.IgnoreTrailer("trailer-date"),
		grpcdump.IgnoreMetadata("metadata-date", "user-agent"),
	)

	ctx = metadata.AppendToOutgoingContext(ctx,
		"metadata-date", time.Now().Format(time.RFC3339),
	)

	_, err = client.SayHello(ctx, &pb.SayHelloRequest{
		Name: time.Now().Format(time.RFC3339), // Dynamic.
	})
	assert.Nil(t, err)
}

type server struct {
	pb.UnimplementedGreeterServiceServer
	dynamic bool
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.SayHelloRequest) (*pb.SayHelloResponse, error) {
	ctx = metadata.NewOutgoingContext(ctx, nil)

	headers := []string{
		"header-key", "header-val",
		"header-key-bin", "header-val-bin",
	}
	if s.dynamic {
		headers = append(headers, "header-date", time.Now().Format(time.RFC3339))
	}
	ctx = metadata.AppendToOutgoingContext(ctx, headers...)
	header, _ := metadata.FromOutgoingContext(ctx)

	// For unary.
	if err := grpc.SendHeader(ctx, header); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to send gRPC header: %v", err)
	}

	trailer := metadata.New(map[string]string{
		"trailer-key":     "trailer-val",
		"trailer-key-bin": "trailer-val-bin",
	})

	if s.dynamic {
		trailer.Set("trailer-date", time.Now().Format(time.RFC3339))
	}

	// For unary.
	if err := grpc.SetTrailer(ctx, trailer); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to send gRPC trailer: %v", err)
	}

	return &pb.SayHelloResponse{Message: "Hello " + in.GetName()}, nil
}

func (s *server) RecordGreetings(stream pb.GreeterService_RecordGreetingsServer) error {
	ctx := stream.Context()
	ctx = metadata.AppendToOutgoingContext(ctx,
		"header-key", "header-val",
		"header-key-bin", "header-val-bin",
	)

	if header, ok := metadata.FromOutgoingContext(ctx); ok {
		// For stream.
		if err := stream.SendHeader(header); err != nil {
			return status.Errorf(codes.Internal, "unable to send gRPC header: %v", err)
		}
	}

	trailer := metadata.New(map[string]string{
		"trailer-key":     "trailer-val",
		"trailer-key-bin": "trailer-val-bin",
	})

	// For unary.
	stream.SetTrailer(trailer)

	var msgs []string
	for {
		msg, err := stream.Recv()
		if err == io.EOF {

			return stream.SendAndClose(&pb.RecordGreetingsResponse{
				Count: int64(len(msgs)),
			})
		}
		if err != nil {
			return err
		}
		msgs = append(msgs, msg.GetMessage())
	}
}

func (s *server) ListGreetings(in *pb.ListGreetingsRequest, stream pb.GreeterService_ListGreetingsServer) error {
	ctx := stream.Context()
	ctx = metadata.AppendToOutgoingContext(ctx,
		"header-key", "header-val",
		"header-key-bin", "header-val-bin",
	)

	header, _ := metadata.FromOutgoingContext(ctx)

	// For stream.
	if err := stream.SendHeader(header); err != nil {
		return status.Errorf(codes.Internal, "unable to send gRPC header: %v", err)
	}

	n := in.GetCount()
	if n <= 0 {
		// Implement custom error using errdetails.
		var d errdetails.BadRequest
		d.FieldViolations = append(d.FieldViolations, &errdetails.BadRequest_FieldViolation{
			Field:       "Count",
			Description: "Count cannot be negative",
		})

		st := status.New(codes.InvalidArgument, "Failed to get count")
		st, err := st.WithDetails(&d)
		if err != nil {
			panic(err)
		}

		return st.Err()
	}

	for i := 0; i < int(n); i++ {
		err := stream.Send(&pb.ListGreetingsResponse{
			Message: fmt.Sprintf("hi sir (%d)", i+1),
		})
		if err != nil {
			return err
		}
	}

	trailer := metadata.New(map[string]string{
		"trailer-key":     "trailer-val",
		"trailer-key-bin": "trailer-val-bin",
	})

	// For stream.
	stream.SetTrailer(trailer)

	return nil
}

func (s *server) Chat(stream pb.GreeterService_ChatServer) error {
	ctx := stream.Context()
	ctx = metadata.AppendToOutgoingContext(ctx,
		"header-key", "header-val",
		"header-key-bin", "header-val-bin",
	)

	header, _ := metadata.FromOutgoingContext(ctx)
	if err := stream.SendHeader(header); err != nil {
		return status.Errorf(codes.Internal, "unable to send gRPC header: %v", err)
	}

	trailer := metadata.New(map[string]string{
		"trailer-key":     "trailer-val",
		"trailer-key-bin": "trailer-val-bin",
	})
	stream.SetTrailer(trailer)

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		if err := stream.Send(&pb.ChatResponse{
			Message: "REPLY: " + in.GetMessage(),
		}); err != nil {
			return err
		}
	}
}

func testServerStreaming(t *testing.T, req *pb.ListGreetingsRequest) error {
	t.Helper()

	ctx := context.Background()
	conn := grpcDialContext(t, ctx)

	// Create a new client.
	client := pb.NewGreeterServiceClient(conn)

	// Create a new recorder.
	ctx = grpcdump.NewRecorder(t, ctx, grpcdump.IgnoreMetadata("user-agent"))

	ctx = metadata.AppendToOutgoingContext(ctx,
		"md-val", "md-val",
		"md-val-bin", "md-val-bin",
	)

	assert := assert.New(t)
	stream, err := client.ListGreetings(ctx, req)
	assert.Nil(err)

	ch := make(chan error)

	go func() {
		for {
			_, err := stream.Recv()
			if err == io.EOF {
				ch <- nil
				return
			}
			if err != nil {
				ch <- err
				return
			}
		}
	}()

	return <-ch
}

func ensureValidToken(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "missing metadata")
	}

	authz := md["authorization"]
	if len(authz) < 1 {
		return nil, status.Error(codes.Unauthenticated, "no token present")
	}

	token := strings.TrimPrefix(authz[0], "Bearer ")
	if token != "xyz" {
		return nil, status.Error(codes.Unauthenticated, "token expired")
	}

	return handler(ctx, req)
}

func grpcDialContext(t *testing.T, ctx context.Context, opts ...grpc.DialOption) *grpc.ClientConn {
	t.Helper()

	token, ok := ctx.Value(tokenCtxKey).(string)
	if !ok {
		// Assign a default token.
		token = "xyz"
	}

	perRPC := oauth.TokenSource{
		TokenSource: oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: token,
		}),
	}

	// Setup credentials for the connection.
	creds, err := credentials.NewClientTLSFromFile(data.Path("x509/ca_cert.pem"), "x.test.example.com")
	if err != nil {
		t.Fatal(err)
	}

	conn, err := s.DialContext(ctx,
		// Setup grpcdump on the client side.
		grpcdump.WithUnaryInterceptor(),
		grpc.WithPerRPCCredentials(perRPC),
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		conn.Close()
	})

	return conn
}

type ctxKey string

var tokenCtxKey ctxKey = "token"
