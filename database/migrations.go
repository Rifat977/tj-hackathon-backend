package database

import (
	"log"
)

// RunMigrations handles database schema migrations
func RunMigrations() {
	log.Println("Running database migrations...")

	// Check if users table has 'name' column and migrate to 'first_name'/'last_name'
	migrateUsersTable()

	log.Println("Database migrations completed!")
}

// migrateUsersTable migrates the users table from 'name' column to 'first_name'/'last_name' columns
func migrateUsersTable() {
	// Check if users table exists
	var tableExists bool
	err := DB.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'users'
		)
	`).Scan(&tableExists).Error

	if err != nil {
		log.Printf("Error checking if users table exists: %v", err)
		return
	}

	if !tableExists {
		log.Println("Users table does not exist, will be created by AutoMigrate")
		return
	}

	// Check if 'name' column exists
	var nameColumnExists bool
	err = DB.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'users' AND column_name = 'name'
		)
	`).Scan(&nameColumnExists).Error

	if err != nil {
		log.Printf("Error checking for 'name' column: %v", err)
		return
	}

	// Check if first_name column exists
	var firstNameExists bool
	err = DB.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'users' AND column_name = 'first_name'
		)
	`).Scan(&firstNameExists).Error

	if err != nil {
		log.Printf("Error checking for 'first_name' column: %v", err)
		return
	}

	// If name column exists but first_name doesn't, we need to migrate
	if nameColumnExists && !firstNameExists {
		log.Println("Migrating users table from 'name' to 'first_name'/'last_name'...")

		// Add first_name and last_name columns
		err = DB.Exec(`
			ALTER TABLE users 
			ADD COLUMN first_name VARCHAR(255),
			ADD COLUMN last_name VARCHAR(255)
		`).Error

		if err != nil {
			log.Printf("Error adding first_name/last_name columns: %v", err)
			return
		}

		// Split existing name data into first_name and last_name
		err = DB.Exec(`
			UPDATE users 
			SET 
				first_name = CASE 
					WHEN name LIKE '% %' THEN SPLIT_PART(name, ' ', 1)
					ELSE name
				END,
				last_name = CASE 
					WHEN name LIKE '% %' THEN SUBSTRING(name FROM POSITION(' ' IN name) + 1)
					ELSE ''
				END
			WHERE first_name IS NULL OR last_name IS NULL
		`).Error

		if err != nil {
			log.Printf("Error splitting name data: %v", err)
			return
		}

		// Make first_name and last_name NOT NULL
		err = DB.Exec(`
			ALTER TABLE users 
			ALTER COLUMN first_name SET NOT NULL,
			ALTER COLUMN last_name SET NOT NULL
		`).Error

		if err != nil {
			log.Printf("Error making columns NOT NULL: %v", err)
			return
		}

		// Drop the old name column
		err = DB.Exec(`ALTER TABLE users DROP COLUMN name`).Error

		if err != nil {
			log.Printf("Error dropping 'name' column: %v", err)
			return
		}

		log.Println("Successfully migrated users table from 'name' to 'first_name'/'last_name'")
	} else if nameColumnExists && firstNameExists {
		// Both columns exist, drop the name column
		log.Println("Both 'name' and 'first_name' columns exist. Dropping 'name' column...")
		err = DB.Exec(`ALTER TABLE users DROP COLUMN name`).Error
		if err != nil {
			log.Printf("Error dropping 'name' column: %v", err)
			return
		}
		log.Println("Successfully dropped 'name' column")
	} else if !nameColumnExists && !firstNameExists {
		// Neither column exists, table needs to be recreated
		log.Println("Users table exists but missing both 'name' and 'first_name' columns. Dropping and recreating...")
		err = DB.Exec(`DROP TABLE IF EXISTS users CASCADE`).Error
		if err != nil {
			log.Printf("Error dropping users table: %v", err)
			return
		}
		log.Println("Users table dropped. Will be recreated with correct schema.")
	} else {
		log.Println("Users table schema is correct (has first_name/last_name columns)")
	}
}
