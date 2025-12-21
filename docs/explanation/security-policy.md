# Security Policy & Stance

**Last Updated**: December 2025

This document outlines the security posture of the myHealth platform, covering infrastructure, application, and supply chain security.

---

## 1. Infrastructure Security (AWS & EKS)

### Network Isolation
*   **VPC**: All resources reside in a custom VPC.
*   **Private Subnets**: Worker nodes and databases are strictly in private subnets with no direct internet access.
*   **Public Access**: The EKS Control Plane public endpoint is restricted to specific allow-listed IPs (VPN/Office).

### Identity & Access Management (IAM)
*   **IRSA (IAM Roles for Service Accounts)**: Pods assume IAM roles via OIDC. No long-lived AWS credentials (access keys) are ever stored in the cluster.
*   **Least Privilege**: Roles are scoped to the exact permissions needed (e.g., `s3:GetObject` on a specific bucket).

### Encryption
*   **At Rest**: All EBS volumes, RDS databases, and S3 buckets are encrypted with AWS KMS.
*   **In Transit**: TLS 1.3 is enforced for all ingress traffic. Service-to-service communication is encrypted via Cilium (WireGuard) or Istio mTLS.
*   **Secrets**: Kubernetes Secrets are encrypted at rest using KMS envelope encryption.

---

## 2. Supply Chain Security

### Provenance & Signing
*   **Cosign**: All container images pushed to ECR are signed using Cosign (Sigstore).
*   **Verification**: The cluster uses an admission controller (Kyverno/Gatekeeper) to verify image signatures before allowing pods to start.

### Vulnerability Scanning
*   **CI Pipeline**: Trivy scans source code and container images during the build process. Builds fail on `CRITICAL` or `HIGH` vulnerabilities.
*   **Runtime**: Trivy Operator scans running images in the cluster daily.

---

## 3. Application Security

### Authentication & Authorization
*   **OIDC**: User authentication is handled via an external OIDC provider (e.g., Auth0, Cognito).
*   **RBAC**: Internal service access is controlled via Kubernetes RBAC.

### Runtime Protection
*   **ReadOnly Root Filesystem**: All pods run with a read-only root filesystem where possible.
*   **Non-Root User**: All containers run as a non-root user (UID > 1000).
*   **Capabilities**: All Linux capabilities are dropped (`ALL`), with only essential ones added back explicitly.

---

## 4. Incident Response

*   **Logs**: All logs are shipped to CloudWatch/OpenSearch and retained for 90 days.
*   **Audit**: Kubernetes Audit Logs are enabled and archived to S3 for compliance.
