package sysservicelogic

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/model"
	"github.com/gcc798/microservice-kit/application/sys-rpc/internal/svc"
	"github.com/gcc798/microservice-kit/application/sys-rpc/pb"
	"gorm.io/gorm"
)

type dictRow = model.SDictData

func parseDictID(id string) (int64, error) {
	return strconv.ParseInt(id, 10, 64)
}

func getDictByID(ctx context.Context, svcCtx *svc.ServiceContext, id int64) (*dictRow, error) {
	row, err := gorm.G[model.SDictData](svcCtx.DB).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("字典不存在")
		}
		return nil, err
	}
	return &row, nil
}

func dictValueExists(ctx context.Context, svcCtx *svc.ServiceContext, dictType, dictValue string, excludeID int64) (bool, error) {
	query := gorm.G[model.SDictData](svcCtx.DB).Where("dict_type = ? AND dict_value = ?", dictType, dictValue)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	count, err := query.Count(ctx, "id")
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func dictDescendantExists(ctx context.Context, svcCtx *svc.ServiceContext, id, parentID int64) (bool, error) {
	var count int64
	err := svcCtx.DB.WithContext(ctx).Raw(`
		with recursive dict_tree as (
			select id, parent_id from public.s_dict_data where id = ?
			union all
			select d.id, d.parent_id from public.s_dict_data d
			inner join dict_tree dt on d.parent_id = dt.id
		)
		select count(1) from dict_tree where id = ?
	`, id, parentID).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func toDictPB(row dictRow) *pb.Dict {
	return &pb.Dict{
		Id:          row.Id,
		ParentId:    row.ParentId,
		DictType:    nullString(row.DictType),
		DictLabel:   nullString(row.DictLabel),
		DictValue:   nullString(row.DictValue),
		Sort:        row.Sort,
		IsDefault:   row.IsDefault,
		Status:      int32(row.Status),
		Remark:      nullString(row.Remark),
		CreateBy:    nullInt64(row.CreateBy),
		UpdateBy:    nullInt64(row.UpdateBy),
		CreatedTime: nullTime(row.CreatedTime),
		UpdatedTime: nullTime(row.UpdatedTime),
	}
}

func toDictList(rows []dictRow) []*pb.Dict {
	list := make([]*pb.Dict, 0, len(rows))
	for _, row := range rows {
		list = append(list, toDictPB(row))
	}
	return list
}
