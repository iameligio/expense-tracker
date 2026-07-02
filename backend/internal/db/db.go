package db

import (
	"fmt"
	"log"

	"expense-tracker/backend/internal/config"
	"expense-tracker/backend/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect opens a GORM connection to MariaDB using the given config.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	gormCfg := &gorm.Config{}
	if cfg.IsProduction {
		gormCfg.Logger = logger.Default.LogMode(logger.Warn)
	} else {
		gormCfg.Logger = logger.Default.LogMode(logger.Info)
	}

	gdb, err := gorm.Open(mysql.Open(cfg.DSN()), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	return gdb, nil
}

// Migrate runs AutoMigrate for all models.
func Migrate(gdb *gorm.DB) error {
	if err := gdb.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Expense{},
		&models.AppSetting{},
		&models.RefreshToken{},
	); err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}
	return nil
}

// Seed ensures the global AppSetting row exists (default: 20% savings target).
func Seed(gdb *gorm.DB) error {
	var count int64
	if err := gdb.Model(&models.AppSetting{}).
		Where("id = ?", models.GlobalSettingID).Count(&count).Error; err != nil {
		return fmt.Errorf("check app setting: %w", err)
	}
	if count == 0 {
		setting := models.AppSetting{
			ID:                 models.GlobalSettingID,
			SavingsTargetType:  models.SavingsPercent,
			SavingsTargetValue: decimal.NewFromInt(20),
		}
		if err := gdb.Create(&setting).Error; err != nil {
			return fmt.Errorf("seed app setting: %w", err)
		}
		log.Println("seeded default app setting (20% savings target)")
	}
	return nil
}
