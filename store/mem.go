package store

import (
    "errors"

    "github.com/reshane/gorest/types"
)

type MemStore struct {
    users map[int]types.User
}

func NewMemStore() *MemStore {
    user := types.User{ ID: 1, Name: "Shane" }
    return &MemStore { users: map[int]types.User{ user.ID: user } }
}

func (s *MemStore) GetUser(id int) (*types.User, error) {
    if user, exists := s.users[id]; exists {
        return &user, nil
    }
    return nil, errors.New("Could not find user")
}

func (s *MemStore) CreateUser(user *types.User) (*types.User, error) {
    if _, exists := s.users[user.ID]; exists {
        return nil, errors.New("Could not create user: User with specified id already exists")
    }

    s.users[user.ID] = *user
    return user, nil
}

func (s *MemStore) UpdateUser(user *types.User) (*types.User, error) {
    if _, exists := s.users[user.ID]; !exists {
        return nil, errors.New("Could not find user to update")
    }

    s.users[user.ID] = *user
    return user, nil
}

func (s *MemStore) DeleteUser(id int) error {
    if _, exists := s.users[id]; exists {
        delete(s.users, id)
        return nil
    }
    return errors.New("Could not find user to delete")
}
