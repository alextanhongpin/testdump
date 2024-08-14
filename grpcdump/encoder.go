package grpcdump

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rogpeppe/go-internal/txtar"
	"google.golang.org/grpc/metadata"
)

var (
	ErrInvalidMetadata = fmt.Errorf("grpcdump: invalid metadata")
)

type encoder struct {
	marshalFns []func(*GRPC) error
}

func (e *encoder) Marshal(v any) ([]byte, error) {
	g := v.(*GRPC)

	for _, fn := range e.marshalFns {
		if err := fn(g); err != nil {
			return nil, err
		}
	}

	return Write(g)
}

func (e *encoder) Unmarshal(b []byte) (a any, err error) {
	return Read(b)
}

const (
	lineFile         = "line"
	headerFile       = "header"
	metadataFile     = "metadata"
	clientFile       = "client"
	clientStreamFile = "client stream"
	serverFile       = "server"
	serverStreamFile = "server stream"
	statusFile       = "status"
	trailerFile      = "trailer"
)

// GRPC is a struct that represents a gRPC message.
// It contains all the necessary fields for a complete gRPC message.
type GRPC struct {
	Addr           string      `json:"addr"`
	FullMethod     string      `json:"full_method"`
	Messages       []Message   `json:"messages"`
	Status         *Status     `json:"status"`
	Metadata       metadata.MD `json:"metadata"` // The server receives metadata.
	Header         metadata.MD `json:"header"`   // The client receives header and trailer.
	Trailer        metadata.MD `json:"trailer"`
	IsServerStream bool        `json:"isServerStream"`
	IsClientStream bool        `json:"isClientStream"`
	//HeaderIdx      int         `json:"-"`
}

// Service is a method on the GRPC struct.
// It extracts and returns the service name from the FullMethod field.
func (g *GRPC) Service() string {
	return filepath.Dir(g.FullMethod)
}

// Method is a method on the GRPC struct.
// It extracts and returns the method name from the FullMethod field.
func (g *GRPC) Method() string {
	return filepath.Base(g.FullMethod)
}

// Read is a function that takes a byte slice and returns a GRPC object and an error.
// It attempts to unmarshal the byte slice into a GRPC object.
// If the unmarshalling is successful, it returns the GRPC object and nil.
// If the unmarshalling fails, it returns nil and the error.
func Read(b []byte) (*GRPC, error) {
	arc := txtar.Parse(b)

	g := new(GRPC)
	for _, f := range arc.Files {
		name, typ, _ := strings.Cut(f.Name, "/")
		data := bytes.TrimSpace(f.Data)
		switch name {
		case lineFile:
			text := string(data)
			text = strings.TrimPrefix(text, "GRPC ")
			i := strings.Index(text, "/")
			addr, fullMethod := text[:i], text[i:]
			g.Addr = addr
			g.FullMethod = fullMethod
		case metadataFile:
			md, err := readMetadata(data)
			if err != nil {
				return nil, err
			}
			g.Metadata = md
		case headerFile:
			md, err := readMetadata(data)
			if err != nil {
				return nil, err
			}
			g.Header = md
		case trailerFile:
			md, err := readMetadata(data)
			if err != nil {
				return nil, err
			}
			g.Trailer = md
		case statusFile:
			if err := json.Unmarshal(data, &g.Status); err != nil {
				return nil, err
			}
		case
			serverFile,
			serverStreamFile,
			clientFile,
			clientStreamFile:
			origin := name
			if origin == clientStreamFile {
				g.IsClientStream = true
			} else if origin == serverStreamFile {
				g.IsServerStream = true
			}

			var a any
			if err := json.Unmarshal(data, &a); err != nil {
				return nil, err
			}

			// We don't need to know if it is stream or not.
			origin = strings.TrimSuffix(origin, " stream")

			g.Messages = append(g.Messages, Message{
				Origin:  origin,
				Message: a,
				Name:    typ,
			})
		}
	}

	return g, nil
}

