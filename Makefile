run:
	@docker compose up --build

clear: 
	@docker compose down --volumes --remove-orphans