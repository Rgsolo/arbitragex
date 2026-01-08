// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package handler

import (
	"net/http"

	"arbitragex/restful/price/internal/logic"
	"arbitragex/restful/price/internal/svc"
	"arbitragex/restful/price/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 健康检查
func healthCheckHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.HealthCheckResponse
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewHealthCheckLogic(r.Context(), svcCtx)
		err := l.HealthCheck(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.Ok(w)
		}
	}
}
