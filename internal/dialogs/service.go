package dialogs

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(user1, user2 string) (int, error) {
	return s.repo.CreateDialog(user1, user2)
}

func (s *Service) Get(user string) ([]Dialog, error) {
	return s.repo.GetDialogs(user)
}