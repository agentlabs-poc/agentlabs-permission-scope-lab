# Handbook of Authorization — Core Concepts & JSON

**PPT-001 · Approved presentation direction**

The user requested a PowerPoint focused on canonical vocabulary and JSON, then
approved the outline with revision mechanics left aside. The deck has no revision
or adoption lesson. Required revision fields remain in approved examples and are
visually muted; simplifying the presentation does not change the contracts.

- [Editable PowerPoint](authorization-core-concepts.pptx)
- [Slide-layout preview](preview.html)
- [Presenter notes and source links](speaker-notes.md)
- [Canonical Foundations reference](../theory/canonical-terms.md)

The 30-slide flow covers identity and tenant context; permission/scope/self;
role/grant/assignment; teams and administration; lineage and delegation;
enablement, validity and orphans; registration/bootstrap; endpoint declarations;
request/resolved views; results; and a connected Team1/Team2/Nutan read.

Text, JSON and diagrams are editable PowerPoint objects. Native SVG previews
are generated from the same drawing primitives for layout review; they are not
screenshots rendered by PowerPoint. Presenter notes supply rationale, assumptions,
source references and the distinction between framework choices and general concepts.

JSON is labeled as approved core shape, an identity-related JWT excerpt, or a
nested fragment. Pending role-publication, membership/owner, registration/root,
delegation and resolved-view formats are not invented. Example controls and
required upstream support must be valid in the stated scenario; examples do not
create live authority. Error codes/messages are illustrative.

## Status after the 19 September recompilation — STALE

The manuscript was recompiled against roughly two dozen decisions taken after the
first draft. **This deck was not**, and the generated artefacts here — the `.pptx`,
`deck-manifest.json` and the `preview/` SVGs — are behind the chapters they teach
from.

`build-deck.cjs` has had one correction applied: the evaluation-error example now
uses the published `AUTHORITY_TIMEOUT` rather than the retired illustrative
`AUTH_SERVICE_TIMEOUT`. The generated files still contain the old spelling, because
regenerating requires the isolated `pptxgenjs` install described below. **Source and
artefacts therefore disagree until a rebuild is run.**

Content that the deck does not yet carry, and would need before it is presented:

- administration is an ordinary grant on the Auth root chain, and what bounds each side
- a tenant's two namespaces, and why root establishment belongs to the first
- ownership as a grant, with the earlier relation superseded
- scope values opaque and matched exactly, with subtree excluded, and the one platform
  key whose values are resolved
- the authority-loading question and answer, which are no longer pending
- the policy's trusted correlations and its mount-time structural validation
- the ordered contributing chain, and the published code catalogue
- an unusable route not establishing a denial

Until then, treat the [Foundations reference](../theory/canonical-terms.md) and the
[implementation chapters](../implementation/06-auth-service.md) as authoritative and
this deck as a prior snapshot.

## Rebuild

The generator uses `pptxgenjs` 3.12.0. Install it into an isolated temporary
directory, not the website's package manifest, then run the generator with that
directory's `node_modules` on `NODE_PATH`:

```sh
NODE_PATH=/absolute/path/to/isolated/node_modules node handbook/slides/build-deck.cjs
```

The generator regenerates only its named PowerPoint, notes, manifest and preview
outputs in this folder. It preserves the handbook's source and discussion files.
`deck-manifest.json` is editorial verification metadata, not an authorization
contract or a complete schema validation test. No commit/push or site integration
is included in this presentation request.

## Verification

The file contains 30 editable slides and 30 sourced presenter notes. All 18 JSON
examples/fragments were parsed from the generated PowerPoint and compared with
the editorial manifest. ZIP integrity and all XML parts passed validation.
Text bounds were checked in the generated layouts, including measured font
widths for 661 preview text lines. The full layout overview and dense PUT/result
slides were visually inspected in the browser.

The PowerPoint uses Arial and Courier New; SVG previews use their metrically
compatible Liberation counterparts available in this environment. Native
PowerPoint/LibreOffice rendering was not available, so the visual check covers
the shared-layout SVG previews, not a claim of native application rendering.
Existing decision, runtime, test and package-manifest files were not changed.
