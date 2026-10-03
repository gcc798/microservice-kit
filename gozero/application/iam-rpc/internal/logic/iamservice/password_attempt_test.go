package iamservicelogic

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/redis/go-redis/v9"
)

func TestPasswordAttemptLimit(t *testing.T) {
	server := miniredis.RunT(t)
	svcCtx := &svc.ServiceContext{Redis: redis.NewClient(&redis.Options{Addr: server.Addr()})}
	ctx := context.Background()
	for range passwordErrorLimit {
		incrementPasswordAttempts(ctx, svcCtx, "tester")
	}
	if err := checkPasswordAttempts(ctx, svcCtx, "tester"); err == nil {
		t.Fatal("password attempts were not limited")
	}
	server.FastForward(passwordLockTTL)
	if err := checkPasswordAttempts(ctx, svcCtx, "tester"); err != nil {
		t.Fatalf("password limit did not expire: %v", err)
	}
}

func TestParseUserAgent(t *testing.T) {
	browser, osName := parseUserAgent("Mozilla/5.0 (Mac OS X) AppleWebKit Chrome/125 Safari/537.36")
	if browser != "Chrome" || osName != "macOS" {
		t.Fatalf("parseUserAgent() = %q, %q", browser, osName)
	}
}
