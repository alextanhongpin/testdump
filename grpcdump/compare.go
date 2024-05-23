package grpcdump

import (
	"fmt"

	"github.com/alextanhongpin/testdump/pkg/diff"
	"github.com/google/go-cmp/cmp"
)

type CompareOption struct {
	Message  []cmp.Option
	Metadata []cmp.Option
	Trailer  []cmp.Option
	Header   []cmp.Option
}

func (snapshot *GRPC) Compare(received *GRPC, opt CompareOption, comparer func(a, b any, opts ...cmp.Option) error) error {
	x := snapshot
	y := received

	compare := diff.ANSI
	if err := compare(x.Addr, y.Addr); err != nil {
		return fmt.Errorf("Addr: %w", err)
	}

	if err := compare(x.FullMethod, y.FullMethod); err != nil {
		return fmt.Errorf("Full Method: %w", err)
	}

	if err := compare(x.Messages, y.Messages, opt.Message...); err != nil {
		return fmt.Errorf("Message: %w", err)
	}

	if err := compare(x.Status, y.Status); err != nil {
		return fmt.Errorf("Status: %w", err)
	}

	if err := compare(x.Metadata, y.Metadata, opt.Metadata...); err != nil {
		return fmt.Errorf("Metadata: %w", err)
	}

	if err := compare(x.Header, y.Header, opt.Header...); err != nil {
		return fmt.Errorf("Header: %w", err)
	}

	if err := compare(x.Trailer, y.Trailer, opt.Trailer...); err != nil {
		return fmt.Errorf("Trailer: %w", err)
	}

	return nil
}
