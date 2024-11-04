package database

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/ferho20u12/nd.go/foundation"

	"github.com/gookit/color"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/dbresolver"
)

var (
	DB Database

	ErrNoConnections             = errors.New("please add at least one or more connections to the database config") // HTTP 500 Internal Server Error
	ErrNoDefaultConnection       = errors.New("please specify a default connection in the database config")         // HTTP 500 Internal Server Error
	ErrDefaultConnectionNotFound = errors.New("default connection not found in the database config")                // HTTP 500 Internal Server Error
	ErrConnectionFailed          = errors.New("connection failed")                                                  // HTTP 500 Internal Server Error
	ErrFailedToGetSQLDB          = errors.New("failed to get sqlDB from gorm.DB")                                   // HTTP 500 Internal Server Error
	ErrBindingConnection         = errors.New("error binding connection")                                           // HTTP 500 Internal Server Error
)

type Database struct {
	Ctx *gorm.DB
}

func isIPAddress(host string) bool {
	return net.ParseIP(host) != nil
}

func buildDSN(DBServer, DBUsername, DBPassword, DBDatabase, DBSsl, DBCharset string) string {
	var dsn string
	if isIPAddress(DBServer) {
		dsn = fmt.Sprintf("server=%s;user id=%s;password=%s;database=%s;encrypt=%s;connection timeout=30;charset=%s", DBServer, DBUsername, DBPassword, DBDatabase, DBSsl, DBCharset)
	} else {
		dsn = fmt.Sprintf("sqlserver://%s:%s@%s?database=%s&encrypt=%s&connection+timeout=30&charset=%s", DBUsername, DBPassword, DBServer, DBDatabase, DBSsl, DBCharset)
	}
	return dsn
}

func buildDSNFromConfig(connection map[string]interface{}) string {
	return buildDSN(
		connection["server"].(string),
		connection["username"].(string),
		connection["password"].(string),
		connection["database"].(string),
		connection["ssl"].(string),
		connection["charset"].(string),
	)
}

func Boot() {
	connections, ok := foundation.App.Config.Get("database.connections", nil).(map[string]interface{})
	if !ok || connections == nil {
		color.Redln(ErrNoConnections)
		os.Exit(1)
	}

	defaultNameConnection, ok := foundation.App.Config.Get("database.default", "").(string)
	if !ok || defaultNameConnection == "" {
		color.Redln(ErrNoDefaultConnection)
		os.Exit(1)
	}

	connectionMain, ok := connections[defaultNameConnection].(map[string]interface{})
	if !ok {
		color.Redln(ErrDefaultConnectionNotFound)
		os.Exit(1)
	}

	dsnMain := buildDSNFromConfig(connectionMain)

	// Determinar el nivel de logger
	var debug bool
	if isDebug, ok := foundation.App.Config.Get("app.debug", true).(bool); ok && isDebug {
		debug = isDebug
	}
	var logLevel logger.LogLevel
	if debug {
		logLevel = logger.Info // Modo debug
	} else {
		logLevel = logger.Silent // Modo no debug (solo errores)
	}

	db, err := gorm.Open(sqlserver.Open(dsnMain), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
			NoLowerCase:   true,
		},
	})
	if err != nil {
		color.Redf("%s: %s: %v\n", ErrConnectionFailed, defaultNameConnection, err)
		os.Exit(1)
	}

	// Configure connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		color.Redf("%s: %v\n", ErrFailedToGetSQLDB, err)
		os.Exit(1)
	}

	poolConfig, ok := foundation.App.Config.Get("database.pool", nil).(map[string]interface{})
	if ok {
		if maxIdleConns, exists := poolConfig["max_idle_conns"].(int); exists {
			sqlDB.SetMaxIdleConns(maxIdleConns)
		}
		if maxOpenConns, exists := poolConfig["max_open_conns"].(int); exists {
			sqlDB.SetMaxOpenConns(maxOpenConns)
		}
		if connMaxIdleTime, exists := poolConfig["conn_max_idletime"].(int); exists {
			sqlDB.SetConnMaxIdleTime(time.Duration(connMaxIdleTime) * time.Second)
		}
		if connMaxLifetime, exists := poolConfig["conn_max_lifetime"].(int); exists {
			sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)
		}
	}
	var resolver *dbresolver.DBResolver

	for nameConnection, connection := range connections {
		connectionMap, ok := connection.(map[string]interface{})
		if !ok {
			color.Redf("%s: %s\n", ErrBindingConnection, nameConnection)
			os.Exit(1)
		}

		if nameConnection != defaultNameConnection {
			dsn := buildDSNFromConfig(connectionMap)
			datas := connectionMap["datas"].([]interface{})
			datas = append(datas, nameConnection)

			// Verifica si resolver es nil y asigna directamente el primer register
			if resolver == nil {
				resolver = dbresolver.Register(dbresolver.Config{
					Sources:           []gorm.Dialector{sqlserver.Open(dsn)},
					Policy:            dbresolver.RandomPolicy{},
					TraceResolverMode: debug,
				}, datas...)
			} else {
				// Si resolver ya está inicializado, solo registra más configuraciones
				resolver = resolver.Register(dbresolver.Config{
					Sources:           []gorm.Dialector{sqlserver.Open(dsn)},
					Policy:            dbresolver.RandomPolicy{},
					TraceResolverMode: debug,
				}, datas...)
			}
		}
	}

	if resolver != nil {
		if err := db.Use(resolver); err != nil {
			color.Redf("Error registration: %v\n", err)
			os.Exit(1)
		} else {
			color.Greenf("Connections registered successfully.\n")
		}
	} else {
		color.Yellowf("No connections to register.\n")
	}

	DB = Database{Ctx: db}
}
