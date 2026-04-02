.PHONY: setup up down logs build shell-backend shell-frontend

setup:
	@if [ ! -f backend/.env ]; then \
		cp backend/.env.example backend/.env; \
		echo "Created backend/.env from .env.example"; \
	fi

up: setup
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

build:
	docker-compose build

shell-backend:
	docker exec -it kafe-api sh

shell-frontend:
	docker exec -it kafe-web sh

restart: down up
