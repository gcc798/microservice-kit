package sys

import (
	"context"
	"time"

	"github.com/gcc798/microservice-kit/application/sys/internal/domain/model"
	sysv1 "github.com/gcc798/microservice-kit/internal/api/sys/v1"
	logging "github.com/gcc798/microservice-kit/internal/logger"
	"github.com/gcc798/microservice-kit/internal/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type LogCleanupService interface {
	Clean(ctx context.Context, retentionDays int) (loginLogs, operationLogs int64, err error)
}

type logCleanupService struct {
	login     LoginLogService
	operation OperLogService
}

func NewLogCleanupService(login LoginLogService, operation OperLogService) LogCleanupService {
	return &logCleanupService{login: login, operation: operation}
}

func (s *logCleanupService) Clean(ctx context.Context, retentionDays int) (int64, int64, error) {
	loginLogs, err := s.login.CleanOldLogs(ctx, retentionDays)
	if err != nil {
		return 0, 0, err
	}
	operationLogs, err := s.operation.CleanOldLogs(ctx, retentionDays)
	if err != nil {
		return loginLogs, 0, err
	}
	return loginLogs, operationLogs, nil
}

type API struct {
	db     *gorm.DB
	logger logging.Logger
}

func NewAPI(db *gorm.DB, logger logging.Logger) *API { return &API{db: db, logger: logger} }

func (a *API) RecordLogin(ctx context.Context, request *sysv1.RecordLoginRequest) error {
	entry := &model.LoginLog{
		ID: request.Id, UserName: request.UserName, Ipaddr: request.IpAddress,
		Browser: request.Browser, Os: request.Os, Status: request.Status, Msg: request.Message,
		LoginTime: utils.LocalTime(time.UnixMilli(request.LoginTimeUnixMilli)), ClientId: request.ClientId,
	}
	return a.db.WithContext(ctx).Create(entry).Error
}

func (a *API) RecordOperations(ctx context.Context, request *sysv1.RecordOperationsRequest) error {
	if len(request.Logs) == 0 {
		return nil
	}
	entries := make([]model.OperLog, 0, len(request.Logs))
	for _, item := range request.Logs {
		entries = append(entries, model.OperLog{
			ID: item.Id, Title: item.Title, BusinessType: item.BusinessType, Method: item.Method,
			RequestMethod: item.RequestMethod, DeviceType: item.DeviceType, OperName: item.OperatorName,
			OperUrl: item.Url, OperIp: item.IpAddress, OperParam: item.Parameters, Status: item.Status,
			ErrorMsg: item.ErrorMessage, OperTime: utils.LocalTime(time.UnixMilli(item.OperationTimeUnixMilli)),
			CostTime: item.CostMillis, UserAgent: item.UserAgent,
		})
	}
	return a.db.WithContext(ctx).Create(&entries).Error
}

type GRPCServer struct {
	sysv1.UnimplementedSystemServiceServer
	api sysv1.API
}

func NewGRPCServer(api sysv1.API) *GRPCServer { return &GRPCServer{api: api} }

func (s *GRPCServer) RecordLogin(ctx context.Context, request *sysv1.RecordLoginRequest) (*sysv1.Empty, error) {
	if err := s.api.RecordLogin(ctx, request); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &sysv1.Empty{}, nil
}

func (s *GRPCServer) RecordOperations(ctx context.Context, request *sysv1.RecordOperationsRequest) (*sysv1.Empty, error) {
	if err := s.api.RecordOperations(ctx, request); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &sysv1.Empty{}, nil
}
