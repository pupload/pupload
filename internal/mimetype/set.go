package mimetypes

import (
	"fmt"
	"strings"

	"github.com/pupload/pupload/internal/models"
)

type MimeSet struct {
	Set map[models.MimeType]struct{}
}

// Creates and Validates a given mimeset
func CreateMimeSet(mimes []models.MimeType) (*MimeSet, error) {

	var mime_set MimeSet
	mime_set.Set = make(map[models.MimeType]struct{})

	for _, mime := range mimes {

		if err := Validate(mime); err != nil {
			return nil, fmt.Errorf("CreateMimeSet: Error found in type %s, %w", mime, err)
		}

		mime_set.Add(mime)
	}

	return &mime_set, nil
}

func (m *MimeSet) Add(t models.MimeType) {
	m.Set[t] = struct{}{}
}

func (m *MimeSet) Contains(t models.MimeType) bool {

	prefixString := strings.Split(string(t), "/")[0]
	wildcard := models.MimeType(fmt.Sprintf("%s/*", prefixString))

	if _, ok := m.Set[wildcard]; ok {
		return true
	}

	_, ok := m.Set[t]
	return ok
}

func (m1 *MimeSet) Join(m2 MimeSet) {
	for k := range m2.Set {
		m1.Add(k)
	}
}

func (m1 *MimeSet) Intersection(m2 MimeSet) *MimeSet {
	m := new(MimeSet)
	m.Set = make(map[models.MimeType]struct{})

	for k := range m2.Set {
		if m1.Contains(k) {
			m.Add(k)
		}
	}

	return m
}

func (m *MimeSet) IsEmpty() bool {
	return len(m.Set) == 0
}
