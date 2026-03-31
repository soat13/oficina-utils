package fiber

import (
	"github.com/gofiber/fiber/v2"
	"github.com/soat13/oficina-utils/pkg/pagination"
	stringHelper "github.com/soat13/oficina-utils/pkg/utils/helpers/string"
)

func NewPagination(ctx *fiber.Ctx, defaultLimit, defaultOffset int) *pagination.Pagination {
	limit := stringHelper.StringToIntOrDefault(ctx.Query("limit"), defaultLimit)
	offset := stringHelper.StringToIntOrDefault(ctx.Query("offset"), defaultOffset)
	return pagination.New(limit, offset)
}
