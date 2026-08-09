package service

import "net/http"

type Service struct {
	name string
}

func New() *Service {
	return &Service{name: "bad-default-client"}
}

func (s *Service) Name() string {
	return s.name
}

func (s *Service) Ping(url string) (*http.Response, error) {
	return http.DefaultClient.Get(url)
}
