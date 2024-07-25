package storage

import (
	"time"
	"users-service/storage/entities"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgreStorage struct {
	db *gorm.DB
}

func NewPostgreStore(dsn string) (*PostgreStorage, error) {
	store := &PostgreStorage{}

	err := store.init(dsn)
	if err != nil {
		return nil, err
	}

	store.db.AutoMigrate(&entities.User{})

	return store, nil
}

func (s *PostgreStorage) init(dsn string) error {
	var db *gorm.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}

		time.Sleep(time.Second)
	}

	if err != nil {
		return err
	}

	s.db = db

	return nil
}

func (s *PostgreStorage) CreateUser(user *entities.User) (int64, error) {
	err := s.db.Create(user).Error

	return user.ID, err
}

func (s *PostgreStorage) GetUserByID(id uint64) (*entities.User, error) {
	var user *entities.User

	if err := s.db.First(user, id).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (s *PostgreStorage) GetUserByUsername(username string) (*entities.User, error) {
	var user *entities.User

	if err := s.db.First(user, "username = ?", username).Error; err != nil {
		return nil, err
	}

	return user, nil
}
