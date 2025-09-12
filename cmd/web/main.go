package main

import (
	"database/sql"
	"flag"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"snippetbox.leyasofficial.net/internal/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	snippets      *models.SnippetModel
	templateCache map[string]*template.Template
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	godotenv.Load()

	// port setup
	addr := os.Getenv("SNIPPETBOX_ADDR")
	if addr == "" {
		addr = ":4000"
	}

	dsnEnv := os.Getenv("DSN")
	dsnFlag := flag.String("dsn", dsnEnv, "MySQL data source name")

	flag.Parse()

	dsn := *dsnFlag

	// log file creation and initialization
	f, err := os.OpenFile("./tmp/info.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	infoWriter := io.MultiWriter(os.Stdout, f)
	errorWriter := io.MultiWriter(os.Stderr, f)

	infoLog := log.New(infoWriter, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog := log.New(errorWriter, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Database connection
	db, err := openDB(dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	// Cache on server initialization
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}

	// app dependencies
	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		snippets:      &models.SnippetModel{DB: db},
		templateCache: templateCache,
	}

	serve := &http.Server{
		Addr:     addr,
		Handler:  app.routes(),
		ErrorLog: errorLog,
	}

	infoLog.Printf("Starting server on %s", addr)
	err = serve.ListenAndServe()

	errorLog.Fatal(err)
}
