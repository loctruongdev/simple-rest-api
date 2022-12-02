package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"strconv"
)

//CREATE TABLE `restaurants` (
//`id` int NOT NULL AUTO_INCREMENT,
//`owner_id` int DEFAULT NULL,
//`name` varchar(50) NOT NULL,
//`addr` varchar(255) NOT NULL,
//`city_id` int DEFAULT NULL,
//`lat` double DEFAULT NULL,
//`lng` double DEFAULT NULL,
//`cover` json DEFAULT NULL,
//`logo` json DEFAULT NULL,
//`shipping_fee_per_km` double DEFAULT '0',
//`status` int NOT NULL DEFAULT '1',
//`created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
//`updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
//PRIMARY KEY (`id`),
//KEY `owner_id` (`owner_id`) USING BTREE,
//KEY `city_id` (`city_id`) USING BTREE,
//KEY `status` (`status`) USING BTREE
//) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3

type Restaurant struct {
	Id   int    `json:"id" gorm:"column:id;"`
	Name string `json:"name" gorm:"column:name;"`
	Addr string `json:"addr" gorm:"column:addr;"`
}

func (Restaurant) TableName() string {
	return "restaurants"
}

type RestaurantUpdate struct {
	Name *string `json:"name" gorm:"column:name;"`
	Addr *string `json:"addr" gorm:"column:addr;"`
}

func (RestaurantUpdate) TableName() string {
	return Restaurant{}.TableName()
}

func main() {
	//dsn := "dbadmin:dbadmin@tcp(127.0.0.1:3306)/simple-rest-api?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := os.Getenv("DB_CONN_STR")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}

	if err := runService(db); err != nil {
		log.Fatalln(err)
	}

}

func runService(db *gorm.DB) error {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	restaurants := r.Group("/restaurants")
	{
		restaurants.POST("", func(c *gin.Context) {
			var data Restaurant

			if err := c.ShouldBind(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})

				return
			}

			if data.Name == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": errors.New("restaurant name can not be blank"),
				})

				return
			}

			if err := db.Create(&data).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})

				return
			}

			c.JSON(http.StatusOK, data)
		})

		restaurants.GET("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})

				return
			}

			var data Restaurant

			if err := db.Where("id =?", id).First(&data).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})

				return
			}

			c.JSON(http.StatusOK, data)
		})

		restaurants.GET("", func(c *gin.Context) {
			var data []Restaurant

			type Filter struct {
				Cityid int `json:"city_id" form:"city_id"`
			}

			var filter Filter

			c.ShouldBind(&filter)

			newDb := db

			if filter.Cityid > 0 {
				newDb = db.Where("city_id = ?", filter.Cityid)
			}

			if err := newDb.Find(&data).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})

				return
			}

			c.JSON(http.StatusOK, data)
		})

		restaurants.PATCH("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})

				return
			}

			var data RestaurantUpdate

			if err := c.ShouldBind(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})

				return
			}

			if err := db.Where("id =?", id).Updates(&data).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})

				return
			}

			c.JSON(http.StatusOK, gin.H{"ok": 1})
		})

		restaurants.DELETE("/:id", func(c *gin.Context) {
			id, err := strconv.Atoi(c.Param("id"))

			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})

				return
			}

			if err := db.Table(Restaurant{}.TableName()).
				Where("id = ?", id).
				Delete(nil).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})

				return
			}
			c.JSON(http.StatusOK, "ok")
		})
	}

	return r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
