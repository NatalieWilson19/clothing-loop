package app

import (
	"fmt"
	"os"

	"github.com/the-clothing-loop/website/server/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func DatabaseInit() *gorm.DB {
	// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", Config.DB_USER, Config.DB_PASS, Config.DB_HOST, Config.DB_PORT, Config.DB_NAME)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("error connecting to db: %s", err))
	}

	if os.Getenv("SERVER_NO_MIGRATE") == "" {
		DatabaseAutoMigrate(db)
	}

	return db
}

func DatabaseAutoMigrate(db *gorm.DB) {
	hadIsApprovedColumn := db.Migrator().HasColumn(&models.UserChain{}, "is_approved")

	if db.Migrator().HasTable("user_tokens") {
		columnTypes, err := db.Migrator().ColumnTypes("user_tokens")
		if err == nil {
			for _, ct := range columnTypes {
				if ct.Name() == "created_at" {
					t, _ := ct.ColumnType()
					if t == "bigint(20)" {
						tx := db.Begin()

						if !db.Migrator().HasColumn(&models.UserToken{}, "new_created_at") {
							err := tx.Exec(`ALTER TABLE user_tokens ADD new_created_at datetime(3) DEFAULT NULL`).Error
							if err != nil {
								tx.Rollback()
								break
							}
						}
						err = tx.Exec(`UPDATE user_tokens SET new_created_at = FROM_UNIXTIME(created_at)`).Error
						if err != nil {
							tx.Rollback()
							break
						}
						err = tx.Exec(`ALTER TABLE user_tokens DROP created_at`).Error
						if err != nil {
							tx.Rollback()
							break
						}
						err = tx.Exec(`ALTER TABLE user_tokens RENAME COLUMN new_created_at TO created_at`).Error
						if err != nil {
							tx.Rollback()
							break
						}

						tx.Commit()
					}
					break
				}
			}
		}
	}
	if db.Migrator().HasTable("bags") {
		columnTypes, err := db.Migrator().ColumnTypes("bags")
		if err == nil {
			for _, ct := range columnTypes {
				if ct.Name() == "number" {
					t, _ := ct.ColumnType()
					if t == "bigint(20)" {
						fmt.Printf("column number found in bags")
						db.Exec(`ALTER TABLE bags MODIFY number longtext`)
					}
				}
			}
		}
	}

	db.AutoMigrate(
		&models.Chain{},
		&models.Mail{},
		&models.Newsletter{},
		&models.User{},
		&models.Event{},
		&models.UserToken{},
		&models.UserChain{},
		&models.Bag{},
		&models.BulkyItem{},
		&models.Payment{},
	)

	if !db.Migrator().HasConstraint("user_chains", "uci_user_id_chain_id") {
		db.Exec(`
ALTER TABLE user_chains
ADD CONSTRAINT uci_user_id_chain_id
UNIQUE (user_id, chain_id)
		`)
	}
	if !hadIsApprovedColumn {
		db.Exec(`
UPDATE user_chains SET is_approved = TRUE WHERE id IN (
	SELECT uc.id FROM user_chains AS uc
	LEFT JOIN users AS u ON u.id = uc.user_id
	WHERE u.is_email_verified = TRUE && uc.is_approved IS NULL 
)
		`)

		db.Exec(`
UPDATE user_chains SET is_approved = FALSE WHERE id IN (
	SELECT uc.id FROM user_chains AS uc
	LEFT JOIN users AS u ON u.id = uc.user_id
	WHERE u.is_email_verified = FALSE && uc.is_approved IS NULL 
)
		`)
	}
}
