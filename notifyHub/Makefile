.PHONY: up down logs ps kafka-topics kafka-groups run-api run-worker tidy

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

# Inspect Kafka from the host (broker must be up).
kafka-topics:
	docker exec notifyhub-kafka /opt/kafka/bin/kafka-topics.sh \
		--bootstrap-server localhost:9092 --list

kafka-groups:
	docker exec notifyhub-kafka /opt/kafka/bin/kafka-consumer-groups.sh \
		--bootstrap-server localhost:9092 --list

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

tidy:
	go mod tidy
