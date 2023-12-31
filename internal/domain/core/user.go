package core

import (
	"birthday-bot/internal/domain/entities"
	"birthday-bot/internal/domain/errs"
	"context"
)

const DefaultListSize = int64(100)

type User struct {
	r *St
}

func NewUser(r *St) *User {
	return &User{r: r}
}

func (c *User) Validate(ctx context.Context, obj *entities.UserCUSt) error {
	return nil
}

func (c *User) Get(ctx context.Context, id int64, errNE bool) (*entities.UserSt, error) {
	result, err := c.r.repo.UserGet(ctx, id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		if errNE {
			return nil, errs.ObjectNotFound
		}
		return nil, nil
	}

	return result, nil
}

func (c *User) GetByEmail(ctx context.Context, email string, errNE bool) (*entities.UserSt, error) {
	result, err := c.r.repo.UserGetByEmail(ctx, id)
	if err != nil {
		return nil, err
	}
	if result == nil {
		if errNE {
			return nil, errs.ObjectNotFound
		}
		return nil, nil
	}

	return result, nil
}

func (c *User) Create(ctx context.Context, obj *entities.UserCUSt) (int64, error) {
	var err error

	err = c.Validate(ctx, obj)
	if err != nil {
		return -1, err
	}

	// create
	result, err := c.r.repo.UserCreate(ctx, obj)
	if err != nil {
		return -1, err
	}

	return result, nil
}

func (c *User) Update(ctx context.Context, id int64, obj *entities.UserCUSt) error {
	var err error

	err = c.Validate(ctx, obj)
	if err != nil {
		return err
	}

	err = c.r.repo.UserUpdate(ctx, id, obj)
	if err != nil {
		return err
	}

	return nil
}

func (c *User) Delete(ctx context.Context, id int64) error {
	return c.r.repo.UserDelete(ctx, id)
}
