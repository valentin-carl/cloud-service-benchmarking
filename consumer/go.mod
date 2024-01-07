module consumer

go 1.21

require (
	benchmark/lib v0.0.0-00010101000000-000000000000
	github.com/rabbitmq/amqp091-go v1.9.0
)

replace benchmark/lib => ../lib
