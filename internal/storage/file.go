package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type FileStore struct {
	Dir string
}

func (s *FileStore) tokenPath(profile string) string {
	return filepath.Join(s.Dir, fmt.Sprintf("%s.json", profile))
}

func validProfile(profile string) bool {
	return profile != "" && profile != "." && profile != ".." && filepath.Base(profile) == profile
}

func (s *FileStore) Get(profile string) (*Token, error) {
	if !validProfile(profile) {
		return nil, errors.New("invalid profile name")
	}
	data, err := os.ReadFile(s.tokenPath(profile))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

func (s *FileStore) Set(profile string, token *Token) error {
	if !validProfile(profile) {
		return errors.New("invalid profile name")
	}
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.tokenPath(profile), data, 0o600)
}

func (s *FileStore) Delete(profile string) error {
	if !validProfile(profile) {
		return errors.New("invalid profile name")
	}
	if err := os.Remove(s.tokenPath(profile)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
