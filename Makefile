.PHONY: up down build logs \
	k8s-load \
	k8s-namespace k8s-infra k8s-app k8s-job k8s-deploy k8s-redeploy \
	k8s-clean-jobs k8s-clean-app k8s-clean-infra k8s-clean-pvc k8s-clean k8s-nuke \
	k8s-status k8s-logs port-forward

# ─── Docker Compose (local dev) ───────────────────────────────────────────────

up:
	docker compose --env-file deploy/.env.docker -f deploy/docker-compose.yml up --build

down:
	docker compose --env-file deploy/.env.docker -f deploy/docker-compose.yml down

build:
	docker compose --env-file deploy/.env.docker -f deploy/docker-compose.yml build

logs:
	docker compose --env-file deploy/.env.docker -f deploy/docker-compose.yml logs -f deepforge-app

# ─── K8s Image Management ─────────────────────────────────────────────────────

k8s-load:
	docker build -t deepforge:0.1.0 -f deploy/Dockerfile .
	docker save deepforge:0.1.0 | sudo k0s ctr images import -

# ─── K8s Setup ────────────────────────────────────────────────────────────────

k8s-namespace:
	kubectl apply -f deploy/k8s/namespace.yaml

k8s-infra: k8s-namespace
	kubectl apply -R -f deploy/k8s/infra

k8s-app: k8s-namespace
	kubectl apply -f deploy/k8s/configmap.yaml
	kubectl apply -f deploy/k8s/secret.yaml

k8s-job:
	kubectl delete job deepforge -n deepforge --ignore-not-found
	kubectl apply -f deploy/k8s/job.yaml

# Deploy everything — infra + app config + job
k8s-deploy: k8s-infra k8s-app k8s-job

# Rebuild image and redeploy everything
k8s-redeploy: k8s-load k8s-deploy

# ─── K8s Cleanup ──────────────────────────────────────────────────────────────

k8s-clean-jobs:
	kubectl delete jobs -n deepforge --all --ignore-not-found

k8s-clean-app:
	kubectl delete -f deploy/k8s/configmap.yaml --ignore-not-found
	kubectl delete -f deploy/k8s/secret.yaml --ignore-not-found

k8s-clean-infra:
	kubectl delete -R -f deploy/k8s/infra --ignore-not-found

k8s-clean-pvc:
	kubectl delete pvc -n deepforge --all --ignore-not-found

# Clean everything except PVCs and namespace — Tempo/Grafana data survives
k8s-clean: k8s-clean-jobs k8s-clean-app k8s-clean-infra

# Full wipe — PVCs and namespace included, completely clean slate
k8s-nuke: k8s-clean k8s-clean-pvc
	kubectl delete namespace deepforge --ignore-not-found

# ─── K8s Observability ────────────────────────────────────────────────────────

k8s-status:
	kubectl get all -n deepforge

k8s-logs:
	kubectl logs -n deepforge -l app=deepforge --tail=100 -f

port-forward:
	kubectl port-forward -n deepforge svc/mailhog-service 8025:8025 &
	kubectl port-forward -n deepforge svc/grafana-service 3000:3000 &