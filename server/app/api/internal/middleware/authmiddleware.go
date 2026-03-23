package middleware

import (
	"net/http"
	"strings"

	"server/app/api/internal/authctx"
	"server/app/api/internal/svc"
	"server/app/rpc/user/userservice"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type AuthMiddleware struct {
	svcCtx *svc.ServiceContext
}

func NewAuthMiddleware(svcCtx *svc.ServiceContext) *AuthMiddleware {
	return &AuthMiddleware{svcCtx: svcCtx}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, map[string]any{
				"code": http.StatusUnauthorized,
				"msg":  "missing or invalid authorization header",
			})
			return
		}

		validateResp, err := m.svcCtx.UserRpc.ValidateToken(r.Context(), &userservice.ValidateTokenReq{
			AccessToken: token,
		})
		if err != nil {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, map[string]any{
				"code": http.StatusUnauthorized,
				"msg":  err.Error(),
			})
			return
		}
		if !validateResp.Valid {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusUnauthorized, map[string]any{
				"code": http.StatusUnauthorized,
				"msg":  "token invalid or expired",
			})
			return
		}

		ctx := authctx.WithUserID(r.Context(), validateResp.UserId)
		ctx = authctx.WithUserStatus(ctx, validateResp.Status)
		ctx = authctx.WithAccessToken(ctx, token)
		next(w, r.WithContext(ctx))
	}
}

func bearerToken(header string) (string, bool) {
	if header == "" {
		return "", false
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", false
	}

	return strings.TrimSpace(parts[1]), true
}
