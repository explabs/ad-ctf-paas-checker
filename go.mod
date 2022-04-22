module github.com/explabs/ad-ctf-paas-checker

go 1.16

require (
	github.com/explabs/ad-ctf-paas-api v1.0.9
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/rabbitmq/amqp091-go v1.3.0
	go.mongodb.org/mongo-driver v1.7.3
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

//replace github.com/explabs/ad-ctf-paas-api => ../ad-ctf-paas-api
