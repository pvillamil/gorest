package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pilinux/gorest/config"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// Import MySQL database driver
	// _ "github.com/jinzhu/gorm/dialects/mysql"
	"gorm.io/driver/mysql"

	// Import PostgreSQL database driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"gorm.io/driver/postgres"

	// Import SQLite3 database driver
	// _ "github.com/jinzhu/gorm/dialects/sqlite"
	"gorm.io/driver/sqlite"

	// Import Redis Driver
	"github.com/mediocregopher/radix/v4"

	// Import Mongo driver
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	log "github.com/sirupsen/logrus"
)

// DB global variable to access gorm
var DB *gorm.DB

var sqlDB *sql.DB
var err error

// RedisClient global variable to access the redis client
var RedisClient radix.Client

// RedisConnTTL - context deadline in second
var RedisConnTTL int

// MongoClient instance
var MongoClient *mongo.Client

// InitDB - function to initialize db
func InitDB() *gorm.DB {
	var db = DB

	configureDB := config.Database().RDBMS

	driver := configureDB.Env.Driver
	username := configureDB.Access.User
	password := configureDB.Access.Pass
	database := configureDB.Access.DbName
	host := configureDB.Env.Host
	port := configureDB.Env.Port
	sslmode := configureDB.Ssl.Sslmode
	timeZone := configureDB.Env.TimeZone
	maxIdleConns := configureDB.Conn.MaxIdleConns
	maxOpenConns := configureDB.Conn.MaxOpenConns
	connMaxLifetime := configureDB.Conn.ConnMaxLifetime
	logLevel := configureDB.Log.LogLevel

	switch driver {
	case "mysql":
		dsn := username + ":" + password + "@tcp(" + host + ":" + port + ")/" + database + "?charset=utf8mb4&parseTime=True&loc=Local"
		sqlDB, err = sql.Open(driver, dsn)
		if err != nil {
			log.WithError(err).Panic("panic code: 151")
		}
		sqlDB.SetMaxIdleConns(maxIdleConns)       // max number of connections in the idle connection pool
		sqlDB.SetMaxOpenConns(maxOpenConns)       // max number of open connections in the database
		sqlDB.SetConnMaxLifetime(connMaxLifetime) // max amount of time a connection may be reused

		db, err = gorm.Open(mysql.New(mysql.Config{
			Conn: sqlDB,
		}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.LogLevel(logLevel)),
		})
		if err != nil {
			log.WithError(err).Panic("panic code: 152")
		}
		// Only for debugging
		if err == nil {
			fmt.Println("DB connection successful!")
		}

	case "postgres":
		dsn := "host=" + host + " port=" + port + " user=" + username + " dbname=" + database + " password=" + password + " sslmode=" + sslmode + " TimeZone=" + timeZone
		sqlDB, err = sql.Open(driver, dsn)
		if err != nil {
			log.WithError(err).Panic("panic code: 153")
		}
		sqlDB.SetMaxIdleConns(maxIdleConns)       // max number of connections in the idle connection pool
		sqlDB.SetMaxOpenConns(maxOpenConns)       // max number of open connections in the database
		sqlDB.SetConnMaxLifetime(connMaxLifetime) // max amount of time a connection may be reused

		db, err = gorm.Open(postgres.New(postgres.Config{
			Conn: sqlDB,
		}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.LogLevel(logLevel)),
		})
		if err != nil {
			log.WithError(err).Panic("panic code: 154")
		}
		// Only for debugging
		if err == nil {
			fmt.Println("DB connection successful!")
		}

	case "sqlite3":
		db, err = gorm.Open(sqlite.Open(database), &gorm.Config{
			Logger:                                   logger.Default.LogMode(logger.Silent),
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			log.WithError(err).Panic("panic code: 155")
		}
		// Only for debugging
		if err == nil {
			fmt.Println("DB connection successful!")
		}

	default:
		log.Fatal("The driver " + driver + " is not implemented yet")
	}

	DB = db

	return DB
}

// GetDB - get a connection
func GetDB() *gorm.DB {
	return DB
}

// InitRedis - function to initialize redis client
func InitRedis() radix.Client {
	configureRedis := config.Database().REDIS
	RedisConnTTL = configureRedis.Conn.ConnTTL

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(RedisConnTTL)*time.Second)
	defer cancel()

	rClient, err := (radix.PoolConfig{
		Size: configureRedis.Conn.PoolSize,
	}).New(ctx, "tcp", fmt.Sprintf("%v:%v",
		configureRedis.Env.Host,
		configureRedis.Env.Port))
	if err != nil {
		log.WithError(err).Panic("panic code: 161")
		fmt.Println(err)
	}
	// Only for debugging
	if err == nil {
		fmt.Println("REDIS pool connection successful!")
	}

	RedisClient = rClient

	return RedisClient
}

// GetRedis - get a connection
func GetRedis() radix.Client {
	return RedisClient
}

// InitMongo - function to initialize mongo client
func InitMongo() (*mongo.Client, error) {
	configureMongo := config.Database().MongoDB

	// Connect to the database or cluster
	URI := configureMongo.Env.URI

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(URI).
		SetServerAPIOptions(serverAPIOptions).
		SetMaxPoolSize(configureMongo.Env.PoolSize)

	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return client, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(configureMongo.Env.ConnTTL)*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return client, err
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return client, err
	}

	fmt.Println("MongoDB pool connection successful!")

	MongoClient = client

	return MongoClient, nil
}

// GetMongo - get a connection
func GetMongo() *mongo.Client {
	return MongoClient
}
