package main

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog/log"
	"github.com/shoyo10/wgorm"
	"github.com/shoyo10/wgorm/examples/repolayer/model"
	"github.com/shoyo10/wgorm/examples/repolayer/repository"
	"github.com/shoyo10/wzerolog"
	"github.com/spf13/viper"
)

type Config struct {
	Log      wzerolog.Config `yaml:"log"`
	Database wgorm.Config    `yaml:"database"`
}

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	var config Config
	err = viper.Unmarshal(&config, func(cfg *mapstructure.DecoderConfig) {
		cfg.TagName = "yaml"
	})
	if err != nil {
		panic(fmt.Errorf("Fatal error unmarshal config: %s", err))
	}

	wzerolog.Init(config.Log)

	ctx := context.Background()
	ctx = log.Logger.WithContext(ctx)

	g, err := wgorm.New(&config.Database)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("new wgorm failed: %+v", err)
		return
	}

	repo := repository.New(g)
	age := 12
	err = repo.CreateUser(ctx, model.User{
		Email:    "test@gmail.com",
		Password: "12345678",
		Age:      &age,
	})
	if err != nil {
		log.Ctx(ctx).Error().Msgf("create user failed: %+v", err)
	}

	user, err := repo.GetUser(ctx, repository.WhereCondUser{
		User: model.User{
			ID: 1,
		},
	})
	if err != nil {
		log.Ctx(ctx).Error().Msgf("get user failed: %+v", err)
	} else {
		log.Ctx(ctx).Debug().Msgf("user: %+v, age: %d", user, *user.Age)
	}

	users, err := repo.ListUser(ctx, repository.WhereCondUser{
		AgeNULL: true,
	})
	if err != nil {
		log.Ctx(ctx).Error().Msgf("list user failed: %+v", err)
	} else {
		log.Ctx(ctx).Debug().Msgf("user: %+v", users)
	}

	err = repo.Transaction(ctx, func(txRepo repository.IRepository) error {
		u, err := txRepo.GetUser(ctx, repository.WhereCondUser{
			User: model.User{
				ID: 1,
			},
		}, wgorm.SetForUpdate())
		if err != nil {
			return err
		}
		age := 12
		err = txRepo.UpdateUser(ctx, repository.UpdateUserReq{
			User: model.User{
				Password: "0987654321",
				Age:      &age,
			},
			Where: repository.WhereCondUser{
				User: model.User{
					ID: u.ID,
				},
			},
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Ctx(ctx).Error().Msgf("%+v", err)
	}

	err = repo.DeleteUser(ctx, repository.WhereCondUser{
		User: model.User{
			ID: user.ID,
		},
	})
	if err != nil {
		log.Ctx(ctx).Error().Msgf("delete user failed: %+v", err)
	}
}
