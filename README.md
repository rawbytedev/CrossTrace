# CrossTrace — Schema‑Driven Journaling & NFT Receipt Agent

![Top Language](https://img.shields.io/github/languages/top/rawbytedev/CrossTrace) ![Tests](https://github.com/rawbytedev/CrossTrace/actions/workflows/tests.yml/badge.svg)

CrossTrace is a manifest‑driven Coral agent that transforms raw user or system events into verifiable, on‑chain receipts and NFT proofs.  
It combines Go‑powered journaling, MistraAI orchestration, and Crossmint NFT minting to deliver a fully auditable, schema‑evolving event pipeline.

---

## Features

- Safe Intake Pipeline — Sanitizes and validates raw messages from Coral before processing.
- Schema‑Driven Events — Enforces structured event formats with YAML‑defined schemas.
- AI‑Powered Decisions — Uses MistraAI to classify actions (mint, rollback, escalate) and suggest schema evolution.
- On‑Chain Anchoring — Stores event checksums on Solana for immutable verification.
- NFT Receipts — Mints claimable NFT proofs via Crossmint with embedded semantic metadata.
- Audit & Replay — Journaling engine supports rollback, replay, and integrity checks.

---

## Architecture

1. Setup Interface — Configure cache size, database dir, log dir, max message size.
2. Sanitization — Validate UTF‑8, enforce size limits, strip suspicious characters.
3. PostEntry Creation — Generate checksums and safe event objects.
4. MistraAI Orchestration — Interpret events, decide actions, suggest schema patches.
5. Action Execution — Mint receipts, rollback state, or escalate for review.
6. Crossmint Integration — Mint NFT receipts with claim URLs for end‑users.

---

## Tech Stack

- Language: Go
- Storage: BadgerDB/PebbleDB for journaling and indexing(allows plugging custom db)
- Blockchain: Solana Devnet for anchoring receipts
- AI: MistraAI for event interpretation and decision‑making
- NFTs: Crossmint for minting and distribution

---

## Getting Started

### Clone repo

`bash
git clone https://github.com/rawbytedev/crosstrace.git
`

`bash
cd crosstrace
`

### Build

`bash
go build -o crosstrace ./cmd/agent
`

### Start agent

`bash
./crosstrace.exe
`

---

## Hackathon Context

Built for the Internet of Agents Hackathon @ Solana Skyline (Sept 14–21, 2025)

---

## License

All right reserved.
No license is granted for reuse, distrubution, or modification without written permission.
