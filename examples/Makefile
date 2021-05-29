postgres-master-slave-start:
	docker-compose up -d

postgres-master-slave-stop:
	docker-compose down

goose.up:
	goose --dir ./databases postgres "user=postgres password=my_password host=127.0.0.1 port=32780 dbname=my_database sslmode=disable" up

goose.down:
	goose --dir ./databases postgres "user=postgres password=my_password host=127.0.0.1 port=32780 dbname=my_database sslmode=disable" down
