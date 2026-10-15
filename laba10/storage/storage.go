package storage

import "laba10/models"

type UserStorage interface {
	List(filter UserFilter) ([]models.User, error)
	Get(id int) (models.User, error)
	Create(user models.User) (models.User, error)
	Update(id int, u models.User) (models.User, error)
	Delete(id int) error

	CreateWithPassword(u models.User, passwordHash string) (models.User, error)
	FindByEmail(email string) (models.User, string, error)
}
