# Architecture Issues — Overview

This folder is the engineer-facing catalog of layered backend concerns we observed across three reference codebases — `intranet-cms`, `workflows`, and `id-services v3` — and the changes `no-frame-works` makes to fix each one. It exists to answer one question: **why is the strict layered architecture worth the discipline it demands?**

## How to read this folder

The folder has two parts:

1. **The layer files** (`01`, `02`, `03`) define the responsibility of each layer of the architecture — Handler, Application, Repository — and compare what each of the three reference codebases does at that layer. Every layer file follows the same six-section template: responsibility, what each repo does, what is good, what is bad, and per-repo improvement lists.
2. **The pain-point files** (`04+`) go deeper into specific cross-cutting concerns — permission duplication, validation scattering, error vocabulary, transaction management, and so on — that affect more than one layer or that warrant their own focused treatment.

If you are new to the template, read the three layer files in order. They give you the top-down map. The pain-point files are referenced from the layer files wherever they apply; follow the links as you read, or come back to them later.

## The core thesis

Layered architecture exists to minimize the **scope of understanding** required for any change. A concern is severe in proportion to *how much of the codebase a reader has to hold in their head* to add a feature, review a PR, or trace a bug.

The three axes every issue is scored against are all downstream of that single question:

- **Maintainability** — the cost of changing the code when requirements shift.
- **Top-level comprehensibility** — how quickly a new engineer can understand what a piece of code does without reading it line by line.
- **Reviewability** — how easy it is to catch problems in a PR without missing them.

If any of those three becomes "read the whole module to be sure," the architecture has failed.

## Severity scale

Each issue gets one of four levels. The scale is tight on purpose — if everything were CRITICAL, the scale would be meaningless.

| Level | Meaning |
|-------|---------|
| **LOW** | Annoying but workable. Shows up occasionally, doesn't block scaling. |
| **MEDIUM** | Affects productivity on every change in the affected area. Visible in review. |
| **HIGH** | Consistently causes bugs, review misses, or rework. A real blocker as the team grows. |
| **CRITICAL** | Makes the affected concern un-auditable or prevents future refactoring entirely. |

## The files

### Layer files — read these first

| # | File | Summary |
|---|------|---------|
| [01](01-handler-layer.md) | **Handler Layer** — the webserver / API handler | Responsibility + per-repo comparison + good / bad / improvements. Covers error mapping, actor extraction, request-struct assembly. |
| [02](02-application-layer.md) | **Application Layer** — the service | Responsibility + per-repo comparison + good / bad / improvements. Covers validation, permission, orchestration, fetch-mutate-save. |
| [03](03-repository-layer.md) | **Repository Layer** — data access gateway | Responsibility + per-repo comparison + good / bad / improvements. Covers go-jet usage, interfaces, `model`/`domain` boundary, "not found" contracts. |

### Pain-point files — cross-cutting concerns

| # | Issue | Severity | Summary |
|---|-------|----------|---------|
| [04](04-permission-duplication.md) | Permission checks duplicated inline at every call site | **CRITICAL** | `intranet-cms` has 265 copies of the permission-check block across 76 files. A missing check is visually identical to one that's there. |

More pain-point files are planned and will slot into `05+` as they are written. Each targets a specific concern that either crosses layer boundaries (so it cannot fit cleanly into one layer file) or is deep enough that it deserves a dedicated treatment.

## The three reference codebases

Every piece of evidence cites one of three real repositories:

- **`ck-andreidamian/intranet-cms`** — thin webserver handlers that `panic` on service errors; 127-line god functions in the `apiservice` package; flat `repository` package with tagged errors and no interface. The most visibly "stage-0" of the three.
- **`ck-andreidamian/workflows`** — partially layered. Handlers type-switch on typed errors. The `app_service/workflow` package decomposes a 120-line function into seven private helpers. `repository/` + `mutation/` split, with multi-table writes in one mutation function.
- **`ck-andreidamian/colorkrew-id-services`** (v3) — the cleanest of the three. Per-operation files with tests. `CreateOp`-struct-and-methods decomposition in the service. Still no interfaces, still no domain layer, still inconsistent "not found" contract.

The three repos are snapshots of the same drift: every attempt to layer the code stopped short of imposing the boundaries that would make the layering enforceable. The `no-frame-works` patterns are the minimal set that makes those boundaries stick.
