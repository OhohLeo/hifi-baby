package main

import (
	"log"

	"github.com/OhohLeo/hifi-baby/app"
)

func main() {
	cfg, err := app.NewConfig()
	if err != nil {
		log.Fatalf("Erreur lors de l'initialisation de la configuration : %v", err)
	}

	app, errApp := app.NewApp(cfg)
	if errApp != nil {
		log.Fatalf("Erreur lors de l'initialisation de l'application : %v", errApp)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("Erreur lors de l'exécution de l'application : %v", err)
	}
}
