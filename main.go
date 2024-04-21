package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/openmymai/assessment-tax/postgres"
	"github.com/openmymai/assessment-tax/tax"
)

type application struct {
	auth struct {
		username string
		password string
	}
}

func main() {
	p, err := postgres.New()
	if err != nil {
		panic(err)
	}

	app := new(application)
	app.auth.username = os.Getenv("ADMIN_USERNAME")
	app.auth.password = os.Getenv("ADMIN_PASSWORD")

	if app.auth.username == "" {
		log.Fatal("Basic auth username must be provided")
	}

	if app.auth.password == "" {
		log.Fatal("Basic auth password must be provided")
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	handler := tax.New(p)
	v1 := e.Group("/api/v1")
	{
		v1.POST("/tax/calculations", handler.TaxCalculationsHandler)
		v1.POST("/tax/calculations/upload-csv", handler.TaxCalculationsCSVHandler)
	}

	admin := e.Group("/admin")

	admin.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if app.auth.username != "" && app.auth.password != "" {
			username = app.auth.username
			password = app.auth.password
		}

		if username == "adminTax" && password == "admin!" {

			return true, nil
		}

		return false, nil
	}))
	{
		admin.POST("/deductions/personal", handler.SetPersonalDeductionHandler)
	}

	go func() {
		if err := e.Start(os.Getenv("PORT")); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	fmt.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
