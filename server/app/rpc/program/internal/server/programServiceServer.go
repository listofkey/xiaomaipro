package server

import (
	"context"

	"server/app/rpc/program/internal/logic"
	"server/app/rpc/program/internal/svc"
	"server/app/rpc/program/programpb/programpb"
)

type ProgramServiceServer struct {
	svcCtx *svc.ServiceContext
	programpb.UnimplementedProgramServiceServer
}

func NewProgramServiceServer(svcCtx *svc.ServiceContext) *ProgramServiceServer {
	return &ProgramServiceServer{
		svcCtx: svcCtx,
	}
}

// --- 活动相关 ---

func (s *ProgramServiceServer) ListEvent(ctx context.Context, in *programpb.ListEventReq) (*programpb.ListEventResp, error) {
	l := logic.NewListEventLogic(ctx, s.svcCtx)
	return l.ListEvent(in)
}

func (s *ProgramServiceServer) GetEventDetail(ctx context.Context, in *programpb.GetEventDetailReq) (*programpb.GetEventDetailResp, error) {
	l := logic.NewGetEventDetailLogic(ctx, s.svcCtx)
	return l.GetEventDetail(in)
}

func (s *ProgramServiceServer) SearchEvent(ctx context.Context, in *programpb.SearchEventReq) (*programpb.SearchEventResp, error) {
	l := logic.NewSearchEventLogic(ctx, s.svcCtx)
	return l.SearchEvent(in)
}

// --- 分类相关 ---

func (s *ProgramServiceServer) ListCategory(ctx context.Context, in *programpb.ListCategoryReq) (*programpb.ListCategoryResp, error) {
	l := logic.NewListCategoryLogic(ctx, s.svcCtx)
	return l.ListCategory(in)
}

// --- 推荐8个最新的活动

func (s *ProgramServiceServer) GetHotRecommend(ctx context.Context, in *programpb.GetHotRecommendReq) (*programpb.GetHotRecommendResp, error) {
	l := logic.NewGetHotRecommendLogic(ctx, s.svcCtx)
	return l.GetHotRecommend(in)
}
