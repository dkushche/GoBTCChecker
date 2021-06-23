package store

import (
	"encoding/csv"
	"io"
	"os"
)

type Store struct {
	config   *Config
	Database map[string]string
}

func New(config *Config) (*Store, error) {
	var file *os.File
	database := make(map[string]string)

	if file, err := os.Open(config.DatabasePath); err != nil {
		if _, err = os.Create(config.DatabasePath); err != nil {
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

	return &Store{
		config:   config,
		Database: database,
	}, nil
}

func AddUser(s *Store, email string, password string) error {
	return nil
}
