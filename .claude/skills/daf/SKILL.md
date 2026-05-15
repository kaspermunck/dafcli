---
name: daf
description: >
  Query the Danish state's grunddata distribution platform — Datafordeler —
  via the `dafcli` binary. Wraps MAT (Matriklen / cadastre), BBR (buildings),
  DAR/DAWA (addresses), and DAGI (administrative geography); EJF is gated
  pending Klimadatastyrelsen approval. Trigger when the user asks about a
  Danish address, parcel, building, BFE, matrikelnummer, ejerlav, or property
  context. Examples: "find matrikel for Vesterbrogade 1, København", "what
  buildings are on parcel 17a Frederiksberg", "look up BFE 9907883",
  "fortæl mig alt om denne adresse". For mortgages, deeds, byrder (Tinglysning)
  Datafordeler doesn't cover it — use tinglysning.dk (free electronic
  tingbogsattest with MitID) or the Domstolsstyrelsen eTL system (OCES III +
  tilslutningsaftale).
tools: [Bash]
---

# daf — Danish grunddata via dafcli

You query Datafordeler through the `dafcli` binary (built from `~/dev/dafcli`).
The binary owns the GraphQL conventions; you compose subcommands and present
the results in Danish.

## Prerequisites

- `dafcli` on PATH (built from `~/dev/dafcli`: `go build -o dafcli .`).
- `DATAFORDELER_API_KEY` env var, or the macOS Keychain entry
  service `datafordeler` / account `DATAFORDELER_API_KEY`. `dafcli` reads
  the env var first and falls back to keychain automatically.

If the binary is missing, tell the user to build it. **Never accept a pasted
API key in chat** — secrets stay in keychain / shell rc.

## Subcommands

| Subcommand | What |
|---|---|
| `dafcli adresse <text>` | DAWA address resolution (no auth). Returns `adgangsadresseUUID`, matrikel, ejerlav, kommunekode, coords. The join-key starting point for everything else. |
| `dafcli jordstykke <matrikelnummer> [--ejerlav <kode>]` | MAT_Jordstykke lookup. Returns `id_lokalId`, `registreretAreal`, status, and crucially `samletFastEjendomLokalId` — the BFE-nummer. |
| `dafcli sfe <bfe-nummer>` | MAT_SamletFastEjendom by BFE. Confirms the SFE exists and is in status `Gældende`. |
| `dafcli bygning --husnummer <UUID>` | BBR_Bygning by access-address UUID. Returns bygningsnummer + anvendelseskode + label. |
| `dafcli inspect <address>` | One-shot DAWA → MAT → BBR chain. The default "tell me everything about this address" path. |
| `dafcli probe <Type> <field>... [--register MAT\|BBR\|...]` | Schema discovery. Datafordeler blocks GraphQL introspection; this probe submits a wide selection and reports which fields the server accepts. Use when extending the binary to new types. |

All subcommands accept `--json` for structured output. `dafcli adresse` also
has `--raw`.

## Entity triage

| User has / asks about | First call |
|---|---|
| A Danish address (text) | `dafcli adresse "<text>"` (or `inspect` for full chain) |
| `adgangsadresseUUID` | `dafcli bygning --husnummer <UUID>` |
| matrikel + ejerlav | `dafcli jordstykke <matrikel> --ejerlav <kode>` |
| BFE-nummer | `dafcli sfe <bfe>` |
| "fortæl mig alt om denne adresse" | `dafcli inspect "<text>"` |
| New field on an unknown type | `dafcli probe <Type> <candidate-fields>...` |

## Hard limits — say no clearly

- **Tinglysning** (deeds, mortgages, byrder, pant) — not on Datafordeler.
  Redirect: `tinglysning.dk` (free electronic tingbogsattest with MitID), or
  the Domstolsstyrelsen eTL SOAP system (OCES III + tilslutningsaftale, weeks
  of provisioning).
- **EJF** (Ejerfortegnelsen / owners) — gated. Requires "Anmodning om adgang
  til Ejerfortegnelsen" approved by Klimadatastyrelsen + OAuth2
  client_credentials. The `inspect` command says explicitly that ejer is
  unavailable when EJF isn't approved.
- **CPR / VUR / SVR** — gated and usually paid. Out of scope.
- **CVR** — Datafordeler has it but `/virkcli` is simpler for ad-hoc lookups.

## Output discipline

- Render all chat output in Danish.
- For BBR anvendelseskoder, the binary already resolves to Danish labels
  (e.g. `321 (Bygning til kontor)`) — relay verbatim.
- For an "inspect" result, summarize in Danish with the address as headline,
  then matrikel/BFE/areal, building count + primary anvendelse, and an
  explicit "ejer/pant ikke tilgængelig" footnote.
- Don't dump raw GraphQL JSON. The binary's plain output is already user-
  ready; only fall back to `--json` when the user explicitly asks.
- Cite "Klimadatastyrelsen / Datafordeler" (or "DAWA") so the user can
  trace back.

## Error recovery

- **`DATAFORDELER_API_KEY must be set`** — tell user to set the env var or
  add the keychain entry. Never paste the key into chat.
- **HTTP 401** — API key rejected. Verify the IT-system's tjeneste
  subscription in `portal.datafordeler.dk` → Administration → IT-system.
- **`graphql error [DAF-GQL-0009]`** — missing bitemporal arg. Shouldn't
  happen from `dafcli` (it always sends them), so this signals the user has
  invoked the binary with an unusual query. Re-run the relevant subcommand
  default.
- **"context deadline exceeded"** — Datafordeler can be flaky. Retry once.
- **`bygning` returns "no buildings"** — the husnummer UUID may be wrong, or
  the address truly has no registered buildings (vacant lot, samlet
  ejendom on a different husnummer). Cross-check with `jordstykke` to see
  the BFE, then look at the SFE's other jordstykker.
- **`probe` shows all fields invalid** — the type name itself is wrong, or
  the register is wrong. Try a different `--register`.

## Discovery workflow (for extending the binary)

When the user asks for a Datafordeler field the binary doesn't yet expose:

1. Run `dafcli probe <Type> <candidate1> <candidate2> ...` with reasonable
   Danish-grunddata field-name guesses.
2. Note the valid set.
3. Add the new fields to `~/dev/dafcli/daf/<register>.go` (struct + query
   string) and rebuild.
4. Update this SKILL.md if a new subcommand or join key emerges.

## Cite sources

For grunddata results, cite "Klimadatastyrelsen / Datafordeler". For DAWA
results, cite "DAWA / Styrelsen for Dataforsyning". Open license — attribution
is polite, not required.
