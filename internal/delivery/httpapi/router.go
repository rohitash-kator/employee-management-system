package httpapi

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/rohitashk/golang-rest-api/internal/delivery/httpapi/handlers"
	"github.com/rohitashk/golang-rest-api/internal/delivery/httpapi/middleware"
	employeeUC "github.com/rohitashk/golang-rest-api/internal/usecase/employee"
)

type RouterDeps struct {
	Logger         *slog.Logger
	RequestTimeout time.Duration
	EmployeeSvc    *employeeUC.Service
}

func NewRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(deps.Logger))

	health := handlers.NewHealthHandler()
	r.GET("/healthz", health.Get)

	v1 := r.Group("/v1")
	{
		eh := handlers.NewEmployeeHandler(deps.EmployeeSvc, deps.RequestTimeout)
		v1.POST("/employees", eh.Create)
		v1.GET("/employees", eh.List)
		v1.GET("/employees/:id", eh.Get)
		v1.PATCH("/employees/:id", eh.Update)
		v1.DELETE("/employees/:id", eh.Delete)
	}

	return r
}
