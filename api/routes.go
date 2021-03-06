package api

import (
	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/container"
	"github.com/mylxsw/glacier/web"
	"github.com/mylxsw/graphviz-server/api/controller"
)

func routers(cc container.Container) func(router *web.Router, mw web.RequestMiddleware) {
	return func(router *web.Router, mw web.RequestMiddleware) {
		mws := make([]web.HandlerDecorator, 0)
		mws = append(mws, mw.AccessLog(log.Module("api")), mw.CORS("*"))

		router.WithMiddleware(mws...).Controllers(
			"/api",
			controller.NewWelcomeController(cc),
			controller.NewGraphvizController(cc),
		)
	}
}
