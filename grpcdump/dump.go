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

// https://github.com/bradleyjkemp/grpc-tools/blob/master/grpc-dump/README.md
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

func (g *GRPC) Service() string {
	return filepath.Dir(g.FullMethod)
}

func (g *GRPC) Method() string {
	return filepath.Base(g.FullMethod)
}

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
			md, err := ReadMetadata(data)
			if err != nil {
				return nil, err
			}
			g.Metadata = md
		case headerFile:
			md, err := ReadMetadata(data)
			if err != nil {
				return nil, err
			}
			g.Header = md
		case trailerFile:
			md, err := ReadMetadata(data)
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

func Write(g *GRPC, transformers ...Transformer) ([]byte, error) {
	for _, transform := range transformers {
		if err := transform(g); err != nil {
			return nil, err
		}
	}

	arc := new(txtar.Archive)
	arc.Files = append(arc.Files, txtar.File{lineFile, appendNewLine([]byte(WriteLine(g.Addr, g.FullMethod)))})
	arc.Files = append(arc.Files, txtar.File{metadataFile, appendNewLine([]byte(WriteMetadata(g.Metadata)))})

	//i := g.HeaderIdx
	msgs, err := WriteMessages(g.IsClientStream, g.IsServerStream, g.Messages...)
	if err != nil {
		return nil, err
	}
	for i, f := range msgs {
		msgs[i].Data = appendNewLine(f.Data)
	}
	arc.Files = append(arc.Files, msgs...)
	arc.Files = append(arc.Files, txtar.File{headerFile, appendNewLine([]byte(WriteMetadata(g.Header)))})
	//clientMsgs, err := WriteMessages(g.IsClientStream, g.IsServerStream, g.Messages[i:]...)
	//if err != nil {
	//return nil, err
	//}
	//for i, f := range clientMsgs {
	//clientMsgs[i].Data = appendNewLine(f.Data)
	//}

	//arc.Files = append(arc.Files, clientMsgs...)

	status, err := WriteStatus(g.Status)
	if err != nil {
		return nil, err
	}
	if status != nil {
		arc.Files = append(arc.Files, txtar.File{statusFile, appendNewLine(status)})
	}

	arc.Files = append(arc.Files, txtar.File{trailerFile, appendNewLine([]byte(WriteMetadata(g.Trailer)))})
	return bytes.TrimSpace(txtar.Format(arc)), nil
}

func WriteLine(addr, fullMethod string) string {
	return fmt.Sprintf("GRPC %s", filepath.Join(addr, fullMethod))
}

func WriteMetadata(md metadata.MD) string {
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

	return strings.Join(result, "\n")
}

func WriteMessages(isClientStream, isServerStream bool, msgs ...Message) ([]txtar.File, error) {
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
			return nil, UnknownMessageOriginError(msg.Origin)
		}

		// E.,g.
		// server/helloworld.v1.ChatRequest
		// server stream/helloworld.v1.ChatRequest
		header := filepath.Join(prefix, msg.Name)

		files[i] = txtar.File{
			Name: header,
			Data: b,
		}
	}

	return files, nil
}

func WriteStatus(status *Status) ([]byte, error) {
	if status == nil {
		return nil, nil
	}

	return json.MarshalIndent(status, "", " ")
}

func ReadMetadata(b []byte) (metadata.MD, error) {
	data := string(bytes.TrimSpace(b))
	if len(data) == 0 {
		return nil, nil
	}

	kvs := strings.Split(data, "\n")
	m := make(map[string]string)
	for _, kv := range kvs {
		k, v, ok := strings.Cut(kv, ": ")
		if !ok {
			fmt.Println("kv", kv)
			return nil, InvalidMetadataError(kv)
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
