# sops key is ~/.config/sops/age/keys.txt by default
sops-encrypt-all:
	sops --encrypt playbooks/.secrets/restic-ansible-vault.yaml > playbooks/.secrets/restic-ansible-vault.enc.yaml && \
	sops --encrypt k8s/foundry/.secret.yaml > .secrets/foundry.enc.yaml && \
	sops --encrypt k8s/monitoring/.secret.yaml > .secrets/monitoring.enc.yaml && \
	sops --encrypt .secrets/escrow > .secrets/escrow.enc && \
	sops --encrypt helm/grafana-loki/values.yaml > .secrets/grafana-loki-values.enc.yaml && \
	sops --encrypt helm/grafana-alloy/values.yaml > .secrets/grafana-alloy-values.enc.yaml

sops-decrypt-all:
	sops --decrypt playbooks/.secrets/restic-ansible-vault.enc.yaml > playbooks/.secrets/restic-ansible-vault.yaml && \
	sops --decrypt .secrets/foundry.enc.yaml > k8s/foundry/.secret.yaml && \
	sops --decrypt .secrets/monitoring.enc.yaml > k8s/monitoring/.secret.yaml && \
	sops --decrypt .secrets/escrow.enc > .secrets/escrow && \
	sops --decrypt .secrets/grafana-loki-values.enc.yaml > helm/grafana-loki/values.yaml && \
	sops --decrypt .secrets/grafana-alloy-values.enc.yaml > helm/grafana-alloy/values.yaml

setup-cluster:
	ansible-playbook -i inventory.ini playbooks/setup_cluster.yaml --extra-vars "config=~/.kube/foundry_config"

manual-backup:
	kubectl apply -f k8s/foundry/jobs/manual/manual-backup.yaml
manual-restore-latest:
	kubectl apply -f k8s/foundry/jobs/manual/restore-from-latest.yaml

mount-archive:
	ansible-playbook -i inventory.ini playbooks/mount_archive.yaml
unmount-archive:
	umount ~/mnt && rm -r ~/mnt
manual-download-archive:
	ansible-playbook -i inventory.ini playbooks/download_archive_nas.yaml --extra-vars "destination=~/Downloads"