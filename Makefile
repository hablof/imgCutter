build:
	go mod vendor
	docker build -t imgcutter ./
run:
	docker run -d -p 8080:8080 --rm --name imgcutter imgcutter
stop:
	docker stop imgcutter