.PHONY: dev backend web

dev:
	@trap 'kill 0' INT TERM; \
	( cd xendit && go run . ) & \
	( cd web && python3 -m http.server 5500 ) & \
	wait

backend:
	cd xendit && go run .

web:
	cd web && python3 -m http.server 5500
