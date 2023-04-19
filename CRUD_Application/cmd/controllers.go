package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// Struct des user
type user struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Picture  string `json:"picture"`
	State    bool   `json:"state"`
	UserType int    `json:"userType"`
}

var userCollection = db().Database("goDatabaseCrud").Collection("users")

// Création d'un utilisateur

func CreateProfile(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json") // on définit le type de contenu de la réponse

	var person user
	err := json.NewDecoder(r.Body).Decode(&person) // On stocke le body de la requête dans la variable person
	if err != nil {
		fmt.Println(err)
	}

	// On vérifie si l'email est déjà utilisé
	var result primitive.M                                                                                    // une représentation non ordonnée d'un document BSON qui est une Map
	err = userCollection.FindOne(context.TODO(), bson.D{{Key: "email", Value: person.Email}}).Decode(&result) // on cherche un document avec l'email donné
	if err == nil {                                                                                           // si on trouve un document, on renvoie une erreur
		fmt.Println("Erreur : l'email est déjà utilisé")
		return
	}

	// On hash le mot de passe avec bcrypt et les fonctions en bas
	hash, _ := hashPassword(person.Password)
	person.Password = hash

	// Vérification de l'usertype si autre que prévu, on le met à 1 par défaut (user)
	if person.UserType != 1 && person.UserType != 2 && person.UserType != 3 {
		person.UserType = 1
	}

	insertResult, err := userCollection.InsertOne(context.Background(), person)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Inserted a single document: ", insertResult)
	json.NewEncoder(w).Encode(insertResult.InsertedID) // on renvoie l'id du document créé (on peut envoyé autre chose si besoin)

}

// Récupération d'un utilisateur avec son email

func GetUserProfile(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var body user
	e := json.NewDecoder(r.Body).Decode(&body)
	if e != nil {

		fmt.Print(e)
	}
	var result primitive.M
	err := userCollection.FindOne(context.TODO(), bson.D{{Key: "email", Value: body.Email}}).Decode(&result)
	if err != nil {

		fmt.Println(err)

	}
	json.NewEncoder(w).Encode(result)

}

// Update d'un utilisateur sur son état

func UpdateProfile(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	type updateBody struct {
		Email string `json:"email"` // l'email de l'utilisateur pour le trouver et le modifier
		State bool   `json:"state"` // le nouvel état de l'utilisateur qui sera mis à jour
	}
	var body updateBody
	e := json.NewDecoder(r.Body).Decode(&body)
	if e != nil {

		fmt.Print(e)
	}
	filter := bson.D{{Key: "email", Value: body.Email}} // on filtre sur l'email pour trouver l'utilisateur à modifier
	after := options.After                              // on veut que le document soit retourné après la modification
	returnOpt := options.FindOneAndUpdateOptions{

		ReturnDocument: &after,
	}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "state", Value: body.State}}}} // on met à jour l'état de l'utilisateur
	updateResult := userCollection.FindOneAndUpdate(context.TODO(), filter, update, &returnOpt)

	var result primitive.M
	_ = updateResult.Decode(&result)

	json.NewEncoder(w).Encode(result)
}

// Suppression d'un utilisateur

func DeleteProfile(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)["id"] // on récupère l'id de l'utilisateur à supprimer dans l'url

	_id, err := primitive.ObjectIDFromHex(params) // on convertit l'id en ObjectID
	if err != nil {
		fmt.Printf(err.Error())
	}
	opts := options.Delete().SetCollation(&options.Collation{})
	res, err := userCollection.DeleteOne(context.TODO(), bson.D{{Key: "_id", Value: _id}}, opts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("deleted %v documents\n", res.DeletedCount)
	json.NewEncoder(w).Encode(res.DeletedCount) // on renvoie le nombre de documents supprimés (1 si tout s'est bien passé)

}

// Récupération de tous les utilisateurs

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var results []primitive.M
	cur, err := userCollection.Find(context.TODO(), bson.D{{}}) // on récupère tous les documents de la collection users
	if err != nil {

		fmt.Println(err)

	}
	for cur.Next(context.TODO()) { // itère sur le curseur jusqu'à ce qu'il n'y ait plus de documents

		var elem primitive.M
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, elem) // on ajoute chaque document à la liste results
	}
	cur.Close(context.TODO()) // on ferme le curseur pour libérer les ressources
	json.NewEncoder(w).Encode(results)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14) // on hash le mot de passe avec bcrypt
	return string(bytes), err
}
