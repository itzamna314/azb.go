package azb

import (
	"errors"
	"strings"
)

var (
	ErrBadBlobSpec = errors.New("malformed blobspec")
)

type BlobSpec struct {
	Container   string
	Path        string
	PathPresent bool
}

func ParseBlobSpec(s string) (*BlobSpec, error) {
	if s == "" {
		return &BlobSpec{"", "", false}, nil
	}

	if i := strings.Index(s, "/"); i != -1 {
		z := strings.SplitN(s, "/", 2)
		return &BlobSpec{z[0], z[1], true}, nil
	}

	return &BlobSpec{s, "", false}, nil
}
