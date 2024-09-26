
dev:
	air

docker.chi.build:
	docker build -t chi .

docker-chi: docker.chi.build
	docker run --rm -d \
		--name presigned-multipart-upload-server \
		-p 8080:8080 \
		chi