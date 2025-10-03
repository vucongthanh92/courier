package v1

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func MapRoutes(
	router *gin.Engine,
	productHandler *ProductHandler,
	categoryHandler *CategoryHandler,
	supplierHandler *SupplierHandler,
) {
	v1 := router.Group("/api/v1")
	{
		// api for category
		v1.POST("/category", categoryHandler.CreateCategory)
		v1.PUT("/category/:id", categoryHandler.UpdateCategory)
		v1.DELETE("/category/:id", categoryHandler.DeleteCategoryByID)
		v1.GET("/categories", categoryHandler.GetCategoryList)
		v1.GET("/category/:id", categoryHandler.GetCategoryByID)

		// api for product
		v1.GET("/products", productHandler.GetProductList)
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
