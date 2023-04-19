package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func db() *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*1000)
	defer cancel()

	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")

	// Connexoon à MongoDB
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal("ERREUR : Impossible de se connecter à MongoDB", err)
	}

	// Vérification de la connexion
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal("ERREUR : Impossible de ping MongoDB", err)
	}

	fmt.Println("Connecté à MongoDB !")

	return client
}
