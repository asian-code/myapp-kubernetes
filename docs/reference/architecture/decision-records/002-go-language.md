# ADR 002: Go as the Primary Language

**Date**: 2025-12-21  
**Status**: Accepted

## Context
We are building high-performance microservices that need to handle concurrent requests and background processing efficiently. We need a language that is strongly typed, compiles to a single binary, and has a strong cloud-native ecosystem.

## Decision
We have decided to use **Go (Golang)** for all backend microservices.

## Consequences
*   **Positive**:
    *   **Performance**: Near C-level performance with memory safety.
    *   **Concurrency**: Goroutines make handling concurrent operations (like data processing) trivial.
    *   **Deployment**: Compiles to a small, static binary (scratch containers), ideal for Kubernetes.
    *   **Ecosystem**: The language of the cloud (Kubernetes, Docker, Terraform are written in Go).
*   **Negative**:
    *   Steeper learning curve for developers coming from interpreted languages (Python/JS).
    *   Verbose error handling (`if err != nil`).

## Alternatives Considered
*   **Python**: Good for data, but slower and harder to manage dependencies in containers.
*   **Node.js**: Good for I/O, but less performant for CPU-bound tasks and larger container images.
