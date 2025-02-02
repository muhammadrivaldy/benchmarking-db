create-migration:
	@read -p "input the migration name: " name; \
	migrate create -ext sql -dir migrations $$name