package grpcdump_test

import (
	"testing"

	"github.com/alextanhongpin/testdump/grpcdump"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func TestDump(t *testing.T) {
	d := &grpcdump.GRPC{
		Addr:       "bufconn",
		FullMethod: "/helloworld.v1.GreeterService/Chat",
		Status: &grpcdump.Status{
			Code:    codes.Unauthenticated.String(),
			Message: "not authenticated",
		},
		Metadata: metadata.New(map[string]string{
			"md-key":     "md-val",
			"md-key-bin": "md-val-bin",
		}),
		Header: metadata.New(map[string]string{
			"header-key":     "header-val",
			"header-key-bin": "header-val-bin",
		}),
		Trailer: metadata.New(map[string]string{
			"trailer-key":     "trailer-val",
			"trailer-key-bin": "trailer-val-bin",
		}),
		Messages: []grpcdump.Message{
			{
				Origin: grpcdump.OriginClient,
				Message: map[string]any{
					"msg": "Hello",
				},
				Name: "Message",
			},
			{
				Origin: grpcdump.OriginServer,
				Message: map[string]any{
					"msg": "Hi",
				},
				Name: "Message",
			},
		},
	}

	b, err := grpcdump.Write(d)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(b))
	g, err := grpcdump.Read(b)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(g)
}
