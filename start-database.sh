#!/run/current-system/sw/bin/bash
podman run \
	--name fs-postgres \
	-e POSTGRES_PASSWORD=letmein \
	-e POSTGRES_USER=fieldseeker \
	-p 5432:5432 \
	--rm \
	-v ./database:/var/lib/postgresql/data \
	-d \
	docker.io/postgres:17.5 \
	postgres -c log_statement=all
