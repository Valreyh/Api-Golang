package main

import (
	"log"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

////////////////////////
///// PARTIE MONGO /////
////////////////////////

type mongoDB_Interface interface {
	CreateProfileMongo(w http.ResponseWriter, r *http.Request)
	GetAllUsersMongo(w http.ResponseWriter, r *http.Request)
	GetUserProfileMongo(w http.ResponseWriter, r *http.Request)
	UpdateProfileMongo(w http.ResponseWriter, r *http.Request)
	DeleteProfileMongo(w http.ResponseWriter, r *http.Request)
	UploadProfileImageMongo(w http.ResponseWriter, r *http.Request)
	GetProfileImageMongo(w http.ResponseWriter, r *http.Request)
	CreateHTLMPageMongo(w http.ResponseWriter, r *http.Request)
}

type mongodb_struct struct {
	client *mongo.Client
}

func (m *mongodb_struct) CreateProfileMongo(w http.ResponseWriter, r *http.Request) {
	CreateProfileMongo(w, r)
}

func (m *mongodb_struct) GetAllUsersMongo(w http.ResponseWriter, r *http.Request) {
	GetAllUsersMongo(w, r)
}

func (m *mongodb_struct) GetUserProfileMongo(w http.ResponseWriter, r *http.Request) {
	GetUserProfileMongo(w, r)
}

func (m *mongodb_struct) UpdateProfileMongo(w http.ResponseWriter, r *http.Request) {
	UpdateProfileMongo(w, r)
}

func (m *mongodb_struct) DeleteProfileMongo(w http.ResponseWriter, r *http.Request) {
	DeleteProfileMongo(w, r)
}

func (m *mongodb_struct) UploadProfileImageMongo(w http.ResponseWriter, r *http.Request) {
	UploadProfileImageMongo(w, r)
}

func (m *mongodb_struct) GetProfileImageMongo(w http.ResponseWriter, r *http.Request) {
	GetProfileImageMongo(w, r)
}

func (m *mongodb_struct) CreateHTLMPageMongo(w http.ResponseWriter, r *http.Request) {
	CreateHTLMPageMongo(w, r)
}

///////////////////////////
///// PARTIE SCYLLADB /////
///////////////////////////

type scyllaDB_Interface interface {
	CreateProfileScylla(w http.ResponseWriter, r *http.Request)
	GetAllUsersScylla(w http.ResponseWriter, r *http.Request)
	GetUserProfileScylla(w http.ResponseWriter, r *http.Request)
	UpdateProfileScylla(w http.ResponseWriter, r *http.Request)
	DeleteProfileScylla(w http.ResponseWriter, r *http.Request)
	UploadProfileImageScylla(w http.ResponseWriter, r *http.Request)
	GetProfileImageScylla(w http.ResponseWriter, r *http.Request)
	CreateHTLMPageScylla(w http.ResponseWriter, r *http.Request)
	DeleteAllDatabaseScylla(w http.ResponseWriter, r *http.Request)
	getAllUsersTypeScylla(w http.ResponseWriter, r *http.Request)
}

type scylladb_struct struct {
	cluster *gocql.ClusterConfig
	session *gocql.Session
}

func (s *scylladb_struct) CreateProfileScylla(w http.ResponseWriter, r *http.Request) {
	CreateProfileScylla(w, r)
}

func (s *scylladb_struct) GetAllUsersScylla(w http.ResponseWriter, r *http.Request) {
	GetAllUsersScylla(w, r)
}

func (s *scylladb_struct) GetUserProfileScylla(w http.ResponseWriter, r *http.Request) {
	GetUserProfileScylla(w, r)
}

func (s *scylladb_struct) UpdateProfileScylla(w http.ResponseWriter, r *http.Request) {
	UpdateProfileScylla(w, r)
}

func (s *scylladb_struct) DeleteProfileScylla(w http.ResponseWriter, r *http.Request) {
	DeleteProfileScylla(w, r)
}

func (s *scylladb_struct) UploadProfileImageScylla(w http.ResponseWriter, r *http.Request) {
	UploadProfileImageScylla(w, r)
}

func (s *scylladb_struct) GetProfileImageScylla(w http.ResponseWriter, r *http.Request) {
	GetProfileImageScylla(w, r)
}

func (s *scylladb_struct) CreateHTLMPageScylla(w http.ResponseWriter, r *http.Request) {
	CreateHTLMPageScylla(w, r)
}

func (s *scylladb_struct) DeleteAllDatabaseScylla(w http.ResponseWriter, r *http.Request) {
	DeleteAllDatabaseScylla(w, r)
}

func (s *scylladb_struct) getAllUsersTypeScylla(w http.ResponseWriter, r *http.Request) {
	getAllUsersTypeScylla(w, r)
}

/////////////////////////////
///// PARTIE COCKCROACH /////
/////////////////////////////

type cockroachDB_Interface interface {
	CreateProfileCockroach(w http.ResponseWriter, r *http.Request)
	GetAllUsersCockroach(w http.ResponseWriter, r *http.Request)
	GetUserProfileCockroach(w http.ResponseWriter, r *http.Request)
	UpdateProfileCockroach(w http.ResponseWriter, r *http.Request)
	DeleteProfileCockroach(w http.ResponseWriter, r *http.Request)
	UploadProfileImageCockroach(w http.ResponseWriter, r *http.Request)
	GetProfileImageCockroach(w http.ResponseWriter, r *http.Request)
	CreateHTLMPageCockroach(w http.ResponseWriter, r *http.Request)
	getAllUsersTypeCockroach(w http.ResponseWriter, r *http.Request)
	DropTableAndRecreateCockroach(w http.ResponseWriter, r *http.Request)
}

type cockroachdb_struct struct {
	db *gorm.DB
}

func (c *cockroachdb_struct) CreateProfileCockroach(w http.ResponseWriter, r *http.Request) {
	CreateProfileCockroach(w, r)
}

func (c *cockroachdb_struct) GetAllUsersCockroach(w http.ResponseWriter, r *http.Request) {
	GetAllUsersCockroach(w, r)
}

func (c *cockroachdb_struct) GetUserProfileCockroach(w http.ResponseWriter, r *http.Request) {
	GetUserProfileCockroach(w, r)
}

func (c *cockroachdb_struct) UpdateProfileCockroach(w http.ResponseWriter, r *http.Request) {
	UpdateProfileCockroach(w, r)
}

func (c *cockroachdb_struct) DeleteProfileCockroach(w http.ResponseWriter, r *http.Request) {
	DeleteProfileCockroach(w, r)
}

func (c *cockroachdb_struct) UploadProfileImageCockroach(w http.ResponseWriter, r *http.Request) {
	UploadProfileImageCockroach(w, r)
}

func (c *cockroachdb_struct) GetProfileImageCockroach(w http.ResponseWriter, r *http.Request) {
	GetProfileImageCockroach(w, r)
}

func (c *cockroachdb_struct) CreateHTLMPageCockroach(w http.ResponseWriter, r *http.Request) {
	CreateHTMLPageCockroach(w, r)
}

func (c *cockroachdb_struct) getAllUsersTypeCockroach(w http.ResponseWriter, r *http.Request) {
	getAllUsersTypeCockroach(w, r)
}

func (c *cockroachdb_struct) DropTableAndRecreateCockroach(w http.ResponseWriter, r *http.Request) {
	DropTableAndRecreateCockroach(w, r)
}

///////////////////////////
////// PARTIE INIT ////////
///////////////////////////

func initMongoDB(m mongoDB_Interface) {
	route := mux.NewRouter()
	log.Println("On créer le routeur")
	s := route.PathPrefix("/api").Subrouter() // on créer un sous-routeur pour les routes de l'api

	log.Println("On créer les routes")
	s.HandleFunc("/createProfile", m.CreateProfileMongo).Methods("POST")
	s.HandleFunc("/getAllUsers", m.GetAllUsersMongo).Methods("GET")
	s.HandleFunc("/getUserProfile", m.GetUserProfileMongo).Methods("POST")
	s.HandleFunc("/updateProfile", m.UpdateProfileMongo).Methods("PUT")
	s.HandleFunc("/deleteProfile/{id}", m.DeleteProfileMongo).Methods("DELETE")
	s.HandleFunc("/uploadProfileImage", m.UploadProfileImageMongo).Methods("POST")
	s.HandleFunc("/getProfileImage", m.GetProfileImageMongo).Methods("POST")
	s.HandleFunc("/createHtmlPage", m.CreateHTLMPageMongo).Methods("POST")

	log.Println("On lance le serveur sur le port 8080")
	log.Fatal(http.ListenAndServe(":8080", s)) // on lance le serveur sur le port 8080
}

func initScyllaDB(s scyllaDB_Interface) {
	route := mux.NewRouter()
	log.Println("On créer le routeur")
	s2 := route.PathPrefix("/api").Subrouter() // on créer un sous-routeur pour les routes de l'api

	log.Println("On créer les routes")
	s2.HandleFunc("/createProfile", s.CreateProfileScylla).Methods("POST")
	s2.HandleFunc("/getAllUsers", s.GetAllUsersScylla).Methods("GET")
	s2.HandleFunc("/getUserProfile", s.GetUserProfileScylla).Methods("POST")
	s2.HandleFunc("/updateProfile", s.UpdateProfileScylla).Methods("PUT")
	s2.HandleFunc("/deleteProfile/{id}", s.DeleteProfileScylla).Methods("DELETE")
	s2.HandleFunc("/uploadProfileImage", s.UploadProfileImageScylla).Methods("POST")
	s2.HandleFunc("/getProfileImage", s.GetProfileImageScylla).Methods("POST")
	s2.HandleFunc("/createHtmlPage", s.CreateHTLMPageScylla).Methods("POST")
	s2.HandleFunc("/deleteAllDatabase", s.DeleteAllDatabaseScylla).Methods("DELETE")
	s2.HandleFunc("/getAllUsersState", s.getAllUsersTypeScylla).Methods("POST")

	log.Println("On lance le serveur sur le port 8080")
	log.Fatal(http.ListenAndServe(":8080", s2)) // on lance le serveur sur le port 8080
}

func initCockroachDB(c cockroachDB_Interface) {
	route := mux.NewRouter()
	log.Println("On créer le routeur")
	s3 := route.PathPrefix("/api").Subrouter() // on créer un sous-routeur pour les routes de l'api

	log.Println("On créer les routes")
	s3.HandleFunc("/createProfile", c.CreateProfileCockroach).Methods("POST")
	s3.HandleFunc("/getAllUsers", c.GetAllUsersCockroach).Methods("GET")
	s3.HandleFunc("/getUserProfile", c.GetUserProfileCockroach).Methods("POST")
	s3.HandleFunc("/updateProfile", c.UpdateProfileCockroach).Methods("PUT")
	s3.HandleFunc("/deleteProfile", c.DeleteProfileCockroach).Methods("DELETE")
	s3.HandleFunc("/uploadProfileImage", c.UploadProfileImageCockroach).Methods("POST")
	s3.HandleFunc("/getProfileImage", c.GetProfileImageCockroach).Methods("POST")
	s3.HandleFunc("/createHtmlPage", c.CreateHTLMPageCockroach).Methods("POST")
	s3.HandleFunc("/getAllUsersState", c.getAllUsersTypeCockroach).Methods("POST")
	s3.HandleFunc("/deleteAllDatabase", c.DropTableAndRecreateCockroach).Methods("DELETE")

	log.Println("On lance le serveur sur le port 8080")
	log.Fatal(http.ListenAndServe(":8080", s3)) // on lance le serveur sur le port 8080
}

////////////////
///// MAIN /////
////////////////

func main() {

	//var scylladb_interface scyllaDB_Interface = &scylladb_struct{}
	//var mongodb_interface mongoDB_Interface = &mongodb_struct{}
	var cockroachdb_interface cockroachDB_Interface = &cockroachdb_struct{}

	//initMongoDB(mongodb_interface)
	initCockroachDB(cockroachdb_interface)
	//initScyllaDB(scylladb_interface)
}
