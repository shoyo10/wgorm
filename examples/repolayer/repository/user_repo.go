package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
	"github.com/shoyo10/wgorm"
	"github.com/shoyo10/wgorm/examples/repolayer/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepo interface {
	CreateUser(ctx context.Context, user model.User, opts ...wgorm.Option) error
	GetUser(ctx context.Context, condition WhereCondUser, opts ...wgorm.Option) (model.User, error)
	ListUser(ctx context.Context, condition WhereCondUser, opts ...wgorm.Option) ([]model.User, error)
	UpdateUser(ctx context.Context, req UpdateUserReq, opts ...wgorm.Option) error
	DeleteUser(ctx context.Context, condition WhereCondUser, opts ...wgorm.Option) error
}

type User struct {
	ID        int64          `json:"id"`
	Email     string         `json:"email"`
	Password  string         `json:"password"`
	Age       sql.NullInt32  `json:"age"`
	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at"`
}

func (u *User) convertModelToRepo(user model.User) {
	u.ID = user.ID
	u.Email = user.Email
	u.Password = user.Password
	if user.Age != nil {
		u.Age = sql.NullInt32{
			Int32: int32(*user.Age),
			Valid: true,
		}
	}
}

func (u *User) convertRepoToModel(user *model.User) {
	user.ID = u.ID
	user.Email = u.Email
	user.Password = u.Password
	if u.Age.Valid {
		var age int = int(u.Age.Int32)
		user.Age = &age
	}
	user.CreatedAt = u.CreatedAt
	user.UpdatedAt = u.UpdatedAt
	if u.DeletedAt.Valid {
		*user.DeletedAt = u.DeletedAt.Time
	}
}

type WhereCondUser struct {
	User       model.User
	user       User
	AgeGte     int
	AgeLte     int
	AgeNULL    bool
	AgeNotNULL bool
}

func (w WhereCondUser) Scope(tx *gorm.DB) *gorm.DB {
	tx = tx.Where(w.User)
	if w.AgeGte > 0 {
		tx = tx.Where("age >= ?", w.AgeGte)
	}
	if w.AgeLte > 0 {
		tx = tx.Where("age <= ?", w.AgeLte)
	}
	if w.AgeNULL {
		tx = tx.Where("age IS NULL")
	}
	if w.AgeNotNULL {
		tx = tx.Where("age IS NOT NULL")
	}
	return tx
}

type UpdateUserReq struct {
	User          model.User
	user          User
	Where         WhereCondUser
	UpdateColumns map[string]interface{}
}

func (u *UpdateUserReq) Scope(tx *gorm.DB) *gorm.DB {
	tx = tx.Scopes(u.Where.Scope)
	return tx
}

func (r *repo) CreateUser(ctx context.Context, user model.User, opts ...wgorm.Option) error {
	var u User
	u.convertModelToRepo(user)
	err := r.db.WithContext(ctx).Options(opts...).GormDB().Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{
				Name: "email",
			},
		},
		Where: clause.Where{
			Exprs: []clause.Expression{
				clause.AndConditions{
					Exprs: []clause.Expression{
						clause.Expr{SQL: "users.deleted_at IS NOT NULL"},
					},
				},
			},
		},
		UpdateAll: true,
	}).Create(&u).Error
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r *repo) GetUser(ctx context.Context, condition WhereCondUser, opts ...wgorm.Option) (model.User, error) {
	var repoUser User
	var user model.User

	condition.user.convertModelToRepo(condition.User)
	err := r.db.WithContext(ctx).Options(opts...).GormDB().Scopes(condition.Scope).First(&repoUser).Error
	if err != nil {
		return user, errors.WithStack(err)
	}
	repoUser.convertRepoToModel(&user)

	return user, nil
}

func (r *repo) ListUser(ctx context.Context, condition WhereCondUser, opts ...wgorm.Option) ([]model.User, error) {
	var repoUsers []User
	var users []model.User

	condition.user.convertModelToRepo(condition.User)
	err := r.db.WithContext(ctx).Options(opts...).GormDB().Scopes(condition.Scope).Find(&repoUsers).Error
	if err != nil {
		return users, errors.WithStack(err)
	}
	for i := range repoUsers {
		var user model.User
		u := repoUsers[i]
		u.convertRepoToModel(&user)
		users = append(users, user)
	}
	return users, nil
}

func (r *repo) UpdateUser(ctx context.Context, req UpdateUserReq, opts ...wgorm.Option) error {
	req.user.convertModelToRepo(req.User)
	req.Where.user.convertModelToRepo(req.Where.User)
	tx := r.db.WithContext(ctx).Options(opts...).GormDB()
	err := tx.Scopes(req.Scope).Omit("id", "email").Updates(&req.user).Error
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (r *repo) DeleteUser(ctx context.Context, condition WhereCondUser, opts ...wgorm.Option) error {
	err := r.db.WithContext(ctx).Options(opts...).GormDB().Scopes(condition.Scope).Delete(&User{}).Error
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
