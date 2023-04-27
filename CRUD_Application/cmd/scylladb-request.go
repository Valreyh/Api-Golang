package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx"
	"github.com/scylladb/gocqlx/qb"
	"github.com/scylladb/gocqlx/table"

	"golang.org/x/crypto/bcrypt"
)

// Définition d'un nouveau type pour représenter l'image sous forme de données binaires
type ImageBinaryScylla struct {
	Data      []byte `db:"data" json:"data"`
	Extension string `db:"extension" json:"extension"`
}

// MarshalCQL implémente la méthode de marshall pour la structure ImageBinaryScylla
func (ib ImageBinaryScylla) MarshalCQL(info gocql.TypeInfo) ([]byte, error) {
	// Utiliser un type intermédiaire pour la sérialisation
	type ImageBinaryScyllaJSON struct {
		Extension string `json:"extension"`
		Data      []byte `json:"data"`
	}
	ibJSON := ImageBinaryScyllaJSON{
		Data:      ib.Data,
		Extension: ib.Extension,
	}
	return json.Marshal(ibJSON)
}

func (ib *ImageBinaryScylla) UnmarshalCQL(info gocql.TypeInfo, data []byte) error {
	type ImageBinaryScyllaJSON struct {
		Extension string `json:"extension"`
		Data      []byte `json:"data"`
	}

	var ibJSON ImageBinaryScyllaJSON
	err := json.Unmarshal(data, &ibJSON)
	if err != nil {
		return err
	}

	ib.Extension = ibJSON.Extension
	ib.Data = ibJSON.Data

	return nil
}

type query struct {
	stmt  string
	names []string
}

type statements struct {
	del query
	ins query
	sel query
	upd query
}

type Record struct {
	Email    string             `db:"email"`
	Password string             `db:"password"`
	Picture  *ImageBinaryScylla `db:"picture" json:"picture"`
	State    bool               `db:"state"`
	UserType int                `db:"usertype"`
}

var stmts = createStatements()

func createStatements() *statements {
	m := table.Metadata{
		Name:    "users",
		Columns: []string{"email", "password", "picture", "state", "usertype"},
		PartKey: []string{"email"},
	}
	tbl := table.New(m)
	deleteStmt, deleteUser := tbl.Delete()
	insertStmt, insertUser := tbl.Insert()
	updateStmt, updateUser := tbl.Update(m.Columns[2], m.Columns[3])
	// Normally a select statement such as this would use `tbl.Select()` to select by
	// primary key but now we just want to display all the records...
	selectStmt, selectUser := qb.Select(m.Name).Columns(m.Columns...).ToCql()

	return &statements{
		del: query{
			stmt:  deleteStmt,
			names: deleteUser,
		},
		ins: query{
			stmt:  insertStmt,
			names: insertUser,
		},
		sel: query{
			stmt:  selectStmt,
			names: selectUser,
		},
		upd: query{
			stmt:  updateStmt,
			names: updateUser,
		},
	}
}

var session = db_scylladb()

