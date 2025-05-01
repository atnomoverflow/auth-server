package database

import (
	"database/sql"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

const (
	_defaultConnTimeout = 5 * time.Second
)

type Postgres struct {
	conntimeout time.Duration

	DB *bun.DB
}

func New(url string, ops ...Options) *Postgres {

	pg := &Postgres{
		conntimeout: _defaultConnTimeout,
	}

	for _, op := range ops {
		op(pg)
	}
	driver := pgdriver.NewConnector(
		pgdriver.WithDSN(url),
		pgdriver.WithTimeout(pg.conntimeout),
	)
	sqldb := sql.OpenDB(driver)
	db := bun.NewDB(sqldb, pgdialect.New())

	pg.DB = db

	return pg
}

func (pg *Postgres) Close() {
	pg.DB.Close()
}
