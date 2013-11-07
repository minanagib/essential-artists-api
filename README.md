essential-artists-api
=====================

This project simply prototypes the process/workflow of leveraging RabbitMQ, PHP and GO to fulfill 
merch orders integrating with essential artists API

[Requirements]
[System]
- Go
- PHP
- nix/mac os
- RabbitMQ

[Go 3rd party libraries]
- AMQP Client https://github.com/streadway/amqp


[PHP 3rd party libraries]
- AMQP Client https://github.com/videlalvaro/php-amqplib



[Source Code]
- src/consumer.go - AMQP client responsible for serializing JSON messages from RabbitMQ and submitting orders to the third-party API
- src/server-essential-artists-api/fake_server.go - HTTP server that accepts requests from the consumer and replies back with a SOAP formatted result
