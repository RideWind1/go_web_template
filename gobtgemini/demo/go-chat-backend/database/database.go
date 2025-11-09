package database

import (
	"fmt"
	"go-chat-backend/config"
	"go-chat-backend/models"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() (*gorm.DB, error) {
	cfg := config.Get()

	// 构建数据库连接字符串
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	// GORM配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %w", err)
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取数据库实例失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// 运行数据库迁移
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %w", err)
	}

	DB = db
	logrus.Info("数据库连接成功")
	return db, nil
}

// runMigrations 运行数据库迁移
func runMigrations(db *gorm.DB) error {
	logrus.Info("开始数据库迁移...")

	// 自动迁移所有模型
	err := db.AutoMigrate(
		&models.User{},
		&models.ChatSession{},
		&models.ChatMessage{},
		&models.UserPreference{},
		&models.RefreshToken{},
	)

	if err != nil {
		return fmt.Errorf("模型迁移失败: %w", err)
	}

	logrus.Info("数据库迁移完成")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}
