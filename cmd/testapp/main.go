package main

import (
	"log"
	"path/filepath"

	"gorm.io/gorm"

	"testapp/internal/handlers"
	repPgSQL "testapp/internal/repositories/pgsql"
	"testapp/internal/services"
	"testapp/pkg/config"
	"testapp/pkg/pgsql"

	// repsPgSQL "testapp/internal/repositories/pgsql"
	"testapp/pkg/http"
)

const (
	CONFIG_NAME = "config"
	CONFIG_EXTENSION = "yaml" 
	CONFIG_PATH = "."
)

func main() {
	// Setting configs
	conf, err := config.LoadConfig(filepath.Join(".", "configs", CONFIG_NAME), CONFIG_EXTENSION, CONFIG_PATH)
	if err != nil {
		log.Fatalf("Error loading a config: %v", err)
	}

	// Connecting to database
	db, err := pgsql.NewPgSQLConnection(conf.PgSQL)
	if err != nil {
		log.Fatalf("Error connecting to pgsql db: %v", err)
	}

	log.Println("Database connected succesfully!")
	
	// Check database size
	DatabaseSize, err := CheckDBSize(db)
	if err != nil {
		log.Fatalf("Failed to check database size: %v", err)
	}

	log.Printf("Database Size: %s\n", DatabaseSize)

	imageRep := repPgSQL.NewImageRepository(db)
	imageServ := services.NewImageService(imageRep)
	imageHandler := handlers.NewImageHandler(imageServ)
	formatHandler := handlers.NewFormatHandler()

	// Creating new server and starting to listen
	srv := http.NewServer(conf.HTTP, imageHandler, formatHandler)
	
	log.Printf("We are starting on %v", srv.Addr)
	
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

type DBSize struct {
	DatabaseSize string
}

func CheckDBSize(db *gorm.DB) (DatabaseSize string, err error) {
	var dbSize DBSize 

	err = db.Raw("SELECT pg_size_pretty(pg_database_size(current_database())) AS database_size").Scan(&dbSize).Error
	if err != nil {
		return "", err
	}

	return dbSize.DatabaseSize, nil
}