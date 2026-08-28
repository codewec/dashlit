.PHONY: frontend backend build run dev clean

DATA_DIR ?= ./data
export DATA_DIR

frontend:
	cd frontend && npm install && npm run build
	rm -rf backend/cmd/server/static
	mkdir -p backend/cmd/server/static
	cp -r frontend/dist/* backend/cmd/server/static/

backend:
	cd backend && go build -o ../app ./cmd/server

build: frontend backend

run: build
	./app

dev-backend:
	cd backend && DEV_MODE=1 go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

seed:
	cd backend && go run ./cmd/seed

clean:
	rm -rf frontend/dist app data/*.db
	mkdir -p backend/cmd/server/static
	find backend/cmd/server/static -mindepth 1 ! -path backend/cmd/server/static/.gitkeep -delete
	touch backend/cmd/server/static/.gitkeep
