.PHONY: up down logs ps migrate kafka-topics kafka-groups kafka-describe-group kafka-dlq run-api run-worker tidy

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

migrate:
	@for f in migrations/*.sql; do \
		echo "Applying $$f"; \
		docker exec -i notifyhub-postgres psql -U notifyhub -d notifyhub < $$f; \
	done

# Inspect Kafka from the host (broker must be up).
kafka-topics:
	docker exec notifyhub-kafka /opt/kafka/bin/kafka-topics.sh \
		--bootstrap-server localhost:9092 --list

kafka-groups:
	docker exec notifyhub-kafka /opt/kafka/bin/kafka-consumer-groups.sh \
		--bootstrap-server localhost:9092 --list

kafka-describe-group:
	docker exec notifyhub-kafka /opt/kafka/bin/kafka-consumer-groups.sh \
		--bootstrap-server localhost:9092 \
		--describe --group notifyhub-worker

# Peek at dead-letter messages (newest first from beginning for local debug).
kafka-dlq:
	docker exec notifyhub-kafka /opt/kafka/bin/kafka-console-consumer.sh \
		--bootstrap-server localhost:9092 \
		--topic notifications.dlq \
		--from-beginning \
		--timeout-ms 3000

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

tidy:
	go mod tidy
