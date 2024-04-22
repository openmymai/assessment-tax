package main

import (
	"context"
	"fmt"
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

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	handler := tax.New(p)
	v1 := e.Group("/tax")
	{
		v1.POST("/calculations", handler.TaxCalculationsHandler)
		v1.POST("/calculations/upload-csv", handler.TaxCalculationsCSVHandler)
	}

	admin := e.Group("/admin")
	admin.Use(middleware.BasicAuth(func(username string, password string, c echo.Context) (bool, error) {
		if username == "adminTax" && password == "admin!" {
			return true, nil
		}

		return false, nil
	}))
	{
		admin.POST("/deductions/personal", handler.SetPersonalDeductionHandler)
		admin.POST("/deductions/k-receipt", handler.SetKreceiptDeductionHandler)
	}

	go func() {
		if err := e.Start(":" + os.Getenv("PORT")); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	fmt.Println("shutting down the server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
