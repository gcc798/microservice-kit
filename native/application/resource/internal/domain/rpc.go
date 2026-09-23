package resource

import resourcev1 "github.com/gcc798/microservice-kit/internal/api/resource/v1"

type API struct{}

func NewAPI() *API { return &API{} }

type GRPCServer struct {
	resourcev1.UnimplementedResourceServiceServer
}

func NewGRPCServer() *GRPCServer { return &GRPCServer{} }
