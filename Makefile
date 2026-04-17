.PHONY: proto build run-order run-payment run-stream db-up db-down

proto:
	@echo "Generating Go code from proto files..."
	protoc \
		--go_out=gen --go_opt=module=github.com/MegaSlime-2407/generated \
		--go-grpc_out=gen --go-grpc_opt=module=github.com/MegaSlime-2407/generated \
		proto/payment/payment.proto proto/order/order.proto
	@echo "Done."

build:
	cd order-service && go build -o ../bin/order-service ./cmd/order-service
	cd payment-service && go build -o ../bin/payment-service ./cmd/payment-service
	cd cmd/stream-client && go build -o ../../bin/stream-client .

db-up:
	docker-compose up -d

db-down:
	docker-compose down

run-payment:
	cd payment-service && go run ./cmd/payment-service

run-order:
	cd order-service && go run ./cmd/order-service

run-stream:
	cd cmd/stream-client && go run . $(ORDER_ID)
