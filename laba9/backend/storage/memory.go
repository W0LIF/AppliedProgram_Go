package storage

import (
	"errors"
	"sort"
	"strings"
	"sync"

	"laba9/models"
)

var ErrNotFound = errors.New("user not found")

type MemoryStorage struct {
	mu        sync.RWMutex
	users     map[int]models.User
	passwords map[int]string
	nextID    int
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		users:     make(map[int]models.User),
		passwords: make(map[int]string),
		nextID:    1,
	}
}

func (s *MemoryStorage) List(f UserFilter) ([]models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	filtered := make([]models.User, 0, len(s.users))
	for _, u := range s.users {
		if f.Name != "" && !strings.Contains(strings.ToLower(u.Name), strings.ToLower(f.Name)) {
			continue
		}
		if f.MinAge > 0 && u.Age < f.MinAge {
			continue
		}
		if f.MaxAge > 0 && u.Age > f.MaxAge {
			continue
		}
		filtered = append(filtered, u)
	}

	sort.Slice(filtered, func(i, j int) bool { return filtered[i].ID < filtered[j].ID })

	start := f.Offset
	if start > len(filtered) {
		return []models.User{}, nil
	}
	end := len(filtered)
	if f.Limit > 0 && start+f.Limit < end {
		end = start + f.Limit
	}
	return filtered[start:end], nil
}

func (s *MemoryStorage) Get(id int) (models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.users[id]
	if !ok {
		return models.User{}, ErrNotFound
	}
	return u, nil
}

func (s *MemoryStorage) Create(u models.User) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u.ID = s.nextID
	s.nextID++
	s.users[u.ID] = u
	return u, nil
}

func (s *MemoryStorage) Update(id int, u models.User) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[id]; !ok {
		return models.User{}, ErrNotFound
	}
	u.ID = id
	s.users[id] = u
	return u, nil
}

func (s *MemoryStorage) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[id]; !ok {
		return ErrNotFound
	}
	delete(s.users, id)
	delete(s.passwords, id)
	return nil
}

func (s *MemoryStorage) CreateWithPassword(u models.User, hash string) (models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u.ID = s.nextID
	s.nextID++
	s.users[u.ID] = u
	s.passwords[u.ID] = hash
	return u, nil
}

func (s *MemoryStorage) FindByEmail(email string) (models.User, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, u := range s.users {
		if strings.EqualFold(u.Email, email) {
			return u, s.passwords[u.ID], nil
		}
	}
	return models.User{}, "", ErrNotFound
}

var _ UserStorage = (*MemoryStorage)(nil)
