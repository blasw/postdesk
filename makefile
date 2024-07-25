.PHONY: entry-up entry-down entry-r test

# disables kafka's and zookeeper's logs
up:
	docker-compose up --no-attach kafka --no-attach zookeeper --build

down:
	docker-compose down

re:
	docker-compose down && docker-compose up --no-attach kafka --no-attach zookeeper --build

test:
	cd tester && gotestsum --format=short-verbose