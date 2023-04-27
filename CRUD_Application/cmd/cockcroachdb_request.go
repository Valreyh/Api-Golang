package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"gorm.io/gorm"

	"golang.org/x/crypto/bcrypt"
)

type UserCockroach struct {
	Email    string                `gorm:"type:VARCHAR(255);primaryKey" json:"email"`
	Password string                `gorm:"type:VARCHAR(255);not null" json:"password"`
	Picture  *ImageBinaryCockroach `json:"picture"`
	State    bool                  `gorm:"type:BOOLEAN;default:true" json:"state"`
	UserType int                   `gorm:"type:INTEGER;default:1" json:"userType"`
}

// Définition d'un nouveau type pour représenter l'image sous forme de données binaires
type ImageBinaryCockroach struct {
	Data          []byte `gorm:"type:BYTEA" json:"data"`                           // les données binaires de l'image
	FileExtension string `gorm:"type:VARCHAR(255);not null" json:"file_extension"` // l'extension de l'image
}

// Implémentation de l'interface Scanner pour la structure ImageBinaryCockroach
func (i *ImageBinaryCockroach) Scan(value interface{}) error {
	// Vérifier si la valeur est nil
	if value == nil {
		return nil
	}

	// Convertir la valeur en []byte
	data, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan ImageBinaryCockroach: value is not []byte")
	}

	// Copier les données vers le champ Data de la structure
	i.Data = make([]byte, len(data))
	copy(i.Data, data)

	return nil
}

// Implémentation de l'interface Valuer pour la structure ImageBinaryCockroach
func (i ImageBinaryCockroach) Value() (driver.Value, error) {
	return i.Data, nil
}

var db = db_cockroach()

// Création d'un utilisateur

func CreateProfileCockroach(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// Vérifier si la table "user_cockroaches" existe déjà
	if !db.Migrator().HasTable(&UserCockroach{}) {
		// AutoMigrate pour créer la table "user_cockroaches" dans la base de données
		err := db.AutoMigrate(&UserCockroach{})
		if err != nil {
			log.Fatalf("Impossible de créer la table 'user_cockroaches': %s", err.Error())
			return
		}

		fmt.Println("Table 'user_cockroaches' créée avec succès")
	} else {
		fmt.Println("La table 'user_cockroaches' existe déjà")
	}

	var person UserCockroach
	err := json.NewDecoder(r.Body).Decode(&person)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Erreur": "Erreur lors de la lecture du corps de la requête"}`))
		return
	}

	// Vérification si l'e-mail est déjà utilisé
	var result UserCockroach
	err = db.Where("email = ?", person.Email).First(&result).Error
	if err == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Erreur": "Email déjà utilisé"}`))
		return
	}

	// Hashage du mot de passe
	hashedPassword, err := hashPassword(person.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Erreur": "Erreur lors du hashage du mot de passe"}`))
		return
	}
	person.Password = hashedPassword

	// Vérification du type d'utilisateur
	if person.UserType != 1 && person.UserType != 2 && person.UserType != 3 {
		person.UserType = 1
	}

	err = db.Create(&person).Error
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Erreur": "Erreur lors de la création du profil"}`))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(person)

}

func CreateHTMLPageCockroach(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	var body UserCockroach
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		fmt.Println(err)
	}

	// On récupère les informations de l'utilisateur et on crée la page HTML
	var user UserCockroach
	err = db.Where("email = ?", body.Email).First(&user).Error
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

	// On écrit le contenu du fichier HTML
	_, err = io.WriteString(file, "<html><head><title>Page de profil</title></head><body><h1>Page de profil</h1><p>Email : "+user.Email+"</p><p>Etat : "+fmt.Sprint(user.State)+"</p><p>Type d'utilisateur : "+fmt.Sprint(user.UserType)+"</p><img src='../images/"+user.Email+user.Picture.FileExtension+"' /></body></html>")
	if err != nil {
		log.Fatal(err)
	}

	// On met un message de succès
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Message": "Fichier HTML créé avec succès, dans le répertoire html_pages"}`))

}

// Récupération d'un utilisateur avec son email

func GetUserProfileCockroach(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// Lire le corps de la requête pour obtenir les données
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// Gérer l'erreur de lecture du corps de la requête
		http.Error(w, "ERREUR : erreur lors de la lecture de la requête", http.StatusBadRequest)
		return
	}

	// Définir une struct pour extraire les données du corps de la requête
	type RequestData struct {
		Email string `json:"email"`
	}

	// Décoder le corps de la requête dans la struct RequestData
	var requestData RequestData
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		// Gérer l'erreur de décodage du corps de la requête
		http.Error(w, "ERREUR : erreur lors du décodage de la requête", http.StatusBadRequest)
		return
	}

	var user UserCockroach
	err = db.Where("email = ?", requestData.Email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"Erreur": "Utilisateur non trouvé"}`))
		} else {
			log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"Erreur": "Erreur lors de la récupération du profil d'utilisateur"}`))
		}
		return
	}

	json.NewEncoder(w).Encode(user)

}

