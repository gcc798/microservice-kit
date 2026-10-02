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

type loginLogRow = model.SLoginLog

func getLoginLogByID(ctx context.Context, svcCtx *svc.ServiceContext, id int64) (*loginLogRow, error) {
	row, err := gorm.G[model.SLoginLog](svcCtx.DB).Where("id = ?", id).First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("登录日志不存在")
		}
		return nil, err
	}
	return &row, nil
}

func toLoginLogPB(row loginLogRow) *pb.LoginLog {
	return &pb.LoginLog{
		Id:            row.Id,
		UserName:      nullString(row.UserName),
		Ipaddr:        nullString(row.Ipaddr),
		LoginLocation: nullString(row.LoginLocation),
		Browser:       nullString(row.Browser),
		Os:            nullString(row.Os),
		Status:        int32(row.Status),
		Msg:           nullString(row.Msg),
		LoginTime:     nullTime(row.LoginTime),
		ClientId:      nullString(row.ClientId),
	}
}

func toLoginLogList(rows []loginLogRow) []*pb.LoginLog {
	list := make([]*pb.LoginLog, 0, len(rows))
	for _, row := range rows {
		list = append(list, toLoginLogPB(row))
	}
	return list
}
