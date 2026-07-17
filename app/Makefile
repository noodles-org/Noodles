dev-backend:
	cd backend && npm run dev-server
dev-frontend:
	cd frontend && npm run dev
dev:
	$(MAKE) dev-backend & $(MAKE) dev-frontend & wait

install-all:
	cd backend && npm ci
	cd frontend && npm ci

update-dashboard:
	cd .. && docker buildx build --platform linux/amd64 -t mephalrith/noodles-dashboard .
	docker push mephalrith/noodles-dashboard:latest
	kubectl rollout restart -n dashboard deployment/noodles-dashboard
