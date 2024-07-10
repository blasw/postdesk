.PHONY: entry-up entry-down entry-r test

entry-up:
	docker-compose -f docker-compose-entry.yaml up --build

entry-down:
	docker-compose -f docker-compose-entry.yaml down

entry-r:
	docker-compose -f docker-compose-entry.yaml down && 	docker-compose -f docker-compose-entry.yaml up --build

test:
	cd tester && gotestsum --format=short-verbose