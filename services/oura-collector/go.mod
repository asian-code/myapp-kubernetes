module github.com/asian-code/myapp-kubernetes/services/oura-collector

go 1.21

require (
	github.com/asian-code/myapp-kubernetes/services/shared v0.0.0
	github.com/asian-code/myapp-kubernetes/services/pkg v0.0.0
	github.com/sirupsen/logrus v1.9.3
	github.com/joho/godotenv v1.5.1
)

replace github.com/asian-code/myapp-kubernetes/services/shared => ../shared
replace github.com/asian-code/myapp-kubernetes/services/pkg => ../pkg
