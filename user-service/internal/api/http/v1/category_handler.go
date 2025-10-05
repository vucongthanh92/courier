package v1

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vucongthanh92/courier/user-service/helper/constants"
	errHandler "github.com/vucongthanh92/courier/user-service/helper/error_handler"
	httpcommon "github.com/vucongthanh92/courier/user-service/helper/http_common"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

type CategoryHandler struct {
	categoryService interfaces.CategoryServiceI
}

func NewCategoryHandler(categoryService interfaces.CategoryServiceI) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// API CreateCategory godoc
// @Tags Category
// @Summary create category by name
// @Accept json
// @Produce json
// @Param params body models.CreateCategoryReq true "CreateCategoryReq"
// @Router /api/v1/category [post]
// @Success	200 {object} entities.Category
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	req := models.CreateCategoryReq{}

	err := httpcommon.ValidatorParams(req)
	if err != nil {
		resErr := errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusBadRequest).
			SetArrayError(err)
		resErr.ExposeHttpError(c)
		return
	}

	if err := httpcommon.GetBodyParamsHTTP(c, &req); err != nil {
		return
	}

	res, resErr := h.categoryService.CreateCategory(c, req)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}

// API UpdateCategory godoc
// @Tags Category
// @Summary update category by id
// @Accept json
// @Produce json
// @Param params body models.CreateCategoryReq true "UpdateCategoryReq"
// @Router /api/v1/category [put]
// @Success	200 {object} entities.Category
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	var (
		req = models.UpdateCategoryReq{}
		err error
	)

	if err := httpcommon.GetBodyParamsHTTP(c, &req); err != nil {
		return
	}

	objectID := c.Param("id")
	req.ID, err = strconv.ParseUint(objectID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, httpcommon.NewErrorResponse("Invalid ID", constants.REQUEST_INVALID, ""))
		return
	}

	if err := httpcommon.ValidatorParams(req); len(err) > 0 {
		errHandler.InitErrorBuilder(c).
			SetLogError(errors.New(constants.InvalidValue)).
			SetStatus(http.StatusBadRequest).
			SetArrayError(err).
			ExposeHttpError(c)
		return
	}

	res, resErr := h.categoryService.UpdateCategory(c, req)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}

// API DeleteCategoryByID godoc
// @Tags Category
// @Summary delete category by id
// @Accept json
// @Produce json
// @Param params body models.CreateCategoryReq true "UpdateCategoryReq"
// @Router /api/v1/category [put]
// @Success	200 {object} entities.Category
func (h *CategoryHandler) DeleteCategoryByID(c *gin.Context) {
	var req = struct {
		UpdatedAt time.Time `json:"updated_at"`
	}{}

	if err := httpcommon.GetBodyParamsHTTP(c, &req); err != nil {
		return
	}

	objectID := c.Param("id")
	categoryID, err := strconv.ParseUint(objectID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, httpcommon.NewErrorResponse("Invalid ID", constants.REQUEST_INVALID, ""))
		return
	}

	resErr := h.categoryService.DeleteCategoryByID(c, categoryID, req.UpdatedAt)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse[any](nil))
}

// API GetCategoryList godoc
// @Tags Category
// @Summary get list categories
// @Accept json
// @Produce json
// @Router /api/v1/category [get]
// @Success 200 {object} []entities.Category
func (h *CategoryHandler) GetCategoryList(c *gin.Context) {

	res, resErr := h.categoryService.GetCategoryList(c)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}

// API GetCategoryByID godoc
// @Tags Category
// @Summary get list categories
// @Accept json
// @Produce json
// @Router /api/v1/category/:id [get]
// @Success 200 {object} entities.Category
func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	param := c.Param("id")
	categoryID, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, httpcommon.NewErrorResponse("Invalid ID", constants.REQUEST_INVALID, ""))
		return
	}

	res, resErr := h.categoryService.GetCategoryByID(c, categoryID)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	c.JSON(http.StatusOK, httpcommon.NewSuccessResponse(res))
}
