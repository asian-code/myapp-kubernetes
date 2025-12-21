# ADR 003: GitOps with ArgoCD

**Date**: 2025-12-21  
**Status**: Accepted

## Context
We need a reliable, auditable, and automated way to deploy applications to our Kubernetes clusters. We want to avoid manual `kubectl apply` commands and ensure the cluster state always matches our git repository.

## Decision
We have decided to use **ArgoCD** for GitOps-based continuous delivery.

## Consequences
*   **Positive**:
    *   **Single Source of Truth**: Git is the only source of truth for infrastructure and applications.
    *   **Drift Detection**: ArgoCD automatically detects and corrects configuration drift.
    *   **Visibility**: Provides a UI to visualize application health and sync status.
    *   **Security**: No need to store cluster credentials in CI/CD pipelines (pull-based model).
*   **Negative**:
    *   Additional component to manage and secure.
    *   Learning curve for the GitOps workflow.

## Alternatives Considered
*   **Flux**: Similar capabilities, but ArgoCD's UI provides better visibility for developers.
*   **Jenkins/GitHub Actions (Push-based)**: Requires storing high-privilege cluster credentials in CI, which is a security risk.
