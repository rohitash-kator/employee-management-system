package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/rohitashk/golang-rest-api/internal/delivery/httpapi/response"
	domainEmployee "github.com/rohitashk/golang-rest-api/internal/domain/employee"
	employeeUC "github.com/rohitashk/golang-rest-api/internal/usecase/employee"
)

type EmployeeHandler struct {
	svc            *employeeUC.Service
	requestTimeout time.Duration
}

func NewEmployeeHandler(svc *employeeUC.Service, requestTimeout time.Duration) *EmployeeHandler {
	if requestTimeout <= 0 {
		requestTimeout = 5 * time.Second
	}
	return &EmployeeHandler{svc: svc, requestTimeout: requestTimeout}
}

type createEmployeeReq struct {
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	Email      string  `json:"email"`
	Department string  `json:"department"`
	Position   string  `json:"position"`
	Salary     float64 `json:"salary"`
	Status     string  `json:"status"`
}

type updateEmployeeReq struct {
	FirstName  *string  `json:"first_name"`
	LastName   *string  `json:"last_name"`
	Email      *string  `json:"email"`
	Department *string  `json:"department"`
	Position   *string  `json:"position"`
	Salary     *float64 `json:"salary"`
	Status     *string  `json:"status"`
}

type employeeDTO struct {
	ID         string  `json:"id"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	Email      string  `json:"email"`
	Department string  `json:"department"`
	Position   string  `json:"position"`
	Salary     float64 `json:"salary"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

func toDTO(e *domainEmployee.Employee) employeeDTO {
	return employeeDTO{
		ID:         e.ID,
		FirstName:  e.FirstName,
		LastName:   e.LastName,
		Email:      e.Email,
		Department: e.Department,
		Position:   e.Position,
		Salary:     e.Salary,
		Status:     string(e.Status),
		CreatedAt:  e.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:  e.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func (h *EmployeeHandler) Create(c *gin.Context) {
	var req createEmployeeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.requestTimeout)
	defer cancel()

	e, err := h.svc.Create(ctx, employeeUC.CreateInput{
		FirstName:  strings.TrimSpace(req.FirstName),
		LastName:   strings.TrimSpace(req.LastName),
		Email:      req.Email,
		Department: strings.TrimSpace(req.Department),
		Position:   strings.TrimSpace(req.Position),
		Salary:     req.Salary,
		Status:     strings.TrimSpace(req.Status),
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, toDTO(e))
}

func (h *EmployeeHandler) Get(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.requestTimeout)
	defer cancel()

	e, err := h.svc.Get(ctx, id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, toDTO(e))
}

func (h *EmployeeHandler) List(c *gin.Context) {
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 64)

	dept := strings.TrimSpace(c.Query("department"))
	var deptPtr *string
	if dept != "" {
		deptPtr = &dept
	}

	status := strings.TrimSpace(c.Query("status"))
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	q := strings.TrimSpace(c.Query("q"))
	var qPtr *string
	if q != "" {
		qPtr = &q
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.requestTimeout)
	defer cancel()

	items, total, err := h.svc.List(ctx, employeeUC.ListInput{
		Department: deptPtr,
		Status:     statusPtr,
		Query:      qPtr,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	out := make([]employeeDTO, 0, len(items))
	for i := range items {
		e := items[i]
		out = append(out, toDTO(&e))
	}

	c.JSON(http.StatusOK, gin.H{
		"data": out,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *EmployeeHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req updateEmployeeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.requestTimeout)
	defer cancel()

	e, err := h.svc.Update(ctx, id, employeeUC.UpdateInput{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		Department: req.Department,
		Position:   req.Position,
		Salary:     req.Salary,
		Status:     req.Status,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, toDTO(e))
}

func (h *EmployeeHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.requestTimeout)
	defer cancel()

	if err := h.svc.Delete(ctx, id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}
