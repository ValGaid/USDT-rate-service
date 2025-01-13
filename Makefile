
lint:
	golangci-lint run --disable-all \
		--enable=revive \
		--enable=errcheck \
		--enable=staticcheck \
		--enable=unused \
		--enable=gofmt

test:
	go test ./... -v

run-f:
	go run cmd/main.go \
		-host=localhost \
		-port=50051 \
		-log-level=DEBUG \
		-trace-host=localhost \
		-trace-port=14268 \
		-db-host=localhost \
		-db-port=5432 \
		-db-user=postgres \
		-db-password=password \
		-db-name=postgres \
		-db-sslmode=disable


run-e:
	export HOST=localhost && \
	export PORT=50051 && \
	export TRACE_HOST=localhost && \
	export TRACE_PORT=14268 && \
	export DB_HOST=localhost && \
	export DB_PORT=5432 && \
	export DB_USER=postgres && \
	export DB_PASSWORD=password && \
	export DB_NAME=postgres && \
	export DB_SSLMODE=disable && \
	export GF_SECURITY_ADMIN_PASSWORD=admin && \
	go run cmd/*.go

run-jaeger:
	docker run -d --name jaeger \
		-e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
		-p 5775:5775/udp \
		-p 6831:6831/udp \
		-p 6832:6832/udp \
		-p 5778:5778 \
		-p 16686:16686 \
		-p 14268:14268 \
		-p 14250:14250 \
		-p 9411:9411 \
		jaegertracing/all-in-one:1.38
build:
	go build -o ./usdt cmd/*.go


.PHONY: run-prometheus stop-prometheus

run-prometheus:
	docker run -d --name prometheus \
		-p 9090:9090 \
		-v ./prometheus.yml:/etc/prometheus/prometheus.yml \
		--add-host=host.docker.internal:host-gateway \
		prom/prometheus:latest

stop-prometheus:
	docker stop prometheus || true
	docker rm prometheus || true

up:
	docker-compose up
