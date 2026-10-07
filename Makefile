APP_URL := http://localhost:8080

.PHONY: test test-validate-valid test-validate-too-long test-validate-invalid-json test-metrics test-app test-metrics-reset test-chirp-flow sql-regenerate

test: test-app test-validate-valid test-validate-too-long test-validate-invalid-json test-metrics test-metrics-reset test-chirp-flow

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

test-chirp-flow:
	USER_ID=$$(curl -s -X POST $(APP_URL)/api/users \
			-H 'Content-Type: application/json' \
			-d "{\"email\":\"$$(date +%s)@example.com\"}" | jq -r .id); \
	CHIRP_ID=$$(curl -s -X POST $(APP_URL)/api/chirps \
			-H 'Content-Type: application/json' \
			-d "{\"body\":\"hello\",\"user_id\":\"$$USER_ID\"}" | jq -r .id); \
	echo "user=$$USER_ID chirp=$$CHIRP_ID"; \
	curl -i $(APP_URL)/api/chirps/$$CHIRP_ID; \
	curl -i $(APP_URL)/api/chirps/00000000-0000-0000-0000-000000000000; \
	curl -i $(APP_URL)/api/chirps/notauuid

sql-regenerate:
	sqlc generate
