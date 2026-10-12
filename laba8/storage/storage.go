package storage

import "laba8/models"

type UserStorage interface {
	List(filter UserFilter) ([]models.User, error)
	Get(id int) (models.User, error)
	Create(user models.User) (models.User, error)
	Update(id int, u models.User) (models.User, error)
	Delete(id int) error
}
