package service

type Service struct {
	name string
}

func New(name string) *Service {
	return &Service{name: name}
}

func (s *Service) Name() string {
	return s.name
}
