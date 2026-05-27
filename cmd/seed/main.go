package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mytheresa/go-hiring-challenge/app/config"
	"github.com/mytheresa/go-hiring-challenge/app/database"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Invalid configuration: %s", err)
	}

	db, closeDB, err := database.Open(context.Background(), cfg.Database)
	if err != nil {
		log.Fatalf("Database connection failed: %s", err)
	}
	defer closeDB()

	dir := os.Getenv("POSTGRES_SQL_DIR")
	if strings.TrimSpace(dir) == "" {
		log.Fatal("POSTGRES_SQL_DIR is required")
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("reading directory failed: %v", err)
	}

	var sqlFiles []os.DirEntry
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file)
		}
	}
	sort.Slice(sqlFiles, func(i, j int) bool {
		return sqlFiles[i].Name() < sqlFiles[j].Name()
	})

	for _, file := range sqlFiles {
		path := filepath.Join(dir, file.Name())

		content, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("reading file %s failed: %v", file.Name(), err)
		}

		if err := db.Exec(string(content)).Error; err != nil {
			log.Fatalf("executing %s failed: %v", file.Name(), err)
		}

		log.Printf("Executed %s successfully", file.Name())
	}
}
