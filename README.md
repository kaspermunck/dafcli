# dafcli

CLI for [Datafordeler](https://datafordeler.dk) — Klimadatastyrelsens distribuering af danske grunddata. Targets the modern GraphQL stack (in prod since November 2025; the legacy webbruger/tjenestebruger REST flow is phased out 30 June 2026).

## Coverage

| Register | Subcommand | Notes |
|---|---|---|
| **DAWA** (open, no auth) | `dafcli adresse` | Adresse-tekst → adgangsadresseUUID + matrikel + ejerlav |
| **MAT** (Matriklen2) | `dafcli jordstykke`, `dafcli sfe` | Parcel by matrikelnummer; SFE/BFE lookup |
| **BBR** | `dafcli bygning` | Buildings by access-address UUID |
| (all) | `dafcli inspect` | DAWA → MAT → BBR chain in one shot |
| (all) | `dafcli probe` | Schema discovery — bypasses Datafordeler's introspection block |

**Out of scope** (for now): EJF (gated, OAuth2 + Klimadatastyrelsen approval), DAGI (legacy REST works), CPR/VUR/SVR (gated, paid). Tinglysning is not on Datafordeler — it lives in the Domstolsstyrelsen eTL system and requires OCES + tilslutningsaftale; use [tinglysning.dk](https://tinglysning.dk) for ad-hoc lookups with MitID.

## Install

Via Homebrew (recommended):

```sh
brew install kaspermunck/tap/dafcli
```

From source:

```sh
go install github.com/kaspermunck/dafcli@latest
```

## Claude Code skill

A [Claude Code](https://claude.com/claude-code) skill ships inside the Homebrew formula. After `brew install`, enable it once with the symlink Homebrew prints in its caveats:

```sh
mkdir -p ~/.claude/skills && ln -sfn "$(brew --prefix dafcli)/share/dafcli/skill" ~/.claude/skills/daf
```

Re-run `brew upgrade dafcli` to update both the binary and the skill atomically.

## Auth

Datafordeler's GraphQL endpoint requires an API key. Sign up at [portal.datafordeler.dk](https://portal.datafordeler.dk) → Administration → IT-system → API key to obtain one.

Set it in your environment:

```sh
export DAF_API_KEY="<your-key>"
```

macOS Keychain pattern (recommended — keeps the secret off disk):

```sh
security add-generic-password -s "dafcli" -a "DAF_API_KEY" -w '<your-key>' -U
```

`dafcli` reads the env var first; if empty, it falls back to the same keychain entry directly, so no shell-rc export is needed.

## Usage

```sh
# Address → matrikel + ejerlav (DAWA, no auth)
dafcli adresse "Novo Allé 1, 2880 Bagsværd"

# Parcel by matrikelnummer (narrow with --ejerlav)
dafcli jordstykke 2hq --ejerlav 12751

# SFE by BFE-nummer
dafcli sfe 7870540

# Buildings on an access-address
dafcli bygning --husnummer 0a3f507c-3f47-32b8-e044-0003ba298018

# Full chain in one shot
dafcli inspect "Novo Allé 1, 2880 Bagsværd"

# Schema discovery (Datafordeler blocks GraphQL introspection)
dafcli probe Jordstykke matrikelnummer ejerlavskode bfeNummer status
# → prints which fields the server accepts vs. rejects

# --json for programmatic output
dafcli jordstykke 2hq --ejerlav 12751 --json
```

## Datafordeler GraphQL quirks worth knowing

Empirically verified — the docs are vague:

- **Auth is a query parameter, not a header.** `?apiKey=<KEY>`. `Authorization: apikey/Bearer <KEY>` returns 401.
- **Every query requires bitemporal arguments**: `registreringstid` + `virkningstid` (ISO-8601 with `.000Z`). Missing them returns `DAF-GQL-0009`.
- **Root queries return Relay Connection types.** Always wrap selections in `{ nodes { ... } }`.
- **Filtering uses HotChocolate `where:`** — `where: { fieldName: { eq: "value" } }`. Operators: `eq, neq, in, nin, contains, startsWith, gt, gte, lt, lte`.
- **Schema introspection is blocked** (`HC0046`). Use `dafcli probe` instead.
- **Field naming is unforgiving**: `id_lokalId` (capital L), `matrikelnummer` (full word, lowercase). Many obvious guesses (`bfeNummer`, `ejerlav`, `samletFastEjendom`) don't exist.
- **Danish chars in field names** (`byg026Opførelsesår`) trigger `HC0011` unless Unicode-escaped; v1 omits them.

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
                                                  │
                                                  └─→ EJF (gated)
                                                  └─→ Tinglysning (separate stack)

DAWA adgangsadresseUUID  →  BBR_Bygning (where husnummer eq UUID)
                                       │
                                       └─ bygningsnummer + anvendelse + BFE-relation
```

## Security note

`dafcli` scrubs the API key from error messages (`net/http` transport errors otherwise embed the full request URL — including the `?apiKey=` parameter — which would leak the secret into logs).

If you suspect the key has been exposed, rotate it in the portal and overwrite the keychain entry:

```sh
security add-generic-password -s "dafcli" -a "DAF_API_KEY" -w '<new-key>' -U
```

## Acknowledgements

`dafcli` accesses data published via the [Datafordeler](https://datafordeler.dk) platform (operated by [Klimadatastyrelsen](https://klimadatastyrelsen.dk)) and via [DAWA](https://dawa.aws.dk) (Danmarks Adressers Web API, also Klimadatastyrelsen). The individual registers are owned by:

- **MAT** (Matriklen) — [Geodatastyrelsen](https://gst.dk)
- **BBR** (Bygnings- og Boligregistret) — [Vurderingsstyrelsen](https://vurderingsportalen.dk) in collaboration with the Danish municipalities
- **DAR** (Danmarks Adresseregister) and **DAGI** (Danmarks Administrative Geografiske Inddeling) — [Klimadatastyrelsen](https://klimadatastyrelsen.dk)

`dafcli` is an independent open-source tool and is not affiliated with or endorsed by any of these authorities.

When publishing work derived from this data, credit the originating register, e.g.:

> Kilde: Matriklen / Geodatastyrelsen
> Kilde: BBR / Vurderingsstyrelsen og kommunerne

## License

This software is released under the [MIT License](LICENSE). The data it fetches is published by the registers listed above under their respective terms — most are open data, but consult each authority's terms before commercial bulk reuse.
