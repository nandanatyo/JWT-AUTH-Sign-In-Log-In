package initializers

import (
	"auth/models"
)

func SyncDatabase() {
	DB.AutoMigrate(&models.User{})
}