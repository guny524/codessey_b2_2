# Auth is copied INTO the isolated codex_home volume after the containers start
# (one-way: host -> container). The host ~/.codex is never mounted, so a
# misbehaving container can't touch it. CODEX_DIR is just the copy source;
# override with `make up CODEX_DIR=/path/to/.codex`.
CODEX_DIR ?= $(HOME)/.codex

up:
	docker compose up -d
	docker compose cp "$(CODEX_DIR)/auth.json" image_proxy:/root/.codex/auth.json

down:
	docker compose down -v

.PHONY: up down
