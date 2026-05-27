package user

import (
	"bytes"
	"io"
	"net/http"

	"project1/internal/response"

	"github.com/gin-gonic/gin"
)

const maxImportFileSize = 2 * 1024 * 1024

type ImportHandler struct {
	service *ImportService
}

func NewImportHandler(service *ImportService) *ImportHandler {
	return &ImportHandler{service: service}
}

func (h *ImportHandler) ImportUsers(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "请选择CSV文件")
		return
	}
	if fileHeader.Size <= 0 {
		response.Error(c, http.StatusBadRequest, "CSV文件不能为空")
		return
	}
	if fileHeader.Size > maxImportFileSize {
		response.Error(c, http.StatusRequestEntityTooLarge, "CSV文件不能超过2MB")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无法读取CSV文件")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无法读取CSV文件内容")
		return
	}

	result, err := h.service.ImportCSV(c.Request.Context(), bytes.NewReader(data))
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, result)
}
