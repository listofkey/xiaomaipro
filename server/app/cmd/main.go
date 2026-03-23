package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
)

const defaultPostgresDSN = "host=127.0.0.1 user=root password=123456 dbname=xiaomaipro port=5432 sslmode=disable TimeZone=Asia/Shanghai"

func main() {
	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		dsn = defaultPostgresDSN
	}

	gormdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{PrepareStmt: true})
	if err != nil {
		log.Fatalf("failed to connect postgres: %v\n", err)
	}

	g := gen.NewGenerator(gen.Config{
		OutPath:      "../rpc/dao",
		ModelPkgPath: "./model",
		Mode:         gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	g.UseDB(gormdb)

	fmt.Println("starting gorm/gen code generation...")
	g.ApplyBasic(g.GenerateAllTable()...)
	g.Execute()
	fmt.Println("code generation completed")
}
