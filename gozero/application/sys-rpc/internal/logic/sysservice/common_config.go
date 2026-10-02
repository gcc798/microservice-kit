package sysservicelogic

import (
	"context"
	"errors"
	"fmt"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"
	"gorm.io/gorm"
)

type configRow = model.SConfig

func getConfigByID(ctx context.Context, svcCtx *svc.ServiceContext, id int64) (*configRow, error) {
	row, err := gorm.G[model.SConfig](svcCtx.DB).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("配置不存在")
		}
		return nil, err
	}
	return &row, nil
}

func configNameExists(ctx context.Context, svcCtx *svc.ServiceContext, name string, excludeID int64) (bool, error) {
	query := gorm.G[model.SConfig](svcCtx.DB).Where("name = ?", name)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	count, err := query.Count(ctx, "id")
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func toConfigPB(row configRow) *pb.Config {
	return &pb.Config{
		Id:          row.Id,
		Name:        row.Name,
		Code:        row.Code,
		DataJson:    row.Data,
		Remark:      nullString(row.Remark),
		CreateBy:    nullInt64(row.CreateBy),
		CreatedTime: nullTime(row.CreatedTime),
		UpdateBy:    nullInt64(row.UpdateBy),
		UpdatedTime: nullTime(row.UpdatedTime),
	}
}

func toConfigList(rows []configRow) []*pb.Config {
	list := make([]*pb.Config, 0, len(rows))
	for _, row := range rows {
		list = append(list, toConfigPB(row))
	}
	return list
}
