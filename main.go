package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Student struct {
	ID       uint      `json:"id" gorm:"primary_key"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Password string    `json:"-" gorm:"column:password"`
	Absences []Absence `json:"absences"`
}

type Absence struct {
	ID            uint      `json:"id" gorm:"primary_key"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	Justification string    `json:"justification"`
	StudentID     uint      `json:"-"`
}

func connectDatabase() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root@tcp(localhost:3306)/uml?charset=utf8&parseTime=True"))
	if err != nil {
		panic("Failed to connect to database!")
	}

	db.AutoMigrate(&Student{}, &Absence{})

	return db
}

func main() {
	r := gin.Default()
	db := connectDatabase()

	r.GET("/students", func(c *gin.Context) {
		var students []Student

		if err := db.Find(&students).Limit(19).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, students)
	})

	r.POST("/students", func(c *gin.Context) {
		var student Student

		if err := c.ShouldBindJSON(&student); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		db.Create(&student)

		c.JSON(http.StatusOK, student)
	})

	r.GET("/students/:id/absences", func(c *gin.Context) {
		var student Student
		var absences []Absence

		if err := db.Preload("Absences").Limit(19).First(&student, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Student not found!"})
			return
		}

		for _, absence := range student.Absences {
			absences = append(absences, absence)
		}

		c.JSON(http.StatusOK, absences)
	})

	r.POST("/students/:id/absences", func(c *gin.Context) {
		var student Student
		var absence Absence

		if err := db.First(&student, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Student not found!"})
			return
		}

		if err := c.ShouldBindJSON(&absence); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		absence.StudentID = student.ID

		db.Create(&absence)
		c.JSON(http.StatusOK, absence)
	})

	r.POST("/students/:id/absences/:absenceId/justification", func(c *gin.Context) {
		var student Student
		var absence Absence

		if err := db.First(&student, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Student not found!"})
			return
		}

		if err := db.Where("student_id = ?", student.ID).First(&absence, c.Param("absenceId")).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Absence not found!"})
			return
		}

		file, err := c.FormFile("justification")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		filename := fmt.Sprintf("%d_%d_%s", student.ID, absence.ID, file.Filename)

		if err := c.SaveUploadedFile(file, filename); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		absence.Justification = filename

		db.Save(&absence)

		c.JSON(http.StatusOK, absence)
	})

	r.GET("/students/:id/absences/:absenceId/justification", func(c *gin.Context) {
		var student Student
		var absence Absence

		if err := db.First(&student, c.Param("id")).Limit(19).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Student not found!"})
			return
		}

		if err := db.Where("student_id = ?", student.ID).First(&absence, c.Param("absenceId")).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Absence not found!"})
			return
		}

		c.File(absence.Justification)
	})

	r.Run(":8080")
}
