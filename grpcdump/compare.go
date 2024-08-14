package grpcdump

import (
	"fmt"

	"github.com/alextanhongpin/testdump/pkg/diff"
	"github.com/google/go-cmp/cmp"
)

// CompareOption is a struct that holds comparison options for different parts of a gRPC message.
// It includes options for comparing the Message, Metadata, Trailer, and Header.
type CompareOption struct {
	Message  []cmp.Option // Options for comparing the message part of a gRPC message.
	Metadata []cmp.Option // Options for comparing the metadata part of a gRPC message.
	Trailer  []cmp.Option // Options for comparing the trailer part of a gRPC message.
	Header   []cmp.Option // Options for comparing the header part of a gRPC message.
}

type comparer struct {
	colors bool
	opt    CompareOption
}

func (c *comparer) Compare(a, b any) error {
	return c.compare(a.(*GRPC), b.(*GRPC))
}

// Compare is a method on the GRPC struct.
// It compares the current GRPC object (snapshot) with another GRPC object (received) using the provided CompareOption and comparer function.
// The comparer function should take two any objects and a slice of cmp.Option objects, and return an error.
// If the comparison is successful, the method should return nil.
// If the comparison fails, the method should return an error.
func (c *comparer) compare(snapshot, received *GRPC) error {
	x := snapshot
	y := received
	opt := c.opt

	comparer := diff.Text
	if c.colors {
		comparer = diff.ANSI
	}

	if err := comparer(x.Addr, y.Addr); err != nil {
		return fmt.Errorf("Addr: %w", err)
	}

	if err := comparer(x.FullMethod, y.FullMethod); err != nil {
		return fmt.Errorf("Full Method: %w", err)
	}

	if err := comparer(x.Messages, y.Messages, opt.Message...); err != nil {
		return fmt.Errorf("Message: %w", err)
	}

	if err := comparer(x.Status, y.Status); err != nil {
		return fmt.Errorf("Status: %w", err)
	}

	if err := comparer(x.Metadata, y.Metadata, opt.Metadata...); err != nil {
		return fmt.Errorf("Metadata: %w", err)
	}

	if err := comparer(x.Header, y.Header, opt.Header...); err != nil {
		return fmt.Errorf("Header: %w", err)
	}

	if err := comparer(x.Trailer, y.Trailer, opt.Trailer...); err != nil {
		return fmt.Errorf("Trailer: %w", err)
	}

	return nil
}
