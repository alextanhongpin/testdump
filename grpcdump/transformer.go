package grpcdump

import (
	"fmt"

	"google.golang.org/grpc/metadata"
)

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
