package repository

type UserRepository interface {
}

type InMemoryUserRepository struct {
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{}
}
