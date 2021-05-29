package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shoyo10/wgorm"
	"github.com/shoyo10/wzerolog"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type User struct {
	ID        int64          `json:"id"`
	Email     string         `json:"email"`
	Password  string         `json:"password"`
	Age       sql.NullInt32  `json:"age"`
	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at"`
}

type Config struct {
	Database wgorm.Config `yaml:"database"`
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

	wzerolog.Init(wzerolog.Config{
		LogLevel:     zerolog.DebugLevel,
		PrettyOutput: true,
	})

	ctx := context.Background()
	ctx = log.Logger.WithContext(ctx)

	g, err := wgorm.New(&config.Database)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("new wgorm failed: %v", err)
		return
	}
	err = g.WithContext(ctx).Create(&User{
		Email:    "test@gmail.com",
		Password: "12345678",
	}).Error
	if err != nil {
		log.Ctx(ctx).Error().Msgf("create user failed: %v", err)
	}

	var u User
	err = g.WithContext(ctx).Model(&User{}).Where("id = ?", 1).First(&u).Error
	if err != nil {
		log.Ctx(ctx).Error().Msgf("get user failed: %v", err)
	} else {
		log.Ctx(ctx).Debug().Msgf("user: %+v", u)
	}
}
