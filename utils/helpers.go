package helpers

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-pg/pg"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// Dial ... connect to grpc
func Dial(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return conn
}

// CorsHandler ... handle cors
func CorsHandler(router http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:     []string{"*"},
		AllowCredentials:   true,
		AllowedMethods:     []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Authorization"},
		OptionsPassthrough: false,
	})

	return c.Handler(router)
}

// Wait ... Wait for a port to be avilable
func Wait(addr string, log *logrus.Entry) {
	start := time.Now()
	retries := 0
	for {
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err == nil {
			conn.Close()
			log.WithFields(logrus.Fields{
				"waiting": addr,
				"status":  "avilable",
				"took":    time.Since(start),
				"retries": retries,
			}).Info("Service is now avilable")
			return
		}
		retries = retries + 1
		log.WithFields(logrus.Fields{
			"waiting": addr,
			"status":  "waiting",
			"retries": retries,
		}).Info("Waiting for service")
		time.Sleep(5 * time.Second)
	}
}

// Logger ... Return a new configurated logrus logger
func Logger(name string) *logrus.Entry {
	log := logrus.New()

	log.SetLevel(logrus.InfoLevel)

	file, err := os.OpenFile("/logs/"+name+".log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}

	contextLogger := log.WithFields(logrus.Fields{
		"service": name,
	})

	return contextLogger
}

// QueryLogger ... A query logger for postgres
type QueryLogger func(*pg.QueryProcessedEvent)

// CreateQueryLogger ... Create a new query logger for pg
func CreateQueryLogger(logger *logrus.Entry) QueryLogger {
	return func(event *pg.QueryProcessedEvent) {
		query, err := event.FormattedQuery()
		if err != nil {
			panic(err)
		}
		logger.WithFields(logrus.Fields{
			"took":  time.Since(event.StartTime),
			"query": query,
		}).Info("Executed Query")
	}
}
