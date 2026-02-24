.PHONY: build run clean docker-up docker-down docker-logs load-test

build:
	go build -o bin/server ./cmd/root.go

run:
	go run ./cmd/root.go

clean:
	rm -rf bin/
	rm -f loadtest/results.json

docker-up:
	docker compose -f docker-compose.yml up -d

docker-down:
	docker compose -f docker-compose.yml down

docker-logs:
	docker compose -f docker-compose.yml logs -f

load-test:
	docker compose -f docker-compose.yml --profile test up k6

test-event:
	@curl -X POST http://localhost:8080/events \
		-H "Content-Type: application/json" \
		-d '{"event_name":"test_event","user_id":"user123","timestamp":'$$(date +%s)',"channel":"web","tags":["test"]}'

test-metrics:
	@curl "http://localhost:8080/metrics?event_name=test_event"

health:
	@curl http://localhost:8080/health

ready:
	@curl http://localhost:8080/ready
