package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"shortly/controllers"
	"shortly/generator"
	"shortly/repository"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type containers struct {
	UrlShorten *controllers.ShortnerController
}

func buildContainer(ctx context.Context) *containers {

	r := repository.Redis{}
	redisClient := r.GetRedisConn()
	slugGen, err := generator.NewSlugGenerator(ctx,
		generator.NewRedisStateStore(redisClient.Client))

	if err != nil {
		log.Panic("Error in slug gen instance init", err)
	}

	return &containers{
		UrlShorten: &controllers.ShortnerController{
			SlugGen: slugGen,
		},
	}

}

func init() {
	LoadEnv()
}

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	c := buildContainer(ctx)
	r := mux.NewRouter()

	r.HandleFunc("/shorten", c.UrlShorten.UrlShortnerCreate)

	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Server started at : %s", port)

	if listenErr := http.ListenAndServe(addr, r); listenErr != nil {
		log.Fatal("Error starting http server", listenErr)
	}
}

func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Panic("Error loading env configs", err)
	}
}
