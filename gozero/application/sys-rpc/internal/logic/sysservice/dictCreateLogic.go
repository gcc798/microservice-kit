package sysservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

type DictCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDictCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DictCreateLogic {
	return &DictCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DictCreateLogic) DictCreate(in *pb.DictCreateReq) (*pb.Ack, error) {
	if in.DictType == "" || in.DictLabel == "" || in.DictValue == "" {
		return nil, errors.New("字典类型、标签和值不能为空")
	}
	exists, err := dictValueExists(l.ctx, l.svcCtx, in.DictType, in.DictValue, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("字典值已存在")
	}
	if in.ParentId > 0 {
		parent, err := getDictByID(l.ctx, l.svcCtx, in.ParentId)
		if err != nil {
			return nil, errors.New("父字典不存在")
		}
		if parent.DictType.Valid && parent.DictType.String != in.DictType {
			return nil, errors.New("父字典类型不匹配")
		}
	}
	now := time.Now()
	row := model.SDictData{
		ParentId: in.ParentId, DictType: sql.NullString{String: in.DictType, Valid: true},
		DictLabel: sql.NullString{String: in.DictLabel, Valid: true}, DictValue: sql.NullString{String: in.DictValue, Valid: true},
		Sort: in.Sort, IsDefault: in.IsDefault, Status: int64(in.Status), Remark: sql.NullString{String: in.Remark, Valid: in.Remark != ""},
		CreateBy: sql.NullInt64{Int64: in.CreateBy, Valid: in.CreateBy != 0}, UpdateBy: sql.NullInt64{Int64: in.UpdateBy, Valid: in.UpdateBy != 0},
		CreatedTime: sql.NullTime{Time: now, Valid: true}, UpdatedTime: sql.NullTime{Time: now, Valid: true},
	}
	if err := gorm.G[model.SDictData](l.svcCtx.DB).Create(l.ctx, &row); err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
