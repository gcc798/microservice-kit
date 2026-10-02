package iamservicelogic

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/iam-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/iam-rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type OrgUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOrgUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrgUpdateLogic {
	return &OrgUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *OrgUpdateLogic) OrgUpdate(in *pb.OrgUpdateReq) (*pb.Ack, error) {
	row, err := getOrgByID(l.ctx, l.svcCtx, in.Id)
	if err != nil {
		return nil, err
	}
	if in.ParentId == in.Id && in.Id != 0 {
		return nil, errors.New("不能将组织设置为自己的子组织")
	}
	if in.OrgCode != "" && in.OrgCode != nullString(row.OrgCode) {
		count, err := gorm.G[model.SOrg](l.svcCtx.DB).Where("org_code = ? AND id <> ?", in.OrgCode, in.Id).Count(l.ctx, "id")
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("组织编码已被占用")
		}
	}
	parentID := row.ParentId
	ancestors := nullString(row.Ancestors)
	if in.ParentId != row.ParentId {
		nextAncestors, err := buildOrgAncestors(l.ctx, l.svcCtx, in.ParentId)
		if err != nil {
			return nil, err
		}
		if strings.Contains(","+nextAncestors+",", ","+strconv.FormatInt(in.Id, 10)+",") {
			return nil, errors.New("不能将组织移动到其子组织下")
		}
		parentID = in.ParentId
		ancestors = nextAncestors
	}
	orgName := row.OrgName
	if in.OrgName != "" {
		orgName = in.OrgName
	}
	orgCode := row.OrgCode
	if in.OrgCode != "" {
		orgCode = sql.NullString{String: in.OrgCode, Valid: true}
	}
	orgType := row.OrgType
	if in.OrgType != "" {
		orgType = in.OrgType
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.SOrg{}).Where("id = ?", in.Id).Updates(map[string]any{
		"parent_id": parentID, "ancestors": ancestors, "org_name": orgName, "org_code": orgCode,
		"org_type": orgType, "leader": nullableString(in.Leader), "phone": nullableString(in.Phone),
		"email": nullableString(in.Email), "status": in.Status, "sort": in.Sort,
		"remark": nullableString(in.Remark), "updated_time": time.Now(),
	}).Error; err != nil {
		return nil, err
	}
	return &pb.Ack{Msg: "ok"}, nil
}
