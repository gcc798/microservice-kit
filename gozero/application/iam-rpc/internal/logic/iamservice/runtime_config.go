package iamservicelogic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/client/sysservice"
)

func loadRuntimeConfig(ctx context.Context, svcCtx *svc.ServiceContext, code string, target any) error {
	result, err := svcCtx.SysRpcClient.ConfigData(ctx, &sysservice.ConfigCodeQueryReq{Code: code})
	if err != nil {
		return fmt.Errorf("读取运行时配置 %s 失败: %w", code, err)
	}
	if err := json.Unmarshal([]byte(result.DataJson), target); err != nil {
		return fmt.Errorf("解析运行时配置 %s 失败: %w", code, err)
	}
	return nil
}
