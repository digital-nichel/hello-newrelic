package main

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/newrelic/go-agent/v3/integrations/logcontext-v2/logWriter"
	"github.com/newrelic/go-agent/v3/integrations/nrecho-v4"
	_ "github.com/newrelic/go-agent/v3/integrations/nrpq"
	"github.com/newrelic/go-agent/v3/newrelic"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"os"
)

type Account struct {
	ID    string `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

var (
	db     *sqlx.DB
	nrApp  *newrelic.Application
	logger *log.Logger
)

func main() {
	db, err := sqlx.Open("nrpostgres", "user=postgres password=postgres dbname=postgres search_path=nichel sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	nrApp, err := newrelic.NewApplication(
		newrelic.ConfigAppName("hello-newrelic"),
		newrelic.ConfigLicense("eu01xxcfea77eb4f2c2efc0bf9023f5e091aNRAL"),
		newrelic.ConfigAppLogEnabled(true),
		newrelic.ConfigAppLogForwardingEnabled(true),
	)
	if err != nil {
		log.Fatal(err)
	}

	writer := logWriter.New(os.Stdout, nrApp)
	logger = log.New(&writer, "", log.Default().Flags())

	e := echo.New()
	e.Use(nrecho.Middleware(nrApp))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Printf("request: [%s] %s -> %d", c.Request().Method, c.Request().URL.Path, c.Response().Status)
			return nil
		},
	}))

	e.GET("/hello", func(c echo.Context) error {
		txn := nrecho.FromContext(c)
		ctx := newrelic.NewContext(context.Background(), txn)

		var accounts []Account
		err := db.SelectContext(ctx, &accounts, "SELECT * FROM accounts")
		if err != nil {
			log.Fatal(err)
		}

		return c.String(http.StatusOK, "Hello, Echo v4!")
	})

	e.Logger.Fatal(e.Start(":1323"))
}
