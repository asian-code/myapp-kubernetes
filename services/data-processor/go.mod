module github.com/asian-code/myapp-kubernetes/services/data-processor

go 1.21

require (
	github.com/asian-code/myapp-kubernetes/services/shared v0.0.0
	github.com/sirupsen/logrus v1.9.3
	github.com/gorilla/mux v1.8.1
	github.com/jackc/pgx/v5 v5.5.1
	github.com/prometheus/client_golang v1.18.0
)

replace github.com/asian-code/myapp-kubernetes/services/shared => ../shared
