package handler

import (
	"log/slog"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	department   DepartmentUsecase
	eduDirection EduDirectionUsecase
	eduGroup     EduGroupUsecase
	eduPlan      EduPlanUsecase
	schedule     ScheduleUsecase
	teacher      TeacherUsecase
	logger       *slog.Logger
}

func NewHandler(
	department DepartmentUsecase,
	eduDirection EduDirectionUsecase,
	eduGroup EduGroupUsecase,
	eduPlan EduPlanUsecase,
	schedule ScheduleUsecase,
	teacher TeacherUsecase,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		department:   department,
		eduDirection: eduDirection,
		eduGroup:     eduGroup,
		eduPlan:      eduPlan,
		teacher:      teacher,
		schedule:     schedule,
		logger:       logger,
	}
}

func (h *Handler) InitRouter() *echo.Echo {
	router := echo.New()
	router.HTTPErrorHandler = NewHttpErrorHandler(h.logger)

	api := router.Group("/v1")

	departments := api.Group("/departments")
	{
		departments.POST("", h.CreateDepartment)
		departments.GET("", h.ListDepartment)
		departments.GET("/:id", h.GetDepartment)
		departments.PUT("/:id", h.UpdateDepartment)
		departments.DELETE("/:id", h.DeleteDepartment)
	}

	directions := api.Group("/edu-directions")
	{
		directions.POST("", h.CreateEduDirection)
		directions.GET("", h.ListEduDirection)
		directions.GET("/:id", h.GetEduDirection)
		directions.PUT("/:id", h.UpdateEduDirection)
		directions.DELETE("/:id", h.DeleteEduDirection)
	}

	groups := api.Group("/edu-groups")
	{
		groups.POST("", h.CreateEduGroup)
		groups.GET("", h.ListEduGroup)
		groups.GET("/:id", h.GetEduGroup)
		groups.PUT("/:id", h.UpdateEduGroup)
		groups.DELETE("/:id", h.DeleteEduGroup)
	}

	plans := api.Group("/edu-plans")
	{
		plans.POST("", h.CreateEduPlan)
		plans.GET("", h.ListEduPlan)
		plans.GET("/:id", h.GetEduPlan)
		plans.DELETE("/:id", h.DeleteEduPlan)
	}

	schedules := api.Group("/schedules")
	{
		schedules.POST("", h.CreateSchedule)
		schedules.GET("", h.ListSchedule)
		schedules.GET("/:id", h.GetSchedule)
		schedules.GET("/:id/export", h.ExportSchedule)
		schedules.PUT("/:id/items", h.AddScheduleItem)
		schedules.DELETE("/:id/items", h.RemoveScheduleItem)
	}

	teachers := api.Group("/teachers")
	{
		teachers.POST("", h.CreateTeacher)
		teachers.GET("", h.ListTeacher)
		teachers.GET("/:id", h.GetTeacher)
		teachers.PUT("/:id", h.UpdateTeacher)
		teachers.DELETE("/:id", h.DeleteTeacher)
	}

	return router
}
