package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ZnNr/Go-GopherMart.git/internal/database"
	"github.com/ZnNr/Go-GopherMart.git/internal/flags"
	"github.com/ZnNr/Go-GopherMart.git/internal/logger"
	"github.com/ZnNr/Go-GopherMart.git/internal/loyalty"
	"github.com/ZnNr/Go-GopherMart.git/internal/router"
	runner2 "github.com/ZnNr/Go-GopherMart.git/internal/runner"
	"github.com/ZnNr/Go-GopherMart.git/internal/server"
	_ "github.com/jackc/pgx/v5/stdlib"
	"os"
)

const logLevel = "info"

func main() {
	ctx := context.Background()
	log, err := logger.New(logLevel)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	// Инициализируем флаги приложения
	params := flags.Init(
		flags.WithAddr(),
		flags.WithDatabase(),
		flags.WithAccrual(),
	)
	// Открываем соединение с базой данных
	db, err := sql.Open("pgx", params.Database.ConnectionString)
	if err != nil {
		log.Sugar().Errorf("error while init db: %s", err.Error())
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Sugar().Errorf("error while closing db: %s", err.Error())
			os.Exit(1)
		}
	}()
	// Инициализируем менеджер базы данных
	dbManager, err := database.New(ctx, db)
	if err != nil {
		log.Sugar().Errorf("error while init db: %s", err.Error())
		os.Exit(1)
	}
	// Создаем экземпляр сервера приложения
	appServer := server.New(params.Server.Address, router.SetupRouter(dbManager, log.Sugar()))
	// Создаем экземпляр системы начисления бонусных баллов
	loyaltyPointsSystem := loyalty.New(params.AccrualSystem.Address, dbManager, log.Sugar())
	// Создаем экземпляр runner и запускаем приложение
	runner := runner2.New(appServer, loyaltyPointsSystem, log.Sugar())
	if err = runner.Run(ctx); err != nil {
		log.Sugar().Errorf("error while running runner: %s", err.Error())
		return
	}
}
