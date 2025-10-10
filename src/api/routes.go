package api

import (
	"github.com/gin-gonic/gin"
)

type RouterRegistrar interface {
	RegisterRoutes(router *gin.Engine)
}

type DefaultRouterRegistrar struct {
	handler *Handler
}

func NewRouterRegistrar(handler *Handler) *DefaultRouterRegistrar {
	return &DefaultRouterRegistrar{
		handler: handler,
	}
}

func (r *DefaultRouterRegistrar) RegisterRoutes(router *gin.Engine) {
	router.POST("/check", r.handler.CheckPasswordHandler)

	internalRoutes := router.Group("/internal")
	{
		debugRoutes := internalRoutes.Group("/debug")
		{
			debugRoutes.GET("/byHash/:hash", r.handler.GetByHash)
			debugRoutes.GET("/getAll", r.handler.GetAllHashes)
		}
	}
}
