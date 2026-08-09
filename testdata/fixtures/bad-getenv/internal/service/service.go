package service

import "os"

type Service struct {
	name string
}

func New() *Service {
	// getenv в бизнес-коде — cfg-01 должен сработать
	return &Service{name: os.Getenv("SERVICE_NAME")}
}

func (s *Service) Name() string {
	return s.name
}
