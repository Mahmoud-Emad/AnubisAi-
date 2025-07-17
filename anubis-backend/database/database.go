package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"anubis-backend/config"
	"anubis-backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the database connection
func InitDatabase(cfg *config.Config) error {
	var err error
	var dialector gorm.Dialector

	// Configure GORM logger
	var gormLogger logger.Interface
	if cfg.Env == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	switch cfg.Database.Type {
	case "sqlite":
		// Ensure directory exists for SQLite
		if err := ensureDir(filepath.Dir(cfg.Database.SQLitePath)); err != nil {
			return fmt.Errorf("failed to create SQLite directory: %w", err)
		}
		dialector = sqlite.Open(cfg.Database.SQLitePath)
		log.Printf("Using SQLite database: %s", cfg.Database.SQLitePath)

	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			cfg.Database.Host,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Name,
			cfg.Database.Port,
			cfg.Database.SSLMode,
		)
		dialector = postgres.Open(dsn)
		log.Printf("Using PostgreSQL database: %s@%s:%d/%s",
			cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	default:
		return fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}

	// Connect to database
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("Database connection established successfully")
	return nil
}

// RunMigrations runs database migrations
func RunMigrations() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	log.Println("Running database migrations...")

	// Auto-migrate all models
	err := DB.AutoMigrate(
		&models.User{},
		&models.UserSetting{},
		&models.UserMemory{},
		&models.TaskExecution{},
		&models.PasswordReset{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// CloseDatabase closes the database connection
func CloseDatabase() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	log.Println("Database connection closed")
	return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// HealthCheck performs a database health check
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// ensureDir creates directory if it doesn't exist
func ensureDir(dir string) error {
	if dir == "" || dir == "." {
		return nil
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		log.Printf("Created directory: %s", dir)
	}
	return nil
}

// SeedDatabase seeds the database with initial data (for development)
func SeedDatabase() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	log.Println("Seeding database with initial data...")

	// Check if admin user already exists
	var adminUser models.User
	result := DB.Where("email = ?", "admin@anubis.local").First(&adminUser)
	if result.Error == nil {
		log.Println("Admin user already exists, skipping seed")
		return nil
	}

	// Create admin user (password should be hashed in real implementation)
	adminUser = models.User{
		Email:     "admin@anubis.local",
		Username:  "admin",
		Password:  "admin123", // This should be hashed
		FirstName: "Admin",
		LastName:  "User",
		IsActive:  true,
		IsAdmin:   true,
	}

	if err := DB.Create(&adminUser).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	log.Println("Database seeded successfully")
	return nil
}
