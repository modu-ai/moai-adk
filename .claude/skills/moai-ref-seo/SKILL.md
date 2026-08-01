---
name: moai-ref-seo
description: >
  Search-visibility and crawlability reference for web output: canonical URL
  discipline, per-page title and meta description uniqueness, robots.txt and
  sitemap.xml as host-derived artifacts, JSON-LD structured data with entity
  consistency, and the document-semantics rules that decide whether a machine can
  read a page at all. Agent-extending skill that amplifies web-output implementation
  and pre-ship review with production-grade indexing and structured-data patterns.
  NOT for: keyboard operability, visible focus indicators, and form-control labeling
  (accessibility owns those; delegated to the accessibility surface); generative-engine
  optimization, deliberately excluded as insufficiently settled; visual polish and
  interface detail (see moai-ref-ui-polish); API contract design (see
  moai-ref-api-patterns); security headers and hardening (see moai-ref-owasp-checklist
  and moai-ref-secops).

when_to_use: >
  Use when a project emits pages that a search engine or an automated reader will
  fetch: choosing canonical addresses, writing per-page metadata, emitting robots.txt
  or sitemap.xml, adding or reviewing JSON-LD, keeping entity naming consistent
  between the page and its serialized data, or auditing heading structure, image alt
  text, anchor text, and in-page fragment targets before shipping. Guidance stays at
  the protocol and output layer — addresses, response headers, markup, serialized
  data — so it applies to any stack that produces HTML.

user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-08-01"
  tags: "seo, canonical, structured-data, json-ld, sitemap, robots, metadata, crawlability, reference"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000
---

# Search Visibility Reference

## Target Agents

- `manager-develop` — applies these rules while implementing pages, routes, and serialized metadata output
- `/moai review` — pre-ship indexing review surface; equivalently available as a per-spawn `Agent(general-purpose)` with web-output review instructions

## Core Principle

A page is indexed on what a machine can fetch and parse, not on what a person sees
once everything has finished loading. Two failures dominate everything below: the
same content reachable at more than one address, and a claim asserted in serialized
data that nothing on the page corroborates. Every rule here is a specialization of
one of those two.

Figures that engines publish and then quietly revise — truncation widths, ranking
weights, crawl allowances — are deliberately absent. Where a limit matters, this
reference gives the decision rule and the measurement to take, so the guidance
outlives the figure.

## Document Semantics

The structure a parser reads before it reaches any content signal. Cheapest class of
defect to introduce, and the cheapest to detect.

| Rule | How to check it | Failure it prevents |
|---|---|---|
| Exactly one `h1` per document | Count `h1` elements in the rendered output of each page | Competing topic signals leave the subject ambiguous |
| No skipped heading level | Walk the heading sequence; each descent moves by at most one level | Section nesting a parser cannot reconstruct |
| Every image carries `alt` text | Each image element has a non-empty `alt` attribute, or an empty one paired with an explicit decorative marker | Meaning locked inside a binary the parser cannot open |
| Anchor text names its destination | Read each link's text with the surrounding sentence hidden, then ask where it goes | Instruction-shaped link text carries no signal about the target |
| Fragment targets resolve | For each in-page fragment link, confirm a matching identifier exists in the same rendered output | Navigation that silently lands nowhere |

Derive `alt` text from what the image depicts, never from its file name. Where an
image genuinely adds nothing a reader needs, mark it decorative on purpose instead
of inventing a description for it.

## Identity and Canonical Address

One resource, one address; everything else redirects to it.

| Decision | Rule |
|---|---|
| Which address is canonical | Pick one form per resource and declare it. The declared value must equal the address actually served |
| Trailing-slash variants | Choose one form and redirect the other permanently. Serving both splits one resource across two addresses |
| Case, query parameters, tracking suffixes | Normalize before serving. A parameter that does not change the response must not mint a second address |
| Parameterized routes | Build the declared value from the resolved parameters, never from a fixed string shared across the whole route |
| Retiring an address | Redirect permanently from the old form and update the declaration in the same change |

A declaration pointing at an address that redirects, errors, or serves different
content is worse than no declaration at all: it actively steers the reader away from
the page it appears on.

## Per-Page Metadata

| Field | Rule | Recurring defect |
|---|---|---|
| `title` | Unique per page, composed from a page-specific part plus a stable identifier for the site | The scaffold's default survives to production on every route |
| meta description | Unique per page, written for a person reading a result snippet | One description copied everywhere, or prose that describes nothing |
| Indexing directive | Set a default for public pages, then narrow per page class. Authenticated areas and internal tools are excluded deliberately, not merely left unlinked | A private surface gets indexed because nothing declared otherwise |
| Social preview fields | Present and absolute. A relative address in a field a third party fetches resolves against the wrong host | A share card that renders blank off-site |

Length budgets are a decision rule, not a constant: draft the text, observe where the
consuming surface truncates it, then cut to what you observed. Placeholder detection
belongs in the pre-ship check — scan the emitted output for the scaffold's own
default strings and fail when one survives.

## Structured Data

Serialized JSON-LD states what a page *is*. Its one hard constraint is correspondence
with the visible page.

