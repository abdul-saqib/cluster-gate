### Contributing to Cluster-Gate

Thank you for your interest in contributing to cluster-gate!

Our goal is to provide a simple, reliable, and extensible Kubernetes controller that exposes Deployments to the outside world using NodePort Services. We appreciate your time and contributions toward improving this project.

This guide outlines how to contribute effectively and collaboratively.

At present, cluster-gate is distributed solely as a Community Edition (CE)—fully open source and free. In the future, an Enterprise Edition may be introduced, and this document will be updated accordingly.

Contributor License Agreement (CLA)

By contributing to this repository:

- Individuals agree to the [Individual Contributor License Agreement][./agreements/individual_contributor.md]

- Organizations agree to the [Corporate Contributor License Agreement][./agreements/corporate_contributor.md]

Submitting a pull request or patch indicates acceptance of the appropriate agreement.

### Security Vulnerability Disclosure

If you discover or suspect a security vulnerability—especially anything involving Kubernetes API access, privilege escalation, or unintended workload exposure—please report it privately.

Email: [saqib.abdul@infracloud.io][mailto:saqib.abdul@infracloud.io]

- Do NOT create public GitHub issues for potential security vulnerabilities.

- Expect a response within two business days.

- Most issues are typically resolved within a week, but timelines may vary depending on severity and complexity.

Responsible disclosure helps keep the Kubernetes ecosystem safe.

### Closing Policy for Issues and Merge Requests

We aim for cluster-gate to become a widely-used tool in the cloud-native community. As project activity grows, the maintainers may need to close issues or merge requests that do not follow our guidelines:

- Issues and merge requests not aligned with this document may be closed without notice.

- Maintainers will try to provide a reason, but this may not always be possible.

- Please treat all volunteers, contributors, and maintainers with respect.

All issues and pull requests must be written in English and remain appropriate for all audiences.

### Development Guidelines

#### 🔧 Setting Up Your Environment

Before contributing, ensure you have:

- Go (latest supported version)

- Docker

- Make

- A Kubernetes cluster for testing (KinD, Minikube, k3d, etc.)

#### 🧪 Testing Changes

If your contribution affects controller behavior—such as reconciliation logic, NodePort handling, or RBAC—please:

Add or update unit tests where appropriate.

Validate your changes in a real or local Kubernetes cluster.

#### 📦 Commit Message Standards

Use clear and concise commit messages. Conventional commit style is preferred:

~~~
feat: add label-based filtering for watched Deployments
fix: correct NodePort collision handling
docs: update configuration examples
~~~

#### 📂 Code Style and Patterns

Please follow:

- Standard go fmt formatting

- Idiomatic Go error handling

- controller-runtime best practices

Deterministic, idempotent reconciliation loops

#### 📝 Pull Request Guidelines

A good PR should include:

- A clear explanation of the change

- Updated documentation, if needed

- Relevant tests

A description of how you validated the update

### Have Questions?

If you need help getting started or want clarification on anything:

Email: [saqib.abdul@infracloud.io](mailto:saqib.abdul@infracloud.io)

We’re happy to help.
Thank you for contributing to cluster-gate!