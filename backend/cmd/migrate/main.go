package main

import (
	"flag"
	"log"

	"chatapp/internal/database"
)

func main() {
	var (
		migrate = flag.Bool("migrate", false, "Run database migrations")
		seed    = flag.Bool("seed", false, "Seed initial data")
		help    = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	if *migrate {
		log.Println("Running migrations...")
		if err := database.Migrate(); err != nil {
			log.Fatal("Migration failed:", err)
		}
		log.Println("Migrations completed successfully")
	}

	if *seed {
		log.Println("Seeding data...")
		if err := database.SeedData(); err != nil {
			log.Fatal("Seeding failed:", err)
		}
		log.Println("Seeding completed successfully")
	}

	if !*migrate && !*seed {
		log.Println("No action specified. Use -help for usage information.")
	}
}

func showHelp() {
	log.Println("Database Migration Tool")
	log.Println("Usage:")
	log.Println("  go run cmd/migrate/main.go [options]")
	log.Println("")
	log.Println("Options:")
	log.Println("  -migrate    Run database migrations")
	log.Println("  -seed       Seed initial data")
	log.Println("  -help       Show this help message")
	log.Println("")
	log.Println("Examples:")
	log.Println("  go run cmd/migrate/main.go -migrate")
	log.Println("  go run cmd/migrate/main.go -migrate -seed")
}