// Update d'un utilisateur sur son état

func UpdateProfileCockroach(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	type updateBody struct {
		Email string `json:"email"` // l'email de l'utilisateur pour le trouver et le modifier
		State bool   `json:"state"` // le nouvel état de l'utilisateur qui sera mis à jour
	}

	var requestData updateBody
	e := json.NewDecoder(r.Body).Decode(&requestData)
	if e != nil {

		fmt.Print(e)
	}

	var user UserCockroach
	err := db.Where("email = ?", requestData.Email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"Erreur": "Utilisateur non trouvé"}`))
		} else {
			log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"Erreur": "Erreur lors de la recherche de l'utilisateur à mettre à jour"}`))
		}
		return
	}

	user.State = requestData.State

	err = db.Save(&user).Error
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Erreur": "Erreur lors de la mise à jour du profil d'utilisateur"}`))
		return
	}

	w.Write([]byte(`{"Message": "Profil d'utilisateur mis à jour avec succès"}`))
}

func UploadProfileImageCockroach(w http.ResponseWriter, r *http.Request) {

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
		fmt.Println("Erreur : récupération du fichier impossible")
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

	// Créer un nouveau modèle UserCockroach avec les données de l'image
	imageBinary := ImageBinaryCockroach{
		Data:          imageBytes,
		FileExtension: fileExtension,
	}

	fmt.Println("L'extension de l'image qu'on veut enregistrer est : ", imageBinary.FileExtension)

	// On récupère l'email de l'utilisateur
	email := r.FormValue("email")
	// Si l'email n'existe pas dans la table user_cockroach, on renvoie une erreur
	var user UserCockroach
	result := db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Erreur": "Email non trouvé"}`))
		return
	}

	// On met à jour l'image de l'utilisateur
	user.Picture = &imageBinary

	result = db.Save(&user)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Erreur": "Échec de la mise à jour de l'image de l'utilisateur"}`))
		return
	}

	// on envoie un message de succès
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Message": "Image envoyée"}`))
}

func GetProfileImageCockroach(w http.ResponseWriter, r *http.Request) {
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
	var userCockroach UserCockroach
	err = db.Where("email = ?", email).First(&userCockroach).Error
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Erreur": "Utilisateur non trouvé"}`))
		return
	}

	// Récupérer les données d'image
	imageBinary := userCockroach.Picture
	imageBytes := imageBinary.Data

	// Récupérer l'extension du fichier
	fileExtension := imageBinary.FileExtension
	fmt.Println("Extension du fichier image à charger :", fileExtension)

	// Créer le nom du fichier
	fileName := userCockroach.Email + fileExtension
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

func DeleteProfileCockroach(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Lire le corps de la requête pour obtenir les données
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "ERREUR : erreur lors de la lecture de la requête", http.StatusBadRequest)
		return
	}

	// Définir une struct pour extraire les données du corps de la requête
	type RequestData struct {
		Email string `json:"email"`
	}

	// Décoder le corps de la requête dans la struct RequestData
	var requestData RequestData
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		http.Error(w, "ERREUR : erreur lors du décodage de la requête", http.StatusBadRequest)
		return
	}

	var user UserCockroach
	err = db.Where("email = ?", requestData.Email).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"Erreur": "Utilisateur non trouvé"}`))
		} else {
			log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"Erreur": "Erreur lors de la recherche de l'utilisateur à supprimer"}`))
		}
		return
	}

	err = db.Delete(&user).Error
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Erreur": "Erreur lors de la suppression de l'utilisateur"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Message": "Utilisateur supprimé"}`))
}

// Récupération de tous les utilisateurs

func GetAllUsersCockroach(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var users []UserCockroach
	err := db.Find(&users).Error
	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Erreur": "Erreur lors de la récupération des utilisateurs"}`))
		return
	}

	json.NewEncoder(w).Encode(users)
}

func getAllUsersTypeCockroach(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// On récupère le UserType à filtrer depuis le corps de la requête
	var requestBody struct {
		UserType int `json:"user_type"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Erreur": "Requête JSON invalide"}`))
		return
	}

	// On récupère tous les utilisateurs avec le UserType spécifié
	var users []UserCockroach
	err = db.Where("user_type = ?", requestBody.UserType).Find(&users).Error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Erreur": "Erreur lors de la récupération des utilisateurs"}`))
		return
	}

	// On retourne les utilisateurs en JSON
	json.NewEncoder(w).Encode(users)
}

func DropTableAndRecreateCockroach(w http.ResponseWriter, r *http.Request) {
	// Supprimer la table "users"
	err := db.Migrator().DropTable("user_cockroaches")
	if err != nil {
		return
	}
	fmt.Println("Table 'users' supprimée avec succès")

	// Recréer la table "users"
	err = db.AutoMigrate(&UserCockroach{})
	if err != nil {
		return
	}
	fmt.Println("Table 'users' recréée avec succès")
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14) // on hash le mot de passe avec bcrypt
	return string(bytes), err
}
