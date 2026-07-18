dev-backend:
	cd backend && go run ./cmd/server
dev-frontend:
	cd frontend && npm run dev
dev:
	$(MAKE) dev-backend & $(MAKE) dev-frontend & wait

build-backend:
	cd backend && CGO_ENABLED=0 go build -o server ./cmd/server
test-backend:
	cd backend && go test ./...

install-all:
	cd backend && go mod download
	cd frontend && npm ci

update-dashboard:
	cd .. && docker buildx build --platform linux/amd64 -t mephalrith/noodles-dashboard .
	docker push mephalrith/noodles-dashboard:latest
	kubectl rollout restart -n dashboard deployment/noodles-dashboard
