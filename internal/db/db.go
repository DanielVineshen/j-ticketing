// FILE: internal/db/db.go
package db

import (
	"database/sql"
	"fmt"
	"j-ticketing/internal/db/migrations"
	"j-ticketing/internal/db/models" // Update this import path if needed
	"j-ticketing/pkg/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config for controlling migration behavior
type DBConfig struct {
	AutoMigrate       bool // Equivalent to Spring Boot's ddl-auto=update when true
	CreateConstraints bool // Whether to create foreign key constraints
	LogLevel          logger.LogLevel
}

// GetDBConnection returns a GORM db connection with applied settings
func GetDBConnection(config *config.Config, dbConfig *DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.DB.User,
		config.DB.Password,
		config.DB.Host,
		config.DB.Port,
		config.DB.Name,
	)

	// Set default config if not provided
	if dbConfig == nil {
		dbConfig = &DBConfig{
			AutoMigrate:       false,
			CreateConstraints: true,
			LogLevel:          logger.Info,
		}
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(dbConfig.LogLevel),
		DisableForeignKeyConstraintWhenMigrating: true, // Disable foreign key constraints during migration
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Apply auto migrations if enabled
	if dbConfig.AutoMigrate {
		err = AutoMigrateSchema(db, dbConfig.CreateConstraints)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-migrate schema: %w", err)
		}
	}

	return db, nil
}

// AutoMigrateSchema automatically migrates all model schemas
// This is equivalent to Spring Boot's ddl-auto=update functionality
func AutoMigrateSchema(db *gorm.DB, createConstraints bool) error {
	// PHASE 1: Create all tables WITHOUT foreign key constraints
	fmt.Println("Phase 1: Creating tables without foreign keys...")

	// Disable foreign key checks during migration
	db.Exec("SET FOREIGN_KEY_CHECKS=0")

	// Create tables in order of dependency (independent tables first)
	baseTables := []interface{}{
		&models.Customer{},
		&models.TicketGroup{},
		&models.Tag{},
		&models.Admin{},
		&models.Token{},
		&models.AuditLog{},
	}

	for _, model := range baseTables {
		tableName := ""
		if t, ok := model.(interface{ TableName() string }); ok {
			tableName = t.TableName()
		}
		fmt.Printf("Creating table: %s\n", tableName)
		if err := db.Migrator().CreateTable(model); err != nil {
			db.Exec("SET FOREIGN_KEY_CHECKS=1") // Make sure to re-enable before returning
			return fmt.Errorf("failed to create table %s: %w", tableName, err)
		}
	}

	// Create dependent tables
	dependentTables := []interface{}{
		&models.Banner{},
		&models.OrderTicketGroup{},
		&models.GroupGallery{},
		&models.TicketTag{},
		&models.OrderTicketInfo{},
		&models.TicketDetail{},
	}

	for _, model := range dependentTables {
		tableName := ""
		if t, ok := model.(interface{ TableName() string }); ok {
			tableName = t.TableName()
		}
		fmt.Printf("Creating table: %s\n", tableName)
		if err := db.Migrator().CreateTable(model); err != nil {
			db.Exec("SET FOREIGN_KEY_CHECKS=1") // Make sure to re-enable before returning
			return fmt.Errorf("failed to create table %s: %w", tableName, err)
		}
	}

	// Re-enable foreign key checks
	db.Exec("SET FOREIGN_KEY_CHECKS=1")

	// PHASE 2: Add foreign key constraints (optional)
	if createConstraints {
		fmt.Println("Phase 2: Adding foreign key constraints...")

		// Define the constraints to add
		constraints := []struct {
			table       string
			constraint  string
			foreignKey  string
			reference   string
			referenceID string
		}{
			{
				table:       "Order_Ticket_Group",
				constraint:  "fk_order_ticket_group_ticket_group",
				foreignKey:  "ticket_group_id",
				reference:   "Ticket_Group",
				referenceID: "ticket_group_id",
			},
			{
				table:       "Order_Ticket_Group",
				constraint:  "fk_order_ticket_group_customer",
				foreignKey:  "cust_id",
				reference:   "Customer",
				referenceID: "cust_id",
			},
			{
				table:       "Order_Ticket_Info",
				constraint:  "fk_order_ticket_info_order_ticket_group",
				foreignKey:  "order_ticket_group_id",
				reference:   "Order_Ticket_Group",
				referenceID: "order_ticket_group_id",
			},
			{
				table:       "Group_Gallery",
				constraint:  "fk_group_gallery_ticket_group",
				foreignKey:  "ticket_group_id",
				reference:   "Ticket_Group",
				referenceID: "ticket_group_id",
			},
			{
				table:       "Ticket_Tag",
				constraint:  "fk_ticket_tag_ticket_group",
				foreignKey:  "ticket_group_id",
				reference:   "Ticket_Group",
				referenceID: "ticket_group_id",
			},
			{
				table:       "Ticket_Tag",
				constraint:  "fk_ticket_tag_tag",
				foreignKey:  "tag_id",
				reference:   "Tag",
				referenceID: "tag_id",
			},
			{
				table:       "Ticket_Detail",
				constraint:  "fk_ticket_detail_ticket_group",
				foreignKey:  "ticket_group_id",
				reference:   "Ticket_Group",
				referenceID: "ticket_group_id",
			},
		}

		for _, c := range constraints {
			sql := fmt.Sprintf(
				"ALTER TABLE `%s` ADD CONSTRAINT `%s` FOREIGN KEY (`%s`) REFERENCES `%s`(`%s`)",
				c.table,
				c.constraint,
				c.foreignKey,
				c.reference,
				c.referenceID,
			)

			fmt.Printf("Adding constraint: %s\n", c.constraint)
			if err := db.Exec(sql).Error; err != nil {
				// Log the error but continue with other constraints
				fmt.Printf("Warning: Failed to add constraint %s: %v\n", c.constraint, err)
			}
		}
	}

	fmt.Println("Migration completed successfully")
	return nil
}

// GetRawDB gets the underlying SQL DB connection
// This is useful when working with migrations
func GetRawDB(db *gorm.DB) (*sql.DB, error) {
	return db.DB()
}

// RunMigrations runs SQL migrations when auto-migration is off
// This is similar to Spring Boot with ddl-auto=none where you'd handle migrations separately
func RunMigrations(db *gorm.DB) error {
	sqlDB, err := GetRawDB(db)
	if err != nil {
		return err
	}

	// Run migrations using golang-migrate
	err = migrations.Migrate(sqlDB)
	if err != nil {
		return err
	}

	return nil
}
