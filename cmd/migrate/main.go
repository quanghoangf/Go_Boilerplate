package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"boilerplate/pkg/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq" // postgres driver
)

func main() {
	var (
		configPath = flag.String("conf", "config/local.yml", "config path, eg: -conf ./config/local.yml")
		direction  = flag.String("dir", "up", "migration direction: up, down, version")
		steps      = flag.Int("steps", 0, "number of steps to migrate (0 = all)")
		version    = flag.Uint("version", 0, "migrate to specific version")
	)
	flag.Parse()

	// Load configuration
	conf := config.NewConfig(*configPath)
	
	// Get database connection string
	driver := conf.GetString("data.db.user.driver")
	dsn := conf.GetString("data.db.user.dsn")
	
	if driver != "postgres" {
		log.Fatal("This migration tool only supports PostgreSQL. Please update your config to use postgres driver.")
	}

	// Connect to database
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Create postgres driver instance
	driver_instance, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("Failed to create postgres driver:", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", 
		driver_instance,
	)
	if err != nil {
		log.Fatal("Failed to create migrate instance:", err)
	}
	defer m.Close()

	// Execute migration based on direction
	switch *direction {
	case "up":
		if *version > 0 {
			err = m.Migrate(*version)
		} else if *steps > 0 {
			err = m.Steps(*steps)
		} else {
			err = m.Up()
		}
		
	case "down":
		if *steps > 0 {
			err = m.Steps(-(*steps))
		} else {
			err = m.Down()
		}
		
	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatal("Failed to get version:", err)
		}
		fmt.Printf("Current version: %d, Dirty: %v\n", version, dirty)
		return
		
	case "force":
		if *version == 0 {
			log.Fatal("Version is required for force command")
		}
		err = m.Force(int(*version))
		
	default:
		log.Fatal("Invalid direction. Use: up, down, version, or force")
	}

	// Handle migration results
	if err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("No changes to apply")
		} else {
			log.Fatal("Migration failed:", err)
		}
	} else {
		fmt.Printf("Migration %s completed successfully\n", *direction)
	}
}