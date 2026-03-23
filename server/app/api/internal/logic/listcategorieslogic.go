package logic

import (
	"context"

	"server/app/api/internal/svc"
	"server/app/api/internal/types"
	"server/app/rpc/program/programservice"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCategoriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListCategoriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCategoriesLogic {
	return &ListCategoriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListCategoriesLogic) ListCategories(req *types.ListCategoriesReq) (*types.ProgramCategoryListResp, error) {
	status := req.Status
	if status == 0 {
		status = 1
	}

	resp, err := l.svcCtx.ProgramRpc.ListCategory(l.ctx, &programservice.ListCategoryReq{
		Status: status,
	})
	if err != nil {
		return nil, err
	}

	return &types.ProgramCategoryListResp{
		Categories: mapProgramCategoryList(resp.Categories),
		Cities:     mapProgramCityList(resp.Cities),
	}, nil
}