| Decision | Rule |
|---|---|
| Which types to emit | Derive them from the page's actual role — an organization, a product, a piece of software, a place, an article. Emit the site-level type everywhere and the page-specific type only where it applies |
| Required fields | Each type has a small mandatory core. Emit that core completely rather than emitting several types partially |
| Addresses inside the data | Absolute, and resolving. Relative values are discarded without a warning |
| Several entities on one page | Emit one block holding all of them, give each a stable identifier, and cross-reference by identifier instead of nesting duplicate copies |
| Placement | Keep each type on the page it describes. A type repeated site-wide dilutes rather than reinforces |

**Mirroring rule** — every claim in the serialized data needs a visible counterpart on
the same page. A name, a description, a rating, or a question-and-answer pair that
exists only in the markup is an assertion the reader cannot check, and it is now
scored as a negative signal rather than a neutral one.

## Entity Consistency

Structured data identifies an entity. Consistency is what makes that identity hold
across independent sources.

| Surface | Requirement |
|---|---|
| Canonical name | One spelling, one casing, one suffix — the same in the serialized data, the page copy, and the headings |
| Owned profile links | Restricted to accounts this entity itself administers, each confirmed to resolve rather than bounce to a sign-in wall |
| Contact details | Name, address, and contact number identical across the serialized data, the visible page, and any third-party listings |
| Omission over approximation | A field with no accurate value is left out. A partial or invented one fragments the identity it was meant to establish |

Formatting differences that read as trivial to a person are distinct values to a
matcher. Fix one format per field and hold it everywhere the entity appears.

## Host-Derived Crawl Artifacts

`robots.txt` and `sitemap.xml` are output the site serves, not files carried between
environments.

| Artifact | Rule |
|---|---|
| `robots.txt` | Derive the absolute addresses it contains from the incoming request's host, so every environment serves a correct description of itself |
| `sitemap.xml` | Enumerate the addresses actually served. An entry for a removed page spends crawl attempts and signals staleness |
| Staying in sync | Adding or removing a route updates the enumeration in the same change. A hand-maintained list drifts within one iteration |
| Per-entry hints | Freshness and weighting hints attached to entries are advisory and widely discounted. Correctness of the address set matters far more than the hints on it |

Hardcoding one environment's host into either artifact is the defect that keeps
recurring, because it passes review while being correct in exactly one place.

## Delivery Chokepoints

| Concern | Rule |
|---|---|
| Response headers | Apply them at a single wrapping point every response passes through, error and redirect responses included. A header set only on the success path is missing exactly when a reader is most likely to be misled |
| Rendering mode | Treat "what does a reader receive in the first response?" as a question the project answers explicitly. Content that appears only after client-side work is content a non-executing reader may never see |
| Redirect chains | One hop to the canonical address. Each extra hop is another opportunity to lose the reader |
| Third-party origins | Connection hints for origins the page depends on belong ahead of the resources that use them. Measure the effect on this project's own pages rather than assuming one |

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "The framework already sets sensible metadata defaults" | A default is identical on every route by construction. Uniqueness is the entire purpose of the field, so it cannot be inherited. |
| "Both address forms serve the same page, so it makes no difference" | Two addresses serving one resource are two resources to a crawler. Which one survives is guessed unless a declaration and a redirect settle it deliberately. |
| "Serialized data may say more than the page shows — it is extra context" | Data with no visible counterpart is an unverifiable claim, and it is scored as one. |
| "Heading structure and `alt` text belong to accessibility, not to indexing" | They are the parse structure a machine uses to segment and describe a page. Sharing a mechanism with accessibility is not a reason to defer them. |
| "The address enumeration can be updated in a follow-up change" | A list maintained apart from the routes drifts immediately and then silently. Deriving it from the routes is the only version that stays true. |
| "The staging host in the artifact is fine, we swap it at deploy" | A hardcoded host is right in one environment and wrong in every other, including the one that gets indexed. |

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- Two or more `h1` elements on one page, or a heading sequence that skips a level
- A canonical declaration that does not match the address actually serving the page
- The same `title` or meta description reused across more than one route
- An image with no `alt` attribute, or `alt` text that repeats the file name
- Link text that reads as an instruction instead of naming a destination
- An in-page fragment link with no matching identifier in the rendered output
- `robots.txt` or `sitemap.xml` carrying a hardcoded host instead of one derived from the request
- Serialized JSON-LD asserting a name, description, or offer with no visible counterpart
- A relative address in a field a third party fetches — a share-preview image, a structured-data logo
- Response headers applied on the success path only, leaving redirect and error responses uncovered

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Every rendered page has exactly one `h1` and a heading sequence with no skipped level — report the counts observed, not the intent
- [ ] Each route's canonical declaration was fetched and returned that same page with no redirect in between
- [ ] `title` and meta description were collected across all routes; no duplicates and no scaffold default strings remain
- [ ] Every image element has an `alt` attribute, and decorative images are marked as such on purpose
- [ ] Every in-page fragment link was resolved against the rendered output of the page carrying it
- [ ] `robots.txt` and `sitemap.xml` were fetched from the deployed host and every address in them resolves — quote the address set observed
- [ ] Serialized JSON-LD parses, its addresses are absolute, and each claim it makes has a visible counterpart quoted from the page
- [ ] Response headers were observed on a success response, on a redirect response, and on an error response

<!-- moai:evolvable-end -->
