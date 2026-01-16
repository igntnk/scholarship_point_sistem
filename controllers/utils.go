package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/igntnk/scholarship_point_system/errors/parsing"
	"github.com/igntnk/scholarship_point_system/errors/unexpected"
	"github.com/igntnk/scholarship_point_system/errors/validation"
	"net/http"
)

func processHttpError(c *gin.Context, err error) {

	switch {
	case errors.Is(err, parsing.OutputDataErr), errors.Is(err, unexpected.RequestErr):
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	case errors.Is(err, parsing.InputDataErr), errors.Is(err, validation.RecordAlreadyExistsErr):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, validation.NoDataFoundErr):
		c.JSON(http.StatusNoContent, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func createResponse(data any) gin.H {
	return gin.H{
		"data": data,
	}
}

func createResponseWithPagination[T any](data []T, limit, offset, totalRecords int) gin.H {
	return gin.H{
		"data": data,
		"pagination": gin.H{
			"total_records": totalRecords,
			"limit":         limit,
			"offset":        offset,
			"selected":      len(data),
		},
	}
}
