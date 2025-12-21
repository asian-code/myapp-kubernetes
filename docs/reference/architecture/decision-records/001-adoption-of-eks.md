# ADR 001: Adoption of Amazon EKS

**Date**: 2025-12-21  
**Status**: Accepted

## Context
We need a container orchestration platform to host our microservices (api-service, data-processor, oura-collector). The platform must support high availability, scalability, and integration with AWS services.

## Decision
We have decided to use **Amazon Elastic Kubernetes Service (EKS)**.

## Consequences
*   **Positive**: 
    *   Managed control plane reduces operational overhead.
    *   Industry standard for container orchestration.
    *   Rich ecosystem of tools (Helm, ArgoCD, Prometheus).
    *   Seamless integration with AWS IAM for security (IRSA).
*   **Negative**:
    *   Higher complexity compared to ECS or App Runner.
    *   Requires Kubernetes expertise.
    *   Cost of the control plane ($73/month) + worker nodes.

## Alternatives Considered
*   **Amazon ECS**: Simpler, but less portable and has a smaller ecosystem of open-source tools.
*   **AWS App Runner**: Too opinionated for our complex networking and background processing needs.
