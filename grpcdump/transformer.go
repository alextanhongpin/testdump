package grpcdump

import (
	"fmt"

	"google.golang.org/grpc/metadata"
)

// Transformer is a type that defines a function that transforms a GRPC object.
// The function takes a pointer to a GRPC object and returns an error.
// If the transformation is successful, the function should return nil.
// If the transformation fails, the function should return an error.
type Transformer func(g *GRPC) error

func maskMetadata(md metadata.MD, mask string, keys []string) (metadata.MD, error) {
	if md.Len() == 0 {
		return md, nil
	}

	md = md.Copy()

	for _, k := range keys {
		if len(md.Get(k)) == 0 {
			return nil, fmt.Errorf("metadata key %q not found", k)
		}

		md[k] = sliceRepeat(mask, len(md[k]))
	}

	return md, nil
}

func sliceRepeat[T any](a T, n int) []T {
	res := make([]T, n)
	for i := 0; i < n; i++ {
		res[i] = a
	}

	return res
}
