module github.com/explabs/ad-ctf-paas-checker

go 1.16

require (
	github.com/explabs/ad-ctf-paas-api v1.0.8-0.20220222195202-059a960a5816
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/rabbitmq/amqp091-go v1.3.0
	go.mongodb.org/mongo-driver v1.7.3
)

//replace github.com/explabs/ad-ctf-paas-api => ../ad-ctf-paas-api
