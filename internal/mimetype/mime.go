package mimetypes

import (
	_ "embed"
	"encoding/json"
	"errors"
	"strings"

	"github.com/pupload/pupload/internal/models"
)

type MimeDBRow struct {
	Source       *string   `json:"source"`
	CharSet      *string   `json:"charset"`
	Compressible *string   `json:"compressible"`
	Extensions   *[]string `json:"extensions"`
}

type MimeDB map[models.MimeType]MimeDBRow

//go:embed db.json
var mime_db_raw string
var mime_db MimeDB

func createDB() {
	json.Unmarshal([]byte(mime_db_raw), &mime_db)
}

func GetDB() MimeDB {
	if mime_db == nil {
		createDB()
	}

	return mime_db
}

var ErrNoMimeDB error = errors.New("MimeDB was not initalized")
var ErrInvalidMime error = errors.New("MimeType does not exist")

func Validate(m models.MimeType) error {

	db := GetDB()

	_, mimeExists := db[m]

	if !mimeExists && !strings.HasSuffix(string(m), "*") {
		return ErrInvalidMime
	}

	return nil
}

func GetExtensionFromMime(m models.MimeType) string {
	db := GetDB()

	entry, ok := db[m]
	if !ok {
		return ""
	}

	if entry.Extensions == nil || len(*entry.Extensions) == 0 {
		return ""
	}

	return ("." + (*entry.Extensions)[0])

}
