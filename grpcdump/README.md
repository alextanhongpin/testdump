# grpcdump

[![Go Reference](https://pkg.go.dev/badge/github.com/alextanhongpin/testdump/grpcdump.svg)](https://pkg.go.dev/github.com/alextanhongpin/testdump/grpcdump)

## Purpose

`grpcdump` is a specialized testing utility for Go that simplifies gRPC service testing by providing tools to capture, serialize, and verify Protocol Buffer messages. It enables snapshot testing for gRPC communications, making it easier to detect regressions and validate message structures in your microservices architecture.

## Example Usage

```go
package main

import (
    "testing"
    "context"
    
    "github.com/alextanhongpin/testdump/grpcdump"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "your/protobuf/package"
)

func TestUserService(t *testing.T) {
    // Setup a test gRPC server with grpcdump interceptors
    server := grpcdump.NewServer(
        grpcdump.StreamInterceptor(),
        grpcdump.UnaryInterceptor(),
    )
    
    // Register your service
    pb.RegisterUserServiceServer(server.Server, &yourServiceImplementation{})
    stop := server.ListenAndServe()
    defer stop()
    
    // Setup a test gRPC client with grpcdump
    ctx := context.Background()
    conn, err := server.DialContext(ctx,
        grpcdump.WithUnaryInterceptor(),
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        t.Fatal(err)
    }
    defer conn.Close()
    
    client := pb.NewUserServiceClient(conn)
    
    // Create a recorder with options
    ctx = grpcdump.NewRecorder(t, ctx, 
        grpcdump.IgnoreMetadata("user-agent"),
        grpcdump.MaskMessageFields("[MASKED]", []string{"password"}),
    )
    
    // Create a request
    req := &pb.GetUserRequest{
        UserId: "123",
    }
    
    // Make the gRPC call - everything is recorded automatically
    resp, err := client.GetUser(ctx, req)
    if err != nil {
        t.Fatalf("GetUser failed: %v", err)
    }
    
    // No need to explicitly call Dump() - the recorder captures everything
    // Assertions can be made as normal
    if resp.GetUsername() != "expected_username" {
        t.Errorf("Expected username %q, got %q", "expected_username", resp.GetUsername())
    }
}
```

## Benefits

- **Simplified gRPC Testing**: Makes it easy to verify that your gRPC services are sending and receiving the correct Protocol Buffer messages.
- **Integration with Snapshot Testing**: Create and compare snapshots of your gRPC messages to catch unintended changes.
- **Compatible with protobuf**: Works seamlessly with Protocol Buffer messages, including complex nested structures.
- **Flexible Options**:
  - Ignore dynamic fields like timestamps and IDs
  - Mask sensitive information in the snapshots
  - Customize output with transformers
  - Target specific paths within complex protobuf messages
- **Developer-Friendly**: Minimal setup required to start testing gRPC services effectively.
- **Works with Existing Tests**: Easily integrate into your existing test suites without changing your testing approach.

## Advanced Features

- Apply transformations to Protocol Buffer messages before dumping
- Register type-specific options for consistent handling across tests
- Compare different versions of the same message type
- Use with mock gRPC servers for complete service testing
- Validate message structure against expected schemas
