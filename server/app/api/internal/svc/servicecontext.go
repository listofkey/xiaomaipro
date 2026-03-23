package svc

import (
	"server/app/api/internal/config"
	"server/app/rpc/order/orderservice"
	"server/app/rpc/payment/paymentservice"
	"server/app/rpc/program/programservice"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	UserRpc    userservice.UserService
	ProgramRpc programservice.ProgramService
	OrderRpc   orderservice.OrderService
	PaymentRpc paymentservice.PaymentService
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		UserRpc:    userservice.NewUserService(zrpc.MustNewClient(c.UserRpc)),
		ProgramRpc: programservice.NewProgramService(zrpc.MustNewClient(c.ProgramRpc)),
		OrderRpc:   orderservice.NewOrderService(zrpc.MustNewClient(c.OrderRpc)),
		PaymentRpc: paymentservice.NewPaymentService(zrpc.MustNewClient(c.PaymentRpc)),
	}
}
