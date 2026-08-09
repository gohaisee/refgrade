package service

import "os"

type Service struct {
	name string
}

func New() *Service {
	return &Service{name: "bad-ignored-error"}
}

func (s *Service) Name() string {
	return s.name
}

func (s *Service) Save(path string, data []byte) error {
	err := os.WriteFile(path, data, 0o644)
	_ = err
	return nil
}

func (s *Service) Load(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
	}
	return data, nil
}
