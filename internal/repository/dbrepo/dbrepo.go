package dbrepo

import (
	"database/sql"
	"nastenka_udalosti/internal/config"
	"nastenka_udalosti/internal/repository"
)

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

// TODO: Odkomentuj až budeš dělat testy
// type testDBRepo struct {
// 	App *config.AppConfig
// 	DB  *sql.DB
// }

func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &postgresDBRepo{
		App: a,
		DB:  conn,
	}
}

// TODO: Odkomentuj až budeš dělat testy
// func NewTestingRepo(a *config.AppConfig) repository.DatabaseRepo {
// 	return &testDBRepo{
// 		App: a,
// 	}
// }
