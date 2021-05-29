package wgorm

import (
	"fmt"
	"strings"
	"time"

	"github.com/shoyo10/wgorm/logger"

	"github.com/cenk/backoff"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// DatabaseDriver 類型
type DatabaseDriver string

const (
	// Postgres ...
	Postgres DatabaseDriver = "postgres"
)

type ConnConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int32  `yaml:"port" mapstructure:"port"`
	Username string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
	DBName   string `yaml:"dbname" mapstructure:"dbname"`

	// for postgresql
	SearchPath string `yaml:"search_path" mapstructure:"search_path"`
	SSLEnable  bool   `yaml:"ssl_enable" mapstructure:"ssl_enable"`

	dsn string
}

type Config struct {
	Driver             DatabaseDriver `yaml:"driver" mapstructure:"driver"`
	MaxIdleConns       int            `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`
	MaxOpenConns       int            `yaml:"max_open_conns" mapstructure:"max_open_conns"`
	ConnMaxLifeTimeSec int            `yaml:"conn_max_life_time_sec" mapstructure:"conn_max_life_time_sec"`
	Master             ConnConfig     `yaml:"master" mapstructure:"master"`
	Slave              []ConnConfig   `yaml:"slave" mapstructure:"slave"`
	Log                logger.Config  `yaml:"log" mapstructure:"log"`
}

func (cfg *Config) clone() (*Config, error) {
	var config Config
	switch cfg.Driver {
	case Postgres:
		config = *cfg
	default:
		return nil, errors.WithStack(fmt.Errorf("not support driver:%s", cfg.Driver))
	}
	return &config, nil
}

func (cfg *Config) newGormDB() (*gorm.DB, error) {
	if err := cfg.setConnectionInfo(); err != nil {
		return nil, err
	}

	db, err := cfg.connectMasterDB()
	if err != nil {
		return nil, err
	}

	if len(cfg.Slave) > 0 {
		if err := cfg.connecSlaveDB(db); err != nil {
			return nil, err
		}
	}

	return db, nil
}

func (cfg *Config) setConnectionInfo() error {
	if err := cfg.setDSN(); err != nil {
		return err
	}

	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 50
	}
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 100
	}
	if cfg.ConnMaxLifeTimeSec == 0 {
		cfg.ConnMaxLifeTimeSec = 3600
	}

	return nil
}

func (cfg *Config) setDSN() error {
	switch cfg.Driver {
	case Postgres:
		cfg.Master.setPostgresDSN()
		for i, cc := range cfg.Slave {
			cc.setPostgresDSN()
			cfg.Slave[i] = cc
		}
	default:
		return errors.WithStack(fmt.Errorf("not support driver:%s", cfg.Driver))
	}

	return nil
}

func (cc *ConnConfig) setPostgresDSN() {
	dsn := fmt.Sprintf(`user=%s password=%s host=%s port=%d dbname=%s`, cc.Username, cc.Password, cc.Host, cc.Port, cc.DBName)
	if cc.SSLEnable {
		dsn += " sslmode=require"
	} else {
		dsn += " sslmode=disable"
	}
	if strings.TrimSpace(cc.SearchPath) != "" {
		dsn = fmt.Sprintf("%s search_path=%s", dsn, cc.SearchPath)
	}
	cc.dsn = dsn
}

func (cfg *Config) getDialector(dsn string) (gorm.Dialector, error) {
	var dialector gorm.Dialector
	switch cfg.Driver {
	case Postgres:
		dialector = postgres.Open(dsn)
	default:
		return nil, errors.WithStack(fmt.Errorf("not support driver:%s", cfg.Driver))
	}

	return dialector, nil
}

func (cfg *Config) connectMasterDB() (*gorm.DB, error) {
	dialector, err := cfg.getDialector(cfg.Master.dsn)
	if err != nil {
		return nil, err
	}

	newLogger := logger.New(cfg.Log)

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = time.Duration(180) * time.Second
	var db *gorm.DB

	err = backoff.Retry(func() error {
		db, err = gorm.Open(dialector, &gorm.Config{
			Logger: newLogger,
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		})
		if err != nil {
			return err
		}

		sqlDB, err := db.DB()
		if err != nil {
			return err
		}

		err = sqlDB.Ping()
		return err
	}, bo)

	if err != nil {
		return nil, errors.WithStack(fmt.Errorf("connect db failed: %v", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifeTimeSec) * time.Second)

	return db, nil
}

func (cfg *Config) connecSlaveDB(db *gorm.DB) error {
	var dialectors []gorm.Dialector
	for _, cc := range cfg.Slave {
		d, err := cfg.getDialector(cc.dsn)
		if err != nil {
			return err
		}
		dialectors = append(dialectors, d)
	}
	err := db.Use(
		dbresolver.Register(dbresolver.Config{
			Replicas: dialectors,
		}).
			SetMaxIdleConns(cfg.MaxIdleConns).
			SetMaxOpenConns(cfg.MaxOpenConns).
			SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifeTimeSec) * time.Second))
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
