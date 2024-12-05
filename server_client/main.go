package main

import (
	"log"
	"testapp/pgsql"
	// repsPgSQL "testapp/repositories/pgsql"
	"testapp/server"

	"gorm.io/gorm"
)


func main() {
	config := pgsql.Config{
		Host: "localhost",
		User: "user",
		Password: "password",
		DBname: "mydb",
		Port: 5432,
		SSLmode: "disable",
		Timezone: "Europe/Kyiv",
	}

	db, err := pgsql.NewPgSQLConnection(config)
	if err != nil {
		log.Fatalf("Error connecting to pgsql db: %v", err)
	}

	log.Println("Database connected succesfully!")
	
	DatabaseSize, err := CheckDBSize(db)
	if err != nil {
		log.Fatalf("Failed to check database size: %v", err)
	}

	log.Printf("Database Size: %s\n", DatabaseSize)

	// Server
	var port uint16 = 8080
	srv := server.NewServer(port)
	
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