package storage

import (
	"encoding/csv"
	"errors"
	"io"
	"os"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"golang.org/x/crypto/bcrypt"
)

type Storage struct {
	database_path string
	database      map[string]string
}

func New(database_path string) (*Storage, error) {
	var file *os.File
	database := make(map[string]string)

	file, err := os.Open(database_path)
	if err != nil {
		if file, err = os.Create(database_path); err != nil {
			return nil, err
		}
	} else {
		reader := csv.NewReader(file)
		reader.FieldsPerRecord = 2

		for {
			record, err := reader.Read()
			if err != nil {
				if err != io.EOF {
					return nil, err
				}
				break
			}
			database[record[0]] = record[1]
		}
	}
	file.Close()

	return &Storage{
		database_path: database_path,
		database:      database,
	}, nil
}

func (s *Storage) Find(email string) (string, error) {
	hash, exists := s.database[email]
	if exists {
		return hash, nil
	}
	return "", errors.New("incorrect email or password")
}

func (s *Storage) UserAuth(email string, password string) error {
	if err := s.validate(email, password); err != nil {
		return err
	}

	hash, err := s.Find(email)
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return errors.New("incorrect email or password")
	}

	return nil
}

func (s *Storage) AddUser(email string, password string) error {
	if err := s.validate(email, password); err != nil {
		return err
	}

	_, err := s.Find(email)
	if err == nil {
		return errors.New("incorrect email or password")
	}

	hashed_password, err := s.encrypt(password)
	if err != nil {
		return err
	}

	if err := s.save(email, hashed_password); err != nil {
		return err
	}

	return nil
}

func (s *Storage) validate(email string, password string) error {
	if len(email) == 0 || len(password) == 0 {
		return errors.New("empty email or password")
	}

	if err := validation.Validate(email, validation.Required, is.Email); err != nil {
		return err
	}
	if err := validation.Validate(password, validation.Required, validation.Length(6, 15)); err != nil {
		return err
	}

	return nil
}

func (s *Storage) encrypt(password string) (string, error) {
	hashed_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(hashed_password), nil
}

func (s *Storage) save(email string, hashed_password string) error {
	file, err := os.OpenFile(s.database_path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	_, err = file.WriteString(email + "," + hashed_password + "\n")
	if err != nil {
		return err
	}
	file.Close()

	s.database[email] = hashed_password

	return nil
}
