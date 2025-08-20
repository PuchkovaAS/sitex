package main

import (
	"os/user"
	"sitex/config"
	"sitex/internal/association"
	"sitex/internal/project"
	"sitex/pkg/db"
)

func main() {
	config.Init()

	dbConfig := config.NewDatabaseConfig()
	db := db.NewDb(dbConfig)
	db.AutoMigrate(
		&user.User{},
		&project.Project{},
		&association.UserProject{},
	)
}
