module github.com/asian-code/myapp-kubernetes/services/api-service

go 1.21

require (
	github.com/asian-code/myapp-kubernetes/services/shared v0.0.0
	github.com/sirupsen/logrus v1.9.3
	github.com/gorilla/mux v1.8.1
	github.com/jackc/pgx/v5 v5.5.1
	github.com/prometheus/client_golang v1.18.0
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/rs/cors v1.10.1
	golang.org/x/crypto v0.17.0
)

replace github.com/asian-code/myapp-kubernetes/services/shared => ../shared
