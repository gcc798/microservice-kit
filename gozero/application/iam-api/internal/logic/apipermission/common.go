package apipermission

import (
	"context"

	"github.com/gcc798/microservice-kit/application/iam-api/internal/logic/commonutil"
	"github.com/gcc798/microservice-kit/application/iam-api/internal/types"
	"github.com/gcc798/microservice-kit/application/iam-rpc/client/iamservice"
)

func success(data interface{}) *types.CommonResp {
	return &types.CommonResp{Code: 200, Msg: "操作成功", Data: data}
}

func failure(err error) *types.CommonResp {
	return &types.CommonResp{Code: 500, Msg: err.Error()}
}

func saveReq(ctx context.Context, req *types.ApiPermissionSaveReq) *iamservice.ApiPermissionSaveReq {
	return &iamservice.ApiPermissionSaveReq{
		ParentId: req.ParentId,
		Module:   req.Module,
		Code:     req.Code,
		Name:     req.Name,
		NodeType: int32(req.NodeType),
		Action:   req.Action,
		Method:   req.Method,
		Path:     req.Path,
		Sort:     req.Sort,
		Status:   int32(req.Status),
		Remark:   req.Remark,
		UserId:   commonutil.UserIDFromContext(ctx),
	}
}

func updateReq(ctx context.Context, req *types.ApiPermissionUpdateReq) *iamservice.ApiPermissionSaveReq {
	return &iamservice.ApiPermissionSaveReq{
		Id:       req.Id,
		ParentId: req.ParentId,
		Module:   req.Module,
		Code:     req.Code,
		Name:     req.Name,
		NodeType: int32(req.NodeType),
		Action:   req.Action,
		Method:   req.Method,
		Path:     req.Path,
		Sort:     req.Sort,
		Status:   int32(req.Status),
		Remark:   req.Remark,
		UserId:   commonutil.UserIDFromContext(ctx),
	}
}
