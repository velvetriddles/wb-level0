publish:
	@echo "Publishing random order..."
	@go run ./cmd/natspublish/publishValid/publish_order.go
app: 
	@go run ./cmd/app/main.go

publish-invalid:
	@echo "Publishing invalid order..."
	@go run ./cmd/natspublish/publishInvalid/publish_invalid_order.go

vegeta-run:
	@echo "Vegeta test is running..."
	@go run ./vegeta/vegeta.go

.PHONY:
	vegeta