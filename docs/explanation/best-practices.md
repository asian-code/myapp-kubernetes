# Golden Standards & Best Practices (Late 2025 Edition)

This document outlines the architectural decisions, tooling choices, and "golden standards" adopted for the myHealth Kubernetes Platform. It reflects the state of the art in Platform Engineering, Cloud Native development, and Security as of late 2025.

---

## 1. Platform Engineering & Kubernetes

### The Shift to Platform Engineering
We have moved beyond traditional "DevOps" to **Platform Engineering**. This platform is treated as a product, providing "Golden Paths" for developers to self-serve infrastructure without needing deep Kubernetes expertise.

### Kubernetes (EKS) Standards
*   **EKS Auto Mode & Karpenter**: We utilize **Karpenter** for just-in-time, high-performance node provisioning. This replaces the legacy Cluster Autoscaler, allowing us to schedule the right compute resources (Spot/On-Demand) based on pending pod requirements within seconds.
*   **eBPF Networking (Cilium)**: We use **Cilium** as our CNI. It leverages eBPF to provide high-performance networking, transparent encryption (WireGuard), and deep observability without the overhead of sidecars.
*   **Namespace Isolation**: Strict multi-tenancy is enforced via Namespaces, ResourceQuotas, and NetworkPolicies.

### Security First
*   **Supply Chain Security**:
    *   **Provenance**: All artifacts are signed using **Cosign (Sigstore)** to ensure provenance.
    *   **Scanning**: Continuous vulnerability scanning is performed by **Trivy** in the CI pipeline and inside the cluster.
*   **Secrets Management**: We strictly avoid native Kubernetes Secrets for sensitive data. We use the **External Secrets Operator (ESO)** to sync secrets from AWS Secrets Manager directly into memory-backed volumes or Kubernetes secrets only when necessary.
*   **Runtime Security**: **Tetragon** (eBPF-based) is used for runtime enforcement, detecting and blocking abnormal process executions or syscalls.

---

## 2. Microservices Architecture (Go)

### Go Language Patterns (Go 1.24+)
*   **Standard Library Routing**: We leverage the enhanced `net/http` `ServeMux` introduced in Go 1.22+, avoiding heavy third-party frameworks.
*   **Project Layout**: We adhere to the **Standard Go Project Layout**:
    *   `cmd/`: Main applications.
    *   `internal/`: Private application and business logic.
    *   `pkg/`: Library code safe for external use.
*   **Domain-Driven Design (DDD)**: Services are structured around business domains, strictly separating transport layers (HTTP/gRPC) from core business logic.

### Observability (OpenTelemetry)
*   **OTel Everywhere**: **OpenTelemetry (OTel)** is the single standard for traces, metrics, and logs. We do not use vendor-specific agents.
*   **Distributed Tracing**: Trace context propagation is mandatory across all synchronous (HTTP/gRPC) and asynchronous (Queue) boundaries.
*   **Golden Signals**: All services must emit the four golden signals: Latency, Traffic, Errors, and Saturation.

### Service Mesh
*   **Ambient Mesh**: We adopt **Istio Ambient Mesh** (sidecar-less mode). This moves Layer 4-7 processing to a per-node ztunnel and Waypoint proxy, significantly reducing resource cost and operational complexity compared to the traditional sidecar model.

---

## 3. GitOps & Continuous Delivery

### ArgoCD & Progressive Delivery
*   **App of Apps Pattern**: We use ArgoCD ApplicationSets to dynamically manage our multi-cluster, multi-environment fleet.
*   **Progressive Delivery**: Critical services use **Argo Rollouts** for Canary deployments. We automate traffic shifting based on real-time Prometheus metrics (e.g., error rates), automatically rolling back if health checks fail.
*   **GitOps Flow**:
    *   **App Repo**: Source code + Helm Chart.
    *   **Config Repo**: Environment-specific values (Helm values.yaml).
    *   Changes to `main` in App Repo trigger a CI build -> push new image -> update Config Repo -> ArgoCD syncs.

### CI Pipelines (GitHub Actions)
*   **Reusable Workflows**: CI logic is centralized in a governance repository. Services call reusable workflows to ensure consistent quality gates (linting, testing, security scans).
*   **Ephemeral Runners**: We use ephemeral runners to ensure a clean build environment and improve security.

---

## 4. Infrastructure as Code (Terraform/OpenTofu)

### Modern IaC Practices
*   **OpenTofu Compatibility**: All modules are compatible with OpenTofu.
*   **Testing**: Infrastructure is software. We use `terraform test` and **Terratest** to validate modules before they are released.
*   **State Isolation**: State is strictly isolated by **Environment** (Prod/Staging) and **Layer** (Network/Data/Compute) to minimize blast radius.
*   **Module Composition**: We avoid "God Modules". We build small, focused modules (`aws-eks`, `aws-rds`) and compose them in the `live/` directory.

---

## 5. Documentation Strategy

### The "Diátaxis" Framework
We structure our documentation to solve the "wall of text" problem by categorizing docs into four quadrants:
1.  **Tutorials**: Learning-oriented lessons (e.g., "Deploy your first microservice").
2.  **How-to Guides**: Problem-oriented steps (e.g., "How to rotate database credentials").
3.  **Reference**: Information-oriented specs (e.g., API Swagger docs, Env Var lists).
4.  **Explanation**: Understanding-oriented background (e.g., "Why we chose EKS over ECS").

### Internal Developer Portal (IDP)
*   **Backstage**: We use Backstage as our IDP.
*   **TechDocs**: Documentation lives **with the code** (Markdown in `/docs` folder of the repo) and is rendered centrally in Backstage.
*   **Software Templates**: Developers create new services using approved Backstage templates that scaffold the repo with all best practices (Dockerfile, Helm chart, CI workflows) pre-configured.
