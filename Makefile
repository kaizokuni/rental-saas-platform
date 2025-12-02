.PHONY: dev deploy down logs db-shell backup

dev:
	docker compose up -d

deploy:
	docker compose -f docker-compose.prod.yml up -d --build --remove-orphans

down:
	docker compose down
	@echo "To stop production: docker compose -f docker-compose.prod.yml down"

logs:
	docker compose -f docker-compose.prod.yml logs -f app

db-shell:
	docker exec -it rental_postgres psql -U postgres -d rental_saas

backup:
	./scripts/backup.sh
