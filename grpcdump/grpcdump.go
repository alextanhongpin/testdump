package grpcdump

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/alextanhongpin/testdump/grpcdump/internal"
	"github.com/alextanhongpin/testdump/pkg/diff"
)

var ErrMetadataNotFound = errors.New("grpcdump: metadata not found")

const OriginServer = "server"
const OriginClient = "client"

const grpcdumpTestID = "x-grpcdump-testid"

// NOTE: hackish implementation to extract the dump from the grpc server.
var testIds = make(map[string]*GRPC)
var mu sync.Mutex

var d *Dumper

func init() {
	d = new(Dumper)
}

type Dumper struct {
	opts []Option
}

func New(opts ...Option) *Dumper {
	return &Dumper{
		opts: opts,
	}
}

// Record is a method on the Dumper struct.
// It takes a testing.T object, a context, and a slice of Option objects, and returns a new context.
// The method configures the Dumper according to the provided options, then starts recording gRPC calls in the provided context.
// The returned context should be used in subsequent gRPC calls that should be recorded.
func (d *Dumper) Record(t *testing.T, ctx context.Context, opts ...Option) context.Context {
	id := uuid.New().String()

	t.Cleanup(func() {
		mu.Lock()
		g2c := testIds[id]
		delete(testIds, id)
		mu.Unlock()

		if err := d.dump(t, g2c, opts...); err != nil {
			t.Error(err)
		}
	})

	return metadata.AppendToOutgoingContext(ctx, grpcdumpTestID, id)
}

func (d *Dumper) dump(t *testing.T, received *GRPC, opts ...Option) error {
	opt := newOption(append(d.opts, opts...)...)

	for _, transform := range opt.transformers {
		if err := transform(received); err != nil {
			return err
		}
	}

	file := filepath.Join("testdata", fmt.Sprintf("%s.grpc", t.Name()))

	receivedBytes, err := Write(received)
	if err != nil {
		return err
	}

	overwrite, _ := strconv.ParseBool(os.Getenv(opt.env))
	written, err := internal.WriteFile(file, receivedBytes, overwrite)
	if err != nil {
		return err
	}

	// First write, there's nothing to compare.
	if written {
		return nil
	}

	// Read the snapshot data from the file.
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	snapshot, err := Read(b)
	if err != nil {
		return err
	}

	comparer := diff.Text
	if opt.colors {
		comparer = diff.ANSI
	}

	return snapshot.Compare(received, opt.cmpOpt, comparer)
}

// NewRecorder is a function that creates a new recorder for gRPC calls.
// It takes a testing.T object, a context, and a slice of Option objects, and returns a new context.
// The function delegates the recording task to the Record method of the Dumper object.
// The returned context should be used in subsequent gRPC calls that should be recorded.
func NewRecorder(t *testing.T, ctx context.Context, opts ...Option) context.Context {
	return d.Record(t, ctx, opts...)
}

// Message is a struct that represents a gRPC message.
// It contains all the necessary fields for a complete gRPC message.
type Message struct {
	Origin  string `json:"origin"` // server or client
	Name    string `json:"name"`   // message type (protobuf name)
	Message any    `json:"message"`
}

type serverStreamInterceptor struct {
	grpc.ServerStream
	header   metadata.MD
	messages []Message
	trailer  metadata.MD
}

func (s *serverStreamInterceptor) SetTrailer(md metadata.MD) {
	s.ServerStream.SetTrailer(md)

	s.trailer = metadata.Join(s.trailer, md)
}

func (s *serverStreamInterceptor) SendHeader(md metadata.MD) error {
	err := s.ServerStream.SendHeader(md)
	s.header = metadata.Join(s.header, md)

	return err
}

func (s *serverStreamInterceptor) SetHeader(md metadata.MD) error {
	err := s.ServerStream.SetHeader(md)
	s.header = metadata.Join(s.header, md)

	return err
}

func (s *serverStreamInterceptor) SendMsg(m interface{}) error {
	err := s.ServerStream.SendMsg(m)
	if err == nil {
		s.messages = append(s.messages, origin(OriginServer, m))
	}

	return err
}

func (s *serverStreamInterceptor) RecvMsg(m interface{}) error {
	err := s.ServerStream.RecvMsg(m)
	if err == nil {
		s.messages = append(s.messages, origin(OriginClient, m))
	}

	return err
}

// StreamInterceptor is a function that returns a grpc.ServerOption.
// This ServerOption, when applied, configures the server to use the StreamServerInterceptor function as the stream interceptor.
// The stream interceptor is a function that intercepts incoming streaming RPCs on the server.
func StreamInterceptor() grpc.ServerOption {
	return grpc.StreamInterceptor(StreamServerInterceptor)
}

