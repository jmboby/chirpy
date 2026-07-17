APP_URL := http://localhost:8080

.PHONY: test test-validatevalid test-validate-too-long test-validate-invalid-json test-metrics sql-regenerate

test: test-validate-valid test-validate-too-long test-validate-invalid-json test-metrics

test-validate-valid:
	curl -i -X POST $(APP_URL)/api/validate_chirp \
		-H "Content-Type: application/json" \
		-d '{"body":"This is valid"}'

test-validate-too-long:
	curl -i -X POST $(APP_URL)/api/validate_chirp \
		-H "Content-Type: application/json" \
		-d '{"body":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}'

test-validate-invalid-json:
	curl -i -X POST $(APP_URL)/api/validate_chirp \
		-H "Content-Type: application/json" \
		-d '{"body":'

test-metrics:
	curl -i $(APP_URL)/admin/metrics

test-app:
	curl -i $(APP_URL)/app/

test-metrics-reset:
	curl -iX POST $(APP_URL)/admin/reset 

sql-regenerate:
	sqlc generate
