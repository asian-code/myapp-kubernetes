# Deployment Guide

This guide covers the end-to-end process of deploying the myHealth platform. It combines infrastructure provisioning (Terraform) and application deployment (ArgoCD).

---

## Prerequisites

*   **AWS CLI** configured with Administrator credentials.
*   **Terraform** v1.6+ installed.
*   **kubectl** installed.
*   **Helm** v3+ installed.
*   **Domain Name** (e.g., `eric-n.com`) managed by a DNS provider (Route53, Cloudflare, etc.).

---

## Phase 1: Infrastructure (Terraform)

**Goal**: Provision the physical AWS resources (VPC, EKS, RDS, ECR, IAM).

1.  **Initialize & Apply**:
    ```bash
    cd terraform
    terraform init
    terraform apply
    ```
    *Review the plan and type `yes`.*

2.  **DNS & SSL Validation (Critical)**:
    *   Terraform will output `acm_validation_records`.
    *   **Action**: Add these CNAME records to your DNS provider to validate the SSL certificate.
    *   **Action**: Note the `acm_certificate_arn` output.

3.  **Record Outputs**:
    Save the following outputs for later steps:
    *   `eks_cluster_name`
    *   `rds_endpoint`
    *   `ecr_registry_id`

---

## Phase 2: Cluster Configuration

**Goal**: Prepare the Kubernetes cluster for the application.

1.  **Connect to Cluster**:
    ```bash
    aws eks update-kubeconfig --name <eks_cluster_name> --region us-east-1
    ```

2.  **Install External Secrets Operator (ESO)**:
    The application needs ESO to fetch secrets from AWS Secrets Manager.
    ```bash
    helm repo add external-secrets https://charts.external-secrets.io
    helm install external-secrets external-secrets/external-secrets -n external-secrets --create-namespace
    ```

3.  **Populate Secrets**:
    Go to the **AWS Secrets Manager Console** and update the secrets created by Terraform:
    *   `myhealth/db-credentials`: Ensure `host`, `username`, `password`, `dbname` are correct (use `rds_endpoint`).
    *   `myhealth/oura-credentials`: Add your Oura Ring API keys.

---

## Phase 3: GitOps Setup (ArgoCD)

**Goal**: Install ArgoCD to manage the application lifecycle.

1.  **Install ArgoCD**:
    ```bash
    kubectl create namespace argocd
    kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
    ```

2.  **Deploy the "Root App"**:
    This triggers the "App of Apps" pattern, deploying the entire platform.
    ```bash
    kubectl apply -f argocd/root-app.yaml
    ```

---

## Phase 4: Verification

1.  **Check ArgoCD**:
    Port-forward the UI to see the sync status.
    ```bash
    kubectl port-forward svc/argocd-server -n argocd 8080:443
    ```
    *Login with `admin` and the initial password (found in `argocd-initial-admin-secret`).*

2.  **Check Pods**:
    ```bash
    kubectl get pods -n myhealth
    ```
    You should see `api-service`, `data-processor`, and `oura-collector` running.

3.  **Check Ingress**:
    Get the Load Balancer URL:
    ```bash
    kubectl get svc -n istio-system istio-ingressgateway
    ```
    *Create a CNAME record in your DNS provider pointing `api.eric-n.com` to this Load Balancer.*

---

## Troubleshooting

*   **Pods Pending**: Check `kubectl describe pod <pod>` for resource issues or PVC binding errors.
*   **Secrets Missing**: Check `kubectl get externalsecrets` and `kubectl get secretstore`. Ensure the IAM role for ESO is correct.
*   **ArgoCD Sync Failed**: Check the ArgoCD UI for error messages (often related to CRDs or missing Helm values).