func DeleteProfileScylla(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete Profile Scylla")

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

	record := Record{
		Email: requestData.Email,
	}

	err = gocqlx.Query(session.Query(stmts.del.stmt), stmts.del.names).BindStruct(record).ExecRelease()
	if err != nil {
		fmt.Println("delete catalog.users", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Message": "Profile supprimé"}`))
}

func CreateProfileScylla(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Inserting Profile Scylla")

	w.Header().Set("Content-Type", "application/json") // on définit le type de contenu de la réponse

	// Lire le corps de la requête pour obtenir les données
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// Gérer l'erreur de lecture du corps de la requête
		http.Error(w, "ERREUR : erreur lors de la lecture de la requête", http.StatusBadRequest)
		return
	}

	// Définir une struct pour extraire les données du corps de la requête
	type RequestData struct {
		Email    string             `json:"email"`
		Password string             `json:"password"`
		Picture  *ImageBinaryScylla `json:"picture,omitempty"`
		State    bool               `json:"state"`
		UserType int                `json:"usertype"`
	}

	// Décoder le corps de la requête dans la struct RequestData
	var requestData RequestData
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		// Gérer l'erreur de décodage du corps de la requête
		http.Error(w, "ERREUR : erreur lors du décodage de la requête", http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	// Hash du mot de passe
	hash, err := bcrypt.GenerateFromPassword([]byte(requestData.Password), 10)
	if err != nil {
		http.Error(w, "ERREUR : erreur lors du hashage du mot de passe", http.StatusBadRequest)
		return
	}

	requestData.Password = string(hash)

	record := Record{
		Email:    requestData.Email,
		Password: requestData.Password,
		State:    requestData.State,
		UserType: requestData.UserType,
	}

	if requestData.Picture != nil {
		record.Picture = requestData.Picture
	}

	err = gocqlx.Query(session.Query(stmts.ins.stmt),
		stmts.ins.names).BindStruct(record).ExecRelease()
	if err != nil {
		fmt.Println("insert catalog.users", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Message": "Profil créé avec succès !"}`))
}

func GetUserProfileScylla(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Affichage du profil utilisateur")

	w.Header().Set("Content-Type", "application/json") // on définit le type de contenu de la réponse

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

	// Utiliser les données récupérées dans la struct pour effectuer votre recherche dans la base de données
	var record Record
	err = gocqlx.Query(session.Query(stmts.sel.stmt), stmts.sel.names).BindMap(qb.M{
		"email": requestData.Email, // Utiliser l'e-mail récupéré dans la requête
	}).GetRelease(&record)
	if err != nil {
		fmt.Println("select catalog.users", err)
		return
	}

	// Vérifier si l'e-mail de l'enregistrement correspond à l'e-mail recherché dans la requête
	if record.Email == requestData.Email {
		// Encoder les données dans un format JSON
		jsonBytes, err := json.Marshal(record)
		if err != nil {
			http.Error(w, "ERREUR : erreur lors de l'encodage de la réponse", http.StatusInternalServerError)
			return
		}

		// Écrire les données encodées dans la réponse HTTP
		w.Write(jsonBytes)
	} else {
		fmt.Println("ERREUR : aucune données pour cette email" + requestData.Email)
	}
}

func GetAllUsersScylla(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Displaying Results:")

	w.Header().Set("Content-Type", "application/json") // on définit le type de contenu de la réponse

	// Utiliser les données récupérées dans la struct pour effectuer votre recherche dans la base de données
	var records []Record // Utiliser un slice de Record pour stocker les enregistrements
	err := gocqlx.Query(session.Query(stmts.sel.stmt), stmts.sel.names).SelectRelease(&records)
	if err != nil {
		fmt.Println("select catalog.users", err)
		http.Error(w, "ERREUR : erreur lors de la recherche des profils", http.StatusInternalServerError)
		return
	}

	// Convertir les valeurs VARCHAR pour le champ `picture` en *main.ImageBinaryScylla
	for i := range records {
		if records[i].Picture == nil {
			records[i].Picture = &ImageBinaryScylla{}
		}
	}

	// Encoder les données dans un format JSON
	jsonBytes, err := json.Marshal(records)
	if err != nil {
		fmt.Println("ERREUR : erreur lors de l'encodage de la réponse", err)
		http.Error(w, "ERREUR : erreur lors de l'encodage de la réponse", http.StatusInternalServerError)
		return
	}

	// Écrire les données encodées dans la réponse HTTP
	w.Write(jsonBytes)
}

func getAllUsersTypeScylla(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json") // on définit le type de contenu de la réponse

	// Lire le corps de la requête pour obtenir les données
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// Gérer l'erreur de lecture du corps de la requête
		http.Error(w, "ERREUR : erreur lors de la lecture de la requête", http.StatusBadRequest)
		return
	}

	// Définir une struct pour extraire les données du corps de la requête
	type RequestData struct {
		UserType int `json:"usertype"`
	}

	// Décoder le corps de la requête dans la struct RequestData
	var requestData RequestData
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		// Gérer l'erreur de décodage du corps de la requête
		http.Error(w, "ERREUR : erreur lors du décodage de la requête", http.StatusBadRequest)
		return
	}

	var records []Record // Utiliser un slice de Record pour stocker les enregistrements
	err = gocqlx.Query(session.Query(stmts.sel.stmt), stmts.sel.names).BindMap(qb.M{
		"usertype": requestData.UserType,
	}).SelectRelease(&records)
	if err != nil {
		fmt.Println("select catalog.users usertype : ", err)
		http.Error(w, "ERREUR : erreur lors de la recherche des profils", http.StatusInternalServerError)
		return
	}

	var filteredRecords []Record
	for _, r := range records {
		if r.UserType == requestData.UserType {
			filteredRecords = append(filteredRecords, r)
		}
	}

	// Convertir les valeurs VARCHAR pour le champ `picture` en *main.ImageBinaryScylla
	for i := range filteredRecords {
		if filteredRecords[i].Picture == nil {
			filteredRecords[i].Picture = &ImageBinaryScylla{}
		}
	}

	// Encoder les données dans un format JSON
	jsonBytes, err := json.Marshal(filteredRecords)
	if err != nil {
		fmt.Println("ERREUR : erreur lors de l'encodage de la réponse", err)
		http.Error(w, "ERREUR : erreur lors de l'encodage de la réponse", http.StatusInternalServerError)
		return
	}

	// Écrire les données encodées dans la réponse HTTP
	w.Write(jsonBytes)
}

func UpdateProfileScylla(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Updating Profile Scylla")

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
		State bool   `json:"state"`
	}

	// Décoder le corps de la requête dans la struct RequestData
	var requestData RequestData
	err = json.Unmarshal(body, &requestData)
	if err != nil {
		// Gérer l'erreur de décodage du corps de la requête
		http.Error(w, "ERREUR : erreur lors du décodage de la requête", http.StatusBadRequest)
		return
	}

	record := Record{
		Email: requestData.Email,
		State: requestData.State,
	}

	err = gocqlx.Query(session.Query(stmts.upd.stmt),
		stmts.upd.names).BindStruct(record).ExecRelease()
	if err != nil {
		fmt.Println("update catalog.users", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Message": "Profil mis à jour avec succès !"}`))
}

func CreateHTLMPageScylla(w http.ResponseWriter, r *http.Request) {

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

	record := Record{
		Email: requestData.Email,
	}

	// Utiliser les données récupérées dans la struct pour effectuer votre recherche dans la base de données
	var userScylla []Record // Modifier ici pour déclarer une slice de Record
	err = gocqlx.Query(session.Query(stmts.sel.stmt), stmts.sel.names).BindStruct(record).SelectRelease(&userScylla)
	if err != nil {
		fmt.Println("select catalog.users", err)
		http.Error(w, "ERREUR : erreur lors de la recherche des profils", http.StatusInternalServerError)
		return
	}

	// Vérifier si le résultat de la requête est vide
	if userScylla[0].Email == "" {
		// Gérer le cas où aucun profil n'a été trouvé
		http.Error(w, "ERREUR : aucun profil trouvé", http.StatusNotFound)
		return
	}

	// On crée le fichier HTML
	file, err := os.Create("./html_pages/" + userScylla[0].Email + ".html")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// On écrit le contenu du fichier HTML
	_, err = io.WriteString(file, "<html><head><title>Page de profil</title></head><body><h1>Page de profil</h1><p>Email : "+userScylla[0].Email+"</p><p>Etat : "+fmt.Sprint(userScylla[0].State)+"</p><p>Type d'utilisateur : "+fmt.Sprint(userScylla[0].UserType)+"</p><img src='../images/"+userScylla[0].Email+userScylla[0].Picture.Extension+"' /></body></html>")
	if err != nil {
		log.Fatal(err)
	}

	// On met un message de succès
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Message": "Fichier HTML créer avec succès, dans le répertoire html_pages"}`))

}

func UploadProfileImageScylla(w http.ResponseWriter, r *http.Request) {

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

	// Lire les données binaires du fichier
	var pictureData bytes.Buffer
	_, err = io.Copy(&pictureData, file)
	if err != nil {
		fmt.Println("ERREUR : impossible de lire les données binaires de l'image", err)
		http.Error(w, "ERREUR : impossible de lire les données binaires de l'image", http.StatusInternalServerError)
		return
	}

	fileName := handler.Filename
	fileExtension := filepath.Ext(fileName)

	fmt.Println("Extension  : ", fileExtension)

	// On récupère l'email de l'utilisateur
	email := r.FormValue("email")

	fmt.Println("Email : ", email)

	// Charger l'utilisateur existant depuis la base de données
	existingRecord := &Record{}
	err = gocqlx.Query(session.Query(stmts.sel.stmt), stmts.sel.names).BindMap(qb.M{
		"email": email,
	}).GetRelease(existingRecord)
	if err != nil {
		fmt.Println("ERREUR : impossible de charger l'utilisateur depuis la base de données", err)
		http.Error(w, "ERREUR : impossible de charger l'utilisateur depuis la base de données", http.StatusInternalServerError)
		return
	}

	// Initialize existingRecord.Picture if it's nil
	if existingRecord.Picture == nil {
		existingRecord.Picture = &ImageBinaryScylla{}
	}

	// Mettre à jour les données de l'image dans l'utilisateur existant
	existingRecord.Picture.Data = pictureData.Bytes()
	existingRecord.Picture.Extension = fileExtension

	// On créer un record pour mettre à jour l'image dans la base de données
	newRecord := &Record{
		Email:    email,
		Password: existingRecord.Password,
		Picture:  existingRecord.Picture,
		State:    existingRecord.State,
		UserType: existingRecord.UserType,
	}

	// Mettre à jour l'image dans la base de données de l'utilisateur
	err = gocqlx.Query(session.Query(stmts.upd.stmt), stmts.upd.names).BindStruct(newRecord).ExecRelease()
	if err != nil {
		fmt.Println("ERREUR : impossible de mettre à jour l'image dans la base de données", err)
		http.Error(w, "ERREUR : impossible de mettre à jour l'image dans la base de données", http.StatusInternalServerError)
		return
	}

	fmt.Println("Image du profil mise à jour avec succès")
	w.Write([]byte("Image du profil mise à jour avec succès"))
}

func GetProfileImageScylla(w http.ResponseWriter, r *http.Request) {

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

	// On récupère l'email de l'utilisateur
	email := requestData.Email

	// On récupère l'image de profil de l'utilisateur dans la base de données
	stmt := fmt.Sprintf(`SELECT picture FROM catalog.users WHERE email = ?`)

	var imageBinary ImageBinaryScylla
	err = session.Query(stmt, email).Scan(&imageBinary)
	if err != nil {
		fmt.Println("Erreur lors de la récupération de l'image de profil :", err)
		return
	}

	// On convertit les données de l'image de profil en bytes

	// On crée le fichier image
	fileExtension := imageBinary.Extension
	fileName := email + fileExtension

	filePath := path.Join("./images", fileName) // Chemin du fichier dans l'arborescence du projet
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

	// On écrit les bytes de l'image dans le fichier
	_, err = file.Write(imageBinary.Data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Erreur": "Impossible d'écrire les données d'image dans le fichier"}`))
		fmt.Println(err)
		return
	}

	// On envoie un message de succès
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Message": "Image envoyée"}`))
}

func DeleteAllDatabaseScylla(w http.ResponseWriter, r *http.Request) {

	// Exécuter la requête CQL de suppression de tous les enregistrements dans la table
	query := "TRUNCATE catalog.users"
	if err := session.Query(query).Exec(); err != nil {
		fmt.Println("Erreur lors de la suppression de tous les enregistrements :", err)
	}

	// On envoie un message de succès
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Message": "Tous les enregistrements ont été supprimés"}`))
}
