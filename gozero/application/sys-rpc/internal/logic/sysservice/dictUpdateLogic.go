package sysservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DictUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDictUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictUpdateLogic {
	return &DictUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DictUpdateLogic) DictUpdate(in *pb.DictUpdateReq) (*pb.Ack, error) {
	oldRow, err := getDictByID(l.ctx, l.svcCtx, in.Id)
	if err != nil {
		return nil, err
	}
	if in.ParentId == in.Id {
		return nil, errors.New("不能将自己设置为父字典")
	}
	if in.ParentId > 0 {
		parent, err := getDictByID(l.ctx, l.svcCtx, in.ParentId)
		if err != nil {
			return nil, errors.New("父字典不存在")
		}
		dType := in.DictType
		if dType == "" {
			dType = oldRow.DictType.String
		}
		if parent.DictType.Valid && parent.DictType.String != dType {
			return nil, errors.New("父字典类型不匹配")
		}
		isDesc, err := dictDescendantExists(l.ctx, l.svcCtx, in.Id, in.ParentId)
		if err != nil {
			return nil, err
		}
		if isDesc {
			return nil, errors.New("不能将子节点设置为父字典")
		}
	}
	dictType := in.DictType
	if dictType == "" {
		dictType = oldRow.DictType.String
	}
	dictValue := in.DictValue
	if dictValue == "" {
		dictValue = oldRow.DictValue.String
	}
	exists, err := dictValueExists(l.ctx, l.svcCtx, dictType, dictValue, in.Id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("字典值已存在")
	}
	dictLabel := in.DictLabel
	if dictLabel == "" {
		dictLabel = oldRow.DictLabel.String
	}
	updates := map[string]any{
		"parent_id": in.ParentId, "dict_type": dictType, "dict_label": dictLabel, "dict_value": dictValue,
		"sort": in.Sort, "is_default": in.IsDefault, "status": in.Status,
		"remark":    sql.NullString{String: in.Remark, Valid: in.Remark != ""},
		"update_by": sql.NullInt64{Int64: in.UpdateBy, Valid: in.UpdateBy != 0}, "updated_time": time.Now(),
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SDictData{}).Where("id = ?", in.Id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
