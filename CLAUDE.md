# dafcli

A CLI for querying Danish grunddata via the Datafordeler GraphQL APIs (Klimadatastyrelsen).

## Stack

- Go, structured as a cobra CLI
- `daf/` package: GraphQL client, DAWA client, typed queries per register
- `cmd/` package: cobra subcommands
- Auth via env var `DAF_API_KEY`, with macOS Keychain fallback (service `dafcli` / account `DAF_API_KEY`)

## CLI commands

```
dafcli adresse <text>                       # DAWA address resolution (no auth)
dafcli jordstykke <matrikel> [--ejerlav]    # MAT_Jordstykke; returns BFE via samletFastEjendomLokalId
dafcli sfe <bfe>                            # MAT_SamletFastEjendom by BFE-nummer
dafcli bygning --husnummer <UUID>           # BBR_Bygning by access-address UUID
dafcli inspect <text>                       # one-shot DAWA → MAT → BBR chain
dafcli probe <Type> <field>... [--register MAT|BBR|...]   # schema discovery (introspection blocked)
```

All subcommands accept `--json`. `adresse` also has `--raw`.

Build: `go build -o dafcli .`
Install to PATH: `go install ./...` (binary lands in `~/go/bin`).

## Skill

A Claude Code skill named `daf` (in your skills directory) delegates to this
binary and handles the natural-language → subcommand triage. Same pattern as
the sibling CLIs in this suite.

## Datafordeler GraphQL conventions

These are non-obvious — the `daf` package encapsulates them so subcommands stay
clean. They were verified empirically against prod (Lunar Bank A/S, Hack
Kampmanns Plads 10) — Datafordeler's own docs are vague.

- **Auth**: `?apiKey=<KEY>` query parameter (NOT `Authorization` header — that returns 401).
- **Bitemporal**: every query requires `registreringstid` + `virkningstid` (ISO-8601 with `.000Z`). Missing them → `DAF-GQL-0009`.
- **Connection types**: root queries return `{ nodes { ... } }` Relay envelopes.
- **Filtering**: HotChocolate `where: { field: { eq: "value" } }`. Operators: `eq, neq, in, nin, contains, startsWith, gt, gte, lt, lte`.
- **Schema introspection blocked** (`HC0046`) — use `dafcli probe` to discover field names.
- **Field naming**: `id_lokalId` (capital L), `matrikelnummer` (full word). Many obvious guesses (`bfeNummer`, `ejerlav`, `samletFastEjendom`) don't exist on Jordstykke.
- **Danish chars in field names** (ø/å/æ) trigger `HC0011` unless Unicode-escaped — v1 omits them.
- **Error messages MUST scrub the API key**: Go's `net/http` embeds the request URL in transport errors, leaking `?apiKey=…` into logs. See `daf.Client.scrubError`.

## Cross-register navigation

```
DAWA adresse text  →  adgangsadresseUUID  +  matrikelnummer  +  ejerlavskode
                                    │
                                    ▼
                       MAT_Jordstykke (where matrikelnummer eq …)
                                    │
                                    └─ samletFastEjendomLokalId  =  BFE-nummer
                                                  │
                                                  ▼
                                  MAT_SamletFastEjendom (where id_lokalId eq BFE)

DAWA adgangsadresseUUID  =  DAR Husnummer.id_lokalId  =  BBR Bygning husnummer
```

## Endpoints

- GraphQL prod: `https://graphql.datafordeler.dk/<REGISTER>/v1` (REGISTER ∈ {MAT, BBR, DAR, DAGI, EJF, ...})
- DAWA (open): `https://api.dataforsyningen.dk/{adresser, adgangsadresser, jordstykker, ...}`
- Legacy REST (phasing out 30 June 2026): `https://services.datafordeler.dk/<REGISTER>/...`
- OAuth2 token (for EJF / certificate auth): `https://auth.datafordeler.dk/realms/distribution/protocol/openid-connect/token`

## Out of scope

- **EJF** (Ejerfortegnelsen / owners) — gated; requires Klimadatastyrelsen approval + OAuth2 client_credentials.
- **Tinglysning** (deeds, mortgages, byrder, pant) — not on Datafordeler. Domstolsstyrelsen eTL system; OCES III + tilslutningsaftale.
- **CPR / VUR / SVR** — gated and usually paid.
- **CVR** — Datafordeler exposes it, but for ad-hoc company lookups [`virkcli`](https://github.com/kaspermunck/virkcli) wraps the dedicated VIRK distribution endpoint more ergonomically.

## Memory

**Always read `.claude/MEMORY.md` at the start of every session.**
Write new memories to `.claude/memory/` and update the index in `.claude/MEMORY.md`.
