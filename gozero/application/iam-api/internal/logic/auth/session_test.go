package auth

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/config"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/svc"
	commonauth "github.com/gcc798/microservice-kit/common/auth"
	"github.com/redis/go-redis/v9"
)

func TestSingleSessionRevokesPreviousToken(t *testing.T) {
	server := miniredis.RunT(t)
	ctx := context.Background()
	svcCtx := &svc.ServiceContext{Config: config.Config{Auth: config.AuthConf{AllowConcurrent: false}}, Redis: redis.NewClient(&redis.Options{Addr: server.Addr()})}
	user := &loginUser{Id: 42, UserName: "tester"}
	client := &authClient{ClientId: "web", ClientKey: "web-admin", Timeout: 3600, ActiveTimeout: 300}

	if err := storeTokenSession(ctx, svcCtx, user, client, "access-1", "refresh-1", true); err != nil {
		t.Fatal(err)
	}
	if err := storeTokenSession(ctx, svcCtx, user, client, "access-2", "refresh-2", true); err != nil {
		t.Fatal(err)
	}
	if server.Exists(commonauth.AccessTokenKey("access-1")) || server.Exists(commonauth.RefreshTokenKey("refresh-1")) {
		t.Fatal("previous session remained active")
	}
	if !server.Exists(commonauth.AccessTokenKey("access-2")) || !server.Exists(commonauth.RefreshTokenKey("refresh-2")) {
		t.Fatal("new session was not stored")
	}

	invalidateByToken(ctx, svcCtx, "access-2")
	if server.Exists(commonauth.AccessTokenKey("access-2")) || server.Exists(commonauth.RefreshTokenKey("refresh-2")) {
		t.Fatal("logout did not revoke the session")
	}
}
