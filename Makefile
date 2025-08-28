MODEL_VECTOR_SIZE ?= 1024
CONTAINER_NAME ?= my-postgres
SQL_FILE ?= /Users/abhayyadav/Desktop/appleNotesRag/migrations/001_init.sql

migrate:
	sed -i '' "s/VECTOR([0-9]\+)/VECTOR($(MODEL_VECTOR_SIZE))/g" $(SQL_FILE)
	docker cp $(SQL_FILE) $(CONTAINER_NAME):/001_init.sql
	docker exec -it $(CONTAINER_NAME) psql -U myuser -d mydatabase -f /001_init.sql

clean-model:
	ollama rm $(MODEL)

.PHONY: migrate clean-model