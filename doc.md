# CrossTrace Documentation

## CrossTrace

CrossTrace is a manifest‑driven Coral agent that transforms raw user or system events into verifiable, on‑chain receipts and NFT proofs.  
It combines Go‑powered journaling, MistraAI orchestration, and Crossmint NFT minting to deliver a fully auditable, schema‑evolving event pipeline.

## Design

CrossTrace was designed as modular. Each components does a operation and doesn't interfere with each other but instead support each other. 
When a user or app wants to register an events it sends the raw message related to event to CrossTrace upon reception, it first perform a basic sanitization to ensure that the message isn't dangerous to the Ai or malicious that include checking for hidden characters or symbols
only sanitized message are handled by Ai once the Ai receive the message it is presented with two tools LogEvent and SealBatch. LogEvent only action is to store the given event in memory and nothing more, SealBatch takes all the events stored by LogEvent based on the amount of event create a Batch(if events>1)or not and generate a receipt and anchoring it on solana using crossmint

### SealBatch

SealBatch handles operations such as batching, generating merkletree, storing into localDB
and anchoring receipt on solana

## Usage
