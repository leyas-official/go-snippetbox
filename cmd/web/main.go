package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type application struct {
	errorLog *log.Logger
	infoLog  *log.Logger
}

func main() {
	godotenv.Load() // load .env into os.Environ
	addr := os.Getenv("SNIPPETBOX_ADDR")
	if addr == "" {
		addr = ":4000" // fallback
	}

	f, err := os.OpenFile("./tmp/info.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	infoLog := log.New(f, "INFO\t", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog := log.New(f, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
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
