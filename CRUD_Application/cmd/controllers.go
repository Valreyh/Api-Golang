package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// Struct des user
type user struct {
	Email    string      `json:"email"`
	Password string      `json:"password"`
	Picture  ImageBinary `json:"picture"`
	State    bool        `json:"state"`
	UserType int         `json:"userType"`
}

// Définition d'un nouveau type pour représenter l'image sous forme de données binaires
type ImageBinary struct {
	Data      []byte           `bson:"data"`      // les données binaires de l'image
	Type      primitive.Binary `bson:"type"`      // le type de données de l'image
	Extension string           `bson:"extension"` // l'extension de l'image
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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Erreur": "Email déjà utilisé"}`))
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

	fmt.Println("Création du profile : ", insertResult)
	json.NewEncoder(w).Encode(insertResult.InsertedID) // on renvoie l'id du document créé (on peut envoyé autre chose si besoin)

}

func CreateHTLMPage(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var body user
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		fmt.Println(err)
	}

	// On récupère les informations de l'utilisateur et on crée la page HTML
	var user user
	err = userCollection.FindOne(context.TODO(), bson.D{{Key: "email", Value: body.Email}}).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Erreur": "Email non trouvé"}`))
		return
	}

	// On crée le fichier HTML
	file, err := os.Create("./html_pages/" + user.Email + ".html")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Convertir le tableau d'octets en une chaîne de caractères base64
	pictureBase64 := base64.StdEncoding.EncodeToString(user.Picture.Data)

	// On détermine le type d'image en fonction de l'extension de fichier
	var imgType string
	switch user.Picture.Extension {
	case "jpg", "jpeg":
		imgType = "image/jpeg"
	case "png":
		imgType = "image/png"
	}

	// On écrit le contenu du fichier HTML en utilisant le type d'image déterminé
	_, err = io.WriteString(file, "<html><head><title>Page de profil</title></head><body><h1>Page de profil</h1><p>Email : "+user.Email+"</p><p>Etat : "+fmt.Sprint(user.State)+"</p><p>Type d'utilisateur : "+fmt.Sprint(user.UserType)+"</p><img src='data:"+imgType+";base64,"+pictureBase64+"' /></body></html>")
	if err != nil {
		log.Fatal(err)
	}

	// On met un message de succès
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Message": "Fichier HTML créer avec succès, dans le répertoir html_pages"}`))

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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Erreur": "Email non trouvé"}`))
		return
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

	// Si l'email n'existe pas, on renvoie une erreur
	var resultEmail primitive.M
	err := userCollection.FindOne(context.TODO(), bson.D{{Key: "email", Value: body.Email}}).Decode(&resultEmail)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Erreur": "Email non trouvé"}`))
		return
	}

	// Si l'état est autre que true ou false, on renvoie une erreur
	if body.State != true && body.State != false {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Erreur": "Etat non valide"}`))
		return
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

func UploadProfileImage(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "multipart/form-data")

	// Parse le corps de la requête pour récupérer le formulaire multipart
	err := r.ParseMultipartForm(16 << 20) // taille maximale du fichier : 16 Mo
	if err != nil {
		http.Error(w, "Erreur lors de la lecture du formulaire", http.StatusBadRequest)
		return
	}

	// On lit le fichier image envoyé
	file, handler, err := r.FormFile("image")
	if err != nil {
		fmt.Println("Erreur : recupération du fichier impossible")
		fmt.Println(err)
		return
	}
	defer file.Close()

	// On vérifie que le fichier est bien une image
	if handler.Header.Get("Content-Type") != "image/jpeg" && handler.Header.Get("Content-Type") != "image/png" && handler.Header.Get("Content-Type") != "image/jpg" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Erreur": "Le fichier n'est pas une image"}`))
		return
	}

	// Lire les bytes de l'image
	var imageBytes []byte
	buf := make([]byte, 1024) // créer un tampon de 1 Ko pour lire les données du fichier par morceaux
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println("Erreur : lecture des bytes de l'image impossible")
			fmt.Println(err)
			return
		}
		if n == 0 {
			break
		}
		imageBytes = append(imageBytes, buf[:n]...)
	}

	fileName := handler.Filename
	fileExtension := filepath.Ext(fileName)

	fmt.Println("Extension  : ", fileExtension)

	// Créer un nouveau document ImageBinary avec les données de l'image
	imageBinary := ImageBinary{
		Data:      imageBytes,
		Extension: fileExtension,
		Type: primitive.Binary{
			Subtype: 0x00,
			Data:    imageBytes,
		},
	}

	fmt.Println("L'extension de l'image qu'on veut enregistrer est : ", imageBinary.Extension)

	// On récupère l'email de l'utilisateur
	email := r.FormValue("email")
	// Si l'email n'existe pas dans userCollection, on renvoie une erreur
	var resultEmail primitive.M
	err = userCollection.FindOne(context.TODO(), bson.D{{Key: "email", Value: email}}).Decode(&resultEmail)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Erreur": "Email non trouvé"}`))
		return
	}

	// On met à jour l'image de l'utilisateur
	filter := bson.D{{Key: "email", Value: email}} // on filtre sur l'email pour trouver l'utilisateur à modifier
	after := options.After                         // on veut que le document soit retourné après la modification
	returnOpt := options.FindOneAndUpdateOptions{

		ReturnDocument: &after,
	}

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "picture", Value: imageBinary}}}} // on met à jour l'image de l'utilisateur
	updateResult := userCollection.FindOneAndUpdate(context.TODO(), filter, update, &returnOpt)

	var result primitive.M
	_ = updateResult.Decode(&result)

	// on envoie un message de succès
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Message": "Image envoyée"}`))
}

func GetProfileImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	type response struct {
		Email string `json:"email"`
	}

	// Récupérer l'email de l'utilisateur
	var body response
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Erreur": "Email non trouvé"}`))
		return
	}

	email := body.Email

	// Rechercher l'utilisateur dans la base de données
	var user user
	filter := bson.D{{Key: "email", Value: email}}
	err = userCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Erreur": "Utilisateur non trouvé"}`))
		return
	}

	// Récupérer les données d'image
	imageBinary := user.Picture
	imageBytes := imageBinary.Data

	// Récupérer l'extension du fichier
	fileExtension := imageBinary.Extension
	fmt.Println("Extension du fichier image à charger :", fileExtension)

	// Créer le nom du fichier
	fileName := fmt.Sprintf("%s-%d.%s", strings.Replace(email, "@", "_", -1), time.Now().Unix(), fileExtension)
	cleanFileName := filepath.Clean(fileName)

	// Créer le fichier dans l'arborescence du projet
	filePath := path.Join("./images", cleanFileName) // Chemin du fichier dans l'arborescence du projet
	fmt.Println("Chemin du dossier images :", path.Dir(filePath))

	// Créer le fichier
	file, err := os.Create(filePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Erreur": "Impossible de créer le fichier d'image"}`))
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Écrire les données d'image dans le fichier
	_, err = file.Write(imageBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Erreur": "Impossible d'écrire les données d'image dans le fichier"}`))
		fmt.Println(err)
		return
	}

	// On met un message de succès
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Message": "Image créée"}`))
}

// Suppression d'un utilisateur

func DeleteProfile(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)["id"] // on récupère l'id de l'utilisateur à supprimer dans l'url

	// Si l'id n'existe pas, on renvoie une erreur
	var result primitive.M
	err := userCollection.FindOne(context.TODO(), bson.D{{Key: "_id", Value: params}}).Decode(&result)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Erreur": "Id non trouvé"}`))
		return
	}

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
