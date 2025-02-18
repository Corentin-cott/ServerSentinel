package models

type Server struct {
	ID          int
	Nom         string
	Jeu         string
	Version     string
	Modpack     string
	ModpackURL  string
	NomMonde    string
	EmbedColor  string
	PathServ    string
	StartScript string
	Actif       bool
	Global      bool
}
