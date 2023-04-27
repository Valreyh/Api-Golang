package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gocql/gocql"

	_ "github.com/cockroachdb/cockroach-go/v2/crdb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func db_mongodb() *mongo.Client {
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

func db_scylladb() *gocql.Session {

	fmt.Println("Attente du démarrage de ScyllaDB...")
	time.Sleep(25 * time.Second)

	// Configuration de la connexion ScyllaDB
	cluster := gocql.NewCluster("scylla") // Adresse IP ou nom d'hôte de votre cluster ScyllaDB7
	cluster.Keyspace = "catalog"          // Nom du keyspace

	session, err := cluster.CreateSession()

	if err != nil {
		log.Fatal("ERREUR : Impossible de se connecter à ScyllaDB ", err)
		return nil
	}

	fmt.Println("Connecté à ScyllaDB !")

	// Création du keyspace 'catalog'
	err = session.Query(`
		CREATE KEYSPACE IF NOT EXISTS catalog
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
	`).Exec()
	if err != nil {
		fmt.Println("Erreur lors de la création du keyspace :", err)
		return nil
	}
	fmt.Println("Keyspace 'catalog' créé avec succès!")

	// Clear la table 'users' si elle existe
	err = session.Query(`DROP TABLE IF EXISTS catalog.users`).Exec()
	if err != nil {
		fmt.Println("Erreur lors du drop de la table 'users' :", err)
		return nil
	}

	// Requête de création de table si elle n'existe pas
	createTableQuery := `CREATE TABLE IF NOT EXISTS catalog.users (
		email TEXT PRIMARY KEY,
		password TEXT,
		picture VARCHAR,
		state BOOLEAN,
		userType INT
	)`

	if err := session.Query(createTableQuery).Exec(); err != nil {
		fmt.Printf("Erreur lors de la création de la table catalog.users : %v", err)
		return nil
	}

	return session
}

func db_cockroach() *gorm.DB {
	// Chaîne de connexion pour CockroachDB
	connectionString := "postgres://root:toto@cockroach:26257/catalog?sslmode=disable"

	// Configurer les options de connexion
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	}

	fmt.Println("Attente du démarrage de Cockroach...")
	time.Sleep(27 * time.Second)

	// Ouvrir une connexion à CockroachDB en utilisant GORM
	db, err := gorm.Open(postgres.Open(connectionString), gormConfig)
	if err != nil {
		log.Fatalf("Impossible de se connecter à CockroachDB: %s", err.Error())
		return nil
	}

	// Vérifier si la connexion est établie en exécutant une requête SQL simple
	var version string
	err = db.Raw("SELECT version()").Scan(&version).Error
	if err != nil {
		log.Fatalf("Impossible de se connecter à CockroachDB: %s", err.Error())
		return nil
	}

	// Creer la database 'catalog' si elle n'existe pas
	err = db.Exec("CREATE DATABASE IF NOT EXISTS catalog").Error
	if err != nil {
		log.Fatalf("Impossible de créer la database 'catalog': %s", err.Error())
		return nil
	}

	// Utiliser la base de données "catalog"
	err = db.Exec("USE catalog").Error
	if err != nil {
		log.Fatalf("Impossible d'utiliser la base de données 'catalog': %s", err.Error())
		return nil
	}

	fmt.Println("Connexion réussie à CockroachDB")

	return db
}
