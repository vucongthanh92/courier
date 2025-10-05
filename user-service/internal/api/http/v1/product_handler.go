package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	httpcommon "github.com/vucongthanh92/courier/user-service/helper/http_common"
	"github.com/vucongthanh92/courier/user-service/internal/domain/interfaces"
	"github.com/vucongthanh92/courier/user-service/internal/domain/models"
)

type ProductHandler struct {
	productService interfaces.ProductService
}

func NewProductHandler(
	productService interfaces.ProductService,
) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// API get products list godoc
// @Tags Product
// @Summary search products with filter and return pagination
// @Accept json
// @Produce json
// @Param  params body models.CreateCategoryReq true "CreateCategoryReq"
// @Router 	/api/v1/products [get]
// @Success	200
func (h *ProductHandler) GetProductList(c *gin.Context) {
	var (
		req    models.ProductListFilter
		paging = httpcommon.ParseParams(c)
	)

	err := httpcommon.GetQueryParamsHTTP(c, &req)
	if err != nil {
		return
	}

	req.Limit = paging.Limit
	req.Offset = paging.Offset

	res, totalRows, resErr := h.productService.GetProductsByFilter(c, req)
	if resErr != nil {
		resErr.ExposeHttpError(c)
		return
	}

	c.JSON(http.StatusOK, httpcommon.NewPagingSuccessResponse(res, int(totalRows), nil, req.Limit))
}