// UnaryInterceptor is a function that returns a grpc.ServerOption.
// This ServerOption, when applied, configures the server to use the UnaryServerInterceptor function as the unary interceptor.
// The unary interceptor is a function that intercepts incoming unary RPCs on the server.
func UnaryInterceptor() grpc.ServerOption {
	return grpc.UnaryInterceptor(UnaryServerInterceptor)
}

// WithUnaryInterceptor is a function that returns a grpc.DialOption.
// This DialOption, when applied, configures the client to use the UnaryClientInterceptor function as the unary interceptor.
// The unary interceptor is a function that intercepts outgoing unary RPCs on the client.
func WithUnaryInterceptor() grpc.DialOption {
	return grpc.WithUnaryInterceptor(UnaryClientInterceptor)
}

// StreamServerInterceptor is a function that intercepts incoming streaming RPCs on the server.
// It takes a server, a grpc.ServerStream, a grpc.StreamServerInfo, and a grpc.StreamHandler, and returns an error.
// If the interception is successful, it should return nil.
// If the interception fails, it should return an error.
func StreamServerInterceptor(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := stream.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ErrMetadataNotFound
	}

	// Extract the test-id from the header.
	// We do not want to log this, so delete it from the
	// existing header.
	id := md.Get(grpcdumpTestID)[0]
	md.Delete(grpcdumpTestID)

	w := &serverStreamInterceptor{
		ServerStream: stream,
	}
	err := handler(srv, w)

	mu.Lock()
	testIds[id] = &GRPC{
		Addr:           addrFromContext(ctx),
		FullMethod:     info.FullMethod,
		Metadata:       md,
		Messages:       w.messages,
		Trailer:        w.trailer,
		Header:         w.header,
		Status:         newStatus(err),
		IsServerStream: info.IsServerStream,
		IsClientStream: info.IsClientStream,
	}
	mu.Unlock()

	return err
}

// UnaryServerInterceptor is a function that intercepts incoming unary RPCs on the server.
// It takes a context, a request, a grpc.UnaryServerInfo, and a grpc.UnaryHandler, and returns a response and an error.
// If the interception is successful, it should return the response and nil.
// If the interception fails, it should return nil and the error.
func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrMetadataNotFound
	}

	// Extract the test-id from the header.
	// We do not want to log this, so delete it from the
	// existing header.
	id := md.Get(grpcdumpTestID)[0]
	md.Delete(grpcdumpTestID)

	res, err := handler(ctx, req)
	messages := []Message{origin(OriginClient, req)}

	if err == nil {
		messages = append(messages, origin(OriginServer, res))
	}

	mu.Lock()
	testIds[id] = &GRPC{
		Addr:       addrFromContext(ctx),
		FullMethod: info.FullMethod,
		Metadata:   md,
		Messages:   messages,
		Status:     newStatus(err),
	}
	mu.Unlock()

	return res, err
}

// UnaryClientInterceptor is a function that intercepts outgoing unary RPCs on the client.
// It takes a context, a method string, a request, a response, a grpc.ClientConn, a grpc.UnaryInvoker, and a slice of grpc.CallOption, and returns an error.
// If the interception is successful, it should return nil.
// If the interception fails, it should return the error.
func UnaryClientInterceptor(ctx context.Context, method string, req, res any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return ErrMetadataNotFound
	}

	testID := md.Get(grpcdumpTestID)[0]

	ctx = metadata.NewOutgoingContext(ctx, md)

	var header, trailer metadata.MD
	opts = append(opts, grpc.Header(&header), grpc.Trailer(&trailer))

	if err := invoker(ctx, method, req, res, cc, opts...); err != nil {
		return err
	}

	header.Delete(grpcdumpTestID)

	mu.Lock()
	testIds[testID].Trailer = trailer
	testIds[testID].Header = header
	mu.Unlock()

	return nil
}

func addrFromContext(ctx context.Context) string {
	var addr string
	if pr, ok := peer.FromContext(ctx); ok {
		if tcpAddr, ok := pr.Addr.(*net.TCPAddr); ok {
			addr = tcpAddr.IP.String()
		} else {
			addr = pr.Addr.String()
		}
	}
	return addr
}

func origin(origin string, v any) Message {
	msg, ok := v.(interface {
		ProtoReflect() protoreflect.Message
	})
	if !ok {
		panic("grpcdump: message is not valid")
	}
	m, err := toMap(v)
	if err != nil {
		panic(err)
	}

	return Message{
		Origin:  origin,
		Name:    fmt.Sprint(msg.ProtoReflect().Descriptor().FullName()),
		Message: m,
	}
}

func toMap(v any) (any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var a any
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, err
	}
	return a, nil
}
