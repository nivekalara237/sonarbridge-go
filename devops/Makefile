default: deploy

ENV ?= dev
OVERLAYS := k8s/overlays/$(ENV)
NAMESPACE := sonarbridge-$(ENV)
DEPLOY_CONTAINER_IMAGE_NAME := sonarbridge-container-image
IMAGE_TAG ?= latest
SONAR_CA_ROOT ?= ""
GITLAB_CA_ROOT ?= ""

.PHONY: create-ns deploy rollback status render install-cert-manager uninstall-cert-manager set-image

create-ns:
	@echo "Preparing env by creating necessary namespaces..."
	kubectl create ns sonarbridge-dev && kubectl create ns sonarbridge-prod

render:
	@echo ">>> Rendu Kustomize pour $(ENV)..."
	cd $(OVERLAYS) && $(MAKE) set-image ENV=$(ENV) IMAGE_TAG=$(IMAGE_TAG)
	kubectl kustomize $(OVERLAYS)

status:
	@echo ">>> Attente que les pods soient prêts ($(ENV))..."
	kubectl rollout status -n $(NAMESPACE) --timeout=120s

set-image:
	@echo "Mettre ç jour l'image dans kustomize..."
	kubectl kustomize edit set image $(DEPLOY_CONTAINER_IMAGE_NAME)=$(IMAGE_TAG)

deploy:
	@echo ">>>1. Creation de la class ingress (Traefik Controller)..."
	kubectl apply -f k8s/ingress-controller.yml -n $(NAMESPACE)
	yq e -i '.images[0].newTag = "$(IMAGE_TAG)"' k8s/base/kustomization.yml

	@echo ">>> 2. Création des ConfigMap et Secret..."
	kubectl kustomize $(OVERLAYS) | kubectl apply -f - --selector="deploy-order=0"
	## kubectl create secret generic sonarbridge-external-ca-files -n ${NAMESPACE} --from-file=gilab.cavom.lan.crt=${GITLAB_CA_ROOT} --from-file=sonarqube.cavom.lan.crt=${SONAR_CA_ROOT} --dry-run=client -oyaml | kubectl apply -f -

	@echo ">>> 3. Application complète..."
	kubectl kustomize $(OVERLAYS) | kubectl apply -f - --selector="deploy-order!=0"


rollback:
	@echo ">>> Rollback sur $(ENV)..."
	kubectl rollback undp deployment -n $(NAMESPACE)

install-cert-manager:
	@echo ">>> Installation cert-manager..."
	kubectl apply -f k8s/tls/cert-manager.yml
	@echo ">>> Attente que cert-manager soit prêt..."
	kubectl wait --for=condition=Available deployment \
		--all -n cert-manager \
		--timeout=120s
	@echo "Deployment du Gestionnaire de Certificat tls - acme-cert-manager"
	kubectl apply -f k8s/tls/letsencrypt-prod.yml

uninstall-cert-manager:
	@echo ">>> Uninstalling lestencrypt-prod issuer..."
	kubectl delete -f k8s/tls/letsencrypt-prod.yml
	@echo ">>> Uninstalling cert-manager..."
	kubectl delete -f k8s/tls/cert-manager.yml