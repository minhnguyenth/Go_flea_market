package infra

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	dbConnections        prometheus.Gauge
	dbMaxOpenConnections prometheus.Gauge
	dbIdleConnections    prometheus.Gauge
	dbInUseConnections   prometheus.Gauge
)

func init() {
	dbConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_total",
			Help: "Total number of database connections in the pool",
		},
	)
	dbMaxOpenConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_max_open_connections",
			Help: "Maximum number of open connections to the database",
		},
	)
	dbIdleConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_idle_connections",
			Help: "Number of idle connections in the pool",
		},
	)
	dbInUseConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_in_use_connections",
			Help: "Number of connections currently in use",
		},
	)

	prometheus.MustRegister(dbConnections)
	prometheus.MustRegister(dbMaxOpenConnections)
	prometheus.MustRegister(dbIdleConnections)
	prometheus.MustRegister(dbInUseConnections)
}

func updateDBConnections(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}

	dbConnections.Set(float64(sqlDB.Stats().MaxOpenConnections))
	dbMaxOpenConnections.Set(float64(sqlDB.Stats().MaxOpenConnections))
	dbIdleConnections.Set(float64(sqlDB.Stats().Idle))
	dbInUseConnections.Set(float64(sqlDB.Stats().InUse))
}

func SetupDB() *gorm.DB {
	dns := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable timezone=Asia/Tokyo",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"))

	log.Println(dns)
	db, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to get database connection")
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Update connection statistics periodically
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		for range ticker.C {
			updateDBConnections(db)
		}
	}()

	return db
}
