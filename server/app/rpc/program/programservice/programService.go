package programservice

import (
	"context"

	"server/app/rpc/program/programpb/programpb"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	CategoryInfo        = programpb.CategoryInfo
	CityInfo            = programpb.CityInfo
	EventBrief          = programpb.EventBrief
	EventDetail         = programpb.EventDetail
	GetEventDetailReq   = programpb.GetEventDetailReq
	GetEventDetailResp  = programpb.GetEventDetailResp
	GetHotRecommendReq  = programpb.GetHotRecommendReq
	GetHotRecommendResp = programpb.GetHotRecommendResp
	ListCategoryReq     = programpb.ListCategoryReq
	ListCategoryResp    = programpb.ListCategoryResp
	ListEventReq        = programpb.ListEventReq
	ListEventResp       = programpb.ListEventResp
	SearchEventReq      = programpb.SearchEventReq
	SearchEventResp     = programpb.SearchEventResp
	TicketTierInfo      = programpb.TicketTierInfo
	VenueInfo           = programpb.VenueInfo

	ProgramService interface {
		ListEvent(ctx context.Context, in *ListEventReq, opts ...grpc.CallOption) (*ListEventResp, error)
		GetEventDetail(ctx context.Context, in *GetEventDetailReq, opts ...grpc.CallOption) (*GetEventDetailResp, error)
		SearchEvent(ctx context.Context, in *SearchEventReq, opts ...grpc.CallOption) (*SearchEventResp, error)
		ListCategory(ctx context.Context, in *ListCategoryReq, opts ...grpc.CallOption) (*ListCategoryResp, error)
		GetHotRecommend(ctx context.Context, in *GetHotRecommendReq, opts ...grpc.CallOption) (*GetHotRecommendResp, error)
	}

	defaultProgramService struct {
		cli zrpc.Client
	}
)

func NewProgramService(cli zrpc.Client) ProgramService {
	return &defaultProgramService{cli: cli}
}

func (m *defaultProgramService) ListEvent(ctx context.Context, in *ListEventReq, opts ...grpc.CallOption) (*ListEventResp, error) {
	client := programpb.NewProgramServiceClient(m.cli.Conn())
	return client.ListEvent(ctx, in, opts...)
}

func (m *defaultProgramService) GetEventDetail(ctx context.Context, in *GetEventDetailReq, opts ...grpc.CallOption) (*GetEventDetailResp, error) {
	client := programpb.NewProgramServiceClient(m.cli.Conn())
	return client.GetEventDetail(ctx, in, opts...)
}

func (m *defaultProgramService) SearchEvent(ctx context.Context, in *SearchEventReq, opts ...grpc.CallOption) (*SearchEventResp, error) {
	client := programpb.NewProgramServiceClient(m.cli.Conn())
	return client.SearchEvent(ctx, in, opts...)
}

func (m *defaultProgramService) ListCategory(ctx context.Context, in *ListCategoryReq, opts ...grpc.CallOption) (*ListCategoryResp, error) {
	client := programpb.NewProgramServiceClient(m.cli.Conn())
	return client.ListCategory(ctx, in, opts...)
}

func (m *defaultProgramService) GetHotRecommend(ctx context.Context, in *GetHotRecommendReq, opts ...grpc.CallOption) (*GetHotRecommendResp, error) {
	client := programpb.NewProgramServiceClient(m.cli.Conn())
	return client.GetHotRecommend(ctx, in, opts...)
}
