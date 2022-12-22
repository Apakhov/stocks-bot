build:
	docker-compose build --no-cache 

down:
	docker-compose down --rmi all -v

up:
	docker-compose up -d

logs:
	docker-compose logs -f

reload: down build up logs

connect_bot:
	docker-compose run bot sh

connect_stockserver:
	docker-compose run stockserver sh

