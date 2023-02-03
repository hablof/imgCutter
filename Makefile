build:
	go mod vendor
	docker build -t hablof/imgcutter ./
run:
	docker run -d -p 8080:8080 --rm --name imgcutter hablof/imgcutter
stop:
	docker stop imgcutter