// Write is a function that takes a pointer to a GRPC object and returns a byte
// slice and an error.
func Write(g *GRPC) ([]byte, error) {
	var files []txtar.File
	files = append(files, writeLine(g.Addr, g.FullMethod))
	files = append(files, writeMetadata(metadataFile, g.Metadata))

	msgs, err := writeMessages(g.IsClientStream, g.IsServerStream, g.Messages...)
	if err != nil {
		return nil, err
	}
	files = append(files, msgs...)
	files = append(files, writeMetadata(headerFile, g.Header))

	status, err := writeStatus(g.Status)
	if err != nil {
		return nil, err
	}
	files = append(files, status)
	files = append(files, writeMetadata(trailerFile, g.Trailer))

	arc := new(txtar.Archive)
	for _, f := range files {
		if len(bytes.TrimSpace(f.Data)) == 0 {
			continue
		}
		arc.Files = append(arc.Files, f)
	}

	return bytes.TrimSpace(txtar.Format(arc)), nil
}

func writeLine(addr, fullMethod string) txtar.File {
	return txtar.File{
		Name: lineFile,
		Data: appendNewLine([]byte(fmt.Sprintf("GRPC %s", filepath.Join(addr, fullMethod)))),
	}
}

func writeMetadata(file string, md metadata.MD) txtar.File {
	var keys []string
	for k := range md {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result []string
	for _, k := range keys {
		vs := md[k]
		fn := func(s string) string {
			if isBinHeader(k) {
				// https://github.com/grpc/grpc-go/pull/1209/files
				return encodeBinHeader([]byte(s))
			}

			return s
		}

		for _, v := range vs {
			result = append(result, fmt.Sprintf("%s: %s", k, fn(v)))
		}
	}

	return txtar.File{
		Name: file,
		Data: appendNewLine([]byte(strings.Join(result, "\n"))),
	}
}

func writeMessages(isClientStream, isServerStream bool, msgs ...Message) ([]txtar.File, error) {
	files := make([]txtar.File, len(msgs))
	for i, msg := range msgs {
		// Pretty print the message.
		b, err := json.MarshalIndent(msg.Message, "", " ")
		if err != nil {
			return nil, err
		}

		var prefix string
		isServer := msg.Origin == OriginServer
		if isServer && isServerStream {
			prefix = serverStreamFile
		} else if isServer && !isServerStream {
			prefix = serverFile
		} else if !isServer && isClientStream {
			prefix = clientStreamFile
		} else if !isServer && !isClientStream {
			prefix = clientFile
		} else {
			return nil, fmt.Errorf("grpcdump: unknown message origin")
		}

		// E.,g.
		// server/helloworld.v1.ChatRequest
		// server stream/helloworld.v1.ChatRequest
		header := filepath.Join(prefix, msg.Name)

		files[i] = txtar.File{
			Name: header,
			Data: appendNewLine(b),
		}
	}

	return files, nil
}

func writeStatus(status *Status) (txtar.File, error) {
	if status == nil {
		return txtar.File{}, nil
	}

	b, err := json.MarshalIndent(status, "", " ")
	if err != nil {
		return txtar.File{}, err
	}

	return txtar.File{
		Name: statusFile,
		Data: appendNewLine(b),
	}, nil
}

func readMetadata(b []byte) (metadata.MD, error) {
	data := string(bytes.TrimSpace(b))
	if len(data) == 0 {
		return nil, nil
	}

	kvs := strings.Split(data, "\n")
	m := make(map[string]string)
	for _, kv := range kvs {
		k, v, ok := strings.Cut(kv, ": ")
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrInvalidMetadata, kv)
		}

		if isBinHeader(k) {
			b, err := decodeBinHeader(v)
			if err != nil {
				return nil, err
			}
			v = string(b)
		}

		m[k] = v
	}

	return metadata.New(m), nil
}

func appendNewLine(b []byte) []byte {
	b = append(b, '\n', '\n')
	return b
}

func encodeBinHeader(v []byte) string {
	return base64.RawStdEncoding.EncodeToString(v)
}

func decodeBinHeader(v string) ([]byte, error) {
	if len(v)%4 == 0 {
		// Input was padded, or padding was not necessary.
		return base64.StdEncoding.DecodeString(v)
	}

	return base64.RawStdEncoding.DecodeString(v)
}

func isBinHeader(key string) bool {
	return strings.HasSuffix(key, "-bin")
}
