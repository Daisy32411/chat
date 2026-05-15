package dialogs

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetOrCreate(user1, user2 string) (int, error) {
	return s.repo.GetOrCreateDialog(user1, user2)
}

func (s *Service) Get(user string) ([]Dialog, error) {
	return s.repo.GetDialogs(user)
}