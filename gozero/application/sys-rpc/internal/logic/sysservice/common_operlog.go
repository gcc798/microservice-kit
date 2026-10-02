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

type operLogRow = model.SOperLog

func getOperLogByID(ctx context.Context, svcCtx *svc.ServiceContext, id int64) (*operLogRow, error) {
	row, err := gorm.G[model.SOperLog](svcCtx.DB).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("操作日志不存在")
		}
		return nil, err
	}
	return &row, nil
}

func toOperLogPB(row operLogRow) *pb.OperLog {
	return &pb.OperLog{
		Id:            row.Id,
		Title:         nullString(row.Title),
		BusinessType:  nullString(row.BusinessType),
		Method:        nullString(row.Method),
		RequestMethod: nullString(row.RequestMethod),
		DeviceType:    nullString(row.DeviceType),
		OperName:      nullString(row.OperName),
		OperUrl:       nullString(row.OperUrl),
		OperIp:        nullString(row.OperIp),
		OperLocation:  nullString(row.OperLocation),
		OperParam:     nullString(row.OperParam),
		JsonResult:    nullString(row.JsonResult),
		Status:        nullString(row.Status),
		ErrorMsg:      nullString(row.ErrorMsg),
		OperTime:      nullTime(row.OperTime),
		CostTime:      nullInt64(row.CostTime),
		UserAgent:     nullString(row.UserAgent),
	}
}

func toOperLogList(rows []operLogRow) []*pb.OperLog {
	list := make([]*pb.OperLog, 0, len(rows))
	for _, row := range rows {
		list = append(list, toOperLogPB(row))
	}
	return list
}
