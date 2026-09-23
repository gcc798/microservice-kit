package transport

import (
	"context"
	"testing"

	"github.com/gcc798/microservice-kit/internal/registry"
	"google.golang.org/grpc"
)

func TestRegisterService(t *testing.T) {
	reg := registry.NewInProcess()
	defer reg.Close()

	instance, err := RegisterService(context.Background(), reg, "example", "example-test", nil)
	if err != nil {
		t.Fatal(err)
	}
	if instance.ID != "example-test" {
		t.Fatalf("instance ID = %q", instance.ID)
	}
	instances, err := reg.Resolve(context.Background(), "example")
	if err != nil || len(instances) != 1 || instances[0].ID != instance.ID {
		t.Fatalf("instances = %v, err = %v", instances, err)
	}
	if _, err := RegisterService(context.Background(), reg, "example", "", nil); err == nil {
		t.Fatal("RegisterService accepted an empty instance ID")
	}
}

func TestStartRegisteredGRPCRequiresServiceID(t *testing.T) {
	reg := registry.NewInProcess()
	defer reg.Close()
	if _, err := StartRegisteredGRPC(context.Background(), reg, RegisteredGRPCOptions{}, func(*grpc.Server) {}); err == nil {
		t.Fatal("StartRegisteredGRPC accepted an empty service ID")
	}
}
