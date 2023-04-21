package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// on se connecte à la base de données

	route := mux.NewRouter()
	log.Println("On créer le routeur")
	s := route.PathPrefix("/api").Subrouter() // on créer un sous-routeur pour les routes de l'api

	log.Println("On créer les routes")
	s.HandleFunc("/createProfile", CreateProfile).Methods("POST")
	s.HandleFunc("/getAllUsers", GetAllUsers).Methods("GET")
	s.HandleFunc("/getUserProfile", GetUserProfile).Methods("POST")
	s.HandleFunc("/updateProfile", UpdateProfile).Methods("PUT")
	s.HandleFunc("/deleteProfile/{id}", DeleteProfile).Methods("DELETE")
	s.HandleFunc("/uploadProfileImage", UploadProfileImage).Methods("POST")
	s.HandleFunc("/getProfileImage", GetProfileImage).Methods("POST")

	log.Println("On lance le serveur sur le port 8080")
	log.Fatal(http.ListenAndServe(":8080", s)) // on lance le serveur sur le port 8080
}
