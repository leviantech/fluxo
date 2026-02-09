package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/leviantech/fluxo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Product model for GORM
type Product struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Code  string `gorm:"uniqueIndex" json:"code"`
	Price uint   `json:"price"`
	Name  string `json:"name"`
}

var db *gorm.DB

// Request/Response types
type CreateProductRequest struct {
	Code  string `json:"code" validate:"required,alphanum,len=4"` // Complex validation: alphanumeric and length 4
	Price uint   `json:"price" validate:"required,gt=0"`         // Must be greater than 0
	Name  string `json:"name" validate:"required,min=2,max=100"`
}

type UpdateProductRequest struct {
	ID    uint   `uri:"id" validate:"required"`
	Price uint   `json:"price" validate:"omitempty,gt=0"`
	Name  string `json:"name" validate:"omitempty,min=2"`
}

type GetProductRequest struct {
	ID uint `uri:"id" validate:"required"`
}

type ProductListResponse struct {
	Total int64     `json:"total"`
	Items []Product `json:"items"`
}

type ListProductsRequest struct{}

// Handlers
func listProducts(ctx *fluxo.Context, req ListProductsRequest) (ProductListResponse, error) {
	var products []Product
	var total int64

	db.Model(&Product{}).Count(&total)
	db.Find(&products)

	if products == nil {
		products = []Product{}
	}

	return ProductListResponse{
		Total: total,
		Items: products,
	}, nil
}

func createProduct(ctx *fluxo.Context, req CreateProductRequest) (Product, error) {
	product := Product{
		Code:  req.Code,
		Price: req.Price,
		Name:  req.Name,
	}

	if err := db.Create(&product).Error; err != nil {
		return Product{}, fluxo.BadRequest(fmt.Sprintf("failed to create product: %v", err))
	}

	return product, nil
}

func getProduct(ctx *fluxo.Context, req GetProductRequest) (Product, error) {
	var product Product
	if err := db.First(&product, req.ID).Error; err != nil {
		return Product{}, fluxo.NotFound(fmt.Sprintf("product %d not found", req.ID))
	}
	return product, nil
}

func updateProduct(ctx *fluxo.Context, req UpdateProductRequest) (Product, error) {
	var product Product
	if err := db.First(&product, req.ID).Error; err != nil {
		return Product{}, fluxo.NotFound(fmt.Sprintf("product %d not found", req.ID))
	}

	if req.Price > 0 {
		product.Price = req.Price
	}
	if req.Name != "" {
		product.Name = req.Name
	}

	db.Save(&product)
	return product, nil
}

func deleteProduct(ctx *fluxo.Context, req GetProductRequest) (gin.H, error) {
	if err := db.Delete(&Product{}, req.ID).Error; err != nil {
		return nil, fluxo.BadRequest(fmt.Sprintf("failed to delete: %v", err))
	}
	return gin.H{"status": "deleted"}, nil
}

func setupDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Migrate the schema
	db.AutoMigrate(&Product{})
}

func setupApp() *fluxo.App {
	setupDB()

	app := fluxo.New().WithSwagger("GORM SQLite API", "1.0.0")

	app.GET("/products", fluxo.Handle(listProducts))
	app.POST("/products", fluxo.Handle(createProduct))
	app.GET("/products/:id", fluxo.Handle(getProduct))
	app.PUT("/products/:id", fluxo.Handle(updateProduct))
	app.DELETE("/products/:id", fluxo.Handle(deleteProduct))

	return app
}

func main() {
	app := setupApp()
	fmt.Println("🚀 GORM API starting on :8080")
	app.Start(":8080")
}
